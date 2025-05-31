package handler

import (
	"context"
	"time"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// initializeProcessingMetrics creates initial processing metrics
func (h *Turn2Handler) initializeProcessingMetrics() *schema.ProcessingMetrics {
	return &schema.ProcessingMetrics{
		WorkflowTotal: &schema.WorkflowMetrics{
			StartTime:     schema.FormatISO8601(),
			FunctionCount: 0,
		},
		Turn2: &schema.TurnMetrics{
			StartTime:     schema.FormatISO8601(),
			RetryAttempts: 0,
		},
	}
}

// createContextLogger creates a logger with context fields for Turn2
func (h *Turn2Handler) createContextLogger(req *models.Turn2Request) logger.Logger {
	return h.log.WithCorrelationId(req.VerificationID).WithFields(map[string]interface{}{
		"verificationId": req.VerificationID,
		"turnId":         2,
		"schemaVersion":  h.validator.GetSchemaVersion(),
		"functionName":   "ExecuteTurn2Combined",
	})
}

// updateStatus updates status with error handling
func (h *Turn2Handler) updateStatus(ctx context.Context, verificationID, status, stage string, metadata map[string]interface{}) {
	if err := h.statusTracker.UpdateStatusWithHistory(ctx, verificationID, status, stage, metadata); err != nil {
		h.log.Warn("failed to update status", map[string]interface{}{
			"error":  err.Error(),
			"status": status,
		})
	}
}

// handleContextLoadError handles errors during context loading
func (h *Turn2Handler) handleContextLoadError(ctx context.Context, verificationID string, loadResult *LoadResult, contextLogger logger.Logger) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("context_loading", "failed", loadResult.Duration, map[string]interface{}{
		"s3_operations": 2,
		"error_type":    "s3_retrieval_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn2Error, "context_loading_failed", map[string]interface{}{
		"error_details": loadResult.Error.Error(),
	})

	if workflowErr, ok := loadResult.Error.(*errors.WorkflowError); ok {
		contextLogger.Error("turn2_context_loading_error", map[string]interface{}{
			"error_type":    string(workflowErr.Type),
			"error_code":    workflowErr.Code,
			"retryable":     workflowErr.Retryable,
			"severity":      string(workflowErr.Severity),
			"s3_operations": 2,
		})
	}

	return nil, loadResult.Error
}

// recordContextLoadSuccess records successful context loading
func (h *Turn2Handler) recordContextLoadSuccess(ctx context.Context, verificationID string, loadResult *LoadResult) {
	h.processingTracker.RecordStage("context_loading", "completed", loadResult.Duration, map[string]interface{}{
		"s3_operations":        2,
		"concurrent_loading":   true,
		"system_prompt_length": len(loadResult.SystemPrompt),
		"image_data_length":    len(loadResult.Base64Image),
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn2ContextLoaded, "context_loading", map[string]interface{}{
		"system_prompt_size":  len(loadResult.SystemPrompt),
		"image_size":          len(loadResult.Base64Image),
		"loading_duration_ms": loadResult.Duration.Milliseconds(),
	})
}

// PromptResult contains prompt generation results
type PromptResult struct {
	Prompt            string
	TemplateProcessor *schema.TemplateProcessor
	Duration          time.Duration
	Error             error
}

// generateTurn2Prompt generates the Turn2 prompt with Turn1 context
func (h *Turn2Handler) generateTurn2Prompt(ctx context.Context, req *models.Turn2Request, systemPrompt string, turn1Response *schema.TurnResponse, turn1RawResponse []byte) *PromptResult {
	startTime := time.Now()

	// Create verification context for Turn2
	vCtx := &schema.VerificationContext{
		VerificationId:   req.VerificationID,
		VerificationAt:   req.VerificationContext.VerificationAt,
		VerificationType: req.VerificationContext.VerificationType,
		VendingMachineId: req.VerificationContext.VendingMachineId,
		LayoutId:         req.VerificationContext.LayoutId,
		LayoutPrefix:     req.VerificationContext.LayoutPrefix,
	}

	prompt, templateProcessor, err := h.promptService.GenerateTurn2PromptWithMetrics(
		ctx,
		vCtx,
		systemPrompt,
		turn1Response,
		turn1RawResponse,
		nil, // TODO: Pass layout metadata when available
	)

	return &PromptResult{
		Prompt:            prompt,
		TemplateProcessor: templateProcessor,
		Duration:          time.Since(startTime),
		Error:             err,
	}
}

// handlePromptError handles errors during prompt generation
func (h *Turn2Handler) handlePromptError(ctx context.Context, verificationID string, result *PromptResult, contextLogger logger.Logger) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("prompt_generation", "failed", result.Duration, map[string]interface{}{
		"template_version": h.cfg.Prompts.TemplateVersion,
		"error_type":       "prompt_generation_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTemplateProcessingError, "prompt_generation_failed", map[string]interface{}{
		"error_details": result.Error.Error(),
	})

	promptErr := errors.NewInternalError("prompt_service", result.Error).
		WithContext("template_version", h.cfg.Prompts.TemplateVersion)

	enrichedErr := errors.SetVerificationID(promptErr, verificationID)

	contextLogger.Error("turn2_prompt_generation_error", map[string]interface{}{
		"template_version": h.cfg.Prompts.TemplateVersion,
		"error":            result.Error.Error(),
	})

	return nil, enrichedErr
}

// handleBedrockError handles errors during Bedrock invocation
func (h *Turn2Handler) handleBedrockError(ctx context.Context, verificationID string, result *InvokeResult) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("bedrock_invocation", "failed", result.Duration, map[string]interface{}{
		"model_id":   h.cfg.AWS.BedrockModel,
		"max_tokens": h.cfg.Processing.MaxTokens,
		"error_type": "bedrock_api_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn2Error, "bedrock_invocation_failed", map[string]interface{}{
		"error_details": result.Error.Error(),
	})

	return nil, result.Error
}

// recordBedrockSuccess records successful Bedrock invocation
func (h *Turn2Handler) recordBedrockSuccess(ctx context.Context, verificationID string, result *InvokeResult, templateProcessor *schema.TemplateProcessor) {
	metadata := h.bedrockInvoker.GetInvocationMetadata(result.Response, result.Duration)
	h.processingTracker.RecordStage("bedrock_invocation", "completed", result.Duration, metadata)

	h.updateStatus(ctx, verificationID, schema.StatusTurn2BedrockCompleted, "bedrock_completion", map[string]interface{}{
		"token_usage":        result.Response.TokenUsage,
		"bedrock_request_id": result.Response.RequestID,
		"latency_ms":         result.Duration.Milliseconds(),
	})

	if templateProcessor != nil {
		templateProcessor.InputTokens = result.Response.TokenUsage.InputTokens
		templateProcessor.OutputTokens = result.Response.TokenUsage.OutputTokens
	}
}

// updateProcessingMetrics updates processing metrics with final values for Turn2
func (h *Turn2Handler) updateProcessingMetrics(metrics *schema.ProcessingMetrics, totalDuration time.Duration, invokeResult *InvokeResult) {
	metrics.WorkflowTotal.EndTime = schema.FormatISO8601()
	metrics.WorkflowTotal.TotalTimeMs = totalDuration.Milliseconds()
	metrics.WorkflowTotal.FunctionCount = h.processingTracker.GetStageCount()

	metrics.Turn2.EndTime = schema.FormatISO8601()
	metrics.Turn2.TotalTimeMs = totalDuration.Milliseconds()
	metrics.Turn2.BedrockLatencyMs = invokeResult.Duration.Milliseconds()
	metrics.Turn2.ProcessingTimeMs = totalDuration.Milliseconds() - invokeResult.Duration.Milliseconds()
	metrics.Turn2.TokenUsage = &invokeResult.Response.TokenUsage
}

// updateInitializationFile writes the final status back to the input initialization.json
func (h *Turn2Handler) updateInitializationFile(ctx context.Context, req *models.Turn2Request, status string, contextLogger logger.Logger) {
	ref := req.InputInitializationFileRef
	if ref.Bucket == "" || ref.Key == "" {
		contextLogger.Warn("initialization reference missing, skipping update", nil)
		return
	}

	initData, err := h.s3.LoadInitializationData(ctx, ref)
	if err != nil {
		contextLogger.Warn("failed to load initialization.json for update", map[string]interface{}{
			"error": err.Error(),
			"key":   ref.Key,
		})
		return
	}

	initData.VerificationContext.Status = status
	initData.VerificationContext.LastUpdatedAt = schema.FormatISO8601()
	initData.SchemaVersion = schema.SchemaVersion

	if _, err := h.s3.StoreJSONAtReference(ctx, ref, initData); err != nil {
		contextLogger.Warn("failed to store updated initialization.json", map[string]interface{}{
			"error": err.Error(),
			"key":   ref.Key,
		})
		return
	}

	contextLogger.Info("updated initialization.json status", map[string]interface{}{
		"key":    ref.Key,
		"status": status,
	})
}

// validateAndLogCompletion validates response and logs completion for Turn2
func (h *Turn2Handler) validateAndLogCompletion(response *models.Turn2Response, totalDuration time.Duration, bedrockResp *schema.BedrockResponse, contextLogger logger.Logger) {
	// Create Turn2Response for validation (response is already validated)
	turn2Response := &models.Turn2Response{
		S3Refs: models.Turn2ResponseS3Refs{
			RawResponse:       models.S3Reference{}, // Already validated during storage
			ProcessedResponse: models.S3Reference{}, // Already validated during storage
		},
		Status: models.StatusTurn2Completed,
		Summary: models.Summary{
			AnalysisStage:    models.StageProcessing,
			ProcessingTimeMs: totalDuration.Milliseconds(),
			TokenUsage: models.TokenUsage{
				InputTokens:  bedrockResp.InputTokens,
				OutputTokens: bedrockResp.OutputTokens,
				TotalTokens:  bedrockResp.InputTokens + bedrockResp.OutputTokens,
			},
			BedrockRequestID: "", // RequestID not available in BedrockResponse
		},
		Discrepancies:       response.Discrepancies,
		VerificationOutcome: response.VerificationOutcome,
	}

	if err := h.validator.ValidateTurn2Response(turn2Response); err != nil {
		contextLogger.Error("turn2 response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	contextLogger.Info("Completed ExecuteTurn2Combined", map[string]interface{}{
		"duration_ms":          totalDuration.Milliseconds(),
		"processing_stages":    h.processingTracker.GetStageCount(),
		"status_updates":       h.statusTracker.GetHistoryCount(),
		"discrepancy_count":    len(response.Discrepancies),
		"verification_outcome": response.VerificationOutcome,
	})
}

// handleStepFunctionEvent handles Step Functions event format
func (h *Turn2Handler) handleStepFunctionEvent(ctx context.Context, req *models.Turn2Request) (interface{}, error) {
	h.log.Info("processing_step_function_event", map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationContext.VerificationType,
		"checking_image_key": req.S3Refs.Images.CheckingBase64.Key,
	})

	h.log.LogReceivedEvent(req)

	turn2Resp, convRef, promptRef, err := h.ProcessTurn2Request(ctx, req)
	if err != nil {
		return nil, err
	}

	builder := NewResponseBuilder(h.cfg)
	stepFunctionResponse := builder.BuildTurn2StepFunctionResponse(req, turn2Resp, promptRef, convRef)

	h.log.LogOutputEvent(stepFunctionResponse)

	return stepFunctionResponse, nil
}

// handleDirectRequest handles direct request format
func (h *Turn2Handler) handleDirectRequest(ctx context.Context, req *models.Turn2Request) (interface{}, error) {
	h.log.LogReceivedEvent(req)

	response, _, _, err := h.ProcessTurn2Request(ctx, req)
	if err != nil {
		return nil, err
	}

	h.log.LogOutputEvent(response)
	return response, nil
}
