package handler

import (
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/schema"
)

// ResponseBuilder handles building combined turn responses
type ResponseBuilder struct {
	cfg config.Config
}

// NewResponseBuilder creates a new instance of ResponseBuilder
func NewResponseBuilder(cfg config.Config) *ResponseBuilder {
	return &ResponseBuilder{
		cfg: cfg,
	}
}

// BuildCombinedTurnResponse builds the combined turn response with all necessary data
func (r *ResponseBuilder) BuildCombinedTurnResponse(
	req *models.Turn1Request,
	resp *models.BedrockResponse,
	rawRef, procRef models.S3Reference,
	templateProcessor *schema.TemplateProcessor,
	processingMetrics *schema.ProcessingMetrics,
	totalDuration time.Duration,
	processingStages []schema.ProcessingStage,
	statusHistory []schema.StatusHistoryEntry,
) *schema.CombinedTurnResponse {
	
	// Build base turn response
	turnResponse := &schema.TurnResponse{
		TurnId:    1,
		Timestamp: schema.FormatISO8601(),
		Prompt:    "", // Would be filled with actual prompt
		ImageUrls: map[string]string{
			"reference": req.S3Refs.Images.ReferenceBase64.Key,
		},
		Response: schema.BedrockApiResponse{
			Content:    string(resp.Raw), // Simplified - would parse properly
			RequestId:  resp.RequestID,
		},
		LatencyMs:  totalDuration.Milliseconds(),
		TokenUsage: &resp.TokenUsage,
		Stage:      "REFERENCE_ANALYSIS",
		Metadata: map[string]interface{}{
			"model_id":         r.cfg.AWS.BedrockModel,
			"verification_id":  req.VerificationID,
			"function_name":    "ExecuteTurn1Combined",
		},
	}
	
	// Build context enrichment
	contextEnrichment := map[string]interface{}{
		"verification_type":    req.VerificationContext.VerificationType,
		"layout_integrated":    req.VerificationContext.LayoutId != 0,
		"historical_context":   req.VerificationContext.HistoricalContext != nil,
		"processing_stages":    len(processingStages),
		"status_updates":       len(statusHistory),
		"concurrent_loading":   true,
		"enhanced_tracking":    true,
	}
	
	templateUsed := "turn1-layout-vs-checking"
	if req.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		templateUsed = "turn1-previous-vs-current"
	}
	
	// Build combined response
	combinedResponse := &schema.CombinedTurnResponse{
		TurnResponse:      turnResponse,
		ProcessingStages:  processingStages,
		InternalPrompt:    "", // Would be filled with actual internal prompt
		TemplateUsed:      templateUsed,
		ContextEnrichment: contextEnrichment,
	}
	
	return combinedResponse
}