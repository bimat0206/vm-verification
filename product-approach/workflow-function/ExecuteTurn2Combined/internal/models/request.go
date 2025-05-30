// internal/models/request.go
package models

// Note: Turn1Request and Turn1RequestS3Refs are not defined here as ExecuteTurn2Combined
// should not handle Turn1 requests directly. Turn1 artifacts are accessed via Turn1References
// in Turn2RequestS3Refs.

// Turn2Request is the input payload for ExecuteTurn2Combined.
type Turn2Request struct {
	VerificationID      string              `json:"verificationId"`
	VerificationContext VerificationContext `json:"verificationContext"`
	S3Refs              Turn2RequestS3Refs  `json:"s3References"`
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
	CheckingBase64      S3Reference `json:"checkingBase64"`                // checking-base64.base64
	CheckingImageFormat string      `json:"checkingImageFormat,omitempty"` // image content type (e.g. image/png)
}

// PromptRefs holds S3 locations for prompt artifacts.
type PromptRefs struct {
	System      S3Reference `json:"system"` // system-prompt.json
	Turn2Prompt S3Reference `json:"turn2Prompt"`
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

// Note: Turn1ResponseS3Refs is not defined here as it belongs to ExecuteTurn1Combined.
// Turn1 artifacts are accessed via Turn1References in Turn2RequestS3Refs.

// Turn2ResponseS3Refs groups the S3 references created by Turn-2.
type Turn2ResponseS3Refs struct {
	RawResponse       S3Reference `json:"rawResponse"`
	ProcessedResponse S3Reference `json:"processedResponse"`
}

// Summary contains metrics and identifiers for the execution.
type Summary struct {
	AnalysisStage         ExecutionStage `json:"analysisStage"`
	VerificationType      string         `json:"verificationType,omitempty"`
	ProcessingTimeMs      int64          `json:"processingTimeMs"`
	TokenUsage            TokenUsage     `json:"tokenUsage"`
	BedrockLatencyMs      int64          `json:"bedrockLatencyMs,omitempty"`
	BedrockRequestID      string         `json:"bedrockRequestId"`
	DiscrepanciesFound    int            `json:"discrepanciesFound"`
	ComparisonCompleted   bool           `json:"comparisonCompleted"`
	ConversationCompleted bool           `json:"conversationCompleted"`
	DynamodbUpdated       bool           `json:"dynamodbUpdated"`
	S3StorageCompleted    bool           `json:"s3StorageCompleted,omitempty"`
}

// Discrepancy represents a single discrepancy found during verification.
type Discrepancy struct {
	Item     string `json:"item"`               // Product name
	Expected string `json:"expected"`           // Expected location
	Found    string `json:"found"`              // Actual location (empty if missing)
	Type     string `json:"type"`               // Type of discrepancy (MISSING, MISPLACED, INCORRECT_PRODUCT)
	Severity string `json:"severity,omitempty"` // Severity of the discrepancy (LOW, MEDIUM, HIGH)
}

// Note: Turn1Response is not defined here as it belongs to ExecuteTurn1Combined.
// Turn1 response data is accessed via the shared schema types.

// Turn2Response is the output payload from ExecuteTurn2Combined.
type Turn2Response struct {
	S3Refs              Turn2ResponseS3Refs `json:"s3References"`
	Status              VerificationStatus  `json:"status"`
	Summary             Summary             `json:"summary"`
	Discrepancies       []Discrepancy       `json:"discrepancies"`
	VerificationOutcome string              `json:"verificationOutcome"` // CORRECT or INCORRECT
}
