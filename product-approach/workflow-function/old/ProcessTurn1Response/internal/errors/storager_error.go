// Package errors provides custom error handling for the ProcessTurn1Response Lambda
package errors

// StorageError represents errors related to storage operations
type StorageError struct {
	// Type of storage error (S3, DynamoDB, etc.)
	Type string
	
	// Message describes the error
	Message string
	
	// Code is a storage-specific error code
	Code string
	
	// Details contains additional error context
	Details map[string]interface{}
	
	// Retryable indicates if the operation can be retried
	Retryable bool
	
	// Severity indicates the error impact
	Severity string
	
	// OriginalError contains the underlying error
	OriginalError error
}

// Error implements the error interface
func (e *StorageError) Error() string {
	if e.OriginalError != nil {
		return e.Message + ": " + e.OriginalError.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.OriginalError
}

// NewStorageError creates a new storage error
func NewStorageError(errorType, message, code string, retryable bool, severity string, originalError error) *StorageError {
	return &StorageError{
		Type:          errorType,
		Message:       message,
		Code:          code,
		Details:       make(map[string]interface{}),
		Retryable:     retryable,
		Severity:      severity,
		OriginalError: originalError,
	}
}

// WithDetails adds details to a storage error
func (e *StorageError) WithDetails(details map[string]interface{}) *StorageError {
	if details != nil {
		for k, v := range details {
			e.Details[k] = v
		}
	}
	return e
}

// NewS3Error creates a new S3-specific error
func NewS3Error(message, code string, retryable bool, originalError error) *StorageError {
	return NewStorageError("S3Exception", message, code, retryable, "HIGH", originalError)
}

// NewDynamoDBError creates a new DynamoDB-specific error
func NewDynamoDBError(message, code string, retryable bool, originalError error) *StorageError {
	return NewStorageError("DynamoDBException", message, code, retryable, "HIGH", originalError)
}

// IsStorageError checks if an error is a StorageError
func IsStorageError(err error) bool {
	var se *StorageError
	return As(err, &se)
}

// IsS3Error checks if an error is an S3-specific StorageError
func IsS3Error(err error) bool {
	var se *StorageError
	return As(err, &se) && se.Type == "S3Exception"
}

// IsDynamoDBError checks if an error is a DynamoDB-specific StorageError
func IsDynamoDBError(err error) bool {
	var se *StorageError
	return As(err, &se) && se.Type == "DynamoDBException"
}

// IsRetryableError checks if a storage error is retryable
func IsRetryableError(err error) bool {
	var se *StorageError
	return As(err, &se) && se.Retryable
}