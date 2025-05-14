package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"workflow-function/shared/schema"
)

// Bedrock model constants
const (
	DefaultAnthropicVersion = "bedrock-2023-05-31"
	DefaultModelID          = "anthropic.claude-3-7-sonnet-20250219-v1:0"
	DefaultMaxTokens        = 24000
	DefaultBudgetTokens     = 16000
)

// Supported image media types for Bedrock
var supportedImageMediaTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// BedrockRequest represents the request structure for Bedrock
type BedrockRequest struct {
	AnthropicVersion string            `json:"anthropic_version"`
	MaxTokens        int               `json:"max_tokens"`
	Messages         []BedrockMessage  `json:"messages"`
	System           string            `json:"system,omitempty"`
	Thinking         BedrockThinking   `json:"thinking,omitempty"`
	Temperature      float64           `json:"temperature,omitempty"`
	TopP             float64           `json:"top_p,omitempty"`
	TopK             int               `json:"top_k,omitempty"`
	StopSequences    []string          `json:"stop_sequences,omitempty"`
	Meta             map[string]string `json:"meta,omitempty"`
}

// BedrockMessage represents a message in the Bedrock request
type BedrockMessage struct {
	Role    string        `json:"role"`
	Content []MessagePart `json:"content"`
}

// MessagePart represents a part of a message (text or image)
type MessagePart struct {
	Type    string       `json:"type"`
	Text    string       `json:"text,omitempty"`
	Source  *ImageSource `json:"source,omitempty"`
}

// ImageSource represents an image source
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data,omitempty"`
	URI       string `json:"uri,omitempty"`
}

// BedrockThinking configures Claude's thinking behavior
type BedrockThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// BedrockResponse represents a response from Bedrock
type BedrockResponse struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Content      []MessagePart `json:"content"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason"`
	StopSequence string        `json:"stop_sequence"`
	Usage        BedrockUsage  `json:"usage"`
}

// BedrockUsage represents token usage information
type BedrockUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ConfigureBedrockSettings creates a Bedrock configuration object based on environment settings
func ConfigureBedrockSettings() *schema.BedrockConfig {
	// Get environment variables
	anthropicVersion := getEnv("ANTHROPIC_VERSION", DefaultAnthropicVersion)
	maxTokens := getIntEnv("MAX_TOKENS", DefaultMaxTokens)
	thinkingType := getEnv("THINKING_TYPE", "enabled")
	budgetTokens := getIntEnv("BUDGET_TOKENS", DefaultBudgetTokens)
	
	// Create default temperature and topP if not provided
	temperature := getFloatEnv("TEMPERATURE", 0.7)
	topP := getFloatEnv("TOP_P", 0.9)
	
	// Create shared schema BedrockConfig
	return &schema.BedrockConfig{
		AnthropicVersion: anthropicVersion,
		MaxTokens:        maxTokens,
		Temperature:      temperature,
		TopP:             topP,
		Thinking: &schema.Thinking{
			Type:         thinkingType,
			BudgetTokens: budgetTokens,
		},
	}
}

// PrepareBedrockRequest creates a BedrockRequest object for system prompt
func PrepareBedrockRequest(systemPrompt string, bedrockConfig *schema.BedrockConfig) (BedrockRequest, error) {
	// Validate system prompt
	if systemPrompt == "" {
		return BedrockRequest{}, fmt.Errorf("system prompt cannot be empty")
	}
	
	// Validate config
	if bedrockConfig == nil {
		return BedrockRequest{}, fmt.Errorf("bedrock config cannot be nil")
	}
	
	// Create request
	request := BedrockRequest{
		AnthropicVersion: bedrockConfig.AnthropicVersion,
		MaxTokens:        bedrockConfig.MaxTokens,
		System:           systemPrompt,
		Messages:         []BedrockMessage{},
		Temperature:      bedrockConfig.Temperature,
		TopP:             bedrockConfig.TopP,
	}
	
	// Set thinking if available
	if bedrockConfig.Thinking != nil {
		request.Thinking = BedrockThinking{
			Type:         bedrockConfig.Thinking.Type,
			BudgetTokens: bedrockConfig.Thinking.BudgetTokens,
		}
	} else {
		// Default thinking settings
		request.Thinking = BedrockThinking{
			Type:         "enabled",
			BudgetTokens: DefaultBudgetTokens,
		}
	}
	
	// Add metadata
	request.Meta = map[string]string{
		"usage":      "vending_machine_verification",
		"version":    getEnv("PROMPT_VERSION", "1.0.0"),
		"timestamps": "true",
	}
	
	return request, nil
}

// PrepareBedrockTurnRequest creates a BedrockRequest for a specific turn
func PrepareBedrockTurnRequest(systemPrompt string, userMessage BedrockMessage, bedrockConfig *schema.BedrockConfig) (BedrockRequest, error) {
	// Get base request
	request, err := PrepareBedrockRequest(systemPrompt, bedrockConfig)
	if err != nil {
		return BedrockRequest{}, err
	}
	
	// Add user message
	request.Messages = append(request.Messages, userMessage)
	
	return request, nil
}

// CreateUserMessageWithImage creates a user message with text and optional image
func CreateUserMessageWithImage(text, imageData, imageType string) BedrockMessage {
	message := BedrockMessage{
		Role: "user",
		Content: []MessagePart{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	
	// Add image if provided
	if imageData != "" {
		imagePart := MessagePart{
			Type: "image",
			Source: &ImageSource{
				Type:      "base64",
				MediaType: getImageMediaType(imageType),
				Data:      imageData,
			},
		}
		message.Content = append(message.Content, imagePart)
	}
	
	return message
}

// getImageMediaType returns the media type for an image format
func getImageMediaType(imageType string) string {
	imageType = strings.ToLower(imageType)
	
	switch imageType {
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	default:
		// Default to JPEG if unknown, but this should be validated earlier
		return "image/jpeg"
	}
}

// ValidateImageMediaType checks if the media type is supported by Bedrock
func ValidateImageMediaType(mediaType string) bool {
	return supportedImageMediaTypes[mediaType]
}

// EstimateTokenUsage estimates token usage for a system prompt
func EstimateTokenUsage(systemPrompt string) int {
	// Rough estimate: average of 4 characters per token for English text
	return len(systemPrompt) / 4
}

// ExtractBedrockResponse parses a JSON response from Bedrock
func ExtractBedrockResponse(responseBody []byte) (*BedrockResponse, error) {
	var response BedrockResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Bedrock response: %w", err)
	}
	
	return &response, nil
}

// ParseImageFromResponse extracts image data from a Bedrock response
func ParseImageFromResponse(response *BedrockResponse) (string, string, error) {
	if response == nil {
		return "", "", fmt.Errorf("response is nil")
	}
	
	for _, content := range response.Content {
		if content.Type == "image" && content.Source != nil {
			return content.Source.Data, content.Source.MediaType, nil
		}
	}
	
	return "", "", fmt.Errorf("no image found in response")
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

// getFloatEnv retrieves a float environment variable with a default value
func getFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	
	return floatValue
}