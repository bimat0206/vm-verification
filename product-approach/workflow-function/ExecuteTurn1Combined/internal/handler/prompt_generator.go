package handler

import (
	"context"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/schema"
)

// PromptGenerator handles Turn1 prompt generation with enhanced template processing
type PromptGenerator struct {
	promptService services.PromptService
	cfg           config.Config
}

// NewPromptGenerator creates a new instance of PromptGenerator
func NewPromptGenerator(promptService services.PromptService, cfg config.Config) *PromptGenerator {
	return &PromptGenerator{
		promptService: promptService,
		cfg:           cfg,
	}
}

// GetTemplateUsed returns the template name based on verification type
func (p *PromptGenerator) GetTemplateUsed(vCtx models.VerificationContext) string {
	if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent {
		return "turn1-previous-vs-current"
	}
	return "turn1-layout-vs-checking"
}

// GenerateTurn1PromptEnhanced generates Turn1 prompt with enhanced template processing
func (p *PromptGenerator) GenerateTurn1PromptEnhanced(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, *schema.TemplateProcessor, error) {
	startTime := time.Now()

	// Use the existing prompt service while capturing processing info
	prompt, err := p.promptService.GenerateTurn1Prompt(ctx, vCtx, systemPrompt)
	if err != nil {
		return "", nil, err
	}

	processingDuration := time.Since(startTime)

	// Create template processor info for tracking
	templateProcessor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      p.GetTemplateUsed(vCtx),
			TemplateVersion: p.cfg.Prompts.TemplateVersion,
			TemplateType:    p.GetTemplateUsed(vCtx),
			Content:         prompt,
		},
		ContextData: map[string]interface{}{
			"verificationType":  vCtx.VerificationType,
			"layoutMetadata":    vCtx.LayoutMetadata,
			"historicalContext": vCtx.HistoricalContext,
			"systemPromptSize":  len(systemPrompt),
		},
		ProcessedPrompt: prompt,
		ProcessingTime:  processingDuration.Milliseconds(),
		InputTokens:     0,
		OutputTokens:    0,
	}

	return prompt, templateProcessor, nil
}
