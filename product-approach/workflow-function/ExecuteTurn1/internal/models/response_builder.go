package models

import (
	"fmt"
	"time"
	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// ResponseBuilder handles the construction of ExecuteTurn1Response objects
type ResponseBuilder struct {
	logger logger.Logger
}

// NewResponseBuilder creates a new response builder instance
func NewResponseBuilder(log logger.Logger) *ResponseBuilder {
	return &ResponseBuilder{
		logger: log.WithFields(map[string]interface{}{
			"component": "ResponseBuilder",
		}),
	}
}

// BuildSuccessResponse creates a successful response wrapper for Step Functions
func (rb *ResponseBuilder) BuildSuccessResponse(state *schema.WorkflowState) *ExecuteTurn1Response {
	// Safe extraction of verification ID for logging
	verificationId := ""
	status := ""
	if state != nil && state.VerificationContext != nil {
		verificationId = state.VerificationContext.VerificationId
		status = state.VerificationContext.Status
	}

	rb.logger.Debug("Building successful response", map[string]interface{}{
		"verificationId": verificationId,
		"status":         status,
	})

	// Ensure we have a valid state to return
	if state == nil {
		rb.logger.Warn("Received nil state for success response, creating minimal state", nil)
		state = &schema.WorkflowState{
			SchemaVersion: schema.SchemaVersion,
			VerificationContext: &schema.VerificationContext{
				Status:         "UNKNOWN",
				VerificationAt: time.Now().UTC().Format(time.RFC3339),
			},
		}
	}

	return &ExecuteTurn1Response{
		WorkflowState: *state,
		Error:         nil,
	}
}

// BuildErrorResponse creates an error response that attaches the WorkflowError to the state
func (rb *ResponseBuilder) BuildErrorResponse(
	state *schema.WorkflowState,
	wfErr *wferrors.WorkflowError,
	status string,
) *ExecuteTurn1Response {
	// Ensure we have a WorkflowError
	if wfErr == nil {
		rb.logger.Error("Received nil error for error response, creating default error", nil)
		wfErr = wferrors.NewInternalError("UnknownError", fmt.Errorf("nil error provided"))
	}

	// Safe extraction of verification ID for logging
	verificationId := ""
	if state != nil && state.VerificationContext != nil {
		verificationId = state.VerificationContext.VerificationId
	}

	rb.logger.Error("Building error response", map[string]interface{}{
		"error":          wfErr.Error(),
		"errorCode":      wfErr.Code,
		"retryable":      wfErr.Retryable,
		"status":         status,
		"verificationId": verificationId,
	})

	// Ensure we have a valid state to update
	if state == nil {
		rb.logger.Warn("Received nil state for error response, creating minimal state", nil)
		state = &schema.WorkflowState{
			SchemaVersion: schema.SchemaVersion,
			VerificationContext: &schema.VerificationContext{
				VerificationId: "unknown",
				Status:         status,
				VerificationAt: time.Now().UTC().Format(time.RFC3339),
			},
		}
	}

	// Update verification context with error information
	if state.VerificationContext != nil {
		state.VerificationContext.Status = status
		state.VerificationContext.Error = &schema.ErrorInfo{
			Code:      wfErr.Code,
			Message:   wfErr.Message,
			Details:   wfErr.Context,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		state.VerificationContext.VerificationAt = time.Now().UTC().Format(time.RFC3339)
	} else {
		// Create verification context if it doesn't exist
		rb.logger.Warn("Creating VerificationContext for error response", nil)
		state.VerificationContext = &schema.VerificationContext{
			VerificationId: "unknown",
			Status:         status,
			VerificationAt: time.Now().UTC().Format(time.RFC3339),
			Error: &schema.ErrorInfo{
				Code:      wfErr.Code,
				Message:   wfErr.Message,
				Details:   wfErr.Context,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		}
	}

	return &ExecuteTurn1Response{
		WorkflowState: *state,
		Error:         wfErr,
	}
}

// BuildErrorResponseWithoutState creates an error response when no state is available
// This is useful for early errors during parsing or initialization
func (rb *ResponseBuilder) BuildErrorResponseWithoutState(
	wfErr *wferrors.WorkflowError,
	status string,
) *ExecuteTurn1Response {
	if wfErr == nil {
		rb.logger.Error("Received nil error for error response without state", nil)
		wfErr = wferrors.NewInternalError("UnknownError", fmt.Errorf("nil error provided"))
	}

	rb.logger.Error("Building error response without state", map[string]interface{}{
		"error":     wfErr.Error(),
		"errorCode": wfErr.Code,
		"retryable": wfErr.Retryable,
		"status":    status,
	})

	// Create minimal workflow state for the error response
	minimalState := &schema.WorkflowState{
		SchemaVersion: schema.SchemaVersion,
		VerificationContext: &schema.VerificationContext{
			VerificationId: "unknown",
			Status:         status,
			VerificationAt: time.Now().UTC().Format(time.RFC3339),
			Error: &schema.ErrorInfo{
				Code:      wfErr.Code,
				Message:   wfErr.Message,
				Details:   wfErr.Context,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	return &ExecuteTurn1Response{
		WorkflowState: *minimalState,
		Error:         wfErr,
	}
}

// EnsureWorkflowError converts any error to a WorkflowError
// This is a utility method to ensure consistent error handling
func (rb *ResponseBuilder) EnsureWorkflowError(err error) *wferrors.WorkflowError {
	if err == nil {
		return nil
	}

	// If already a WorkflowError, return as-is
	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
		return wfErr
	}

	// Convert to WorkflowError
	rb.logger.Debug("Converting error to WorkflowError", map[string]interface{}{
		"originalError": err.Error(),
		"errorType":     fmt.Sprintf("%T", err),
	})

	return wferrors.WrapError(err, wferrors.ErrorTypeInternal, "unexpected error", false)
}

// Backward compatibility functions for existing code

// NewResponse creates a successful response wrapper for Step Functions
func NewResponse(state *schema.WorkflowState, log logger.Logger) *ExecuteTurn1Response {
	builder := NewResponseBuilder(log)
	return builder.BuildSuccessResponse(state)
}

// NewErrorResponse creates an error response that attaches the WorkflowError to the state
func NewErrorResponse(
	state *schema.WorkflowState,
	wfErr *wferrors.WorkflowError,
	log logger.Logger,
	status string,
) *ExecuteTurn1Response {
	builder := NewResponseBuilder(log)
	return builder.BuildErrorResponse(state, wfErr, status)
}