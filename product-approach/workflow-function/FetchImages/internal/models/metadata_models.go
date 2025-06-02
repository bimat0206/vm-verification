// Package models provides enhanced metadata models for the FetchImages function
package models

import (
	"time"
	"workflow-function/shared/schema"
)

// EnhancedImageMetadata represents the expected metadata structure for both reference and checking images
type EnhancedImageMetadata struct {
	VerificationId   string                    `json:"verificationId"`
	VerificationType string                    `json:"verificationType"`
	ReferenceImage   *EnhancedImageInfo        `json:"referenceImage"`
	CheckingImage    *EnhancedImageInfo        `json:"checkingImage"`
	ProcessingMetadata *ProcessingMetadata     `json:"processingMetadata"`
	Version          string                    `json:"version"`
}

// EnhancedImageInfo represents the nested structure for each image
type EnhancedImageInfo struct {
	OriginalMetadata *OriginalMetadata `json:"originalMetadata"`
	Base64Metadata   *Base64Metadata   `json:"base64Metadata"`
	StorageMetadata  *StorageMetadata  `json:"storageMetadata"`
	ImageType        string            `json:"imageType"`
	Validation       *ValidationInfo   `json:"validation"`
}

// OriginalMetadata contains the original image metadata
type OriginalMetadata struct {
	SourceUrl        string           `json:"sourceUrl"`
	SourceBucket     string           `json:"sourceBucket"`
	SourceKey        string           `json:"sourceKey"`
	ContentType      string           `json:"contentType"`
	OriginalSize     int64            `json:"originalSize"`
	ImageDimensions  *ImageDimensions `json:"imageDimensions"`
}

// ImageDimensions contains image dimension information
type ImageDimensions struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio float64 `json:"aspectRatio"`
}

// Base64Metadata contains Base64 encoding information
type Base64Metadata struct {
	OriginalSize        int64   `json:"originalSize"`
	EncodedSize         int64   `json:"encodedSize"`
	EncodingTimestamp   string  `json:"encodingTimestamp"`
	EncodingMethod      string  `json:"encodingMethod"`
	CompressionRatio    float64 `json:"compressionRatio"`
	QualitySettings     *QualitySettings `json:"qualitySettings"`
}

// QualitySettings contains quality optimization information
type QualitySettings struct {
	Optimized bool `json:"optimized"`
}

// StorageMetadata contains S3 storage information
type StorageMetadata struct {
	Bucket       string      `json:"bucket"`
	Key          string      `json:"key"`
	StoredSize   int64       `json:"storedSize"`
	StorageClass string      `json:"storageClass"`
	Encryption   *Encryption `json:"encryption"`
}

// Encryption contains encryption information
type Encryption struct {
	Method string `json:"method"`
}

// ValidationInfo contains validation results
type ValidationInfo struct {
	IsValid            bool `json:"isValid"`
	BedrockCompatible  bool `json:"bedrockCompatible"`
	SizeWithinLimits   bool `json:"sizeWithinLimits"`
}

// ProcessingMetadata contains processing information
type ProcessingMetadata struct {
	ProcessedAt          string   `json:"processedAt"`
	ProcessingTimeMs     int64    `json:"processingTimeMs"`
	TotalImagesProcessed int      `json:"totalImagesProcessed"`
	ProcessingSteps      []string `json:"processingSteps"`
	ParallelProcessing   bool     `json:"parallelProcessing"`
}

// ConvertToEnhancedMetadata converts the flat ImageMetadata to the expected nested structure
func ConvertToEnhancedMetadata(
	verificationId string,
	verificationType string,
	flatMetadata *ImageMetadata,
	processingStartTime time.Time,
) *EnhancedImageMetadata {
	processingEndTime := time.Now()
	processingTimeMs := processingEndTime.Sub(processingStartTime).Milliseconds()

	enhanced := &EnhancedImageMetadata{
		VerificationId:   verificationId,
		VerificationType: verificationType,
		ProcessingMetadata: &ProcessingMetadata{
			ProcessedAt:          processingEndTime.UTC().Format(time.RFC3339),
			ProcessingTimeMs:     processingTimeMs,
			TotalImagesProcessed: 2,
			ProcessingSteps: []string{
				"IMAGE_DOWNLOAD",
				"IMAGE_VALIDATION",
				"BASE64_ENCODING",
				"S3_STORAGE",
				"METADATA_EXTRACTION",
			},
			ParallelProcessing: true,
		},
		Version: "1.0",
	}

	// Convert reference image
	if flatMetadata.Reference != nil {
		enhanced.ReferenceImage = convertImageInfo(flatMetadata.Reference, "layout_reference")
	}

	// Convert checking image
	if flatMetadata.Checking != nil {
		enhanced.CheckingImage = convertImageInfo(flatMetadata.Checking, "current_checking")
	}

	return enhanced
}

// convertImageInfo converts a schema.ImageInfo to EnhancedImageInfo
func convertImageInfo(imageInfo *schema.ImageInfo, imageType string) *EnhancedImageInfo {
	// Calculate aspect ratio
	var aspectRatio float64
	if imageInfo.Height > 0 {
		aspectRatio = float64(imageInfo.Width) / float64(imageInfo.Height)
	}

	// Calculate compression ratio
	var compressionRatio float64
	if imageInfo.Base64Size > 0 && imageInfo.Size > 0 {
		compressionRatio = float64(imageInfo.Size) / float64(imageInfo.Base64Size)
	}

	enhanced := &EnhancedImageInfo{
		OriginalMetadata: &OriginalMetadata{
			SourceUrl:    imageInfo.URL,
			SourceBucket: imageInfo.S3Bucket,
			SourceKey:    imageInfo.S3Key,
			ContentType:  imageInfo.ContentType,
			OriginalSize: imageInfo.Size,
			ImageDimensions: &ImageDimensions{
				Width:       imageInfo.Width,
				Height:      imageInfo.Height,
				AspectRatio: aspectRatio,
			},
		},
		Base64Metadata: &Base64Metadata{
			OriginalSize:      imageInfo.Size,
			EncodedSize:       imageInfo.Base64Size,
			EncodingTimestamp: imageInfo.StorageDecisionAt,
			EncodingMethod:    "standard_base64",
			CompressionRatio:  compressionRatio,
			QualitySettings: &QualitySettings{
				Optimized: true,
			},
		},
		StorageMetadata: &StorageMetadata{
			Bucket:       imageInfo.Base64S3Bucket,
			Key:          imageInfo.GetBase64S3Key(),
			StoredSize:   imageInfo.Base64Size,
			StorageClass: "STANDARD",
			Encryption: &Encryption{
				Method: "SSE-S3",
			},
		},
		ImageType: imageType,
		Validation: &ValidationInfo{
			IsValid:           true,
			BedrockCompatible: imageInfo.ValidateBase64Size() == nil,
			SizeWithinLimits:  imageInfo.ValidateBase64Size() == nil,
		},
	}

	return enhanced
}
