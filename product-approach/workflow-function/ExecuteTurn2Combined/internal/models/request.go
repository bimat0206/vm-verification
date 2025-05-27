package models

// VerificationContext carries metadata to drive prompt generation.
// This is compatible with the schema package but maintains local structure
type VerificationContext struct {
	// Timestamp when the verification was first initialized
	VerificationAt string `json:"verificationAt,omitempty"`
	// Type of verification: e.g. "LAYOUT_VS_CHECKING" or "PREVIOUS_VS_CURRENT"
	VerificationType string `json:"verificationType"`
	// Arbitrary planogram/layout details for LAYOUT_VS_CHECKING
	LayoutMetadata map[string]interface{} `json:"layoutMetadata,omitempty"`
	// Historical context (e.g. prior image analysis) for PREVIOUS_VS_CURRENT
	HistoricalContext map[string]interface{} `json:"historicalContext,omitempty"`
	// Additional fields for schema compatibility
	VendingMachineId string `json:"vendingMachineId,omitempty"`
	LayoutId         int    `json:"layoutId,omitempty"`
	LayoutPrefix     string `json:"layoutPrefix,omitempty"`
}

// PromptRefs holds S3 locations for prompt artifacts.
type PromptRefs struct {
	System S3Reference `json:"system"` // system-prompt.json
}

// Turn2Request represents the input payload for ExecuteTurn2Combined
// It references artifacts produced by Turn1
type Turn2Request struct {
	VerificationID      string              `json:"verificationId"`
	VerificationContext VerificationContext `json:"verificationContext"`
	S3Refs              Turn2RequestS3Refs  `json:"s3Refs"`
	InputInitializationFileRef S3Reference  `json:"-"`
}

type Turn2RequestS3Refs struct {
	Prompts    PromptRefs           `json:"prompts"`
	Images     Turn2ImageRefs       `json:"images"`
	Processing Turn2ProcessingRefs  `json:"processing"`
}

type Turn2ImageRefs struct {
	CheckingBase64 S3Reference `json:"checkingBase64"`
}

type Turn2ProcessingRefs struct {
	Turn1Markdown S3Reference `json:"turn1Markdown"`
}

// TokenUsage represents token usage metrics
type TokenUsage struct {
	InputTokens    int `json:"inputTokens"`
	OutputTokens   int `json:"outputTokens"`
	ThinkingTokens int `json:"thinkingTokens,omitempty"`
	TotalTokens    int `json:"totalTokens"`
}

// StepFunctionResponse is a simplified output for Step Functions
// Provided here for main wrapper
type StepFunctionResponse struct {
	VerificationID string                  `json:"verificationId"`
	Status         string                  `json:"status"`
	S3References   map[string]S3Reference  `json:"s3References"`
}
