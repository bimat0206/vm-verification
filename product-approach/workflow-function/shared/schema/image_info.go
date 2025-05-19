// Package schema provides image information structures and methods
package schema

import (
	"fmt"
)

// ImageInfo contains details about a single image with S3 Base64 storage support
// Supports both legacy field names and new unified format
type ImageInfo struct {
	// S3 References (for traceability and logging)
	URL      string `json:"url"`
	S3Key    string `json:"s3Key"`
	S3Bucket string `json:"s3Bucket"`
	
	// Image Properties
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Format      string `json:"format,omitempty"`           // png, jpeg, etc.
	ContentType string `json:"contentType,omitempty"`      // image/png, image/jpeg
	Size        int64  `json:"size,omitempty"`             // Original file size in bytes
	
	Base64Size           int64  `json:"base64Size,omitempty"`           // Size of Base64 string
	
	// S3 Storage for Base64 - supports legacy field names
	Base64S3Bucket       string `json:"base64S3Bucket,omitempty"`       // Temporary bucket for Base64
	Base64S3Key          string `json:"base64S3Key,omitempty"`          // S3 key for Base64 data
	
	// Legacy S3 field names for specific image types
	CheckingImageBase64S3Key string `json:"checkingImageBase64S3Key,omitempty"` // Legacy field for checking image S3 key
	ReferenceImageBase64S3Key string `json:"referenceImageBase64S3Key,omitempty"` // Legacy field for reference image S3 key
	
	Base64S3Metadata     map[string]string `json:"base64S3Metadata,omitempty"` // Additional S3 metadata
	
	// Storage Method Indicators
	StorageMethod        string `json:"storageMethod"`                  // Always "s3-temporary"
	Base64Generated      bool   `json:"base64Generated"`                // Indicates Base64 conversion completed
	StorageDecisionAt    string `json:"storageDecisionAt,omitempty"`    // When storage method was decided
	
	// Retrieval metadata
	LastBase64Access     string `json:"lastBase64Access,omitempty"`     // Last time Base64 was accessed
	Base64AccessCount    int    `json:"base64AccessCount,omitempty"`    // Number of times Base64 was accessed
	
	// Existing Metadata
	LastModified string                 `json:"lastModified,omitempty"`
	ETag         string                 `json:"etag,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EnsureBase64Generated checks if Base64 data is present and valid
func (img *ImageInfo) EnsureBase64Generated() bool {
	return img.Base64Generated && img.Format != "" && img.HasBase64Data()
}

// HasBase64Data checks if Base64 data is available in S3
func (img *ImageInfo) HasBase64Data() bool {
	// Check S3 storage (unified and legacy fields)
	if img.Base64S3Bucket != "" && (img.Base64S3Key != "" || 
		img.CheckingImageBase64S3Key != "" || img.ReferenceImageBase64S3Key != "") {
		return true
	}
	
	return false
}

// GetBase64Data returns an empty string for S3-only storage
// This method is kept for backward compatibility
// Use HybridBase64Retriever.RetrieveBase64Data instead
func (img *ImageInfo) GetBase64Data() string {
	// In S3-only storage, we don't store Base64 data inline
	// This method is kept for backward compatibility
	// and returns an empty string
	return ""
}

// GetBase64S3Key returns the S3 key for Base64 data, handling legacy field names
func (img *ImageInfo) GetBase64S3Key() string {
	// Return unified field if available
	if img.Base64S3Key != "" {
		return img.Base64S3Key
	}
	
	// Fall back to legacy fields
	if img.CheckingImageBase64S3Key != "" {
		return img.CheckingImageBase64S3Key
	}
	
	if img.ReferenceImageBase64S3Key != "" {
		return img.ReferenceImageBase64S3Key
	}
	
	return ""
}

// SetBase64S3Key sets the S3 key for Base64 data using the appropriate field based on image type
func (img *ImageInfo) SetBase64S3Key(s3Key string, imageType string) {
	// Set unified field
	img.Base64S3Key = s3Key
	
	// Also set legacy field for backward compatibility
	switch imageType {
	case "reference":
		img.ReferenceImageBase64S3Key = s3Key
	case "checking":
		img.CheckingImageBase64S3Key = s3Key
	}
}

// GetBase64SizeEstimate estimates the Base64 size
func (img *ImageInfo) GetBase64SizeEstimate() int64 {
	if img.Base64Size > 0 {
		return img.Base64Size
	}
	
	// Base64 encoding increases size by ~33%
	return img.Size * 4 / 3
}

// ValidateBase64Size checks if Base64 data is within reasonable limits
func (img *ImageInfo) ValidateBase64Size() error {
	base64Size := img.GetBase64SizeEstimate()
	if base64Size > BedrockMaxImageSize {
		return fmt.Errorf("Base64 data size (%d bytes) exceeds Bedrock limit (%d bytes)", 
			base64Size, BedrockMaxImageSize)
	}
	return nil
}

// IsS3TemporaryStorage returns true if image uses S3 temporary storage
func (img *ImageInfo) IsS3TemporaryStorage() bool {
	return img.StorageMethod == StorageMethodS3Temporary
}

// IsInlineStorage returns false for S3-only storage
// This method is kept for backward compatibility
func (img *ImageInfo) IsInlineStorage() bool {
	// In S3-only storage, we don't store Base64 data inline
	// This method is kept for backward compatibility
	return false
}

// GetStorageInfo returns storage method information
func (img *ImageInfo) GetStorageInfo() map[string]interface{} {
	info := map[string]interface{}{
		"storageMethod":    img.StorageMethod,
		"base64Generated":  img.Base64Generated,
		"base64Size":       img.GetBase64SizeEstimate(),
		"hasBase64Data":    img.HasBase64Data(),
	}
	
	// S3 storage information (supporting legacy fields)
	info["base64S3Bucket"] = img.Base64S3Bucket
	info["base64S3Key"] = img.GetBase64S3Key() // Uses helper that handles legacy fields
	info["base64S3Metadata"] = img.Base64S3Metadata
	
	// Legacy S3 field information
	if img.CheckingImageBase64S3Key != "" {
		info["checkingImageBase64S3Key"] = img.CheckingImageBase64S3Key
	}
	if img.ReferenceImageBase64S3Key != "" {
		info["referenceImageBase64S3Key"] = img.ReferenceImageBase64S3Key
	}
	
	if img.StorageDecisionAt != "" {
		info["storageDecisionAt"] = img.StorageDecisionAt
	}
	
	return info
}

// UpdateLastAccess updates the last access information for Base64 data
func (img *ImageInfo) UpdateLastAccess() {
	img.LastBase64Access = FormatISO8601()
	img.Base64AccessCount++
}
