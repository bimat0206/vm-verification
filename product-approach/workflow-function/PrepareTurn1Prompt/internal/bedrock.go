package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Bedrock model constants
const (
	DefaultAnthropicVersion = "bedrock-2023-05-31"
	DefaultModelID          = "anthropic.claude-3-7-sonnet-20250219-v1:0"
	DefaultMaxTokens        = 24000
	DefaultBudgetTokens     = 16000
)

// Supported image media types for Bedrock
var supportedImageFormats = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
}

// ConfigureBedrockSettings creates a Bedrock configuration object based on environment settings
func ConfigureBedrockSettings() BedrockConfig {
	return BedrockConfig{
		AnthropicVersion: getEnv("ANTHROPIC_VERSION", DefaultAnthropicVersion),
		MaxTokens:        getIntEnv("MAX_TOKENS", DefaultMaxTokens),
		Thinking: ThinkingConfig{
			Type:         getEnv("THINKING_TYPE", "enabled"),
			BudgetTokens: getIntEnv("BUDGET_TOKENS", DefaultBudgetTokens),
		},
	}
}

// CreateTurn1Message creates a Bedrock message for the first turn with reference image
func CreateTurn1Message(promptText, referenceImageUrl, bucketOwner string) (BedrockMessage, error) {
	if promptText == "" {
		return BedrockMessage{}, fmt.Errorf("prompt text cannot be empty")
	}

	if referenceImageUrl == "" {
		return BedrockMessage{}, fmt.Errorf("reference image URL cannot be empty")
	}

	// Parse the S3 URL
	bucket, key, err := ExtractS3BucketAndKey(referenceImageUrl)
	if err != nil {
		return BedrockMessage{}, fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Determine image format from file extension
	imageFormat := getImageFormatFromKey(key)
	if imageFormat == "" {
		return BedrockMessage{}, fmt.Errorf("unsupported or missing image format for file: %s", key)
	}

	// Create the S3 location
	s3Location := S3Location{
		URI: fmt.Sprintf("s3://%s/%s", bucket, key),
	}

	// Add bucket owner if provided
	if bucketOwner != "" {
		s3Location.BucketOwner = bucketOwner
	}

	// Create message with text and image
	message := BedrockMessage{
		Role: "user",
		Content: []MessagePart{
			{
				Type: "text",
				Text: promptText,
			},
			{
				Type: "image",
				Image: &ImagePart{
					Format: imageFormat,
					Source: ImageSource{
						S3Location: s3Location,
					},
				},
			},
		},
	}
	
	return message, nil
}

// getImageFormatFromKey extracts the image format from an S3 key
func getImageFormatFromKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ext == "" {
		return ""
	}
	
	// Remove leading dot
	ext = ext[1:]
	
	// Check if format is supported
	if _, ok := supportedImageFormats[ext]; ok {
		return ext
	}
	
	return ""
}

// EstimateTokenUsage estimates token usage for a text prompt
func EstimateTokenUsage(text string) int {
	// Rough estimate: average of 4 characters per token for English text
	return len(text) / 4
}

// EstimateImageTokenUsage estimates token usage for an image
// This is a simplified estimation based on typical usage patterns
func EstimateImageTokenUsage() int {
	// Claude typically uses around 85 tokens per image (may vary)
	return 85
}

// EstimateTotalTurn1TokenUsage estimates the total token usage for Turn 1
func EstimateTotalTurn1TokenUsage(promptText string) int {
	// System prompt tokens + user prompt tokens + image tokens
	return EstimateTokenUsage(promptText) + EstimateImageTokenUsage()
}

// FormatBedrockRequest formats a full Bedrock request with system prompt and messages
func FormatBedrockRequest(systemPrompt string, messages []BedrockMessage, bedrockConfig BedrockConfig) ([]byte, error) {
	// Validate inputs
	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}
	
	// Create request
	request := BedrockRequest{
		AnthropicVersion: bedrockConfig.AnthropicVersion,
		MaxTokens:        bedrockConfig.MaxTokens,
		System:           systemPrompt,
		Messages:         messages,
		Thinking: BedrockThinking{
			Type:         bedrockConfig.Thinking.Type,
			BudgetTokens: bedrockConfig.Thinking.BudgetTokens,
		},
		Temperature:   0.7, // Default temperature for deterministic outputs
		TopP:          0.9, // Default topP for deterministic outputs
	}
	
	// Add metadata if available
	request.Meta = map[string]string{
		"usage":      "vending_machine_verification",
		"version":    getEnv("PROMPT_VERSION", "1.0.0"),
		"timestamps": "true",
	}
	
	// Marshal to JSON
	return json.Marshal(request)
}

// FormatTurn1BedrockRequest formats a Turn 1 Bedrock request with messages only
func FormatTurn1BedrockRequest(userMessage BedrockMessage, bedrockConfig BedrockConfig) ([]byte, error) {
	// Create request
	request := BedrockRequest{
		AnthropicVersion: bedrockConfig.AnthropicVersion,
		MaxTokens:        bedrockConfig.MaxTokens,
		Messages:         []BedrockMessage{userMessage},
		Thinking: BedrockThinking{
			Type:         bedrockConfig.Thinking.Type,
			BudgetTokens: bedrockConfig.Thinking.BudgetTokens,
		},
		Temperature:   0.7, // Default temperature for deterministic outputs
		TopP:          0.9, // Default topP for deterministic outputs
	}
	
	// Add metadata if available
	request.Meta = map[string]string{
		"usage":      "vending_machine_verification_turn1",
		"version":    getEnv("PROMPT_VERSION", "1.0.0"),
		"timestamps": "true",
	}
	
	// Marshal to JSON
	return json.Marshal(request)
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv retrieves an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}