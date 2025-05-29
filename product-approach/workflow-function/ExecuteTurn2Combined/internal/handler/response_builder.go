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

// BuildCombinedTurnResponse builds the combined turn response with all necessary data
func (r *ResponseBuilder) BuildCombinedTurnResponse(
	req *models.Turn1Request,
	renderedPrompt string,
	promptRef, rawRef, procRef models.S3Reference,
	invoke *models.BedrockResponse,
	stages []schema.ProcessingStage,
	totalDurationMs int64,
	bedrockLatencyMs int64,
	dynamoOK bool,
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
		"summary":            buildSummary(totalDurationMs, invoke, req.VerificationContext.VerificationType, bedrockLatencyMs, dynamoOK),
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

// BuildStepFunctionResponse builds a response formatted for Step Functions
func (r *ResponseBuilder) BuildStepFunctionResponse(
	req *models.Turn1Request,
	promptRef, rawRef, procRef models.S3Reference,
	invoke *models.BedrockResponse,
	totalDurationMs int64,
	bedrockLatencyMs int64,
	dynamoOK bool,
) *models.StepFunctionResponse {
	// Build S3 reference tree in the expected format
	s3RefTree := buildS3RefTree(req.S3Refs, promptRef, rawRef, procRef)

	// Convert S3RefTree to map[string]interface{} for Step Functions
	s3References := map[string]interface{}{
		"processing_initialization": s3RefTree.Initialization,
		"images_metadata":           s3RefTree.Images.Metadata,
		"prompts_system":            s3RefTree.Prompts.SystemPrompt,
		"responses": map[string]interface{}{
			"turn1Raw":       rawRef,
			"turn1Processed": procRef,
		},
	}

	if s3RefTree.Processing.LayoutMetadata.Key != "" {
		s3References["processing_layout-metadata"] = s3RefTree.Processing.LayoutMetadata
	}

	if req.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		if s3RefTree.Processing.HistoricalContext.Key != "" {
			s3References["processing_historical-context"] = s3RefTree.Processing.HistoricalContext
		}
	}

	// Build summary in the expected format
	summary := buildSummary(totalDurationMs, invoke, req.VerificationContext.VerificationType, bedrockLatencyMs, dynamoOK)

	// Convert ExecutionSummary to map[string]interface{}
	summaryMap := map[string]interface{}{
		"analysisStage":    summary.AnalysisStage,
		"verificationType": summary.VerificationType,
		"processingTimeMs": summary.ProcessingTimeMs,
		"tokenUsage": map[string]interface{}{
			"input":    summary.TokenUsage.Input,
			"output":   summary.TokenUsage.Output,
			"thinking": summary.TokenUsage.Thinking,
			"total":    summary.TokenUsage.Total,
		},
		"bedrockLatencyMs":    summary.BedrockLatencyMs,
		"bedrockRequestId":    summary.BedrockRequestId,
		"dynamodbUpdated":     summary.DynamodbUpdated,
		"conversationTracked": summary.ConversationTracked,
		"s3StorageCompleted":  summary.S3StorageCompleted,
	}

	return &models.StepFunctionResponse{
		VerificationID: req.VerificationID,
		S3References:   s3References,
		Status:         schema.StatusTurn1Completed,
		Summary:        summaryMap,
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
