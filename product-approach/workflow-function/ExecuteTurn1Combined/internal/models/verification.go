// internal/models/verification.go
package models

import (
	"time"
)

// VerificationContext carries metadata to drive prompt generation.
// This is compatible with the schema package but maintains local structure
type VerificationContext struct {
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