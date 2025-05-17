package storage

import (
	"errors"
	"fmt"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

// ErrorSeverity indicates the severity of the error
type ErrorSeverity string

const (
	// Error types
	ErrorTypeValidation   ErrorType = "ValidationException"
	ErrorTypeDynamoDB     ErrorType = "DynamoDBException"
	ErrorTypeS3           ErrorType = "S3Exception"
	ErrorTypeInternal     ErrorType = "InternalException"
	ErrorTypeTimeout      ErrorType = "TimeoutException"
	ErrorTypeRetryable    ErrorType = "RetryableException"
)

// Error severities
const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// StorageError represents a structured error for storage operations
type StorageError struct {
	Type           ErrorType              `json:"errorType"`
	Message        string                 `json:"errorMessage"`
	Code           string                 `json:"errorCode"`
	Details        map[string]interface{} `json:"details"`
	Timestamp      time.Time              `json:"timestamp"`
	VerificationID string                 `json:"verificationId,omitempty"`
	Retryable      bool                   `json:"retryable"`
	Severity       ErrorSeverity          `json:"severity"`
	OriginalError  error                  `json:"-"`
}

// Error implements the error interface
func (e *StorageError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("%s: %s (Code: %s) - %v", e.Type, e.Message, e.Code, e.OriginalError)
	}
	return fmt.Sprintf("%s: %s (Code: %s)", e.Type, e.Message, e.Code)
}

// Unwrap returns the original error
func (e *StorageError) Unwrap() error {
	return e.OriginalError
}

// WithVerificationID adds verification ID to the error
func (e *StorageError) WithVerificationID(id string) *StorageError {
	e.VerificationID = id
	return e
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var storageErr *StorageError
	if errors.As(err, &storageErr) {
		return storageErr.Retryable
	}
	return false
}

// As is a wrapper around errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *StorageError {
	return &StorageError{
		Type:      ErrorTypeValidation,
		Message:   message,
		Code:      "INVALID_INPUT",
		Details:   details,
		Timestamp: time.Now(),
		Retryable: false,
		Severity:  ErrorSeverityMedium,
	}
}

// NewDynamoDBError creates a new DynamoDB error
func NewDynamoDBError(message, code string, retryable bool, originalErr error) *StorageError {
	severity := ErrorSeverityMedium
	if retryable {
		severity = ErrorSeverityLow
	} else {
		severity = ErrorSeverityHigh
	}

	return &StorageError{
		Type:          ErrorTypeDynamoDB,
		Message:       message,
		Code:          code,
		Details:       make(map[string]interface{}),
		Timestamp:     time.Now(),
		Retryable:     retryable,
		Severity:      severity,
		OriginalError: originalErr,
	}
}

// NewS3Error creates a new S3 error
func NewS3Error(message, code string, retryable bool, originalErr error) *StorageError {
	severity := ErrorSeverityMedium
	if retryable {
		severity = ErrorSeverityLow
	} else {
		severity = ErrorSeverityHigh
	}

	return &StorageError{
		Type:          ErrorTypeS3,
		Message:       message,
		Code:          code,
		Details:       make(map[string]interface{}),
		Timestamp:     time.Now(),
		Retryable:     retryable,
		Severity:      severity,
		OriginalError: originalErr,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(component string, originalErr error) *StorageError {
	return &StorageError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf("Internal error in component: %s", component),
		Code:    "INTERNAL_ERROR",
		Details: map[string]interface{}{
			"component": component,
			"error":     originalErr.Error(),
		},
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityHigh,
		OriginalError: originalErr,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, duration time.Duration) *StorageError {
	return &StorageError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("Operation timed out: %s", operation),
		Code:    "TIMEOUT",
		Details: map[string]interface{}{
			"operation": operation,
			"duration":  duration.String(),
		},
		Timestamp: time.Now(),
		Retryable: true,
		Severity:  ErrorSeverityMedium,
	}
}
