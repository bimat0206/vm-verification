
// Package internal provides shared types across ExecuteTurn1 functions
package internal

import (
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// S3 State Management Category Structure - matching design specifications
const (
	// Categories
	CategoryProcessing = "processing"
	CategoryImages     = "images"
	CategoryPrompts    = "prompts" 
	CategoryResponses  = "responses"
)

// Standard filenames for each category - matching design specifications
const (
	// Processing category files
	FileInitialization   = "initialization.json"
	FileLayoutMetadata   = "layout-metadata.json"
	FileHistoricalContext = "historical-context.json"
	FileConversationState = "conversation-state.json"
	
	// Images category files
	FileImageMetadata    = "metadata.json"
	FileReferenceBase64  = "reference-base64.base64"
	FileCheckingBase64   = "checking-base64.base64"
	
	// Prompts category files
	FileSystemPrompt     = "system-prompt.json"
	
	// Responses category files
	FileTurn1Response    = "turn1-response.json"
	FileTurn1Thinking    = "turn1-thinking.json"
	FileTurn2Response    = "turn2-response.json"
	FileTurn2Thinking    = "turn2-thinking.json"
)

// S3 path date components
const (
	DatePathFormat = "%04d/%02d/%02d/%s" // YYYY/MM/DD/verificationId
)

// Storage structure information
const (
	Base64Extension = ".base64"
	JsonExtension = ".json"
)

// StateReferences contains all S3 object references needed for the workflow
// Updated to match the correct folder structure
type StateReferences struct {
	VerificationId      string              `json:"verificationId"`
	DatePartition       string              `json:"datePartition"` // YYYY/MM/DD format
	
	// Processing category references
	Initialization      *s3state.Reference  `json:"initialization"`
	LayoutMetadata      *s3state.Reference  `json:"layoutMetadata,omitempty"`
	HistoricalContext   *s3state.Reference  `json:"historicalContext,omitempty"`
	ConversationState   *s3state.Reference  `json:"conversationState,omitempty"`
	
	// Images category references
	ImageMetadata       *s3state.Reference  `json:"imageMetadata"`
	ReferenceBase64     *s3state.Reference  `json:"referenceBase64,omitempty"`
	CheckingBase64      *s3state.Reference  `json:"checkingBase64,omitempty"`
	
	// Prompts category references
	SystemPrompt        *s3state.Reference  `json:"systemPrompt"`
	
	// Responses category references
	Turn1Response       *s3state.Reference  `json:"turn1Response,omitempty"`
	Turn1Thinking       *s3state.Reference  `json:"turn1Thinking,omitempty"`
	Turn2Response       *s3state.Reference  `json:"turn2Response,omitempty"`
	Turn2Thinking       *s3state.Reference  `json:"turn2Thinking,omitempty"`
	FinalResults        *s3state.Reference  `json:"finalResults,omitempty"`
	StorageResult       *s3state.Reference  `json:"storageResult,omitempty"`
	NotificationResult  *s3state.Reference  `json:"notificationResult,omitempty"`
}

// HybridStorageConfig defines configuration for hybrid storage of Base64 data
type HybridStorageConfig struct {
	EnableHybridStorage     bool   `json:"enableHybridStorage"`
	TempBase64Bucket        string `json:"tempBase64Bucket"`
	Base64SizeThreshold     int64  `json:"base64SizeThreshold"`
	Base64RetrievalTimeout  int    `json:"base64RetrievalTimeout"`
	MaxInlineBase64Size     int64  `json:"maxInlineBase64Size"`
	Base64RetrievalRetry    int    `json:"base64RetrievalRetry"`
	Base64RetrievalBackoff  int    `json:"base64RetrievalBackoff"`
}

// StepFunctionInput represents the input to the step function
type StepFunctionInput struct {
	StateReferences *StateReferences       `json:"stateReferences"`
	S3References *StateReferences       `json:"s3References"` // Added for compatibility with step function
	Config          map[string]interface{} `json:"config,omitempty"`
}

// StepFunctionOutput represents the output from the step function
type StepFunctionOutput struct {
	StateReferences *StateReferences       `json:"stateReferences"`
	S3References *StateReferences       `json:"s3References"` // Added for compatibility with step function
	Status          string                 `json:"status"`
	Summary         map[string]interface{} `json:"summary,omitempty"`
	Error           *schema.ErrorInfo      `json:"error,omitempty"`
}
