package handler

import (
	"encoding/base64"
	"fmt"

	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// generateBase64Images processes images using hybrid storage approach
func (h *Handler) generateBase64Images(state *schema.WorkflowState, log logger.Logger) (*schema.HybridBase64Retriever, error) {
	log.Debug("Starting Base64 image generation", map[string]interface{}{
		"hasImages": state.Images != nil,
	})

	retriever := schema.NewHybridBase64Retriever(h.s3Client, h.hybridConfig)
	processor := schema.HybridImageProcessor(retriever, nil)

	if err := processor.EnsureHybridBase64Generated(state.Images); err != nil {
		log.Error("Failed to generate Base64 for images", map[string]interface{}{
			"error": err.Error(),
		})
		wfErr := wferrors.NewInternalError("HybridBase64Generation", err)
		return nil, h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	log.Debug("Base64 image generation completed successfully", nil)
	return retriever, nil
}

// buildImageContent creates the image block for Bedrock request
func (h *Handler) buildImageContent(
	images *schema.ImageData,
	retriever *schema.HybridBase64Retriever,
	log logger.Logger,
) (map[string]interface{}, error) {
	imageInfo := h.getReferenceImageInfo(images)
	if imageInfo == nil {
		return nil, fmt.Errorf("no reference image found")
	}

	// Retrieve Base64 data
	b64Data, err := retriever.RetrieveBase64Data(imageInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Base64 data for %s: %w", imageInfo.URL, err)
	}

	// Decode Base64 to raw bytes
	rawBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("invalid Base64 data: %w", err)
	}

	log.Debug("Image content prepared", map[string]interface{}{
		"imageUrl":   imageInfo.URL,
		"format":     imageInfo.Format,
		"sizeBytes":  len(rawBytes),
	})

	// Create image block according to Bedrock API specification
	return map[string]interface{}{
		"image": map[string]interface{}{
			"format": imageInfo.Format,
			"source": map[string]interface{}{
				"bytes": rawBytes,
			},
		},
	}, nil
}

// getReferenceImageInfo retrieves the reference image information
func (h *Handler) getReferenceImageInfo(images *schema.ImageData) *schema.ImageInfo {
	if images == nil {
		return nil
	}
	
	if images.Reference != nil {
		return images.Reference
	}
	
	// Legacy fallback
	return images.ReferenceImage
}