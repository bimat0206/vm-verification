package main

import "time"

// -------------------------
// Request/Response Models
// -------------------------

// InitRequest represents the input payload to the Lambda function
type InitRequest struct {
	ReferenceImageUrl  string             `json:"referenceImageUrl"`
	CheckingImageUrl   string             `json:"checkingImageUrl"`
	VendingMachineId   string             `json:"vendingMachineId"` // Optional
	LayoutId           int                `json:"layoutId"`
	LayoutPrefix       string             `json:"layoutPrefix"`
	ConversationConfig *ConversationConfig `json:"conversationConfig,omitempty"`
	RequestId          string             `json:"requestId,omitempty"`
	RequestTimestamp   string             `json:"requestTimestamp,omitempty"`
}

// VerificationContext represents the output returned from the Lambda function
type VerificationContext struct {
	VerificationId    string           `json:"verificationId"`
	VerificationAt    string           `json:"verificationAt"`
	Status            string           `json:"status"`
	ConversationType  string           `json:"conversationType,omitempty"`
	VendingMachineId  string           `json:"vendingMachineId,omitempty"`
	LayoutId          int              `json:"layoutId"`
	LayoutPrefix      string           `json:"layoutPrefix"`
	ReferenceImageUrl string           `json:"referenceImageUrl"`
	CheckingImageUrl  string           `json:"checkingImageUrl"`
	TurnConfig        *TurnConfig      `json:"turnConfig,omitempty"`
	TurnTimestamps    *TurnTimestamps  `json:"turnTimestamps,omitempty"`
	RequestMetadata   *RequestMetadata `json:"requestMetadata,omitempty"`
}

// -------------------------
// Conversation Models
// -------------------------

// ConversationConfig defines the conversation structure for verification
type ConversationConfig struct {
	Type     string `json:"type"`
	MaxTurns int    `json:"maxTurns"`
}

// TurnConfig defines the turn structure for the verification process
type TurnConfig struct {
	MaxTurns           int `json:"maxTurns"`
	ReferenceImageTurn int `json:"referenceImageTurn"`
	CheckingImageTurn  int `json:"checkingImageTurn"`
}

// TurnTimestamps tracks the timestamps of each turn in the verification process
type TurnTimestamps struct {
	Initialized string  `json:"initialized"`
	Turn1       *string `json:"turn1"`
	Turn2       *string `json:"turn2"`
	Completed   *string `json:"completed"`
}

// RequestMetadata stores metadata about the original request
type RequestMetadata struct {
	RequestId         string `json:"requestId"`
	RequestTimestamp  string `json:"requestTimestamp"`
	ProcessingStarted string `json:"processingStarted"`
}

// -------------------------
// Storage Models
// -------------------------

// DynamoDBVerificationItem represents the DynamoDB item structure
type DynamoDBVerificationItem struct {
	VerificationId    string `dynamodbav:"verificationId"`
	VerificationAt    string `dynamodbav:"verificationAt"`
	Status            string `dynamodbav:"status"`
	VendingMachineId  string `dynamodbav:"vendingMachineId,omitempty"`
	LayoutId          int    `dynamodbav:"layoutId"`
	LayoutPrefix      string `dynamodbav:"layoutPrefix"`
	ReferenceImageUrl string `dynamodbav:"referenceImageUrl"`
	CheckingImageUrl  string `dynamodbav:"checkingImageUrl"`
	RequestId         string `dynamodbav:"requestId,omitempty"`
	TTL               int64  `dynamodbav:"ttl,omitempty"`
}

// -------------------------
// Utility Models
// -------------------------

// S3URL represents a parsed S3 URL with bucket and key components
type S3URL struct {
	Bucket string
	Key    string
}

// Constants for verification status
const (
	StatusInitialized  = "INITIALIZED"
	StatusProcessing   = "PROCESSING"
	StatusImagesFetched = "IMAGES_FETCHED"
	StatusCompleted    = "COMPLETED"
	StatusFailed       = "FAILED"
)

// -------------------------
// Helper Functions
// -------------------------

// FormatISO8601 formats the current time as an ISO8601 string
func FormatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GetCurrentTimestamp returns the current time formatted for use in IDs
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format("20060102150405")
}