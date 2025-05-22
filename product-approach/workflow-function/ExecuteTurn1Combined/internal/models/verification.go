// internal/models/verification.go
package models

import "time"

// ExecutionStage represents the current stage of the verification workflow.
type ExecutionStage string

const (
	StageValidation        ExecutionStage = "validation"
	StageContextLoading    ExecutionStage = "context_loading"
	StagePromptGeneration  ExecutionStage = "prompt_generation"
	StageBedrockCall       ExecutionStage = "bedrock_invocation"
	StageProcessing        ExecutionStage = "response_processing"
	StageStorage           ExecutionStage = "state_storage"
	StageDynamoDB          ExecutionStage = "dynamodb_update"
	StageReferenceAnalysis ExecutionStage = "reference_analysis"
)

// VerificationStatus indicates the lifecycle status of a verification.
type VerificationStatus string

const (
	StatusTurn1Started        VerificationStatus = "TURN1_STARTED"
	StatusTurn1PromptPrepared VerificationStatus = "TURN1_PROMPT_PREPARED"
	StatusTurn1Completed      VerificationStatus = "TURN1_COMPLETED"
)

// VerificationContext carries metadata to drive prompt generation.
type VerificationContext struct {
	// Type of verification: e.g. "LAYOUT_VS_CHECKING" or "PREVIOUS_VS_CURRENT"
	VerificationType string `json:"verificationType"`
	// Arbitrary planogram/layout details for LAYOUT_VS_CHECKING
	LayoutMetadata map[string]interface{} `json:"layoutMetadata,omitempty"`
	// Historical context (e.g. prior image analysis) for PREVIOUS_VS_CURRENT
	HistoricalContext map[string]interface{} `json:"historicalContext,omitempty"`
}

// ConversationTurn records a single step in the verification dialogue.
type ConversationTurn struct {
	VerificationID   string      `dynamodbav:"verificationId" json:"verificationId"`
	TurnID           int         `dynamodbav:"turnId" json:"turnId"`
	RawResponseRef   S3Reference `dynamodbav:"rawResponseRef" json:"rawResponseRef"`
	ProcessedRef     S3Reference `dynamodbav:"processedRef" json:"processedRef"`
	TokenUsage       TokenUsage  `dynamodbav:"tokenUsage" json:"tokenUsage"`
	BedrockRequestID string      `dynamodbav:"bedrockRequestId" json:"bedrockRequestId"`
	Timestamp        time.Time   `dynamodbav:"timestamp" json:"timestamp"`
}
