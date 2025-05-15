package bedrock

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Validation constants
const (
	// Expected turn number for Turn 1
	ExpectedTurnNumber = 1

	// Expected image included value for Turn 1
	ExpectedImageIncluded = "reference"
)

var (
	// S3 URI regex pattern
	s3URIPattern = regexp.MustCompile(`^s3://([^/]+)/(.+)$`)

	// Supported image formats
	supportedImageFormats = []string{"png", "jpeg", "jpg", "gif", "webp"}
)

// ValidateConverseRequest validates a Converse API request
func ValidateConverseRequest(request *ConverseRequest) error {
	if request == nil {
		return fmt.Errorf("converse request cannot be nil")
	}

	if request.ModelId == "" {
		return fmt.Errorf("model ID cannot be empty")
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	// Validate messages
	for i, msg := range request.Messages {
		if err := ValidateMessageWrapper(msg); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	// Validate inference config
	if request.InferenceConfig.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}

	return nil
}

// ValidateMessageWrapper validates a message wrapper
func ValidateMessageWrapper(msg MessageWrapper) error {
	if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
		return fmt.Errorf("invalid role: %s", msg.Role)
	}

	if len(msg.Content) == 0 {
		return fmt.Errorf("content cannot be empty")
	}

	// Validate content blocks
	for i, content := range msg.Content {
		if err := ValidateContentBlock(content); err != nil {
			return fmt.Errorf("invalid content block at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateContentBlock validates a content block
func ValidateContentBlock(content ContentBlock) error {
	switch content.Type {
	case "text":
		if content.Text == "" {
			return fmt.Errorf("text content cannot be empty")
		}
	case "image":
		if content.Image == nil {
			return fmt.Errorf("image content cannot be nil")
		}
		if err := ValidateImageBlock(content.Image); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported content type: %s", content.Type)
	}

	return nil
}

// ValidateImageBlock validates an image block
func ValidateImageBlock(img *ImageBlock) error {
	if img == nil {
		return fmt.Errorf("image block cannot be nil")
	}

	// Validate format
	if !isValidImageFormat(img.Format) {
		return fmt.Errorf("invalid image format: %s", img.Format)
	}

	// Validate source
	if err := ValidateImageSource(img.Source); err != nil {
		return err
	}

	return nil
}

// ValidateImageSource validates an image source
func ValidateImageSource(source ImageSource) error {
	// Validate S3 location
	if err := ValidateS3Location(source.S3Location); err != nil {
		return err
	}

	return nil
}

// ValidateS3Location validates an S3 location
func ValidateS3Location(loc S3Location) error {
	if loc.URI == "" {
		return fmt.Errorf("S3 URI cannot be empty")
	}

	if err := ValidateS3URI(loc.URI); err != nil {
		return err
	}

	if loc.BucketOwner == "" {
		return fmt.Errorf("bucket owner cannot be empty")
	}

	return nil
}

// ValidateS3URI validates an S3 URI format
func ValidateS3URI(uri string) error {
	if uri == "" {
		return fmt.Errorf("S3 URI cannot be empty")
	}

	if !s3URIPattern.MatchString(uri) {
		return fmt.Errorf("invalid S3 URI format, expected: s3://bucket/key")
	}

	// Parse as URL to ensure it's valid
	_, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid S3 URI: %w", err)
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

// ExtractBucketFromS3URI extracts the bucket name from an S3 URI
func ExtractBucketFromS3URI(uri string) (string, error) {
	matches := s3URIPattern.FindStringSubmatch(uri)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid S3 URI format")
	}
	return matches[1], nil
}

// ExtractKeyFromS3URI extracts the key from an S3 URI
func ExtractKeyFromS3URI(uri string) (string, error) {
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

// ExtractAndValidateCurrentPrompt extracts the current prompt from the input
func ExtractAndValidateCurrentPrompt(messages []MessageWrapper) error {
	if len(messages) == 0 {
		return fmt.Errorf("no messages in current prompt")
	}

	// Validate the first (and should be only) message
	if err := ValidateMessageWrapper(messages[0]); err != nil {
		return err
	}

	return nil
}
