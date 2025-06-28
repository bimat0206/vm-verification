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
	promptRef, rawRef, procRef, convRef models.S3Reference,
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
			InputTokens:    invoke.InputTokens,
			OutputTokens:   invoke.OutputTokens,
			ThinkingTokens: invoke.ThinkingTokens,
			TotalTokens:    invoke.TotalTokens,
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
	s3RefTree := buildTurn2S3RefTree(req.S3Refs, promptRef, rawRef, procRef, convRef)

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
	promptRef models.S3Reference,
	convRef models.S3Reference,
) *models.StepFunctionResponse {
	tree := buildTurn2S3RefTree(req.S3Refs, promptRef, turn2Resp.S3Refs.RawResponse, turn2Resp.S3Refs.ProcessedResponse, convRef)

	s3References := map[string]interface{}{}
	// carry over all input references first
	for k, v := range req.InputS3References {
		s3References[k] = v
	}

	// overwrite / add updated turn2 references
	s3References["prompts_system"] = tree.Prompts.SystemPrompt
	s3References["processing_initialization"] = tree.Initialization
	s3References["images_metadata"] = tree.Images.Metadata
	s3References["responses"] = map[string]interface{}{
		"turn2Raw":       tree.Responses.Turn2Raw,
		"turn2Processed": tree.Responses.Turn2Processed,
		"turn1Raw":       tree.Responses.Turn1Raw,
		"turn1Processed": tree.Responses.Turn1Processed,
	}

	if tree.Prompts.Turn1Prompt.Key != "" {
		s3References["prompts_turn1"] = tree.Prompts.Turn1Prompt
	}
	if tree.Prompts.Turn2Prompt.Key != "" {
		s3References["prompts_turn2"] = tree.Prompts.Turn2Prompt
	}

	convMap := map[string]interface{}{}
	if tree.Conversation.Turn1.Key != "" {
		convMap["turn1"] = tree.Conversation.Turn1
	}
	if tree.Conversation.Turn2.Key != "" {
		convMap["turn2"] = tree.Conversation.Turn2
	}
	if len(convMap) > 0 {
		s3References["conversation"] = convMap
	}

	if tree.Processing.LayoutMetadata.Key != "" {
		s3References["processing_layout-metadata"] = tree.Processing.LayoutMetadata
	}
	if tree.Processing.HistoricalContext.Key != "" {
		s3References["processing_historical_context"] = tree.Processing.HistoricalContext
	}

	summaryMap := map[string]interface{}{
		"analysisStage":       turn2Resp.Summary.AnalysisStage,
		"processingTimeMs":    turn2Resp.Summary.ProcessingTimeMs,
		"verificationOutcome": turn2Resp.VerificationOutcome,
		"tokenUsage": map[string]interface{}{
			"input":    turn2Resp.Summary.TokenUsage.InputTokens,
			"output":   turn2Resp.Summary.TokenUsage.OutputTokens,
			"thinking": turn2Resp.Summary.TokenUsage.ThinkingTokens,
			"total":    turn2Resp.Summary.TokenUsage.TotalTokens,
		},
		"bedrockRequestId": turn2Resp.Summary.BedrockRequestID,
	}
	summaryMap["discrepanciesFound"] = turn2Resp.Summary.DiscrepanciesFound
	summaryMap["dynamodbUpdated"] = turn2Resp.Summary.DynamodbUpdated
	summaryMap["comparisonCompleted"] = turn2Resp.Summary.ComparisonCompleted
	summaryMap["conversationCompleted"] = turn2Resp.Summary.ConversationCompleted
	if turn2Resp.Summary.VerificationType != "" {
		summaryMap["verificationType"] = turn2Resp.Summary.VerificationType
	}
	if turn2Resp.Summary.BedrockLatencyMs > 0 {
		summaryMap["bedrockLatencyMs"] = turn2Resp.Summary.BedrockLatencyMs
	}
	if turn2Resp.Summary.S3StorageCompleted {
		summaryMap["s3StorageCompleted"] = turn2Resp.Summary.S3StorageCompleted
	}

	return &models.StepFunctionResponse{
		VerificationID: req.VerificationID,
		S3References:   s3References,
		Status:         string(turn2Resp.Status),
		Summary:        summaryMap,
	}
}
