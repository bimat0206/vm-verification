package errors

import "fmt"

// AppError represents an application error
type AppError struct {
	Type    string
	Message string
	Cause   error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	AppError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		AppError: AppError{
			Type:    "ValidationError",
			Message: message,
		},
	}
}

// ResourceNotFoundError represents a resource not found error
type ResourceNotFoundError struct {
	AppError
}

// NewResourceNotFoundError creates a new resource not found error
func NewResourceNotFoundError(message string) *ResourceNotFoundError {
	return &ResourceNotFoundError{
		AppError: AppError{
			Type:    "ResourceNotFoundError",
			Message: message,
		},
	}
}

// ServiceError represents a service error
type ServiceError struct {
	AppError
}

// NewServiceError creates a new service error
func NewServiceError(message string, cause error) *ServiceError {
	return &ServiceError{
		AppError: AppError{
			Type:    "ServiceError",
			Message: message,
			Cause:   cause,
		},
	}
}

// BedrockError represents a Bedrock API error
type BedrockError struct {
	AppError
	Retryable bool
}

// NewBedrockError creates a new Bedrock error
func NewBedrockError(message string, cause error, retryable bool) *BedrockError {
	return &BedrockError{
		AppError: AppError{
			Type:    "BedrockError",
			Message: message,
			Cause:   cause,
		},
		Retryable: retryable,
	}
}

// S3Error represents an S3 error
type S3Error struct {
	AppError
}

// NewS3Error creates a new S3 error
func NewS3Error(message string, cause error) *S3Error {
	return &S3Error{
		AppError: AppError{
			Type:    "S3Error",
			Message: message,
			Cause:   cause,
		},
	}
}

// DynamoDBError represents a DynamoDB error
type DynamoDBError struct {
	AppError
}

// NewDynamoDBError creates a new DynamoDB error
func NewDynamoDBError(message string, cause error) *DynamoDBError {
	return &DynamoDBError{
		AppError: AppError{
			Type:    "DynamoDBError",
			Message: message,
			Cause:   cause,
		},
	}
}