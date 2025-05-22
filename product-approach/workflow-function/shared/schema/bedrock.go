// Package schema provides Bedrock-related types and functions
package schema

// BedrockMessage represents a message in Bedrock API format
type BedrockMessage struct {
	Role    string           `json:"role"`    // "user", "assistant"
	Content []BedrockContent `json:"content"`
}

// BedrockContent represents content within a Bedrock message
type BedrockContent struct {
	Type  string            `json:"type"`            // "text" or "image"
	Text  string            `json:"text,omitempty"`  // For text content
	Image *BedrockImageData `json:"image,omitempty"` // For image content
}

// BedrockImageData represents image data for Bedrock API
type BedrockImageData struct {
	Format string             `json:"format"`         // "png", "jpeg", etc.
	Source BedrockImageSource `json:"source"`
}

// BedrockImageSource represents the source of image data for Bedrock
type BedrockImageSource struct {
	Type      string `json:"type"`           // "base64" for v2.0.0 schema
	Media_type string `json:"media_type"`    // "image/jpeg", "image/png", etc.
	Data      string `json:"data"`           // Base64-encoded image data
}

// BedrockConfig contains configuration for Bedrock API calls
type BedrockConfig struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	Thinking         *Thinking `json:"thinking,omitempty"`
}

// Thinking configures the thinking feature
type Thinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// TurnResponse represents response from a Bedrock turn with enhanced structure
type TurnResponse struct {
	TurnId      int                    `json:"turnId"`
	Timestamp   string                 `json:"timestamp"`
	Prompt      string                 `json:"prompt"`
	ImageUrls   map[string]string      `json:"imageUrls,omitempty"`  // S3 URLs for reference/traceability
	Response    BedrockApiResponse     `json:"response"`
	LatencyMs   int64                  `json:"latencyMs"`
	TokenUsage  *TokenUsage            `json:"tokenUsage,omitempty"`
	Stage       string                 `json:"analysisStage"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BedrockApiResponse represents the full response from Bedrock API
type BedrockApiResponse struct {
	Content    string `json:"content"`              // Main response text
	Thinking   string `json:"thinking,omitempty"`   // Internal reasoning (if enabled)
	StopReason string `json:"stop_reason,omitempty"`
	ModelId    string `json:"modelId,omitempty"`
	RequestId  string `json:"requestId,omitempty"`
}

// TokenUsage represents token usage metrics
type TokenUsage struct {
	InputTokens   int `json:"inputTokens"`
	OutputTokens  int `json:"outputTokens"`
	ThinkingTokens int `json:"thinkingTokens,omitempty"`
	TotalTokens   int `json:"totalTokens"`
}

// SystemPrompt contains the system prompt configuration
type SystemPrompt struct {
	Content       string        `json:"content"`
	BedrockConfig *BedrockConfig `json:"-"` // Marked as ignored in JSON as it will be at top level
	PromptId      string        `json:"promptId,omitempty"`
	PromptVersion string        `json:"promptVersion,omitempty"`
}

// CurrentPrompt contains the user prompt for a specific turn with Bedrock message support
type CurrentPrompt struct {
	// Text prompt (backward compatibility)
	Text       string `json:"text,omitempty"`
	TurnNumber int    `json:"turnNumber"`
	
	// Image configuration
	IncludeImage string `json:"includeImage"`  // "reference", "checking", "both", "none"
	
	// Bedrock-formatted messages (with Base64 images)
	Messages []BedrockMessage `json:"messages,omitempty"`
	
	// Metadata
	PromptId      string                 `json:"promptId,omitempty"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	PromptVersion string                 `json:"promptVersion,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BuildBedrockMessage creates a Bedrock message with text and optional image
// For S3-only storage, this requires a retriever to get the Base64 data from S3
// This function is kept for backward compatibility but should be replaced with BedrockMessageBuilder
func BuildBedrockMessage(text string, image *ImageInfo) BedrockMessage {
	content := []BedrockContent{
		{
			Type: "text",
			Text: text,
		},
	}
	
	// Note: In S3-only storage, we can't add the image directly here
	// as we need a retriever to get the Base64 data from S3
	// This function is kept for backward compatibility
	// Use BedrockMessageBuilder.AddImageWithS3Retrieval instead
	
	return BedrockMessage{
		Role:    "user",
		Content: content,
	}
}

// BuildBedrockMessages creates the complete messages array for Bedrock API
// Supports both legacy and new field formats
func BuildBedrockMessages(prompt string, includeImage string, images *ImageData) []BedrockMessage {
	var messages []BedrockMessage
	
	var imageToInclude *ImageInfo
	switch includeImage {
	case "reference":
		if images != nil {
			imageToInclude = images.GetReference() // Uses helper that handles both formats
		}
	case "checking":
		if images != nil {
			imageToInclude = images.GetChecking() // Uses helper that handles both formats
		}
	}
	
	message := BuildBedrockMessage(prompt, imageToInclude)
	messages = append(messages, message)
	
	return messages
}
// Add these structures for combined function response handling

// CombinedTurnResponse represents response from combined turn functions
type CombinedTurnResponse struct {
    *TurnResponse // Embed existing TurnResponse
    
    // Additional fields for combined functions
    ProcessingStages  []ProcessingStage      `json:"processingStages,omitempty"`
    InternalPrompt    string                 `json:"internalPrompt,omitempty"`
    TemplateUsed      string                 `json:"templateUsed,omitempty"`
    ContextEnrichment map[string]interface{} `json:"contextEnrichment,omitempty"`
}

// ProcessingStage represents individual processing stages within combined functions
type ProcessingStage struct {
    StageName    string                 `json:"stageName"`
    StartTime    string                 `json:"startTime"`
    EndTime      string                 `json:"endTime"`
    Duration     int64                  `json:"durationMs"`
    Status       string                 `json:"status"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateContext represents context for template processing
type TemplateContext struct {
    VerificationType  string                 `json:"verificationType"`
    TurnNumber        int                    `json:"turnNumber"`
    MachineStructure  map[string]interface{} `json:"machineStructure,omitempty"`
    LayoutMetadata    map[string]interface{} `json:"layoutMetadata,omitempty"`
    HistoricalContext map[string]interface{} `json:"historicalContext,omitempty"`
}