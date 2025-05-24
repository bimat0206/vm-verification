package handler

import (
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
	renderedPrompt string,
	promptRef, rawRef, procRef models.S3Reference,
	invoke *models.BedrockResponse,
	stages []schema.ProcessingStage,
	totalDurationMs int64,
	dynamoOK *bool,
) *schema.CombinedTurnResponse {
	
	// Build base turn response
	turnResponse := &schema.TurnResponse{
		TurnId:    1,
		Timestamp: schema.FormatISO8601(),
		Prompt:    renderedPrompt,
		ImageUrls: map[string]string{
			"reference": req.S3Refs.Images.ReferenceBase64.Key,
		},
		Response: schema.BedrockApiResponse{
			Content:   string(invoke.Raw),
			RequestId: invoke.RequestID,
		},
		LatencyMs:  totalDurationMs,
		TokenUsage: &invoke.TokenUsage,
		Stage:      "REFERENCE_ANALYSIS",
		Metadata: map[string]interface{}{
			"model_id":        r.cfg.AWS.BedrockModel,
			"verification_id": req.VerificationID,
			"function_name":   "ExecuteTurn1Combined",
		},
	}

	// Determine template used based on verification type
	templateUsed := "turn1-layout-vs-checking"
	if req.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		templateUsed = "turn1-previous-vs-current"
	}

	// Build S3 reference tree
	s3RefTree := buildS3RefTree(req.S3Refs, promptRef, rawRef, procRef)

	// Build context enrichment with schema version and other required fields
	contextEnrichment := map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationContext.VerificationType,
		"s3_references":      s3RefTree,
		"status":             schema.StatusTurn1Completed,
		"summary":            buildSummary(totalDurationMs, invoke, req.VerificationContext.VerificationType, dynamoOK),
		"schema_version":     schema.SchemaVersion,
		"layout_integrated":  req.VerificationContext.LayoutId != 0,
		"historical_context": req.VerificationContext.HistoricalContext != nil,
	}

	// Build combined response according to the schema structure
	resp := &schema.CombinedTurnResponse{
		TurnResponse:      turnResponse,
		ProcessingStages:  stages,
		InternalPrompt:    renderedPrompt,
		TemplateUsed:      templateUsed,
		ContextEnrichment: contextEnrichment,
	}

	return resp
}
