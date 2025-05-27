// Package errors provides standardized error types for workflow functions
package errors

import (
	"encoding/json"
	"fmt"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

// ErrorSeverity indicates the severity of the error
type ErrorSeverity string

// APIErrorSource indicates which API caused the error
type APIErrorSource string

// Error types
const (
	ErrorTypeValidation ErrorType = "ValidationException"
	ErrorTypeAPI        ErrorType = "APIException"
	ErrorTypeBedrock    ErrorType = "BedrockException"
	ErrorTypeConverse   ErrorType = "ConverseException"
	ErrorTypeS3         ErrorType = "S3Exception"
	ErrorTypeDynamoDB   ErrorType = "DynamoDBException"
	ErrorTypeInternal   ErrorType = "InternalException"
	// ErrorTypeTemplate represents errors encountered during template processing
	ErrorTypeTemplate   ErrorType = "TemplateException"
	ErrorTypeTimeout    ErrorType = "TimeoutException"
	ErrorTypeRetryable  ErrorType = "RetryableException"
	ErrorTypeConversion ErrorType = "ConversionException"
	ErrorTypeConfig     ErrorType = "ConfigException"
)

// Error severities
const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// API sources
const (
	APISourceLegacy   APIErrorSource = "InvokeModel"
	APISourceConverse APIErrorSource = "Converse"
	APISourceUnknown  APIErrorSource = "Unknown"
)

// WorkflowError represents a structured error for workflow functions
type WorkflowError struct {
	Type           ErrorType              `json:"errorType"`
	Message        string                 `json:"errorMessage"`
	Code           string                 `json:"errorCode"`
	Details        map[string]interface{} `json:"details"`
	HTTPStatusCode int                    `json:"httpStatusCode"`
	Timestamp      time.Time              `json:"timestamp"`
	VerificationID string                 `json:"verificationId,omitempty"`
	RequestID      string                 `json:"requestId,omitempty"`
	Retryable      bool                   `json:"retryable"`
	Severity       ErrorSeverity          `json:"severity"`
	APISource      APIErrorSource         `json:"apiSource"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *WorkflowError) Error() string {
	return fmt.Sprintf("%s: %s (Code: %s, Source: %s)", e.Type, e.Message, e.Code, e.APISource)
}

// ToJSON converts the error to JSON for API responses
func (e *WorkflowError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// WithContext adds context information to the error
func (e *WorkflowError) WithContext(key string, value interface{}) *WorkflowError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithAPISource sets the API source for the error
func (e *WorkflowError) WithAPISource(source APIErrorSource) *WorkflowError {
	e.APISource = source
	return e
}

// IsAPISpecific checks if the error is specific to a particular API
func (e *WorkflowError) IsAPISpecific() bool {
	return e.APISource != APISourceUnknown
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Retryable
	}
	return false
}

// IsConfigError checks if an error is a configuration error
func IsConfigError(err error) bool {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Type == ErrorTypeConfig
	}
	return false
}

// GetErrorType extracts error type from error
func GetErrorType(err error) ErrorType {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Type
	}
	return ErrorTypeInternal
}

// GetErrorSeverity extracts error severity from error
func GetErrorSeverity(err error) ErrorSeverity {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Severity
	}
	return ErrorSeverityMedium
}

// GetAPISource extracts API source from error
func GetAPISource(err error) APIErrorSource {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.APISource
	}
	return APISourceUnknown
}

// ExtractVerificationID extracts verification ID from error
func ExtractVerificationID(err error) string {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.VerificationID
	}
	return ""
}

// SetVerificationID sets verification ID on error
func SetVerificationID(err error, verificationID string) error {
	if workflowErr, ok := err.(*WorkflowError); ok {
		workflowErr.VerificationID = verificationID
		return workflowErr
	}
	return err
}

// SetRequestID sets request ID on error
func SetRequestID(err error, requestID string) error {
	if workflowErr, ok := err.(*WorkflowError); ok {
		workflowErr.RequestID = requestID
		return workflowErr
	}
	return err
}

// WrapError wraps a generic error with WorkflowError
func WrapError(err error, errorType ErrorType, message string, retryable bool) *WorkflowError {
	return &WorkflowError{
		Type:      errorType,
		Message:   message,
		Code:      "WRAPPED_ERROR",
		Details:   map[string]interface{}{"originalError": err.Error()},
		Retryable: retryable,
		Timestamp: time.Now(),
		Severity:  ErrorSeverityMedium,
		APISource: APISourceUnknown,
	}
}

// Error factory functions

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *WorkflowError {
	return &WorkflowError{
		Type:           ErrorTypeValidation,
		Message:        message,
		Code:           "INVALID_INPUT",
		Details:        details,
		HTTPStatusCode: 400,
		Timestamp:      time.Now(),
		Retryable:      false,
		Severity:       ErrorSeverityMedium,
		APISource:      APISourceUnknown,
	}
}

// NewMissingFieldError creates a new missing field error
func NewMissingFieldError(field string) *WorkflowError {
	return NewValidationError(
		fmt.Sprintf("Missing required field: %s", field),
		map[string]interface{}{"field": field},
	)
}

// NewInvalidFieldError creates a new invalid field error
func NewInvalidFieldError(field string, value interface{}, expected string) *WorkflowError {
	return NewValidationError(
		fmt.Sprintf("Invalid value for field %s: got %v, expected %s", field, value, expected),
		map[string]interface{}{
			"field":    field,
			"value":    value,
			"expected": expected,
		},
	)
}

// NewBedrockError creates a new Bedrock error
func NewBedrockError(message string, code string, retryable bool) *WorkflowError {
	severity := ErrorSeverityMedium
	if retryable {
		severity = ErrorSeverityLow
	} else {
		severity = ErrorSeverityHigh
	}

	return &WorkflowError{
		Type:      ErrorTypeBedrock,
		Message:   message,
		Code:      code,
		Details:   make(map[string]interface{}),
		Retryable: retryable,
		Timestamp: time.Now(),
		Severity:  severity,
		APISource: APISourceLegacy,
	}
}

// NewBedrockThrottlingError creates a new Bedrock throttling error
func NewBedrockThrottlingError() *WorkflowError {
	return &WorkflowError{
		Type:           ErrorTypeBedrock,
		Message:        "Request was throttled by Bedrock API",
		Code:           "THROTTLING",
		HTTPStatusCode: 429,
		Retryable:      true,
		Timestamp:      time.Now(),
		Severity:       ErrorSeverityLow,
		APISource:      APISourceLegacy,
		Details: map[string]interface{}{
			"retryAfter": "2-3 seconds",
			"backoff":    "exponential",
		},
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, duration time.Duration) *WorkflowError {
	return &WorkflowError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("Operation timed out: %s", operation),
		Code:    "TIMEOUT",
		Details: map[string]interface{}{
			"operation": operation,
			"duration":  duration.String(),
		},
		HTTPStatusCode: 408,
		Retryable:      true,
		Timestamp:      time.Now(),
		Severity:       ErrorSeverityMedium,
		APISource:      APISourceUnknown,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(component string, err error) *WorkflowError {
	return &WorkflowError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf("Internal error in component: %s", component),
		Code:    "INTERNAL_ERROR",
		Details: map[string]interface{}{
			"component": component,
			"error":     err.Error(),
		},
		HTTPStatusCode: 500,
		Retryable:      false,
		Timestamp:      time.Now(),
		Severity:       ErrorSeverityHigh,
		APISource:      APISourceUnknown,
	}
}

// NewParsingError creates a new parsing error
func NewParsingError(format string, err error) *WorkflowError {
	return &WorkflowError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf("Failed to parse %s", format),
		Code:    "PARSING_ERROR",
		Details: map[string]interface{}{
			"format": format,
			"error":  err.Error(),
		},
		HTTPStatusCode: 500,
		Retryable:      false,
		Timestamp:      time.Now(),
		Severity:       ErrorSeverityMedium,
		APISource:      APISourceUnknown,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(code string, message string, variable string) *WorkflowError {
	return &WorkflowError{
		Type:    ErrorTypeConfig,
		Message: message,
		Code:    code,
		Details: map[string]interface{}{
			"variable": variable,
		},
		HTTPStatusCode: 500,
		Retryable:      false,
		Timestamp:      time.Now(),
		Severity:       ErrorSeverityCritical,
		APISource:      APISourceUnknown,
	}
}
