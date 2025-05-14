package internal

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Validation configuration and constants
const (
	// Supported image formats for Bedrock
	ImageFormatPNG  = "png"
	ImageFormatJPEG = "jpeg"
	ImageFormatJPG  = "jpg"

	// Verification types
	VerificationTypeLayoutVsChecking = "LAYOUT_VS_CHECKING"
	VerificationTypePreviousVsCurrent = "PREVIOUS_VS_CURRENT"

	// Expected turn number for Turn 1
	ExpectedTurnNumber = 1

	// Expected image included value for Turn 1
	ExpectedImageIncluded = "reference"

	// Valid status values for verification context
	StatusTurn1PromptReady = "TURN1_PROMPT_READY"
	StatusImagesFetched    = "IMAGES_FETCHED"
)

var (
	// S3 URI regex pattern
	s3URIPattern = regexp.MustCompile(`^s3://([^/]+)/(.+)$`)

	// Supported image formats
	supportedImageFormats = []string{ImageFormatPNG, ImageFormatJPEG, ImageFormatJPG}
)

// ValidateExecuteTurn1Input validates the entire input structure
func ValidateExecuteTurn1Input(input *ExecuteTurn1Input) error {
	if input == nil {
		return NewValidationError("Input cannot be nil", nil)
	}

	// Validate verification context
	if err := ValidateVerificationContext(&input.VerificationContext); err != nil {
		return err
	}

	// Validate current prompt using the extraction helper
	_, err := ExtractAndValidateCurrentPrompt(input)
	if err != nil {
		return err
	}

	// Validate Bedrock configuration
	if err := ValidateBedrockConfig(&input.BedrockConfig); err != nil {
		return err
	}

	// Validation based on verification type
	switch input.VerificationContext.VerificationType {
	case VerificationTypeLayoutVsChecking:
		if err := ValidateLayoutVsCheckingInputs(input); err != nil {
			return err
		}
	case VerificationTypePreviousVsCurrent:
		if err := ValidatePreviousVsCurrentInputs(input); err != nil {
			return err
		}
	default:
		return NewInvalidFieldError("verificationType", 
			input.VerificationContext.VerificationType, 
			"LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT")
	}

	return nil
}

// ValidateVerificationContext validates the verification context
func ValidateVerificationContext(ctx *VerificationContext) error {
	if ctx == nil {
		return NewMissingFieldError("verificationContext")
	}

	// Required fields
	if ctx.VerificationID == "" {
		return NewMissingFieldError("verificationId")
	}

	if ctx.VerificationAt == "" {
		return NewMissingFieldError("verificationAt")
	}

	if ctx.Status != StatusTurn1PromptReady && ctx.Status != StatusImagesFetched {
		return NewInvalidFieldError("status", ctx.Status, "TURN1_PROMPT_READY or IMAGES_FETCHED")
	}

	if ctx.VerificationType == "" {
		return NewMissingFieldError("verificationType")
	}

	if ctx.VerificationType != VerificationTypeLayoutVsChecking && 
	   ctx.VerificationType != VerificationTypePreviousVsCurrent {
		return NewInvalidFieldError("verificationType", 
			ctx.VerificationType, 
			"LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT")
	}

	if ctx.ReferenceImageURL == "" {
		return NewMissingFieldError("referenceImageUrl")
	}

	if ctx.CheckingImageURL == "" {
		return NewMissingFieldError("checkingImageUrl")
	}

	// Validate S3 URIs
	if err := ValidateS3URI(ctx.ReferenceImageURL); err != nil {
		return NewValidationError(
			fmt.Sprintf("Invalid referenceImageUrl: %v", err),
			map[string]interface{}{"referenceImageUrl": ctx.ReferenceImageURL},
		)
	}

	if err := ValidateS3URI(ctx.CheckingImageURL); err != nil {
		return NewValidationError(
			fmt.Sprintf("Invalid checkingImageUrl: %v", err),
			map[string]interface{}{"checkingImageUrl": ctx.CheckingImageURL},
		)
	}

	return nil
}

// ValidateCurrentPrompt validates the current prompt structure
func ValidateCurrentPrompt(prompt *CurrentPrompt) error {
	if prompt == nil {
		return NewMissingFieldError("currentPrompt")
	}

	// Validate turn number
	if prompt.TurnNumber != ExpectedTurnNumber {
		return NewInvalidFieldError("turnNumber", prompt.TurnNumber, fmt.Sprintf("%d", ExpectedTurnNumber))
	}

	// Validate image included
	if prompt.ImageIncluded != ExpectedImageIncluded {
		return NewInvalidFieldError("imageIncluded", prompt.ImageIncluded, ExpectedImageIncluded)
	}

	// Validate messages
	if len(prompt.Messages) == 0 {
		return NewValidationError("Messages cannot be empty", nil)
	}

	// Validate the first (and should be only) message
	if err := ValidateBedrockMessage(&prompt.Messages[0]); err != nil {
		return err
	}

	// Additional validation for prompt metadata
	if prompt.PromptID == "" {
		return NewMissingFieldError("promptId")
	}

	if prompt.PromptVersion == "" {
		return NewMissingFieldError("promptVersion")
	}

	return nil
}

// ValidateBedrockMessage validates a Bedrock message structure
func ValidateBedrockMessage(msg *BedrockMessage) error {
	if msg == nil {
		return NewValidationError("Bedrock message cannot be nil", nil)
	}

	if msg.Role != "user" {
		return NewInvalidFieldError("role", msg.Role, "user")
	}

	if len(msg.Content) == 0 {
		return NewValidationError("Message content cannot be empty", nil)
	}

	// Validate content items
	hasText := false
	hasImage := false

	for i, content := range msg.Content {
		if err := ValidateMessageContent(&content); err != nil {
			return NewValidationError(
				fmt.Sprintf("Invalid content at index %d: %v", i, err),
				map[string]interface{}{"index": i},
			)
		}

		if content.Type == "text" {
			hasText = true
		} else if content.Type == "image" {
			hasImage = true
		}
	}

	// For Turn 1, we expect both text and image content
	if !hasText {
		return NewValidationError("Turn 1 message must contain text content", nil)
	}

	if !hasImage {
		return NewValidationError("Turn 1 message must contain image content", nil)
	}

	return nil
}

// ValidateMessageContent validates individual message content
func ValidateMessageContent(content *MessageContent) error {
	if content == nil {
		return NewValidationError("Message content cannot be nil", nil)
	}

	switch content.Type {
	case "text":
		if content.Text == nil || *content.Text == "" {
			return NewValidationError("Text content cannot be empty", nil)
		}
	case "image":
		if content.Image == nil {
			return NewValidationError("Image content cannot be nil", nil)
		}
		if err := ValidateImageData(content.Image); err != nil {
			return err
		}
	default:
		return NewInvalidFieldError("type", content.Type, "text or image")
	}

	return nil
}

// ValidateImageData validates image data structure
func ValidateImageData(img *ImageData) error {
	if img == nil {
		return NewValidationError("Image data cannot be nil", nil)
	}

	// Validate format
	if !isValidImageFormat(img.Format) {
		return NewInvalidImageFormatError(img.Format, supportedImageFormats)
	}

	// Validate S3 location
	if err := ValidateS3Location(&img.Source.S3Location); err != nil {
		return err
	}

	return nil
}

// ValidateS3Location validates S3 location structure
func ValidateS3Location(loc *S3Location) error {
	if loc == nil {
		return NewValidationError("S3 location cannot be nil", nil)
	}

	if loc.URI == "" {
		return NewMissingFieldError("uri")
	}

	if err := ValidateS3URI(loc.URI); err != nil {
		return err
	}

	if loc.BucketOwner == "" {
		return NewMissingFieldError("bucketOwner")
	}

	return nil
}

// ValidateBedrockConfig validates Bedrock configuration
func ValidateBedrockConfig(config *BedrockConfig) error {
	if config == nil {
		return NewMissingFieldError("bedrockConfig")
	}

	if config.AnthropicVersion == "" {
		return NewMissingFieldError("anthropic_version")
	}

	if config.MaxTokens <= 0 {
		return NewInvalidFieldError("max_tokens", config.MaxTokens, "positive integer")
	}

	// Validate thinking configuration
	if config.Thinking.Type != "enabled" && config.Thinking.Type != "enable" {
		return NewInvalidFieldError("thinking.type", config.Thinking.Type, "enabled or enable")
	}

	if config.Thinking.BudgetTokens <= 0 {
		return NewInvalidFieldError("thinking.budget_tokens", 
			config.Thinking.BudgetTokens, 
			"positive integer")
	}

	// Check if budget tokens is reasonable compared to max tokens
	if config.Thinking.BudgetTokens > config.MaxTokens {
		return NewValidationError(
			"Budget tokens cannot exceed max tokens",
			map[string]interface{}{
				"budgetTokens": config.Thinking.BudgetTokens,
				"maxTokens":    config.MaxTokens,
			},
		)
	}

	return nil
}

// ValidateLayoutVsCheckingInputs validates inputs specific to LAYOUT_VS_CHECKING
func ValidateLayoutVsCheckingInputs(input *ExecuteTurn1Input) error {
	ctx := &input.VerificationContext

	// Layout ID and prefix are required for this verification type
	if ctx.LayoutID == nil {
		return NewMissingFieldError("layoutId")
	}

	if *ctx.LayoutID <= 0 {
		return NewInvalidFieldError("layoutId", *ctx.LayoutID, "positive integer")
	}

	if ctx.LayoutPrefix == nil || *ctx.LayoutPrefix == "" {
		return NewMissingFieldError("layoutPrefix")
	}

	// Layout metadata should be present
	if input.LayoutMetadata == nil {
		return NewValidationError(
			"Layout metadata is required for LAYOUT_VS_CHECKING verification",
			map[string]interface{}{"verificationType": VerificationTypeLayoutVsChecking},
		)
	}

	// Validate layout metadata
	if err := ValidateLayoutMetadata(input.LayoutMetadata); err != nil {
		return err
	}

	return nil
}

// ValidatePreviousVsCurrentInputs validates inputs specific to PREVIOUS_VS_CURRENT
func ValidatePreviousVsCurrentInputs(input *ExecuteTurn1Input) error {
	// For PREVIOUS_VS_CURRENT, both images should be from checking bucket
	// This is validated by checking the bucket name in the S3 URIs
	
	// Extract bucket names from URIs
	refBucket, err := extractBucketFromS3URI(input.VerificationContext.ReferenceImageURL)
	if err != nil {
		return NewValidationError(
			fmt.Sprintf("Failed to extract bucket from referenceImageUrl: %v", err),
			map[string]interface{}{"referenceImageUrl": input.VerificationContext.ReferenceImageURL},
		)
	}

	checkBucket, err := extractBucketFromS3URI(input.VerificationContext.CheckingImageURL)
	if err != nil {
		return NewValidationError(
			fmt.Sprintf("Failed to extract bucket from checkingImageUrl: %v", err),
			map[string]interface{}{"checkingImageUrl": input.VerificationContext.CheckingImageURL},
		)
	}

	// Both should be from the same bucket (checking bucket)
	if refBucket != checkBucket {
		return NewValidationError(
			"For PREVIOUS_VS_CURRENT verification, both images must be from the same bucket",
			map[string]interface{}{
				"referenceImageBucket": refBucket,
				"checkingImageBucket":  checkBucket,
			},
		)
	}

	return nil
}

// ValidateLayoutMetadata validates layout metadata structure
func ValidateLayoutMetadata(metadata *LayoutMetadata) error {
	if metadata == nil {
		return NewValidationError("Layout metadata cannot be nil", nil)
	}

	if metadata.LayoutID <= 0 {
		return NewInvalidFieldError("layoutId", metadata.LayoutID, "positive integer")
	}

	if metadata.LayoutPrefix == "" {
		return NewMissingFieldError("layoutPrefix")
	}

	// Validate machine structure
	if err := ValidateMachineStructure(&metadata.MachineStructure); err != nil {
		return err
	}

	return nil
}

// ValidateMachineStructure validates machine structure
func ValidateMachineStructure(structure *MachineStructure) error {
	if structure == nil {
		return NewValidationError("Machine structure cannot be nil", nil)
	}

	if structure.RowCount <= 0 {
		return NewInvalidFieldError("rowCount", structure.RowCount, "positive integer")
	}

	if structure.ColumnsPerRow <= 0 {
		return NewInvalidFieldError("columnsPerRow", structure.ColumnsPerRow, "positive integer")
	}

	// Validate row order
	if len(structure.RowOrder) != structure.RowCount {
		return NewValidationError(
			fmt.Sprintf("Row order length (%d) does not match row count (%d)", 
				len(structure.RowOrder), structure.RowCount),
			map[string]interface{}{
				"rowOrderLength": len(structure.RowOrder),
				"rowCount":       structure.RowCount,
			},
		)
	}

	// Validate column order
	if len(structure.ColumnOrder) != structure.ColumnsPerRow {
		return NewValidationError(
			fmt.Sprintf("Column order length (%d) does not match columns per row (%d)", 
				len(structure.ColumnOrder), structure.ColumnsPerRow),
			map[string]interface{}{
				"columnOrderLength": len(structure.ColumnOrder),
				"columnsPerRow":     structure.ColumnsPerRow,
			},
		)
	}

	return nil
}

// ValidateS3URI validates an S3 URI format
func ValidateS3URI(uri string) error {
	if uri == "" {
		return NewValidationError("S3 URI cannot be empty", nil)
	}

	if !s3URIPattern.MatchString(uri) {
		return NewValidationError(
			"Invalid S3 URI format, expected: s3://bucket/key",
			map[string]interface{}{"uri": uri},
		)
	}

	// Parse as URL to ensure it's valid
	_, err := url.Parse(uri)
	if err != nil {
		return NewValidationError(
			fmt.Sprintf("Invalid S3 URI: %v", err),
			map[string]interface{}{"uri": uri},
		)
	}

	return nil
}

// Helper functions

// isValidImageFormat checks if the image format is supported
func isValidImageFormat(format string) bool {
	format = strings.ToLower(format)
	for _, supported := range supportedImageFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// extractBucketFromS3URI extracts the bucket name from an S3 URI
func extractBucketFromS3URI(uri string) (string, error) {
	matches := s3URIPattern.FindStringSubmatch(uri)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid S3 URI format")
	}
	return matches[1], nil
}

// extractKeyFromS3URI extracts the key from an S3 URI
func extractKeyFromS3URI(uri string) (string, error) {
	matches := s3URIPattern.FindStringSubmatch(uri)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid S3 URI format")
	}
	return matches[2], nil
}

// ParseS3URI parses an S3 URI and returns bucket and key
func ParseS3URI(uri string) (bucket, key string, err error) {
	matches := s3URIPattern.FindStringSubmatch(uri)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("invalid S3 URI format: %s", uri)
	}
	return matches[1], matches[2], nil
}