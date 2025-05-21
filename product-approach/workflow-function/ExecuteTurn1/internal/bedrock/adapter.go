package bedrock

import (
	"context"
	
	sharedBedrock "workflow-function/shared/bedrock"
)

// BedrockAdapter adapts the shared Bedrock client to our internal interface
type BedrockAdapter struct {
	client *sharedBedrock.BedrockClient
}

// NewBedrockAdapter creates a new adapter for the shared Bedrock client
func NewBedrockAdapter(client *sharedBedrock.BedrockClient) *BedrockClient {
	adapter := &BedrockAdapter{
		client: client,
	}
	
	return &BedrockClient{
		Converse:            adapter.Converse,
		GetTextFromResponse: adapter.GetTextFromResponse,
	}
}

// Converse adapts the shared Bedrock client Converse method to our internal types
func (a *BedrockAdapter) Converse(ctx context.Context, req *ConverseRequest) (*ConverseResponse, int64, error) {
	// Adapt our request to shared request
	bedrockMessages := make([]sharedBedrock.MessageWrapper, 0, len(req.Messages))
	
	for _, msg := range req.Messages {
		// Convert content blocks
		contentBlocks := make([]sharedBedrock.ContentBlock, 0, len(msg.Content))
		
		for _, block := range msg.Content {
			if block.Type == "text" {
				contentBlocks = append(contentBlocks, sharedBedrock.ContentBlock{
					Type: "text",
					Text: block.Text,
				})
			} else if block.Type == "image" && block.Image != nil {
				contentBlocks = append(contentBlocks, sharedBedrock.ContentBlock{
					Type: "image",
					Image: &sharedBedrock.ImageBlock{
						Format: block.Image.Format,
						Source: sharedBedrock.ImageSource{
							Type: "bytes",
							Bytes: block.Image.Source.Bytes,
						},
					},
				})
			}
		}
		
		bedrockMessages = append(bedrockMessages, sharedBedrock.MessageWrapper{
			Role:    msg.Role,
			Content: contentBlocks,
		})
	}
	
	// Create shared inference config
	inferenceConfig := sharedBedrock.InferenceConfig{
		MaxTokens:    req.InferenceConfig.MaxTokens,
		Temperature:  req.InferenceConfig.Temperature,
		TopP:         req.InferenceConfig.TopP,
		StopSequences: req.InferenceConfig.StopSequences,
	}
	
	// Create shared request
	sharedReq := &sharedBedrock.ConverseRequest{
		System:          req.System,
		Messages:        bedrockMessages,
		InferenceConfig: inferenceConfig,
		ModelId:         req.ModelID, // Fix: Correct field name is "ModelId" (lowercase "d")
	}
	
	// Call shared client
	resp, latencyMs, err := a.client.Converse(ctx, sharedReq)
	if err != nil {
		return nil, 0, err
	}
	
	// Extract text from response content
	content := ""
	if len(resp.Content) > 0 {
		for _, block := range resp.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}
	}

	// Convert response
	return &ConverseResponse{
		Content:    content,
		StopReason: resp.StopReason,
		ModelID:    resp.ModelID,
		Usage: &TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, latencyMs, nil
}

// GetTextFromResponse adapts the shared Bedrock client's text extraction method
func (a *BedrockAdapter) GetTextFromResponse(resp *ConverseResponse) string {
	// The shared client might have more complex extraction logic,
	// but we just return content directly in this simple adapter
	return resp.Content
}