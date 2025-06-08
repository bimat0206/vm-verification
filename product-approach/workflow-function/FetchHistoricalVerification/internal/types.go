package internal

import (
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// InputEvent represents the Lambda function input event
type InputEvent struct {
	VerificationID string                        `json:"verificationId"`
	S3References   map[string]*s3state.Reference `json:"s3References"`
	Status         string                        `json:"status"`
}

// OutputEvent represents the Lambda function output event
type OutputEvent struct {
	VerificationID      string                        `json:"verificationId"`
	S3References        map[string]*s3state.Reference `json:"s3References"`
	Status              string                        `json:"status"`
	VerificationContext *EnhancedVerificationContext  `json:"verificationContext,omitempty"`
}

// EnhancedVerificationContext extends the standard VerificationContext with historical data fields
type EnhancedVerificationContext struct {
	*schema.VerificationContext
	PreviousVerificationAt string `json:"previousVerificationAt,omitempty"`
	Turn2Processed         string `json:"turn2Processed,omitempty"`
	HistoricalDataFound    bool   `json:"historicalDataFound"`
	SourceType             string `json:"sourceType"`
	PreviousStatus         string `json:"previousStatus,omitempty"`
}

// HistoricalContext represents the historical verification data
type HistoricalContext struct {
	VerificationID              string                      `json:"verificationId"`
	VerificationType            string                      `json:"verificationType"`
	ReferenceImageUrl           string                      `json:"referenceImageUrl"`
	CheckingImageUrl            string                      `json:"checkingImageUrl"`
	HistoricalDataFound         bool                        `json:"historicalDataFound"`
	Turn2Processed              string                      `json:"turn2Processed,omitempty"`
	SourceType                  string                      `json:"sourceType"`
	PreviousVerification        *PreviousVerification       `json:"previousVerification"`
	TemporalContext             *TemporalContext            `json:"temporalContext"`
	MachineStructure            *EnhancedMachineStructure   `json:"machineStructure,omitempty"`
	PreviousVerificationSummary *PreviousVerificationSummary `json:"previousVerificationSummary"`
}

// EnhancedMachineStructure extends the schema.MachineStructure with totalPositions
type EnhancedMachineStructure struct {
	RowCount        int      `json:"rowCount"`
	ColumnsPerRow   int      `json:"columnsPerRow"`
	RowOrder        []string `json:"rowOrder"`
	ColumnOrder     []string `json:"columnOrder"`
	TotalPositions  int      `json:"totalPositions"`
}

// PreviousVerification represents the previous verification details
type PreviousVerification struct {
	VerificationID     string `json:"verificationId"`
	VerificationAt     string `json:"verificationAt"`
	VerificationStatus string `json:"verificationStatus"`
	VendingMachineID   string `json:"vendingMachineId"`
	Location           string `json:"location"`
	LayoutID           string `json:"layoutId"`
	LayoutPrefix       string `json:"layoutPrefix"`
}

// TemporalContext represents temporal information about the verification
type TemporalContext struct {
	HoursSinceLastVerification float64 `json:"hoursSinceLastVerification"`
	DaysSinceLastVerification  float64 `json:"daysSinceLastVerification"`
	BusinessDaysSince          int     `json:"businessDaysSince"`
}

// PreviousVerificationSummary represents the summary of the previous verification
type PreviousVerificationSummary struct {
	TotalPositionsChecked   int                    `json:"total_positions_checked"`
	CorrectPositions        int                    `json:"correct_positions"`
	DiscrepantPositions     int                    `json:"discrepant_positions"`
	DiscrepancyDetails      *DiscrepancyDetails    `json:"discrepancy_details"`
	EmptyPositionsInChecking int                   `json:"empty_positions_in_checking_image"`
	OverallAccuracy         string                 `json:"overall_accuracy"`
	OverallConfidence       string                 `json:"overall_confidence"`
	VerificationOutcome     string                 `json:"verification_outcome"`
}

// DiscrepancyDetails represents detailed discrepancy information
type DiscrepancyDetails struct {
	MissingProducts        int `json:"missing_products"`
	IncorrectProductTypes  int `json:"incorrect_product_types"`
	UnexpectedProducts     int `json:"unexpected_products"`
}

// VerificationRecord represents the DynamoDB record for verification results
type VerificationRecord struct {
	VerificationID      string                     `json:"verificationId" dynamodbav:"verificationId"`
	VerificationAt      string                     `json:"verificationAt" dynamodbav:"verificationAt"`
	VerificationType    string                     `json:"verificationType" dynamodbav:"verificationType"`
	VendingMachineID    string                     `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	CheckingImageURL    string                     `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
	ReferenceImageURL   string                     `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	VerificationStatus  string                     `json:"verificationStatus" dynamodbav:"verificationStatus"`
	MachineStructure    schema.MachineStructure    `json:"machineStructure" dynamodbav:"machineStructure"`
	CheckingStatus      map[string]string          `json:"checkingStatus" dynamodbav:"checkingStatus"`
	VerificationSummary schema.VerificationSummary `json:"verificationSummary" dynamodbav:"verificationSummary"`
}
