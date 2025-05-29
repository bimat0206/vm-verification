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
	if systemPrompt == "" {
		return "", nil, errors.NewValidationError("system prompt cannot be empty", map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	promptText := "Analyze the provided checking image against the reference context established in the previous turn. Follow all instructions in the System Prompt to identify discrepancies and generate the full verification report."

	templateData := map[string]interface{}{
		"VerificationContext": vCtx,
		"SystemPrompt":        systemPrompt,
		"Turn1Response":       turn1Response,
	}

	processor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      "turn2-static",
			TemplateVersion: p.cfg.Prompts.Turn2TemplateVersion,
			TemplateType:    "turn2-static",
			Content:         promptText,
		},
		ContextData:     templateData,
		ProcessedPrompt: promptText,
		ProcessingTime:  time.Since(start).Milliseconds(),
		InputTokens:     0,
		OutputTokens:    0,
	}

	p.log.Info("turn2_prompt_generated", map[string]interface{}{
		"verification_id":  vCtx.VerificationId,
		"template_type":    "turn2-static",
		"template_version": p.cfg.Prompts.Turn2TemplateVersion,
		"prompt_length":    len(promptText),
	})

	return promptText, processor, nil
}
