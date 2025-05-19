// Package internal provides shared types across ExecuteTurn1 functions
package internal

import (
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// StateReferences contains all S3 object references needed for the workflow
type StateReferences struct {
	VerificationId      string              `json:"verificationId"`
	VerificationContext *s3state.Reference  `json:"verificationContext"`
	SystemPrompt        *s3state.Reference  `json:"systemPrompt"`
	Images              *s3state.Reference  `json:"images,omitempty"`
	BedrockConfig       *s3state.Reference  `json:"bedrockConfig"`
	ConversationState   *s3state.Reference  `json:"conversationState,omitempty"`
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
	Config          map[string]interface{} `json:"config,omitempty"`
}

// StepFunctionOutput represents the output from the step function
type StepFunctionOutput struct {
	StateReferences *StateReferences       `json:"stateReferences"`
	Status          string                 `json:"status"`
	Summary         map[string]interface{} `json:"summary,omitempty"`
	Error           *schema.ErrorInfo      `json:"error,omitempty"`
}