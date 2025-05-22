package s3state

import (
	"fmt"
	"strings"
)

// StateError represents an error in state management operations
type StateError struct {
	Type      string
	Operation string
	Message   string
	Cause     error
}

// Error implements the error interface
func (e *StateError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s operation '%s': %s (caused by: %v)", e.Type, e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s operation '%s': %s", e.Type, e.Operation, e.Message)
}

// Unwrap returns the underlying error
func (e *StateError) Unwrap() error {
	return e.Cause
}

// Error types
const (
	ErrorTypeS3Operation   = "S3_OPERATION_ERROR"
	ErrorTypeValidation    = "VALIDATION_ERROR"
	ErrorTypeJSONOperation = "JSON_OPERATION_ERROR"
	ErrorTypeReference     = "REFERENCE_ERROR"
	ErrorTypeCategory      = "CATEGORY_ERROR"
	ErrorTypeInternal      = "INTERNAL_ERROR"
)

// NewS3Error creates a new S3 operation error
func NewS3Error(operation, message string, cause error) *StateError {
	return &StateError{
		Type:      ErrorTypeS3Operation,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(operation, message string) *StateError {
	return &StateError{
		Type:      ErrorTypeValidation,
		Operation: operation,
		Message:   message,
	}
}

// NewJSONError creates a new JSON operation error
func NewJSONError(operation, message string, cause error) *StateError {
	return &StateError{
		Type:      ErrorTypeJSONOperation,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewReferenceError creates a new reference error
func NewReferenceError(operation, message string) *StateError {
	return &StateError{
		Type:      ErrorTypeReference,
		Operation: operation,
		Message:   message,
	}
}

// NewCategoryError creates a new category error
func NewCategoryError(operation, message string) *StateError {
	return &StateError{
		Type:      ErrorTypeCategory,
		Operation: operation,
		Message:   message,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(operation, message string, cause error) *StateError {
	return &StateError{
		Type:      ErrorTypeInternal,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// IsS3Error checks if an error is an S3 operation error
func IsS3Error(err error) bool {
	if stateErr, ok := err.(*StateError); ok {
		return stateErr.Type == ErrorTypeS3Operation
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if stateErr, ok := err.(*StateError); ok {
		return stateErr.Type == ErrorTypeValidation
	}
	return false
}

// IsJSONError checks if an error is a JSON operation error
func IsJSONError(err error) bool {
	if stateErr, ok := err.(*StateError); ok {
		return stateErr.Type == ErrorTypeJSONOperation
	}
	return false
}

// IsReferenceError checks if an error is a reference error
func IsReferenceError(err error) bool {
	if stateErr, ok := err.(*StateError); ok {
		return stateErr.Type == ErrorTypeReference
	}
	return false
}

// IsCategoryError checks if an error is a category error
func IsCategoryError(err error) bool {
	if stateErr, ok := err.(*StateError); ok {
		return stateErr.Type == ErrorTypeCategory
	}
	return false
}

// IsRetryable determines if an error is likely transient and retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common retryable error patterns
	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection",
		"network",
		"throttl",
		"rate limit",
		"service unavailable",
		"internal error",
		"temporary",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// WrapError wraps an existing error with additional context
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	if stateErr, ok := err.(*StateError); ok {
		// If it's already a StateError, just update the operation if empty
		if stateErr.Operation == "" {
			stateErr.Operation = operation
		}
		return stateErr
	}
	
	// Wrap non-StateError with appropriate type
	return NewInternalError(operation, "operation failed", err)
}

// ValidateReference checks if a reference is valid and returns appropriate error
func ValidateReference(ref *Reference, operation string) error {
	if ref == nil {
		return NewReferenceError(operation, "reference is nil")
	}
	
	if ref.Bucket == "" {
		return NewReferenceError(operation, "bucket name is required")
	}
	
	if ref.Key == "" {
		return NewReferenceError(operation, "key is required")
	}
	
	return nil
}

// ValidateCategory checks if a category is valid and returns appropriate error
func ValidateCategory(category, operation string) error {
	if category == "" {
		return NewCategoryError(operation, "category is required")
	}
	
	if !IsValidCategory(category) {
		return NewCategoryError(operation, fmt.Sprintf("invalid category: %s", category))
	}
	
	return nil
}

// ValidateEnvelope checks if an envelope is valid and returns appropriate error
func ValidateEnvelope(envelope *Envelope, operation string) error {
	if envelope == nil {
		return NewValidationError(operation, "envelope is nil")
	}
	
	if envelope.VerificationID == "" {
		return NewValidationError(operation, "verification ID is required")
	}
	
	// Validate all references in the envelope
	for name, ref := range envelope.References {
		if err := ValidateReference(ref, operation); err != nil {
			return NewValidationError(operation, fmt.Sprintf("invalid reference '%s': %v", name, err))
		}
	}
	
	return nil
}

// ErrorList holds multiple errors
type ErrorList struct {
	Errors []error
}

// Error implements the error interface for ErrorList
func (el *ErrorList) Error() string {
	if len(el.Errors) == 0 {
		return "no errors"
	}
	
	if len(el.Errors) == 1 {
		return el.Errors[0].Error()
	}
	
	var messages []string
	for _, err := range el.Errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("multiple errors: [%s]", strings.Join(messages, "; "))
}

// Add adds an error to the list if it's not nil
func (el *ErrorList) Add(err error) {
	if err != nil {
		el.Errors = append(el.Errors, err)
	}
}

// HasErrors returns true if there are any errors in the list
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// First returns the first error in the list, or nil if empty
func (el *ErrorList) First() error {
	if len(el.Errors) > 0 {
		return el.Errors[0]
	}
	return nil
}

// ToError returns the ErrorList as an error interface, or nil if no errors
func (el *ErrorList) ToError() error {
	if !el.HasErrors() {
		return nil
	}
	return el
}

// NewErrorList creates a new ErrorList
func NewErrorList() *ErrorList {
	return &ErrorList{
		Errors: make([]error, 0),
	}
}