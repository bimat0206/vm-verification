// Package schema provides Base64 helper functions for image processing
package schema

import (
	"encoding/base64"
	"fmt"
	//"mime"
	"strings"
)

// Base64ImageHelpers provides utilities for working with Base64-encoded images
type Base64ImageHelpers struct{}

// ConvertToBase64 converts image bytes to Base64 string
func (h *Base64ImageHelpers) ConvertToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}

// DecodeBase64 converts Base64 string back to image bytes
func (h *Base64ImageHelpers) DecodeBase64(base64Data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Data)
}

// DetectImageFormat detects image format from content type or file extension
func (h *Base64ImageHelpers) DetectImageFormat(contentType, filename string) string {
	// Try content type first
	if contentType != "" {
		switch contentType {
		case "image/png":
			return "png"
		case "image/jpeg", "image/jpg":
			return "jpeg"
		case "image/gif":
			return "gif"
		case "image/webp":
			return "webp"
		}
	}
	
	// Try file extension
	if filename != "" {
		ext := strings.ToLower(filename)
		if strings.HasSuffix(ext, ".png") {
			return "png"
		}
		if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
			return "jpeg"
		}
		if strings.HasSuffix(ext, ".gif") {
			return "gif"
		}
		if strings.HasSuffix(ext, ".webp") {
			return "webp"
		}
	}
	
	return "unknown"
}

// GetContentTypeFromFormat returns appropriate content type for image format
func (h *Base64ImageHelpers) GetContentTypeFromFormat(format string) string {
	switch strings.ToLower(format) {
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// ValidateImageFormat checks if the format is supported by Bedrock
func (h *Base64ImageHelpers) ValidateImageFormat(format string) error {
	supportedFormats := []string{"png", "jpeg", "jpg"}
	for _, supported := range supportedFormats {
		if strings.ToLower(format) == strings.ToLower(supported) {
			return nil
		}
	}
	return fmt.Errorf("unsupported image format: %s (supported: %v)", format, supportedFormats)
}

// EstimateBase64Size estimates the Base64 string size from original image size
func (h *Base64ImageHelpers) EstimateBase64Size(originalSize int64) int64 {
	// Base64 encoding increases size by approximately 33%
	return originalSize * 4 / 3
}

// EstimateOriginalSize estimates the original image size from Base64 string
func (h *Base64ImageHelpers) EstimateOriginalSize(base64Size int64) int64 {
	// Reverse of the Base64 encoding size increase
	return base64Size * 3 / 4
}

// CheckBase64SizeLimit validates that Base64 data is within Bedrock limits
func (h *Base64ImageHelpers) CheckBase64SizeLimit(base64Data string, maxSizeBytes int64) error {
	if maxSizeBytes <= 0 {
		maxSizeBytes = 20 * 1024 * 1024 // Default 20MB limit
	}
	
	currentSize := int64(len(base64Data))
	if currentSize > maxSizeBytes {
		return fmt.Errorf("Base64 data size (%d bytes) exceeds limit (%d bytes)", 
			currentSize, maxSizeBytes)
	}
	return nil
}

// ImageInfoBuilder helps build ImageInfo structures with Base64 data
type ImageInfoBuilder struct {
	imageInfo *ImageInfo
	helpers   *Base64ImageHelpers
}

// NewImageInfoBuilder creates a new ImageInfo builder
func NewImageInfoBuilder() *ImageInfoBuilder {
	return &ImageInfoBuilder{
		imageInfo: &ImageInfo{},
		helpers:   &Base64ImageHelpers{},
	}
}

// WithS3Info sets S3-related information
func (b *ImageInfoBuilder) WithS3Info(url, s3Key, s3Bucket string) *ImageInfoBuilder {
	b.imageInfo.URL = url
	b.imageInfo.S3Key = s3Key
	b.imageInfo.S3Bucket = s3Bucket
	return b
}

// WithImageData sets image properties and Base64 data
func (b *ImageInfoBuilder) WithImageData(imageBytes []byte, contentType string, filename string) *ImageInfoBuilder {
	// Generate Base64 data
	b.imageInfo.Base64Data = b.helpers.ConvertToBase64(imageBytes)
	b.imageInfo.Base64Size = int64(len(b.imageInfo.Base64Data))
	b.imageInfo.Size = int64(len(imageBytes))
	
	// Detect format
	b.imageInfo.Format = b.helpers.DetectImageFormat(contentType, filename)
	b.imageInfo.ContentType = contentType
	if b.imageInfo.ContentType == "" {
		b.imageInfo.ContentType = b.helpers.GetContentTypeFromFormat(b.imageInfo.Format)
	}
	
	return b
}

// WithMetadata sets additional metadata
func (b *ImageInfoBuilder) WithMetadata(key string, value interface{}) *ImageInfoBuilder {
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata[key] = value
	return b
}

// WithDimensions sets image dimensions
func (b *ImageInfoBuilder) WithDimensions(width, height int) *ImageInfoBuilder {
	b.imageInfo.Width = width
	b.imageInfo.Height = height
	return b
}

// WithS3Metadata sets S3-specific metadata
func (b *ImageInfoBuilder) WithS3Metadata(lastModified, etag string) *ImageInfoBuilder {
	b.imageInfo.LastModified = lastModified
	b.imageInfo.ETag = etag
	return b
}

// Build returns the constructed ImageInfo and validates it
func (b *ImageInfoBuilder) Build() (*ImageInfo, error) {
	// Validate the image format
	if err := b.helpers.ValidateImageFormat(b.imageInfo.Format); err != nil {
		return nil, err
	}
	
	// Validate Base64 size
	if err := b.imageInfo.ValidateBase64Size(); err != nil {
		return nil, err
	}
	
	return b.imageInfo, nil
}

// BedrockMessageBuilder helps build Bedrock messages with Base64 images
type BedrockMessageBuilder struct {
	role     string
	contents []BedrockContent
}

// NewBedrockMessageBuilder creates a new message builder
func NewBedrockMessageBuilder(role string) *BedrockMessageBuilder {
	return &BedrockMessageBuilder{
		role:     role,
		contents: []BedrockContent{},
	}
}

// AddText adds text content to the message
func (b *BedrockMessageBuilder) AddText(text string) *BedrockMessageBuilder {
	b.contents = append(b.contents, BedrockContent{
		Type: "text",
		Text: text,
	})
	return b
}

// AddImage adds image content from ImageInfo
func (b *BedrockMessageBuilder) AddImage(imageInfo *ImageInfo) *BedrockMessageBuilder {
	if imageInfo != nil && imageInfo.Base64Data != "" {
		b.contents = append(b.contents, BedrockContent{
			Type: "image",
			Image: &BedrockImageData{
				Format: imageInfo.Format,
				Source: BedrockImageSource{
					Type:  "bytes",
					Bytes: imageInfo.Base64Data,
				},
			},
		})
	}
	return b
}

// AddImageFromBase64 adds image content directly from Base64 data
func (b *BedrockMessageBuilder) AddImageFromBase64(base64Data, format string) *BedrockMessageBuilder {
	if base64Data != "" && format != "" {
		b.contents = append(b.contents, BedrockContent{
			Type: "image",
			Image: &BedrockImageData{
				Format: format,
				Source: BedrockImageSource{
					Type:  "bytes",
					Bytes: base64Data,
				},
			},
		})
	}
	return b
}

// Build returns the constructed BedrockMessage
func (b *BedrockMessageBuilder) Build() BedrockMessage {
	return BedrockMessage{
		Role:    b.role,
		Content: b.contents,
	}
}

// CurrentPromptBuilder helps build CurrentPrompt structures with Bedrock messages
type CurrentPromptBuilder struct {
	turnNumber   int
	includeImage string
	text         string
	messages     []BedrockMessage
	metadata     map[string]interface{}
}

// NewCurrentPromptBuilder creates a new prompt builder
func NewCurrentPromptBuilder(turnNumber int) *CurrentPromptBuilder {
	return &CurrentPromptBuilder{
		turnNumber: turnNumber,
		metadata:   make(map[string]interface{}),
	}
}

// WithIncludeImage sets which image to include
func (b *CurrentPromptBuilder) WithIncludeImage(includeImage string) *CurrentPromptBuilder {
	b.includeImage = includeImage
	return b
}

// WithText sets the text prompt (for backward compatibility)
func (b *CurrentPromptBuilder) WithText(text string) *CurrentPromptBuilder {
	b.text = text
	return b
}

// WithBedrockMessages sets the Bedrock-formatted messages
func (b *CurrentPromptBuilder) WithBedrockMessages(prompt string, images *ImageData) *CurrentPromptBuilder {
	b.messages = BuildBedrockMessages(prompt, b.includeImage, images)
	return b
}

// WithMetadata adds metadata to the prompt
func (b *CurrentPromptBuilder) WithMetadata(key string, value interface{}) *CurrentPromptBuilder {
	b.metadata[key] = value
	return b
}

// Build returns the constructed CurrentPrompt
func (b *CurrentPromptBuilder) Build() *CurrentPrompt {
	prompt := &CurrentPrompt{
		TurnNumber:   b.turnNumber,
		IncludeImage: b.includeImage,
		Messages:     b.messages,
		Metadata:     b.metadata,
		CreatedAt:    FormatISO8601(),
	}
	
	// Set text for backward compatibility
	if b.text != "" {
		prompt.Text = b.text
	}
	
	// Generate prompt ID
	prompt.PromptId = fmt.Sprintf("prompt-%s-turn%d", GetCurrentTimestamp(), b.turnNumber)
	
	return prompt
}

// ImageDataProcessor provides utilities for processing ImageData structures
type ImageDataProcessor struct {
	helpers *Base64ImageHelpers
}

// NewImageDataProcessor creates a new image data processor
func NewImageDataProcessor() *ImageDataProcessor {
	return &ImageDataProcessor{
		helpers: &Base64ImageHelpers{},
	}
}

// EnsureBase64Generated ensures both images have Base64 data
func (p *ImageDataProcessor) EnsureBase64Generated(images *ImageData) error {
	if images == nil {
		return fmt.Errorf("images cannot be nil")
	}
	
	if images.Reference != nil && !images.Reference.EnsureBase64Generated() {
		return fmt.Errorf("reference image missing Base64 data")
	}
	
	if images.Checking != nil && !images.Checking.EnsureBase64Generated() {
		return fmt.Errorf("checking image missing Base64 data")
	}
	
	images.Base64Generated = true
	images.ProcessedAt = FormatISO8601()
	
	return nil
}

// ValidateForBedrock validates that ImageData is ready for Bedrock API calls
func (p *ImageDataProcessor) ValidateForBedrock(images *ImageData) error {
	if err := p.EnsureBase64Generated(images); err != nil {
		return err
	}
	
	// Validate image formats
	if images.Reference != nil {
		if err := p.helpers.ValidateImageFormat(images.Reference.Format); err != nil {
			return fmt.Errorf("reference image: %w", err)
		}
	}
	
	if images.Checking != nil {
		if err := p.helpers.ValidateImageFormat(images.Checking.Format); err != nil {
			return fmt.Errorf("checking image: %w", err)
		}
	}
	
	return nil
}

// Global helper instances for easy access
var (
	Base64Helpers     = &Base64ImageHelpers{}
	ImageProcessor    = NewImageDataProcessor()
)