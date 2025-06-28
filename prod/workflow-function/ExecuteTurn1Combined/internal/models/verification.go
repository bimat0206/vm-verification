// internal/models/verification.go
package models

import (
	"fmt"
	"time"
)

// VerificationContext carries metadata to drive prompt generation.
// This is compatible with the schema package but maintains local structure
type VerificationContext struct {
	// Timestamp when the verification was first initialized
	VerificationAt string `json:"verificationAt,omitempty"`
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

// Validate performs validation on the VerificationContext to ensure required fields are present
func (vc *VerificationContext) Validate() error {
	if vc.VerificationType == "" {
		return fmt.Errorf("verification type is required")
	}

	switch vc.VerificationType {
	case "LAYOUT_VS_CHECKING":
		if vc.LayoutMetadata == nil {
			return fmt.Errorf("LayoutMetadata is required for LAYOUT_VS_CHECKING verification type")
		}

	case "PREVIOUS_VS_CURRENT":
		// For PREVIOUS_VS_CURRENT verification type, HistoricalContext is optional
		// The historical context will be loaded from S3 after validation
		// When sourceType is NO_HISTORICAL_DATA, no historical data exists and validation should pass
		// The actual historical context validation will be handled during template processing
	}

	return nil
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
