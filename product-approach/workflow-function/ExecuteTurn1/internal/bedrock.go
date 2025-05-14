package internal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockClient handles Bedrock API interactions
type BedrockClient struct {
	client  *bedrockruntime.Client
	modelID string
}

// NewBedrockClient creates a new Bedrock client
func NewBedrockClient(region, modelID string) (*BedrockClient, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, NewBedrockError(
			"Failed to load AWS configuration",
			"AWS_CONFIG_ERROR",
			false,
		)
	}

	// Create Bedrock runtime client
	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockClient{
		client:  client,
		modelID: modelID,
	}, nil
}

// InvokeModel calls the Bedrock API with the Turn 1 prompt
func (bc *BedrockClient) InvokeModel(ctx context.Context, input *ExecuteTurn1Input) (*BedrockResponse, int64, error) {
	// Start timing
	startTime := time.Now()

	// Construct Bedrock request
	bedrockReq, err := bc.constructBedrockRequest(input)
	if err != nil {
		return nil, 0, err
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, 0, NewInternalError("bedrock request marshaling", err)
	}

	// Prepare invoke model input
	invokeInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(bc.modelID),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	// Call Bedrock API
	result, err := bc.client.InvokeModel(ctx, invokeInput)
	if err != nil {
		return nil, 0, bc.handleBedrockError(err)
	}

	// Calculate latency
	latency := time.Since(startTime)

	// Parse response
	response, err := bc.parseBedrockResponse(result.Body)
	if err != nil {
		return nil, latency.Milliseconds(), err
	}

	return response, latency.Milliseconds(), nil
}

// constructBedrockRequest constructs the Bedrock API request
func (bc *BedrockClient) constructBedrockRequest(input *ExecuteTurn1Input) (*BedrockRequest, error) {
	// Extract current prompt
	currentPrompt, err := ExtractAndValidateCurrentPrompt(input)
	if err != nil {
		return nil, err
	}

	// Get the current prompt messages
	messages := currentPrompt.Messages

	// Validate that we have the expected message structure
	if len(messages) == 0 {
		return nil, NewValidationError("No messages in current prompt", nil)
	}

	// Ensure the image S3 location has bucket owner information
	for i := range messages {
		for j := range messages[i].Content {
			if messages[i].Content[j].Image != nil {
				// Extract bucket owner from verification context if not present
				if messages[i].Content[j].Image.Source.S3Location.BucketOwner == "" {
					// Default to the bucket owner from image metadata if available
					if input.Images != nil {
						messages[i].Content[j].Image.Source.S3Location.BucketOwner = ExtractBucketOwner(input)
					} else {
						// Extract from URI or use default
						_, _, err := ParseS3URI(messages[i].Content[j].Image.Source.S3Location.URI)
						if err != nil {
							return nil, NewValidationError("Invalid S3 URI in image source", 
								map[string]interface{}{"uri": messages[i].Content[j].Image.Source.S3Location.URI})
						}
						// Default bucket owner (this should be set from environment or context)
						messages[i].Content[j].Image.Source.S3Location.BucketOwner = "defaultBucketOwner"
					}
				}
			}
		}
	}

	// Construct the request
	request := &BedrockRequest{
		AnthropicVersion: input.BedrockConfig.AnthropicVersion,
		MaxTokens:       input.BedrockConfig.MaxTokens,
		Thinking:        input.BedrockConfig.Thinking,
		Messages:        messages,
	}

	return request, nil
}

// parseBedrockResponse parses the Bedrock API response
func (bc *BedrockClient) parseBedrockResponse(body []byte) (*BedrockResponse, error) {
	var response BedrockResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, NewParsingError("Bedrock response", err)
	}

	// Validate required fields
	if response.Type != "message" {
		return nil, NewBedrockError(
			"Unexpected response type from Bedrock",
			"INVALID_RESPONSE_TYPE",
			false,
		)
	}

	if response.Role != "assistant" {
		return nil, NewBedrockError(
			"Unexpected role in Bedrock response",
			"INVALID_RESPONSE_ROLE",
			false,
		)
	}

	// Calculate total tokens if not provided
	if response.Usage.Total == 0 {
		response.Usage.Total = response.Usage.Input + response.Usage.Output + response.Usage.Thinking
	}

	return &response, nil
}

// handleBedrockError converts Bedrock API errors to our error types
func (bc *BedrockClient) handleBedrockError(err error) error {
	// Type assertion to check for specific AWS SDK error types
	switch err.(type) {
	case *types.ThrottlingException:
		return NewBedrockThrottlingError()
	case *types.ServiceQuotaExceededException:
		return NewBedrockTokenLimitError(0, 0) // Token counts not available in this context
	case *types.ValidationException:
		return NewBedrockError(
			"Validation error from Bedrock API",
			"BEDROCK_VALIDATION",
			false,
		)
	case *types.AccessDeniedException:
		return NewBedrockError(
			"Access denied to Bedrock API",
			"BEDROCK_ACCESS_DENIED",
			false,
		)
	case *types.ResourceNotFoundException:
		return NewBedrockModelError(bc.modelID)
	case *types.InternalServerException:
		return NewBedrockError(
			"Internal server error in Bedrock API",
			"BEDROCK_INTERNAL_ERROR",
			true, // This is retryable
		)
	default:
		// Generic error handling
		return NewBedrockError(
			err.Error(),
			"BEDROCK_UNKNOWN_ERROR",
			true, // Assume retryable by default
		)
	}
}

// ValidateModel validates that the specified model is available
func (bc *BedrockClient) ValidateModel(ctx context.Context) error {
	// This is a simple ping to verify the model is accessible
	// We'll create a minimal request to test connectivity
	testRequest := &BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:       10,
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []MessageContent{
					{
						Type: "text",
						Text: aws.String("Hello"),
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(testRequest)
	if err != nil {
		return NewInternalError("test request marshaling", err)
	}

	invokeInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(bc.modelID),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	// Create a timeout context for the validation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = bc.client.InvokeModel(ctx, invokeInput)
	if err != nil {
		return bc.handleBedrockError(err)
	}

	return nil
}

// GetModelInfo returns information about the current model
func (bc *BedrockClient) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"modelId": bc.modelID,
		"region":  bc.client.Options().Region,
	}
}

// EstimateTokenCount estimates the token count for a text input
// This is a rough estimation and may not be perfectly accurate
func (bc *BedrockClient) EstimateTokenCount(text string) int {
	// Rough estimation: 1 token per 4 characters for English text
	// This is a simplified estimation and should be replaced with
	// actual tokenization if more accuracy is needed
	return len(text) / 4
}

// ConvertToBedrockFormat converts internal types to Bedrock API format
func ConvertToBedrockFormat(input *ExecuteTurn1Input) (*BedrockRequest, error) {
	// Extract current prompt
	currentPrompt, err := ExtractAndValidateCurrentPrompt(input)
	if err != nil {
		return nil, err
	}

	request := &BedrockRequest{
		AnthropicVersion: input.BedrockConfig.AnthropicVersion,
		MaxTokens:       input.BedrockConfig.MaxTokens,
		Thinking:        input.BedrockConfig.Thinking,
		Messages:        currentPrompt.Messages,
	}

	// Ensure all required fields are present
	if request.AnthropicVersion == "" {
		return nil, NewMissingFieldError("anthropic_version")
	}

	if request.MaxTokens <= 0 {
		return nil, NewInvalidFieldError("max_tokens", request.MaxTokens, "positive integer")
	}

	return request, nil
}

// Helper function to extract text content from Bedrock response
func ExtractTextFromResponse(response *BedrockResponse) string {
	if response == nil || len(response.Content) == 0 {
		return ""
	}

	for _, content := range response.Content {
		if content.Type == "text" && content.Text != nil {
			return *content.Text
		}
	}

	return ""
}