package main

import (
	"errors"
	"fmt"
	"strings"
	"workflow-function/shared/schema"
)

// FetchImagesRequest represents the expected input to the Lambda function.
type FetchImagesRequest struct {
	VerificationContext *schema.VerificationContext `json:"verificationContext,omitempty"`
	// Optional historical context from previous step
	HistoricalContext   map[string]interface{}     `json:"historicalContext,omitempty"`
	// Schema version for compatibility
	SchemaVersion       string                     `json:"schemaVersion,omitempty"`
	
	// Direct fields for backward compatibility or direct invocation
	VerificationId        string `json:"verificationId"`
	VerificationType      string `json:"verificationType"`
	ReferenceImageUrl     string `json:"referenceImageUrl"`
	CheckingImageUrl      string `json:"checkingImageUrl"`
	LayoutId              int    `json:"layoutId,omitempty"`
	LayoutPrefix          string `json:"layoutPrefix,omitempty"`
	PreviousVerificationId string `json:"previousVerificationId,omitempty"`
	VendingMachineId      string `json:"vendingMachineId,omitempty"`
}

// ImageMetadata holds S3 object metadata.
type ImageMetadata struct {
	ContentType   string `json:"contentType"`
	Size          int64  `json:"size"`
	LastModified  string `json:"lastModified"`
	ETag          string `json:"etag"`
	BucketOwner   string `json:"bucketOwner"`
	Bucket        string `json:"bucket"`
	Key           string `json:"key"`
}

// ImagesData contains metadata for both reference and checking images
type ImagesData struct {
	ReferenceImageMeta  ImageMetadata `json:"referenceImageMeta"`
	CheckingImageMeta   ImageMetadata `json:"checkingImageMeta"`
}

// FetchImagesResponse represents the Lambda output.
type FetchImagesResponse struct {
	VerificationContext schema.VerificationContext `json:"verificationContext"`
	Images              ImagesData                 `json:"images"`
	LayoutMetadata      map[string]interface{}     `json:"layoutMetadata,omitempty"`
	HistoricalContext   map[string]interface{}     `json:"historicalContext"` // Always include this field
}

// ParallelFetchResults holds the results of parallel fetches.
type ParallelFetchResults struct {
	ReferenceMeta     ImageMetadata
	CheckingMeta      ImageMetadata
	LayoutMeta        map[string]interface{}
	HistoricalContext map[string]interface{}
	Errors            []error
}

// Validate checks for required fields and basic format.
func (r *FetchImagesRequest) Validate() error {
	// If we have a verificationContext object, validate it
	if r.VerificationContext != nil {
		if r.VerificationContext.VerificationId == "" {
			return errors.New("verificationContext.verificationId is required")
		}
		if r.VerificationContext.VerificationType == "" {
			return errors.New("verificationContext.verificationType is required")
		}
		if r.VerificationContext.ReferenceImageUrl == "" {
			return errors.New("verificationContext.referenceImageUrl is required")
		}
		if r.VerificationContext.CheckingImageUrl == "" {
			return errors.New("verificationContext.checkingImageUrl is required")
		}

		// Validate verification type
		switch r.VerificationContext.VerificationType {
		case schema.VerificationTypeLayoutVsChecking:
			// For LAYOUT_VS_CHECKING, we need layoutId and layoutPrefix
			if r.VerificationContext.LayoutId == 0 {
				return errors.New("verificationContext.layoutId is required for LAYOUT_VS_CHECKING verification type")
			}
			if r.VerificationContext.LayoutPrefix == "" {
				return errors.New("verificationContext.layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
			}
			// previousVerificationId is not required for LAYOUT_VS_CHECKING
		case schema.VerificationTypePreviousVsCurrent:
			// For PREVIOUS_VS_CURRENT, we need previousVerificationId
			if r.VerificationContext.PreviousVerificationId == "" {
				return errors.New("verificationContext.previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
			}
			// Validate that referenceImageUrl is from the checking bucket
			if !strings.Contains(r.VerificationContext.ReferenceImageUrl, "checking") {
				return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
			}
		default:
			return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
				r.VerificationContext.VerificationType)
		}

		// Validation passed
		return nil
	}

	// Otherwise validate direct fields
	if r.VerificationId == "" {
		return errors.New("verificationId is required")
	}
	if r.VerificationType == "" {
		return errors.New("verificationType is required")
	}
	if r.ReferenceImageUrl == "" {
		return errors.New("referenceImageUrl is required")
	}
	if r.CheckingImageUrl == "" {
		return errors.New("checkingImageUrl is required")
	}

	// Validate verification type
	switch r.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		// For LAYOUT_VS_CHECKING, we need layoutId and layoutPrefix
		if r.LayoutId == 0 {
			return errors.New("layoutId is required for LAYOUT_VS_CHECKING verification type")
		}
		if r.LayoutPrefix == "" {
			return errors.New("layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
		}
		// previousVerificationId is not required for LAYOUT_VS_CHECKING
	case schema.VerificationTypePreviousVsCurrent:
		// For PREVIOUS_VS_CURRENT, we need previousVerificationId
		if r.PreviousVerificationId == "" {
			return errors.New("previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
		}
		// Validate that referenceImageUrl is from the checking bucket
		if !strings.Contains(r.ReferenceImageUrl, "checking") {
			return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
		}
	default:
		return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
			r.VerificationType)
	}
	
	return nil
}

// Error helpers
func NewBadRequestError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func NewNotFoundError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}