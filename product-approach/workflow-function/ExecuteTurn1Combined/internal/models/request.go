// internal/models/request.go
package models

// Turn1Request is the input payload for ExecuteTurn1Combined.
type Turn1Request struct {
	VerificationID      string              `json:"verificationId"`
	VerificationContext VerificationContext `json:"verificationContext"`
	S3Refs              Turn1RequestS3Refs  `json:"s3Refs"`
}

// Turn1RequestS3Refs groups the S3 references needed for Turn-1.
type Turn1RequestS3Refs struct {
	Prompts    PromptRefs           `json:"prompts"`
	Images     ImageRefs            `json:"images"`
	Processing ProcessingReferences `json:"processing,omitempty"`
}

// PromptRefs holds S3 locations for prompt artifacts.
type PromptRefs struct {
	System S3Reference `json:"system"` // system-prompt.json
}

// ImageRefs holds S3 locations for image artifacts.
type ImageRefs struct {
	ReferenceBase64 S3Reference `json:"referenceBase64"` // reference-base64.json
}

// ProcessingReferences holds S3 locations for processing artifacts.
type ProcessingReferences struct {
	HistoricalContext S3Reference `json:"historicalContext,omitempty"`
	LayoutMetadata    S3Reference `json:"layoutMetadata,omitempty"`
}

// Turn1ResponseS3Refs groups the S3 references created by Turn-1.
type Turn1ResponseS3Refs struct {
	RawResponse       S3Reference `json:"rawResponse"`
	ProcessedResponse S3Reference `json:"processedResponse"`
}

// Summary contains metrics and identifiers for the Turn-1 execution.
type Summary struct {
	AnalysisStage    ExecutionStage `json:"analysisStage"`
	ProcessingTimeMs int64          `json:"processingTimeMs"`
	TokenUsage       TokenUsage     `json:"tokenUsage"`
	BedrockRequestID string         `json:"bedrockRequestId"`
}

// Turn1Response is the output payload from ExecuteTurn1Combined.
type Turn1Response struct {
	S3Refs  Turn1ResponseS3Refs `json:"s3Refs"`
	Status  VerificationStatus  `json:"status"`
	Summary Summary             `json:"summary"`
}
