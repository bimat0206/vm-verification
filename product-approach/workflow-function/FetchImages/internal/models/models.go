// Package models provides data models for the FetchImages function
package models

import (
	"errors"
	"fmt"
	"strings"
	
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// FetchImagesRequest represents the request format for the FetchImages function
type FetchImagesRequest struct {
	// Core fields for S3 state management
	VerificationId string                          `json:"verificationId"`
	S3References   map[string]*s3state.Reference   `json:"s3References"`
	Status         string                          `json:"status"`

	// Legacy fields for backward compatibility
	VerificationContext *schema.VerificationContext `json:"verificationContext,omitempty"`
	SchemaVersion       string                     `json:"schemaVersion,omitempty"`
}

// FetchImagesResponse represents the response format for the FetchImages function
type FetchImagesResponse struct {
	// Core fields for S3 state management
	VerificationId string                          `json:"verificationId"`
	S3References   map[string]*s3state.Reference   `json:"s3References"`
	Status         string                          `json:"status"`
	Summary        map[string]interface{}          `json:"summary,omitempty"`
}

// ImageMetadata represents the metadata for both reference and checking images
type ImageMetadata struct {
	Reference *schema.ImageInfo `json:"reference"`
	Checking  *schema.ImageInfo `json:"checking"`
}

// Validate checks for required fields and valid format
func (r *FetchImagesRequest) Validate() error {
	// Validate basic request fields
	if r.VerificationId == "" {
		return errors.New("verificationId is required")
	}

	// Ensure we have S3 references
	if r.S3References == nil || len(r.S3References) == 0 {
		// Check for legacy format
		if r.VerificationContext != nil {
			// Validate verification context for legacy format
			if err := validateVerificationContext(r.VerificationContext); err != nil {
				return err
			}
		} else {
			return errors.New("s3References is required but missing or empty")
		}
	}

	// Check for initialization reference
	if ref := r.S3References["processing_initialization"]; ref == nil {
		// If using legacy format, this is ok
		if r.VerificationContext == nil {
			return errors.New("initialization reference is required in s3References")
		}
	}

	return nil
}

// validateVerificationContext validates the legacy verification context format
func validateVerificationContext(vc *schema.VerificationContext) error {
	if vc == nil {
		return errors.New("verificationContext is nil")
	}

	if vc.VerificationId == "" {
		return errors.New("verificationContext.verificationId is required")
	}

	if vc.VerificationType == "" {
		return errors.New("verificationContext.verificationType is required")
	}

	if vc.ReferenceImageUrl == "" {
		return errors.New("verificationContext.referenceImageUrl is required")
	}

	if vc.CheckingImageUrl == "" {
		return errors.New("verificationContext.checkingImageUrl is required")
	}

	// Validate verification type
	switch vc.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		if vc.LayoutId == 0 {
			return errors.New("verificationContext.layoutId is required for LAYOUT_VS_CHECKING verification type")
		}
		if vc.LayoutPrefix == "" {
			return errors.New("verificationContext.layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
		}
	case schema.VerificationTypePreviousVsCurrent:
		if vc.PreviousVerificationId == "" {
			return errors.New("verificationContext.previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
		}
		if !strings.Contains(vc.ReferenceImageUrl, "checking") {
			return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
		}
	default:
		return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
			vc.VerificationType)
	}

	return nil
}

// Error types for standard error reporting

// ValidationError represents input validation errors
type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *ValidationError {
	return &ValidationError{
		Message: message,
		Err:     err,
	}
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Message string
	Err     error
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) *NotFoundError {
	return &NotFoundError{
		Message: message,
		Err:     err,
	}
}

// ProcessingError represents errors during processing
type ProcessingError struct {
	Message string
	Err     error
}

func (e *ProcessingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewProcessingError creates a new processing error
func NewProcessingError(message string, err error) *ProcessingError {
	return &ProcessingError{
		Message: message,
		Err:     err,
	}
}