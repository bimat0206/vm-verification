package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// WorkflowState represents the current state of the verification workflow
type WorkflowState struct {
	VerificationID    string          `json:"verificationId"`
	Status            string          `json:"status"`
	Stage             string          `json:"stage"`
	CorrelationID     string          `json:"correlationId"`
	Timestamp         string          `json:"timestamp"`
	CurrentPrompt     PromptDetails   `json:"currentPrompt"`
	ImageData         ImageDetails    `json:"imageData"`
	TurnHistory       []TurnResponse  `json:"turnHistory"`
	Settings          Settings        `json:"settings"`
	VerificationKeys  []string        `json:"verificationKeys"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// PromptDetails contains information about the prompt to be sent
type PromptDetails struct {
	PromptID     string         `json:"promptId"`
	PromptText   string         `json:"promptText,omitempty"`
	Messages     []BedrockMessage `json:"messages,omitempty"`
	ThinkingType string         `json:"thinkingType,omitempty"`
}

// BedrockMessage represents a message in the Bedrock format
type BedrockMessage struct {
	Role    string `json:"role"`
	Content []MessageContent `json:"content"`
}

// MessageContent represents content in a Bedrock message
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

// ImageSource represents an image source in a Bedrock message
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

// ImageDetails contains information about the image to be analyzed
type ImageDetails struct {
	ImageID    string `json:"imageId"`
	ImageURL   string `json:"imageUrl,omitempty"`
	ImageBase64 string `json:"imageBase64,omitempty"`
	ImageS3Key string `json:"imageS3Key,omitempty"`
	ImageS3Bucket string `json:"imageS3Bucket,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// Settings represents configuration settings for the workflow
type Settings struct {
	ModelID           string  `json:"modelId"`
	Temperature       float64 `json:"temperature"`
	MaxTokens         int     `json:"maxTokens"`
	EnableThinking    bool    `json:"enableThinking"`
	ThinkingBudget    int     `json:"thinkingBudget"`
	HybridBase64Enabled bool  `json:"hybridBase64Enabled"`
}

// BedrockRequest represents a request to the Bedrock Converse API
type BedrockRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	Messages         []BedrockMessage `json:"messages"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature,omitempty"`
	System           string          `json:"system,omitempty"`
}

// TurnResponse represents a response from Bedrock for a specific turn
type TurnResponse struct {
	TurnID      int             `json:"turnId"`
	Timestamp   string          `json:"timestamp"`
	Response    BedrockResponse `json:"response"`
	LatencyMs   int64           `json:"latencyMs"`
	TokenUsage  TokenUsage      `json:"tokenUsage"`
	Stage       string          `json:"stage"`
	Thinking    string          `json:"thinking,omitempty"`
}

// BedrockResponse represents a response from the Bedrock Converse API
type BedrockResponse struct {
	ID           string                `json:"id"`
	Content      []MessageContent      `json:"content"`
	Role         string                `json:"role"`
	StopReason   string                `json:"stop_reason,omitempty"`
	StopSequence string                `json:"stop_sequence,omitempty"`
	Usage        TokenUsage            `json:"usage,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// FormatISO8601 formats current time as ISO8601
func FormatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// HybridBase64Retriever is an interface for retrieving Base64 data
type HybridBase64Retriever interface {
	RetrieveBase64Data(imageInfo ImageDetails) (string, error)
}

// Error represents a structured error response
type Error struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Retryable bool   `json:"retryable"`
	Context  map[string]string `json:"context,omitempty"`
}

// NewError creates a new structured error
func NewError(code, message string, retryable bool) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Context:   make(map[string]string),
	}
}

// WithContext adds context to an error
func (e *Error) WithContext(key, value string) *Error {
	e.Context[key] = value
	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	contextStr := ""
	for k, v := range e.Context {
		contextStr += fmt.Sprintf(" %s=%s", k, v)
	}
	return fmt.Sprintf("%s: %s%s", e.Code, e.Message, contextStr)
}

// ExecuteTurn1Request represents the input to the ExecuteTurn1 Lambda function
type ExecuteTurn1Request struct {
	WorkflowState WorkflowState `json:"workflowState"`
}

// ExecuteTurn1Response represents the output of the ExecuteTurn1 Lambda function
type ExecuteTurn1Response struct {
	WorkflowState WorkflowState `json:"workflowState"`
	Error         *Error        `json:"error,omitempty"`
}