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
		TurnID:        1,
		Timestamp:     timestamp.Format(time.RFC3339),
		Prompt:        promptText,
		Response: TextResponse{
			Content: responseText,
			StopReason: response.StopReason,
		},
		LatencyMs:     latencyMs,
		TokenUsage:    tokenUsage,
		AnalysisStage: "TURN1",
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
// Either bytes or s3Location should be provided, but not both
func CreateImageContentBlock(format string, bytes string, s3URI string, bucketOwner string) ContentBlock {
	// Create image block
	imageBlock := &ImageBlock{
		Format: format,
		Source: ImageSource{},
	}

	// Set source based on inputs
	if bytes != "" {
		imageBlock.Source.Type = "bytes"
		imageBlock.Source.Bytes = bytes
	} else if s3URI != "" {
		imageBlock.Source.Type = "s3Location"
		imageBlock.Source.S3Location = S3Location{
			URI:         s3URI,
			BucketOwner: bucketOwner,
		}
	}

	return ContentBlock{
		Type:  "image",
		Image: imageBlock,
	}
}

// CreateImageContentFromBytes creates an image content block directly from base64 encoded data
func CreateImageContentFromBytes(format string, base64Data string) ContentBlock {
	return CreateImageContentBlock(format, base64Data, "", "")
}

// CreateImageContentFromS3 creates an image content block from an S3 URI
func CreateImageContentFromS3(format string, s3URI string, bucketOwner string) ContentBlock {
	return CreateImageContentBlock(format, "", s3URI, bucketOwner)
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
