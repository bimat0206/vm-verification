package bedrock

import (
	"strings"
)

// API type constants
const (
	APITypeConverse = "Converse"
)

// Analysis stage identifiers
const (
	AnalysisStageTurn1 = "REFERENCE_ANALYSIS"
	AnalysisStageTurn2 = "CHECKING_ANALYSIS"
)

// Expected turn numbers
const (
	ExpectedTurn1Number = 1
	ExpectedTurn2Number = 2
)

// ConverseRequest represents a request to the Bedrock Converse API
type ConverseRequest struct {
	ModelId         string            `json:"modelId"`
	Messages        []MessageWrapper  `json:"messages"`
	System          string            `json:"system,omitempty"`
	InferenceConfig InferenceConfig   `json:"inferenceConfig,omitempty"`
	GuardrailConfig *GuardrailConfig  `json:"guardrailConfig,omitempty"`
	CacheControl    map[string]string `json:"cache_control,omitempty"`
	Reasoning       string            `json:"reasoning,omitempty"` // Added for Claude 3.5 Sonnet thinking support
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
// Only supports bytes (base64 encoded data)
type ImageSource struct {
	Type  string `json:"type"`  // Must be "bytes"
	Bytes string `json:"bytes"` // Base64 encoded image data
}

// InferenceConfig represents configuration for the Bedrock Converse API
type InferenceConfig struct {
	MaxTokens     int      `json:"maxTokens"`
	Temperature   *float64 `json:"temperature,omitempty"`
	TopP          *float64 `json:"topP,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
	Reasoning     string   `json:"reasoning,omitempty"` // Added for Claude 3.5 Sonnet thinking support
}

// GuardrailConfig represents configuration for guardrails in Bedrock
type GuardrailConfig struct {
	GuardrailIdentifier string `json:"guardrailIdentifier"`
	GuardrailVersion    string `json:"guardrailVersion,omitempty"`
}

// ConverseResponse represents a response from the Bedrock Converse API
type ConverseResponse struct {
	RequestID  string           `json:"requestId"`
	ModelID    string           `json:"modelId"`
	StopReason string           `json:"stopReason,omitempty"`
	Content    []ContentBlock   `json:"content"`
	Usage      *TokenUsage      `json:"usage,omitempty"`
	Metrics    *ResponseMetrics `json:"metrics,omitempty"`
}

// TokenUsage represents token usage metrics from Bedrock
type TokenUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// ResponseMetrics represents metrics about the response from Bedrock
type ResponseMetrics struct {
	LatencyMs int64 `json:"latencyMs"`
}

// TextResponse represents the text response portion of a model's output
type TextResponse struct {
	Content    string `json:"content"`
	StopReason string `json:"stop_reason,omitempty"`
}

// BedrockMetadata represents metadata about the Bedrock request
type BedrockMetadata struct {
	ModelID         string `json:"modelId"`
	RequestID       string `json:"requestId"`
	InvokeLatencyMs int64  `json:"invokeLatencyMs"`
	APIType         string `json:"apiType"` // Always "Converse"
}

// Turn1Response represents the response from Turn 1 of a multi-turn conversation
type Turn1Response struct {
	TurnID          int             `json:"turnId"`
	Timestamp       string          `json:"timestamp"`
	Prompt          string          `json:"prompt"`
	Response        TextResponse    `json:"response"`
	LatencyMs       int64           `json:"latencyMs"`
	TokenUsage      TokenUsage      `json:"tokenUsage"`
	AnalysisStage   string          `json:"analysisStage"`
	BedrockMetadata BedrockMetadata `json:"bedrockMetadata"`
	APIType         string          `json:"apiType,omitempty"` // Always "Converse"
}

// Turn2Response represents the response from Turn 2 of a multi-turn conversation
type Turn2Response struct {
	TurnID          int             `json:"turnId"`
	Timestamp       string          `json:"timestamp"`
	Prompt          string          `json:"prompt"`
	Response        TextResponse    `json:"response"`
	LatencyMs       int64           `json:"latencyMs"`
	TokenUsage      TokenUsage      `json:"tokenUsage"`
	AnalysisStage   string          `json:"analysisStage"`
	BedrockMetadata BedrockMetadata `json:"bedrockMetadata"`
	APIType         string          `json:"apiType,omitempty"` // Always "Converse"
	PreviousTurn    *Turn1Response  `json:"previousTurn,omitempty"`
}

// VerificationResult represents the result of a verification operation
type VerificationResult struct {
	VerificationId    string      `json:"verificationId"`
	VerificationAt    string      `json:"verificationAt"`
	Status            string      `json:"status"`
	ReferenceImageUrl string      `json:"referenceImageUrl"`
	CheckingImageUrl  string      `json:"checkingImageUrl"`
	DiscrepancyCount  int         `json:"discrepancyCount"`
	ResultImageUrl    string      `json:"resultImageUrl,omitempty"`
	Metrics           interface{} `json:"metrics,omitempty"`
}

// VerificationStatus represents the status of a verification operation
type VerificationStatus struct {
	Status           string `json:"status"` // PROCESSING, COMPLETED, FAILED
	VerificationId   string `json:"verificationId"`
	VerificationAt   string `json:"verificationAt"`
	Message          string `json:"message,omitempty"`
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	CompletedAt      string `json:"completedAt,omitempty"`
	ProcessingTimeMs int64  `json:"processingTimeMs,omitempty"`
}

// DiscrepancyItem represents a discrepancy between reference and checking images
type DiscrepancyItem struct {
	Position           string  `json:"position"`
	ExpectedProduct    string  `json:"expected"`
	FoundProduct       string  `json:"found"`
	Issue              string  `json:"issue"`
	Confidence         float64 `json:"confidence"`
	Evidence           string  `json:"evidence,omitempty"`
	VerificationResult string  `json:"verificationResult"`
}

// VerificationSummary provides overall metrics for a verification
type VerificationSummary struct {
	TotalPositionsChecked int     `json:"totalPositionsChecked"`
	CorrectPositions      int     `json:"correctPositions"`
	DiscrepantPositions   int     `json:"discrepantPositions"`
	MissingProducts       int     `json:"missingProducts"`
	IncorrectProductTypes int     `json:"incorrectProductTypes"`
	UnexpectedProducts    int     `json:"unexpectedProducts"`
	EmptyPositionsCount   int     `json:"emptyPositionsCount"`
	OverallAccuracy       float64 `json:"overallAccuracy"`
	OverallConfidence     float64 `json:"overallConfidence"`
	VerificationStatus    string  `json:"verificationStatus"`  // CORRECT, INCORRECT
	VerificationOutcome   string  `json:"verificationOutcome"` // Human-readable summary
}

// ProcessedResults represents the processed verification results
type ProcessedResults struct {
	VerificationId      string              `json:"verificationId"`
	VerificationAt      string              `json:"verificationAt"`
	Status              string              `json:"status"`
	ReferenceImageUrl   string              `json:"referenceImageUrl"`
	CheckingImageUrl    string              `json:"checkingImageUrl"`
	Discrepancies       []DiscrepancyItem   `json:"discrepancies"`
	VerificationSummary VerificationSummary `json:"verificationSummary"`
	Turn1Response       *Turn1Response      `json:"turn1Response,omitempty"`
	Turn2Response       *Turn2Response      `json:"turn2Response,omitempty"`
	ProcessingTimeMs    int64               `json:"processingTimeMs"`
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
