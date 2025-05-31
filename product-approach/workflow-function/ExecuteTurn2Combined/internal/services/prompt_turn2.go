package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/templateloader"
)

// PromptServiceTurn2 defines Turn2 prompt generation service
type PromptServiceTurn2 interface {
	GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.TurnResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error)
}

// promptServiceTurn2 implements PromptServiceTurn2
type promptServiceTurn2 struct {
	templateLoader templateloader.TemplateLoader
	cfg            *config.Config
	log            logger.Logger
}

// NewPromptServiceTurn2 creates a new PromptServiceTurn2 instance
func NewPromptServiceTurn2(loader templateloader.TemplateLoader, cfg *config.Config, log logger.Logger) (PromptServiceTurn2, error) {
	if loader == nil {
		return nil, errors.NewInternalError("prompt_service_init", fmt.Errorf("template loader cannot be nil"))
	}
	if cfg == nil {
		return nil, errors.NewInternalError("prompt_service_init", fmt.Errorf("config cannot be nil"))
	}
	if log == nil {
		return nil, errors.NewInternalError("prompt_service_init", fmt.Errorf("logger cannot be nil"))
	}

	return &promptServiceTurn2{
		templateLoader: loader,
		cfg:            cfg,
		log:            log.WithFields(map[string]interface{}{"service_component": "PromptServiceTurn2"}),
	}, nil
}

// getTurn2TemplateType selects the Turn 2 template type based on verification type
func getTurn2TemplateType(verificationType string) string {
	switch verificationType {
	case schema.VerificationTypeLayoutVsChecking:
		return "turn2-layout-vs-checking"
	case schema.VerificationTypePreviousVsCurrent:
		return "turn2-previous-vs-current"
	default:
		return "turn2-default"
	}
}

// GenerateTurn2PromptWithMetrics renders the Turn2 prompt and returns metrics
func (p *promptServiceTurn2) GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.TurnResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error) {
	start := time.Now()
	if vCtx == nil {
		return "", nil, errors.NewValidationError("verification context required", nil)
	}
	if systemPrompt == "" {
		return "", nil, errors.NewValidationError("system prompt cannot be empty", map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	templateType := getTurn2TemplateType(vCtx.VerificationType)

	templateData := map[string]interface{}{
		"VerificationContext": vCtx,
		"SystemPrompt":        systemPrompt,
		"Turn1Response":       turn1Response,
	}

	renderedPrompt, err := p.templateLoader.RenderTemplateWithVersion(
		templateType,
		p.cfg.Prompts.Turn2TemplateVersion,
		templateData,
	)
	if err != nil {
		p.log.Error("turn2_template_rendering_failed", map[string]interface{}{
			"error":            err.Error(),
			"template_type":    templateType,
			"template_version": p.cfg.Prompts.Turn2TemplateVersion,
			"verification_id":  vCtx.VerificationId,
		})
		return "", nil, errors.WrapError(err, errors.ErrorTypeTemplate, "failed to render Turn 2 prompt template", false)
	}

	processor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      templateType,
			TemplateVersion: p.cfg.Prompts.Turn2TemplateVersion,
			TemplateType:    templateType,
			Content:         renderedPrompt,
		},
		ContextData:     templateData,
		ProcessedPrompt: renderedPrompt,
		ProcessingTime:  time.Since(start).Milliseconds(),
		InputTokens:     0,
		OutputTokens:    0,
	}

	p.log.Info("turn2_prompt_generated", map[string]interface{}{
		"verification_id":  vCtx.VerificationId,
		"template_type":    templateType,
		"template_version": p.cfg.Prompts.Turn2TemplateVersion,
		"prompt_length":    len(renderedPrompt),
	})

	return renderedPrompt, processor, nil
}
