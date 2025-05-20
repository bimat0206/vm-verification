// Package processor provides Turn1 processing capabilities
package processor

import (
	"fmt"
	"log/slog"
	
	"workflow-function/ProcessTurn1Response/internal/errors"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// Operation names for error reporting
const (
	OpProcessTurn1Response    = "ProcessTurn1Response"
	OpProcessValidationFlow   = "ProcessValidationFlow"
	OpProcessHistoricalEnhancement = "ProcessHistoricalEnhancement"
	OpProcessFreshExtraction  = "ProcessFreshExtraction"
	OpExtractObservations     = "ExtractObservations"
	OpExtractMachineStructure = "ExtractMachineStructure"
	OpExtractMachineState     = "ExtractMachineState"
)

// handleProcessingError converts and logs processing errors
func handleProcessingError(logger *slog.Logger, operation string, err error, path types.ProcessingPath) error {
	if err == nil {
		return nil
	}
	
	// Check if it's already a function error
	var fe *errors.FunctionError
	if errors.As(err, &fe) {
		return fe
	}
	
	// Get the appropriate error code
	code := errors.ErrProcessingFailed
	
	// Check for specific error types
	if procErr, ok := err.(*ProcessingError); ok {
		code = errors.ErrProcessingFailed
		err = fmt.Errorf("%s: %s", procErr.Message, procErr.Details)
	} else if parseErr, ok := err.(*ParsingError); ok {
		code = errors.ErrParsingFailed
		err = fmt.Errorf("%s: %v", parseErr.Message, parseErr.InnerErr)
	} else if valErr, ok := err.(*ValidationError); ok {
		code = errors.ErrValidationFailed
		err = fmt.Errorf("%s: field '%s', expected: %v, actual: %v", 
			valErr.Message, valErr.Field, valErr.Expected, valErr.Actual)
	}
	
	// Create a new function error with path information
	details := map[string]interface{}{
		"processingPath": string(path),
	}
	
	funcErr := errors.NewWithDetails(
		operation,
		errors.CategoryProcess,
		code,
		err.Error(),
		errors.SeverityError,
		details,
		err,
	)
	
	// Log the error
	errors.LogError(logger, funcErr)
	
	return funcErr
}

// handleParsingError creates a parsing-specific error
func handleParsingError(logger *slog.Logger, operation string, message string, err error) error {
	details := map[string]interface{}{}
	
	funcErr := errors.NewWithDetails(
		operation,
		errors.CategoryProcess,
		errors.ErrParsingFailed,
		message,
		errors.SeverityError,
		details,
		err,
	)
	
	// Log the error
	errors.LogError(logger, funcErr)
	
	return funcErr
}

// handleValidationError creates a validation-specific error
func handleValidationError(logger *slog.Logger, operation string, field, message string, expected, actual interface{}) error {
	details := map[string]interface{}{
		"field": field,
	}
	
	if expected != nil {
		details["expected"] = expected
	}
	
	if actual != nil {
		details["actual"] = actual
	}
	
	funcErr := errors.NewWithDetails(
		operation,
		errors.CategoryInput,
		errors.ErrValidationFailed,
		message,
		errors.SeverityWarning,
		details,
		nil,
	)
	
	// Log the error
	errors.LogError(logger, funcErr)
	
	return funcErr
}

// createProcessingResult creates a processing result with error information
func createProcessingResultWithError(err error, path types.ProcessingPath) *types.Turn1ProcessingResult {
	// Extract error details
	code := errors.GetCode(err)
	category := errors.GetCategory(err)
	
	// Create base processing result
	result := &types.Turn1ProcessingResult{
		Status:     "EXTRACTION_FAILED",
		SourceType: path,
		ProcessingMetadata: &types.ProcessingMetadata{
			ProcessingStartTime: TimeNow(),
			ProcessingEndTime:   TimeNow(),
			ProcessingPath:      path,
			FallbackReason:      err.Error(),
		},
		ReferenceAnalysis: map[string]interface{}{
			"status":     "EXTRACTION_FAILED",
			"sourceType": string(path),
			"error": map[string]interface{}{
				"code":     code,
				"message":  err.Error(),
				"category": string(category),
			},
		},
		ContextForTurn2: map[string]interface{}{
			"readyForTurn2": false,
			"error": map[string]interface{}{
				"code":     code,
				"message":  err.Error(),
				"category": string(category),
			},
		},
		Warnings: []string{err.Error()},
		FallbackUsed: true,
	}
	
	return result
}