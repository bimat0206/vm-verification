// Package state provides S3 state management for the ProcessTurn1Response Lambda
package state

import (
	"log/slog"
	
	"workflow-function/ProcessTurn1Response/internal/errors"
	"workflow-function/shared/s3state"
)

// Operation names for error reporting
const (
	OpLoadWorkflowState     = "LoadWorkflowState"
	OpLoadTurn1Response     = "LoadTurn1Response"
	OpStoreReferenceAnalysis = "StoreReferenceAnalysis"
	OpUpdateWorkflowState   = "UpdateWorkflowState"
	OpExtractHistoricalData = "ExtractHistoricalData"
	OpDetermineProcessingPath = "DetermineProcessingPath"
	OpBuildTurn2Context     = "BuildTurn2Context"
	OpStoreProcessingResult = "StoreProcessingResult"
)

// handleStateError converts and logs state management errors
func handleStateError(logger *slog.Logger, operation string, err error, verificationID string) error {
	if err == nil {
		return nil
	}
	
	// Convert to our internal error type
	funcErr := convertStateError(operation, err, verificationID)
	
	// Log the error with appropriate severity
	errors.LogError(logger, funcErr)
	
	return funcErr
}

// convertStateError converts state errors to function errors
func convertStateError(operation string, err error, verificationID string) error {
	// First check if it's already a FunctionError
	var fe *errors.FunctionError
	if errors.As(err, &fe) {
		// Add verificationID to details if it's not already there
		if verificationID != "" && fe.Details["verificationId"] == nil {
			fe.Details["verificationId"] = verificationID
		}
		return fe
	}
	
	// Add workflow context
	details := map[string]interface{}{}
	if verificationID != "" {
		details["verificationId"] = verificationID
	}
	
	// Check for specific error types
	switch {
	case s3state.IsValidationError(err):
		return errors.NewWithDetails(
			operation,
			errors.CategoryInput,
			errors.ErrValidationFailed,
			err.Error(),
			errors.SeverityWarning,
			details,
			err,
		)
		
	case s3state.IsReferenceError(err):
		return errors.NewWithDetails(
			operation,
			errors.CategoryState,
			errors.ErrReferenceInvalid,
			err.Error(),
			errors.SeverityError,
			details,
			err,
		)
		
	case s3state.IsS3Error(err):
		return errors.NewWithDetails(
			operation,
			errors.CategoryState,
			errors.ErrStateStoreFailed,
			err.Error(),
			errors.SeverityError,
			details,
			err,
		)
		
	default:
		return errors.NewWithDetails(
			operation,
			errors.CategoryState,
			errors.ErrStateStoreFailed,
			err.Error(),
			errors.SeverityError,
			details,
			err,
		)
	}
}

// validateReference validates a reference and returns a standardized error
func validateReference(ref *s3state.Reference, operation string) error {
	if ref == nil {
		return errors.ReferenceError(operation, "reference is nil")
	}
	
	if ref.Bucket == "" {
		return errors.ReferenceError(operation, "reference has empty bucket")
	}
	
	if ref.Key == "" {
		return errors.ReferenceError(operation, "reference has empty key")
	}
	
	return nil
}