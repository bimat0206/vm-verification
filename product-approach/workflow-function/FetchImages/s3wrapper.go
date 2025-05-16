package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	//"strconv"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// S3UtilsWrapper wraps the shared s3utils package with Base64 encoding functionality
type S3UtilsWrapper struct {
	s3utils      *s3utils.S3Utils
	s3Client     *s3.Client
	logger       logger.Logger
	maxImageSize int64
}

// NewS3Utils creates a new S3UtilsWrapper with AWS config and max image size
func NewS3Utils(config aws.Config, log logger.Logger, maxImageSize int64) *S3UtilsWrapper {
	s3Client := s3.NewFromConfig(config)
	
	// Set default max size if not provided (100MB)
	if maxImageSize <= 0 {
		maxImageSize = 104857600 // 100MB default
	}
	
	return &S3UtilsWrapper{
		s3utils:      s3utils.NewWithConfig(config, log),
		s3Client:     s3Client,
		logger:       log.WithFields(map[string]interface{}{
			"component": "s3wrapper",
		}),
		maxImageSize: maxImageSize,
	}
}

// GetS3ImageWithBase64 downloads the image from S3 and returns metadata with Base64 encoded data
func (u *S3UtilsWrapper) GetS3ImageWithBase64(ctx context.Context, s3url string) (ImageMetadata, error) {
	u.logger.Info("Downloading and encoding S3 image", map[string]interface{}{
		"s3url": s3url,
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

	// Download the image
	u.logger.Debug("Starting image download", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key": parsed.Key,
		"size": metadata.Size,
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

	// Update metadata with Base64 data
	metadata.Base64Data = base64Data
	metadata.UpdateImageFormat()
	metadata.UpdateBedrockFormat()

	u.logger.Info("Successfully downloaded and encoded image", map[string]interface{}{
		"contentType":    metadata.ContentType,
		"size":           metadata.Size,
		"base64Length":   len(base64Data),
		"imageFormat":    metadata.ImageFormat,
	})

	return metadata, nil
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