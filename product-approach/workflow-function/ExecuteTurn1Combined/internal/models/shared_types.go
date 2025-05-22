// internal/models/shared_types.go
package models

import (
	"workflow-function/shared/schema"
)

// S3Reference represents a pointer to an object in S3.
// This is a local type that will be standardized across functions
type S3Reference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"etag,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

// ExecutionStage represents the current stage of the verification workflow.
// Using local type with string values for backward compatibility
type ExecutionStage string

// VerificationStatus indicates the lifecycle status of a verification.
// Using local type with string values for backward compatibility
type VerificationStatus string

// Import standardized types from schema package where they exist
type (
	// Core workflow types
	SchemaVerificationContext = schema.VerificationContext
	SchemaWorkflowState      = schema.WorkflowState
	SchemaErrorInfo          = schema.ErrorInfo
	
	// S3 and storage types
	SchemaS3StorageConfig   = schema.S3StorageConfig
	
	// Image related types
	SchemaImageData         = schema.ImageData
	SchemaImageInfo         = schema.ImageInfo
	SchemaImageMetadata     = schema.ImageMetadata
	
	// Bedrock types
	SchemaBedrockMessage    = schema.BedrockMessage
	SchemaBedrockContent    = schema.BedrockContent
	SchemaBedrockImageData  = schema.BedrockImageData
	SchemaBedrockConfig     = schema.BedrockConfig
	SchemaTokenUsage        = schema.TokenUsage
	SchemaCurrentPrompt     = schema.CurrentPrompt
	SchemaSystemPrompt      = schema.SystemPrompt
	SchemaTurnResponse      = schema.TurnResponse
	
	// Layout and conversation types
	SchemaLayoutMetadata    = schema.LayoutMetadata
	SchemaConversationState = schema.ConversationState
	SchemaTurnHistory       = schema.TurnHistory
	
	// Validation types
	SchemaValidationError   = schema.ValidationError
	SchemaValidationErrors  = schema.Errors
)

// Local constants for ExecutionStage - backward compatibility
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

// Local constants for VerificationStatus - backward compatibility
const (
	StatusTurn1Started        VerificationStatus = "TURN1_STARTED"
	StatusTurn1PromptPrepared VerificationStatus = "TURN1_PROMPT_PREPARED"
	StatusTurn1Completed      VerificationStatus = "TURN1_COMPLETED"
)

// Constants from schema package
const (
	// Schema version
	SchemaVersion = schema.SchemaVersion
	
	// Verification types
	VerificationTypeLayoutVsChecking  = schema.VerificationTypeLayoutVsChecking
	VerificationTypePreviousVsCurrent = schema.VerificationTypePreviousVsCurrent
	
	// Schema status constants (for integration)
	SchemaStatusVerificationRequested       = schema.StatusVerificationRequested
	SchemaStatusVerificationInitialized     = schema.StatusVerificationInitialized
	SchemaStatusFetchingImages              = schema.StatusFetchingImages
	SchemaStatusImagesFetched               = schema.StatusImagesFetched
	SchemaStatusPromptPrepared              = schema.StatusPromptPrepared
	SchemaStatusTurn1PromptReady            = schema.StatusTurn1PromptReady
	SchemaStatusTurn1Completed              = schema.StatusTurn1Completed
	SchemaStatusTurn1Processed              = schema.StatusTurn1Processed
	SchemaStatusTurn2PromptReady            = schema.StatusTurn2PromptReady
	SchemaStatusTurn2Completed              = schema.StatusTurn2Completed
	SchemaStatusTurn2Processed              = schema.StatusTurn2Processed
	SchemaStatusResultsFinalized            = schema.StatusResultsFinalized
	SchemaStatusResultsStored               = schema.StatusResultsStored
	SchemaStatusNotificationSent            = schema.StatusNotificationSent
	SchemaStatusCompleted                   = schema.StatusCompleted
	SchemaStatusInitializationFailed        = schema.StatusInitializationFailed
	SchemaStatusHistoricalFetchFailed       = schema.StatusHistoricalFetchFailed
	SchemaStatusImageFetchFailed            = schema.StatusImageFetchFailed
	SchemaStatusBedrockProcessingFailed     = schema.StatusBedrockProcessingFailed
	SchemaStatusVerificationFailed          = schema.StatusVerificationFailed
	
	// Storage constants
	StorageMethodS3Temporary = schema.StorageMethodS3Temporary
	BedrockMaxImageSize      = schema.BedrockMaxImageSize
	TempBase64KeyPrefix      = schema.TempBase64KeyPrefix
	TempBase64TTL            = schema.TempBase64TTL
)

// Validation functions from schema package
var (
	ValidateVerificationContext = schema.ValidateVerificationContext
	ValidateWorkflowState       = schema.ValidateWorkflowState
	ValidateImageInfo           = schema.ValidateImageInfo
	ValidateImageData           = schema.ValidateImageData
	ValidateBedrockMessages     = schema.ValidateBedrockMessages
	ValidateCurrentPrompt       = schema.ValidateCurrentPrompt
	ValidateBedrockConfig       = schema.ValidateBedrockConfig
)

// Helper functions from schema package
var (
	FormatISO8601         = schema.FormatISO8601
	GetCurrentTimestamp   = schema.GetCurrentTimestamp
	BuildBedrockMessage   = schema.BuildBedrockMessage
	BuildBedrockMessages  = schema.BuildBedrockMessages
)

// ConvertToSchemaStatus converts local status to schema status
func ConvertToSchemaStatus(localStatus VerificationStatus) string {
	switch localStatus {
	case StatusTurn1Started:
		return SchemaStatusTurn1PromptReady
	case StatusTurn1PromptPrepared:
		return SchemaStatusTurn1PromptReady
	case StatusTurn1Completed:
		return SchemaStatusTurn1Completed
	default:
		return string(localStatus)
	}
}

// ConvertFromSchemaStatus converts schema status to local status
func ConvertFromSchemaStatus(schemaStatus string) VerificationStatus {
	switch schemaStatus {
	case SchemaStatusTurn1PromptReady:
		return StatusTurn1PromptPrepared
	case SchemaStatusTurn1Completed:
		return StatusTurn1Completed
	default:
		return VerificationStatus(schemaStatus)
	}
}