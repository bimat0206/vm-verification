// Package errors provides custom error handling for the ProcessTurn1Response Lambda
package errors

import (
	"fmt"
	"time"
	
	"workflow-function/shared/s3state"
)

// ConvertS3StateError converts s3state errors to FunctionError
func ConvertS3StateError(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	// Check if it's already a FunctionError
	var fe *FunctionError
	if As(err, &fe) {
		return err
	}
	
	// Check for S3State error types and convert accordingly
	switch {
	case s3state.IsValidationError(err):
		return New(
			operation,
			CategoryInput,
			ErrValidationFailed,
			err.Error(),
			SeverityWarning,
			err,
		)
		
	case s3state.IsReferenceError(err):
		return New(
			operation,
			CategoryState,
			ErrReferenceInvalid,
			err.Error(),
			SeverityError,
			err,
		)
		
	case s3state.IsJSONError(err):
		return New(
			operation,
			CategoryProcess,
			ErrInvalidFormat,
			err.Error(),
			SeverityError,
			err,
		)
		
	case s3state.IsS3Error(err):
		return New(
			operation,
			CategoryState,
			ErrStateStoreFailed,
			err.Error(),
			SeverityError,
			err,
		)
		
	case s3state.IsCategoryError(err):
		return New(
			operation,
			CategoryState,
			ErrStateStoreFailed,
			err.Error(),
			SeverityError,
			err,
		)
		
	default:
		return New(
			operation,
			CategorySystem,
			ErrInternalError,
			err.Error(),
			SeverityError,
			err,
		)
	}
}

// ConvertToErrorInfo converts a FunctionError to an API-friendly error response
func ConvertToErrorInfo(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	
	errorInfo := make(map[string]interface{})
	
	var fe *FunctionError
	if As(err, &fe) {
		errorInfo["code"] = fe.Code
		errorInfo["message"] = fe.Message
		errorInfo["category"] = string(fe.Category)
		
		if len(fe.Details) > 0 {
			errorInfo["details"] = fe.Details
		}
	} else {
		// Generic error handling
		errorInfo["code"] = ErrInternalError
		errorInfo["message"] = err.Error()
		errorInfo["category"] = string(CategorySystem)
	}
	
	// Add timestamp
	errorInfo["timestamp"] = GetCurrentTimestamp()
	
	return errorInfo
}

// GetCurrentTimestamp returns a formatted timestamp for error reporting
func GetCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixNano() / int64(time.Millisecond))
}