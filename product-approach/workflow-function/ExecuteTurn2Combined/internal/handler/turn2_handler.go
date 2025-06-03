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
	bedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Turn2Handler handles the complete Turn2 processing flow
type Turn2Handler struct {
	contextLoader *ContextLoader
	promptService services.PromptServiceTurn2
	bedrock       services.BedrockServiceTurn2
	s3            services.S3StateManager
	dynamo        services.DynamoDBService
	dynamoManager *DynamoManager
	log           logger.Logger
	cfg           config.Config

	// Additional components required by helper utilities
	validator         *Validator
	statusTracker     *StatusTracker
	processingTracker *ProcessingStagesTracker
	bedrockInvoker    *BedrockInvoker
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
		bedrock:        bedrockService,
		s3:             s3Service,
		dynamo:         dynamoService,
		dynamoManager:  NewDynamoManager(dynamoService, cfg, log),
		log:            log,
		cfg:            cfg,
		validator:      NewValidator(),
		bedrockInvoker: NewBedrockInvoker(bedrockService, cfg, log),
	}
}

// ProcessTurn2Request handles the complete Turn2 processing flow
func (h *Turn2Handler) ProcessTurn2Request(ctx context.Context, req *models.Turn2Request) (*models.Turn2Response, models.S3Reference, models.S3Reference, error) {
	startTime := time.Now()
	h.log.Info("turn2_processing_started", map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationContext.VerificationType,
		"checking_image_key": req.S3Refs.Images.CheckingBase64.Key,
		"turn1_response_key": req.S3Refs.Turn1.ProcessedResponse.Key,
	})

	var convRef models.S3Reference
	var promptRef models.S3Reference

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
		return nil, models.S3Reference{}, models.S3Reference{}, wfErr
	}

	// Load Turn1 raw response for Bedrock history
	var loadedTurn1Response *schema.TurnResponse
	if req.S3Refs.Turn1.RawResponse.Key != "" {
		var err error
		loadedTurn1Response, err = h.s3.LoadTurn1SchemaResponse(ctx, req.S3Refs.Turn1.RawResponse)
		if err != nil {
			wfErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load Turn1 raw response", true).
				WithContext("verification_id", req.VerificationID).
				WithContext("stage", "context_loading")
			h.log.Error("turn1_raw_load_failed", map[string]interface{}{
				"error_type": string(wfErr.Type),
				"error_code": wfErr.Code,
				"message":    wfErr.Message,
				"retryable":  wfErr.Retryable,
				"severity":   string(wfErr.Severity),
			})
			h.persistErrorState(ctx, req, wfErr, "context_loading", startTime)
			return nil, models.S3Reference{}, models.S3Reference{}, wfErr
		}
	}

	// Load layout metadata if available for LAYOUT_VS_CHECKING verification type
	var layoutMetadata map[string]interface{}
	if req.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking &&
		req.S3Refs.Processing.LayoutMetadata.Key != "" {
		h.log.Info("loading_layout_metadata_for_turn2", map[string]interface{}{
			"verification_id": req.VerificationID,
			"layout_key":      req.S3Refs.Processing.LayoutMetadata.Key,
		})

		layoutData, err := h.s3.LoadLayoutMetadata(ctx, req.S3Refs.Processing.LayoutMetadata)
		if err != nil {
			h.log.Warn("layout_metadata_load_failed_turn2", map[string]interface{}{
				"error":           err.Error(),
				"verification_id": req.VerificationID,
				"layout_key":      req.S3Refs.Processing.LayoutMetadata.Key,
				"impact":          "continuing_without_layout_metadata",
			})
		} else {
			// Convert schema.LayoutMetadata to map[string]interface{} for template processing
			layoutMetadata = map[string]interface{}{
				"LayoutId":           layoutData.LayoutId,
				"LayoutPrefix":       layoutData.LayoutPrefix,
				"VendingMachineId":   layoutData.VendingMachineId,
				"Location":           layoutData.Location,
				"MachineStructure":   layoutData.MachineStructure,
				"ProductPositionMap": layoutData.ProductPositionMap,
			}

			// Extract RowCount and ColumnCount from MachineStructure if available
			if layoutData.MachineStructure != nil {
				if rowCount, ok := layoutData.MachineStructure["RowCount"]; ok {
					layoutMetadata["RowCount"] = rowCount
				}
				if colCount, ok := layoutData.MachineStructure["ColumnCount"]; ok {
					layoutMetadata["ColumnCount"] = colCount
				}
				if rowLabels, ok := layoutData.MachineStructure["RowLabels"]; ok {
					layoutMetadata["RowLabels"] = rowLabels
				}
			}

			h.log.Info("layout_metadata_loaded_successfully_turn2", map[string]interface{}{
				"verification_id":    req.VerificationID,
				"layout_id":          layoutData.LayoutId,
				"layout_prefix":      layoutData.LayoutPrefix,
				"vending_machine_id": layoutData.VendingMachineId,
				"location":           layoutData.Location,
				"has_row_count":      layoutMetadata["RowCount"] != nil,
				"has_column_count":   layoutMetadata["ColumnCount"] != nil,
				"has_row_labels":     layoutMetadata["RowLabels"] != nil,
			})
		}
	}

	// Create verification context
	vCtx := &schema.VerificationContext{
		VerificationId:   req.VerificationID,
		VerificationAt:   req.VerificationContext.VerificationAt,
		VerificationType: req.VerificationContext.VerificationType,
		VendingMachineId: req.VerificationContext.VendingMachineId,
		LayoutId:         req.VerificationContext.LayoutId,
		LayoutPrefix:     req.VerificationContext.LayoutPrefix,
	}

	// Generate Turn2 prompt
	prompt, promptProcessor, err := h.promptService.GenerateTurn2PromptWithMetrics(
		ctx,
		vCtx,
		loadResult.SystemPrompt,
		loadedTurn1Response,
		nil,
		layoutMetadata,
	)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeInternal,
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
		return nil, models.S3Reference{}, models.S3Reference{}, wfErr
	}

	// Store prompt processor metrics
	_, err = h.s3.StoreTemplateProcessor(ctx, req.VerificationID, promptProcessor)
	if err != nil {
		h.log.Warn("failed_to_store_prompt_processor", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Persist the rendered Turn2 prompt
	promptRef, err = h.s3.StorePrompt(ctx, req.VerificationID, 2, prompt)
	if err != nil {
		h.log.Warn("failed_to_store_turn2_prompt", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	// Invoke Bedrock with conversation history
	bedrockResponse, err := h.bedrock.ConverseWithHistory(
		ctx,
		loadResult.SystemPrompt,
		prompt,
		loadResult.Base64Image,
		loadResult.ImageFormat,
		loadedTurn1Response,
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
		return nil, models.S3Reference{}, models.S3Reference{}, wfErr
	}

	// Build TurnResponse for raw logging
	turn2Raw := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Prompt:    prompt,
		ImageUrls: map[string]string{"checking": req.S3Refs.Images.CheckingBase64.Key},
		Response: schema.BedrockApiResponse{
			Content:    bedrockResponse.Content,
			Thinking:   bedrockResponse.Thinking,
			StopReason: bedrockResponse.CompletionReason,
			ModelId:    bedrockResponse.ModelId,
		},
		LatencyMs: bedrockResponse.LatencyMs,
		TokenUsage: &schema.TokenUsage{
			InputTokens:    bedrockResponse.InputTokens,
			OutputTokens:   bedrockResponse.OutputTokens,
			ThinkingTokens: bedrockResponse.ThinkingTokens,
			TotalTokens:    bedrockResponse.InputTokens + bedrockResponse.OutputTokens + bedrockResponse.ThinkingTokens,
		},
		Stage: bedrock.AnalysisStageTurn2,
		Metadata: map[string]interface{}{
			"verificationId":   req.VerificationID,
			"verificationType": req.VerificationContext.VerificationType,
			"bedrockMetadata": map[string]interface{}{
				"modelId": bedrockResponse.ModelId,
				// "requestId" field removed as it's not available in BedrockResponse
				"stopReason": bedrockResponse.CompletionReason,
			},
			"processingMetadata": map[string]interface{}{
				"executionTimeMs": bedrockResponse.ProcessingTimeMs,
				"retryAttempts":   0,
			},
			"promptMetadata": map[string]interface{}{
				"imageSize":           len(loadResult.Base64Image),
				"imageType":           "checking",
				"promptTokenEstimate": promptProcessor.InputTokens,
			},
		},
	}

	// Prepare to store raw response later using the envelope
	var rawResponseRef models.S3Reference

	// Parse Bedrock response
	markdownResponse, err := bedrockparser.ParseTurn2BedrockResponseAsMarkdown(bedrockResponse.Content)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeInternal,
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
		return nil, models.S3Reference{}, models.S3Reference{}, wfErr
	}

	// Store markdown response
	_, err = h.s3.StoreTurn2Markdown(ctx, req.VerificationID, markdownResponse.ComparisonMarkdown)
	if err != nil {
		h.log.Warn("failed_to_store_markdown_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Build and store conversation history
	bedrockTextOutput := markdownResponse.ComparisonMarkdown

	// Load turn1 conversation history if available
	var turn1Messages []schema.BedrockMessage
	if req.S3Refs.Turn1.Conversation.Key != "" {
		var convData struct {
			Messages []schema.BedrockMessage `json:"messages"`
		}
		if err := h.s3.LoadJSON(ctx, req.S3Refs.Turn1.Conversation, &convData); err != nil {
			h.log.Warn("failed_to_load_turn1_conversation", map[string]interface{}{
				"error":           err.Error(),
				"verification_id": req.VerificationID,
			})
		} else {
			turn1Messages = convData.Messages
		}
	}

	var thinkingBlocks []interface{}
	if blocks, ok := bedrockResponse.Metadata["thinking_blocks"]; ok {
		if arr, ok := blocks.([]interface{}); ok {
			thinkingBlocks = arr
		}
	}

	convRef, convErr := h.s3.StoreTurn2Conversation(ctx, req.VerificationID, turn1Messages, loadResult.SystemPrompt, prompt, loadResult.Base64Image, bedrockTextOutput, bedrockResponse.Thinking, thinkingBlocks, &schema.TokenUsage{
		InputTokens:    bedrockResponse.InputTokens,
		OutputTokens:   bedrockResponse.OutputTokens,
		ThinkingTokens: 0,
		TotalTokens:    bedrockResponse.InputTokens + bedrockResponse.OutputTokens,
	}, bedrockResponse.LatencyMs, "", bedrockResponse.ModelId, map[string]interface{}{
		"stopReason": bedrockResponse.CompletionReason,
	})
	if convErr != nil {
		h.log.Warn("failed_to_store_conversation", map[string]interface{}{
			"error":           convErr.Error(),
			"verification_id": req.VerificationID,
		})
	}

	// Parse structured data from response
	parsedData, err := bedrockparser.ParseTurn2Response(bedrockResponse.Content)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeInternal,
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
		return nil, models.S3Reference{}, models.S3Reference{}, wfErr
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

	// Store raw and processed Turn2 outputs
	var processedRef models.S3Reference
	// Store the raw response as structured JSON without additional encoding
	rawResponseRef, err = h.s3.StoreTurn2RawResponse(ctx, req.VerificationID, turn2Raw)
	if err != nil {
		h.log.Warn("failed_to_store_raw_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	processedRef, err = h.s3.StoreTurn2ProcessedResponse(ctx, req.VerificationID, parsedData)
	if err != nil {
		h.log.Warn("failed_to_store_processed_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	// Create processing metrics
	processingMetrics := &schema.ProcessingMetrics{
		Turn2: &schema.TurnMetrics{
			StartTime:        startTime.Format(time.RFC3339),
			EndTime:          time.Now().Format(time.RFC3339),
			TotalTimeMs:      time.Since(startTime).Milliseconds(),
			BedrockLatencyMs: bedrockResponse.LatencyMs,
			ProcessingTimeMs: time.Since(startTime).Milliseconds() - bedrockResponse.LatencyMs,
			RetryAttempts:    0,
			TokenUsage: &schema.TokenUsage{
				InputTokens:    bedrockResponse.InputTokens,
				OutputTokens:   bedrockResponse.OutputTokens,
				ThinkingTokens: bedrockResponse.ThinkingTokens,
				TotalTokens:    bedrockResponse.InputTokens + bedrockResponse.OutputTokens + bedrockResponse.ThinkingTokens,
			},
		},
	}

	// Store processing metrics
	_, err = h.s3.StoreProcessingMetrics(ctx, req.VerificationID, processingMetrics)
	if err != nil {
		h.log.Warn("failed_to_store_processing_metrics", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Create Turn2 response
	response := &models.Turn2Response{
		S3Refs: models.Turn2ResponseS3Refs{
			RawResponse:       rawResponseRef,
			ProcessedResponse: processedRef,
		},
		Status: models.ConvertFromSchemaStatus(schema.StatusTurn2Completed),
		Summary: models.Summary{
			AnalysisStage:    models.StageProcessing,
			ProcessingTimeMs: processingMetrics.Turn2.TotalTimeMs,
			TokenUsage: models.TokenUsage{
				InputTokens:    bedrockResponse.InputTokens,
				OutputTokens:   bedrockResponse.OutputTokens,
				ThinkingTokens: bedrockResponse.ThinkingTokens,
				TotalTokens:    bedrockResponse.InputTokens + bedrockResponse.OutputTokens + bedrockResponse.ThinkingTokens,
			},
			BedrockRequestID: "", // RequestID not available in BedrockResponse
		},
		Discrepancies:       parsedData.Discrepancies,
		VerificationOutcome: finalStatus,
	}

	h.log.Info("turn2_processing_completed", map[string]interface{}{
		"verification_id":       req.VerificationID,
		"verification_type":     req.VerificationContext.VerificationType,
		"verification_outcome":  finalStatus,
		"discrepancy_count":     len(parsedData.Discrepancies),
		"total_processing_time": processingMetrics.Turn2.TotalTimeMs,
		"input_tokens":          processingMetrics.Turn2.TokenUsage.InputTokens,
		"output_tokens":         processingMetrics.Turn2.TokenUsage.OutputTokens,
		"thinking_tokens":       processingMetrics.Turn2.TokenUsage.ThinkingTokens,
		"total_tokens":          processingMetrics.Turn2.TokenUsage.TotalTokens,
	})

	// Build status entry and turn metrics for DynamoDB update
	statusEntry := schema.StatusHistoryEntry{
		Status:           schema.StatusTurn2Completed,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn2Combined",
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		Stage:            "turn2_completion",
		Metrics: map[string]interface{}{
			"discrepancy_count":    len(parsedData.Discrepancies),
			"verification_outcome": finalStatus,
		},
	}

	turn2Metrics := &schema.TurnMetrics{
		StartTime:        startTime.Format(time.RFC3339),
		EndTime:          time.Now().Format(time.RFC3339),
		TotalTimeMs:      processingMetrics.Turn2.TotalTimeMs,
		BedrockLatencyMs: bedrockResponse.LatencyMs,
		ProcessingTimeMs: processingMetrics.Turn2.TotalTimeMs - bedrockResponse.LatencyMs,
		RetryAttempts:    0,
		TokenUsage:       processingMetrics.Turn2.TokenUsage,
	}

	// Update VerificationResults with Turn2 details
	// Convert discrepancies to schema format
	discrepancies := make([]schema.Discrepancy, 0, len(parsedData.Discrepancies))
	for _, d := range parsedData.Discrepancies {
		desc := fmt.Sprintf("%s expected %s found %s", d.Item, d.Expected, d.Found)
		discrepancies = append(discrepancies, schema.Discrepancy{
			Type:        d.Type,
			Description: desc,
			Severity:    d.Severity,
		})
	}

	// Prepare conversation history entry
	turnEntry := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Prompt:    prompt,
		ImageUrls: map[string]string{
			"checking": req.S3Refs.Images.CheckingBase64.Key,
		},
		Response: schema.BedrockApiResponse{
			Content: bedrockResponse.Content,
			ModelId: bedrockResponse.ModelId,
		},
		LatencyMs: bedrockResponse.LatencyMs,
		TokenUsage: &schema.TokenUsage{
			InputTokens:  bedrockResponse.InputTokens,
			OutputTokens: bedrockResponse.OutputTokens,
			TotalTokens:  bedrockResponse.InputTokens + bedrockResponse.OutputTokens,
		},
		Stage: "CHECKING_ANALYSIS",
	}
	dynamoOK := h.dynamoManager.UpdateTurn2Completion(ctx, Turn2Result{
		VerificationID:     req.VerificationID,
		VerificationAt:     req.VerificationContext.VerificationAt,
		StatusEntry:        statusEntry,
		TurnEntry:          turnEntry,
		Metrics:            turn2Metrics,
		VerificationStatus: finalStatus,
		Discrepancies:      discrepancies,
		ComparisonSummary:  refinedSummary,
		ConversationRef:    &convRef,
	})
	if !dynamoOK {
		h.log.Warn("dynamodb_update_turn2_failed", map[string]interface{}{
			"verification_id": req.VerificationID,
		})
	}

	// Populate summary fields with completion details
	response.Summary.DiscrepanciesFound = len(parsedData.Discrepancies)
	response.Summary.ComparisonCompleted = true
	response.Summary.ConversationCompleted = true
	response.Summary.DynamodbUpdated = dynamoOK
	response.Summary.VerificationType = req.VerificationContext.VerificationType
	response.Summary.BedrockLatencyMs = bedrockResponse.LatencyMs
	response.Summary.S3StorageCompleted = true

	return response, convRef, promptRef, nil
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

// dynamoRetryOperation implements retry logic for DynamoDB operations
func (h *Turn2Handler) dynamoRetryOperation(ctx context.Context, operation func() error, operationName string, verificationID string) error {
	const maxRetries = 3
	const baseDelay = 200 * time.Millisecond
	const maxDelay = 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := operation()
		if err == nil {
			if attempt > 0 {
				h.log.Info("dynamo_retry_successful", map[string]interface{}{
					"operation":       operationName,
					"attempt":         attempt + 1,
					"verification_id": verificationID,
				})
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if wfErr, ok := err.(*errors.WorkflowError); ok && !wfErr.Retryable {
			h.log.Debug("dynamo_non_retryable_error", map[string]interface{}{
				"operation":       operationName,
				"error":           err.Error(),
				"attempt":         attempt + 1,
				"verification_id": verificationID,
			})
			break
		}

		// Don't retry on the last attempt
		if attempt == maxRetries-1 {
			break
		}

		// Calculate delay with exponential backoff
		multiplier := 1
		for i := 0; i < attempt; i++ {
			multiplier *= 2
		}
		delay := time.Duration(int64(baseDelay) * int64(multiplier))
		if delay > maxDelay {
			delay = maxDelay
		}

		h.log.Debug("retrying_dynamo_operation", map[string]interface{}{
			"operation":       operationName,
			"attempt":         attempt + 1,
			"max_attempts":    maxRetries,
			"delay_ms":        delay.Milliseconds(),
			"error":           err.Error(),
			"verification_id": verificationID,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	h.log.Error("dynamo_all_retry_attempts_failed", map[string]interface{}{
		"operation":       operationName,
		"max_attempts":    maxRetries,
		"final_error":     lastErr.Error(),
		"verification_id": verificationID,
	})

	return lastErr
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

	// Update verification status with retry logic
	if err := h.dynamoRetryOperation(ctx, func() error {
		return h.dynamo.UpdateVerificationStatusEnhanced(ctx, req.VerificationID, req.VerificationContext.VerificationAt, statusEntry)
	}, "UpdateVerificationStatusEnhanced", req.VerificationID); err != nil {
		h.log.Warn("dynamodb_status_error_update_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	// Update error tracking with retry logic
	if err := h.dynamoRetryOperation(ctx, func() error {
		return h.dynamo.UpdateErrorTracking(ctx, req.VerificationID, tracking)
	}, "UpdateErrorTracking", req.VerificationID); err != nil {
		h.log.Warn("dynamodb_error_tracking_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}

	// Update conversation turn with retry logic
	turnEntry := &schema.TurnResponse{
		TurnId:    2,
		Timestamp: schema.FormatISO8601(),
		Stage:     stage,
		Metadata: map[string]interface{}{
			"error": wfErr.Message,
		},
	}
	if err := h.dynamoRetryOperation(ctx, func() error {
		return h.dynamo.UpdateConversationTurn(ctx, req.VerificationID, turnEntry)
	}, "UpdateConversationTurn", req.VerificationID); err != nil {
		h.log.Warn("conversation_history_error_update_failed", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
	}
}

// Handle is currently not implemented for Turn2Handler.
func (h *Turn2Handler) Handle(ctx context.Context, req *models.Turn2Request) (*schema.CombinedTurnResponse, error) {
	return nil, fmt.Errorf("Handle not implemented")
}

// HandleForStepFunction processes a Turn2 request and formats the output for Step Functions
func (h *Turn2Handler) HandleForStepFunction(ctx context.Context, req *models.Turn2Request) (*models.StepFunctionResponse, error) {
	startTime := time.Now()
	contextLogger := h.log.WithCorrelationId(req.VerificationID).WithFields(map[string]interface{}{
		"verificationId": req.VerificationID,
		"turnId":         2,
		"schemaVersion":  h.validator.GetSchemaVersion(),
	})

	contextLogger.Info("Starting ExecuteTurn2Combined", map[string]interface{}{
		"verification_type": req.VerificationContext.VerificationType,
		"layout_id":         req.VerificationContext.LayoutId,
	})

	turn2Resp, convRef, promptRef, err := h.ProcessTurn2Request(ctx, req)
	if err != nil {
		contextLogger.Error("turn2_processing_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	builder := NewResponseBuilder(h.cfg)
	stepFunctionResp := builder.BuildTurn2StepFunctionResponse(req, turn2Resp, promptRef, convRef)

	duration := time.Since(startTime)
	contextLogger.Info("Completed ExecuteTurn2Combined", map[string]interface{}{
		"duration_ms":     duration.Milliseconds(),
		"status":          stepFunctionResp.Status,
		"discrepancy_cnt": len(turn2Resp.Discrepancies),
	})

	return stepFunctionResp, nil
}
