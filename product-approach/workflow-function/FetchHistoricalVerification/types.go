package main

// InputEvent represents the Lambda function input event
type InputEvent struct {
	VerificationContext VerificationContext `json:"verificationContext"`
}

// OutputEvent represents the Lambda function output event
type OutputEvent struct {
	HistoricalContext HistoricalContext `json:"historicalContext"`
}

// VerificationContext contains the verification details
type VerificationContext struct {
	VerificationID    string `json:"verificationId"`
	VerificationAt    string `json:"verificationAt"`
	VerificationType  string `json:"verificationType"`
	ReferenceImageURL string `json:"referenceImageUrl"`
	CheckingImageURL  string `json:"checkingImageUrl"`
	VendingMachineID  string `json:"vendingMachineId"`
	LayoutID          int    `json:"layoutId,omitempty"`
	LayoutPrefix      string `json:"layoutPrefix,omitempty"`
}

// HistoricalContext represents the historical verification data
type HistoricalContext struct {
	PreviousVerificationID        string              `json:"previousVerificationId"`
	PreviousVerificationAt        string              `json:"previousVerificationAt"`
	PreviousVerificationStatus    string              `json:"previousVerificationStatus"`
	HoursSinceLastVerification    float64             `json:"hoursSinceLastVerification"`
	MachineStructure              MachineStructure    `json:"machineStructure"`
	CheckingStatus                map[string]string   `json:"checkingStatus"`
	VerificationSummary           VerificationSummary `json:"verificationSummary"`
}

// MachineStructure represents the physical structure of the vending machine
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// VerificationSummary contains the summary of the verification result
type VerificationSummary struct {
	TotalPositionsChecked  int     `json:"totalPositionsChecked"`
	CorrectPositions       int     `json:"correctPositions"`
	DiscrepantPositions    int     `json:"discrepantPositions"`
	MissingProducts        int     `json:"missingProducts"`
	IncorrectProductTypes  int     `json:"incorrectProductTypes"`
	UnexpectedProducts     int     `json:"unexpectedProducts"`
	EmptyPositionsCount    int     `json:"emptyPositionsCount"`
	OverallAccuracy        float64 `json:"overallAccuracy"`
	OverallConfidence      float64 `json:"overallConfidence"`
	VerificationStatus     string  `json:"verificationStatus"`
	VerificationOutcome    string  `json:"verificationOutcome"`
}

// VerificationRecord represents the DynamoDB record for verification results
type VerificationRecord struct {
	VerificationID         string              `json:"verificationId"`
	VerificationAt         string              `json:"verificationAt"`
	VerificationType       string              `json:"verificationType"`
	VendingMachineID       string              `json:"vendingMachineId"`
	CheckingImageURL       string              `json:"checkingImageUrl"`
	ReferenceImageURL      string              `json:"referenceImageUrl"`
	VerificationStatus     string              `json:"verificationStatus"`
	MachineStructure       MachineStructure    `json:"machineStructure"`
	CheckingStatus         map[string]string   `json:"checkingStatus"`
	VerificationSummary    VerificationSummary `json:"verificationSummary"`
}