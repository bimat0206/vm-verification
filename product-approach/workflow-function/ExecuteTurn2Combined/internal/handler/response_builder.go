package handler

import (
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
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

// BuildCombinedTurn2Response builds the combined turn response with all necessary data for Turn2
func (r *ResponseBuilder) BuildCombinedTurn2Response(
	req *models.Turn2Request,
	renderedPrompt string,
	promptRef, rawRef, procRef models.S3Reference,
	invoke *schema.BedrockResponse,
	stages []schema.ProcessingStage,
	totalDurationMs int64,
	bedrockLatencyMs int64,
	dynamoOK bool,
) *schema.CombinedTurnResponse {

	// Build base turn response for Turn2
	turnResponse := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Prompt:    renderedPrompt,
		ImageUrls: map[string]string{
			"checking": req.S3Refs.Images.CheckingBase64.Key,
		},
		Response: schema.BedrockApiResponse{
			Content:   "", // Will be populated from invoke response
			RequestId: "", // RequestID not available in BedrockResponse
		},
		LatencyMs: totalDurationMs,
		TokenUsage: &schema.TokenUsage{
			InputTokens:  invoke.InputTokens,
			OutputTokens: invoke.OutputTokens,
			TotalTokens:  invoke.InputTokens + invoke.OutputTokens,
		},
		Stage: "COMPARISON_ANALYSIS",
		Metadata: map[string]interface{}{
			"model_id":        r.cfg.AWS.BedrockModel,
			"verification_id": req.VerificationID,
			"function_name":   "ExecuteTurn2Combined",
		},
	}

	// Determine template used based on verification type for Turn2
	templateUsed := "turn2-layout-vs-checking"
	if req.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		templateUsed = "turn2-previous-vs-current"
	}

	// Build S3 reference tree for Turn2
	s3RefTree := buildTurn2S3RefTree(req.S3Refs, promptRef, rawRef, procRef)

	// Build context enrichment with schema version and other required fields
	contextEnrichment := map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationContext.VerificationType,
		"s3_references":      s3RefTree,
		"status":             schema.StatusTurn2Completed,
		"summary":            buildTurn2Summary(totalDurationMs, invoke, req.VerificationContext.VerificationType, bedrockLatencyMs, dynamoOK),
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

// Legacy BuildStepFunctionResponse - not used in Turn2 processing
// This method is kept for compatibility but should not be used for Turn2
func (r *ResponseBuilder) BuildStepFunctionResponse(
	req interface{}, // Changed to interface{} to avoid Turn1Request dependency
	promptRef, rawRef, procRef models.S3Reference,
	invoke interface{}, // Changed to interface{} to avoid BedrockResponse dependency
	totalDurationMs int64,
	bedrockLatencyMs int64,
	dynamoOK bool,
) *models.StepFunctionResponse {
	// This method is deprecated for Turn2 processing
	// Use BuildTurn2StepFunctionResponse instead
	return &models.StepFunctionResponse{
		VerificationID: "legacy-method-not-supported",
		S3References:   map[string]interface{}{},
		Status:         "ERROR",
		Summary: map[string]interface{}{
			"error": "Legacy BuildStepFunctionResponse not supported for Turn2",
		},
	}
}

// BuildTurn2StepFunctionResponse builds a Step Function compatible response for Turn2
func (r *ResponseBuilder) BuildTurn2StepFunctionResponse(
	req *models.Turn2Request,
	turn2Resp *models.Turn2Response,
) *models.StepFunctionResponse {
	s3References := map[string]interface{}{
		"prompts_system":  req.S3Refs.Prompts.System,
		"images_checking": req.S3Refs.Images.CheckingBase64,
		"responses": map[string]interface{}{
			"turn2Raw":       turn2Resp.S3Refs.RawResponse,
			"turn2Processed": turn2Resp.S3Refs.ProcessedResponse,
			"turn1Raw":       req.S3Refs.Turn1.RawResponse,
			"turn1Processed": req.S3Refs.Turn1.ProcessedResponse,
		},
	}

	if req.S3Refs.Processing.LayoutMetadata.Key != "" {
		s3References["processing_layout-metadata"] = req.S3Refs.Processing.LayoutMetadata
	}
	if req.S3Refs.Processing.HistoricalContext.Key != "" {
		s3References["processing_historical-context"] = req.S3Refs.Processing.HistoricalContext
	}

	summaryMap := map[string]interface{}{
		"analysisStage":       turn2Resp.Summary.AnalysisStage,
		"processingTimeMs":    turn2Resp.Summary.ProcessingTimeMs,
		"verificationOutcome": turn2Resp.VerificationOutcome,
		"tokenUsage": map[string]interface{}{
			"input":  turn2Resp.Summary.TokenUsage.InputTokens,
			"output": turn2Resp.Summary.TokenUsage.OutputTokens,
			"total":  turn2Resp.Summary.TokenUsage.TotalTokens,
		},
		"bedrockRequestId": turn2Resp.Summary.BedrockRequestID,
	}
	if turn2Resp.Summary.DiscrepanciesFound != nil {
		summaryMap["discrepanciesFound"] = *turn2Resp.Summary.DiscrepanciesFound
	}
	if turn2Resp.Summary.DynamodbUpdated != nil {
		summaryMap["dynamodbUpdated"] = *turn2Resp.Summary.DynamodbUpdated
	}

	return &models.StepFunctionResponse{
		VerificationID: req.VerificationID,
		S3References:   s3References,
		Status:         string(turn2Resp.Status),
		Summary:        summaryMap,
	}
}
