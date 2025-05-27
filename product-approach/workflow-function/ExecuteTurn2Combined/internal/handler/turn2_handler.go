package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Turn2Handler handles the complete Turn2 processing flow
type Turn2Handler struct {
	contextLoader  *ContextLoader
	promptService  services.PromptServiceTurn2
	bedrockService services.BedrockServiceTurn2
	s3Service      services.S3StateManager
	dynamoService  services.DynamoDBService
	log            logger.Logger
	cfg            config.Config
}

// NewTurn2Handler creates a new Turn2Handler instance
func NewTurn2Handler(
	contextLoader *ContextLoader,
	promptService services.PromptServiceTurn2,
	bedrockService services.BedrockServiceTurn2,
	s3Service services.S3StateManager,
	dynamoService services.DynamoDBService,
	log logger.Logger,
	cfg config.Config,
) *Turn2Handler {
	return &Turn2Handler{
		contextLoader:  contextLoader,
		promptService:  promptService,
		bedrockService: bedrockService,
		s3Service:      s3Service,
		dynamoService:  dynamoService,
		log:            log,
		cfg:            cfg,
	}
}

// ProcessTurn2Request handles the complete Turn2 processing flow
func (h *Turn2Handler) ProcessTurn2Request(ctx context.Context, req *models.Turn2Request) (*models.Turn2Response, error) {
	startTime := time.Now()
	h.log.Info("turn2_processing_started", map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationType,
		"checking_image_key": req.S3Refs.Images.CheckingBase64.Key,
		"turn1_response_key": req.S3Refs.Turn1References.ProcessedResponse.Key,
	})

	// Load context (system prompt, checking image, Turn1 results)
	loadResult := h.contextLoader.LoadContextTurn2(ctx, req)
	if loadResult.Error != nil {
		wfErr := errors.WrapError(loadResult.Error, errors.ErrorTypeS3,
			"failed to load Turn2 context", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "context_loading")
		h.log.Error("context_loading_failed", map[string]interface{}{
			"error_type": string(wfErr.Type),
			"error_code": wfErr.Code,
			"message":    wfErr.Message,
			"retryable":  wfErr.Retryable,
			"severity":   string(wfErr.Severity),
		})
		h.persistErrorState(ctx, req, wfErr, "context_loading", startTime)
		return nil, wfErr
	}

	// Create verification context
	vCtx := &schema.VerificationContext{
		VerificationId:   req.VerificationID,
		VerificationType: req.VerificationType,
		LayoutId:         req.LayoutID,
		VendingMachineId: req.VendingMachineID,
		Location:         req.Location,
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	// Generate Turn2 prompt
	prompt, promptProcessor, err := h.promptService.GenerateTurn2PromptWithMetrics(
		ctx,
		vCtx,
		loadResult.SystemPrompt,
		loadResult.Turn1Response,
		loadResult.Turn1RawResponse,
	)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeTemplate,
			"failed to generate Turn2 prompt", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "prompt_generation")
		h.log.Error("prompt_generation_failed", map[string]interface{}{
			"error_type": string(wfErr.Type),
			"error_code": wfErr.Code,
			"message":    wfErr.Message,
			"retryable":  wfErr.Retryable,
			"severity":   string(wfErr.Severity),
		})
		h.persistErrorState(ctx, req, wfErr, "prompt_generation", startTime)
		return nil, wfErr
	}

	// Store prompt processor metrics
	promptProcessorRef, err := h.s3Service.StoreTemplateProcessor(ctx, req.VerificationID, promptProcessor)
	if err != nil {
		h.log.Warn("failed_to_store_prompt_processor", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Invoke Bedrock with conversation history
	bedrockResponse, err := h.bedrockService.ConverseWithHistory(
		ctx,
		loadResult.SystemPrompt,
		prompt,
		loadResult.Base64Image,
		loadResult.Turn1Response,
	)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to invoke Bedrock for Turn2", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "bedrock_invocation")
		h.log.Error("bedrock_invocation_failed", map[string]interface{}{
			"error_type": string(wfErr.Type),
			"error_code": wfErr.Code,
			"message":    wfErr.Message,
			"retryable":  wfErr.Retryable,
			"severity":   string(wfErr.Severity),
		})
		h.persistErrorState(ctx, req, wfErr, "bedrock_invocation", startTime)
		return nil, wfErr
	}

	// Store raw Bedrock response
	rawResponseRef, err := h.s3Service.StoreRawResponse(ctx, req.VerificationID, bedrockResponse)
	if err != nil {
		h.log.Warn("failed_to_store_raw_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Parse Bedrock response
	markdownResponse, err := bedrockparser.ParseBedrockResponseAsMarkdown(bedrockResponse.Content)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeParser,
			"failed to parse Bedrock response as markdown", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "response_parsing")
		h.log.Error("markdown_parsing_failed", map[string]interface{}{
			"error_type": string(wfErr.Type),
			"error_code": wfErr.Code,
			"message":    wfErr.Message,
			"retryable":  wfErr.Retryable,
			"severity":   string(wfErr.Severity),
		})
		h.persistErrorState(ctx, req, wfErr, "response_parsing", startTime)
		return nil, wfErr
	}

	// Store markdown response
	markdownRef, err := h.s3Service.StoreTurn2Markdown(ctx, req.VerificationID, markdownResponse)
	if err != nil {
		h.log.Warn("failed_to_store_markdown_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Parse structured data from response
	parsedData, err := bedrockparser.ParseTurn2Response(bedrockResponse.Content)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeParser,
			"failed to parse Turn2 response", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "response_parsing")
		h.log.Error("turn2_parsing_failed", map[string]interface{}{
			"error_type": string(wfErr.Type),
			"error_code": wfErr.Code,
			"message":    wfErr.Message,
			"retryable":  wfErr.Retryable,
			"severity":   string(wfErr.Severity),
		})
		h.persistErrorState(ctx, req, wfErr, "response_parsing", startTime)
		return nil, wfErr
	}

	// Interpret discrepancies with business rules
	finalStatus, refinedSummary, err := h.interpretDiscrepancies(parsedData, &req.VerificationContext)
	if err != nil {
		h.log.Warn("discrepancy_interpretation_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Fall back to Bedrock outcome if interpretation fails
		finalStatus = parsedData.VerificationOutcome
		refinedSummary = parsedData.ComparisonSummary
	}

	// Store processed Turn2 response
	processedRef, err := h.s3Service.StoreTurn2Response(ctx, req.VerificationID, parsedData)
	if err != nil {
		h.log.Warn("failed_to_store_processed_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Create processing metrics
	processingMetrics := &schema.ProcessingMetrics{
		VerificationId:      req.VerificationID,
		VerificationType:    req.VerificationType,
		Turn:                2,
		TotalTimeMs:         time.Since(startTime).Milliseconds(),
		ContextLoadingMs:    loadResult.Duration.Milliseconds(),
		PromptGenerationMs:  promptProcessor.ProcessingTimeMs,
		BedrockInvocationMs: bedrockResponse.LatencyMs,
		ResponseParsingMs:   0, // Not tracked separately
		TokenUsage: schema.TokenUsage{
			InputTokens:  bedrockResponse.InputTokens,
			OutputTokens: bedrockResponse.OutputTokens,
			TotalTokens:  bedrockResponse.InputTokens + bedrockResponse.OutputTokens,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Store processing metrics
	metricsRef, err := h.s3Service.StoreProcessingMetrics(ctx, req.VerificationID, processingMetrics)
	if err != nil {
		h.log.Warn("failed_to_store_processing_metrics", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Create Turn2 response
	response := &models.Turn2Response{
		VerificationID:      req.VerificationID,
		VerificationType:    req.VerificationType,
		VerificationOutcome: finalStatus,
		Discrepancies:       parsedData.Discrepancies,
		ComparisonSummary:   refinedSummary,
		S3Refs: models.Turn2ResponseS3Refs{
			RawResponse:       rawResponseRef,
			MarkdownResponse:  markdownRef,
			ProcessedResponse: processedRef,
			PromptProcessor:   promptProcessorRef,
			ProcessingMetrics: metricsRef,
		},
		ProcessingMetrics: processingMetrics,
	}

	h.log.Info("turn2_processing_completed", map[string]interface{}{
		"verification_id":       req.VerificationID,
		"verification_type":     req.VerificationType,
		"verification_outcome":  finalStatus,
		"discrepancy_count":     len(parsedData.Discrepancies),
		"total_processing_time": processingMetrics.TotalTimeMs,
		"input_tokens":          processingMetrics.TokenUsage.InputTokens,
		"output_tokens":         processingMetrics.TokenUsage.OutputTokens,
		"total_tokens":          processingMetrics.TokenUsage.TotalTokens,
	})

	// Build status entry and turn metrics for DynamoDB update
	statusEntry := schema.StatusHistoryEntry{
		Status:           schema.StatusTurn2Completed,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn2Combined",
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		Stage:            "turn2_completion",
		Metrics: map[string]interface{}{
			"bedrock_request_id":   bedrockResponse.RequestID,
			"discrepancy_count":    len(parsedData.Discrepancies),
			"verification_outcome": finalStatus,
		},
	}

	turn2Metrics := &schema.TurnMetrics{
		StartTime:        startTime.Format(time.RFC3339),
		EndTime:          time.Now().Format(time.RFC3339),
		TotalTimeMs:      processingMetrics.TotalTimeMs,
		BedrockLatencyMs: bedrockResponse.LatencyMs,
		ProcessingTimeMs: processingMetrics.TotalTimeMs - bedrockResponse.LatencyMs,
		RetryAttempts:    0,
		TokenUsage:       &processingMetrics.TokenUsage,
	}

	// Update VerificationResults with Turn2 details
	err = h.dynamoService.UpdateTurn2CompletionDetails(
		ctx,
		req.VerificationID,
		req.VerificationContext.VerificationAt,
		statusEntry,
		turn2Metrics,
		finalStatus,
		parsedData.Discrepancies,
		refinedSummary,
	)
	if err != nil {
		h.log.Warn("dynamodb_update_turn2_failed", map[string]interface{}{
			"error":           err.Error(),
			"retryable":       errors.IsRetryable(err),
			"verification_id": req.VerificationID,
		})
	}

	// Append conversation history for Turn2
	turnEntry := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Prompt:    prompt,
		ImageUrls: map[string]string{
			"checking": req.S3Refs.Images.CheckingBase64.Key,
		},
		Response: schema.BedrockApiResponse{
			Content:   bedrockResponse.Content,
			ModelId:   bedrockResponse.ModelId,
			RequestId: bedrockResponse.RequestID,
		},
		LatencyMs: bedrockResponse.LatencyMs,
		TokenUsage: &schema.TokenUsage{
			InputTokens:  bedrockResponse.InputTokens,
			OutputTokens: bedrockResponse.OutputTokens,
			TotalTokens:  bedrockResponse.InputTokens + bedrockResponse.OutputTokens,
		},
		Stage: "CHECKING_ANALYSIS",
	}

	if err := h.dynamoService.UpdateConversationTurn(ctx, req.VerificationID, turnEntry); err != nil {
		h.log.Warn("conversation_history_update_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	return response, nil
}

// interpretDiscrepancies applies business rules to refine verification outcome
func (h *Turn2Handler) interpretDiscrepancies(parsedData *bedrockparser.ParsedTurn2Data, vCtx *models.VerificationContext) (string, string, error) {
	if parsedData == nil {
		return schema.VerificationStatusFailed, "", fmt.Errorf("parsed data is nil")
	}

	finalStatus := parsedData.VerificationOutcome
	summary := parsedData.ComparisonSummary

	highSeverity := false
	mismatchCount := 0

	for _, d := range parsedData.Discrepancies {
		if strings.EqualFold(d.Severity, "HIGH") {
			highSeverity = true
		}
		if d.Type == "MISSING" || d.Type == "MISPLACED" {
			mismatchCount++
		}
	}

	if highSeverity {
		finalStatus = schema.VerificationStatusIncorrect
	}

	if h.cfg.Processing.DiscrepancyThreshold > 0 && mismatchCount >= h.cfg.Processing.DiscrepancyThreshold {
		finalStatus = schema.VerificationStatusIncorrect
	}

	if finalStatus != parsedData.VerificationOutcome {
		note := fmt.Sprintf("Assessment: %s due to %d discrepancies.", finalStatus, mismatchCount)
		if summary != "" {
			summary = summary + " " + note
		} else {
			summary = note
		}
	}

	return finalStatus, summary, nil
}

// persistErrorState attempts to record error information in DynamoDB before returning the error.
func (h *Turn2Handler) persistErrorState(ctx context.Context, req *models.Turn2Request, wfErr *errors.WorkflowError, stage string, startTime time.Time) {
	statusEntry := schema.StatusHistoryEntry{
		Status:           schema.StatusTurn2Error,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn2Combined",
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		Stage:            stage,
	}

	errorInfo := schema.ErrorInfo{
		Code:      wfErr.Code,
		Message:   wfErr.Message,
		Details:   wfErr.Details,
		Timestamp: schema.FormatISO8601(),
	}

	tracking := &schema.ErrorTracking{
		HasErrors:    true,
		CurrentError: &errorInfo,
		ErrorHistory: []schema.ErrorInfo{errorInfo},
		LastErrorAt:  schema.FormatISO8601(),
	}

	if err := h.dynamoService.UpdateVerificationStatusEnhanced(ctx, req.VerificationID, req.VerificationContext.VerificationAt, statusEntry); err != nil {
		h.log.Warn("dynamodb_status_error_update_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	if err := h.dynamoService.UpdateErrorTracking(ctx, req.VerificationID, tracking); err != nil {
		h.log.Warn("dynamodb_error_tracking_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	turnEntry := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Stage:     stage,
		Metadata: map[string]interface{}{
			"error": wfErr.Message,
		},
	}
	if err := h.dynamoService.UpdateConversationTurn(ctx, req.VerificationID, turnEntry); err != nil {
		h.log.Warn("conversation_history_error_update_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}
}
