package main

import (
	"context"
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

// Storage method constants
const (
	// StorageMethodInline indicates Base64 data is stored inline in the response
	StorageMethodInline      = "inline"
	// StorageMethodS3Temporary indicates Base64 data is stored in S3 temporarily
	StorageMethodS3Temporary = "s3-temporary"
)

// ImageMetadata holds S3 object metadata and Base64 encoded image data with hybrid storage support.
type ImageMetadata struct {
	ContentType   string `json:"contentType"`
	Size          int64  `json:"size"`
	LastModified  string `json:"lastModified"`
	ETag          string `json:"etag"`
	Bucket        string `json:"bucket"`
	Key           string `json:"key"`
	
	// Storage method indicator
	StorageMethod string `json:"storageMethod"` // "inline" or "s3-temporary"
	
	// Base64 fields for Bedrock integration
	Base64Data    string                 `json:"base64Data,omitempty"`     // For inline storage
	ImageFormat   string                 `json:"imageFormat,omitempty"`    
	BedrockFormat map[string]interface{} `json:"bedrockFormat,omitempty"`  
	
	// For S3 temporary storage
	Base64S3Bucket string `json:"base64S3Bucket,omitempty"` // S3 bucket for Base64 data
	Base64S3Key    string `json:"base64S3Key,omitempty"`    // S3 key for Base64 data
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
	base64Data, err := im.GetBase64Data(context.Background(), nil)
	if err != nil || base64Data == "" {
		return nil
	}
	
	return map[string]interface{}{
		"image": map[string]interface{}{
			"format": im.GetImageFormat(),
			"source": map[string]interface{}{
				"bytes": base64Data,
			},
		},
	}
}

// HasBase64Data checks if Base64 data is available (either inline or in S3)
func (im *ImageMetadata) HasBase64Data() bool {
	return im.HasInlineData() || im.HasS3Storage()
}

// HasInlineData checks if inline Base64 data is available
func (im *ImageMetadata) HasInlineData() bool {
	return im.StorageMethod == StorageMethodInline && im.Base64Data != ""
}

// HasS3Storage checks if S3 temporary storage references are available
func (im *ImageMetadata) HasS3Storage() bool {
	return im.StorageMethod == StorageMethodS3Temporary && 
		   im.Base64S3Bucket != "" && 
		   im.Base64S3Key != ""
}

// GetBase64Data retrieves Base64 data based on storage method
func (im *ImageMetadata) GetBase64Data(ctx context.Context, s3Utils S3Base64Retriever) (string, error) {
	switch im.StorageMethod {
	case StorageMethodInline:
		if im.Base64Data == "" {
			return "", fmt.Errorf("inline Base64 data not available")
		}
		return im.Base64Data, nil
		
	case StorageMethodS3Temporary:
		if im.Base64S3Bucket == "" || im.Base64S3Key == "" {
			return "", fmt.Errorf("S3 temporary storage references not available")
		}
		if s3Utils == nil {
			return "", fmt.Errorf("S3 utilities not provided for temporary storage retrieval")
		}
		return s3Utils.RetrieveBase64FromS3(ctx, im.Base64S3Bucket, im.Base64S3Key)
		
	default:
		return "", fmt.Errorf("unknown storage method: %s", im.StorageMethod)
	}
}

// UpdateImageFormat sets the image format from content type
func (im *ImageMetadata) UpdateImageFormat() {
	im.ImageFormat = im.GetImageFormat()
}

// UpdateBedrockFormat updates the Bedrock format using current Base64 data
func (im *ImageMetadata) UpdateBedrockFormat() {
	im.BedrockFormat = im.CreateBedrockBytesFormat()
}

// ValidateStorageMethod validates the storage method and required fields
func (im *ImageMetadata) ValidateStorageMethod() error {
	switch im.StorageMethod {
	case StorageMethodInline:
		if im.Base64Data == "" {
			return fmt.Errorf("inline storage method requires base64Data field")
		}
		// Clear S3 fields for inline storage
		im.Base64S3Bucket = ""
		im.Base64S3Key = ""
		
	case StorageMethodS3Temporary:
		if im.Base64S3Bucket == "" {
			return fmt.Errorf("s3-temporary storage method requires base64S3Bucket field")
		}
		if im.Base64S3Key == "" {
			return fmt.Errorf("s3-temporary storage method requires base64S3Key field")
		}
		// Clear inline data for S3 storage
		im.Base64Data = ""
		
	case "":
		return fmt.Errorf("storage method is required")
		
	default:
		return fmt.Errorf("invalid storage method: %s (must be 'inline' or 's3-temporary')", im.StorageMethod)
	}
	
	return nil
}

// SetInlineStorage configures the metadata for inline Base64 storage
func (im *ImageMetadata) SetInlineStorage(base64Data string) {
	im.StorageMethod = StorageMethodInline
	im.Base64Data = base64Data
	im.Base64S3Bucket = ""
	im.Base64S3Key = ""
	im.UpdateImageFormat()
	im.UpdateBedrockFormat()
}

// SetS3Storage configures the metadata for S3 temporary storage
func (im *ImageMetadata) SetS3Storage(bucket, key string) {
	im.StorageMethod = StorageMethodS3Temporary
	im.Base64S3Bucket = bucket
	im.Base64S3Key = key
	im.Base64Data = ""
	im.UpdateImageFormat()
	// Note: BedrockFormat will be updated when data is retrieved
}

// GetStorageInfo returns detailed information about the storage method
func (im *ImageMetadata) GetStorageInfo() map[string]interface{} {
	info := map[string]interface{}{
		"storageMethod": im.StorageMethod,
		"hasBase64Data": im.HasBase64Data(),
		"hasInlineData": im.HasInlineData(),
		"hasS3Storage":  im.HasS3Storage(),
	}
	
	if im.HasInlineData() {
		info["inlineDataLength"] = len(im.Base64Data)
	}
	
	if im.HasS3Storage() {
		info["s3Bucket"] = im.Base64S3Bucket
		info["s3Key"] = im.Base64S3Key
	}
	
	return info
}

// S3Base64Retriever interface for retrieving Base64 data from S3
type S3Base64Retriever interface {
	RetrieveBase64FromS3(ctx context.Context, bucket, key string) (string, error)
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

// Base64RetrievalError represents errors during Base64 data retrieval
type Base64RetrievalError struct {
	StorageMethod string
	Message       string
	Err           error
}

func (e *Base64RetrievalError) Error() string {
	return fmt.Sprintf("Base64 retrieval failed [%s]: %s: %v", e.StorageMethod, e.Message, e.Err)
}

// NewBase64RetrievalError creates a new Base64 retrieval error
func NewBase64RetrievalError(method, message string, err error) error {
	return &Base64RetrievalError{
		StorageMethod: method,
		Message:       message,
		Err:           err,
	}
}

// InvalidStorageMethodError represents errors with invalid storage methods
type InvalidStorageMethodError struct {
	StorageMethod string
	Message       string
}

func (e *InvalidStorageMethodError) Error() string {
	return fmt.Sprintf("Invalid storage method [%s]: %s", e.StorageMethod, e.Message)
}

// NewInvalidStorageMethodError creates a new invalid storage method error
func NewInvalidStorageMethodError(method, message string) error {
	return &InvalidStorageMethodError{
		StorageMethod: method,
		Message:       message,
	}
}