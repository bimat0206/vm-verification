package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
	
	"workflow-function/ExecuteTurn1/internal"
	"workflow-function/ExecuteTurn1/internal/bedrock"
	"workflow-function/ExecuteTurn1/internal/state"
	"workflow-function/ExecuteTurn1/internal/validation"
)

// Handler provides the core logic for ExecuteTurn1
type Handler struct {
	stateLoader    *state.Loader
	stateSaver     *state.Saver
	bedrockClient  *bedrock.Client
	validator      *validation.Validator
	s3Client       *s3.Client
	hybridConfig   *internal.HybridStorageConfig
	logger         logger.Logger
}

// NewHandler constructs the ExecuteTurn1 handler with injected dependencies
func NewHandler(
	stateLoader *state.Loader,
	stateSaver *state.Saver,
	bedrockClient *bedrock.Client,
	validator *validation.Validator,
	s3Client *s3.Client,
	hybridConfig *internal.HybridStorageConfig,
	logger logger.Logger,
) *Handler {
	return &Handler{
		stateLoader:    stateLoader,
		stateSaver:     stateSaver,
		bedrockClient:  bedrockClient,
		validator:      validator,
		s3Client:       s3Client,
		hybridConfig:   hybridConfig,
		logger:         logger,
	}
}

// HandleRequest processes the ExecuteTurn1 request using S3 state references
func (h *Handler) HandleRequest(ctx context.Context, input *internal.StepFunctionInput) (*internal.StepFunctionOutput, error) {
	if input == nil {
		return nil, wferrors.NewValidationError("Input is nil", nil)
	}
	
	// Check if we have either type of references
	hasStateRefs := input.StateReferences != nil
	hasS3Refs := input.S3References != nil && len(input.S3References) > 0
	
	if !hasStateRefs && !hasS3Refs {
		return nil, wferrors.NewValidationError("Neither StateReferences nor S3References is provided", nil)
	}
	
	// Map the S3References to StateReferences format using the new method
	if input.StateReferences == nil {
		input.StateReferences = input.MapS3References()
	}
	
	// Validate that we have a proper StateReferences object after mapping
	if input.StateReferences == nil {
		return nil, wferrors.NewValidationError("Failed to map references from input", nil)
	}
	
	// Check if VerificationId is provided and consistent
	verificationId := input.StateReferences.VerificationId
	if verificationId == "" {
		// Try to get from direct input if not in StateReferences
		verificationId = input.VerificationId
		if verificationId == "" {
			return nil, wferrors.NewValidationError("VerificationId is required", nil)
		}
		// Set it in the StateReferences for consistency
		input.StateReferences.VerificationId = verificationId
	}

	// Set up context-aware logger
	requestID := ""
	if reqID, ok := ctx.Value("requestID").(string); ok {
		requestID = reqID
	} else {
		// Use a default request ID if not provided in context
		requestID = "req-" + time.Now().Format("20060102-150405")
	}
	
	log := h.logger.WithCorrelationId(requestID).WithFields(map[string]interface{}{
		"verificationId": verificationId,
		"step":           "ExecuteTurn1",
	})

	log.Info("Starting ExecuteTurn1 with state references", map[string]interface{}{
		"verificationId": verificationId,
		"datePartition":  input.StateReferences.DatePartition,
	})

	// Step 1: Validate state references
	if err := h.validator.ValidateStateReferences(input.StateReferences); err != nil {
		log.Error("State references validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return h.handleError(err, "state_references_validation_failed", log)
	}

	// Step 2: Load state from S3 references
	state, err := h.stateLoader.LoadWorkflowState(ctx, input.StateReferences)
	if err != nil {
		log.Error("Failed to load workflow state", map[string]interface{}{
			"error": err.Error(),
		})
		return h.handleError(err, "state_loading_failed", log)
	}

	// Step 3: Validate complete workflow state
	if err := h.validator.ValidateWorkflowState(state); err != nil {
		log.Error("Workflow state validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return h.handleError(err, "state_validation_failed", log)
	}

	// Step 4: Handle specific use case data based on verification type
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		log.Info("Processing UC1: Layout vs Checking verification", nil)
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		log.Info("Processing UC2: Previous vs Current verification", nil)
	}

	// Step 5: Generate Base64 for images using hybrid storage
	if state.Images != nil {
		if err := h.processImages(ctx, state.Images, log); err != nil {
			log.Error("Image processing failed", map[string]interface{}{
				"error": err.Error(),
			})
			return h.handleError(err, "image_processing_failed", log)
		}

		// Validate images after Base64 generation
		if err := h.validator.ValidateImageData(state.Images, true); err != nil {
			log.Error("Image validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return h.handleError(err, "image_validation_failed", log)
		}
	}

	// Step 6: Process with Bedrock
	turn1Response, err := h.bedrockClient.ProcessTurn1(ctx, state.CurrentPrompt, state.Images)
	if err != nil {
		log.Error("Bedrock processing failed", map[string]interface{}{
			"error": err.Error(),
		})
		return h.handleError(err, "bedrock_processing_failed", log)
	}

	// Validate turn response
	if err := h.validator.ValidateTurnResponse(turn1Response); err != nil {
		log.Error("Turn response validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue despite validation errors
		log.Warn("Continuing despite turn response validation issues", nil)
	}

	// Step 7: Update workflow state with results
	h.updateWorkflowState(state, turn1Response, log)

	// Step 8: Save results to S3 using the correct folder structure
	updatedRefs, err := h.stateSaver.SaveWorkflowState(ctx, state)
	if err != nil {
		log.Error("Failed to save workflow state", map[string]interface{}{
			"error": err.Error(),
		})
		return h.handleError(err, "state_saving_failed", log)
	}

	// Save thinking content separately if available
	thinkingContent := ""
	if turn1Response.Metadata != nil && turn1Response.Metadata["thinking"] != nil {
		thinkingContent, _ = turn1Response.Metadata["thinking"].(string)
		if thinkingContent != "" {
			thinkingRef, err := h.stateSaver.SaveThinkingContent(ctx, state.VerificationContext.VerificationId, thinkingContent)
			if err != nil {
				log.Warn("Failed to save thinking content", map[string]interface{}{
					"error": err.Error(),
				})
				// Continue despite thinking save failure
			} else if thinkingRef != nil {
				// Add thinking reference
				updatedRefs.Turn1Thinking = thinkingRef
			}
		}
	}

	log.Info("ExecuteTurn1 completed successfully", map[string]interface{}{
		"latencyMs":      turn1Response.LatencyMs,
		"tokenUsage":     turn1Response.TokenUsage.TotalTokens,
		"status":         state.VerificationContext.Status,
		"hasThinking":    thinkingContent != "",
		"datePartition":  updatedRefs.DatePartition,
	})

	// Step 9: Return lightweight references with correct organization
	output := &internal.StepFunctionOutput{
		StateReferences: updatedRefs,
		Status:          schema.StatusTurn1Completed,
		Summary: map[string]interface{}{
			"tokenUsage":    turn1Response.TokenUsage.TotalTokens,
			"latencyMs":     turn1Response.LatencyMs,
			"status":        schema.StatusTurn1Completed,
			"datePartition": updatedRefs.DatePartition,
		},
	}
	
	// Ensure S3References is populated for step function compatibility
	output.S3References = output.StateReferences
	
	return output, nil
}

// processImages handles image Base64 generation using hybrid storage
// processImages handles image Base64 generation using hybrid storage
func (h *Handler) processImages(ctx context.Context, images *schema.ImageData, log logger.Logger) error {
    if images == nil {
        return nil
    }

    start := time.Now() // Add this line for the duration calculation below
    
    log.Info("Processing images with hybrid storage", map[string]interface{}{
        "hybridEnabled": false,
    })
    
    // Ensure base64 data is accessible for reference image - this is critical for Turn1
    if images.GetReference() != nil && !images.GetReference().HasBase64Data() {
        err := fmt.Errorf("reference image missing base64 data")
        log.Error("Failed to access Base64 for images", map[string]interface{}{
            "error": err.Error(),
        })
        return wferrors.WrapError(err, wferrors.ErrorTypeInternal, "Base64 access failed", false)
    }
    
    // For checking image, only log a warning since it's not needed in Turn1
    if images.GetChecking() != nil && !images.GetChecking().HasBase64Data() {
        log.Warn("Checking image missing Base64 data, but continuing since it's not needed for Turn1", map[string]interface{}{
            "checkingImageExists": images.GetChecking() != nil,
        })
        // No error return here - we continue processing
    }
    
    // Mark as processed
    images.Base64Generated = true
    images.ProcessedAt = schema.FormatISO8601()

    log.Info("Images processed successfully", map[string]interface{}{
        "durationMs": time.Since(start).Milliseconds(),
    })
    return nil
}

// updateWorkflowState updates the workflow state with Turn 1 results
func (h *Handler) updateWorkflowState(state *schema.WorkflowState, turnResponse *schema.TurnResponse, log logger.Logger) {
	log.Debug("Updating workflow state with results", map[string]interface{}{
		"turnId": turnResponse.TurnId,
	})

	// Update Turn1Response
	state.Turn1Response = map[string]interface{}{"turnResponse": turnResponse}

	// Update verification context
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.VerificationContext.VerificationAt = schema.FormatISO8601()
	state.VerificationContext.Error = nil

	// Update conversation state
	if state.ConversationState == nil {
		state.ConversationState = &schema.ConversationState{
			CurrentTurn: turnResponse.TurnId,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}
	
	// Add turn to history (avoiding duplicates)
	alreadyInHistory := false
	for _, item := range state.ConversationState.History {
		if turn, ok := item.(schema.TurnResponse); ok && turn.TurnId == turnResponse.TurnId {
			alreadyInHistory = true
			break
		}
	}
	
	if !alreadyInHistory {
		state.ConversationState.History = append(state.ConversationState.History, *turnResponse)
	}
	
	state.ConversationState.CurrentTurn = turnResponse.TurnId

	log.Debug("Workflow state updated successfully", map[string]interface{}{
		"status":      state.VerificationContext.Status,
		"currentTurn": state.ConversationState.CurrentTurn,
		"historySize": len(state.ConversationState.History),
	})
}

// handleError creates a consistent error response with the proper status
func (h *Handler) handleError(err error, code string, log logger.Logger) (*internal.StepFunctionOutput, error) {
	// Ensure we have a WorkflowError
	var wfErr *wferrors.WorkflowError
	if e, ok := err.(*wferrors.WorkflowError); ok {
		wfErr = e
	} else {
		wfErr = wferrors.WrapError(err, wferrors.ErrorTypeInternal, fmt.Sprintf("unexpected %s error", code), false)
	}

	// Log the error with full context
	log.Error("ExecuteTurn1 error", map[string]interface{}{
		"errorType":  wfErr.Type,
		"errorCode":  wfErr.Code,
		"retryable":  wfErr.Retryable,
		"errorMsg":   wfErr.Message,
		"context":    wfErr.Context,
	})

	// Return an error response with minimal details
	output := &internal.StepFunctionOutput{
		Status: schema.StatusBedrockProcessingFailed,
		Error: &schema.ErrorInfo{
			Code:      wfErr.Code,
			Message:   wfErr.Message,
			Timestamp: schema.FormatISO8601(),
			Details:   wfErr.Context,
		},
		Summary: map[string]interface{}{
			"error":     wfErr.Message,
			"status":    schema.StatusBedrockProcessingFailed,
			"retryable": wfErr.Retryable,
		},
	}
	
	// Ensure S3References is populated for step function compatibility
	if output.StateReferences != nil {
		output.S3References = output.StateReferences
	}
	
	return output, nil
}