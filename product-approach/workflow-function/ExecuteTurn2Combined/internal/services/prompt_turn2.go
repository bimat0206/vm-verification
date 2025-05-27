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
	GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error)
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

// GenerateTurn2PromptWithMetrics renders the Turn2 prompt and returns metrics
func (p *promptServiceTurn2) GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error) {
	start := time.Now()
	if vCtx == nil {
		return "", nil, errors.NewValidationError("verification context required", nil)
	}
	if turn1Response == nil {
		return "", nil, errors.NewValidationError("turn1Response cannot be nil", map[string]interface{}{"verification_id": vCtx.VerificationId})
	}
	if systemPrompt == "" {
		return "", nil, errors.NewValidationError("system prompt cannot be empty", map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	templateType, err := schema.GetTemplateForUseCase(vCtx.VerificationType, 2)
	if err != nil {
		return "", nil, errors.WrapError(err, errors.ErrorTypeTemplate, "unknown verification type", false)
	}

	templateData := map[string]interface{}{
		"VerificationContext": vCtx,
		"SystemPrompt":        systemPrompt,
		"Turn1Response":       turn1Response,
	}

	if len(turn1RawResponse) > 0 {
		var raw interface{}
		if err := json.Unmarshal(turn1RawResponse, &raw); err == nil {
			templateData["Turn1RawResponse"] = raw
		} else {
			p.log.Warn("failed_to_parse_turn1_raw_response", map[string]interface{}{
				"error":           err.Error(),
				"verification_id": vCtx.VerificationId,
			})
			templateData["Turn1RawResponse"] = string(turn1RawResponse)
		}
	}

	rendered, err := p.templateLoader.RenderTemplateWithVersion(templateType, p.cfg.Prompts.Turn2TemplateVersion, templateData)
	if err != nil {
		return "", nil, errors.WrapError(err, errors.ErrorTypeTemplate, "failed to render turn2 template", false)
	}

	processor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      templateType,
			TemplateVersion: p.cfg.Prompts.Turn2TemplateVersion,
			TemplateType:    templateType,
			Content:         rendered,
		},
		ContextData:     templateData,
		ProcessedPrompt: rendered,
		ProcessingTime:  time.Since(start).Milliseconds(),
		InputTokens:     0,
		OutputTokens:    0,
	}

	p.log.Info("turn2_prompt_generated", map[string]interface{}{
		"verification_id":  vCtx.VerificationId,
		"template_type":    templateType,
		"template_version": p.cfg.Prompts.Turn2TemplateVersion,
		"prompt_length":    len(rendered),
	})

	return rendered, processor, nil
}
