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
	SchemaWorkflowState       = schema.WorkflowState
	SchemaErrorInfo           = schema.ErrorInfo
	SchemaStatusHistoryEntry  = schema.StatusHistoryEntry
	SchemaProcessingMetrics   = schema.ProcessingMetrics
	SchemaErrorTracking       = schema.ErrorTracking
)

// Enhanced response types
type (
	SchemaTurnResponse         = schema.TurnResponse
	SchemaCombinedTurnResponse = schema.CombinedTurnResponse
	SchemaProcessingStage      = schema.ProcessingStage
)

// Conversation and tracking types
