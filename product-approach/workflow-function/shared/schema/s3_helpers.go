// Package schema provides S3 helper functions for image processing with S3-only storage
package schema

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/service/s3"
	//"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// S3Interface defines the S3 client interface for testing
type S3Interface interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// Base64Helpers provides utilities for working with Base64-encoded images
var Base64Helpers = &base64ImageHelpers{}

// base64ImageHelpers provides utilities for working with Base64-encoded images
type base64ImageHelpers struct{}

// ConvertToBase64 converts image bytes to Base64 string
func (h *base64ImageHelpers) ConvertToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}

// DecodeBase64 converts Base64 string back to image bytes
func (h *base64ImageHelpers) DecodeBase64(base64Data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Data)
}

// DetectImageFormat detects image format from content type or file extension
func (h *base64ImageHelpers) DetectImageFormat(contentType, filename string) string {
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
func (h *base64ImageHelpers) GetContentTypeFromFormat(format string) string {
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
func (h *base64ImageHelpers) ValidateImageFormat(format string) error {
	supportedFormats := []string{"png", "jpeg", "jpg"}
	for _, supported := range supportedFormats {
		if strings.EqualFold(format, supported) {
			return nil
		}
	}
	return fmt.Errorf("unsupported image format: %s (supported: %v)", format, supportedFormats)
}

// EstimateBase64Size estimates the Base64 string size from original image size
func (h *base64ImageHelpers) EstimateBase64Size(originalSize int64) int64 {
	// Base64 encoding increases size by approximately 33%
	return originalSize * 4 / 3
}

// EstimateOriginalSize estimates the original image size from Base64 string
func (h *base64ImageHelpers) EstimateOriginalSize(base64Size int64) int64 {
	// Reverse of the Base64 encoding size increase
	return base64Size * 3 / 4
}

// CheckBase64SizeLimit validates that Base64 data is within Bedrock limits
func (h *base64ImageHelpers) CheckBase64SizeLimit(base64Data string, maxSizeBytes int64) error {
	if maxSizeBytes <= 0 {
		maxSizeBytes = BedrockMaxImageSize // Use default Bedrock limit
	}
	
	currentSize := int64(len(base64Data))
	if currentSize > maxSizeBytes {
		return fmt.Errorf("Base64 data size (%d bytes) exceeds limit (%d bytes)", 
			currentSize, maxSizeBytes)
	}
	return nil
}

// S3ImageInfoBuilder helps build ImageInfo structures with S3 storage
type S3ImageInfoBuilder struct {
	imageInfo   *ImageInfo
	helpers     *base64ImageHelpers
	config      *S3StorageConfig
	s3Client    S3Interface
	context     context.Context
}

// NewS3ImageInfoBuilder creates a new ImageInfo builder with S3 storage
func NewS3ImageInfoBuilder(config *S3StorageConfig, s3Client S3Interface) *S3ImageInfoBuilder {
	return &S3ImageInfoBuilder{
		imageInfo: &ImageInfo{},
		helpers:   &base64ImageHelpers{},
		config:    config,
		s3Client:  s3Client,
		context:   context.Background(),
	}
}

// WithContext sets the context for S3 operations
func (b *S3ImageInfoBuilder) WithContext(ctx context.Context) *S3ImageInfoBuilder {
	b.context = ctx
	return b
}

// WithS3Info sets S3-related information
func (b *S3ImageInfoBuilder) WithS3Info(url, s3Key, s3Bucket string) *S3ImageInfoBuilder {
	b.imageInfo.URL = url
	b.imageInfo.S3Key = s3Key
	b.imageInfo.S3Bucket = s3Bucket
	return b
}

// WithImageDataAndS3Storage processes image and applies S3 storage
func (b *S3ImageInfoBuilder) WithImageDataAndS3Storage(imageBytes []byte, contentType string, filename string, imageType string) *S3ImageInfoBuilder {
	// Generate Base64 data
	base64Data := b.helpers.ConvertToBase64(imageBytes)
	base64Size := int64(len(base64Data))
	
	// Set basic image properties
	b.imageInfo.Base64Size = base64Size
	b.imageInfo.Size = int64(len(imageBytes))
	b.imageInfo.Format = b.helpers.DetectImageFormat(contentType, filename)
	b.imageInfo.ContentType = contentType
	if b.imageInfo.ContentType == "" {
		b.imageInfo.ContentType = b.helpers.GetContentTypeFromFormat(b.imageInfo.Format)
	}
	
	// Set storage method to S3 temporary
	b.imageInfo.StorageMethod = StorageMethodS3Temporary
	b.imageInfo.StorageDecisionAt = FormatISO8601()
	
	// Prepare for S3 storage
	b.imageInfo.Base64S3Bucket = b.config.TempBase64Bucket
	
	// Generate and set S3 key using both unified and legacy fields
	s3Key := b.generateTempS3Key(imageType)
	b.imageInfo.SetBase64S3Key(s3Key, imageType)
	
	// Store metadata for S3 upload
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata["_tempBase64Data"] = base64Data
	b.imageInfo.Metadata["_imageType"] = imageType
	
	// Initialize S3 metadata
	b.imageInfo.Base64S3Metadata = map[string]string{
		"format":    b.imageInfo.Format,
		"size":      fmt.Sprintf("%d", base64Size),
		"imageType": imageType,
		"createdAt": FormatISO8601(),
	}
	
	return b
}

// WithMetadata sets additional metadata
func (b *S3ImageInfoBuilder) WithMetadata(key string, value interface{}) *S3ImageInfoBuilder {
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata[key] = value
	return b
}

// WithDimensions sets image dimensions
func (b *S3ImageInfoBuilder) WithDimensions(width, height int) *S3ImageInfoBuilder {
	b.imageInfo.Width = width
	b.imageInfo.Height = height
	return b
}

// WithS3Metadata sets S3-specific metadata
func (b *S3ImageInfoBuilder) WithS3Metadata(lastModified, etag string) *S3ImageInfoBuilder {
	b.imageInfo.LastModified = lastModified
	b.imageInfo.ETag = etag
	return b
}

// WithVerificationContext provides context for S3 key generation
func (b *S3ImageInfoBuilder) WithVerificationContext(verificationId, imageType string) *S3ImageInfoBuilder {
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata["_verificationId"] = verificationId
	b.imageInfo.Metadata["_imageType"] = imageType // "reference" or "checking"
	return b
}

// generateTempS3Key generates a unique S3 key for temporary Base64 storage
func (b *S3ImageInfoBuilder) generateTempS3Key(imageType string) string {
	timestamp := GetCurrentTimestamp()
	
	verificationId := "unknown"
	
	if b.imageInfo.Metadata != nil {
		if vid, ok := b.imageInfo.Metadata["_verificationId"].(string); ok {
			verificationId = vid
		}
	}
	
	return fmt.Sprintf("%s%s/%s-%s.base64", 
		TempBase64KeyPrefix, verificationId, imageType, timestamp)
}

// convertToS3Metadata converts string map to AWS metadata format
func (b *S3ImageInfoBuilder) convertToS3Metadata(metadata map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// buildTaggingString builds the tagging string for S3
func (b *S3ImageInfoBuilder) buildTaggingString(tags map[string]string) *string {
	var pairs []string
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	result := strings.Join(pairs, "&")
	return &result
}

// getVerificationId extracts verification ID from metadata
func (b *S3ImageInfoBuilder) getVerificationId() string {
	if b.imageInfo.Metadata != nil {
		if vid, ok := b.imageInfo.Metadata["_verificationId"].(string); ok {
			return vid
		}
	}
	return "unknown"
}

// Upload uploads the image data to S3
func (b *S3ImageInfoBuilder) Upload() error {
	// Extract temporary Base64 data
	tempData, exists := b.imageInfo.Metadata["_tempBase64Data"].(string)
	if !exists {
		return fmt.Errorf("no temporary Base64 data found for S3 upload")
	}
	
	// Get image type from metadata
	imageType := "image"
	if itype, ok := b.imageInfo.Metadata["_imageType"].(string); ok {
		imageType = itype
	}
	
	// Prepare S3 tags for lifecycle management
	tags := map[string]string{
		"Purpose":        "TempBase64Storage",
		"TTL":           fmt.Sprintf("%d", time.Now().Unix()+TempBase64TTL),
		"VerificationId": b.getVerificationId(),
		"ImageType":      imageType,
	}
	
	// Get the appropriate S3 key (supports legacy fields)
	s3Key := b.imageInfo.GetBase64S3Key()
	if s3Key == "" {
		return fmt.Errorf("no S3 key found for Base64 upload")
	}
	
	// Upload to S3
	_, err := b.s3Client.PutObject(b.context, &s3.PutObjectInput{
		Bucket:      aws.String(b.imageInfo.Base64S3Bucket),
		Key:         aws.String(s3Key),
		Body:        strings.NewReader(tempData),
		ContentType: aws.String("text/plain"),
		Metadata:    b.convertToS3Metadata(b.imageInfo.Base64S3Metadata),
		Tagging:     b.buildTaggingString(tags),
	})
	
	if err != nil {
		return fmt.Errorf("failed to upload Base64 to S3: %w", err)
	}
	
	// Clean up temporary data and mark as generated
	delete(b.imageInfo.Metadata, "_tempBase64Data")
	delete(b.imageInfo.Metadata, "_imageType")
	b.imageInfo.Base64Generated = true
	
	return nil
}

// Build returns the constructed ImageInfo and validates it
func (b *S3ImageInfoBuilder) Build() (*ImageInfo, error) {
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

// S3Base64Retriever handles retrieving Base64 data from S3 storage
type S3Base64Retriever struct {
	s3Client S3Interface
	config   *S3StorageConfig
	context  context.Context
}

// NewS3Base64Retriever creates a new retriever
func NewS3Base64Retriever(s3Client S3Interface, config *S3StorageConfig) *S3Base64Retriever {
	return &S3Base64Retriever{
		s3Client: s3Client,
		config:   config,
		context:  context.Background(),
	}
}

// WithContext sets the context for S3 operations
func (r *S3Base64Retriever) WithContext(ctx context.Context) *S3Base64Retriever {
	r.context = ctx
	return r
}

// RetrieveBase64Data gets Base64 data from S3
func (r *S3Base64Retriever) RetrieveBase64Data(imageInfo *ImageInfo) (string, error) {
	if imageInfo == nil {
		return "", fmt.Errorf("imageInfo cannot be nil")
	}
	
	// Update access tracking
	imageInfo.UpdateLastAccess()
	
	// Retrieve from S3
	return r.retrieveFromS3(imageInfo)
}

// retrieveFromS3 retrieves Base64 data from S3 storage
func (r *S3Base64Retriever) retrieveFromS3(imageInfo *ImageInfo) (string, error) {
	if imageInfo.Base64S3Bucket == "" {
		return "", fmt.Errorf("S3 bucket information is missing")
	}
	
	// Get the appropriate S3 key (supports legacy fields)
	s3Key := imageInfo.GetBase64S3Key()
	if s3Key == "" {
		return "", fmt.Errorf("S3 key information is missing")
	}
	
	// Set timeout for S3 operation
	ctx := r.context
	if r.config != nil && r.config.Base64RetrievalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(r.context, 
			time.Duration(r.config.Base64RetrievalTimeout)*time.Millisecond)
		defer cancel()
	}
	
	// Get object from S3
	result, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(imageInfo.Base64S3Bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve Base64 from S3 (bucket: %s, key: %s): %w", 
			imageInfo.Base64S3Bucket, s3Key, err)
	}
	defer result.Body.Close()
	
	// Read the Base64 data
	base64Data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Base64 data from S3: %w", err)
	}
	
	return string(base64Data), nil
}

// BedrockMessageBuilder helps build Bedrock messages with S3 Base64 images for v2.0.0
type BedrockMessageBuilder struct {
	role      string
	contents  []BedrockContent
	retriever *S3Base64Retriever
}

// NewBedrockMessageBuilder creates a new message builder with S3 Base64 support
func NewBedrockMessageBuilder(role string, retriever *S3Base64Retriever) *BedrockMessageBuilder {
	return &BedrockMessageBuilder{
		role:      role,
		contents:  []BedrockContent{},
		retriever: retriever,
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

// getMediaType returns the appropriate media type for the image format
func (b *BedrockMessageBuilder) getMediaType(format string) string {
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
		return "image/jpeg" // Default to JPEG if unknown
	}
}

// AddImageWithS3Retrieval adds image content with S3 Base64 retrieval using v2.0.0 schema
func (b *BedrockMessageBuilder) AddImageWithS3Retrieval(imageInfo *ImageInfo) error {
	if imageInfo == nil || !imageInfo.EnsureBase64Generated() {
		return fmt.Errorf("image info is nil or Base64 not generated")
	}
	
	// Retrieve Base64 data using S3 retriever
	base64Data, err := b.retriever.RetrieveBase64Data(imageInfo)
	if err != nil {
		return fmt.Errorf("failed to retrieve Base64 data: %w", err)
	}
	
	// Add image content to message using v2.0.0 schema format
	b.contents = append(b.contents, BedrockContent{
		Type: "image",
		Image: &BedrockImageData{
			Format: imageInfo.Format,
			Source: BedrockImageSource{
				Type:      "base64",
				Media_type: b.getMediaType(imageInfo.Format),
				Data:      base64Data,
			},
		},
	})
	
	return nil
}

// AddImage adds image content from ImageInfo
func (b *BedrockMessageBuilder) AddImage(imageInfo *ImageInfo) *BedrockMessageBuilder {
	err := b.AddImageWithS3Retrieval(imageInfo)
	if err != nil {
		// Log error but continue for backward compatibility
		fmt.Printf("Warning: failed to add image with S3 retrieval: %v\n", err)
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

// S3ImageDataProcessor provides utilities for processing ImageData structures with S3 storage
type S3ImageDataProcessor struct {
	helpers    *base64ImageHelpers
	retriever  *S3Base64Retriever
	builder    *S3ImageInfoBuilder
}

// NewS3ImageDataProcessor creates a new S3 image data processor
func NewS3ImageDataProcessor(retriever *S3Base64Retriever, builder *S3ImageInfoBuilder) *S3ImageDataProcessor {
	return &S3ImageDataProcessor{
		helpers:   &base64ImageHelpers{},
		retriever: retriever,
		builder:   builder,
	}
}

// EnsureS3Base64Generated ensures both images have Base64 data in S3 storage
func (p *S3ImageDataProcessor) EnsureS3Base64Generated(images *ImageData) error {
	if images == nil {
		return fmt.Errorf("images cannot be nil")
	}
	
	// Process reference image (supporting both new and legacy field names)
	ref := images.GetReference()
	if ref != nil {
		if !ref.EnsureBase64Generated() {
			return fmt.Errorf("reference image Base64 not properly generated")
		}
	}
	
	// Process checking image (supporting both new and legacy field names)
	checking := images.GetChecking()
	if checking != nil {
		if !checking.EnsureBase64Generated() {
			return fmt.Errorf("checking image Base64 not properly generated")
		}
	}
	
	// Update ImageData metadata
	images.Base64Generated = true
	images.ProcessedAt = FormatISO8601()
	
	// Calculate storage summary
	images.UpdateStorageSummary()
	
	return nil
}

// ValidateForBedrockS3 validates that ImageData is ready for Bedrock API calls with S3 storage
func (p *S3ImageDataProcessor) ValidateForBedrockS3(images *ImageData) error {
	if err := p.EnsureS3Base64Generated(images); err != nil {
		return err
	}
	
	// Validate reference image formats (supporting both formats)
	ref := images.GetReference()
	if ref != nil {
		if err := p.helpers.ValidateImageFormat(ref.Format); err != nil {
			return fmt.Errorf("reference image: %w", err)
		}
		if err := ref.ValidateBase64Size(); err != nil {
			return fmt.Errorf("reference image: %w", err)
		}
	}
	
	// Validate checking image formats (supporting both formats)
	checking := images.GetChecking()
	if checking != nil {
		if err := p.helpers.ValidateImageFormat(checking.Format); err != nil {
			return fmt.Errorf("checking image: %w", err)
		}
		if err := checking.ValidateBase64Size(); err != nil {
			return fmt.Errorf("checking image: %w", err)
		}
	}
	
	return nil
}

// CurrentPromptBuilder helps build CurrentPrompt structures with S3 Base64 messages
type CurrentPromptBuilder struct {
	turnNumber   int
	includeImage string
	text         string
	messages     []BedrockMessage
	metadata     map[string]interface{}
	retriever    *S3Base64Retriever
}

// NewCurrentPromptBuilder creates a new prompt builder with S3 Base64 support
func NewCurrentPromptBuilder(turnNumber int, retriever *S3Base64Retriever) *CurrentPromptBuilder {
	return &CurrentPromptBuilder{
		turnNumber: turnNumber,
		metadata:   make(map[string]interface{}),
		retriever:  retriever,
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

// WithBedrockMessagesS3 sets the Bedrock-formatted messages with S3 Base64 retrieval
func (b *CurrentPromptBuilder) WithBedrockMessagesS3(prompt string, images *ImageData) error {
	if b.retriever == nil {
		return fmt.Errorf("S3 Base64 retriever not configured")
	}
	
	// Determine which image to include using helper methods
	var imageToInclude *ImageInfo
	switch b.includeImage {
	case "reference":
		if images != nil {
			imageToInclude = images.GetReference()
		}
	case "checking":
		if images != nil {
			imageToInclude = images.GetChecking()
		}
	}
	
	// Build message with S3 retrieval
	messageBuilder := NewBedrockMessageBuilder("user", b.retriever)
	messageBuilder.AddText(prompt)
	
	if imageToInclude != nil {
		err := messageBuilder.AddImageWithS3Retrieval(imageToInclude)
		if err != nil {
			return fmt.Errorf("failed to add image to message: %w", err)
		}
	}
	
	message := messageBuilder.Build()
	b.messages = []BedrockMessage{message}
	
	return nil
}

// WithBedrockMessages sets the Bedrock-formatted messages (wrapper for WithBedrockMessagesS3)
func (b *CurrentPromptBuilder) WithBedrockMessages(prompt string, images *ImageData) *CurrentPromptBuilder {
	err := b.WithBedrockMessagesS3(prompt, images)
	if err != nil {
		// Log error but continue for backward compatibility
		fmt.Printf("Warning: failed to build Bedrock messages with S3 storage: %v\n", err)
	}
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

// Legacy wrapper for backward compatibility
// These functions provide compatibility with code that 
// expects the older API while using the new implementation
// ----------------------------------------------------

// ImageProcessor provides a global instance for backward compatibility
var ImageProcessor = &imageProcessor{helpers: Base64Helpers}

type imageProcessor struct {
	helpers *base64ImageHelpers
}

// EnsureBase64Generated ensures both images have Base64 data
func (p *imageProcessor) EnsureBase64Generated(images *ImageData) error {
	if images == nil {
		return fmt.Errorf("images cannot be nil")
	}
	
	// Process reference image
	ref := images.GetReference()
	if ref != nil && !ref.EnsureBase64Generated() {
		return fmt.Errorf("reference image missing Base64 data")
	}
	
	// Process checking image
	checking := images.GetChecking()
	if checking != nil && !checking.EnsureBase64Generated() {
		return fmt.Errorf("checking image missing Base64 data")
	}
	
	images.Base64Generated = true
	images.ProcessedAt = FormatISO8601()
	
	return nil
}

// ValidateForBedrock validates that ImageData is ready for Bedrock API calls
func (p *imageProcessor) ValidateForBedrock(images *ImageData) error {
	if err := p.EnsureBase64Generated(images); err != nil {
		return err
	}
	
	// Validate image formats
	ref := images.GetReference()
	if ref != nil {
		if err := p.helpers.ValidateImageFormat(ref.Format); err != nil {
			return fmt.Errorf("reference image: %w", err)
		}
	}
	
	checking := images.GetChecking()
	if checking != nil {
		if err := p.helpers.ValidateImageFormat(checking.Format); err != nil {
			return fmt.Errorf("checking image: %w", err)
		}
	}
	
	return nil
}

// S3ImageProcessor provides a function to create a new S3ImageDataProcessor
var S3ImageProcessor = func(retriever *S3Base64Retriever, builder *S3ImageInfoBuilder) *S3ImageDataProcessor {
	return NewS3ImageDataProcessor(retriever, builder)
}