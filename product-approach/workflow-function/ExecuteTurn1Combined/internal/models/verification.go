// internal/models/verification.go
package models

import (
	"fmt"
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

// Validate performs validation on the VerificationContext to ensure required fields are present
func (vc *VerificationContext) Validate() error {
	if vc.VerificationType == "" {
		return fmt.Errorf("verification type is required")
	}

	switch vc.VerificationType {
	case "LAYOUT_VS_CHECKING":
		// Ensure layout metadata has minimum required fields
		if vc.LayoutMetadata == nil {
			vc.LayoutMetadata = make(map[string]interface{})
		}

		// Ensure RowCount exists
		if _, ok := vc.LayoutMetadata["RowCount"]; !ok {
			vc.LayoutMetadata["RowCount"] = 6 // Default
		}

		// Ensure ColumnCount exists
		if _, ok := vc.LayoutMetadata["ColumnCount"]; !ok {
			vc.LayoutMetadata["ColumnCount"] = 10 // Default
		}

	case "PREVIOUS_VS_CURRENT":
		// Ensure historical context exists
		if vc.HistoricalContext == nil {
			vc.HistoricalContext = make(map[string]interface{})
		}

		// Ensure required fields have defaults
		if _, ok := vc.HistoricalContext["PreviousVerificationAt"]; !ok {
			vc.HistoricalContext["PreviousVerificationAt"] = "unknown"
		}

		if _, ok := vc.HistoricalContext["HoursSinceLastVerification"]; !ok {
			vc.HistoricalContext["HoursSinceLastVerification"] = 0.0
		}

		if _, ok := vc.HistoricalContext["PreviousVerificationStatus"]; !ok {
			vc.HistoricalContext["PreviousVerificationStatus"] = "unknown"
		}

		// Initialize VerificationSummary if missing
		if _, ok := vc.HistoricalContext["VerificationSummary"]; !ok {
			vc.HistoricalContext["VerificationSummary"] = map[string]interface{}{
				"OverallAccuracy":       0.0,
				"MissingProducts":       0,
				"IncorrectProductTypes": 0,
				"EmptyPositionsCount":   0,
			}
		}
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
