package handler

import (
	"context"
	"encoding/json"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// initializeProcessingMetrics creates initial processing metrics
func (h *Handler) initializeProcessingMetrics() *schema.ProcessingMetrics {
	return &schema.ProcessingMetrics{
		WorkflowTotal: &schema.WorkflowMetrics{
			StartTime:     schema.FormatISO8601(),
			FunctionCount: 0,
		},
		Turn1: &schema.TurnMetrics{
			StartTime:     schema.FormatISO8601(),
			RetryAttempts: 0,
		},
	}
}

// createContextLogger creates a logger with context fields
func (h *Handler) createContextLogger(req *models.Turn1Request) logger.Logger {
	return h.log.WithCorrelationId(req.VerificationID).WithFields(map[string]interface{}{
		"verificationId": req.VerificationID,
		"turnId":         1,
		"schemaVersion":  h.validator.GetSchemaVersion(),
		"functionName":   "ExecuteTurn1Combined",
	})
}

// updateStatus updates status with error handling
func (h *Handler) updateStatus(ctx context.Context, verificationID, status, stage string, metadata map[string]interface{}) {
	if err := h.statusTracker.UpdateStatusWithHistory(ctx, verificationID, status, stage, metadata); err != nil {
		h.log.Warn("failed to update status", map[string]interface{}{
			"error":  err.Error(),
			"status": status,
		})
	}
}

// handleContextLoadError handles errors during context loading with enhanced error details
func (h *Handler) handleContextLoadError(ctx context.Context, verificationID string, loadResult *LoadResult, contextLogger logger.Logger) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("context_loading", "failed", loadResult.Duration, map[string]interface{}{
		"s3_operations": 2,
		"error_type":    "s3_retrieval_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn1Error, "context_loading_failed", map[string]interface{}{
		"error_details": loadResult.Error.Error(),
	})

	// Enhanced error handling with detailed context
	var enhancedErr *errors.WorkflowError
	if workflowErr, ok := loadResult.Error.(*errors.WorkflowError); ok {
		enhancedErr = workflowErr.
			WithVerificationID(verificationID).
			WithComponent("ContextLoader").
			WithOperation("LoadContext").
			WithCategory(errors.CategoryTransient).
			WithRetryStrategy(errors.RetryExponential).
			SetMaxRetries(3).
			WithSeverity(errors.ErrorSeverityHigh).
			WithSuggestions(
				"Check S3 bucket permissions and connectivity",
				"Verify that all required S3 objects exist",
				"Ensure proper IAM roles are configured",
				"Check S3 bucket region configuration",
			).
			WithRecoveryHints(
				"Retry the operation with exponential backoff",
				"Verify S3 object keys and bucket names",
				"Check network connectivity to S3",
			)

		contextLogger.Error("resource loading error", map[string]interface{}{
			"error_type":     string(enhancedErr.Type),
			"error_code":     enhancedErr.Code,
			"message":        enhancedErr.Message,
			"retryable":      enhancedErr.Retryable,
			"severity":       string(enhancedErr.Severity),
			"category":       string(enhancedErr.Category),
			"component":      enhancedErr.Component,
			"operation":      enhancedErr.Operation,
			"suggestions":    enhancedErr.Suggestions,
			"recovery_hints": enhancedErr.RecoveryHints,
			"s3_operations":  2,
		})
	} else {
		enhancedErr = errors.WrapError(loadResult.Error, errors.ErrorTypeS3,
			"context loading failed", true).
			WithVerificationID(verificationID).
			WithComponent("ContextLoader").
			WithOperation("LoadContext").
			WithCategory(errors.CategoryTransient).
			WithRetryStrategy(errors.RetryExponential).
			SetMaxRetries(3).
			WithSeverity(errors.ErrorSeverityHigh)
	}

	return nil, enhancedErr
}

// recordContextLoadSuccess records successful context loading
func (h *Handler) recordContextLoadSuccess(ctx context.Context, verificationID string, loadResult *LoadResult) {
	h.processingTracker.RecordStage("context_loading", "completed", loadResult.Duration, map[string]interface{}{
		"s3_operations":        2,
		"concurrent_loading":   true,
		"system_prompt_length": len(loadResult.SystemPrompt),
		"image_data_length":    len(loadResult.Base64Image),
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn1ContextLoaded, "context_loading", map[string]interface{}{
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

// generatePrompt generates the Turn1 prompt
func (h *Handler) generatePrompt(ctx context.Context, req *models.Turn1Request, systemPrompt string) *PromptResult {
	startTime := time.Now()
	prompt, templateProcessor, err := h.promptGenerator.GenerateTurn1PromptEnhanced(ctx, req.VerificationContext, systemPrompt)

	return &PromptResult{
		Prompt:            prompt,
		TemplateProcessor: templateProcessor,
		Duration:          time.Since(startTime),
		Error:             err,
	}
}

// handlePromptError handles errors during prompt generation with enhanced error details
func (h *Handler) handlePromptError(ctx context.Context, verificationID string, result *PromptResult, contextLogger logger.Logger) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("prompt_generation", "failed", result.Duration, map[string]interface{}{
		"template_version": h.cfg.Prompts.TemplateVersion,
		"error_type":       "prompt_generation_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTemplateProcessingError, "prompt_generation_failed", map[string]interface{}{
		"error_details": result.Error.Error(),
	})

	// Enhanced prompt error with detailed context
	var enhancedErr *errors.WorkflowError
	if workflowErr, ok := result.Error.(*errors.WorkflowError); ok {
		enhancedErr = workflowErr.
			WithVerificationID(verificationID).
			WithComponent("PromptService").
			WithOperation("GenerateTurn1Prompt").
			WithCategory(errors.CategoryPermanent).
			WithRetryStrategy(errors.RetryNone).
			WithSeverity(errors.ErrorSeverityCritical).
			WithContext("template_version", h.cfg.Prompts.TemplateVersion).
			WithContext("template_base_path", h.cfg.Prompts.TemplateBasePath).
			WithSuggestions(
				"Check template syntax and variable bindings",
				"Verify all required template variables are provided",
				"Ensure template files exist and are accessible",
				"Validate template version compatibility",
			).
			WithRecoveryHints(
				"Review template file structure and syntax",
				"Check template variable mappings",
				"Verify template loading configuration",
			)
	} else {
		enhancedErr = errors.NewInternalError("prompt_service", result.Error).
			WithVerificationID(verificationID).
			WithComponent("PromptService").
			WithOperation("GenerateTurn1Prompt").
			WithCategory(errors.CategoryPermanent).
			WithRetryStrategy(errors.RetryNone).
			WithSeverity(errors.ErrorSeverityCritical).
			WithContext("template_version", h.cfg.Prompts.TemplateVersion).
			WithContext("template_base_path", h.cfg.Prompts.TemplateBasePath)
	}

	contextLogger.Error("prompt generation error", map[string]interface{}{
		"error_type":       string(enhancedErr.Type),
		"error_code":       enhancedErr.Code,
		"message":          enhancedErr.Message,
		"severity":         string(enhancedErr.Severity),
		"category":         string(enhancedErr.Category),
		"component":        enhancedErr.Component,
		"operation":        enhancedErr.Operation,
		"suggestions":      enhancedErr.Suggestions,
		"recovery_hints":   enhancedErr.RecoveryHints,
		"template_version": h.cfg.Prompts.TemplateVersion,
	})

	return nil, enhancedErr
}

// handleBedrockError handles errors during Bedrock invocation with enhanced error details
func (h *Handler) handleBedrockError(ctx context.Context, verificationID string, result *InvokeResult) (*schema.CombinedTurnResponse, error) {
	h.processingTracker.RecordStage("bedrock_invocation", "failed", result.Duration, map[string]interface{}{
		"model_id":   h.cfg.AWS.BedrockModel,
		"max_tokens": h.cfg.Processing.MaxTokens,
		"error_type": "bedrock_api_failure",
	})

	h.updateStatus(ctx, verificationID, schema.StatusTurn1Error, "bedrock_invocation_failed", map[string]interface{}{
		"error_details": result.Error.Error(),
	})

	// Enhanced Bedrock error with detailed context
	var enhancedErr *errors.WorkflowError
	if workflowErr, ok := result.Error.(*errors.WorkflowError); ok {
		// Determine retry strategy and category based on error type
		category := errors.CategoryServer
		retryStrategy := errors.RetryExponential
		maxRetries := 3
		severity := errors.ErrorSeverityMedium

		if workflowErr.Type == errors.ErrorTypeThrottling {
			category = errors.CategoryCapacity
			retryStrategy = errors.RetryJittered
			maxRetries = 5
			severity = errors.ErrorSeverityLow
		} else if workflowErr.Type == errors.ErrorTypeValidation {
			category = errors.CategoryValidation
			retryStrategy = errors.RetryNone
			maxRetries = 0
			severity = errors.ErrorSeverityHigh
		}

		enhancedErr = workflowErr.
			WithVerificationID(verificationID).
			WithComponent("BedrockService").
			WithOperation("InvokeBedrock").
			WithCategory(category).
			WithRetryStrategy(retryStrategy).
			SetMaxRetries(maxRetries).
			WithSeverity(severity).
			WithContext("model_id", h.cfg.AWS.BedrockModel).
			WithContext("max_tokens", h.cfg.Processing.MaxTokens).
			WithContext("temperature", h.cfg.Processing.Temperature).
			WithContext("top_p", h.cfg.Processing.TopP).
			WithSuggestions(
				"Check Bedrock service availability and quotas",
				"Verify model permissions and access policies",
				"Ensure prompt and image sizes are within limits",
				"Check for service throttling or rate limits",
			).
			WithRecoveryHints(
				"Retry with exponential backoff if retryable",
				"Check AWS service health dashboard",
				"Review Bedrock quotas and limits",
			)
	} else {
		enhancedErr = errors.WrapError(result.Error, errors.ErrorTypeBedrock,
			"Bedrock invocation failed", true).
			WithVerificationID(verificationID).
			WithComponent("BedrockService").
			WithOperation("InvokeBedrock").
			WithCategory(errors.CategoryServer).
			WithRetryStrategy(errors.RetryExponential).
			SetMaxRetries(3).
			WithSeverity(errors.ErrorSeverityMedium).
			WithContext("model_id", h.cfg.AWS.BedrockModel).
			WithContext("max_tokens", h.cfg.Processing.MaxTokens)
	}

	return nil, enhancedErr
}

// recordBedrockSuccess records successful Bedrock invocation
func (h *Handler) recordBedrockSuccess(ctx context.Context, verificationID string, result *InvokeResult, templateProcessor *schema.TemplateProcessor) {
	metadata := h.bedrockInvoker.GetInvocationMetadata(result.Response, result.Duration)
	h.processingTracker.RecordStage("bedrock_invocation", "completed", result.Duration, metadata)

	h.updateStatus(ctx, verificationID, schema.StatusTurn1BedrockCompleted, "bedrock_completion", map[string]interface{}{
		"token_usage":        result.Response.TokenUsage,
		"bedrock_request_id": result.Response.RequestID,
		"latency_ms":         result.Duration.Milliseconds(),
	})

	if templateProcessor != nil {
		templateProcessor.InputTokens = result.Response.TokenUsage.InputTokens
		templateProcessor.OutputTokens = result.Response.TokenUsage.OutputTokens
	}
}

// recordStorageSuccess records successful storage operations
func (h *Handler) recordStorageSuccess(result *StorageResult) {
	metadata := h.storageManager.GetStorageMetadata(result)
	h.processingTracker.RecordStage("response_processing", "completed", result.Duration, metadata)
}

// updateProcessingMetrics updates processing metrics with final values
func (h *Handler) updateProcessingMetrics(metrics *schema.ProcessingMetrics, totalDuration time.Duration, invokeResult *InvokeResult) {
	metrics.WorkflowTotal.EndTime = schema.FormatISO8601()
	metrics.WorkflowTotal.TotalTimeMs = totalDuration.Milliseconds()
	metrics.WorkflowTotal.FunctionCount = h.processingTracker.GetStageCount()

	metrics.Turn1.EndTime = schema.FormatISO8601()
	metrics.Turn1.TotalTimeMs = totalDuration.Milliseconds()
	metrics.Turn1.BedrockLatencyMs = invokeResult.Duration.Milliseconds()
	metrics.Turn1.ProcessingTimeMs = totalDuration.Milliseconds() - invokeResult.Duration.Milliseconds()
	metrics.Turn1.TokenUsage = &invokeResult.Response.TokenUsage
}

// updateInitializationFile writes the final status back to the input initialization.json
func (h *Handler) updateInitializationFile(ctx context.Context, req *models.Turn1Request, status string, contextLogger logger.Logger) {
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
// validateAndLogCompletion validates response and logs completion
func (h *Handler) validateAndLogCompletion(response *schema.CombinedTurnResponse, totalDuration time.Duration, bedrockResp *models.BedrockResponse, contextLogger logger.Logger) {
	// Create Turn1Response for validation
	turn1Response := &models.Turn1Response{
		S3Refs: models.Turn1ResponseS3Refs{
			RawResponse:       models.S3Reference{}, // Already validated during storage
			ProcessedResponse: models.S3Reference{}, // Already validated during storage
		},
		Status: models.StatusTurn1Completed,
		Summary: models.Summary{
			AnalysisStage:    models.StageReferenceAnalysis,
			ProcessingTimeMs: totalDuration.Milliseconds(),
			TokenUsage:       bedrockResp.TokenUsage,
			BedrockRequestID: bedrockResp.RequestID,
		},
	}

	if err := h.validator.ValidateResponse(turn1Response); err != nil {
		contextLogger.Error("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	contextLogger.Info("Completed ExecuteTurn1Combined", map[string]interface{}{
		"duration_ms":       totalDuration.Milliseconds(),
		"processing_stages": h.processingTracker.GetStageCount(),
		"status_updates":    h.statusTracker.GetHistoryCount(),
		"schema_version":    h.validator.GetSchemaVersion(),
		"template_used":     response.TemplateUsed,
	})
}

// handleStepFunctionEvent handles Step Functions event format
func (h *Handler) handleStepFunctionEvent(ctx context.Context, event StepFunctionEvent) (interface{}, error) {
	h.log.Info("processing_step_function_event", map[string]interface{}{
		"schema_version":      event.SchemaVersion,
		"verification_id":     event.VerificationID,
		"status":              event.Status,
		"s3_references_count": len(event.S3References),
	})

	h.log.LogReceivedEvent(event)

	req, err := h.eventTransformer.TransformStepFunctionEvent(ctx, event)
	if err != nil {
		h.log.Error("failed_to_transform_step_function_event", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": event.VerificationID,
		})
		return nil, err
	}

	// Use the new HandleForStepFunction method that returns StepFunctionResponse
	stepFunctionResponse, err := h.HandleForStepFunction(ctx, req)
	if err != nil {
		return nil, err
	}

	h.log.LogOutputEvent(stepFunctionResponse)

	return stepFunctionResponse, nil
}

// handleDirectRequest handles direct request format with enhanced error handling
func (h *Handler) handleDirectRequest(ctx context.Context, event json.RawMessage) (interface{}, error) {
	var req models.Turn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		validationErr := errors.NewValidationError(
			"invalid input payload",
			map[string]interface{}{
				"payload_size": len(event),
				"parse_error":  err.Error(),
			}).
			WithComponent("RequestParser").
			WithOperation("UnmarshalDirectRequest").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.ErrorSeverityHigh).
			WithSuggestions(
				"Verify request payload format matches Turn1Request schema",
				"Check JSON syntax and structure",
				"Ensure all required fields are present",
				"Validate field data types and values",
			).
			WithRecoveryHints(
				"Review request payload structure",
				"Check API documentation for correct format",
				"Validate JSON syntax",
			)

		h.log.Error("input validation failed", map[string]interface{}{
			"error_type":         string(validationErr.Type),
			"error_code":         validationErr.Code,
			"message":            validationErr.Message,
			"severity":           string(validationErr.Severity),
			"category":           string(validationErr.Category),
			"component":          validationErr.Component,
			"operation":          validationErr.Operation,
			"suggestions":        validationErr.Suggestions,
			"recovery_hints":     validationErr.RecoveryHints,
			"payload_size_bytes": len(event),
		})

		return nil, validationErr
	}

	h.log.LogReceivedEvent(req)

	response, err := h.Handle(ctx, &req)
	if err != nil {
		// Ensure error has proper context for direct requests
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			enrichedErr := workflowErr.WithContext("request_type", "direct")
			return nil, enrichedErr
		}
		return nil, err
	}

	h.log.LogOutputEvent(response)
	return response, nil
}
