// Package errors provides standardized error types for workflow functions
package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
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
	// Enhanced error types
	ErrorTypeNetwork     ErrorType = "NetworkException"
	ErrorTypeAuth        ErrorType = "AuthenticationException"
	ErrorTypePermission  ErrorType = "PermissionException"
	ErrorTypeResource    ErrorType = "ResourceException"
	ErrorTypeTransaction ErrorType = "TransactionException"
	ErrorTypeBatch       ErrorType = "BatchException"
	ErrorTypeThrottling  ErrorType = "ThrottlingException"
	ErrorTypeCapacity    ErrorType = "CapacityException"
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
	APISourceDynamoDB APIErrorSource = "DynamoDB"
	APISourceS3       APIErrorSource = "S3"
)

// ErrorCategory represents the category of error for classification
type ErrorCategory string

const (
	CategoryTransient   ErrorCategory = "TRANSIENT"   // Temporary errors that may resolve
	CategoryPermanent   ErrorCategory = "PERMANENT"   // Errors that won't resolve without intervention
	CategoryClient      ErrorCategory = "CLIENT"      // Client-side errors (4xx)
	CategoryServer      ErrorCategory = "SERVER"      // Server-side errors (5xx)
	CategoryNetwork     ErrorCategory = "NETWORK"     // Network-related errors
	CategoryCapacity    ErrorCategory = "CAPACITY"    // Resource capacity errors
	CategoryAuth        ErrorCategory = "AUTH"        // Authentication/authorization errors
	CategoryValidation  ErrorCategory = "VALIDATION"  // Input validation errors
)

// RetryStrategy defines how errors should be retried
type RetryStrategy string

const (
	RetryNone        RetryStrategy = "NONE"        // No retry
	RetryImmediate   RetryStrategy = "IMMEDIATE"   // Retry immediately
	RetryLinear      RetryStrategy = "LINEAR"      // Linear backoff
	RetryExponential RetryStrategy = "EXPONENTIAL" // Exponential backoff
	RetryJittered    RetryStrategy = "JITTERED"    // Exponential with jitter
)

// DynamoDBErrorCode represents specific DynamoDB error codes
type DynamoDBErrorCode string

const (
	DynamoDBValidationException              DynamoDBErrorCode = "ValidationException"
	DynamoDBConditionalCheckFailedException  DynamoDBErrorCode = "ConditionalCheckFailedException"
	DynamoDBProvisionedThroughputExceeded    DynamoDBErrorCode = "ProvisionedThroughputExceededException"
	DynamoDBResourceNotFoundException        DynamoDBErrorCode = "ResourceNotFoundException"
	DynamoDBInternalServerError              DynamoDBErrorCode = "InternalServerError"
	DynamoDBServiceUnavailableException      DynamoDBErrorCode = "ServiceUnavailableException"
	DynamoDBThrottlingException              DynamoDBErrorCode = "ThrottlingException"
	DynamoDBLimitExceededException           DynamoDBErrorCode = "LimitExceededException"
	DynamoDBItemCollectionSizeLimitExceeded  DynamoDBErrorCode = "ItemCollectionSizeLimitExceededException"
	DynamoDBTransactionConflictException     DynamoDBErrorCode = "TransactionConflictException"
	DynamoDBTransactionCanceledException     DynamoDBErrorCode = "TransactionCanceledException"
	DynamoDBTransactionInProgressException   DynamoDBErrorCode = "TransactionInProgressException"
	DynamoDBDuplicateTransactionException    DynamoDBErrorCode = "DuplicateTransactionException"
	DynamoDBRequestLimitExceeded             DynamoDBErrorCode = "RequestLimitExceeded"
	DynamoDBBackupNotFoundException          DynamoDBErrorCode = "BackupNotFoundException"
	DynamoDBBackupInUseException             DynamoDBErrorCode = "BackupInUseException"
	DynamoDBTableNotFoundException           DynamoDBErrorCode = "TableNotFoundException"
	DynamoDBTableInUseException              DynamoDBErrorCode = "TableInUseException"
	DynamoDBIndexNotFoundException           DynamoDBErrorCode = "IndexNotFoundException"
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
	// Enhanced fields
	Category       ErrorCategory          `json:"category,omitempty"`
	RetryStrategy  RetryStrategy          `json:"retryStrategy,omitempty"`
	RetryCount     int                    `json:"retryCount,omitempty"`
	MaxRetries     int                    `json:"maxRetries,omitempty"`
	StackTrace     string                 `json:"stackTrace,omitempty"`
	CorrelationID  string                 `json:"correlationId,omitempty"`
	Component      string                 `json:"component,omitempty"`
	Operation      string                 `json:"operation,omitempty"`
	TableName      string                 `json:"tableName,omitempty"`
	IndexName      string                 `json:"indexName,omitempty"`
	ItemKey        map[string]interface{} `json:"itemKey,omitempty"`
	Suggestions    []string               `json:"suggestions,omitempty"`
	RecoveryHints  []string               `json:"recoveryHints,omitempty"`
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

// WithCategory sets the error category
func (e *WorkflowError) WithCategory(category ErrorCategory) *WorkflowError {
	e.Category = category
	return e
}

// WithSeverity sets the error severity
func (e *WorkflowError) WithSeverity(severity ErrorSeverity) *WorkflowError {
	e.Severity = severity
	return e
}

// WithRetryStrategy sets the retry strategy
func (e *WorkflowError) WithRetryStrategy(strategy RetryStrategy) *WorkflowError {
	e.RetryStrategy = strategy
	return e
}

// WithComponent sets the component name
func (e *WorkflowError) WithComponent(component string) *WorkflowError {
	e.Component = component
	return e
}

// WithOperation sets the operation name
func (e *WorkflowError) WithOperation(operation string) *WorkflowError {
	e.Operation = operation
	return e
}

// WithTableName sets the DynamoDB table name
func (e *WorkflowError) WithTableName(tableName string) *WorkflowError {
	e.TableName = tableName
	return e
}

// WithIndexName sets the DynamoDB index name
func (e *WorkflowError) WithIndexName(indexName string) *WorkflowError {
	e.IndexName = indexName
	return e
}

// WithItemKey sets the DynamoDB item key
func (e *WorkflowError) WithItemKey(key map[string]interface{}) *WorkflowError {
	e.ItemKey = key
	return e
}

// WithCorrelationID sets the correlation ID for tracing
func (e *WorkflowError) WithCorrelationID(correlationID string) *WorkflowError {
	e.CorrelationID = correlationID
	return e
}

// WithVerificationID sets the verification ID
func (e *WorkflowError) WithVerificationID(verificationID string) *WorkflowError {
	e.VerificationID = verificationID
	return e
}

// WithRequestID sets the request ID
func (e *WorkflowError) WithRequestID(requestID string) *WorkflowError {
	e.RequestID = requestID
	return e
}

// WithSuggestions adds recovery suggestions
func (e *WorkflowError) WithSuggestions(suggestions ...string) *WorkflowError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithRecoveryHints adds recovery hints
func (e *WorkflowError) WithRecoveryHints(hints ...string) *WorkflowError {
	e.RecoveryHints = append(e.RecoveryHints, hints...)
	return e
}

// IncrementRetryCount increments the retry count
func (e *WorkflowError) IncrementRetryCount() *WorkflowError {
	e.RetryCount++
	return e
}

// SetMaxRetries sets the maximum retry count
func (e *WorkflowError) SetMaxRetries(maxRetries int) *WorkflowError {
	e.MaxRetries = maxRetries
	return e
}

// IsRetryLimitExceeded checks if retry limit is exceeded
func (e *WorkflowError) IsRetryLimitExceeded() bool {
	return e.MaxRetries > 0 && e.RetryCount >= e.MaxRetries
}

// GetRetryDelay calculates retry delay based on strategy
func (e *WorkflowError) GetRetryDelay(baseDelay time.Duration) time.Duration {
	switch e.RetryStrategy {
	case RetryNone:
		return 0
	case RetryImmediate:
		return 0
	case RetryLinear:
		return baseDelay * time.Duration(e.RetryCount+1)
	case RetryExponential:
		multiplier := 1
		for i := 0; i < e.RetryCount; i++ {
			multiplier *= 2
		}
		return baseDelay * time.Duration(multiplier)
	case RetryJittered:
		multiplier := 1
		for i := 0; i < e.RetryCount; i++ {
			multiplier *= 2
		}
		delay := baseDelay * time.Duration(multiplier)
		// Add 25% jitter (would need math/rand import)
		return delay
	default:
		return baseDelay
	}
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

// =============================================================================
// DynamoDB Error Functions
// =============================================================================

// NewDynamoDBError creates a generic DynamoDB error
func NewDynamoDBError(operation string, tableName string, err error) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       fmt.Sprintf("DynamoDB %s operation failed", operation),
		Code:          "DYNAMODB_ERROR",
		Details:       map[string]interface{}{"originalError": err.Error()},
		HTTPStatusCode: 500,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryServer,
		RetryStrategy: RetryNone,
		Operation:     operation,
		TableName:     tableName,
	}
}

// NewDynamoDBValidationError creates a DynamoDB validation error
func NewDynamoDBValidationError(operation string, tableName string, message string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       fmt.Sprintf("DynamoDB validation error: %s", message),
		Code:          string(DynamoDBValidationException),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityHigh,
		APISource:     APISourceDynamoDB,
		Category:      CategoryValidation,
		RetryStrategy: RetryNone,
		Operation:     operation,
		TableName:     tableName,
		Details: map[string]interface{}{
			"errorType": "ValidationException",
		},
		Suggestions: []string{
			"Check item structure and required fields",
			"Verify data types match table schema",
			"Ensure primary key attributes are provided",
		},
	}
}

// NewDynamoDBConditionalCheckFailedError creates a conditional check failed error
func NewDynamoDBConditionalCheckFailedError(operation string, tableName string, condition string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       "DynamoDB conditional check failed",
		Code:          string(DynamoDBConditionalCheckFailedException),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryClient,
		RetryStrategy: RetryNone,
		Operation:     operation,
		TableName:     tableName,
		Details: map[string]interface{}{
			"errorType":  "ConditionalCheckFailedException",
			"condition":  condition,
		},
		Suggestions: []string{
			"Item may already exist or condition expression failed",
			"Check if the condition expression is correct",
			"Verify the item state before the operation",
		},
	}
}

// NewDynamoDBThroughputExceededError creates a throughput exceeded error
func NewDynamoDBThroughputExceededError(operation string, tableName string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       "DynamoDB provisioned throughput exceeded",
		Code:          string(DynamoDBProvisionedThroughputExceeded),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityLow,
		APISource:     APISourceDynamoDB,
		Category:      CategoryCapacity,
		RetryStrategy: RetryExponential,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    5,
		Details: map[string]interface{}{
			"errorType": "ProvisionedThroughputExceededException",
		},
		Suggestions: []string{
			"Implement exponential backoff retry strategy",
			"Consider increasing table provisioned capacity",
			"Use auto-scaling for dynamic capacity adjustment",
		},
		RecoveryHints: []string{
			"Retry with exponential backoff",
			"Monitor CloudWatch metrics for capacity utilization",
		},
	}
}

// NewDynamoDBResourceNotFoundError creates a resource not found error
func NewDynamoDBResourceNotFoundError(operation string, resourceType string, resourceName string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       fmt.Sprintf("DynamoDB %s not found: %s", resourceType, resourceName),
		Code:          string(DynamoDBResourceNotFoundException),
		HTTPStatusCode: 404,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityHigh,
		APISource:     APISourceDynamoDB,
		Category:      CategoryClient,
		RetryStrategy: RetryNone,
		Operation:     operation,
		Details: map[string]interface{}{
			"errorType":    "ResourceNotFoundException",
			"resourceType": resourceType,
			"resourceName": resourceName,
		},
		Suggestions: []string{
			"Verify table name and AWS region configuration",
			"Check if the table exists in the specified region",
			"Ensure proper IAM permissions for table access",
		},
	}
}

// NewDynamoDBInternalServerError creates an internal server error
func NewDynamoDBInternalServerError(operation string, tableName string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       "DynamoDB internal server error",
		Code:          string(DynamoDBInternalServerError),
		HTTPStatusCode: 500,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryServer,
		RetryStrategy: RetryExponential,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    3,
		Details: map[string]interface{}{
			"errorType": "InternalServerError",
		},
		Suggestions: []string{
			"Retry the operation with exponential backoff",
			"Check AWS service health dashboard",
		},
		RecoveryHints: []string{
			"This is a temporary AWS service issue",
			"Retry should resolve the issue",
		},
	}
}

// NewDynamoDBThrottlingError creates a throttling error
func NewDynamoDBThrottlingError(operation string, tableName string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeDynamoDB,
		Message:       "DynamoDB request throttled",
		Code:          string(DynamoDBThrottlingException),
		HTTPStatusCode: 429,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityLow,
		APISource:     APISourceDynamoDB,
		Category:      CategoryCapacity,
		RetryStrategy: RetryJittered,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    5,
		Details: map[string]interface{}{
			"errorType": "ThrottlingException",
		},
		Suggestions: []string{
			"Implement exponential backoff with jitter",
			"Reduce request rate",
			"Consider using batch operations",
		},
		RecoveryHints: []string{
			"Retry with exponential backoff and jitter",
			"Monitor request patterns and optimize",
		},
	}
}

// NewDynamoDBTransactionConflictError creates a transaction conflict error
func NewDynamoDBTransactionConflictError(operation string, tableName string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeTransaction,
		Message:       "DynamoDB transaction conflict detected",
		Code:          string(DynamoDBTransactionConflictException),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryTransient,
		RetryStrategy: RetryExponential,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    3,
		Details: map[string]interface{}{
			"errorType": "TransactionConflictException",
		},
		Suggestions: []string{
			"Retry the transaction with exponential backoff",
			"Consider reducing transaction scope",
			"Implement optimistic locking patterns",
		},
		RecoveryHints: []string{
			"Transaction conflicts are common in high-concurrency scenarios",
			"Retry should resolve most conflicts",
		},
	}
}

// NewDynamoDBTransactionCancelledError creates a transaction cancelled error
func NewDynamoDBTransactionCancelledError(operation string, tableName string, reason string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeTransaction,
		Message:       fmt.Sprintf("DynamoDB transaction cancelled: %s", reason),
		Code:          string(DynamoDBTransactionCanceledException),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityHigh,
		APISource:     APISourceDynamoDB,
		Category:      CategoryClient,
		RetryStrategy: RetryNone,
		Operation:     operation,
		TableName:     tableName,
		Details: map[string]interface{}{
			"errorType":           "TransactionCanceledException",
			"cancellationReason":  reason,
		},
		Suggestions: []string{
			"Check transaction conditions and constraints",
			"Verify all items in transaction exist and meet conditions",
			"Review transaction logic for conflicts",
		},
	}
}

// NewDynamoDBLimitExceededError creates a limit exceeded error
func NewDynamoDBLimitExceededError(operation string, tableName string, limitType string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeCapacity,
		Message:       fmt.Sprintf("DynamoDB limit exceeded: %s", limitType),
		Code:          string(DynamoDBLimitExceededException),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryCapacity,
		RetryStrategy: RetryLinear,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    3,
		Details: map[string]interface{}{
			"errorType":  "LimitExceededException",
			"limitType":  limitType,
		},
		Suggestions: []string{
			"Reduce batch size or request rate",
			"Implement pagination for large queries",
			"Consider using parallel processing with smaller batches",
		},
		RecoveryHints: []string{
			"Retry with smaller batch sizes",
			"Implement rate limiting",
		},
	}
}

// NewDynamoDBItemCollectionSizeLimitExceededError creates an item collection size limit error
func NewDynamoDBItemCollectionSizeLimitExceededError(operation string, tableName string, partitionKey string) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeCapacity,
		Message:       "DynamoDB item collection size limit exceeded",
		Code:          string(DynamoDBItemCollectionSizeLimitExceeded),
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     false,
		Severity:      ErrorSeverityHigh,
		APISource:     APISourceDynamoDB,
		Category:      CategoryClient,
		RetryStrategy: RetryNone,
		Operation:     operation,
		TableName:     tableName,
		Details: map[string]interface{}{
			"errorType":    "ItemCollectionSizeLimitExceededException",
			"partitionKey": partitionKey,
		},
		Suggestions: []string{
			"Redesign partition key to distribute data more evenly",
			"Consider using composite partition keys",
			"Archive or delete old items to reduce collection size",
		},
		RecoveryHints: []string{
			"This indicates a data modeling issue",
			"Partition key design needs to be reconsidered",
		},
	}
}

// =============================================================================
// Enhanced DynamoDB Error Analysis Functions
// =============================================================================

// AnalyzeDynamoDBError analyzes a DynamoDB error and returns an enhanced WorkflowError
func AnalyzeDynamoDBError(operation string, tableName string, err error) *WorkflowError {
	if err == nil {
		return nil
	}

	errorStr := err.Error()

	// Check for specific DynamoDB error patterns
	if containsIgnoreCase(errorStr, "ValidationException") {
		return NewDynamoDBValidationError(operation, tableName, errorStr)
	}

	if containsIgnoreCase(errorStr, "ConditionalCheckFailedException") {
		return NewDynamoDBConditionalCheckFailedError(operation, tableName, "condition check failed")
	}

	if containsIgnoreCase(errorStr, "ProvisionedThroughputExceededException") {
		return NewDynamoDBThroughputExceededError(operation, tableName)
	}

	if containsIgnoreCase(errorStr, "ResourceNotFoundException") {
		return NewDynamoDBResourceNotFoundError(operation, "table", tableName)
	}

	if containsIgnoreCase(errorStr, "InternalServerError") {
		return NewDynamoDBInternalServerError(operation, tableName)
	}

	if containsIgnoreCase(errorStr, "ThrottlingException") {
		return NewDynamoDBThrottlingError(operation, tableName)
	}

	if containsIgnoreCase(errorStr, "TransactionConflictException") {
		return NewDynamoDBTransactionConflictError(operation, tableName)
	}

	if containsIgnoreCase(errorStr, "TransactionCanceledException") {
		return NewDynamoDBTransactionCancelledError(operation, tableName, "transaction cancelled")
	}

	if containsIgnoreCase(errorStr, "LimitExceededException") {
		return NewDynamoDBLimitExceededError(operation, tableName, "request limit")
	}

	if containsIgnoreCase(errorStr, "ItemCollectionSizeLimitExceededException") {
		return NewDynamoDBItemCollectionSizeLimitExceededError(operation, tableName, "unknown")
	}

	// Default to generic DynamoDB error
	return NewDynamoDBError(operation, tableName, err)
}

// IsDynamoDBRetryableError checks if a DynamoDB error is retryable
func IsDynamoDBRetryableError(err error) bool {
	if workflowErr, ok := err.(*WorkflowError); ok {
		if workflowErr.Type != ErrorTypeDynamoDB && workflowErr.Type != ErrorTypeTransaction && workflowErr.Type != ErrorTypeCapacity {
			return false
		}
		return workflowErr.Retryable
	}

	// Check raw error strings for retryable patterns
	errorStr := err.Error()
	retryablePatterns := []string{
		"ProvisionedThroughputExceededException",
		"InternalServerError",
		"ServiceUnavailableException",
		"ThrottlingException",
		"TransactionConflictException",
		"LimitExceededException",
	}

	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errorStr, pattern) {
			return true
		}
	}

	return false
}

// GetDynamoDBRetryStrategy returns the appropriate retry strategy for a DynamoDB error
func GetDynamoDBRetryStrategy(err error) RetryStrategy {
	if workflowErr, ok := err.(*WorkflowError); ok {
		if workflowErr.RetryStrategy != "" {
			return workflowErr.RetryStrategy
		}
	}

	errorStr := err.Error()

	if containsIgnoreCase(errorStr, "ThrottlingException") {
		return RetryJittered
	}

	if containsIgnoreCase(errorStr, "ProvisionedThroughputExceededException") {
		return RetryExponential
	}

	if containsIgnoreCase(errorStr, "TransactionConflictException") {
		return RetryExponential
	}

	if containsIgnoreCase(errorStr, "InternalServerError") {
		return RetryExponential
	}

	if containsIgnoreCase(errorStr, "LimitExceededException") {
		return RetryLinear
	}

	return RetryNone
}

// =============================================================================
// Utility Functions
// =============================================================================

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     findSubstring(s, substr)))
}

// findSubstring is a simple case-insensitive substring search
func findSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLowerCase(s[i+j]) != toLowerCase(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLowerCase converts a byte to lowercase
func toLowerCase(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// =============================================================================
// Enhanced Error Context and Monitoring Functions
// =============================================================================

// WithStackTrace adds stack trace to error (requires runtime import)
func (e *WorkflowError) WithStackTrace() *WorkflowError {
	// Get stack trace
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			e.StackTrace = string(buf[:n])
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	return e
}

// GetErrorMetrics returns error metrics for monitoring
func GetErrorMetrics(err error) map[string]interface{} {
	metrics := make(map[string]interface{})

	if workflowErr, ok := err.(*WorkflowError); ok {
		metrics["error_type"] = string(workflowErr.Type)
		metrics["error_code"] = workflowErr.Code
		metrics["severity"] = string(workflowErr.Severity)
		metrics["category"] = string(workflowErr.Category)
		metrics["retryable"] = workflowErr.Retryable
		metrics["retry_count"] = workflowErr.RetryCount
		metrics["api_source"] = string(workflowErr.APISource)
		metrics["component"] = workflowErr.Component
		metrics["operation"] = workflowErr.Operation
		metrics["table_name"] = workflowErr.TableName
		metrics["http_status_code"] = workflowErr.HTTPStatusCode

		if workflowErr.VerificationID != "" {
			metrics["verification_id"] = workflowErr.VerificationID
		}

		if workflowErr.CorrelationID != "" {
			metrics["correlation_id"] = workflowErr.CorrelationID
		}
	} else {
		metrics["error_type"] = "unknown"
		metrics["error_message"] = err.Error()
	}

	return metrics
}

// IsTransientError checks if an error is transient (temporary)
func IsTransientError(err error) bool {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Category == CategoryTransient ||
			   workflowErr.Category == CategoryCapacity ||
			   workflowErr.Category == CategoryNetwork
	}
	return false
}

// IsPermanentError checks if an error is permanent
func IsPermanentError(err error) bool {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Category == CategoryPermanent ||
			   workflowErr.Category == CategoryValidation ||
			   (workflowErr.Category == CategoryClient && !workflowErr.Retryable)
	}
	return false
}

// GetErrorSuggestions returns suggestions for error resolution
func GetErrorSuggestions(err error) []string {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.Suggestions
	}
	return []string{}
}

// GetRecoveryHints returns recovery hints for error resolution
func GetRecoveryHints(err error) []string {
	if workflowErr, ok := err.(*WorkflowError); ok {
		return workflowErr.RecoveryHints
	}
	return []string{}
}

// =============================================================================
// Batch and Transaction Error Functions
// =============================================================================

// NewBatchOperationError creates an error for batch operations
func NewBatchOperationError(operation string, tableName string, failedItems int, totalItems int) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeBatch,
		Message:       fmt.Sprintf("Batch operation partially failed: %d/%d items failed", failedItems, totalItems),
		Code:          "BATCH_PARTIAL_FAILURE",
		HTTPStatusCode: 207, // Multi-status
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryTransient,
		RetryStrategy: RetryExponential,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    3,
		Details: map[string]interface{}{
			"failed_items": failedItems,
			"total_items":  totalItems,
			"success_rate": float64(totalItems-failedItems) / float64(totalItems),
		},
		Suggestions: []string{
			"Retry failed items individually",
			"Check for throttling or capacity issues",
			"Consider reducing batch size",
		},
		RecoveryHints: []string{
			"Extract failed items and retry",
			"Monitor for patterns in failures",
		},
	}
}

// NewTransactionError creates a generic transaction error
func NewTransactionError(operation string, tableName string, transactionType string, err error) *WorkflowError {
	return &WorkflowError{
		Type:          ErrorTypeTransaction,
		Message:       fmt.Sprintf("Transaction %s failed: %s", transactionType, err.Error()),
		Code:          "TRANSACTION_ERROR",
		HTTPStatusCode: 400,
		Timestamp:     time.Now(),
		Retryable:     true,
		Severity:      ErrorSeverityMedium,
		APISource:     APISourceDynamoDB,
		Category:      CategoryTransient,
		RetryStrategy: RetryExponential,
		Operation:     operation,
		TableName:     tableName,
		MaxRetries:    3,
		Details: map[string]interface{}{
			"transaction_type": transactionType,
			"original_error":   err.Error(),
		},
		Suggestions: []string{
			"Retry the transaction with exponential backoff",
			"Check for concurrent modifications",
			"Consider optimistic locking patterns",
		},
	}
}

// =============================================================================
// Error Aggregation and Reporting Functions
// =============================================================================

// ErrorSummary represents a summary of multiple errors
type ErrorSummary struct {
	TotalErrors    int                    `json:"total_errors"`
	ErrorsByType   map[string]int         `json:"errors_by_type"`
	ErrorsByCode   map[string]int         `json:"errors_by_code"`
	RetryableCount int                    `json:"retryable_count"`
	CriticalCount  int                    `json:"critical_count"`
	MostCommon     string                 `json:"most_common_error"`
	Suggestions    []string               `json:"suggestions"`
	Timestamp      time.Time              `json:"timestamp"`
}

// AggregateErrors creates a summary of multiple errors
func AggregateErrors(errors []error) *ErrorSummary {
	summary := &ErrorSummary{
		TotalErrors:  len(errors),
		ErrorsByType: make(map[string]int),
		ErrorsByCode: make(map[string]int),
		Suggestions:  []string{},
		Timestamp:    time.Now(),
	}

	suggestionSet := make(map[string]bool)

	for _, err := range errors {
		if workflowErr, ok := err.(*WorkflowError); ok {
			summary.ErrorsByType[string(workflowErr.Type)]++
			summary.ErrorsByCode[workflowErr.Code]++

			if workflowErr.Retryable {
				summary.RetryableCount++
			}

			if workflowErr.Severity == ErrorSeverityCritical {
				summary.CriticalCount++
			}

			// Collect unique suggestions
			for _, suggestion := range workflowErr.Suggestions {
				if !suggestionSet[suggestion] {
					suggestionSet[suggestion] = true
					summary.Suggestions = append(summary.Suggestions, suggestion)
				}
			}
		} else {
			summary.ErrorsByType["unknown"]++
			summary.ErrorsByCode["unknown"]++
		}
	}

	// Find most common error type
	maxCount := 0
	for errorType, count := range summary.ErrorsByType {
		if count > maxCount {
			maxCount = count
			summary.MostCommon = errorType
		}
	}

	return summary
}
