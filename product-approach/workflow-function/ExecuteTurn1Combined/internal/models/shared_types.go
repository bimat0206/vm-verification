// internal/models/shared_types.go - CLEAN AND SIMPLE VERSION
package models

import (
	"workflow-function/shared/schema"
)

// ===================================================================
// LOCAL TYPE DEFINITIONS (for this function's specific needs)
// ===================================================================

// S3Reference represents a pointer to an object in S3
type S3Reference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"etag,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

// ExecutionStage represents the current stage of processing
type ExecutionStage string

// VerificationStatus represents the current status of verification
type VerificationStatus string

// ===================================================================
// LOCAL CONSTANTS (for backward compatibility)
// ===================================================================

// Execution stages
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

// Local verification statuses
const (
	StatusTurn1Started        VerificationStatus = "TURN1_STARTED"
	StatusTurn1PromptPrepared VerificationStatus = "TURN1_PROMPT_PREPARED"
	StatusTurn1Completed      VerificationStatus = "TURN1_COMPLETED"
)

// ===================================================================
// SCHEMA TYPE ALIASES (direct imports from shared schema)
// ===================================================================

// Core workflow types
type (
	SchemaVerificationContext = schema.VerificationContext
	SchemaWorkflowState      = schema.WorkflowState
	SchemaErrorInfo          = schema.ErrorInfo
	SchemaStatusHistoryEntry = schema.StatusHistoryEntry
	SchemaProcessingMetrics  = schema.ProcessingMetrics
	SchemaErrorTracking      = schema.ErrorTracking
)

// Enhanced response types
type (
	SchemaTurnResponse         = schema.TurnResponse
	SchemaCombinedTurnResponse = schema.CombinedTurnResponse
	SchemaProcessingStage      = schema.ProcessingStage
)

// Conversation and tracking types
type (
	SchemaConversationTracker = schema.ConversationTracker
	SchemaTemplateProcessor   = schema.TemplateProcessor
)

// Bedrock and AI types
type (
	SchemaBedrockMessage   = schema.BedrockMessage
	SchemaBedrockConfig    = schema.BedrockConfig
	SchemaTokenUsage       = schema.TokenUsage
	SchemaCurrentPrompt    = schema.CurrentPrompt
	SchemaSystemPrompt     = schema.SystemPrompt
)

// Image and storage types
type (
	SchemaImageData       = schema.ImageData
	SchemaImageInfo       = schema.ImageInfo
	SchemaS3StorageConfig = schema.S3StorageConfig
)

// Layout and metadata types
type (
	SchemaLayoutMetadata = schema.LayoutMetadata
)

// ===================================================================
// SCHEMA CONSTANTS (commonly used ones)
// ===================================================================

const (
	// Schema version
	SchemaVersion = schema.SchemaVersion
	
	// Verification types
	VerificationTypeLayoutVsChecking  = schema.VerificationTypeLayoutVsChecking
	VerificationTypePreviousVsCurrent = schema.VerificationTypePreviousVsCurrent
	
	// Enhanced status constants (for combined functions)
	SchemaStatusTurn1Started              = schema.StatusTurn1Started
	SchemaStatusTurn1ContextLoaded        = schema.StatusTurn1ContextLoaded
	SchemaStatusTurn1PromptPrepared       = schema.StatusTurn1PromptPrepared
	SchemaStatusTurn1ImageLoaded          = schema.StatusTurn1ImageLoaded
	SchemaStatusTurn1BedrockInvoked       = schema.StatusTurn1BedrockInvoked
	SchemaStatusTurn1BedrockCompleted     = schema.StatusTurn1BedrockCompleted
	SchemaStatusTurn1ResponseProcessing   = schema.StatusTurn1ResponseProcessing
	SchemaStatusTurn1Completed            = schema.StatusTurn1Completed
	SchemaStatusTurn1Error                = schema.StatusTurn1Error
	SchemaStatusTemplateProcessingError   = schema.StatusTemplateProcessingError
	
	// General workflow statuses
	SchemaStatusVerificationInitialized  = schema.StatusVerificationInitialized
	SchemaStatusImagesFetched            = schema.StatusImagesFetched
	SchemaStatusCompleted                = schema.StatusCompleted
	SchemaStatusVerificationFailed       = schema.StatusVerificationFailed
)

// ===================================================================
// VALIDATION AND HELPER FUNCTIONS (direct imports)
// ===================================================================

var (
	// Validation functions
	ValidateVerificationContext = schema.ValidateVerificationContext
	ValidateWorkflowState       = schema.ValidateWorkflowState
	ValidateImageData           = schema.ValidateImageData
	ValidateBedrockMessages     = schema.ValidateBedrockMessages
	ValidateCurrentPrompt       = schema.ValidateCurrentPrompt
	
	// Helper functions
	FormatISO8601        = schema.FormatISO8601
	GetCurrentTimestamp  = schema.GetCurrentTimestamp
	BuildBedrockMessage  = schema.BuildBedrockMessage
	BuildBedrockMessages = schema.BuildBedrockMessages
)

// ===================================================================
// SIMPLE CONVERSION FUNCTIONS
// ===================================================================

// ConvertToSchemaStatus converts local status to schema status
func ConvertToSchemaStatus(localStatus VerificationStatus) string {
	switch localStatus {
	case StatusTurn1Started:
		return SchemaStatusTurn1Started
	case StatusTurn1PromptPrepared:
		return SchemaStatusTurn1PromptPrepared
	case StatusTurn1Completed:
		return SchemaStatusTurn1Completed
	default:
		return string(localStatus)
	}
}

// ConvertFromSchemaStatus converts schema status to local status
func ConvertFromSchemaStatus(schemaStatus string) VerificationStatus {
	switch schemaStatus {
	case SchemaStatusTurn1Started:
		return StatusTurn1Started
	case SchemaStatusTurn1PromptPrepared:
		return StatusTurn1PromptPrepared
	case SchemaStatusTurn1Completed:
		return StatusTurn1Completed
	default:
		return VerificationStatus(schemaStatus)
	}
}

// ConvertS3ReferenceToSchema converts local S3Reference to schema format (if needed)
func ConvertS3ReferenceToSchema(localRef S3Reference) map[string]interface{} {
	return map[string]interface{}{
		"bucket": localRef.Bucket,
		"key":    localRef.Key,
		"etag":   localRef.ETag,
		"size":   localRef.Size,
	}
}

// CreateStatusHistoryEntry creates a schema status history entry
func CreateStatusHistoryEntry(status, functionName, stage string, processingTimeMs int64, metrics map[string]interface{}) SchemaStatusHistoryEntry {
	return SchemaStatusHistoryEntry{
		Status:           status,
		Timestamp:        FormatISO8601(),
		FunctionName:     functionName,
		ProcessingTimeMs: processingTimeMs,
		Stage:            stage,
		Metrics:          metrics,
	}
}

// CreateProcessingStage creates a schema processing stage
func CreateProcessingStage(stageName, status string, startTime, endTime string, durationMs int64, metadata map[string]interface{}) SchemaProcessingStage {
	return SchemaProcessingStage{
		StageName: stageName,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  durationMs,
		Status:    status,
		Metadata:  metadata,
	}
}

// CreateVerificationContext creates a basic verification context
func CreateVerificationContext(verificationID, verificationType string) *SchemaVerificationContext {
	return &SchemaVerificationContext{
		VerificationId:   verificationID,
		VerificationAt:   FormatISO8601(),
		Status:           SchemaStatusVerificationInitialized,
		VerificationType: verificationType,
		CurrentStatus:    SchemaStatusVerificationInitialized,
		LastUpdatedAt:    FormatISO8601(),
		StatusHistory:    make([]SchemaStatusHistoryEntry, 0),
		ProcessingMetrics: &SchemaProcessingMetrics{},
		ErrorTracking:    &SchemaErrorTracking{
			HasErrors:        false,
			ErrorHistory:     make([]schema.ErrorInfo, 0),
			RecoveryAttempts: 0,
		},
		NotificationEnabled: false,
	}
}

// CreateConversationTracker creates a basic conversation tracker
func CreateConversationTracker(verificationID string, maxTurns int) *SchemaConversationTracker {
	return &SchemaConversationTracker{
		ConversationId: verificationID,
		CurrentTurn:    0,
		MaxTurns:       maxTurns,
		TurnStatus:     "INITIALIZED",
		ConversationAt: FormatISO8601(),
		History:        make([]interface{}, 0),
		Metadata:       make(map[string]interface{}),
	}
}

// ===================================================================
// UTILITY FUNCTIONS FOR TYPE CHECKING
// ===================================================================

// IsEnhancedStatus checks if a status is an enhanced schema status
func IsEnhancedStatus(status string) bool {
	enhancedStatuses := []string{
		SchemaStatusTurn1Started,
		SchemaStatusTurn1ContextLoaded,
		SchemaStatusTurn1PromptPrepared,
		SchemaStatusTurn1ImageLoaded,
		SchemaStatusTurn1BedrockInvoked,
		SchemaStatusTurn1BedrockCompleted,
		SchemaStatusTurn1ResponseProcessing,
		SchemaStatusTurn1Error,
		SchemaStatusTemplateProcessingError,
	}
	
	for _, enhancedStatus := range enhancedStatuses {
		if status == enhancedStatus {
			return true
		}
	}
	return false
}

// IsVerificationComplete checks if verification is in a completed state
func IsVerificationComplete(status string) bool {
	completedStatuses := []string{
		SchemaStatusTurn1Completed,
		SchemaStatusCompleted,
		SchemaStatusVerificationFailed,
	}
	
	for _, completedStatus := range completedStatuses {
		if status == completedStatus {
			return true
		}
	}
	return false
}

// IsErrorStatus checks if status indicates an error state
func IsErrorStatus(status string) bool {
	return status == SchemaStatusTurn1Error || 
		   status == SchemaStatusTemplateProcessingError || 
		   status == SchemaStatusVerificationFailed
}