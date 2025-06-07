package models

// Input to the Lambda function

type S3Reference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"etag,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

type InputS3References struct {
	InitializationS3Ref S3Reference `json:"processing_initialization"`
	Turn1ProcessedS3Ref S3Reference `json:"turn1Processed"`
	Turn2ProcessedS3Ref S3Reference `json:"turn2Processed"`
}

type LambdaInput struct {
	VerificationID string            `json:"verificationId"`
	References     InputS3References `json:"s3References"`
	Status         string            `json:"status"`
}

// Output structs

type OutputDiscrepancyDetails struct {
	MissingProducts       int `json:"missing_products" dynamodbav:"missing_products"`
	IncorrectProductTypes int `json:"incorrect_product_types" dynamodbav:"incorrect_product_types"`
	UnexpectedProducts    int `json:"unexpected_products" dynamodbav:"unexpected_products"`
}

type OutputVerificationSummary struct {
	TotalPositionsChecked         int                      `json:"total_positions_checked" dynamodbav:"total_positions_checked"`
	CorrectPositions              int                      `json:"correct_positions" dynamodbav:"correct_positions"`
	DiscrepantPositions           int                      `json:"discrepant_positions" dynamodbav:"discrepant_positions"`
	DiscrepancyDetails            OutputDiscrepancyDetails `json:"discrepancy_details" dynamodbav:"discrepancy_details"`
	EmptyPositionsInCheckingImage int                      `json:"empty_positions_in_checking_image" dynamodbav:"empty_positions_in_checking_image"`
	OverallAccuracy               string                   `json:"overall_accuracy" dynamodbav:"overall_accuracy"`
	OverallConfidence             string                   `json:"overall_confidence" dynamodbav:"overall_confidence"`
	VerificationOutcome           string                   `json:"verification_outcome" dynamodbav:"verification_outcome"`
}

type LambdaOutput struct {
	VerificationID      string                    `json:"verificationId"`
	VerificationAt      string                    `json:"verificationAt"`
	Status              string                    `json:"status"`
	VerificationStatus  string                    `json:"verificationStatus"`
	VerificationSummary OutputVerificationSummary `json:"verificationSummary"`
	Message             string                    `json:"message"`
}

// InitializationData from initialization.json with nested verificationContext structure

type InitializationData struct {
	SchemaVersion       string                   `json:"schemaVersion"`
	VerificationContext *VerificationContextData `json:"verificationContext"`
	SystemPrompt        map[string]interface{}   `json:"systemPrompt,omitempty"`
}

// VerificationContextData represents the nested verification context
type VerificationContextData struct {
	VerificationID         string                 `json:"verificationId"`
	VerificationAt         string                 `json:"verificationAt"`
	Status                 string                 `json:"status"`
	VerificationType       string                 `json:"verificationType"`
	LayoutID               int                    `json:"layoutId,omitempty"`
	LayoutPrefix           string                 `json:"layoutPrefix,omitempty"`
	ReferenceImageUrl      string                 `json:"referenceImageUrl"`
	CheckingImageUrl       string                 `json:"checkingImageUrl"`
	VendingMachineID       string                 `json:"vendingMachineId,omitempty"`
	PreviousVerificationID string                 `json:"previousVerificationId,omitempty"`
	ResourceValidation     map[string]interface{} `json:"resourceValidation,omitempty"`
	LastUpdatedAt          string                 `json:"lastUpdatedAt,omitempty"`
}

// Parsed Turn2 data

type Turn2ParsedData struct {
	VerificationSummary OutputVerificationSummary
	InitialConfirmation string
	VerificationStatus  string
}

// DynamoDB item

type VerificationResultItem struct {
	VerificationID         string                    `dynamodbav:"verificationId"`
	VerificationAt         string                    `dynamodbav:"verificationAt"`
	VerificationType       string                    `dynamodbav:"verificationType"`
	LayoutID               *int                      `dynamodbav:"layoutId,omitempty"`
	LayoutPrefix           string                    `dynamodbav:"layoutPrefix,omitempty"`
	VendingMachineID       string                    `dynamodbav:"vendingMachineId,omitempty"`
	Location               string                    `dynamodbav:"location,omitempty"`
	ReferenceImageUrl      string                    `dynamodbav:"referenceImageUrl"`
	CheckingImageUrl       string                    `dynamodbav:"checkingImageUrl"`
	VerificationStatus     string                    `dynamodbav:"verificationStatus"`
	CurrentStatus          string                    `dynamodbav:"currentStatus"`
	LastUpdatedAt          string                    `dynamodbav:"lastUpdatedAt"`
	ProcessingStartedAt    string                    `dynamodbav:"processingStartedAt,omitempty"`
	StatusHistory          []map[string]interface{}  `dynamodbav:"statusHistory,omitempty"`
	ProcessingMetrics      map[string]interface{}    `dynamodbav:"processingMetrics,omitempty"`
	ErrorTracking          map[string]interface{}    `dynamodbav:"errorTracking,omitempty"`
	MachineStructure       map[string]interface{}    `dynamodbav:"machineStructure,omitempty"`
	InitialConfirmation    string                    `dynamodbav:"initialConfirmation,omitempty"`
	CorrectedRows          []string                  `dynamodbav:"correctedRows,omitempty"`
	EmptySlotReport        map[string]interface{}    `dynamodbav:"emptySlotReport,omitempty"`
	ReferenceStatus        map[string]string         `dynamodbav:"referenceStatus,omitempty"`
	CheckingStatus         map[string]string         `dynamodbav:"checkingStatus,omitempty"`
	Discrepancies          []map[string]interface{}  `dynamodbav:"discrepancies,omitempty"`
	VerificationSummary    OutputVerificationSummary `dynamodbav:"verificationSummary"`
	Metadata               map[string]interface{}    `dynamodbav:"metadata,omitempty"`
	PreviousVerificationID string                    `dynamodbav:"previousVerificationId,omitempty"`
}
