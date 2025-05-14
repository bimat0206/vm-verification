package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

// Custom error types for different scenarios
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "ValidationException"
	ErrorTypeAPI         ErrorType = "APIException"
	ErrorTypeBedrock     ErrorType = "BedrockException"
	ErrorTypeS3          ErrorType = "S3Exception"
	ErrorTypeDynamoDB    ErrorType = "DynamoDBException"
	ErrorTypeInternal    ErrorType = "InternalException"
	ErrorTypeTimeout     ErrorType = "TimeoutException"
	ErrorTypeRetryable   ErrorType = "RetryableException"
)

// ExecuteTurn1Error represents a structured error for the ExecuteTurn1 function
type ExecuteTurn1Error struct {
	Type           ErrorType              `json:"errorType"`
	Message        string                 `json:"errorMessage"`
	Code           string                 `json:"errorCode"`
	Details        map[string]interface{} `json:"details"`
	HTTPStatusCode int                    `json:"httpStatusCode"`
	Timestamp      time.Time              `json:"timestamp"`
	VerificationID string                 `json:"verificationId,omitempty"`
	RequestID      string                 `json:"requestId,omitempty"`
	Retryable      bool                   `json:"retryable"`
}

// Error implements the error interface
func (e *ExecuteTurn1Error) Error() string {
	return fmt.Sprintf("%s: %s (Code: %s)", e.Type, e.Message, e.Code)
}

// ToJSON converts the error to JSON for API responses
func (e *ExecuteTurn1Error) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// Validation Errors
func NewValidationError(message string, details map[string]interface{}) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:           ErrorTypeValidation,
		Message:        message,
		Code:           "INVALID_INPUT",
		Details:        details,
		HTTPStatusCode: 400,
		Timestamp:      time.Now(),
		Retryable:      false,
	}
}

func NewMissingFieldError(field string) *ExecuteTurn1Error {
	return NewValidationError(
		fmt.Sprintf("Missing required field: %s", field),
		map[string]interface{}{"field": field},
	)
}

func NewInvalidFieldError(field string, value interface{}, expected string) *ExecuteTurn1Error {
	return NewValidationError(
		fmt.Sprintf("Invalid value for field %s: got %v, expected %s", field, value, expected),
		map[string]interface{}{
			"field":    field,
			"value":    value,
			"expected": expected,
		},
	)
}

func NewInvalidImageFormatError(format string, supportedFormats []string) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeValidation,
		Message: fmt.Sprintf("Invalid image format: %s", format),
		Code:    "INVALID_IMAGE_FORMAT",
		Details: map[string]interface{}{
			"providedFormat":   format,
			"supportedFormats": supportedFormats,
		},
		HTTPStatusCode: 400,
		Timestamp:      time.Now(),
		Retryable:      false,
	}
}

// Bedrock API Errors
func NewBedrockError(message string, code string, retryable bool) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:      ErrorTypeBedrock,
		Message:   message,
		Code:      code,
		Details:   make(map[string]interface{}),
		Retryable: retryable,
		Timestamp: time.Now(),
	}
}

func NewBedrockThrottlingError() *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:           ErrorTypeBedrock,
		Message:        "Request was throttled by Bedrock API",
		Code:           "THROTTLING",
		HTTPStatusCode: 429,
		Retryable:      true,
		Timestamp:      time.Now(),
	}
}

func NewBedrockTokenLimitError(inputTokens, maxTokens int) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeBedrock,
		Message: "Input exceeds maximum token limit",
		Code:    "TOKEN_LIMIT_EXCEEDED",
		Details: map[string]interface{}{
			"inputTokens": inputTokens,
			"maxTokens":   maxTokens,
		},
		HTTPStatusCode: 400,
		Retryable:      false,
		Timestamp:      time.Now(),
	}
}

func NewBedrockModelError(modelID string) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeBedrock,
		Message: fmt.Sprintf("Model not available: %s", modelID),
		Code:    "MODEL_UNAVAILABLE",
		Details: map[string]interface{}{
			"modelId": modelID,
		},
		HTTPStatusCode: 400,
		Retryable:      false,
		Timestamp:      time.Now(),
	}
}

// S3 Errors
func NewS3Error(operation string, bucket string, key string, err error) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeS3,
		Message: fmt.Sprintf("S3 operation failed: %s", operation),
		Code:    "S3_OPERATION_FAILED",
		Details: map[string]interface{}{
			"operation": operation,
			"bucket":    bucket,
			"key":       key,
			"error":     err.Error(),
		},
		HTTPStatusCode: 500,
		Retryable:      true,
		Timestamp:      time.Now(),
	}
}

func NewS3AccessDeniedError(bucket string, key string) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeS3,
		Message: "Access denied to S3 resource",
		Code:    "S3_ACCESS_DENIED",
		Details: map[string]interface{}{
			"bucket": bucket,
			"key":    key,
		},
		HTTPStatusCode: 403,
		Retryable:      false,
		Timestamp:      time.Now(),
	}
}

// DynamoDB Errors
func NewDynamoDBError(operation string, tableName string, err error) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeDynamoDB,
		Message: fmt.Sprintf("DynamoDB operation failed: %s", operation),
		Code:    "DYNAMODB_OPERATION_FAILED",
		Details: map[string]interface{}{
			"operation": operation,
			"table":     tableName,
			"error":     err.Error(),
		},
		HTTPStatusCode: 500,
		Retryable:      true,
		Timestamp:      time.Now(),
	}
}

// Timeout Errors
func NewTimeoutError(operation string, duration time.Duration) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
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
	}
}

// Internal Processing Errors
func NewInternalError(component string, err error) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
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
	}
}

func NewParsingError(format string, err error) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
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
	}
}

// Response formatting errors
func NewResponseFormattingError(stage string, err error) *ExecuteTurn1Error {
	return &ExecuteTurn1Error{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf("Failed to format response at stage: %s", stage),
		Code:    "RESPONSE_FORMATTING_ERROR",
		Details: map[string]interface{}{
			"stage": stage,
			"error": err.Error(),
		},
		HTTPStatusCode: 500,
		Retryable:      false,
		Timestamp:      time.Now(),
	}
}

// Check if an error is retryable
func IsRetryable(err error) bool {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		return execErr.Retryable
	}
	return false
}

// Extract verification ID from error
func ExtractVerificationID(err error) string {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		return execErr.VerificationID
	}
	return ""
}

// Set verification ID on error
func SetVerificationID(err error, verificationID string) error {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		execErr.VerificationID = verificationID
		return execErr
	}
	return err
}

// Set request ID on error
func SetRequestID(err error, requestID string) error {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		execErr.RequestID = requestID
		return execErr
	}
	return err
}

// Common error response structure for Lambda
type LambdaErrorResponse struct {
	Error struct {
		ErrorType      string `json:"errorType"`
		ErrorMessage   string `json:"errorMessage"`
		ErrorCode      string `json:"errorCode"`
		Details        map[string]interface{} `json:"details"`
	} `json:"error"`
	VerificationContext struct {
		VerificationID string `json:"verificationId"`
		Status         string `json:"status"`
	} `json:"verificationContext"`
}

// Convert error to Lambda error response
func ToLambdaErrorResponse(err error, verificationID string) *LambdaErrorResponse {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		return &LambdaErrorResponse{
			Error: struct {
				ErrorType      string `json:"errorType"`
				ErrorMessage   string `json:"errorMessage"`
				ErrorCode      string `json:"errorCode"`
				Details        map[string]interface{} `json:"details"`
			}{
				ErrorType:    string(execErr.Type),
				ErrorMessage: execErr.Message,
				ErrorCode:    execErr.Code,
				Details:      execErr.Details,
			},
			VerificationContext: struct {
				VerificationID string `json:"verificationId"`
				Status         string `json:"status"`
			}{
				VerificationID: verificationID,
				Status:         "TURN1_FAILED",
			},
		}
	}

	// Fallback for non-ExecuteTurn1Error types
	return &LambdaErrorResponse{
		Error: struct {
			ErrorType      string `json:"errorType"`
			ErrorMessage   string `json:"errorMessage"`
			ErrorCode      string `json:"errorCode"`
			Details        map[string]interface{} `json:"details"`
		}{
			ErrorType:    string(ErrorTypeInternal),
			ErrorMessage: err.Error(),
			ErrorCode:    "UNKNOWN_ERROR",
			Details:      map[string]interface{}{},
		},
		VerificationContext: struct {
			VerificationID string `json:"verificationId"`
			Status         string `json:"status"`
		}{
			VerificationID: verificationID,
			Status:         "TURN1_FAILED",
		},
	}
}