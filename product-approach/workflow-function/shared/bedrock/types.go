package bedrock

import (
	"strings"
)

// API type constants
const (
	APITypeConverse = "Converse"
	APITypeLegacy   = "InvokeModel"
)

// ConverseRequest represents a request to the Bedrock Converse API
type ConverseRequest struct {
	ModelId         string                `json:"modelId"`
	Messages        []MessageWrapper      `json:"messages"`
	System          string                `json:"system,omitempty"`
	InferenceConfig InferenceConfig       `json:"inferenceConfig,omitempty"`
	GuardrailConfig *GuardrailConfig      `json:"guardrailConfig,omitempty"`
}

// MessageWrapper represents a message in the Converse API
type MessageWrapper struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type  string      `json:"type"`
	Text  string      `json:"text,omitempty"`
	Image *ImageBlock `json:"image,omitempty"`
}

// ImageBlock represents an image in a content block
type ImageBlock struct {
	Format string      `json:"format"`
	Source ImageSource `json:"source"`
}

// ImageSource represents the source of an image
type ImageSource struct {
	Type       string     `json:"type"`
	S3Location S3Location `json:"s3Location"`
}

// S3Location represents an S3 location
type S3Location struct {
	URI         string `json:"uri"`
	BucketOwner string `json:"bucketOwner,omitempty"`
}

// InferenceConfig represents configuration for inference
type InferenceConfig struct {
	MaxTokens     int       `json:"maxTokens"`
	Temperature   *float64  `json:"temperature,omitempty"`
	TopP          *float64  `json:"topP,omitempty"`
	StopSequences []string  `json:"stopSequences,omitempty"`
}

// GuardrailConfig represents configuration for guardrails
type GuardrailConfig struct {
	GuardrailIdentifier string `json:"guardrailIdentifier"`
	GuardrailVersion    string `json:"guardrailVersion,omitempty"`
}

// ConverseResponse represents a response from the Bedrock Converse API
type ConverseResponse struct {
	RequestID  string         `json:"requestId"`
	ModelID    string         `json:"modelId"`
	StopReason string         `json:"stopReason,omitempty"`
	Content    []ContentBlock `json:"content"`
	Usage      *TokenUsage    `json:"usage,omitempty"`
	Metrics    *ResponseMetrics `json:"metrics,omitempty"`
}

// TokenUsage represents token usage metrics
type TokenUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// ResponseMetrics represents metrics about the response
type ResponseMetrics struct {
	LatencyMs int64 `json:"latencyMs"`
}

// Turn1Response represents the response from Turn 1
type Turn1Response struct {
	TurnID          int              `json:"turnId"`
	Timestamp       string           `json:"timestamp"`
	Prompt          string           `json:"prompt"`
	Response        TextResponse     `json:"response"`
	LatencyMs       int64            `json:"latencyMs"`
	TokenUsage      TokenUsage       `json:"tokenUsage"`
	AnalysisStage   string           `json:"analysisStage"`
	BedrockMetadata BedrockMetadata  `json:"bedrockMetadata"`
	APIType         string           `json:"apiType"`
}

// TextResponse represents the text response portion
type TextResponse struct {
	Content    string  `json:"content"`
	StopReason string  `json:"stop_reason"`
}

// BedrockMetadata represents metadata about the Bedrock request
type BedrockMetadata struct {
	ModelID        string `json:"modelId"`
	RequestID      string `json:"requestId"`
	InvokeLatencyMs int64 `json:"invokeLatencyMs"`
	APIType        string `json:"apiType"`
}

// ExtractTextFromResponse extracts text content from a Converse response
func ExtractTextFromResponse(response *ConverseResponse) string {
	if response == nil || len(response.Content) == 0 {
		return ""
	}

	var textParts []string
	for _, content := range response.Content {
		if content.Type == "text" {
			textParts = append(textParts, content.Text)
		}
	}

	return strings.Join(textParts, "")
}
