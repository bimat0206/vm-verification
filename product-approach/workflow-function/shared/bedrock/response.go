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
