// Package schema provides Turn1 response types
package schema

// Turn1ProcessedResponse represents the processed response from Turn1 analysis
type Turn1ProcessedResponse struct {
	InitialConfirmation string `json:"initialConfirmation"`
	MachineStructure    string `json:"machineStructure"`
	ReferenceRowStatus  string `json:"referenceRowStatus"`

	// Additional fields that might be present in Turn1 responses
	ProductAnalysis    map[string]interface{} `json:"productAnalysis,omitempty"`
	LayoutValidation   map[string]interface{} `json:"layoutValidation,omitempty"`
	QualityAssessment  map[string]interface{} `json:"qualityAssessment,omitempty"`
	ProcessingMetadata map[string]interface{} `json:"processingMetadata,omitempty"`
}

// BedrockResponse represents a standardized response from Bedrock
type BedrockResponse struct {
	Content          string                 `json:"content"`
	Thinking         string                 `json:"thinking,omitempty"`
	CompletionReason string                 `json:"completionReason"`
	InputTokens      int                    `json:"inputTokens"`
	OutputTokens     int                    `json:"outputTokens"`
	ThinkingTokens   int                    `json:"thinkingTokens,omitempty"`
	TotalTokens      int                    `json:"totalTokens,omitempty"`
	LatencyMs        int64                  `json:"latencyMs"`
	ModelId          string                 `json:"modelId"`
	Timestamp        string                 `json:"timestamp"`
	Turn             int                    `json:"turn,omitempty"`
	ProcessingTimeMs int64                  `json:"processingTimeMs,omitempty"`
	TokenUsage       *TokenUsage            `json:"tokenUsage,omitempty"`
	ModelConfig      *ModelConfig           `json:"modelConfig,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ModelConfig represents Bedrock model configuration
type ModelConfig struct {
	ModelId     string  `json:"modelId"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"topP"`
	MaxTokens   int     `json:"maxTokens"`
}
