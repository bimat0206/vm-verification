// internal/models/request.go
package models

// Turn1Request is the input payload for ExecuteTurn1Combined.
type Turn1Request struct {
	VerificationID      string              `json:"verificationId"`
	VerificationContext VerificationContext `json:"verificationContext"`
	S3Refs              Turn1RequestS3Refs  `json:"s3Refs"`
	// InputInitializationFileRef stores the S3 location of the initialization.json
	// that was used to create this request. It is not part of the incoming
	// JSON payload but is populated internally for status updates.
	InputInitializationFileRef S3Reference `json:"-"`
}

// Turn1RequestS3Refs groups the S3 references needed for Turn-1.
type Turn1RequestS3Refs struct {
	Prompts    PromptRefs           `json:"prompts"`
	Images     ImageRefs            `json:"images"`
	Processing ProcessingReferences `json:"processing,omitempty"`
}

// Turn2Request is the input payload for ExecuteTurn2Combined.
type Turn2Request struct {
	VerificationID      string              `json:"verificationId"`
	VerificationContext VerificationContext `json:"verificationContext"`
	S3Refs              Turn2RequestS3Refs  `json:"s3Refs"`
	// InputInitializationFileRef stores the S3 location of the initialization.json
	// that was used to create this request. It is not part of the incoming
	// JSON payload but is populated internally for status updates.
	InputInitializationFileRef S3Reference `json:"-"`
}

// Turn2RequestS3Refs groups the S3 references needed for Turn-2.
type Turn2RequestS3Refs struct {
	Prompts    PromptRefs           `json:"prompts"`
	Images     Turn2ImageRefs       `json:"images"`
	Processing ProcessingReferences `json:"processing,omitempty"`
	Turn1      Turn1References      `json:"turn1"`
}

// Turn1References holds S3 locations for Turn-1 artifacts needed for Turn-2.
type Turn1References struct {
	ProcessedResponse S3Reference `json:"processedResponse"` // turn1-processed-response.json
	RawResponse       S3Reference `json:"rawResponse"`       // turn1-raw-response.json
}

// Turn2ImageRefs holds S3 locations for image artifacts for Turn-2.
type Turn2ImageRefs struct {
	CheckingBase64 S3Reference `json:"checkingBase64"` // checking-base64.json
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

// Turn2ResponseS3Refs groups the S3 references created by Turn-2.
type Turn2ResponseS3Refs struct {
	RawResponse       S3Reference `json:"rawResponse"`
	ProcessedResponse S3Reference `json:"processedResponse"`
}

// Summary contains metrics and identifiers for the execution.
type Summary struct {
	AnalysisStage    ExecutionStage `json:"analysisStage"`
	ProcessingTimeMs int64          `json:"processingTimeMs"`
	TokenUsage       TokenUsage     `json:"tokenUsage"`
	BedrockRequestID string         `json:"bedrockRequestId"`
}

// Discrepancy represents a single discrepancy found during verification.
type Discrepancy struct {
	Item     string `json:"item"`               // Product name
	Expected string `json:"expected"`           // Expected location
	Found    string `json:"found"`              // Actual location (empty if missing)
	Type     string `json:"type"`               // Type of discrepancy (MISSING, MISPLACED, INCORRECT_PRODUCT)
	Severity string `json:"severity,omitempty"` // Severity of the discrepancy (LOW, MEDIUM, HIGH)
}

// Turn1Response is the output payload from ExecuteTurn1Combined.
type Turn1Response struct {
	S3Refs  Turn1ResponseS3Refs `json:"s3Refs"`
	Status  VerificationStatus  `json:"status"`
	Summary Summary             `json:"summary"`
}

// Turn2Response is the output payload from ExecuteTurn2Combined.
type Turn2Response struct {
	S3Refs              Turn2ResponseS3Refs `json:"s3Refs"`
	Status              VerificationStatus  `json:"status"`
	Summary             Summary             `json:"summary"`
	Discrepancies       []Discrepancy       `json:"discrepancies"`
	VerificationOutcome string              `json:"verificationOutcome"` // CORRECT or INCORRECT
}
