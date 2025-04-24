package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	//"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockProvider provides access to Amazon Bedrock
type BedrockProvider struct {
	client     *bedrockruntime.Client
	modelID    string
	maxRetries int
}

// NewBedrockProvider creates a new Bedrock provider
func NewBedrockProvider(region, modelID string, maxRetries int) (*BedrockProvider, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Bedrock client
	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockProvider{
		client:     client,
		modelID:    modelID,
		maxRetries: maxRetries,
	}, nil
}

// AnthropicRequest represents the request structure for Claude models
type AnthropicRequest struct {
	AnthropicVersion string       `json:"anthropic_version"`
	MaxTokens        int          `json:"max_tokens"`
	Thinking         ThinkingConfig `json:"thinking"`
	System           string       `json:"system"`
	Messages         []Message    `json:"messages"`
}

// ThinkingConfig configures the model's thinking behavior
type ThinkingConfig struct {
	Type       string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// Message represents a conversation message
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Content represents content in a message
type Content struct {
	Type   string          `json:"type"`
	Text   string          `json:"text,omitempty"`
	Source *ImageSource    `json:"source,omitempty"`
}

// ImageSource represents the source of an image
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// AnthropicResponse represents the response structure from Claude models
type AnthropicResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Content []Content `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// InvokeModel calls the Bedrock API with the provided prompts and images
func (p *BedrockProvider) InvokeModel(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	images []string, // Base64 encoded
	conversationContext map[string]interface{},
) (string, map[string]interface{}, error) {
	// Prepare message content
	contents := []Content{
		{
			Type: "text",
			Text: userPrompt,
		},
	}

	// Add images to content
	for _, imageBase64 := range images {
		contents = append(contents, Content{
			Type: "image",
			Source: &ImageSource{
				Type:      "base64",
				MediaType: "image/jpeg", // Assuming JPEG; adjust based on actual content
				Data:      imageBase64,
			},
		})
	}

	// Prepare request
	request := AnthropicRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        24000,
		Thinking: ThinkingConfig{
			Type:        "enabled",
			BudgetTokens: 16000,
		},
		System: systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: contents,
			},
		},
	}

	// Convert request to bytes
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Initialize retry counter
	retryCount := 0
	var response *bedrockruntime.InvokeModelOutput
	var invocationErr error

	// Invoke model with retry logic
	for retryCount <= p.maxRetries {
		// Call Bedrock API
		response, invocationErr = p.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(p.modelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        requestBytes,
		})

		// Check for success
		if invocationErr == nil {
			break
		}

		// Decide whether to retry
		if retryCount >= p.maxRetries {
			return "", nil, fmt.Errorf("failed to invoke model after %d retries: %w", p.maxRetries, invocationErr)
		}

		// Exponential backoff
		backoffTime := time.Duration(1<<uint(retryCount)) * time.Second
		time.Sleep(backoffTime)
		retryCount++
	}

	// Parse response
	var anthropicResponse AnthropicResponse
	if err := json.Unmarshal(response.Body, &anthropicResponse); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract text from response
	responseText := ""
	for _, content := range anthropicResponse.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"model":        anthropicResponse.Model,
		"id":           anthropicResponse.ID,
		"inputTokens":  anthropicResponse.Usage.InputTokens,
		"outputTokens": anthropicResponse.Usage.OutputTokens,
		"totalTokens":  anthropicResponse.Usage.InputTokens + anthropicResponse.Usage.OutputTokens,
		"retryCount":   retryCount,
	}

	return responseText, metadata, nil
}