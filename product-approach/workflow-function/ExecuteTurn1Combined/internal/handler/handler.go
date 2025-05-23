// internal/handler/handler.go - ENHANCED VERSION WITH COMPREHENSIVE DESIGN INTEGRATION
package handler

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/ExecuteTurn1Combined/internal/validation"
	
	// Using shared packages correctly
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Handler orchestrates the ExecuteTurn1Combined workflow with enhanced tracking and observability.
type Handler struct {
	cfg           config.Config
	s3            services.S3StateManager
	bedrock       services.BedrockService
	dynamo        services.DynamoDBService
	promptService services.PromptService
	validator     *validation.SchemaValidator
	log           logger.Logger
	
	// Enhanced tracking
	processingStages []schema.ProcessingStage
	statusHistory    []schema.StatusHistoryEntry
	startTime        time.Time
}

// NewHandler wires together all dependencies for the Lambda with enhanced capabilities.
func NewHandler(
	s3Mgr services.S3StateManager,
	bedrockClient services.BedrockService,
	dynamoClient services.DynamoDBService,
	promptGen services.PromptService,
	log logger.Logger,
	cfg *config.Config,
) (*Handler, error) {
	return &Handler{
		cfg:              *cfg,
		s3:               s3Mgr,
		bedrock:          bedrockClient,
		dynamo:           dynamoClient,
		promptService:    promptGen,
		validator:        validation.NewSchemaValidator(),
		log:              log,
		processingStages: make([]schema.ProcessingStage, 0),
		statusHistory:    make([]schema.StatusHistoryEntry, 0),
	}, nil
}

// Handle executes a single Turn-1 verification cycle with comprehensive tracking.
func (h *Handler) Handle(ctx context.Context, req *models.Turn1Request) (*schema.CombinedTurnResponse, error) {
	h.startTime = time.Now()
	
	// Initialize processing metrics
	processingMetrics := &schema.ProcessingMetrics{
		WorkflowTotal: &schema.WorkflowMetrics{
			StartTime:     schema.FormatISO8601(),
			FunctionCount: 0,
		},
		Turn1: &schema.TurnMetrics{
			StartTime:        schema.FormatISO8601(),
			RetryAttempts:    0,
		},
	}
	
	// Schema validation using standardized validation
	if err := h.validator.ValidateRequest(req); err != nil {
		h.recordProcessingStage("validation", "failed", time.Since(h.startTime), map[string]interface{}{
			"validation_error": err.Error(),
		})
		return nil, errors.NewValidationError("request validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	
	h.recordProcessingStage("validation", "completed", time.Since(h.startTime), nil)
	
	// Create context logger with enhanced tracking
	contextLogger := h.log.WithCorrelationId(req.VerificationID).
		WithFields(map[string]interface{}{
			"verificationId": req.VerificationID,
			"turnId":         1,
			"schemaVersion":  h.validator.GetSchemaVersion(),
			"functionName":   "ExecuteTurn1Combined",
		})
	
	contextLogger.Info("Starting ExecuteTurn1Combined with enhanced tracking", map[string]interface{}{
		"verification_type": req.VerificationContext.VerificationType,
		"s3_system_prompt":  req.S3Refs.Prompts.System.Key,
		"s3_reference_img":  req.S3Refs.Images.ReferenceBase64.Key,
		"processing_mode":   "combined_function",
	})

	// STAGE 1: Update status to TURN1_STARTED
	if err := h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1Started, "initialization", map[string]interface{}{
		"function_start_time": schema.FormatISO8601(),
		"verification_type":   req.VerificationContext.VerificationType,
	}); err != nil {
		contextLogger.Warn("failed to update initial status", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// STAGE 2: Concurrent context loading (system prompt & base64 image)
	contextLoadStart := time.Now()
	var (
		systemPrompt string
		base64Img    string
		loadErr      error
	)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		sp, err := h.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3, 
				"failed to load system prompt", true)
			
			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading")
			
			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}
		systemPrompt = sp
	}()

	go func() {
		defer wg.Done()
		img, err := h.s3.LoadBase64Image(ctx, req.S3Refs.Images.ReferenceBase64)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3, 
				"failed to load reference image", true)
			
			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Images.ReferenceBase64.Key).
				WithContext("stage", "context_loading").
				WithContext("image_size", req.S3Refs.Images.ReferenceBase64.Size)
			
			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}
		base64Img = img
	}()

	wg.Wait()
	contextLoadDuration := time.Since(contextLoadStart)
	
	if loadErr != nil {
		h.recordProcessingStage("context_loading", "failed", contextLoadDuration, map[string]interface{}{
			"s3_operations": 2,
			"error_type":    "s3_retrieval_failure",
		})
		
		h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1Error, "context_loading_failed", map[string]interface{}{
			"error_details": loadErr.Error(),
		})
		
		if workflowErr, ok := loadErr.(*errors.WorkflowError); ok {
			contextLogger.Error("resource loading error", map[string]interface{}{
				"error_type":    string(workflowErr.Type),
				"error_code":    workflowErr.Code,
				"retryable":     workflowErr.Retryable,
				"severity":      string(workflowErr.Severity),
				"s3_operations": 2,
			})
		}
		
		return nil, loadErr
	}

	// Record successful context loading
	h.recordProcessingStage("context_loading", "completed", contextLoadDuration, map[string]interface{}{
		"s3_operations":        2,
		"concurrent_loading":   true,
		"system_prompt_length": len(systemPrompt),
		"image_data_length":    len(base64Img),
	})
	
	// Update status to TURN1_CONTEXT_LOADED
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1ContextLoaded, "context_loading", map[string]interface{}{
		"system_prompt_size": len(systemPrompt),
		"image_size":         len(base64Img),
		"loading_duration_ms": contextLoadDuration.Milliseconds(),
	})

	// STAGE 3: Generate Turn-1 prompt with template processing
	promptStart := time.Now()
	turnPrompt, templateProcessor, err := h.generateTurn1PromptEnhanced(ctx, req.VerificationContext, systemPrompt)
	promptDuration := time.Since(promptStart)
	
	if err != nil {
		h.recordProcessingStage("prompt_generation", "failed", promptDuration, map[string]interface{}{
			"template_version": h.cfg.Prompts.TemplateVersion,
			"error_type":       "prompt_generation_failure",
		})
		
		h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTemplateProcessingError, "prompt_generation_failed", map[string]interface{}{
			"error_details": err.Error(),
		})
		
		promptErr := errors.NewInternalError("prompt_service", err).
			WithContext("template_version", h.cfg.Prompts.TemplateVersion).
			WithContext("verification_type", req.VerificationContext.VerificationType)
		
		enrichedErr := errors.SetVerificationID(promptErr, req.VerificationID)
		
		contextLogger.Error("prompt generation error", map[string]interface{}{
			"template_version":    h.cfg.Prompts.TemplateVersion,
			"verification_type":   req.VerificationContext.VerificationType,
			"system_prompt_size":  len(systemPrompt),
		})
		
		return nil, enrichedErr
	}

	// Record successful prompt generation
	promptMetadata := map[string]interface{}{
		"template_version":     h.cfg.Prompts.TemplateVersion,
		"prompt_length":        len(turnPrompt),
		"template_used":        "turn1-combined",
		"context_enrichment":   true,
	}
	
	if templateProcessor != nil {
		promptMetadata["template_processing_time_ms"] = templateProcessor.ProcessingTime
		promptMetadata["token_estimate"] = templateProcessor.TokenEstimate
	}
	
	h.recordProcessingStage("prompt_generation", "completed", promptDuration, promptMetadata)
	
	// Update status to TURN1_PROMPT_PREPARED
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1PromptPrepared, "prompt_generation", map[string]interface{}{
		"prompt_length":    len(turnPrompt),
		"template_version": h.cfg.Prompts.TemplateVersion,
		"generation_duration_ms": promptDuration.Milliseconds(),
	})

	// STAGE 4: Bedrock invocation
	bedrockStart := time.Now()
	
	// Update status to TURN1_BEDROCK_INVOKED
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1BedrockInvoked, "bedrock_invocation", map[string]interface{}{
		"model_id":      h.cfg.AWS.BedrockModel,
		"max_tokens":    h.cfg.Processing.MaxTokens,
		"invocation_time": schema.FormatISO8601(),
	})
	
	resp, err := h.bedrock.Converse(ctx, systemPrompt, turnPrompt, base64Img)
	bedrockDuration := time.Since(bedrockStart)
	
	if err != nil {
		h.recordProcessingStage("bedrock_invocation", "failed", bedrockDuration, map[string]interface{}{
			"model_id":     h.cfg.AWS.BedrockModel,
			"max_tokens":   h.cfg.Processing.MaxTokens,
			"error_type":   "bedrock_api_failure",
		})
		
		h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1Error, "bedrock_invocation_failed", map[string]interface{}{
			"error_details": err.Error(),
		})
		
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			enrichedErr := workflowErr.WithContext("model_id", h.cfg.AWS.BedrockModel).
				WithContext("max_tokens", h.cfg.Processing.MaxTokens).
				WithContext("prompt_size", len(turnPrompt)).
				WithContext("image_size", len(base64Img))
			
			finalErr := errors.SetVerificationID(enrichedErr, req.VerificationID)
			
			if workflowErr.Retryable {
				contextLogger.Warn("bedrock retryable error", map[string]interface{}{
					"error_code":     workflowErr.Code,
					"api_source":     string(workflowErr.APISource),
					"retry_attempt":  "will_be_retried_by_step_functions",
				})
			} else {
				contextLogger.Error("bedrock non-retryable error", map[string]interface{}{
					"error_code":   workflowErr.Code,
					"api_source":   string(workflowErr.APISource),
					"severity":     string(workflowErr.Severity),
				})
			}
			
			return nil, finalErr
		} else {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeBedrock, 
				"bedrock invocation failed", false).
				WithAPISource(errors.APISourceConverse)
			
			enrichedErr := errors.SetVerificationID(wrappedErr, req.VerificationID)
			
			contextLogger.Error("bedrock unexpected error", map[string]interface{}{
				"original_error": err.Error(),
			})
			
			return nil, enrichedErr
		}
	}

	// Record successful Bedrock invocation
	h.recordProcessingStage("bedrock_invocation", "completed", bedrockDuration, map[string]interface{}{
		"model_id":            h.cfg.AWS.BedrockModel,
		"input_tokens":        resp.TokenUsage.InputTokens,
		"output_tokens":       resp.TokenUsage.OutputTokens,
		"thinking_tokens":     resp.TokenUsage.ThinkingTokens,
		"total_tokens":        resp.TokenUsage.TotalTokens,
		"bedrock_request_id":  resp.RequestID,
		"latency_ms":          bedrockDuration.Milliseconds(),
	})
	
	// Update status to TURN1_BEDROCK_COMPLETED
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1BedrockCompleted, "bedrock_completion", map[string]interface{}{
		"token_usage":         resp.TokenUsage,
		"bedrock_request_id":  resp.RequestID,
		"latency_ms":          bedrockDuration.Milliseconds(),
	})

	// STAGE 5: Response processing and S3 storage
	processingStart := time.Now()
	
	// Update status to TURN1_RESPONSE_PROCESSING
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1ResponseProcessing, "response_processing", nil)
	
	rawRef, err := h.s3.StoreRawResponse(ctx, req.VerificationID, resp.Raw)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3, 
			"store raw response failed", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("response_size", len(resp.Raw))
		
		enrichedErr := errors.SetVerificationID(s3Err, req.VerificationID)
		
		contextLogger.Warn("s3 raw-store warning", map[string]interface{}{
			"response_size_bytes": len(resp.Raw),
			"bucket":              h.cfg.AWS.S3Bucket,
		})
		
		return nil, enrichedErr
	}
	
	procRef, err := h.s3.StoreProcessedAnalysis(ctx, req.VerificationID, resp.Processed)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3, 
			"store processed analysis failed", true).
			WithContext("verification_id", req.VerificationID)
		
		enrichedErr := errors.SetVerificationID(s3Err, req.VerificationID)
		
		contextLogger.Warn("s3 processed-store warning", map[string]interface{}{
			"bucket": h.cfg.AWS.S3Bucket,
		})
		
		return nil, enrichedErr
	}
	
	processingDuration := time.Since(processingStart)
	h.recordProcessingStage("response_processing", "completed", processingDuration, map[string]interface{}{
		"s3_objects_created":   2,
		"raw_response_size":    len(resp.Raw),
		"processed_ref_key":    procRef.Key,
		"raw_ref_key":          rawRef.Key,
	})

	// STAGE 6: Update processing metrics
	totalDuration := time.Since(h.startTime)
	processingMetrics.WorkflowTotal.EndTime = schema.FormatISO8601()
	processingMetrics.WorkflowTotal.TotalTimeMs = totalDuration.Milliseconds()
	processingMetrics.WorkflowTotal.FunctionCount = len(h.processingStages)
	
	processingMetrics.Turn1.EndTime = schema.FormatISO8601()
	processingMetrics.Turn1.TotalTimeMs = totalDuration.Milliseconds()
	processingMetrics.Turn1.BedrockLatencyMs = bedrockDuration.Milliseconds()
	processingMetrics.Turn1.ProcessingTimeMs = totalDuration.Milliseconds() - bedrockDuration.Milliseconds()
	processingMetrics.Turn1.TokenUsage = &resp.TokenUsage

	// STAGE 7: DynamoDB updates and conversation history (enhanced)
	// Create a channel to track async update completion
	updateComplete := make(chan error, 2)
	
	go func() {
		// Update verification status
		updateErr := h.dynamo.UpdateVerificationStatus(ctx, req.VerificationID, models.StatusTurn1Completed, resp.TokenUsage)
		if updateErr != nil {
			asyncLogger := contextLogger.WithFields(map[string]interface{}{
				"async_operation": "verification_status_update",
				"table":           h.cfg.AWS.DynamoDBVerificationTable,
			})
			
			if workflowErr, ok := updateErr.(*errors.WorkflowError); ok {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error_type": string(workflowErr.Type),
					"error_code": workflowErr.Code,
					"retryable":  workflowErr.Retryable,
				})
			} else {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error": updateErr.Error(),
				})
			}
		}
		updateComplete <- updateErr
		
		// Record conversation history
		conversationTurn := &models.ConversationTurn{
			VerificationID:   req.VerificationID,
			TurnID:           1,
			RawResponseRef:   rawRef,
			ProcessedRef:     procRef,
			TokenUsage:       resp.TokenUsage,
			BedrockRequestID: resp.RequestID,
			Timestamp:        time.Now(),
		}
		
		historyErr := h.dynamo.RecordConversationTurn(ctx, conversationTurn)
		if historyErr != nil {
			contextLogger.Warn("conversation history recording failed", map[string]interface{}{
				"error": historyErr.Error(),
				"table": h.cfg.AWS.DynamoDBConversationTable,
			})
		}
		updateComplete <- historyErr
	}()

	// Final status update
	h.updateStatusWithHistory(ctx, req.VerificationID, schema.StatusTurn1Completed, "completion", map[string]interface{}{
		"total_duration_ms": totalDuration.Milliseconds(),
		"processing_stages": len(h.processingStages),
		"status_updates":    len(h.statusHistory),
	})

	// STAGE 8: Build enhanced combined response
	combinedResponse := h.buildCombinedTurnResponse(req, resp, rawRef, procRef, templateProcessor, processingMetrics, totalDuration)

	// Wait for async DynamoDB updates to complete (with timeout)
	updateTimeout := time.After(5 * time.Second)
	updatesReceived := 0
	for updatesReceived < 2 {
		select {
		case err := <-updateComplete:
			updatesReceived++
			if err != nil {
				contextLogger.Warn("async update error received", map[string]interface{}{
					"update_number": updatesReceived,
					"error": err.Error(),
				})
			}
		case <-updateTimeout:
			contextLogger.Warn("async updates timed out", map[string]interface{}{
				"updates_received": updatesReceived,
				"timeout_seconds": 5,
			})
			goto ContinueProcessing
		}
	}
	
ContinueProcessing:
	// Validate response before returning
	if err := h.validator.ValidateResponse(&models.Turn1Response{
		S3Refs: models.Turn1ResponseS3Refs{
			RawResponse:       rawRef,
			ProcessedResponse: procRef,
		},
		Status: models.StatusTurn1Completed,
		Summary: models.Summary{
			AnalysisStage:    models.StageReferenceAnalysis,
			ProcessingTimeMs: totalDuration.Milliseconds(),
			TokenUsage:       resp.TokenUsage,
			BedrockRequestID: resp.RequestID,
		},
	}); err != nil {
		contextLogger.Error("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return nil, errors.NewValidationError("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	contextLogger.Info("Completed ExecuteTurn1Combined with enhanced tracking", map[string]interface{}{
		"duration_ms":        totalDuration.Milliseconds(),
		"input_tokens":       resp.TokenUsage.InputTokens,
		"output_tokens":      resp.TokenUsage.OutputTokens,
		"thinking_tokens":    resp.TokenUsage.ThinkingTokens,
		"total_tokens":       resp.TokenUsage.TotalTokens,
		"bedrock_request_id": resp.RequestID,
		"s3_objects_created": 2,
		"dynamo_updates":     2,
		"processing_stages":  len(h.processingStages),
		"status_updates":     len(h.statusHistory),
		"status":             string(models.StatusTurn1Completed),
		"schema_version":     h.validator.GetSchemaVersion(),
		"template_used":      combinedResponse.TemplateUsed,
	})
	
	return combinedResponse, nil
}

// Helper method to generate Turn1 prompt with enhanced template processing
func (h *Handler) generateTurn1PromptEnhanced(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, *schema.TemplateProcessor, error) {
	// For now, use the existing prompt service but capture processing info
	prompt, err := h.promptService.GenerateTurn1Prompt(ctx, vCtx, systemPrompt)
	if err != nil {
		return "", nil, err
	}
	
	// Create template processor info for tracking
	templateProcessor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      "turn1-combined",
			TemplateVersion: h.cfg.Prompts.TemplateVersion,
			TemplateType:    "turn1-layout-vs-checking",
			Content:         prompt,
		},
		ContextData: map[string]interface{}{
			"verificationType": vCtx.VerificationType,
			"layoutMetadata":   vCtx.LayoutMetadata,
			"systemPromptSize": len(systemPrompt),
		},
		ProcessedPrompt: prompt,
		ProcessingTime:  10, // Placeholder - would be actual processing time
		TokenEstimate:   len(prompt) / 4,
	}
	
	return prompt, templateProcessor, nil
}

// Helper method to record processing stages
func (h *Handler) recordProcessingStage(stageName, status string, duration time.Duration, metadata map[string]interface{}) {
	stage := schema.ProcessingStage{
		StageName: stageName,
		StartTime: h.startTime.Add(duration - duration).Format(time.RFC3339), // Approximate start time
		EndTime:   h.startTime.Add(duration).Format(time.RFC3339),
		Duration:  duration.Milliseconds(),
		Status:    status,
		Metadata:  metadata,
	}
	
	h.processingStages = append(h.processingStages, stage)
}

// Helper method to update status with history tracking
func (h *Handler) updateStatusWithHistory(ctx context.Context, verificationID, status, stage string, metadata map[string]interface{}) error {
	statusEntry := schema.StatusHistoryEntry{
		Status:           status,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn1Combined",
		ProcessingTimeMs: time.Since(h.startTime).Milliseconds(),
		Stage:            stage,
		Metrics:          metadata,
	}
	
	h.statusHistory = append(h.statusHistory, statusEntry)
	
	// In a full implementation, this would also update DynamoDB
	// For now, we just track locally
	return nil
}

// Helper method to build the combined turn response
func (h *Handler) buildCombinedTurnResponse(
	req *models.Turn1Request,
	resp *models.BedrockResponse,
	rawRef, procRef models.S3Reference,
	templateProcessor *schema.TemplateProcessor,
	processingMetrics *schema.ProcessingMetrics,
	totalDuration time.Duration,
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
			"model_id":         h.cfg.AWS.BedrockModel,
			"verification_id":  req.VerificationID,
			"function_name":    "ExecuteTurn1Combined",
		},
	}
	
	// Build context enrichment
	contextEnrichment := map[string]interface{}{
		"verification_type":    req.VerificationContext.VerificationType,
		"layout_integrated":    req.VerificationContext.LayoutId != 0,
		"historical_context":   req.VerificationContext.HistoricalContext != nil,
		"processing_stages":    len(h.processingStages),
		"status_updates":       len(h.statusHistory),
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
		ProcessingStages:  h.processingStages,
		InternalPrompt:    "", // Would be filled with actual internal prompt
		TemplateUsed:      templateUsed,
		ContextEnrichment: contextEnrichment,
	}
	
	return combinedResponse
}

// HandleTurn1Combined is the Lambda entrypoint invoked by Step Functions.
func (h *Handler) HandleTurn1Combined(ctx context.Context, event json.RawMessage) (interface{}, error) {
	var req models.Turn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		validationErr := errors.NewValidationError(
			"invalid input payload",
			map[string]interface{}{
				"payload_size": len(event),
				"parse_error":  err.Error(),
			})
		
		h.log.Error("input validation failed", map[string]interface{}{
			"payload_size_bytes": len(event),
			"error_details":      err.Error(),
		})
		
		return nil, validationErr
	}
	
	// Log received event
	h.log.LogReceivedEvent(req)
	
	// Call enhanced handle method
	response, err := h.Handle(ctx, &req)
	if err != nil {
		return nil, err
	}
	
	// Log output event
	h.log.LogOutputEvent(response)
	
	return response, nil
}