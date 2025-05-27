package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/templateloader"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// PromptServiceTurn2 extends PromptService with Turn2-specific functionality
type PromptServiceTurn2 interface {
	PromptService

	// Turn2 prompt generation methods
	GenerateTurn2Prompt(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, error)
	GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error)
}

// promptServiceTurn2 implements PromptServiceTurn2
type promptServiceTurn2 struct {
	*promptService
}

// NewPromptServiceTurn2 creates a new PromptServiceTurn2 instance
func NewPromptServiceTurn2(templateLoader templateloader.TemplateLoader, cfg config.Config, log logger.Logger) PromptServiceTurn2 {
	baseService := NewPromptService(templateLoader, cfg, log).(*promptService)
	return &promptServiceTurn2{
		promptService: baseService,
	}
}

// GenerateTurn2Prompt generates a Turn2 prompt based on the verification type and Turn1 results
func (p *promptServiceTurn2) GenerateTurn2Prompt(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, error) {
	prompt, _, err := p.GenerateTurn2PromptWithMetrics(ctx, vCtx, systemPrompt, turn1Response, turn1RawResponse)
	return prompt, err
}

// GenerateTurn2PromptWithMetrics generates a Turn2 prompt with metrics
func (p *promptServiceTurn2) GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.Turn1ProcessedResponse, turn1RawResponse json.RawMessage) (string, *schema.TemplateProcessor, error) {
	startTime := time.Now()

	// Validate inputs
	if err := p.validateInputs(vCtx, systemPrompt); err != nil {
		return "", nil, err
	}

	// Validate Turn1 response
	if turn1Response == nil {
		return "", nil, errors.NewValidationError(
			"turn1Response cannot be nil",
			map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	// Determine template type based on verification type
	templateType := ""
	switch vCtx.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		templateType = "turn2-layout-vs-checking"
	case schema.VerificationTypePreviousVsCurrent:
		templateType = "turn2-previous-vs-current"
	default:
		return "", nil, errors.NewValidationError(
			fmt.Sprintf("unsupported verification type for Turn2: %s", vCtx.VerificationType),
			map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	// Build template context
	templateContext := map[string]interface{}{
		"VerificationContext": vCtx,
		"SystemPrompt":        systemPrompt,
		"Turn1Response":       turn1Response,
	}

	// Add raw Turn1 response if available
	if len(turn1RawResponse) > 0 {
		var rawResponseObj interface{}
		if err := json.Unmarshal(turn1RawResponse, &rawResponseObj); err == nil {
			templateContext["Turn1RawResponse"] = rawResponseObj
		} else {
			p.log.Warn("failed_to_parse_turn1_raw_response", map[string]interface{}{
				"error":           err.Error(),
				"verification_id": vCtx.VerificationId,
			})
		}
	}

	// Render template
	var renderedTemplate string
	var err error

	if p.cfg.TemplateVersioning.Enabled {
		renderedTemplate, err = p.templateLoader.RenderTemplateWithVersion(
			templateType,
			p.cfg.TemplateVersioning.Version,
			templateContext,
		)
	} else {
		renderedTemplate, err = p.templateLoader.RenderTemplate(
			templateType,
			templateContext,
		)
	}

	if err != nil {
		return "", nil, errors.WrapError(err, errors.ErrorTypeTemplate,
			"failed to render Turn2 prompt template", false).
			WithContext("template_type", templateType).
			WithContext("verification_id", vCtx.VerificationId).
			WithContext("verification_type", vCtx.VerificationType)
	}

	// Quick token estimate check
	tokenEstimate := len(renderedTemplate) / 4 // Rough estimate: ~4 chars per token
	if tokenEstimate > p.cfg.TokenBudget.Turn2Prompt {
		p.log.Warn("turn2_prompt_exceeds_token_budget", map[string]interface{}{
			"verification_id":    vCtx.VerificationId,
			"estimated_tokens":  tokenEstimate,
			"token_budget":      p.cfg.TokenBudget.Turn2Prompt,
			"template_type":     templateType,
			"verification_type": vCtx.VerificationType,
		})
	}

	// Create template processor for metrics
	processor := &schema.TemplateProcessor{
		TemplateType:     templateType,
		TemplateVersion:  p.cfg.TemplateVersioning.Version,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		TokenEstimate:    tokenEstimate,
		CharCount:        len(renderedTemplate),
		VerificationId:   vCtx.VerificationId,
		VerificationType: vCtx.VerificationType,
		Turn:             2,
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	p.log.Info("turn2_prompt_generated_successfully", map[string]interface{}{
		"verification_id":    vCtx.VerificationId,
		"verification_type":  vCtx.VerificationType,
		"template_type":      templateType,
		"template_version":   p.cfg.TemplateVersioning.Version,
		"char_count":         len(renderedTemplate),
		"estimated_tokens":   tokenEstimate,
		"processing_time_ms": processor.ProcessingTimeMs,
	})

	return renderedTemplate, processor, nil
}