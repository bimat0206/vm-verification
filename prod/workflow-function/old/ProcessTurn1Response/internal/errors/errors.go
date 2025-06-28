// Package errors provides custom error handling for the ProcessTurn1Response Lambda
package errors

import (
	"errors"
	"fmt"
	"strings"
	
	"log/slog"
)

// Common error codes used across the application
const (
	// Input validation error codes
	ErrInvalidInput         = "INVALID_INPUT"
	ErrMissingField         = "MISSING_FIELD"
	ErrInvalidFormat        = "INVALID_FORMAT"
	
	// Processing error codes
	ErrProcessingFailed     = "PROCESSING_FAILED"
	ErrParsingFailed        = "PARSING_FAILED"
	ErrExtractionFailed     = "EXTRACTION_FAILED"
	ErrValidationFailed     = "VALIDATION_FAILED"
	
	// State management error codes
	ErrStateLoadFailed      = "STATE_LOAD_FAILED"
	ErrStateStoreFailed     = "STATE_STORE_FAILED"
	ErrReferenceInvalid     = "REFERENCE_INVALID"
	ErrEnvelopeInvalid      = "ENVELOPE_INVALID"
	
	// System error codes
	ErrInternalError        = "INTERNAL_ERROR"
	ErrDependencyFailed     = "DEPENDENCY_FAILED"
	ErrTimeout              = "TIMEOUT"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// Error categories for grouping and handling
	CategoryInput    ErrorCategory = "INPUT"
	CategoryProcess  ErrorCategory = "PROCESS"
	CategoryState    ErrorCategory = "STATE"
	CategorySystem   ErrorCategory = "SYSTEM"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	// Error severity levels for logging and reporting
	SeverityDebug    ErrorSeverity = "DEBUG"
	SeverityInfo     ErrorSeverity = "INFO"
	SeverityWarning  ErrorSeverity = "WARNING"
	SeverityError    ErrorSeverity = "ERROR"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

// FunctionError represents a custom error for the Lambda function
type FunctionError struct {
	// Operation that was being performed when the error occurred
	Operation string
	
	// Category of the error
	Category ErrorCategory
	
	// Code is a stable identifier for the error type
	Code string
	
	// Message is a human-readable description of the error
	Message string
	
	// Details contains additional information about the error
	Details map[string]interface{}
	
	// Severity indicates the error severity
	Severity ErrorSeverity
	
	// Wrapped contains the underlying error
	Wrapped error
}

// Error implements the error interface
func (e *FunctionError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("[%s:%s] %s in %s: %v", 
			e.Category, e.Code, e.Message, e.Operation, e.Wrapped)
	}
	return fmt.Sprintf("[%s:%s] %s in %s", 
		e.Category, e.Code, e.Message, e.Operation)
}

// Unwrap returns the underlying error
func (e *FunctionError) Unwrap() error {
	return e.Wrapped
}

// New creates a new FunctionError
func New(operation string, category ErrorCategory, code, message string, severity ErrorSeverity, wrapped error) *FunctionError {
	return &FunctionError{
		Operation: operation,
		Category:  category,
		Code:      code,
		Message:   message,
		Details:   make(map[string]interface{}),
		Severity:  severity,
		Wrapped:   wrapped,
	}
}

// NewWithDetails creates a new FunctionError with additional details
func NewWithDetails(operation string, category ErrorCategory, code, message string, 
	severity ErrorSeverity, details map[string]interface{}, wrapped error) *FunctionError {
	err := New(operation, category, code, message, severity, wrapped)
	if details != nil {
		err.Details = details
	}
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	// If it's already a FunctionError, just update the operation
	if fe, ok := err.(*FunctionError); ok {
		newErr := *fe
		newErr.Operation = fmt.Sprintf("%s.%s", operation, fe.Operation)
		return &newErr
	}
	
	// Otherwise, wrap it as an internal error
	return New(operation, CategorySystem, ErrInternalError, 
		err.Error(), SeverityError, err)
}

// InputError creates an input validation error
func InputError(operation, message string, wrapped error) error {
	return New(operation, CategoryInput, ErrInvalidInput, 
		message, SeverityWarning, wrapped)
}

// ProcessingError creates a processing error
func ProcessingError(operation, message string, wrapped error) error {
	return New(operation, CategoryProcess, ErrProcessingFailed, 
		message, SeverityError, wrapped)
}

// ParsingError creates a parsing error
func ParsingError(operation, message string, wrapped error) error {
	return New(operation, CategoryProcess, ErrParsingFailed, 
		message, SeverityError, wrapped)
}

// ValidationError creates a validation error
func ValidationError(operation, message string, details map[string]interface{}) error {
	return NewWithDetails(operation, CategoryInput, ErrValidationFailed, 
		message, SeverityWarning, details, nil)
}

// StateError creates a state management error
func StateError(operation, code, message string, wrapped error) error {
	return New(operation, CategoryState, code, 
		message, SeverityError, wrapped)
}

// SystemError creates a system error
func SystemError(operation, message string, wrapped error) error {
	return New(operation, CategorySystem, ErrInternalError, 
		message, SeverityCritical, wrapped)
}

// ReferenceError creates a reference error
func ReferenceError(operation, message string) error {
	return New(operation, CategoryState, ErrReferenceInvalid, 
		message, SeverityError, nil)
}

// Is reports whether any error in the error chain matches the target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in the error chain that matches the target type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetCategory returns the error category of a FunctionError
func GetCategory(err error) ErrorCategory {
	var fe *FunctionError
	if errors.As(err, &fe) {
		return fe.Category
	}
	return CategorySystem
}

// GetCode returns the error code of a FunctionError
func GetCode(err error) string {
	var fe *FunctionError
	if errors.As(err, &fe) {
		return fe.Code
	}
	return ErrInternalError
}

// ToMap converts an error to a map for response formatting
func ToMap(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	
	var fe *FunctionError
	if errors.As(err, &fe) {
		result := map[string]interface{}{
			"code":      fe.Code,
			"message":   fe.Message,
			"category":  string(fe.Category),
			"operation": fe.Operation,
		}
		
		if len(fe.Details) > 0 {
			result["details"] = fe.Details
		}
		
		return result
	}
	
	// Default case for non-FunctionError
	return map[string]interface{}{
		"code":    ErrInternalError,
		"message": err.Error(),
	}
}

// LogError logs an error with appropriate severity
func LogError(log *slog.Logger, err error) {
	if err == nil {
		return
	}
	
	var fe *FunctionError
	if errors.As(err, &fe) {
		// Extract details to add to the log entry
		logAttrs := make([]interface{}, 0, 2*len(fe.Details)+6)
		logAttrs = append(logAttrs,
			"code", fe.Code,
			"operation", fe.Operation,
			"category", string(fe.Category),
		)
		
		for k, v := range fe.Details {
			logAttrs = append(logAttrs, k, v)
		}
		
		// Log with the appropriate severity
		switch fe.Severity {
		case SeverityDebug:
			log.Debug(fe.Message, logAttrs...)
		case SeverityInfo:
			log.Info(fe.Message, logAttrs...)
		case SeverityWarning:
			log.Warn(fe.Message, logAttrs...)
		case SeverityCritical:
			log.Error(fe.Message, logAttrs...)
		default:
			log.Error(fe.Message, logAttrs...)
		}
	} else {
		// Log non-FunctionError as error
		log.Error(err.Error())
	}
}

// ErrorList represents a collection of errors
type ErrorList struct {
	errors []error // lowercase to avoid conflict with method name
}

// NewErrorList creates a new error list
func NewErrorList() *ErrorList {
	return &ErrorList{
		errors: make([]error, 0),
	}
}

// Add adds an error to the error list
func (el *ErrorList) Add(err error) {
	if err != nil {
		el.errors = append(el.errors, err)
	}
}

// ToError returns nil if the error list is empty, otherwise returns an error
func (el *ErrorList) ToError() error {
	if len(el.errors) == 0 {
		return nil
	}
	
	messages := make([]string, 0, len(el.errors))
	for _, err := range el.errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Errorf("multiple errors: %s", strings.Join(messages, "; "))
}

// Errors returns the list of errors
func (el *ErrorList) Errors() []error {
	return el.errors
}

// HasErrors returns true if the error list contains errors
func (el *ErrorList) HasErrors() bool {
	return len(el.errors) > 0
}

// Count returns the number of errors in the list
func (el *ErrorList) Count() int {
	return len(el.errors)
}

// FormatErrors creates a user-friendly formatted error message
func FormatErrors(errs []error) string {
	if len(errs) == 0 {
		return ""
	}
	
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	
	if len(errs) == 1 {
		return messages[0]
	}
	
	return strings.Join(messages, "; ")
}

// WrapWithWorkflow wraps an error with workflow context 
func WrapWithWorkflow(operation string, err error, workflowID string) error {
	details := map[string]interface{}{
		"workflowId": workflowID,
	}
	
	var fe *FunctionError
	if errors.As(err, &fe) {
		newErr := *fe
		newErr.Operation = fmt.Sprintf("%s.%s", operation, fe.Operation)
		for k, v := range details {
			newErr.Details[k] = v
		}
		return &newErr
	}
	
	return NewWithDetails(
		operation,
		CategorySystem,
		ErrInternalError,
		err.Error(),
		SeverityError,
		details,
		err,
	)
}