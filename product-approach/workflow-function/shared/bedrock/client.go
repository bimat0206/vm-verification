// Package bedrock provides a standardized client for Bedrock Converse API interactions
package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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

	// Convert messages to AWS SDK format
	messages := make([]map[string]interface{}, len(request.Messages))
	for i, msg := range request.Messages {
		// Convert content blocks
		contentBlocks := make([]map[string]interface{}, len(msg.Content))
		for j, content := range msg.Content {
			var contentBlock map[string]interface{}
			
			switch content.Type {
			case "text":
				contentBlock = map[string]interface{}{
					"type": "text",
					"text": content.Text,
				}
			case "image":
				if content.Image == nil {
					return nil, 0, fmt.Errorf("image content cannot be nil")
				}
				
				// Ensure we're correctly nested as per Bedrock API requirements
				// The API needs: {type:"image", image:{format:"png", source:{type:"s3", s3Location:{uri:"...", bucketOwner:"..."}}}}
				contentBlock = map[string]interface{}{
					"type": "image",
					"image": map[string]interface{}{
						"format": content.Image.Format,
						"source": map[string]interface{}{
							"type": "s3",
							"s3Location": map[string]interface{}{
								"uri":         content.Image.Source.S3Location.URI,
								"bucketOwner": content.Image.Source.S3Location.BucketOwner,
							},
						},
					},
				}
				
				// Log the structure for debugging
				logBytes, _ := json.Marshal(contentBlock)
				log.Printf("Image content block: %s", string(logBytes))
			default:
				return nil, 0, fmt.Errorf("unsupported content type: %s", content.Type)
			}
			
			contentBlocks[j] = contentBlock
		}
		
		messages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": contentBlocks,
		}
	}
	
	// Create inference config
	inferenceConfig := map[string]interface{}{
		"maxTokens": request.InferenceConfig.MaxTokens,
	}
	
	if request.InferenceConfig.Temperature != nil {
		inferenceConfig["temperature"] = *request.InferenceConfig.Temperature
	}
	
	if request.InferenceConfig.TopP != nil {
		inferenceConfig["topP"] = *request.InferenceConfig.TopP
	}
	
	if len(request.InferenceConfig.StopSequences) > 0 {
		inferenceConfig["stopSequences"] = request.InferenceConfig.StopSequences
	}
	
	// Create request body
	requestBody := map[string]interface{}{
		"modelId":         bc.modelID,
		"messages":        messages,
		"inferenceConfig": inferenceConfig,
	}
	
	// Add system prompt if provided
	if request.System != "" {
		requestBody["system"] = request.System
	}
	
	// Add guardrail config if provided
	if request.GuardrailConfig != nil {
		requestBody["guardrailConfig"] = map[string]interface{}{
			"guardrailIdentifier": request.GuardrailConfig.GuardrailIdentifier,
		}
		
		if request.GuardrailConfig.GuardrailVersion != "" {
			requestBody["guardrailConfig"].(map[string]interface{})["guardrailVersion"] = request.GuardrailConfig.GuardrailVersion
		}
	}
	
	// Marshal request to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Log the full request for debugging
	log.Printf("Full Bedrock request: %s", string(requestJSON))
	
	// Create InvokeModel input
	invokeInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(bc.modelID),
		Body:        requestJSON,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	
	// Call InvokeModel API
	result, err := bc.client.InvokeModel(ctx, invokeInput)
	if err != nil {
		return nil, 0, bc.handleBedrockError(err)
	}
	
	// Calculate latency
	latency := time.Since(startTime)
	
	// Parse response
	var responseBody map[string]interface{}
	if err := json.Unmarshal(result.Body, &responseBody); err != nil {
		return nil, latency.Milliseconds(), fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// Extract response data
	response, err := bc.parseConverseResponse(responseBody, bc.modelID)
	if err != nil {
		return nil, latency.Milliseconds(), fmt.Errorf("failed to parse response: %w", err)
	}
	
	log.Printf("Bedrock API call completed in %v", latency)
	return response, latency.Milliseconds(), nil
}

// parseConverseResponse parses the Converse API response
func (bc *BedrockClient) parseConverseResponse(responseBody map[string]interface{}, modelID string) (*ConverseResponse, error) {
	// Extract output
	output, ok := responseBody["output"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing output in response")
	}
	
	// Extract message
	message, ok := output["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing message in output")
	}
	
	// Extract content
	contentRaw, ok := message["content"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing content in message")
	}
	
	// Parse content blocks
	var content []ContentBlock
	for _, blockRaw := range contentRaw {
		block, ok := blockRaw.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Check for text content
		if textRaw, ok := block["text"].(string); ok {
			content = append(content, ContentBlock{
				Type: "text",
				Text: textRaw,
			})
		}
	}
	
	// Extract usage
	var usage *TokenUsage
	if usageRaw, ok := responseBody["usage"].(map[string]interface{}); ok {
		inputTokens, _ := usageRaw["inputTokens"].(float64)
		outputTokens, _ := usageRaw["outputTokens"].(float64)
		totalTokens, _ := usageRaw["totalTokens"].(float64)
		
		usage = &TokenUsage{
			InputTokens:  int(inputTokens),
			OutputTokens: int(outputTokens),
			TotalTokens:  int(totalTokens),
		}
	}
	
	// Extract stop reason
	var stopReason string
	if stopReasonRaw, ok := responseBody["stopReason"].(string); ok {
		stopReason = stopReasonRaw
	}
	
	// Extract metrics
	var metrics *ResponseMetrics
	if metricsRaw, ok := responseBody["metrics"].(map[string]interface{}); ok {
		if latencyRaw, ok := metricsRaw["latencyMs"].(float64); ok {
			metrics = &ResponseMetrics{
				LatencyMs: int64(latencyRaw),
			}
		}
	}
	
	// Create response
	response := &ConverseResponse{
		RequestID:  "", // Not available in the response body
		ModelID:    modelID,
		StopReason: stopReason,
		Content:    content,
		Usage:      usage,
		Metrics:    metrics,
	}
	
	return response, nil
}

// handleBedrockError converts AWS SDK errors to our error types
func (bc *BedrockClient) handleBedrockError(err error) error {
	log.Printf("Bedrock API error: %v", err)
	return fmt.Errorf("bedrock API error: %w", err)
}

// ValidateModel validates that the specified model is available
func (bc *BedrockClient) ValidateModel(ctx context.Context) error {
	log.Printf("Validating model %s with Bedrock API", bc.modelID)
	
	// Create a minimal request
	requestBody := map[string]interface{}{
		"modelId": bc.modelID,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Test",
					},
				},
			},
		},
		"inferenceConfig": map[string]interface{}{
			"maxTokens": 10,
		},
	}
	
	// Marshal request to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal validation request: %w", err)
	}
	
	// Create InvokeModel input
	invokeInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(bc.modelID),
		Body:        requestJSON,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	
	// Create a timeout context for the validation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Call InvokeModel API
	_, err = bc.client.InvokeModel(ctx, invokeInput)
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