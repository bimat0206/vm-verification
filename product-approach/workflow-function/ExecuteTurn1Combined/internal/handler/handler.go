// internal/handler/handler.go - CORRECTED VERSION
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
)

// Handler orchestrates the ExecuteTurn1Combined workflow.
type Handler struct {
	cfg           config.Config
	s3            services.S3StateManager
	bedrock       services.BedrockService
	dynamo        services.DynamoDBService
	promptService services.PromptService
	validator     *validation.SchemaValidator
	// CORRECT: Logger is an interface, not a pointer
	log           logger.Logger
}

// NewHandler wires together all dependencies for the Lambda.
func NewHandler(
	s3Mgr services.S3StateManager,
	bedrockClient services.BedrockService,
	dynamoClient services.DynamoDBService,
	promptGen services.PromptService,
	// CORRECT: Interface parameter, not pointer
	log logger.Logger,
	cfg *config.Config,
) (*Handler, error) {
	return &Handler{
		cfg:           *cfg,
		s3:            s3Mgr,
		bedrock:       bedrockClient,
		dynamo:        dynamoClient,
		promptService: promptGen,
		validator:     validation.NewSchemaValidator(),
		log:           log,
	}, nil
}

// Handle executes a single Turn-1 verification cycle.
func (h *Handler) Handle(ctx context.Context, req *models.Turn1Request) (*models.Turn1Response, error) {
	start := time.Now()
	
	// Schema validation using standardized validation
	if err := h.validator.ValidateRequest(req); err != nil {
		return nil, errors.NewValidationError("request validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	
	// CORRECT: Using fluent interface properly - both methods return Logger interface
	contextLogger := h.log.WithCorrelationId(req.VerificationID).
		WithFields(map[string]interface{}{
			"verificationId": req.VerificationID,
			"turnId":         1,
			"schemaVersion":  h.validator.GetSchemaVersion(),
		})
	
	contextLogger.Info("Starting ExecuteTurn1Combined", map[string]interface{}{
		"verification_type": req.VerificationContext.VerificationType,
		"s3_system_prompt":  req.S3Refs.Prompts.System.Key,
		"s3_reference_img":  req.S3Refs.Images.ReferenceBase64.Key,
	})

	// --- 1) Concurrently load system prompt & base64 image from S3 ---
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
			// CORRECT: Using WrapError helper function and error modification functions
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3, 
				"failed to load system prompt", true) // true = retryable
			
			// CORRECT: Using helper functions to enrich error context
			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading")
			
			// CORRECT: Using helper function, not method
			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}
		systemPrompt = sp
	}()

	go func() {
		defer wg.Done()
		img, err := h.s3.LoadBase64Image(ctx, req.S3Refs.Images.ReferenceBase64)
		if err != nil {
			// CORRECT: Chaining context methods properly
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
	if loadErr != nil {
		// CORRECT: Checking for WorkflowError type to access properties
		if workflowErr, ok := loadErr.(*errors.WorkflowError); ok {
			contextLogger.Error("resource loading error", map[string]interface{}{
				"error_type":    string(workflowErr.Type),
				"error_code":    workflowErr.Code,
				"retryable":     workflowErr.Retryable,
				"severity":      string(workflowErr.Severity),
				"s3_operations": 2,
			})
		} else {
			contextLogger.Error("resource loading error", map[string]interface{}{
				"error":         loadErr.Error(),
				"s3_operations": 2,
			})
		}
		
		return nil, loadErr
	}

	// --- 2) Generate Turn-1 prompt ---
	turnPrompt, err := h.promptService.GenerateTurn1Prompt(ctx, req.VerificationContext, systemPrompt)
	if err != nil {
		// CORRECT: Creating errors with factory functions
		promptErr := errors.NewInternalError("prompt_service", err).
			WithContext("template_version", h.cfg.Prompts.TemplateVersion).
			WithContext("verification_type", req.VerificationContext.VerificationType)
		
		// CORRECT: Using helper function for verification ID
		enrichedErr := errors.SetVerificationID(promptErr, req.VerificationID)
		
		contextLogger.Error("prompt generation error", map[string]interface{}{
			"template_version":    h.cfg.Prompts.TemplateVersion,
			"verification_type":   req.VerificationContext.VerificationType,
			"system_prompt_size":  len(systemPrompt),
		})
		
		return nil, enrichedErr
	}

	// --- 3) Invoke Bedrock Converse API ---
	resp, err := h.bedrock.Converse(ctx, systemPrompt, turnPrompt, base64Img)
	if err != nil {
		// CORRECT: Proper error type checking and enrichment
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			// Error is already a WorkflowError from the service layer
			enrichedErr := workflowErr.WithContext("model_id", h.cfg.AWS.BedrockModel).
				WithContext("max_tokens", h.cfg.Processing.MaxTokens).
				WithContext("prompt_size", len(turnPrompt)).
				WithContext("image_size", len(base64Img))
			
			// CORRECT: Using helper function
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
			// CORRECT: Wrapping unexpected error types
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

	// --- 4) Persist raw & processed responses to S3 ---
	rawRef, err := h.s3.StoreRawResponse(ctx, req.VerificationID, resp.Raw)
	if err != nil {
		// CORRECT: S3 storage errors are typically retryable
		s3Err := errors.WrapError(err, errors.ErrorTypeS3, 
			"store raw response failed", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("response_size", len(resp.Raw))
		
		// CORRECT: Using helper function
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

	// --- 5) DynamoDB updates (fire & forget) ---
	go func() {
		if err := h.dynamo.UpdateVerificationStatus(ctx, req.VerificationID, models.StatusTurn1Completed, resp.TokenUsage); err != nil {
			// CORRECT: Create scoped logger for async operations
			asyncLogger := contextLogger.WithFields(map[string]interface{}{
				"async_operation": "verification_status_update",
				"table":           h.cfg.AWS.DynamoDBVerificationTable,
			})
			
			if workflowErr, ok := err.(*errors.WorkflowError); ok {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error_type": string(workflowErr.Type),
					"error_code": workflowErr.Code,
					"retryable":  workflowErr.Retryable,
				})
			} else {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}()

	// --- 6) Build and return response envelope ---
	summary := models.Summary{
		AnalysisStage:    models.StageReferenceAnalysis,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
		TokenUsage:       resp.TokenUsage,
		BedrockRequestID: resp.RequestID,
	}
	
	output := &models.Turn1Response{
		S3Refs: models.Turn1ResponseS3Refs{
			RawResponse:       rawRef,
			ProcessedResponse: procRef,
		},
		Status:  models.StatusTurn1Completed,
		Summary: summary,
	}

	// Validate response before returning using schema validation
	if err := h.validator.ValidateResponse(output); err != nil {
		contextLogger.Error("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return nil, errors.NewValidationError("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	// Validate token usage separately for detailed reporting
	if err := h.validator.ValidateTokenUsage(&resp.TokenUsage); err != nil {
		contextLogger.Warn("token usage validation warning", map[string]interface{}{
			"token_validation_error": err.Error(),
		})
	}

	contextLogger.Info("Completed ExecuteTurn1Combined", map[string]interface{}{
		"duration_ms":        summary.ProcessingTimeMs,
		"input_tokens":       resp.TokenUsage.InputTokens,
		"output_tokens":      resp.TokenUsage.OutputTokens,
		"total_tokens":       resp.TokenUsage.TotalTokens,
		"bedrock_request_id": resp.RequestID,
		"s3_objects_created": 2,
		"dynamo_updates":     2,
		"status":             string(models.StatusTurn1Completed),
		"schema_version":     h.validator.GetSchemaVersion(),
	})
	
	return output, nil
}

// HandleTurn1Combined is the Lambda entrypoint invoked by Step Functions.
func (h *Handler) HandleTurn1Combined(ctx context.Context, event json.RawMessage) (*models.Turn1Response, error) {
	var req models.Turn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		// CORRECT: Using validation error factory function
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
	
	// CORRECT: Using shared logger's event logging capability
	h.log.LogReceivedEvent(req)
	
	return h.Handle(ctx, &req)
}