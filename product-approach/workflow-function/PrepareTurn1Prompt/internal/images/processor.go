package images

import (
	"context"
	"fmt"
	"strings"
	"time"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// Processor handles image processing for Bedrock preparation
type Processor struct {
	s3Manager s3state.Manager
	log       logger.Logger
}

// NewProcessor creates a new image processor with the given S3 manager and logger
func NewProcessor(s3Manager s3state.Manager, log logger.Logger) *Processor {
	return &Processor{
		s3Manager: s3Manager,
		log:       log,
	}
}

// ProcessImagesForBedrock processes all images in the workflow state for Bedrock use
func (p *Processor) ProcessImagesForBedrock(ctx context.Context, workflowState *schema.WorkflowState) error {
	if workflowState.Images == nil {
		return errors.NewValidationError("Images data is nil", nil)
	}

	// Process reference image (required for Turn 1)
	if refImage := workflowState.Images.GetReference(); refImage != nil {
		startTime := time.Now()
		if err := p.ProcessImageForBedrock(ctx, refImage); err != nil {
			return fmt.Errorf("failed to process reference image: %w", err)
		}
		p.log.Info("Reference image processed successfully", map[string]interface{}{
			"storageMethod":  refImage.StorageMethod,
			"format":         refImage.Format,
			"hasBase64":      refImage.Base64Generated,
			"processingTime": time.Since(startTime).String(),
		})
	} else {
		return errors.NewValidationError("Reference image is required for Turn 1", nil)
	}

	// Process checking image if available (not needed for Turn 1 but may be present)
	if checkImage := workflowState.Images.GetChecking(); checkImage != nil {
		startTime := time.Now()
		if err := p.ProcessImageForBedrock(ctx, checkImage); err != nil {
			// Log warning but don't fail for checking image in Turn 1
			p.log.Warn("Failed to process checking image", map[string]interface{}{
				"error":         err.Error(),
				"processingTime": time.Since(startTime).String(),
			})
		} else {
			p.log.Info("Checking image processed successfully", map[string]interface{}{
				"storageMethod":  checkImage.StorageMethod,
				"format":         checkImage.Format,
				"hasBase64":      checkImage.Base64Generated,
				"processingTime": time.Since(startTime).String(),
			})
		}
	}

	return nil
}

// ProcessImageForBedrock processes a single image for Bedrock preparation
func (p *Processor) ProcessImageForBedrock(ctx context.Context, imageInfo *schema.ImageInfo) error {
	if imageInfo == nil {
		return errors.NewValidationError("Image info is nil", nil)
	}

	// Handle case where image info is created from URL but doesn't have storage method
	if imageInfo.StorageMethod == "" && imageInfo.URL != "" {
		p.log.Info("Setting storage method to s3-temporary for image with URL", map[string]interface{}{
			"url": imageInfo.URL,
		})
		imageInfo.StorageMethod = "s3-temporary"
	}

	// Check if image already has Base64 data generated
	if imageInfo.Base64Generated {
		p.log.Info("Image already has Base64 data generated", map[string]interface{}{
			"url": imageInfo.URL,
		})
		return nil
	}

	// Determine how to retrieve Base64 data based on storage method
	switch imageInfo.StorageMethod {
	case schema.StorageMethodS3Temporary:
		// If image has a URL in s3:// format, process it directly
		if imageInfo.URL != "" && strings.HasPrefix(imageInfo.URL, "s3://") {
			return p.processFromS3URL(ctx, imageInfo)
		}
		// Otherwise, retrieve Base64 data from S3 temporary storage
		return p.retrieveBase64FromS3Temp(ctx, imageInfo)

	default:
		// Default: download from regular S3 bucket and encode
		return p.downloadAndEncodeFromS3(ctx, imageInfo)
	}
}

// processFromS3URL processes an image from its S3 URL directly
func (p *Processor) processFromS3URL(ctx context.Context, imageInfo *schema.ImageInfo) error {
	if imageInfo.URL == "" {
		return errors.NewValidationError("URL is empty", nil)
	}

	// Extract bucket and key from URL (format: s3://bucket/key)
	s3URL := imageInfo.URL
	if !strings.HasPrefix(s3URL, "s3://") {
		return errors.NewValidationError("Invalid S3 URL format", map[string]interface{}{"url": s3URL})
	}

	// Remove s3:// prefix
	s3URL = strings.TrimPrefix(s3URL, "s3://")

	// Split into bucket and key
	parts := strings.SplitN(s3URL, "/", 2)
	if len(parts) != 2 {
		return errors.NewValidationError("Invalid S3 URL format (missing key)", map[string]interface{}{"url": s3URL})
	}

	bucket := parts[0]
	key := parts[1]

	// Store the bucket and key in the image info
	imageInfo.S3Bucket = bucket
	imageInfo.S3Key = key

	// Now we can use the standard method to download and encode from S3
	err := p.downloadAndEncodeFromS3(ctx, imageInfo)
	if err != nil {
		p.log.Error("Failed to process image from S3 URL", map[string]interface{}{
			"error":  err.Error(),
			"bucket": bucket,
			"key":    key,
			"url":    s3URL,
		})
		return err
	}

	p.log.Info("Successfully processed image from S3 URL", map[string]interface{}{
		"url":     imageInfo.URL,
		"bucket":  bucket,
		"key":     key,
		"format":  imageInfo.Format,
	})

	return nil
}