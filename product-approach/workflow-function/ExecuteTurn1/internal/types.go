package internal

import (
	"time"
)

// VerificationContext represents the verification context from the input
type VerificationContext struct {
	VerificationID       string            `json:"verificationId"`
	VerificationAt       string            `json:"verificationAt"`
	Status               string            `json:"status"`
	VerificationType     string            `json:"verificationType"`
	VendingMachineID     string            `json:"vendingMachineId,omitempty"`
	LayoutID             *int              `json:"layoutId,omitempty"`
	LayoutPrefix         *string           `json:"layoutPrefix,omitempty"`
	ReferenceImageURL    string            `json:"referenceImageUrl"`
	CheckingImageURL     string            `json:"checkingImageUrl"`
	TurnConfig           TurnConfig        `json:"turnConfig"`
	TurnTimestamps       TurnTimestamps    `json:"turnTimestamps"`
	RequestMetadata      RequestMetadata   `json:"requestMetadata"`
	NotificationEnabled  bool              `json:"notificationEnabled"`
	ResourceValidation   ResourceValidation `json:"resourceValidation"`
}

// TurnConfig represents the turn configuration
type TurnConfig struct {
	MaxTurns           int `json:"maxTurns"`
	ReferenceImageTurn int `json:"referenceImageTurn"`
	CheckingImageTurn  int `json:"checkingImageTurn"`
}

// TurnTimestamps represents timestamps for each turn
type TurnTimestamps struct {
	Initialized time.Time  `json:"initialized"`
	Turn1       *time.Time `json:"turn1"`
	Turn2       *time.Time `json:"turn2"`
	Completed   *time.Time `json:"completed"`
}

// RequestMetadata represents request metadata
type RequestMetadata struct {
	RequestID         string    `json:"requestId"`
	RequestTimestamp  time.Time `json:"requestTimestamp"`
	ProcessingStarted time.Time `json:"processingStarted"`
}

// ResourceValidation represents resource validation status
type ResourceValidation struct {
	LayoutExists           bool      `json:"layoutExists"`
	ReferenceImageExists   bool      `json:"referenceImageExists"`
	CheckingImageExists    bool      `json:"checkingImageExists"`
	ValidationTimestamp    time.Time `json:"validationTimestamp"`
}

// LayoutMetadata represents layout metadata from DynamoDB
type LayoutMetadata struct {
	LayoutID       int                `json:"layoutId"`
	LayoutPrefix   string             `json:"layoutPrefix"`
	MachineStructure MachineStructure `json:"machineStructure"`
	Location       string             `json:"location"`
}

// MachineStructure represents the vending machine structure
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// ImageMetadata represents image metadata from S3
type ImageMetadata struct {
	ContentType   string    `json:"contentType"`
	Size          int64     `json:"size"`
	LastModified  time.Time `json:"lastModified"`
	ETag          string    `json:"etag"`
	BucketOwner   string    `json:"bucketOwner"`
}

// BedrockConfig represents Bedrock configuration
type BedrockConfig struct {
	AnthropicVersion string             `json:"anthropic_version"`
	MaxTokens       int                `json:"max_tokens"`
	Thinking        BedrockThinking    `json:"thinking"`
}

// BedrockThinking represents thinking configuration
type BedrockThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// CurrentPrompt represents the prompt structure for Turn 1
type CurrentPrompt struct {
	Messages       []BedrockMessage `json:"messages"`
	TurnNumber     int              `json:"turnNumber"`
	PromptID       string           `json:"promptId"`
	CreatedAt      time.Time        `json:"createdAt"`
	PromptVersion  string           `json:"promptVersion"`
	ImageIncluded  string           `json:"imageIncluded"`
}

// BedrockMessage represents a message in the Bedrock API format
type BedrockMessage struct {
	Role    string          `json:"role"`
	Content []MessageContent `json:"content"`
}

// MessageContent represents content within a message
type MessageContent struct {
	Type  string     `json:"type"`
	Text  *string    `json:"text,omitempty"`
	Image *ImageData `json:"image,omitempty"`
}

// ImageData represents image data in Bedrock format
type ImageData struct {
	Format string     `json:"format"`
	Source ImageSource `json:"source"`
}

// ImageSource represents the source of an image
type ImageSource struct {
	S3Location S3Location `json:"s3Location"`
}

// S3Location represents S3 location details
type S3Location struct {
	URI         string `json:"uri"`
	BucketOwner string `json:"bucketOwner"`
}

// BedrockRequest represents the request to Bedrock API
type BedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens       int              `json:"max_tokens"`
	Thinking        BedrockThinking  `json:"thinking"`
	Messages        []BedrockMessage `json:"messages"`
}

// BedrockResponse represents the response from Bedrock API
type BedrockResponse struct {
	Content       []BedrockContent `json:"content"`
	ID            string           `json:"id"`
	Model         string           `json:"model"`
	Role          string           `json:"role"`
	StopReason    string           `json:"stop_reason"`
	Type          string           `json:"type"`
	Usage         TokenUsage       `json:"usage"`
	Thinking      *string          `json:"thinking,omitempty"`
}

// BedrockContent represents content in Bedrock response
type BedrockContent struct {
	Type string  `json:"type"`
	Text *string `json:"text,omitempty"`
}

// TokenUsage represents token usage metrics
type TokenUsage struct {
	Input     int `json:"input_tokens"`
	Output    int `json:"output_tokens"`
	Thinking  int `json:"thinking_tokens,omitempty"`
	Total     int `json:"total_tokens"`
}

// Turn1Response represents the response from Turn 1
type Turn1Response struct {
	TurnID          int              `json:"turnId"`
	Timestamp       time.Time        `json:"timestamp"`
	Prompt          string           `json:"prompt"`
	Response        BedrockTextResponse `json:"response"`
	LatencyMs       int64            `json:"latencyMs"`
	TokenUsage      TokenUsage       `json:"tokenUsage"`
	AnalysisStage   string           `json:"analysisStage"`
	BedrockMetadata BedrockMetadata  `json:"bedrockMetadata"`
}

// BedrockTextResponse represents the text response portion
type BedrockTextResponse struct {
	Content    string  `json:"content"`
	Thinking   *string `json:"thinking,omitempty"`
	StopReason string  `json:"stop_reason"`
}

// BedrockMetadata represents metadata about the Bedrock request
type BedrockMetadata struct {
	ModelID        string `json:"modelId"`
	RequestID      string `json:"requestId"`
	InvokeLatencyMs int64 `json:"invokeLatencyMs"`
}

// ConversationState represents the conversation state
type ConversationState struct {
	CurrentTurn int        `json:"currentTurn"`
	MaxTurns    int        `json:"maxTurns"`
	History     []TurnHistory `json:"history"`
}

// TurnHistory represents a single turn in conversation history
type TurnHistory struct {
	TurnID        int        `json:"turnId"`
	Timestamp     time.Time  `json:"timestamp"`
	Prompt        string     `json:"prompt"`
	Response      string     `json:"response"`
	LatencyMs     int64      `json:"latencyMs"`
	TokenUsage    TokenUsage `json:"tokenUsage"`
	AnalysisStage string     `json:"analysisStage"`
}

// ExecuteTurn1Input represents the input to ExecuteTurn1 function
type ExecuteTurn1Input struct {
	VerificationContext VerificationContext `json:"verificationContext"`
	CurrentPrompt       CurrentPromptWrapper `json:"currentPrompt"`
	BedrockConfig       BedrockConfig       `json:"bedrockConfig"`
	Images              *Images             `json:"images,omitempty"`
	LayoutMetadata      *LayoutMetadata     `json:"layoutMetadata,omitempty"`
	SystemPrompt        *SystemPromptWrapper `json:"systemPrompt,omitempty"`
	HistoricalContext   map[string]interface{} `json:"historicalContext,omitempty"`
	ConversationState   map[string]interface{} `json:"conversationState,omitempty"`
}

// CurrentPromptWrapper wraps the nested currentPrompt structure
type CurrentPromptWrapper struct {
	CurrentPrompt CurrentPrompt `json:"currentPrompt"`
	Messages      []BedrockMessage `json:"messages,omitempty"`
	TurnNumber    int              `json:"turnNumber,omitempty"`
	PromptID      string           `json:"promptId,omitempty"`
	CreatedAt     time.Time        `json:"createdAt,omitempty"`
	PromptVersion string           `json:"promptVersion,omitempty"`
	ImageIncluded string           `json:"imageIncluded,omitempty"`
}

// SystemPromptWrapper wraps the nested systemPrompt structure
type SystemPromptWrapper struct {
	SystemPrompt SystemPrompt `json:"systemPrompt"`
}

// SystemPrompt represents the system prompt data
type SystemPrompt struct {
	Content      string `json:"content"`
	PromptID     string `json:"promptId"`
	CreatedAt    string `json:"createdAt"`
	PromptVersion string `json:"promptVersion"`
}

// Images represents the image metadata structure
type Images struct {
	ReferenceImageMeta ImageMetadata `json:"referenceImageMeta"`
	CheckingImageMeta  ImageMetadata `json:"checkingImageMeta"`
	BucketOwner string `json:"bucketOwner,omitempty"`
}

// ExecuteTurn1Output represents the output from ExecuteTurn1 function
type ExecuteTurn1Output struct {
	VerificationContext VerificationContext `json:"verificationContext"`
	Turn1Response       Turn1Response       `json:"turn1Response"`
	ConversationState   ConversationState   `json:"conversationState"`
}

// HistoricalContext represents historical verification context for UC2
type HistoricalContext struct {
	PreviousVerificationID string              `json:"previousVerificationId"`
	PreviousVerificationAt string              `json:"previousVerificationAt"`
	PreviousStatus         string              `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64         `json:"hoursSinceLastVerification"`
	MachineStructure       MachineStructure    `json:"machineStructure"`
	CheckingStatus         map[string]string   `json:"checkingStatus"`
	VerificationSummary    map[string]interface{} `json:"verificationSummary"`
}

// Environment variables structure
type Environment struct {
	BedrockModelID        string
	BedrockRegion        string
	AnthropicVersion     string
	MaxTokens            int
	BudgetTokens         int
	ThinkingType         string
	DynamoDBTable        string
	RetryMaxAttempts     int
	RetryBaseDelay       int
}