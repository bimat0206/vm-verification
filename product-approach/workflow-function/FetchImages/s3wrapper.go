package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// Lambda response size constants
const (
	// Lambda response size limit (6MB) minus buffer for other response data
	MaxLambdaResponseSize  = 6291556 // 6MB
	ResponseOverheadBuffer = 500000  // 500KB buffer for other response data
	MaxUsableResponseSize  = MaxLambdaResponseSize - ResponseOverheadBuffer // ~5.5MB
	Base64ExpansionFactor  = 1.33    // Base64 expands original size by ~33%
)

// ResponseSizeTracker tracks the total Base64 size across all images
type ResponseSizeTracker struct {
	mu                    sync.Mutex
	totalBase64Size       int64
	referenceBase64Size   int64
	checkingBase64Size    int64
	estimatedTotalSize    int64
}

// NewResponseSizeTracker creates a new response size tracker
func NewResponseSizeTracker() *ResponseSizeTracker {
	return &ResponseSizeTracker{}
}

// UpdateReferenceSize updates the reference image Base64 size
func (rst *ResponseSizeTracker) UpdateReferenceSize(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.referenceBase64Size = size
	rst.updateTotalSize()
}

// UpdateCheckingSize updates the checking image Base64 size
func (rst *ResponseSizeTracker) UpdateCheckingSize(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.checkingBase64Size = size
	rst.updateTotalSize()
}

// updateTotalSize recalculates the total size (must be called with lock held)
func (rst *ResponseSizeTracker) updateTotalSize() {
	rst.totalBase64Size = rst.referenceBase64Size + rst.checkingBase64Size
}

// GetTotalSize returns the current total Base64 size
func (rst *ResponseSizeTracker) GetTotalSize() int64 {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.totalBase64Size
}

// GetEstimatedTotalSize returns the estimated total size including pending images
func (rst *ResponseSizeTracker) GetEstimatedTotalSize() int64 {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.estimatedTotalSize
}

// SetEstimatedTotal sets the estimated total size for both images
func (rst *ResponseSizeTracker) SetEstimatedTotal(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.estimatedTotalSize = size
}

// WouldExceedLimit checks if adding a new Base64 would exceed the response limit
func (rst *ResponseSizeTracker) WouldExceedLimit(additionalSize int64) bool {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.totalBase64Size+additionalSize > MaxUsableResponseSize
}

// S3UtilsWrapper wraps the shared s3utils package with Base64 encoding functionality
type S3UtilsWrapper struct {
	s3utils             *s3utils.S3Utils
	s3Client            *s3.Client
	logger              logger.Logger
	maxImageSize        int64
	maxInlineBase64Size int64   // Individual image threshold
	tempBase64Bucket    string  // S3 bucket for temporary Base64 storage
	responseSizeTracker *ResponseSizeTracker // Tracks total response size
}

// NewS3Utils creates a new S3UtilsWrapper with AWS config and configuration
func NewS3Utils(config aws.Config, log logger.Logger, maxImageSize int64) *S3UtilsWrapper {
	s3Client := s3.NewFromConfig(config)
	
	// Set default max image size if not provided (100MB)
	if maxImageSize <= 0 {
		maxImageSize = 104857600 // 100MB default
	}
	
	return &S3UtilsWrapper{
		s3utils:             s3utils.NewWithConfig(config, log),
		s3Client:            s3Client,
		logger:              log.WithFields(map[string]interface{}{
			"component": "s3wrapper",
		}),
		maxImageSize:        maxImageSize,
		maxInlineBase64Size: 2097152, // 2MB default (individual threshold)
		tempBase64Bucket:    "",      // Will be set via SetTempBucket
		responseSizeTracker: NewResponseSizeTracker(), // Initialize response size tracker
	}
}

// SetTempBucket sets the temporary S3 bucket for Base64 storage
func (u *S3UtilsWrapper) SetTempBucket(bucket string) {
	u.tempBase64Bucket = bucket
}

// SetMaxInlineSize sets the maximum size for inline Base64 storage
func (u *S3UtilsWrapper) SetMaxInlineSize(size int64) {
	u.maxInlineBase64Size = size
}

// SetResponseSizeTracker sets a shared response size tracker (for parallel operations)
func (u *S3UtilsWrapper) SetResponseSizeTracker(tracker *ResponseSizeTracker) {
	u.responseSizeTracker = tracker
}

// GetResponseSizeTracker returns the current response size tracker
func (u *S3UtilsWrapper) GetResponseSizeTracker() *ResponseSizeTracker {
	return u.responseSizeTracker
}

// GetS3ImageWithBase64 downloads the image from S3 and returns metadata with Base64 encoded data
// using the dynamic response size calculation approach
func (u *S3UtilsWrapper) GetS3ImageWithBase64(ctx context.Context, s3url string) (ImageMetadata, error) {
	u.logger.Info("Downloading and encoding S3 image with dynamic response size calculation", map[string]interface{}{
		"s3url":                s3url,
		"maxInlineBase64Size":  u.maxInlineBase64Size,
		"maxUsableResponseSize": MaxUsableResponseSize,
		"tempBase64Bucket":     u.tempBase64Bucket,
	})

	// Parse S3 URL
	parsed, err := u.s3utils.ParseS3URL(s3url)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to parse S3 URL: %w", err)
	}

	// Validate image size before downloading
	if err := u.validateImageSize(ctx, parsed); err != nil {
		return ImageMetadata{}, fmt.Errorf("image validation failed: %w", err)
	}

	// Get metadata
	metadata, err := u.getImageMetadata(ctx, parsed)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Estimate Base64 size before downloading
	estimatedBase64Size := int64(float64(metadata.Size) * Base64ExpansionFactor)
	
	u.logger.Debug("Estimated Base64 size", map[string]interface{}{
		"originalSize":         metadata.Size,
		"estimatedBase64Size":  estimatedBase64Size,
		"expansionFactor":      Base64ExpansionFactor,
	})

	// Download the image
	u.logger.Debug("Starting image download", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
		"size":   metadata.Size,
	})

	getOutput, err := u.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to download image: %w", err)
	}
	defer getOutput.Body.Close()

	// Read image bytes
	imageBytes, err := io.ReadAll(getOutput.Body)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to read image bytes: %w", err)
	}

	// Convert to Base64
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	actualBase64Size := int64(len(base64Data))

	// Determine storage method with dynamic response size consideration
	storageMethod := u.determineStorageMethodWithResponseSize(actualBase64Size, s3url)
	
	u.logger.Info("Determined storage method with response size calculation", map[string]interface{}{
		"actualBase64Size":      actualBase64Size,
		"estimatedBase64Size":   estimatedBase64Size,
		"storageMethod":         storageMethod,
		"currentTotalSize":      u.responseSizeTracker.GetTotalSize(),
		"maxUsableResponseSize": MaxUsableResponseSize,
		"wouldExceedLimit":      u.responseSizeTracker.WouldExceedLimit(actualBase64Size),
	})

	// Store Base64 data according to determined method
	if err := u.storeBase64Data(ctx, &metadata, base64Data, storageMethod); err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to store Base64 data: %w", err)
	}

	// Update response size tracker
	u.updateResponseSizeTracker(actualBase64Size, storageMethod, s3url)

	// Update metadata with format information
	metadata.UpdateImageFormat()
	metadata.UpdateBedrockFormat()

	u.logger.Info("Successfully processed image with dynamic storage", map[string]interface{}{
		"contentType":           metadata.ContentType,
		"size":                  metadata.Size,
		"actualBase64Size":      actualBase64Size,
		"storageMethod":         metadata.StorageMethod,
		"imageFormat":           metadata.ImageFormat,
		"currentTotalBase64":    u.responseSizeTracker.GetTotalSize(),
		"responseUtilization":   float64(u.responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
	})

	return metadata, nil
}

// determineStorageMethodWithResponseSize determines storage method considering total response size
func (u *S3UtilsWrapper) determineStorageMethodWithResponseSize(base64Length int64, imageURL string) string {
	// First check if this single image would exceed our usable response size
	if base64Length > MaxUsableResponseSize {
		u.logger.Warn("Single image Base64 exceeds usable response size", map[string]interface{}{
			"base64Length":         base64Length,
			"maxUsableResponseSize": MaxUsableResponseSize,
			"imageURL":             imageURL,
		})
		return StorageMethodS3Temporary
	}

	// Check if adding this Base64 would exceed the total response size limit
	if u.responseSizeTracker.WouldExceedLimit(base64Length) {
		u.logger.Info("Using S3 storage due to total response size limit", map[string]interface{}{
			"base64Length":         base64Length,
			"currentTotalSize":     u.responseSizeTracker.GetTotalSize(),
			"wouldBeTotalSize":     u.responseSizeTracker.GetTotalSize() + base64Length,
			"maxUsableResponseSize": MaxUsableResponseSize,
			"imageURL":             imageURL,
		})
		return StorageMethodS3Temporary
	}

	// Apply individual image threshold
	if base64Length <= u.maxInlineBase64Size {
		return StorageMethodInline
	}

	u.logger.Info("Using S3 storage due to individual size threshold", map[string]interface{}{
		"base64Length":         base64Length,
		"maxInlineBase64Size":  u.maxInlineBase64Size,
		"imageURL":             imageURL,
	})
	return StorageMethodS3Temporary
}

// updateResponseSizeTracker updates the response size tracker based on storage method
func (u *S3UtilsWrapper) updateResponseSizeTracker(base64Size int64, storageMethod, imageURL string) {
	// Only count toward response size if stored inline
	if storageMethod == StorageMethodInline {
		// Determine if this is reference or checking image based on URL
		if strings.Contains(imageURL, "reference") || strings.Contains(imageURL, "processed") {
			u.responseSizeTracker.UpdateReferenceSize(base64Size)
		} else {
			u.responseSizeTracker.UpdateCheckingSize(base64Size)
		}
		
		u.logger.Debug("Updated response size tracker", map[string]interface{}{
			"base64Size":         base64Size,
			"storageMethod":      storageMethod,
			"newTotalSize":       u.responseSizeTracker.GetTotalSize(),
			"responseUtilization": float64(u.responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
		})
	} else {
		u.logger.Debug("Not counting S3-stored image in response size", map[string]interface{}{
			"base64Size":    base64Size,
			"storageMethod": storageMethod,
		})
	}
}

// determineStorageMethod decides whether to use inline or S3 storage based on individual image size only
func (u *S3UtilsWrapper) determineStorageMethod(base64Length int64) string {
	if base64Length <= u.maxInlineBase64Size {
		return StorageMethodInline
	}
	return StorageMethodS3Temporary
}

// storeBase64Data stores Base64 data according to the specified method
func (u *S3UtilsWrapper) storeBase64Data(ctx context.Context, metadata *ImageMetadata, base64Data, storageMethod string) error {
	switch storageMethod {
	case StorageMethodInline:
		metadata.SetInlineStorage(base64Data)
		u.logger.Debug("Stored Base64 data inline", map[string]interface{}{
			"length": len(base64Data),
		})
		return nil
		
	case StorageMethodS3Temporary:
		if u.tempBase64Bucket == "" {
			return fmt.Errorf("temporary S3 bucket not configured for large Base64 storage")
		}
		
		// Generate unique key for temporary storage
		tempKey, err := u.generateTempKey(metadata)
		if err != nil {
			return fmt.Errorf("failed to generate temporary key: %w", err)
		}
		
		// Store in S3
		if err := u.storeBase64InS3(ctx, base64Data, u.tempBase64Bucket, tempKey); err != nil {
			return fmt.Errorf("failed to store Base64 in S3: %w", err)
		}
		
		metadata.SetS3Storage(u.tempBase64Bucket, tempKey)
		u.logger.Debug("Stored Base64 data in S3", map[string]interface{}{
			"bucket": u.tempBase64Bucket,
			"key":    tempKey,
			"length": len(base64Data),
		})
		return nil
		
	default:
		return fmt.Errorf("unknown storage method: %s", storageMethod)
	}
}

// storeBase64InS3 stores Base64 data in the temporary S3 bucket
func (u *S3UtilsWrapper) storeBase64InS3(ctx context.Context, base64Data, bucket, key string) error {
	// Add lifecycle metadata for automatic cleanup
	metadata := map[string]string{
		"Content-Type":       "text/plain",
		"x-amz-meta-type":    "base64-data",
		"x-amz-meta-created": time.Now().UTC().Format(time.RFC3339),
		"x-amz-meta-size":    fmt.Sprintf("%d", len(base64Data)),
	}
	
	putInput := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     strings.NewReader(base64Data),
		Metadata: metadata,
		// Set server-side encryption
		ServerSideEncryption: "AES256",
	}
	
	// Add lifecycle configuration for automatic cleanup (24 hours)
	// This would typically be configured at the bucket level
	
	_, err := u.s3Client.PutObject(ctx, putInput)
	if err != nil {
		return fmt.Errorf("failed to put Base64 data to S3: %w", err)
	}
	
	u.logger.Debug("Successfully stored Base64 data in S3", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
		"size":   len(base64Data),
	})
	
	return nil
}

// RetrieveBase64FromS3 retrieves Base64 data from S3 temporary storage
func (u *S3UtilsWrapper) RetrieveBase64FromS3(ctx context.Context, bucket, key string) (string, error) {
	u.logger.Debug("Retrieving Base64 data from S3", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})
	
	getOutput, err := u.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", NewBase64RetrievalError(StorageMethodS3Temporary, 
			"failed to retrieve Base64 from S3", err)
	}
	defer getOutput.Body.Close()
	
	// Read Base64 data
	base64Bytes, err := io.ReadAll(getOutput.Body)
	if err != nil {
		return "", NewBase64RetrievalError(StorageMethodS3Temporary,
			"failed to read Base64 data from S3", err)
	}
	
	base64Data := string(base64Bytes)
	
	u.logger.Debug("Successfully retrieved Base64 data from S3", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
		"length": len(base64Data),
	})
	
	return base64Data, nil
}

// ConvertToS3Storage converts an inline-stored image to S3 storage (for post-processing optimization)
func (u *S3UtilsWrapper) ConvertToS3Storage(ctx context.Context, metadata *ImageMetadata) error {
	if metadata.StorageMethod != StorageMethodInline {
		return fmt.Errorf("image is not stored inline, cannot convert")
	}
	
	if metadata.Base64Data == "" {
		return fmt.Errorf("no inline Base64 data to convert")
	}
	
	u.logger.Info("Converting inline storage to S3 storage", map[string]interface{}{
		"currentSize": len(metadata.Base64Data),
		"bucket":      metadata.Bucket,
		"key":         metadata.Key,
	})
	
	// Store current inline data
	base64Data := metadata.Base64Data
	
	// Generate new temporary key
	tempKey, err := u.generateTempKey(metadata)
	if err != nil {
		return fmt.Errorf("failed to generate temporary key: %w", err)
	}
	
	// Store in S3
	if err := u.storeBase64InS3(ctx, base64Data, u.tempBase64Bucket, tempKey); err != nil {
		return fmt.Errorf("failed to convert to S3 storage: %w", err)
	}
	
	// Update metadata to reflect S3 storage
	metadata.SetS3Storage(u.tempBase64Bucket, tempKey)
	
	// Update response size tracker to remove the inline data size
	imageSize := int64(len(base64Data))
	u.responseSizeTracker.mu.Lock()
	if metadata.Bucket != "" && strings.Contains(metadata.Bucket, "reference") {
		u.responseSizeTracker.referenceBase64Size = 0
	} else {
		u.responseSizeTracker.checkingBase64Size = 0
	}
	u.responseSizeTracker.updateTotalSize()
	u.responseSizeTracker.mu.Unlock()
	
	u.logger.Info("Successfully converted to S3 storage", map[string]interface{}{
		"tempBucket":        u.tempBase64Bucket,
		"tempKey":           tempKey,
		"convertedSize":     imageSize,
		"newTotalBase64":    u.responseSizeTracker.GetTotalSize(),
	})
	
	return nil
}

// generateTempKey generates a unique key for temporary Base64 storage
func (u *S3UtilsWrapper) generateTempKey(metadata *ImageMetadata) (string, error) {
	// Generate random suffix for uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomSuffix := base64.URLEncoding.EncodeToString(randomBytes)[:8]
	
	// Create structured key path
	timestamp := time.Now().UTC().Format("2006-01-02/15-04-05")
	imageFormat := metadata.GetImageFormat()
	
	key := fmt.Sprintf("temp-base64/%s/%s-%s.base64", 
		timestamp, imageFormat, randomSuffix)
	
	return key, nil
}

// validateImageSize checks the image size before downloading
func (u *S3UtilsWrapper) validateImageSize(ctx context.Context, parsed s3utils.S3URL) error {
	// Get object metadata to check size before downloading
	headOutput, err := u.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object metadata: %w", err)
	}
	
	// Check size
	if headOutput.ContentLength != nil && *headOutput.ContentLength > u.maxImageSize {
		return fmt.Errorf("image too large: %d bytes (max %d bytes)", 
			*headOutput.ContentLength, u.maxImageSize)
	}
	
	// Validate content type
	if headOutput.ContentType != nil {
		contentType := *headOutput.ContentType
		if !u.s3utils.IsImageContentType(contentType) {
			return fmt.Errorf("invalid content type: %s (expected image type)", contentType)
		}
	}
	
	return nil
}

// getImageMetadata gets object metadata and creates ImageMetadata struct
func (u *S3UtilsWrapper) getImageMetadata(ctx context.Context, parsed s3utils.S3URL) (ImageMetadata, error) {
	// Get object metadata using HeadObject
	headOutput, err := u.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Create metadata object
	metadata := ImageMetadata{
		Bucket: parsed.Bucket,
		Key:    parsed.Key,
	}

	// Set content type
	if headOutput.ContentType != nil {
		metadata.ContentType = *headOutput.ContentType
	} else {
		metadata.ContentType = "application/octet-stream"
	}
	
	// Set size
	if headOutput.ContentLength != nil {
		metadata.Size = *headOutput.ContentLength
	}
	
	// Set last modified
	if headOutput.LastModified != nil {
		metadata.LastModified = headOutput.LastModified.Format("2006-01-02T15:04:05Z")
	}
	
	// Set ETag (remove quotes if present)
	if headOutput.ETag != nil {
		etag := *headOutput.ETag
		if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		metadata.ETag = etag
	}

	return metadata, nil
}

// ValidateImageForBedrock validates that an image meets Bedrock requirements
func (u *S3UtilsWrapper) ValidateImageForBedrock(ctx context.Context, s3url string) error {
	parsed, err := u.s3utils.ParseS3URL(s3url)
	if err != nil {
		return err
	}
	return u.validateImageSize(ctx, parsed)
}

// CleanupTempBase64 removes temporary Base64 data from S3 (optional cleanup method)
func (u *S3UtilsWrapper) CleanupTempBase64(ctx context.Context, bucket, key string) error {
	u.logger.Debug("Cleaning up temporary Base64 data", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})
	
	_, err := u.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		u.logger.Warn("Failed to cleanup temporary Base64 data", map[string]interface{}{
			"bucket": bucket,
			"key":    key,
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to cleanup temporary Base64 data: %w", err)
	}
	
	u.logger.Debug("Successfully cleaned up temporary Base64 data", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})
	
	return nil
}

// GetStorageConfig returns the current storage configuration
func (u *S3UtilsWrapper) GetStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"maxImageSize":           u.maxImageSize,
		"maxInlineBase64Size":    u.maxInlineBase64Size,
		"tempBase64Bucket":       u.tempBase64Bucket,
		"inlineThresholdMB":      float64(u.maxInlineBase64Size) / 1024 / 1024,
		"maxUsableResponseSize":  MaxUsableResponseSize,
		"responseUtilization":    float64(u.responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
		"currentTotalBase64":     u.responseSizeTracker.GetTotalSize(),
		"base64ExpansionFactor":  Base64ExpansionFactor,
	}
}

// GetResponseSizeInfo returns detailed information about response size tracking
func (u *S3UtilsWrapper) GetResponseSizeInfo() map[string]interface{} {
	return map[string]interface{}{
		"currentTotalSize":       u.responseSizeTracker.GetTotalSize(),
		"referenceBase64Size":    u.responseSizeTracker.referenceBase64Size,
		"checkingBase64Size":     u.responseSizeTracker.checkingBase64Size,
		"maxUsableResponseSize":  MaxUsableResponseSize,
		"responseOverheadBuffer": ResponseOverheadBuffer,
		"maxLambdaResponseSize":  MaxLambdaResponseSize,
		"utilizationPercentage":  float64(u.responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
		"remainingCapacity":      MaxUsableResponseSize - u.responseSizeTracker.GetTotalSize(),
	}
}