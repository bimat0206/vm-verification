// internal/handler/handler.go - ENHANCED VERSION WITH COMPREHENSIVE DESIGN INTEGRATION
package handler

import (
	"context"
	"encoding/json"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"

	"workflow-function/ExecuteTurn1Combined/internal/bedrockparser"
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
	log           logger.Logger

	// Components for better code organization
	processingTracker *ProcessingStagesTracker
	statusTracker     *StatusTracker
	responseBuilder   *ResponseBuilder
	eventTransformer  *EventTransformer
	promptGenerator   *PromptGenerator
	contextLoader     *ContextLoader
	historicalLoader  *HistoricalContextLoader
	bedrockInvoker    *BedrockInvoker
	storageManager    *StorageManager
	dynamoManager     *DynamoManager
	validator         *Validator
	startTime         time.Time
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
		log:              log,
		responseBuilder:  NewResponseBuilder(*cfg),
		eventTransformer: NewEventTransformer(s3Mgr, log),
		promptGenerator:  NewPromptGenerator(promptGen, *cfg),
		contextLoader:    NewContextLoader(s3Mgr, log),
		historicalLoader: NewHistoricalContextLoader(s3Mgr, dynamoClient, log),
		bedrockInvoker:   NewBedrockInvoker(bedrockClient, *cfg, log),
		storageManager:   NewStorageManager(s3Mgr, *cfg, log),
		dynamoManager:    NewDynamoManager(dynamoClient, *cfg, log),
		validator:        NewValidator(),
	}, nil
}

// Handle executes a single Turn-1 verification cycle with comprehensive tracking.
func (h *Handler) Handle(ctx context.Context, req *models.Turn1Request) (resp *schema.CombinedTurnResponse, err error) {
	h.startTime = time.Now()
	h.processingTracker = NewProcessingStagesTracker(h.startTime)
	h.statusTracker = NewStatusTracker(h.startTime)

	// Initialize processing metrics
	processingMetrics := h.initializeProcessingMetrics()
	// Create context logger
	contextLogger := h.createContextLogger(req)
	defer func() {
		finalStatus := schema.StatusTurn1Completed
		if err != nil {
			if wfErr, ok := err.(*errors.WorkflowError); ok {
				finalStatus = string(wfErr.Type)
			} else {
				finalStatus = schema.StatusTurn1Error
			}
		}
		h.updateInitializationFile(ctx, req, finalStatus, contextLogger)
	}()
	// Validate request
	if err := h.validator.ValidateRequest(req); err != nil {
		h.processingTracker.RecordStage("validation", "failed", time.Since(h.startTime), map[string]interface{}{
			"validation_error": err.Error(),
		})
		return nil, err
	}
	h.processingTracker.RecordStage("validation", "completed", time.Since(h.startTime), nil)

	// Update initial status
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1Started, "initialization", map[string]interface{}{
		"function_start_time": schema.FormatISO8601(),
		"verification_type":   req.VerificationContext.VerificationType,
	})

	// STAGE 1: Load context (system prompt & base64 image)
	loadResult := h.contextLoader.LoadContext(ctx, req)
	if loadResult.Error != nil {
		return h.handleContextLoadError(ctx, req.VerificationID, loadResult, contextLogger)
	}
	h.recordContextLoadSuccess(ctx, req.VerificationID, loadResult)

	// STAGE 2: Load historical context if applicable
	historicalDuration, _ := h.historicalLoader.LoadHistoricalContext(ctx, req, contextLogger)
	if historicalDuration > 0 {
		h.processingTracker.RecordStage("historical_context_loading", "completed", historicalDuration, map[string]interface{}{
			"has_historical_context": req.VerificationContext.HistoricalContext != nil,
		})
	}

	// STAGE 3: Generate prompt
	promptResult := h.generatePrompt(ctx, req, loadResult.SystemPrompt)
	if promptResult.Error != nil {
		return h.handlePromptError(ctx, req.VerificationID, promptResult, contextLogger)
	}

	// Record prompt generation success
	h.processingTracker.RecordStage("prompt_generation", "completed", promptResult.Duration, map[string]interface{}{
		"template_version": h.cfg.Prompts.TemplateVersion,
		"prompt_length":    len(promptResult.Prompt),
		"template_used":    "turn1-combined",
	})

	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1PromptPrepared, "prompt_generation", map[string]interface{}{
		"prompt_length":          len(promptResult.Prompt),
		"template_version":       h.cfg.Prompts.TemplateVersion,
		"generation_duration_ms": promptResult.Duration.Milliseconds(),
	})

	// STAGE 4: Invoke Bedrock
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1BedrockInvoked, "bedrock_invocation", map[string]interface{}{
		"model_id":        h.cfg.AWS.BedrockModel,
		"max_tokens":      h.cfg.Processing.MaxTokens,
		"invocation_time": schema.FormatISO8601(),
	})

	invokeResult := h.bedrockInvoker.InvokeBedrock(ctx, loadResult.SystemPrompt, promptResult.Prompt, loadResult.Base64Image, req.VerificationID)
	if invokeResult.Error != nil {
		return h.handleBedrockError(ctx, req.VerificationID, invokeResult)
	}
	h.recordBedrockSuccess(ctx, req.VerificationID, invokeResult, promptResult.TemplateProcessor)

	var bedrockTextOutput string
	if processedMap, ok := invokeResult.Response.Processed.(map[string]interface{}); ok {
		if contentStr, ok := processedMap["content"].(string); ok {
			bedrockTextOutput = contentStr
		}
	}
	if bedrockTextOutput == "" {
		var rawResp struct {
			Content string `json:"content"`
		}
		_ = json.Unmarshal(invokeResult.Response.Raw, &rawResp)
		bedrockTextOutput = rawResp.Content
	}
	parsedTurn1Data, parseErr := bedrockparser.ParseBedrockResponseAsMarkdown(bedrockTextOutput)
	if parseErr != nil {
		contextLogger.Warn("failed to parse bedrock response", map[string]interface{}{
			"error": parseErr.Error(),
		})
	}

	// STAGE 5: Store responses
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1ResponseProcessing, "response_processing", nil)

	storageResult := h.storageManager.StoreResponses(ctx, req, invokeResult, promptResult, len(loadResult.Base64Image), parsedTurn1Data)
	if storageResult.Error != nil {
		return nil, storageResult.Error
	}
	h.recordStorageSuccess(storageResult)

	// STAGE 6: Store prompt
	promptStart := time.Now()
	promptRef, err := h.storageManager.StorePrompt(ctx, req, 1, promptResult)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store prompt", true).
			WithContext("verification_id", req.VerificationID)
	}

	// Record prompt storage success
	h.processingTracker.RecordStage("prompt_storage", "completed", time.Since(promptStart), map[string]interface{}{
		"prompt_length":  len(promptResult.Prompt),
		"prompt_ref_key": promptRef.Key,
	})

	// STAGE 7: Update metrics and async operations
	totalDuration := time.Since(h.startTime)
	h.updateProcessingMetrics(processingMetrics, totalDuration, invokeResult)

	// Prepare DynamoDB entries
	statusEntry := schema.StatusHistoryEntry{
		Status:           schema.StatusTurn1Completed,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn1Combined",
		ProcessingTimeMs: totalDuration.Milliseconds(),
		Stage:            "turn1_completion",
	}

	turnEntry := &schema.TurnResponse{
		TurnId:     1,
		Timestamp:  schema.FormatISO8601(),
		Prompt:     "",
		ImageUrls:  map[string]string{},
		Response:   schema.BedrockApiResponse{RequestId: invokeResult.Response.RequestID},
		LatencyMs:  invokeResult.Duration.Milliseconds(),
		TokenUsage: &invokeResult.Response.TokenUsage,
		Stage:      "REFERENCE_ANALYSIS",
		Metadata: map[string]interface{}{
			"model_id":        h.cfg.AWS.BedrockModel,
			"verification_id": req.VerificationID,
			"function_name":   "ExecuteTurn1Combined",
		},
	}

	var turn1MetricsForDB *schema.TurnMetrics = nil
	if processingMetrics != nil && processingMetrics.Turn1 != nil {
		turn1MetricsForDB = processingMetrics.Turn1
	}

	// Perform DynamoDB updates synchronously
	dynamoOK := h.dynamoManager.UpdateTurn1Completion(ctx, req.VerificationID, req.VerificationContext.VerificationAt, statusEntry, turnEntry, turn1MetricsForDB, &storageResult.ProcessedRef, &convRef)

	// Final status update
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1Completed, "completion", map[string]interface{}{
		"total_duration_ms": totalDuration.Milliseconds(),
		"processing_stages": h.processingTracker.GetStageCount(),
		"status_updates":    h.statusTracker.GetHistoryCount(),
		"dynamodb_updated":  dynamoOK,
	})

	// Build response with all required fields for schema v2.1.0
	response := h.responseBuilder.BuildCombinedTurnResponse(
		req, promptResult.Prompt, promptRef, storageResult.RawRef, storageResult.ProcessedRef, convRef,
		invokeResult.Response, h.processingTracker.GetStages(), totalDuration.Milliseconds(),
		invokeResult.Duration.Milliseconds(), dynamoOK,
	)

	// Validate and log completion
	h.validateAndLogCompletion(response, totalDuration, invokeResult.Response, contextLogger)

	return response, nil
}

// HandleForStepFunction processes Turn1 and returns StepFunctionResponse
func (h *Handler) HandleForStepFunction(ctx context.Context, req *models.Turn1Request) (resp *models.StepFunctionResponse, err error) {
	// Initialize tracking
	h.startTime = time.Now()
	h.processingTracker = NewProcessingStagesTracker(h.startTime)
	h.statusTracker = NewStatusTracker(h.startTime)

	processingMetrics := h.initializeProcessingMetrics()
	contextLogger := h.createContextLogger(req)
	defer func() {
		finalStatus := schema.StatusTurn1Completed
		if err != nil {
			if wfErr, ok := err.(*errors.WorkflowError); ok {
				finalStatus = string(wfErr.Type)
			} else {
				finalStatus = schema.StatusTurn1Error
			}
		}
		h.updateInitializationFile(ctx, req, finalStatus, contextLogger)
	}()
	contextLogger.Info("Starting ExecuteTurn1Combined", map[string]interface{}{
		"verification_type": req.VerificationContext.VerificationType,
		"layout_id":         req.VerificationContext.LayoutId,
		"schema_version":    h.validator.GetSchemaVersion(),
	})

	// STAGE 1: Validation
	h.processingTracker.RecordStage("validation", "started", 0, nil)
	if err := h.validator.ValidateRequest(req); err != nil {
		contextLogger.Error("input validation error", map[string]interface{}{
			"validation_error": err.Error(),
		})
		h.processingTracker.RecordStage("validation", "failed", 0, map[string]interface{}{
			"error": err.Error(),
		})
		h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1Error, "validation_failed", map[string]interface{}{
			"error_details": err.Error(),
		})
		return nil, err
	}
	h.processingTracker.RecordStage("validation", "completed", 0, nil)

	// Update status to Turn1Started
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1Started, "turn1_start", map[string]interface{}{
		"function_name": "ExecuteTurn1Combined",
	})

	// STAGE 2: Load context (system prompt and images)
	contextLogger.Info("Loading context", map[string]interface{}{
		"system_prompt_key": req.S3Refs.Prompts.System.Key,
		"reference_image":   req.S3Refs.Images.ReferenceBase64.Key,
	})

	loadResult := h.contextLoader.LoadContext(ctx, req)
	if loadResult.Error != nil {
		h.handleContextLoadError(ctx, req.VerificationID, loadResult, contextLogger)
		return nil, loadResult.Error
	}
	h.recordContextLoadSuccess(ctx, req.VerificationID, loadResult)

	// STAGE 3: Generate prompt
	promptResult := h.generatePrompt(ctx, req, loadResult.SystemPrompt)
	if promptResult.Error != nil {
		h.handlePromptError(ctx, req.VerificationID, promptResult, contextLogger)
		return nil, promptResult.Error
	}

	h.processingTracker.RecordStage("prompt_generation", "completed", promptResult.Duration, map[string]interface{}{
		"prompt_length":    len(promptResult.Prompt),
		"template_used":    h.promptGenerator.GetTemplateUsed(req.VerificationContext),
		"template_version": h.cfg.Prompts.TemplateVersion,
	})

	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1PromptPrepared, "prompt_prepared", map[string]interface{}{
		"prompt_size":      len(promptResult.Prompt),
		"template_version": h.cfg.Prompts.TemplateVersion,
	})

	// STAGE 4: Invoke Bedrock
	invokeResult := h.bedrockInvoker.InvokeBedrock(ctx, loadResult.SystemPrompt, promptResult.Prompt, loadResult.Base64Image, req.VerificationID)
	if invokeResult.Error != nil {
		h.handleBedrockError(ctx, req.VerificationID, invokeResult)
		return nil, invokeResult.Error
	}
	h.recordBedrockSuccess(ctx, req.VerificationID, invokeResult, promptResult.TemplateProcessor)

	var bedrockTextOutput string
	if processedMap, ok := invokeResult.Response.Processed.(map[string]interface{}); ok {
		if contentStr, ok := processedMap["content"].(string); ok {
			bedrockTextOutput = contentStr
		}
	}
	if bedrockTextOutput == "" {
		var rawResp struct {
			Content string `json:"content"`
		}
		_ = json.Unmarshal(invokeResult.Response.Raw, &rawResp)
		bedrockTextOutput = rawResp.Content
	}
	parsedTurn1Data, parseErr := bedrockparser.ParseBedrockResponseAsMarkdown(bedrockTextOutput)
	if parseErr != nil {
		contextLogger.Warn("failed to parse bedrock response", map[string]interface{}{"error": parseErr.Error()})
	}

	// Build conversation history messages
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": []map[string]interface{}{{"type": "text", "text": loadResult.SystemPrompt}},
		},
		{
			"role": "user",
			"content": []map[string]interface{}{
				{"type": "text", "text": promptResult.Prompt},
				{"type": "image_base64", "text": loadResult.Base64Image},
			},
		},
		{
			"role":    "assistant",
			"content": []map[string]interface{}{{"type": "text", "text": bedrockTextOutput}},
		},
	}

	convRef, convErr := h.storageManager.StoreConversation(ctx, req.VerificationID, messages)
	if convErr != nil {
		return nil, convErr
	}

	// STAGE 5: Store response and processed analysis
	storageResult := h.storageManager.StoreResponses(ctx, req, invokeResult, promptResult, len(loadResult.Base64Image), parsedTurn1Data)
	if storageResult.Error != nil {
		contextLogger.Error("storage error", map[string]interface{}{
			"error": storageResult.Error.Error(),
		})
		return nil, storageResult.Error
	}
	h.recordStorageSuccess(storageResult)

	// STAGE 6: Store prompt
	promptStart := time.Now()
	promptRef, err := h.storageManager.StorePrompt(ctx, req, 1, promptResult)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store prompt", true).
			WithContext("verification_id", req.VerificationID)
	}

	// Record prompt storage success
	h.processingTracker.RecordStage("prompt_storage", "completed", time.Since(promptStart), map[string]interface{}{
		"prompt_length":  len(promptResult.Prompt),
		"prompt_ref_key": promptRef.Key,
	})

	// STAGE 7: Update metrics and async operations
	totalDuration := time.Since(h.startTime)
	h.updateProcessingMetrics(processingMetrics, totalDuration, invokeResult)

	// Prepare DynamoDB entries
	statusEntry := schema.StatusHistoryEntry{
		Status:           schema.StatusTurn1Completed,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn1Combined",
		ProcessingTimeMs: totalDuration.Milliseconds(),
		Stage:            "turn1_completion",
	}

	turnEntry := &schema.TurnResponse{
		TurnId:     1,
		Timestamp:  schema.FormatISO8601(),
		Prompt:     "",
		ImageUrls:  map[string]string{},
		Response:   schema.BedrockApiResponse{RequestId: invokeResult.Response.RequestID},
		LatencyMs:  invokeResult.Duration.Milliseconds(),
		TokenUsage: &invokeResult.Response.TokenUsage,
		Stage:      "REFERENCE_ANALYSIS",
		Metadata: map[string]interface{}{
			"model_id":        h.cfg.AWS.BedrockModel,
			"verification_id": req.VerificationID,
			"function_name":   "ExecuteTurn1Combined",
		},
	}

	var turn1MetricsForDB *schema.TurnMetrics = nil
	if processingMetrics != nil && processingMetrics.Turn1 != nil {
		turn1MetricsForDB = processingMetrics.Turn1
	}

	dynamoOK := h.dynamoManager.UpdateTurn1Completion(ctx, req.VerificationID, req.VerificationContext.VerificationAt, statusEntry, turnEntry, turn1MetricsForDB, &storageResult.ProcessedRef)

	// Final status update
	h.updateStatus(ctx, req.VerificationID, schema.StatusTurn1Completed, "completion", map[string]interface{}{
		"total_duration_ms": totalDuration.Milliseconds(),
		"processing_stages": h.processingTracker.GetStageCount(),
		"status_updates":    h.statusTracker.GetHistoryCount(),
		"dynamodb_updated":  dynamoOK,
	})

	// Build Step Function response
	stepFunctionResponse := h.responseBuilder.BuildStepFunctionResponse(
		req, promptRef, storageResult.RawRef, storageResult.ProcessedRef, convRef,
		invokeResult.Response, totalDuration.Milliseconds(), invokeResult.Duration.Milliseconds(), dynamoOK,
	)

	// Log completion
	contextLogger.Info("Completed ExecuteTurn1Combined", map[string]interface{}{
		"duration_ms":       totalDuration.Milliseconds(),
		"processing_stages": h.processingTracker.GetStageCount(),
		"status_updates":    h.statusTracker.GetHistoryCount(),
		"schema_version":    h.validator.GetSchemaVersion(),
		"verification_id":   req.VerificationID,
		"status":            schema.StatusTurn1Completed,
	})

	return stepFunctionResponse, nil
}

// HandleTurn1Combined is the Lambda entrypoint invoked by Step Functions.
func (h *Handler) HandleTurn1Combined(ctx context.Context, event json.RawMessage) (interface{}, error) {
	var stepFunctionEvent StepFunctionEvent

	if err := json.Unmarshal(event, &stepFunctionEvent); err == nil && stepFunctionEvent.SchemaVersion != "" {
		return h.handleStepFunctionEvent(ctx, stepFunctionEvent)
	}

	return h.handleDirectRequest(ctx, event)
}
