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
	VerificationID string                        `json:"verificationId"`
	S3References   map[string]*s3state.Reference `json:"s3References"`
	Status         string                        `json:"status"`
	Summary        map[string]interface{}        `json:"summary,omitempty"`
}

// HistoricalContext represents the historical verification data
type HistoricalContext struct {
	PreviousVerificationID     string                     `json:"previousVerificationId"`
	PreviousVerificationAt     string                     `json:"previousVerificationAt"`
	PreviousVerificationStatus string                     `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64                    `json:"hoursSinceLastVerification"`
	MachineStructure           schema.MachineStructure    `json:"machineStructure"`
	CheckingStatus             map[string]string          `json:"checkingStatus"`
	VerificationSummary        schema.VerificationSummary `json:"verificationSummary"`
}

// VerificationRecord represents the DynamoDB record for verification results
type VerificationRecord struct {
	VerificationID      string                     `json:"verificationId" dynamodbav:"VerificationId"`
	VerificationAt      string                     `json:"verificationAt" dynamodbav:"VerificationAt"`
	VerificationType    string                     `json:"verificationType" dynamodbav:"VerificationType"`
	VendingMachineID    string                     `json:"vendingMachineId" dynamodbav:"VendingMachineId"`
	CheckingImageURL    string                     `json:"checkingImageUrl" dynamodbav:"CheckingImageUrl"`
	ReferenceImageURL   string                     `json:"referenceImageUrl" dynamodbav:"ReferenceImageUrl"`
	VerificationStatus  string                     `json:"verificationStatus" dynamodbav:"VerificationStatus"`
	MachineStructure    schema.MachineStructure    `json:"machineStructure" dynamodbav:"MachineStructure"`
	CheckingStatus      map[string]string          `json:"checkingStatus" dynamodbav:"CheckingStatus"`
	VerificationSummary schema.VerificationSummary `json:"verificationSummary" dynamodbav:"VerificationSummary"`
}
