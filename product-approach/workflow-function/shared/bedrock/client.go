// Package bedrock provides a standardized client for Bedrock Converse API interactions
package bedrock

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockClient handles Bedrock API interactions using the Converse API
type BedrockClient struct {
	client  *bedrockruntime.Client
	modelID string
	config  *ClientConfig
}

// ClientConfig contains configuration for the Bedrock client
type ClientConfig struct {
	Region           string
	AnthropicVersion string
	MaxTokens        int
	Temperature      float64
	TopP             float64
	ThinkingType     string
	BudgetTokens     int
}

// NewBedrockClient creates a new Bedrock client with Converse API support
func NewBedrockClient(ctx context.Context, modelID string, clientConfig *ClientConfig) (*BedrockClient, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(clientConfig.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create Bedrock runtime client
	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockClient{
		client:  client,
		modelID: modelID,
		config:  clientConfig,
	}, nil
}

// Converse calls the Bedrock Converse API
func (bc *BedrockClient) Converse(ctx context.Context, request *ConverseRequest) (*ConverseResponse, int64, error) {
	// Start timing
	startTime := time.Now()

	log.Printf("Using Converse API for model: %s", bc.modelID)

	// Validate request
	if err := ValidateConverseRequest(request); err != nil {
		return nil, 0, fmt.Errorf("invalid request: %w", err)
	}

	// Convert messages to AWS SDK format
	messages := make([]types.Message, len(request.Messages))
	for i, msg := range request.Messages {
		// Convert content blocks
		contentBlocks := make([]types.ContentBlock, len(msg.Content))
		for j, content := range msg.Content {
			switch content.Type {
			case "text":
				contentBlocks[j] = &types.ContentBlockMemberText{
					Value: content.Text,
				}
				log.Printf("Added text content block: %d chars", len(content.Text))
			case "image":
				if content.Image != nil {
					// Validate image format (Bedrock only supports jpeg and png for Converse API)
					if content.Image.Format != "jpeg" && content.Image.Format != "png" {
						return nil, 0, fmt.Errorf("unsupported image format: %s (only jpeg and png are supported)", content.Image.Format)
					}
					
					// Create the image block with the appropriate source type
					imageBlock := types.ImageBlock{
						Format: types.ImageFormat(content.Image.Format),
					}
					
					// Determine the source type based on what's provided
					if content.Image.Source.Bytes != "" {
						// Decode base64 string to bytes
						decodedBytes, err := base64.StdEncoding.DecodeString(content.Image.Source.Bytes)
						if err != nil {
							return nil, 0, fmt.Errorf("failed to decode base64 image data: %w", err)
						}
						
						// Use bytes source
						imageBlock.Source = &types.ImageSourceMemberBytes{
							Value: decodedBytes,
						}
						log.Printf("Added image content block for format: %s, with bytes source (size: %d bytes)", 
							content.Image.Format, len(decodedBytes))
					} else {
						return nil, 0, fmt.Errorf("image source must be provided as bytes")
					}
					
					contentBlocks[j] = &types.ContentBlockMemberImage{
						Value: imageBlock,
					}
				} else {
					return nil, 0, fmt.Errorf("image content block has nil image field")
				}
			default:
				return nil, 0, fmt.Errorf("unsupported content type: %s", content.Type)
			}
		}
		
		// Check for empty content
		if len(contentBlocks) == 0 {
			return nil, 0, fmt.Errorf("message must contain at least one content block")
		}
		
		messages[i] = types.Message{
			Role:    types.ConversationRole(msg.Role),
			Content: contentBlocks,
		}
	}
	
	// Create inference config
	inferenceConfig := &types.InferenceConfiguration{
		MaxTokens: aws.Int32(int32(request.InferenceConfig.MaxTokens)),
	}
	
	if request.InferenceConfig.Temperature != nil {
		inferenceConfig.Temperature = aws.Float32(float32(*request.InferenceConfig.Temperature))
	}
	
	if request.InferenceConfig.TopP != nil {
		inferenceConfig.TopP = aws.Float32(float32(*request.InferenceConfig.TopP))
	}
	
	if len(request.InferenceConfig.StopSequences) > 0 {
		inferenceConfig.StopSequences = request.InferenceConfig.StopSequences
	}
	
	// Create Converse input
	converseInput := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(bc.modelID),
		Messages:        messages,
		InferenceConfig: inferenceConfig,
	}
	
	// Add system prompt if provided
	if request.System != "" {
		converseInput.System = []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{
				Value: request.System,
			},
		}
		log.Printf("Added system prompt: %d chars", len(request.System))
	}
	
	// Add guardrail config if provided
	if request.GuardrailConfig != nil {
		guardrailConfig := &types.GuardrailConfiguration{
			GuardrailIdentifier: aws.String(request.GuardrailConfig.GuardrailIdentifier),
		}
		
		if request.GuardrailConfig.GuardrailVersion != "" {
			guardrailConfig.GuardrailVersion = aws.String(request.GuardrailConfig.GuardrailVersion)
		}
		
		converseInput.GuardrailConfig = guardrailConfig
		log.Printf("Added guardrail config with identifier: %s", request.GuardrailConfig.GuardrailIdentifier)
	}
	
	// Log request details
	log.Printf("Sending Converse API request to model %s with %d messages", bc.modelID, len(messages))
	
	// Call Bedrock Converse API
	result, err := bc.client.Converse(ctx, converseInput)
	if err != nil {
		return nil, 0, bc.handleBedrockError(err)
	}
	
	// Calculate latency
	latency := time.Since(startTime)
	
	// Convert response to our format
	response, err := bc.convertFromBedrockResponse(result, bc.modelID)
	if err != nil {
		return nil, latency.Milliseconds(), fmt.Errorf("failed to convert response: %w", err)
	}
	
	log.Printf("Bedrock API call completed in %v with %d tokens total", 
		latency, response.Usage.TotalTokens)
	
	return response, latency.Milliseconds(), nil
}

// convertFromBedrockResponse converts Bedrock SDK response to our format
func (bc *BedrockClient) convertFromBedrockResponse(result *bedrockruntime.ConverseOutput, modelID string) (*ConverseResponse, error) {
	var content []ContentBlock
	
	// Extract content from the response
	if result.Output != nil {
		switch v := result.Output.(type) {
		case *types.ConverseOutputMemberMessage:
			for _, contentBlock := range v.Value.Content {
				switch cb := contentBlock.(type) {
				case *types.ContentBlockMemberText:
					content = append(content, ContentBlock{
						Type: "text",
						Text: cb.Value,
					})
				// Note: Image output is not supported by Bedrock currently
				default:
					log.Printf("Unknown content block type in response: %T", cb)
				}
			}
		default:
			log.Printf("Unknown output type in response: %T", v)
		}
	}
	
	// Extract usage information
	var usage *TokenUsage
	if result.Usage != nil {
		usage = &TokenUsage{
			InputTokens:  int(*result.Usage.InputTokens),
			OutputTokens: int(*result.Usage.OutputTokens),
			TotalTokens:  int(*result.Usage.InputTokens + *result.Usage.OutputTokens),
		}
	} else {
		// Provide default usage if not available
		usage = &TokenUsage{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		}
	}
	
	// Extract stop reason
	var stopReason string
	if result.StopReason != "" {
		stopReason = string(result.StopReason)
	}
	
	// Extract metrics
	var metrics *ResponseMetrics
	if result.Metrics != nil && result.Metrics.LatencyMs != nil {
		metrics = &ResponseMetrics{
			LatencyMs: int64(*result.Metrics.LatencyMs),
		}
	}
	
	// Generate request ID if not available
	requestID := "req-" + time.Now().Format("20060102-150405")
	
	return &ConverseResponse{
		RequestID:  requestID,
		ModelID:    modelID,
		StopReason: stopReason,
		Content:    content,
		Usage:      usage,
		Metrics:    metrics,
	}, nil
}

// handleBedrockError converts AWS SDK errors to our error types
func (bc *BedrockClient) handleBedrockError(err error) error {
	log.Printf("Bedrock API error: %v", err)
	
	// Extract useful information from the error
	return fmt.Errorf("bedrock API error: %w", err)
}

// ValidateModel validates that the specified model is available
func (bc *BedrockClient) ValidateModel(ctx context.Context) error {
	log.Printf("Validating model %s with Bedrock API", bc.modelID)
	
	// Create a minimal validation using the proper Converse API
	converseInput := &bedrockruntime.ConverseInput{
		ModelId: aws.String(bc.modelID),
		Messages: []types.Message{
			{
				Role: types.ConversationRole("user"),
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "Test",
					},
				},
			},
		},
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens: aws.Int32(10),
		},
	}
	
	// Create a timeout context for the validation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Call Converse API
	_, err := bc.client.Converse(ctx, converseInput)
	if err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}
	
	log.Printf("Model %s validated successfully with Bedrock API", bc.modelID)
	return nil
}

// GetModelInfo returns information about the current model and API configuration
func (bc *BedrockClient) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"modelId":          bc.modelID,
		"region":           bc.client.Options().Region,
		"apiType":          "Converse",
		"anthropicVersion": bc.config.AnthropicVersion,
		"maxTokens":        bc.config.MaxTokens,
		"temperature":      bc.config.Temperature,
		"topP":             bc.config.TopP,
		"thinkingType":     bc.config.ThinkingType,
		"budgetTokens":     bc.config.BudgetTokens,
	}
}

// EstimateTokenCount estimates the token count for a text input
func (bc *BedrockClient) EstimateTokenCount(text string) int {
	// Rough estimation: 1 token per 4 characters for English text
	// This is a simplified estimation and should be replaced with
	// actual tokenization if more accuracy is needed
	return len(text) / 4
}

// CreateClientConfig creates a client configuration from environment variables
func CreateClientConfig(region, anthropicVersion string, maxTokens int, thinkingType string, budgetTokens int) *ClientConfig {
	return &ClientConfig{
		Region:           region,
		AnthropicVersion: anthropicVersion,
		MaxTokens:        maxTokens,
		Temperature:      0.7, // Default temperature
		TopP:             0.9, // Default topP
		ThinkingType:     thinkingType,
		BudgetTokens:     budgetTokens,
	}
}

// GetTextFromResponse is a convenience method to extract text from a response
func (bc *BedrockClient) GetTextFromResponse(response *ConverseResponse) string {
	return ExtractTextFromResponse(response)
}