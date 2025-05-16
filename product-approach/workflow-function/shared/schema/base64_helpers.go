// Package schema provides Base64 helper functions for image processing with hybrid storage support
package schema

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// S3Interface defines the S3 client interface for testing
type S3Interface interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

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
		maxSizeBytes = BedrockMaxImageSize // Use default Bedrock limit
	}
	
	currentSize := int64(len(base64Data))
	if currentSize > maxSizeBytes {
		return fmt.Errorf("Base64 data size (%d bytes) exceeds limit (%d bytes)", 
			currentSize, maxSizeBytes)
	}
	return nil
}

// DetermineStorageMethod decides storage method based on Base64 size and configuration
func (h *Base64ImageHelpers) DetermineStorageMethod(base64Size int64, config *HybridStorageConfig) string {
	if config == nil || !config.EnableHybridStorage {
		return StorageMethodInline
	}
	
	threshold := config.Base64SizeThreshold
	if threshold <= 0 {
		threshold = DefaultBase64SizeThreshold
	}
	
	if base64Size <= threshold {
		return StorageMethodInline
	}
	return StorageMethodS3Temporary
}

// HybridImageInfoBuilder helps build ImageInfo structures with hybrid Base64 storage
type HybridImageInfoBuilder struct {
	imageInfo   *ImageInfo
	helpers     *Base64ImageHelpers
	config      *HybridStorageConfig
	s3Client    S3Interface
	context     context.Context
}

// NewHybridImageInfoBuilder creates a new ImageInfo builder with hybrid storage support
func NewHybridImageInfoBuilder(config *HybridStorageConfig, s3Client S3Interface) *HybridImageInfoBuilder {
	return &HybridImageInfoBuilder{
		imageInfo: &ImageInfo{},
		helpers:   &Base64ImageHelpers{},
		config:    config,
		s3Client:  s3Client,
		context:   context.Background(),
	}
}

// WithContext sets the context for S3 operations
func (b *HybridImageInfoBuilder) WithContext(ctx context.Context) *HybridImageInfoBuilder {
	b.context = ctx
	return b
}

// WithS3Info sets S3-related information
func (b *HybridImageInfoBuilder) WithS3Info(url, s3Key, s3Bucket string) *HybridImageInfoBuilder {
	b.imageInfo.URL = url
	b.imageInfo.S3Key = s3Key
	b.imageInfo.S3Bucket = s3Bucket
	return b
}

// WithImageDataAndHybridStorage processes image and applies hybrid storage logic
func (b *HybridImageInfoBuilder) WithImageDataAndHybridStorage(imageBytes []byte, contentType string, filename string, imageType string) *HybridImageInfoBuilder {
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
	
	// Determine storage method
	b.imageInfo.StorageMethod = b.helpers.DetermineStorageMethod(base64Size, b.config)
	b.imageInfo.StorageDecisionAt = FormatISO8601()
	
	// Apply storage method
	if b.imageInfo.StorageMethod == StorageMethodInline {
		// Store inline using the appropriate legacy field
		b.imageInfo.SetBase64Data(base64Data, imageType)
		b.imageInfo.Base64Generated = true
	} else {
		// Prepare for S3 temporary storage
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
	}
	
	return b
}

// WithMetadata sets additional metadata
func (b *HybridImageInfoBuilder) WithMetadata(key string, value interface{}) *HybridImageInfoBuilder {
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata[key] = value
	return b
}

// WithDimensions sets image dimensions
func (b *HybridImageInfoBuilder) WithDimensions(width, height int) *HybridImageInfoBuilder {
	b.imageInfo.Width = width
	b.imageInfo.Height = height
	return b
}

// WithS3Metadata sets S3-specific metadata
func (b *HybridImageInfoBuilder) WithS3Metadata(lastModified, etag string) *HybridImageInfoBuilder {
	b.imageInfo.LastModified = lastModified
	b.imageInfo.ETag = etag
	return b
}

// WithVerificationContext provides context for S3 key generation
func (b *HybridImageInfoBuilder) WithVerificationContext(verificationId, imageType string) *HybridImageInfoBuilder {
	if b.imageInfo.Metadata == nil {
		b.imageInfo.Metadata = make(map[string]interface{})
	}
	b.imageInfo.Metadata["_verificationId"] = verificationId
	b.imageInfo.Metadata["_imageType"] = imageType // "reference" or "checking"
	return b
}

// generateTempS3Key generates a unique S3 key for temporary Base64 storage
func (b *HybridImageInfoBuilder) generateTempS3Key(imageType string) string {
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

// Upload uploads the image data to S3 if using temporary storage
func (b *HybridImageInfoBuilder) Upload() error {
	if b.imageInfo.StorageMethod != StorageMethodS3Temporary {
		// Mark as generated for inline storage
		b.imageInfo.Base64Generated = true
		return nil
	}
	
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
	
	var tagSet []types.Tag
	for k, v := range tags {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
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

// convertToS3Metadata converts string map to AWS metadata format
func (b *HybridImageInfoBuilder) convertToS3Metadata(metadata map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// buildTaggingString builds the tagging string for S3
func (b *HybridImageInfoBuilder) buildTaggingString(tags map[string]string) *string {
	var pairs []string
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	result := strings.Join(pairs, "&")
	return &result
}

// getVerificationId extracts verification ID from metadata
func (b *HybridImageInfoBuilder) getVerificationId() string {
	if b.imageInfo.Metadata != nil {
		if vid, ok := b.imageInfo.Metadata["_verificationId"].(string); ok {
			return vid
		}
	}
	return "unknown"
}

// Build returns the constructed ImageInfo and validates it
func (b *HybridImageInfoBuilder) Build() (*ImageInfo, error) {
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

// HybridBase64Retriever handles retrieving Base64 data from various storage methods
type HybridBase64Retriever struct {
	s3Client S3Interface
	config   *HybridStorageConfig
	context  context.Context
}

// NewHybridBase64Retriever creates a new retriever
func NewHybridBase64Retriever(s3Client S3Interface, config *HybridStorageConfig) *HybridBase64Retriever {
	return &HybridBase64Retriever{
		s3Client: s3Client,
		config:   config,
		context:  context.Background(),
	}
}

// WithContext sets the context for S3 operations
func (r *HybridBase64Retriever) WithContext(ctx context.Context) *HybridBase64Retriever {
	r.context = ctx
	return r
}

// RetrieveBase64Data gets Base64 data regardless of storage method
func (r *HybridBase64Retriever) RetrieveBase64Data(imageInfo *ImageInfo) (string, error) {
	if imageInfo == nil {
		return "", fmt.Errorf("imageInfo cannot be nil")
	}
	
	// Update access tracking
	imageInfo.UpdateLastAccess()
	
	if imageInfo.IsInlineStorage() {
		// Return inline Base64 data using the helper method
		base64Data := imageInfo.GetBase64Data()
		if base64Data == "" {
			return "", fmt.Errorf("inline Base64 data is empty")
		}
		return base64Data, nil
	}
	
	if imageInfo.IsS3TemporaryStorage() {
		// Retrieve from S3
		return r.retrieveFromS3(imageInfo)
	}
	
	return "", fmt.Errorf("unknown storage method: %s", imageInfo.StorageMethod)
}

// retrieveFromS3 retrieves Base64 data from S3 temporary storage
func (r *HybridBase64Retriever) retrieveFromS3(imageInfo *ImageInfo) (string, error) {
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

// BedrockMessageBuilder helps build Bedrock messages with hybrid Base64 images
type BedrockMessageBuilder struct {
	role      string
	contents  []BedrockContent
	retriever *HybridBase64Retriever
}

// NewBedrockMessageBuilder creates a new message builder with hybrid Base64 support
func NewBedrockMessageBuilder(role string, retriever *HybridBase64Retriever) *BedrockMessageBuilder {
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

// AddImageWithHybridRetrieval adds image content with hybrid Base64 retrieval
func (b *BedrockMessageBuilder) AddImageWithHybridRetrieval(imageInfo *ImageInfo) error {
	if imageInfo == nil || !imageInfo.EnsureBase64Generated() {
		return fmt.Errorf("image info is nil or Base64 not generated")
	}
	
	// Retrieve Base64 data using hybrid retriever
	base64Data, err := b.retriever.RetrieveBase64Data(imageInfo)
	if err != nil {
		return fmt.Errorf("failed to retrieve Base64 data: %w", err)
	}
	
	// Add image content to message
	b.contents = append(b.contents, BedrockContent{
		Type: "image",
		Image: &BedrockImageData{
			Format: imageInfo.Format,
			Source: BedrockImageSource{
				Type:  "bytes",
				Bytes: base64Data,
			},
		},
	})
	
	return nil
}

// AddImage adds image content from ImageInfo (legacy method for backward compatibility)
func (b *BedrockMessageBuilder) AddImage(imageInfo *ImageInfo) *BedrockMessageBuilder {
	err := b.AddImageWithHybridRetrieval(imageInfo)
	if err != nil {
		// Log error but continue for backward compatibility
		// In production, this should be logged properly
		fmt.Printf("Warning: failed to add image with hybrid retrieval: %v\n", err)
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

// CurrentPromptBuilder helps build CurrentPrompt structures with hybrid Base64 messages
type CurrentPromptBuilder struct {
	turnNumber   int
	includeImage string
	text         string
	messages     []BedrockMessage
	metadata     map[string]interface{}
	retriever    *HybridBase64Retriever
}

// NewCurrentPromptBuilder creates a new prompt builder with hybrid Base64 support
func NewCurrentPromptBuilder(turnNumber int, retriever *HybridBase64Retriever) *CurrentPromptBuilder {
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

// WithBedrockMessagesHybrid sets the Bedrock-formatted messages with hybrid Base64 retrieval
func (b *CurrentPromptBuilder) WithBedrockMessagesHybrid(prompt string, images *ImageData) error {
	if b.retriever == nil {
		return fmt.Errorf("hybrid Base64 retriever not configured")
	}
	
	messages, err := b.buildBedrockMessagesWithRetrieval(prompt, b.includeImage, images)
	if err != nil {
		return fmt.Errorf("failed to build Bedrock messages: %w", err)
	}
	
	b.messages = messages
	return nil
}

// buildBedrockMessagesWithRetrieval creates Bedrock messages with hybrid Base64 retrieval
func (b *CurrentPromptBuilder) buildBedrockMessagesWithRetrieval(prompt string, includeImage string, images *ImageData) ([]BedrockMessage, error) {
	var messages []BedrockMessage
	
	// Determine which image to include using helper methods
	var imageToInclude *ImageInfo
	switch includeImage {
	case "reference":
		if images != nil {
			imageToInclude = images.GetReference()
		}
	case "checking":
		if images != nil {
			imageToInclude = images.GetChecking()
		}
	}
	
	// Build message with hybrid retrieval
	messageBuilder := NewBedrockMessageBuilder("user", b.retriever)
	messageBuilder.AddText(prompt)
	
	if imageToInclude != nil {
		err := messageBuilder.AddImageWithHybridRetrieval(imageToInclude)
		if err != nil {
			return nil, fmt.Errorf("failed to add image to message: %w", err)
		}
	}
	
	message := messageBuilder.Build()
	messages = append(messages, message)
	
	return messages, nil
}

// WithBedrockMessages sets the Bedrock-formatted messages (legacy method)
func (b *CurrentPromptBuilder) WithBedrockMessages(prompt string, images *ImageData) *CurrentPromptBuilder {
	err := b.WithBedrockMessagesHybrid(prompt, images)
	if err != nil {
		// Log error but continue for backward compatibility
		fmt.Printf("Warning: failed to build Bedrock messages with hybrid storage: %v\n", err)
		// Fall back to legacy method
		b.messages = BuildBedrockMessages(prompt, b.includeImage, images)
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

// HybridImageDataProcessor provides utilities for processing ImageData structures with hybrid storage
type HybridImageDataProcessor struct {
	helpers    *Base64ImageHelpers
	retriever  *HybridBase64Retriever
	builder    *HybridImageInfoBuilder
}

// NewHybridImageDataProcessor creates a new hybrid image data processor
func NewHybridImageDataProcessor(retriever *HybridBase64Retriever, builder *HybridImageInfoBuilder) *HybridImageDataProcessor {
	return &HybridImageDataProcessor{
		helpers:   &Base64ImageHelpers{},
		retriever: retriever,
		builder:   builder,
	}
}

// EnsureHybridBase64Generated ensures both images have Base64 data in appropriate storage
func (p *HybridImageDataProcessor) EnsureHybridBase64Generated(images *ImageData) error {
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
	p.updateStorageSummary(images)
	
	return nil
}

// updateStorageSummary updates the storage summary in ImageData
func (p *HybridImageDataProcessor) updateStorageSummary(images *ImageData) {
	var totalInlineSize int64
	var totalS3References int
	
	// Handle reference image (supporting both formats)
	ref := images.GetReference()
	if ref != nil {
		if ref.IsInlineStorage() {
			totalInlineSize += ref.GetBase64SizeEstimate()
		} else if ref.IsS3TemporaryStorage() {
			totalS3References++
		}
	}
	
	// Handle checking image (supporting both formats)
	checking := images.GetChecking()
	if checking != nil {
		if checking.IsInlineStorage() {
			totalInlineSize += checking.GetBase64SizeEstimate()
		} else if checking.IsS3TemporaryStorage() {
			totalS3References++
		}
	}
	
	images.TotalInlineSize = totalInlineSize
	images.TotalS3References = totalS3References
}

// ValidateForBedrockHybrid validates that ImageData is ready for Bedrock API calls with hybrid storage
func (p *HybridImageDataProcessor) ValidateForBedrockHybrid(images *ImageData) error {
	if err := p.EnsureHybridBase64Generated(images); err != nil {
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
	
	// Validate payload limits if using inline storage
	if err := images.ValidateForPayloadLimits(); err != nil {
		return err
	}
	
	return nil
}

// Legacy ImageDataProcessor for backward compatibility
// NOTE: This is a simplified version for backward compatibility
// The full implementation should use the hybrid processor
type ImageDataProcessor struct {
	helpers *Base64ImageHelpers
}

// NewImageDataProcessor creates a new legacy image data processor
func NewImageDataProcessor() *ImageDataProcessor {
	return &ImageDataProcessor{
		helpers: &Base64ImageHelpers{},
	}
}

// EnsureBase64Generated ensures both images have Base64 data (legacy method)
func (p *ImageDataProcessor) EnsureBase64Generated(images *ImageData) error {
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

// ValidateForBedrock validates that ImageData is ready for Bedrock API calls (legacy method)
func (p *ImageDataProcessor) ValidateForBedrock(images *ImageData) error {
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

// Global helper instances for easy access with hybrid storage support
var (
	Base64Helpers        = &Base64ImageHelpers{}
	ImageProcessor       = NewImageDataProcessor() // Legacy processor
	HybridImageProcessor = func(retriever *HybridBase64Retriever, builder *HybridImageInfoBuilder) *HybridImageDataProcessor {
		return NewHybridImageDataProcessor(retriever, builder)
	}
)