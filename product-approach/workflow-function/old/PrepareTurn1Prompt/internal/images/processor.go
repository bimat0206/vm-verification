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
			"base64S3Bucket": refImage.Base64S3Bucket,
			"base64S3Key":    refImage.GetBase64S3Key(),
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
				"error":          err.Error(),
				"processingTime": time.Since(startTime).String(),
			})
		} else {
			p.log.Info("Checking image processed successfully", map[string]interface{}{
				"storageMethod":  checkImage.StorageMethod,
				"format":         checkImage.Format,
				"hasBase64":      checkImage.Base64Generated,
				"base64S3Bucket": checkImage.Base64S3Bucket,
				"base64S3Key":    checkImage.GetBase64S3Key(),
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
		imageInfo.StorageMethod = schema.StorageMethodS3Temporary
	}

	// Check if image already has Base64 data generated
	if imageInfo.Base64Generated {
		// Verify Base64 storage references are valid
		if imageInfo.StorageMethod == schema.StorageMethodS3Temporary {
			if imageInfo.Base64S3Bucket == "" || imageInfo.GetBase64S3Key() == "" {
				p.log.Warn("Image has Base64Generated flag but missing storage references", map[string]interface{}{
					"url":            imageInfo.URL,
					"storageMethod":  imageInfo.StorageMethod,
					"base64S3Bucket": imageInfo.Base64S3Bucket,
					"base64S3Key":    imageInfo.GetBase64S3Key(),
				})

				// Force regeneration by setting Base64Generated to false
				imageInfo.Base64Generated = false
			} else {
				p.log.Info("Image already has Base64 data generated with valid references", map[string]interface{}{
					"url":            imageInfo.URL,
					"base64S3Bucket": imageInfo.Base64S3Bucket,
					"base64S3Key":    imageInfo.GetBase64S3Key(),
				})
				return nil
			}
		} else {
			p.log.Info("Image already has Base64 data generated", map[string]interface{}{
				"url":           imageInfo.URL,
				"storageMethod": imageInfo.StorageMethod,
			})
			return nil
		}
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
		return errors.NewValidationError("Invalid S3 URL format", map[string]interface{}{
			"url": s3URL,
		})
	}

	// Remove s3:// prefix
	s3URL = strings.TrimPrefix(s3URL, "s3://")

	// Split into bucket and key
	parts := strings.SplitN(s3URL, "/", 2)
	if len(parts) != 2 {
		return errors.NewValidationError("Invalid S3 URL format (missing key)", map[string]interface{}{
			"url": s3URL,
		})
	}

	bucket := parts[0]
	key := parts[1]

	// Store the bucket and key in the image info
	imageInfo.S3Bucket = bucket
	imageInfo.S3Key = key

	// Download and encode image from S3
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

	// After downloading and encoding, explicitly ensure Base64 reference fields are set
	if imageInfo.StorageMethod == schema.StorageMethodS3Temporary && imageInfo.Base64Generated {
		if imageInfo.Base64S3Bucket == "" || imageInfo.GetBase64S3Key() == "" {
			p.log.Warn("Base64 storage reference missing", map[string]interface{}{
				"url": imageInfo.URL,
			})
		} else {
			p.log.Info("Ensured Base64 storage reference for image", map[string]interface{}{
				"url":            imageInfo.URL,
				"base64S3Bucket": imageInfo.Base64S3Bucket,
				"base64S3Key":    imageInfo.GetBase64S3Key(),
			})
		}
	}

	p.log.Info("Successfully processed image from S3 URL", map[string]interface{}{
		"url":            imageInfo.URL,
		"bucket":         bucket,
		"key":            key,
		"format":         imageInfo.Format,
		"storageMethod":  imageInfo.StorageMethod,
		"base64S3Bucket": imageInfo.Base64S3Bucket,
		"base64S3Key":    imageInfo.GetBase64S3Key(),
	})

	return nil
}

// retrieveBase64FromS3Temp retrieves Base64 data from S3 temporary storage
func (p *Processor) retrieveBase64FromS3Temp(ctx context.Context, imageInfo *schema.ImageInfo) error {
	// Verify and log Base64 storage references
	if imageInfo.Base64S3Bucket == "" || imageInfo.GetBase64S3Key() == "" {
		return errors.NewValidationError("S3 temporary storage info missing",
			map[string]interface{}{
				"bucket": imageInfo.Base64S3Bucket,
				"key":    imageInfo.GetBase64S3Key(),
				"url":    imageInfo.URL,
			})
	}

	// Create reference
	ref := &s3state.Reference{
		Bucket: imageInfo.Base64S3Bucket,
		Key:    imageInfo.GetBase64S3Key(),
	}

	// Download Base64 string from S3
	data, err := p.s3Manager.Retrieve(ref)
	if err != nil {
		return errors.NewInternalError("s3-temp-download", err)
	}

	// Set Base64 generated flag
	imageInfo.Base64Generated = true
	p.log.Info("Retrieved Base64 data from S3 temporary storage", map[string]interface{}{
		"bucket":   imageInfo.Base64S3Bucket,
		"key":      imageInfo.GetBase64S3Key(),
		"dataSize": len(data),
	})

	// Validate the retrieved Base64 data
	return p.validateExistingBase64Data(imageInfo, string(data))
}

// downloadAndEncodeFromS3 downloads image from S3 and encodes to Base64
func (p *Processor) downloadAndEncodeFromS3(ctx context.Context, imageInfo *schema.ImageInfo) error {
	if imageInfo.S3Bucket == "" || imageInfo.S3Key == "" {
		return errors.NewValidationError("S3 storage info missing",
			map[string]interface{}{
				"bucket": imageInfo.S3Bucket,
				"key":    imageInfo.S3Key,
			})
	}

	// Create reference
	ref := &s3state.Reference{
		Bucket: imageInfo.S3Bucket,
		Key:    imageInfo.S3Key,
	}

	// Download image data
	imageData, err := p.s3Manager.Retrieve(ref)
	if err != nil {
		return errors.NewInternalError("s3-image-download", err)
	}

	// Validate image size before encoding
	if len(imageData) > 10*1024*1024 {
		return errors.NewValidationError("Image size exceeds Bedrock limit",
			map[string]interface{}{
				"sizeMB": len(imageData) / (1024 * 1024),
			})
	}

	// Detect and validate image format
	format := detectImageFormatFromHeader(imageData)
	if !isValidBedrockImageFormat(format) {
		return errors.NewValidationError("Unsupported image format for Bedrock",
			map[string]interface{}{
				"format":    format,
				"supported": []string{"jpeg", "png"},
			})
	}

	// Set format in image info
	imageInfo.Format = format

	// If using S3 temporary storage, set up Base64 references
	if imageInfo.StorageMethod == schema.StorageMethodS3Temporary {
		// Set Base64 S3 bucket if not already set
		// Ensure key information exists
		if imageInfo.Base64S3Bucket == "" || imageInfo.GetBase64S3Key() == "" {
			timestamp := time.Now().Format("20060102150405")
			if imageInfo.Base64S3Bucket == "" {
				imageInfo.Base64S3Bucket = ""
			}
			if imageInfo.GetBase64S3Key() == "" {
				imageInfo.Base64S3Key = fmt.Sprintf("temp/%s-base64.json", timestamp)
			}
		}

		// Upload Base64 encoded image data to S3 using state manager
		ref, err := p.s3Manager.Store(s3state.CategoryImages, imageInfo.GetBase64S3Key(), imageData)
		if err != nil {
			return errors.NewInternalError("s3-base64-upload", err)
		}

		// Update reference with actual key/bucket
		imageInfo.Base64S3Bucket = ref.Bucket
		imageInfo.Base64S3Key = ref.Key

		// Log successful upload
		p.log.Info("Uploaded Base64 encoded image data to S3", map[string]interface{}{
			"bucket":   imageInfo.Base64S3Bucket,
			"key":      imageInfo.GetBase64S3Key(),
			"dataSize": len(imageData),
			"format":   imageInfo.Format,
		})
	}

	// Set Base64 generated flag
	imageInfo.Base64Generated = true

	// Log successful processing
	p.log.Info("Successfully processed image from S3", map[string]interface{}{
		"sourceBucket":   imageInfo.S3Bucket,
		"sourceKey":      imageInfo.S3Key,
		"storageMethod":  imageInfo.StorageMethod,
		"format":         imageInfo.Format,
		"base64S3Bucket": imageInfo.Base64S3Bucket,
		"base64S3Key":    imageInfo.GetBase64S3Key(),
	})

	return nil
}
