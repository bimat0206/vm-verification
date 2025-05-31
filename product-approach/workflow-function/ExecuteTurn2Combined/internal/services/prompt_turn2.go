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
	GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.TurnResponse, turn1RawResponse json.RawMessage, layoutMetadata map[string]interface{}) (string, *schema.TemplateProcessor, error)
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
func (p *promptServiceTurn2) GenerateTurn2PromptWithMetrics(ctx context.Context, vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.TurnResponse, turn1RawResponse json.RawMessage, layoutMetadata map[string]interface{}) (string, *schema.TemplateProcessor, error) {
	start := time.Now()
	if vCtx == nil {
		return "", nil, errors.NewValidationError("verification context required", nil)
	}
	if systemPrompt == "" {
		return "", nil, errors.NewValidationError("system prompt cannot be empty", map[string]interface{}{"verification_id": vCtx.VerificationId})
	}

	templateType := getTurn2TemplateType(vCtx.VerificationType)

	templateData := p.buildTemplateContext(vCtx, systemPrompt, turn1Response, layoutMetadata)

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

	// Create a structured prompt object that matches the Turn1 format
	structuredPrompt := map[string]interface{}{
		"contextualInstructions": map[string]string{
			"analysisObjective": "Compare checking image with reference layout",
		},
		"createdAt": schema.FormatISO8601(),
		"generationMetadata": map[string]interface{}{
			"contextSources":   []string{"TURN1_ANALYSIS", "IMAGE_METADATA", "LAYOUT_METADATA"},
			"processingTimeMs": time.Since(start).Milliseconds(),
			"promptSource":     "TEMPLATE_BASED",
		},
		"imageReference": map[string]interface{}{
			"base64StorageReference": map[string]string{
				"bucket": "kootoro-dev-s3-state-f6d3xl",
				"key":    "checking-base64.base64",
			},
			"imageType":  "checking",
			"sourceUrl": "s3://kootoro-dev-s3-state-f6d3xl/checking-base64.base64",
		},
		"messageStructure": map[string]interface{}{
			"content": []map[string]string{
				{
					"text": renderedPrompt,
					"type": "text",
				},
			},
			"role": "user",
		},
		"promptType":       "TURN2",
		"templateVersion":  p.cfg.Prompts.Turn2TemplateVersion,
		"verificationId":   vCtx.VerificationId,
		"verificationType": vCtx.VerificationType,
	}

	// Convert the structured prompt to JSON
	structuredPromptJSON, err := json.Marshal(structuredPrompt)
	if err != nil {
		p.log.Error("turn2_structured_prompt_marshal_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": vCtx.VerificationId,
		})
		return "", nil, errors.WrapError(err, errors.ErrorTypeValidation, "failed to marshal structured Turn2 prompt", false)
	}

	processor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      templateType,
			TemplateVersion: p.cfg.Prompts.Turn2TemplateVersion,
			TemplateType:    templateType,
			Content:         renderedPrompt,
		},
		ContextData:     templateData,
		ProcessedPrompt: string(structuredPromptJSON),
		ProcessingTime:  time.Since(start).Milliseconds(),
		InputTokens:     0,
		OutputTokens:    0,
	}

	p.log.Info("turn2_prompt_generated", map[string]interface{}{
		"verification_id":  vCtx.VerificationId,
		"template_type":    templateType,
		"template_version": p.cfg.Prompts.Turn2TemplateVersion,
		"prompt_length":    len(renderedPrompt),
		"structured":       true,
	})

	return string(structuredPromptJSON), processor, nil
}

// buildTemplateContext creates the context data for template processing
func (p *promptServiceTurn2) buildTemplateContext(vCtx *schema.VerificationContext, systemPrompt string, turn1Response *schema.TurnResponse, layoutMetadata map[string]interface{}) map[string]interface{} {
	context := map[string]interface{}{
		"VerificationContext": vCtx,
		"SystemPrompt":        systemPrompt,
		"Turn1Response":       turn1Response,
		"VendingMachineId":    vCtx.VendingMachineId,
		"VendingMachineID":    vCtx.VendingMachineId, // Also add with uppercase ID for template compatibility
		"TemplateVersion":     p.cfg.Prompts.Turn2TemplateVersion,
		"CreatedAt":           schema.FormatISO8601(),
	}

	// Extract layout dimensions from layoutMetadata when available
	rowCount := -1
	colCount := -1
	
	if layoutMetadata != nil {
		if rc, ok := layoutMetadata["RowCount"]; ok {
			switch v := rc.(type) {
			case int:
				rowCount = v
			case float64:
				rowCount = int(v)
			case int64:
				rowCount = int(v)
			}
		}
		if cc, ok := layoutMetadata["ColumnCount"]; ok {
			switch v := cc.(type) {
			case int:
				colCount = v
			case float64:
				colCount = int(v)
			case int64:
				colCount = int(v)
			}
		}
	}
	
	if rowCount == -1 || colCount == -1 {
		p.log.Warn("layout_dimensions_missing_in_turn2", map[string]interface{}{
			"rowCount":        rowCount,
			"columnCount":     colCount,
			"verification_id": vCtx.VerificationId,
		})
	}
	
	context["RowCount"] = rowCount
	context["ColumnCount"] = colCount

	// Row labels only generated when row count is valid
	if rowCount > 0 {
		context["RowLabels"] = p.ensureRowLabels(layoutMetadata, rowCount)
	} else {
		context["RowLabels"] = []string{}
	}

	// Add layout-specific context
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking {
		context["LayoutId"] = vCtx.LayoutId
		context["LayoutPrefix"] = vCtx.LayoutPrefix
		context["LayoutMetadata"] = layoutMetadata

		// Add Location if available in layout metadata
		if layoutMetadata != nil {
			if location, ok := layoutMetadata["Location"].(string); ok {
				context["Location"] = location
			}
		}
	}

	return context
}

// ensureRowLabels generates row labels for the template
func (p *promptServiceTurn2) ensureRowLabels(layoutMetadata map[string]interface{}, rowCount int) []string {
	if rowCount <= 0 {
		return []string{}
	}
	
	// Check if RowLabels already exists in layout metadata
	if layoutMetadata != nil {
		if labels, ok := layoutMetadata["RowLabels"].([]string); ok && len(labels) >= rowCount {
			return labels
		}
		// Handle []interface{} case
		if labelsInterface, ok := layoutMetadata["RowLabels"].([]interface{}); ok {
			labels := make([]string, 0, len(labelsInterface))
			for _, l := range labelsInterface {
				if str, ok := l.(string); ok {
					labels = append(labels, str)
				}
			}
			if len(labels) >= rowCount {
				return labels
			}
		}
	}

	// Generate default row labels A, B, C, etc.
	labels := make([]string, rowCount)
	for i := 0; i < rowCount; i++ {
		labels[i] = string(rune('A' + i))
	}

	return labels
}
