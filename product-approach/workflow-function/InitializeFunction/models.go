// models.go
package main

import "time"

// -------------------------
// Request/Response Models
// -------------------------

// Verification types
const (
    VerificationTypeLayoutVsChecking   = "LAYOUT_VS_CHECKING"
    VerificationTypePreviousVsCurrent  = "PREVIOUS_VS_CURRENT"
)

// InitRequest represents the input payload to the Lambda function
type InitRequest struct {
    VerificationType      string              `json:"verificationType"`
    ReferenceImageUrl     string              `json:"referenceImageUrl"`
    CheckingImageUrl      string              `json:"checkingImageUrl"`
    VendingMachineId      string              `json:"vendingMachineId,omitempty"`
    LayoutId              int                 `json:"layoutId,omitempty"`
    LayoutPrefix          string              `json:"layoutPrefix,omitempty"`
    PreviousVerificationId string             `json:"previousVerificationId,omitempty"`
    ConversationConfig    *ConversationConfig `json:"conversationConfig,omitempty"`
    RequestId             string              `json:"requestId,omitempty"`
    RequestTimestamp      string              `json:"requestTimestamp,omitempty"`
    NotificationEnabled   bool                `json:"notificationEnabled"`
}

// InitResponse matches the desired output schema
type InitResponse struct {
    VerificationContext *VerificationContext `json:"verificationContext"`
    Message             string               `json:"message"`
}

// VerificationContext represents the output returned from the Lambda function
type VerificationContext struct {
    VerificationId        string               `json:"verificationId"`
    VerificationAt        string               `json:"verificationAt"`
    Status                string               `json:"status"`
    VerificationType      string               `json:"verificationType"`
    ConversationType      string               `json:"conversationType,omitempty"`
    VendingMachineId      string               `json:"vendingMachineId,omitempty"`
    LayoutId              int                  `json:"layoutId,omitempty"`
    LayoutPrefix          string               `json:"layoutPrefix,omitempty"`
    PreviousVerificationId string              `json:"previousVerificationId,omitempty"`
    ReferenceImageUrl     string               `json:"referenceImageUrl"`
    CheckingImageUrl      string               `json:"checkingImageUrl"`
    TurnConfig            *TurnConfig          `json:"turnConfig,omitempty"`
    TurnTimestamps        *TurnTimestamps      `json:"turnTimestamps,omitempty"`
    RequestMetadata       *RequestMetadata     `json:"requestMetadata,omitempty"`
    ResourceValidation    *ResourceValidation  `json:"resourceValidation,omitempty"`
    NotificationEnabled   bool                 `json:"notificationEnabled"`
}

// -------------------------
// Conversation Models
// -------------------------

type ConversationConfig struct {
    Type     string `json:"type"`
    MaxTurns int    `json:"maxTurns"`
}

type TurnConfig struct {
    MaxTurns           int `json:"maxTurns"`
    ReferenceImageTurn int `json:"referenceImageTurn"`
    CheckingImageTurn  int `json:"checkingImageTurn"`
}

type TurnTimestamps struct {
    Initialized string  `json:"initialized"`
    Turn1       *string `json:"turn1"`
    Turn2       *string `json:"turn2"`
    Completed   *string `json:"completed"`
}

type RequestMetadata struct {
    RequestId         string `json:"requestId"`
    RequestTimestamp  string `json:"requestTimestamp"`
    ProcessingStarted string `json:"processingStarted"`
}

// -------------------------
// Validation Models
// -------------------------

type ResourceValidation struct {
    LayoutExists         bool   `json:"layoutExists,omitempty"`
    ReferenceImageExists bool   `json:"referenceImageExists"`
    CheckingImageExists  bool   `json:"checkingImageExists"`
    ValidationTimestamp  string `json:"validationTimestamp"`
}

// -------------------------
// Historical Context Models
// -------------------------

type HistoricalContext struct {
    PreviousVerificationId     string               `json:"previousVerificationId"`
    PreviousVerificationAt     string               `json:"previousVerificationAt"`
    PreviousVerificationStatus string               `json:"previousVerificationStatus"`
    HoursSinceLastVerification float64              `json:"hoursSinceLastVerification"`
    MachineStructure           *MachineStructure    `json:"machineStructure,omitempty"`
    VerificationSummary        *VerificationSummary `json:"verificationSummary,omitempty"`
    CheckingStatus             map[string]string    `json:"checkingStatus,omitempty"`
}

type MachineStructure struct {
    RowCount      int      `json:"rowCount"`
    ColumnsPerRow int      `json:"columnsPerRow"`
    RowOrder      []string `json:"rowOrder"`
    ColumnOrder   []string `json:"columnOrder"`
}

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

// -------------------------
// Storage Models
// -------------------------

type DynamoDBVerificationItem struct {
    VerificationId         string `dynamodbav:"verificationId"`
    VerificationAt         string `dynamodbav:"verificationAt"`
    Status                 string `dynamodbav:"status"`
    VerificationType       string `dynamodbav:"verificationType"`
    VendingMachineId       string `dynamodbav:"vendingMachineId,omitempty"`
    LayoutId               int    `dynamodbav:"layoutId,omitempty"`
    LayoutPrefix           string `dynamodbav:"layoutPrefix,omitempty"`
    PreviousVerificationId string `dynamodbav:"previousVerificationId,omitempty"`
    ReferenceImageUrl      string `dynamodbav:"referenceImageUrl"`
    CheckingImageUrl       string `dynamodbav:"checkingImageUrl"`
    RequestId              string `dynamodbav:"requestId,omitempty"`
    NotificationEnabled    bool   `dynamodbav:"notificationEnabled"`
    TTL                    int64  `dynamodbav:"ttl,omitempty"`
}

// -------------------------
// Utility Models & Helpers
// -------------------------

type S3URL struct {
    Bucket string
    Key    string
}

// Status constants
const (
    StatusInitialized   = "INITIALIZED"
    StatusProcessing    = "PROCESSING"
    StatusImagesFetched = "IMAGES_FETCHED"
    StatusCompleted     = "COMPLETED"
    StatusFailed        = "FAILED"
)

// FormatISO8601 returns now in RFC3339
func FormatISO8601() string {
    return time.Now().UTC().Format(time.RFC3339)
}

// GetCurrentTimestamp returns time formatted for IDs
func GetCurrentTimestamp() string {
    return time.Now().UTC().Format("20060102150405")
}
