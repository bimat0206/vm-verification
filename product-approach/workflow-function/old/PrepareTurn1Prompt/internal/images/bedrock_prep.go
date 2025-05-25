package images

import (
	"encoding/base64"
	"workflow-function/shared/errors"
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
