package bedrock

import (
	"context"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
)

// Client provides a wrapper around the shared Bedrock client
type Client struct {
	bedrockClient *BedrockClient
	logger        logger.Logger
	config        *Config
}

// BedrockClient is a local interface to represent the shared Bedrock client
type BedrockClient struct {
	Converse           func(context.Context, *ConverseRequest) (*ConverseResponse, int64, error)
	GetTextFromResponse func(*ConverseResponse) string
}

// ConverseRequest is a local interface to represent the shared Bedrock request
type ConverseRequest struct {
	Messages        []Message
	System          string
	InferenceConfig InferenceConfig
}

// Message represents a message in the conversation
type Message struct {
	Role    string
	Content []ContentBlock
}

// ContentBlock represents a content block for text or image
type ContentBlock struct {
	Type  string
	Text  string
	Image *Image
}

// Image represents an image in the message
type Image struct {
	Format string
	Source ImageSource
}

// ImageSource provides the source for an image
type ImageSource struct {
	Bytes string
}

// InferenceConfig provides model configuration for inference
type InferenceConfig struct {
	MaxTokens     int
	Temperature   *float64
	TopP          *float64
	StopSequences []string
}

// ConverseResponse is the response from the Bedrock Converse API
type ConverseResponse struct {
	Content    string
	StopReason string
	ModelID    string
	Usage      *TokenUsage
}

// TokenUsage contains token usage information
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// Config holds Bedrock-specific configuration
type Config struct {
	ModelID          string
	AnthropicVersion string
	MaxTokens        int
	Temperature      float64
	ThinkingType     string
	ThinkingBudget   int
	Timeout          time.Duration
}

// NewClient creates a new Bedrock client
func NewClient(bedrockClient *BedrockClient, logger logger.Logger, config *Config) *Client {
	return &Client{
		bedrockClient: bedrockClient,
		logger:        logger.WithFields(map[string]interface{}{"component": "BedrockClient"}),
		config:        config,
	}
}

// ProcessTurn1 handles the complete Turn 1 processing with Bedrock
func (c *Client) ProcessTurn1(ctx context.Context, prompt *schema.CurrentPrompt, images *schema.ImageData) (*schema.TurnResponse, error) {
	c.logger.Info("Starting Turn1 processing with Bedrock", map[string]interface{}{
		"modelId": c.config.ModelID,
		"timeout": c.config.Timeout,
	})

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Build Bedrock request
	request, err := c.BuildTurn1Request(prompt, images)
	if err != nil {
		return nil, wferrors.WrapError(err, wferrors.ErrorTypeBedrock, 
			"Failed to build Bedrock request", false)
	}

	// Execute Bedrock call
	c.logger.Info("Invoking Bedrock API", map[string]interface{}{
		"promptId": prompt.PromptId,
		"maxTokens": c.config.MaxTokens,
	})
	
	response, latencyMs, err := c.bedrockClient.Converse(ctx, request)
	if err != nil {
		return nil, c.handleBedrockError(err)
	}
	
	// Log successful API call
	c.logger.Info("Bedrock API call successful", map[string]interface{}{
		"latencyMs":   latencyMs,
		"tokenUsage":  response.Usage.TotalTokens,
		"stopReason":  response.StopReason,
	})

	// Parse Bedrock response into Turn response
	promptText := ""
	if len(prompt.Messages) > 0 && len(prompt.Messages[0].Content) > 0 {
		promptText = prompt.Messages[0].Content[0].Text
	} else if prompt.Text != "" {
		promptText = prompt.Text
	}
	
	turnResponse, err := c.ParseTurn1Response(response, latencyMs, promptText, images)
	if err != nil {
		return nil, wferrors.WrapError(err, wferrors.ErrorTypeInternal, 
			"Failed to parse Turn1 response", false)
	}

	return turnResponse, nil
}

// BuildTurn1Request constructs the Bedrock Converse API request
func (c *Client) BuildTurn1Request(prompt *schema.CurrentPrompt, images *schema.ImageData) (*ConverseRequest, error) {
	// Validate inputs
	if prompt == nil {
		return nil, wferrors.NewValidationError("Prompt is nil", nil)
	}

	if len(prompt.Messages) == 0 {
		return nil, wferrors.NewValidationError("Prompt has no messages", nil)
	}

	// Convert schema messages to Bedrock request format
	bedrockMessages := make([]Message, 0, len(prompt.Messages))

	for _, msg := range prompt.Messages {
		// Create content blocks
		contentBlocks := make([]ContentBlock, 0)

		// Add text content
		if len(msg.Content) > 0 && msg.Content[0].Text != "" {
			contentBlocks = append(contentBlocks, ContentBlock{
				Type: "text",
				Text: msg.Content[0].Text,
			})
		}

		// Add image content if present and we have image data
		if len(msg.Content) > 1 && msg.Content[1].Image != nil && images != nil {
			// Get image data from schema.ImageData
			imageBlock, err := c.createImageBlock(images)
			if err != nil {
				return nil, err
			}
			
			if imageBlock != nil {
				contentBlocks = append(contentBlocks, *imageBlock)
			}
		}

		// Add message with content blocks
		bedrockMessages = append(bedrockMessages, Message{
			Role:    msg.Role,
			Content: contentBlocks,
		})
	}

	// Create inference config
	inferenceConfig := InferenceConfig{
		MaxTokens:    c.config.MaxTokens,
		Temperature:  &c.config.Temperature,
		TopP:         nil, // Use Bedrock default
		StopSequences: []string{},
	}

	// Get system prompt
	systemPrompt := ""
	if prompt.Metadata != nil {
		if sp, ok := prompt.Metadata["systemPrompt"].(string); ok {
			systemPrompt = sp
		}
	}

	// Assemble final request
	request := &ConverseRequest{
		Messages:        bedrockMessages,
		System:          systemPrompt,
		InferenceConfig: inferenceConfig,
	}

	return request, nil
}

// createImageBlock creates an image content block from schema.ImageData
func (c *Client) createImageBlock(images *schema.ImageData) (*ContentBlock, error) {
	if images == nil {
		return nil, wferrors.NewValidationError("ImageData is nil", nil)
	}

	// Get the reference image (primary or fallback)
	var imageInfo *schema.ImageInfo
	if images.Reference != nil {
		imageInfo = images.Reference
	} else if images.ReferenceImage != nil {
		imageInfo = images.ReferenceImage
	} else {
		return nil, wferrors.NewValidationError("No reference image found in ImageData", nil)
	}

	// Check if base64 data is available
	if !imageInfo.HasBase64Data() {
		return nil, wferrors.NewValidationError("Image has no Base64 data", map[string]interface{}{
			"imageUrl": imageInfo.URL,
		})
	}

	// Get Base64 data
	base64Data := imageInfo.GetBase64Data()
	if base64Data == "" {
		return nil, wferrors.NewValidationError("Failed to retrieve Base64 data", map[string]interface{}{
			"imageUrl": imageInfo.URL,
		})
	}

	// Create the source structure
	source := ImageSource{
		Bytes: base64Data,
	}

	// Create the image block
	return &ContentBlock{
		Type: "image",
		Image: &Image{
			Format: imageInfo.Format,
			Source: source,
		},
	}, nil
}

// ParseTurn1Response processes the Bedrock response into a schema.TurnResponse
func (c *Client) ParseTurn1Response(bedrockResp *ConverseResponse, latencyMs int64, promptText string, images *schema.ImageData) (*schema.TurnResponse, error) {
	// Extract text from response
	responseText := c.bedrockClient.GetTextFromResponse(bedrockResp)

	// Extract thinking content if enabled
	thinking := ""
	if c.config.ThinkingType != "" {
		thinking = c.extractThinking(responseText)
	}

	// Build image URLs map
	imageUrls := make(map[string]string)
	if images != nil && images.Reference != nil {
		imageUrls["reference"] = images.Reference.URL
	} else if images != nil && images.ReferenceImage != nil {
		imageUrls["reference"] = images.ReferenceImage.URL
	}

	// Create token usage
	tokenUsage := &schema.TokenUsage{
		InputTokens:    bedrockResp.Usage.InputTokens,
		OutputTokens:   bedrockResp.Usage.OutputTokens,
		TotalTokens:    bedrockResp.Usage.TotalTokens,
		ThinkingTokens: 0, // We don't know this yet
	}

	// Create Bedrock API response
	bedrockApiResp := schema.BedrockApiResponse{
		Content:    responseText,
		StopReason: bedrockResp.StopReason,
	}

	// Create Turn response
	turnResp := &schema.TurnResponse{
		TurnId:     1, // Turn 1
		Timestamp:  schema.FormatISO8601(),
		Prompt:     promptText,
		ImageUrls:  imageUrls,
		Response:   bedrockApiResp,
		LatencyMs:  latencyMs,
		TokenUsage: tokenUsage,
		Stage:      schema.StatusTurn1Completed,
		Metadata:   c.buildResponseMetadata(thinking, bedrockResp.ModelID),
	}

	return turnResp, nil
}

// extractThinking extracts thinking content from the response text
func (c *Client) extractThinking(text string) string {
	// Simple implementation to extract thinking content
	// In production, this would use a more sophisticated method
	thinking := ""
	
	// Apply thinking budget if configured
	if thinking != "" && c.config.ThinkingBudget > 0 && len(thinking) > c.config.ThinkingBudget {
		c.logger.Debug("Truncating thinking content", map[string]interface{}{
			"originalLength": len(thinking),
			"budgetLength":   c.config.ThinkingBudget,
		})
		thinking = thinking[:c.config.ThinkingBudget] + "... [truncated]"
	}
	
	return thinking
}

// buildResponseMetadata creates metadata for the turn response
func (c *Client) buildResponseMetadata(thinking, modelId string) map[string]interface{} {
	metadata := map[string]interface{}{
		"modelId":     modelId,
		"processedAt": schema.FormatISO8601(),
	}

	if thinking != "" {
		metadata["thinking"] = thinking
		metadata["hasThinking"] = true
		metadata["thinkingLength"] = len(thinking)
	} else {
		metadata["hasThinking"] = false
	}

	return metadata
}

// handleBedrockError processes Bedrock API errors
func (c *Client) handleBedrockError(err error) error {
	c.logger.Error("Bedrock API error", map[string]interface{}{
		"error": err.Error(),
	})
	
	// Determine if error is retryable - simple implementation
	retryable := false
	
	return wferrors.NewBedrockError("Bedrock API error", "BEDROCK_API_ERROR", retryable).
		WithContext("error", err.Error())
}