package main

import (
	"encoding/json"
	"fmt"
)

// StandardError represents a standardized error response
type StandardError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error returns the error message
func (e StandardError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// MarshalJSON marshals the error to JSON
func (e StandardError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	}{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	})
}

// NewResourceNotFoundError creates a new error for resource not found
func NewResourceNotFoundError(resource, resourceID string) *StandardError {
	return &StandardError{
		Code:    "ResourceNotFoundError",
		Message: fmt.Sprintf("%s not found: %s", resource, resourceID),
		Details: map[string]string{
			"resourceType": resource,
			"resourceId":   resourceID,
		},
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]string) *StandardError {
	return &StandardError{
		Code:    "ValidationError",
		Message: message,
		Details: details,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *StandardError {
	return &StandardError{
		Code:    "InternalError",
		Message: message,
	}
}