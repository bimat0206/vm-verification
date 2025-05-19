package images

import (
	"context"
	"encoding/base64"
	"workflow-function/shared/errors"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// validateExistingBase64Data validates existing Base64 data
func (p *Processor) validateExistingBase64Data(imageInfo *schema.ImageInfo, base64Data string) error {
	// Basic Base64 validation
	if !isValidBase64(base64Data) {
		return errors.NewValidationError("Invalid Base64 data format", nil)
	}

	// Decode to check if it's a valid image (basic check)
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return errors.NewValidationError("Failed to decode Base64 data",
			map[string]interface{}{"error": err.Error()})
	}

	// Validate image size (max 10MB for Bedrock)
	if len(decoded) > 10*1024*1024 {
		return errors.NewValidationError("Image size exceeds Bedrock limit",
			map[string]interface{}{"sizeMB": len(decoded) / (1024 * 1024)})
	}

	// Set format if not already set
	if imageInfo.Format == "" {
		imageInfo.Format = detectImageFormatFromHeader(decoded)
	}

	// Validate format is supported by Bedrock
	if !isValidBedrockImageFormat(imageInfo.Format) {
		return errors.NewValidationError("Unsupported image format for Bedrock",
			map[string]interface{}{
				"format":    imageInfo.Format,
				"supported": []string{"jpeg", "png"},
			})
	}

	return nil
}

// retrieveBase64FromS3Temp retrieves Base64 data from S3 temporary storage
func (p *Processor) retrieveBase64FromS3Temp(ctx context.Context, imageInfo *schema.ImageInfo) error {
	if imageInfo.Base64S3Bucket == "" || imageInfo.GetBase64S3Key() == "" {
		return errors.NewValidationError("S3 temporary storage info missing",
			map[string]interface{}{
				"bucket": imageInfo.Base64S3Bucket,
				"key":    imageInfo.GetBase64S3Key(),
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
			map[string]interface{}{"sizeMB": len(imageData) / (1024 * 1024)})
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

	// Encode to Base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	imageInfo.Format = format
	imageInfo.Base64Generated = true
	imageInfo.StorageMethod = schema.StorageMethodS3Temporary

	// Validate the encoded Base64 data
	return p.validateExistingBase64Data(imageInfo, base64Data)
}