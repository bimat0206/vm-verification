// Package handler provides the entry point for the ProcessTurn1Response Lambda function
package handler

import (
	"context"
	"fmt"
	"time"
	
	"workflow-function/ProcessTurn1Response/internal/errors"
	"workflow-function/ProcessTurn1Response/internal/state"
	"workflow-function/shared/schema"
)

// Operation names for error reporting
const (
	OpHandleRequest   = "Handle"
	OpValidateInput   = "ValidateInput"
	OpProcessResponse = "ProcessResponse"
	OpPrepareOutput   = "PrepareOutput"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	SchemaVersion string            `json:"schemaVersion"`
	Status        string            `json:"status"`
	Error         *schema.ErrorInfo `json:"error,omitempty"`
	RequestID     string            `json:"requestId,omitempty"`
	Timestamp     string            `json:"timestamp"`
}

// handleError creates a standardized error response and logs the error
func (h *Handler) handleError(ctx context.Context, state *state.WorkflowState, errorCategory ErrorCategory, message string, err error) (interface{}, error) {
	// Map our internal error category to schema error code
	errorCode := mapErrorCategoryToCode(errorCategory)
	
	// Create detailed error info
	errorInfo := &schema.ErrorInfo{
		Code:      errorCode,
		Message:   fmt.Sprintf("%s: %v", message, err),
		Timestamp: time.Now().Format(time.RFC3339),
		Details: map[string]interface{}{
			"function": "ProcessTurn1Response",
			"stage":    "Turn1Processing",
		},
	}
	
	// Add correlation information if available
	verificationID := ""
	if state != nil {
		verificationID = state.VerificationID
		errorInfo.Details["verificationId"] = verificationID
	}
	
	// Convert standard error to FunctionError for proper logging
	var funcError error
	if fe, ok := err.(*errors.FunctionError); ok {
		funcError = fe
	} else {
		funcError = errors.NewWithDetails(
			OpHandleRequest,
			mapErrorCategoryToFunctionCategory(errorCategory),
			errorCode,
			message,
			errors.SeverityError,
			errorInfo.Details,
			err,
		)
	}
	
	// Log the error appropriately
	h.logError(funcError, errorInfo)
	
	// If we have a state, update it with the error
	if state != nil {
		state.Status = schema.StatusVerificationFailed
		state.Metadata = map[string]interface{}{
			"error": errorInfo,
		}
		
		// Try to update the state in S3, but don't fail if this fails
		updateErr := h.stateManager.UpdateWorkflowState(context.Background(), state, schema.StatusVerificationFailed)
		if updateErr != nil {
			h.logger.Error("Failed to update error state in S3", 
				"error", updateErr,
				"verificationId", verificationID)
		}
		
		// Return envelope with error information
		envelope := h.stateManager.GetEnvelopeFromState(state)
		if envelope != nil {
			envelope.SetStatus(schema.StatusVerificationFailed)
			envelope.AddSummary("error", errorInfo)
			return envelope, nil
		}
	}
	
	// If we don't have a state, create a minimal error response
	return map[string]interface{}{
		"status": schema.StatusVerificationFailed,
		"error":  errorInfo,
	}, nil
}

// logError logs an error with appropriate contextual information
func (h *Handler) logError(err error, errorInfo *schema.ErrorInfo) {
	// Extract attributes for structured logging
	attrs := []interface{}{
		"errorCode", errorInfo.Code,
		"errorMessage", errorInfo.Message,
	}
	
	// Add details to attributes
	for k, v := range errorInfo.Details {
		attrs = append(attrs, k, v)
	}
	
	// Log at appropriate level based on error category
	var fe *errors.FunctionError
	if errors.As(err, &fe) {
		errors.LogError(h.logger, err)
	} else {
		// Default to error level for unknown error types
		h.logger.Error(errorInfo.Message, attrs...)
	}
}

// mapErrorCategoryToCode maps internal error categories to schema error codes
func mapErrorCategoryToCode(category ErrorCategory) string {
	switch category {
	case ErrorCategoryInput:
		return "INPUT_VALIDATION_ERROR"
	case ErrorCategoryState:
		return "STATE_MANAGEMENT_ERROR"
	case ErrorCategoryProcessing:
		return "PROCESSING_ERROR"
	case ErrorCategoryStorage:
		return "STORAGE_ERROR"
	default:
		return "UNEXPECTED_ERROR"
	}
}

// mapErrorCategoryToFunctionCategory maps handler error categories to function error categories
func mapErrorCategoryToFunctionCategory(category ErrorCategory) errors.ErrorCategory {
	switch category {
	case ErrorCategoryInput:
		return errors.CategoryInput
	case ErrorCategoryState:
		return errors.CategoryState
	case ErrorCategoryProcessing:
		return errors.CategoryProcess
	case ErrorCategoryStorage:
		return errors.CategoryState
	default:
		return errors.CategorySystem
	}
}