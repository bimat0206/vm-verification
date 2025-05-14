package schema

import (
	"fmt"
	"strings"
)

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
}

// Errors is a collection of validation errors
type Errors []ValidationError

// Error implements the error interface for a collection of errors
func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}

	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// ValidateVerificationContext validates the VerificationContext struct
func ValidateVerificationContext(ctx *VerificationContext) Errors {
	var errors Errors

	// Required fields
	if ctx.VerificationId == "" {
		errors = append(errors, ValidationError{Field: "verificationId", Message: "required field missing"})
	}
	if ctx.VerificationAt == "" {
		errors = append(errors, ValidationError{Field: "verificationAt", Message: "required field missing"})
	}
	if ctx.Status == "" {
		errors = append(errors, ValidationError{Field: "status", Message: "required field missing"})
	} else {
		// Validate status is a known value
		validStatus := false
		for _, s := range []string{
			StatusVerificationRequested, StatusVerificationInitialized, StatusFetchingImages,
			StatusImagesFetched, StatusPromptPrepared, StatusTurn1PromptReady, StatusTurn1Completed,
			StatusTurn1Processed, StatusTurn2PromptReady, StatusTurn2Completed, StatusTurn2Processed,
			StatusResultsFinalized, StatusResultsStored, StatusNotificationSent, StatusCompleted,
			StatusInitializationFailed, StatusHistoricalFetchFailed, StatusImageFetchFailed,
			StatusBedrockProcessingFailed, StatusVerificationFailed,
		} {
			if ctx.Status == s {
				validStatus = true
				break
			}
		}
		if !validStatus {
			errors = append(errors, ValidationError{Field: "status", Message: "invalid status value"})
		}
	}
	if ctx.VerificationType == "" {
		errors = append(errors, ValidationError{Field: "verificationType", Message: "required field missing"})
	} else {
		// Validate verification type
		if ctx.VerificationType != VerificationTypeLayoutVsChecking && 
		   ctx.VerificationType != VerificationTypePreviousVsCurrent {
			errors = append(errors, ValidationError{
				Field:   "verificationType",
				Message: "must be either LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT",
			})
		}
	}
	if ctx.ReferenceImageUrl == "" {
		errors = append(errors, ValidationError{Field: "referenceImageUrl", Message: "required field missing"})
	}
	if ctx.CheckingImageUrl == "" {
		errors = append(errors, ValidationError{Field: "checkingImageUrl", Message: "required field missing"})
	}
	if ctx.VendingMachineId == "" {
		errors = append(errors, ValidationError{Field: "vendingMachineId", Message: "required field missing"})
	}

	// Verification type-specific validations
	if ctx.VerificationType == VerificationTypeLayoutVsChecking {
		if ctx.LayoutId == 0 {
			errors = append(errors, ValidationError{
				Field:   "layoutId", 
				Message: "required for LAYOUT_VS_CHECKING verification type",
			})
		}
		if ctx.LayoutPrefix == "" {
			errors = append(errors, ValidationError{
				Field:   "layoutPrefix",
				Message: "required for LAYOUT_VS_CHECKING verification type",
			})
		}
	}

	return errors
}

// ValidateWorkflowState validates the complete workflow state
func ValidateWorkflowState(state *WorkflowState) Errors {
	var errors Errors

	// Check schema version
	if state.SchemaVersion == "" {
		state.SchemaVersion = SchemaVersion
	} else if state.SchemaVersion != SchemaVersion {
		errors = append(errors, ValidationError{
			Field:   "schemaVersion", 
			Message: fmt.Sprintf("unsupported schema version: %s (supported: %s)", 
				state.SchemaVersion, SchemaVersion),
		})
	}

	// Always validate verification context
	if state.VerificationContext == nil {
		errors = append(errors, ValidationError{Field: "verificationContext", Message: "required field missing"})
	} else {
		ctxErrors := ValidateVerificationContext(state.VerificationContext)
		errors = append(errors, ctxErrors...)
	}

	// Add additional validations for other fields based on the current state
	// For example, if in TURN1_COMPLETED, validate Turn1Response exists
	if state.VerificationContext != nil {
		status := state.VerificationContext.Status
		switch status {
		case StatusTurn1Completed, StatusTurn1Processed:
			if state.Turn1Response == nil || len(state.Turn1Response) == 0 {
				errors = append(errors, ValidationError{
					Field:   "turn1Response",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		case StatusTurn2Completed, StatusTurn2Processed:
			if state.Turn2Response == nil || len(state.Turn2Response) == 0 {
				errors = append(errors, ValidationError{
					Field:   "turn2Response",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		case StatusResultsFinalized, StatusResultsStored, StatusCompleted:
			if state.FinalResults == nil {
				errors = append(errors, ValidationError{
					Field:   "finalResults",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		}
	}

	return errors
}