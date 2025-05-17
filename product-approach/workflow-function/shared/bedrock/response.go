package bedrock

import (
	"fmt"
	"time"
)

// ResponseProcessor processes responses from Bedrock
type ResponseProcessor struct{}

// NewResponseProcessor creates a new response processor
func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{}
}

// ProcessTurn1Response processes a response for Turn 1
func (rp *ResponseProcessor) ProcessTurn1Response(
	response *ConverseResponse,
	promptText string,
	latencyMs int64,
	timestamp time.Time,
) (*Turn1Response, error) {
	// Extract text from response
	responseText := ExtractTextFromResponse(response)
	if responseText == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	// Create token usage
	tokenUsage := TokenUsage{
		InputTokens:  0,
		OutputTokens: 0,
		TotalTokens:  0,
	}

	// Copy token usage if available
	if response.Usage != nil {
		tokenUsage.InputTokens = response.Usage.InputTokens
		tokenUsage.OutputTokens = response.Usage.OutputTokens
		tokenUsage.TotalTokens = response.Usage.TotalTokens
	}

	// Create Turn1Response
	turn1Response := &Turn1Response{
		TurnID:        ExpectedTurn1Number,
		Timestamp:     timestamp.Format(time.RFC3339),
		Prompt:        promptText,
		Response: TextResponse{
			Content: responseText,
			StopReason: response.StopReason,
		},
		LatencyMs:     latencyMs,
		TokenUsage:    tokenUsage,
		AnalysisStage: AnalysisStageTurn1,
		BedrockMetadata: BedrockMetadata{
			ModelID:        response.ModelID,
			RequestID:      response.RequestID,
			InvokeLatencyMs: latencyMs,
			APIType:        APITypeConverse,
		},
		APIType:       APITypeConverse,
	}

	return turn1Response, nil
}

// ProcessTurn2Response processes a response for Turn 2
func (rp *ResponseProcessor) ProcessTurn2Response(
	response *ConverseResponse,
	promptText string,
	latencyMs int64,
	timestamp time.Time,
	previousTurn *Turn1Response,
) (*Turn2Response, error) {
	// Extract text from response
	responseText := ExtractTextFromResponse(response)
	if responseText == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	// Create token usage
	tokenUsage := TokenUsage{
		InputTokens:  0,
		OutputTokens: 0,
		TotalTokens:  0,
	}

	// Copy token usage if available
	if response.Usage != nil {
		tokenUsage.InputTokens = response.Usage.InputTokens
		tokenUsage.OutputTokens = response.Usage.OutputTokens
		tokenUsage.TotalTokens = response.Usage.TotalTokens
	}

	// Create Turn2Response
	turn2Response := &Turn2Response{
		TurnID:        ExpectedTurn2Number,
		Timestamp:     timestamp.Format(time.RFC3339),
		Prompt:        promptText,
		Response: TextResponse{
			Content: responseText,
			StopReason: response.StopReason,
		},
		LatencyMs:     latencyMs,
		TokenUsage:    tokenUsage,
		AnalysisStage: AnalysisStageTurn2,
		BedrockMetadata: BedrockMetadata{
			ModelID:        response.ModelID,
			RequestID:      response.RequestID,
			InvokeLatencyMs: latencyMs,
			APIType:        APITypeConverse,
		},
		APIType:       APITypeConverse,
		PreviousTurn:  previousTurn,
	}

	return turn2Response, nil
}

// CreateConverseRequest creates a request for the Converse API
func CreateConverseRequest(
	modelID string,
	messages []MessageWrapper,
	systemPrompt string,
	maxTokens int,
) *ConverseRequest {
	request := &ConverseRequest{
		ModelId:  modelID,
		Messages: messages,
		System:   systemPrompt,
		InferenceConfig: InferenceConfig{
			MaxTokens: maxTokens,
		},
	}

	return request
}

// CreateImageContentBlock creates an image content block for use in a Converse request
// Only bytes should be provided
func CreateImageContentBlock(format string, bytes string) ContentBlock {
	// Create image block
	imageBlock := &ImageBlock{
		Format: format,
		Source: ImageSource{
			Type: "bytes",
			Bytes: bytes,
		},
	}

	return ContentBlock{
		Type:  "image",
		Image: imageBlock,
	}
}

// CreateImageContentFromBytes creates an image content block directly from base64 encoded data
func CreateImageContentFromBytes(format string, base64Data string) ContentBlock {
	return CreateImageContentBlock(format, base64Data)
}

// CreateUserMessageWithContent creates a user message with mixed content (text and/or images)
func CreateUserMessageWithContent(text string, images []ContentBlock) MessageWrapper {
	// Start with text content if provided
	var content []ContentBlock
	if text != "" {
		content = append(content, ContentBlock{
			Type: "text",
			Text: text,
		})
	}
	
	// Add images if provided
	if len(images) > 0 {
		content = append(content, images...)
	}
	
	return MessageWrapper{
		Role:    "user",
		Content: content,
	}
}

// CreateAssistantMessageWithText creates an assistant message with text content
func CreateAssistantMessageWithText(text string) MessageWrapper {
	return MessageWrapper{
		Role: "assistant",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// CreateTurn2ConversationHistory creates a conversation history for Turn 2 based on Turn 1 response
func CreateTurn2ConversationHistory(turn1Response *Turn1Response) []MessageWrapper {
	if turn1Response == nil {
		return nil
	}

	// Create conversation history with user prompt and assistant response from Turn 1
	messages := []MessageWrapper{
		// User message from Turn 1
		{
			Role: "user",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: turn1Response.Prompt,
				},
			},
		},
		// Assistant response from Turn 1
		{
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: turn1Response.Response.Content,
				},
			},
		},
	}

	return messages
}

// CreateConverseRequestForTurn2 creates a request for Turn 2 based on Turn 1 response and new prompt
func CreateConverseRequestForTurn2(
	modelID string,
	turn1Response *Turn1Response,
	turn2Prompt string,
	systemPrompt string,
	maxTokens int,
) *ConverseRequest {
	// Get conversation history from Turn 1
	messages := CreateTurn2ConversationHistory(turn1Response)
	
	// Add the new user message for Turn 2
	messages = append(messages, MessageWrapper{
		Role: "user",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: turn2Prompt,
			},
		},
	})
	
	// Create the request
	return CreateConverseRequest(modelID, messages, systemPrompt, maxTokens)
}

// CreateConverseRequestForTurn2WithImages creates a request for Turn 2 with images
func CreateConverseRequestForTurn2WithImages(
	modelID string,
	turn1Response *Turn1Response,
	turn2Prompt string,
	images []ContentBlock,
	systemPrompt string,
	maxTokens int,
) *ConverseRequest {
	// Get conversation history from Turn 1
	messages := CreateTurn2ConversationHistory(turn1Response)
	
	// Add the new user message for Turn 2 with images
	messages = append(messages, CreateUserMessageWithContent(turn2Prompt, images))
	
	// Create the request
	return CreateConverseRequest(modelID, messages, systemPrompt, maxTokens)
}