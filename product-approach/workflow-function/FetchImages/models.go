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
	HistoricalContext   map[string]interface{}     `json:"historicalContext,omitempty"`
	SchemaVersion       string                     `json:"schemaVersion,omitempty"`
	
	// Direct fields for backward compatibility
	VerificationId         string `json:"verificationId"`
	VerificationType       string `json:"verificationType"`
	ReferenceImageUrl      string `json:"referenceImageUrl"`
	CheckingImageUrl       string `json:"checkingImageUrl"`
	LayoutId               int    `json:"layoutId,omitempty"`
	LayoutPrefix           string `json:"layoutPrefix,omitempty"`
	PreviousVerificationId string `json:"previousVerificationId,omitempty"`
	VendingMachineId       string `json:"vendingMachineId,omitempty"`
}

// ImageMetadata holds S3 object metadata and Base64 encoded image data.
type ImageMetadata struct {
	ContentType   string `json:"contentType"`
	Size          int64  `json:"size"`
	LastModified  string `json:"lastModified"`
	ETag          string `json:"etag"`
	Bucket        string `json:"bucket"`
	Key           string `json:"key"`
	// Base64 fields for Bedrock integration
	Base64Data    string                 `json:"base64Data,omitempty"`     
	ImageFormat   string                 `json:"imageFormat,omitempty"`    
	BedrockFormat map[string]interface{} `json:"bedrockFormat,omitempty"`  
}

// GetImageFormat extracts the image format from content type
func (im *ImageMetadata) GetImageFormat() string {
	switch im.ContentType {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpeg"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}

// CreateBedrockBytesFormat creates the Bedrock-compatible bytes format
func (im *ImageMetadata) CreateBedrockBytesFormat() map[string]interface{} {
	if im.Base64Data == "" {
		return nil
	}
	
	return map[string]interface{}{
		"image": map[string]interface{}{
			"format": im.GetImageFormat(),
			"source": map[string]interface{}{
				"bytes": im.Base64Data,
			},
		},
	}
}

// HasBase64Data checks if Base64 data is available
func (im *ImageMetadata) HasBase64Data() bool {
	return im.Base64Data != ""
}

// UpdateImageFormat sets the image format from content type
func (im *ImageMetadata) UpdateImageFormat() {
	im.ImageFormat = im.GetImageFormat()
}

// UpdateBedrockFormat updates the Bedrock format using current Base64 data
func (im *ImageMetadata) UpdateBedrockFormat() {
	im.BedrockFormat = im.CreateBedrockBytesFormat()
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
	HistoricalContext   map[string]interface{}     `json:"historicalContext"`
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
			if r.VerificationContext.LayoutId == 0 {
				return errors.New("verificationContext.layoutId is required for LAYOUT_VS_CHECKING verification type")
			}
			if r.VerificationContext.LayoutPrefix == "" {
				return errors.New("verificationContext.layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
			}
		case schema.VerificationTypePreviousVsCurrent:
			if r.VerificationContext.PreviousVerificationId == "" {
				return errors.New("verificationContext.previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
			}
			if !strings.Contains(r.VerificationContext.ReferenceImageUrl, "checking") {
				return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
			}
		default:
			return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
				r.VerificationContext.VerificationType)
		}
		return nil
	}

	// Validate direct fields
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

	switch r.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		if r.LayoutId == 0 {
			return errors.New("layoutId is required for LAYOUT_VS_CHECKING verification type")
		}
		if r.LayoutPrefix == "" {
			return errors.New("layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
		}
	case schema.VerificationTypePreviousVsCurrent:
		if r.PreviousVerificationId == "" {
			return errors.New("previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
		}
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