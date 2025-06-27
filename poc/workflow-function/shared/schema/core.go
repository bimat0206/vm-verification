// Package schema provides standardized types and constants for the verification workflow
package schema

import (
	"time"
)

// VerificationContext represents the core context that flows through the Step Functions
type VerificationContext struct {
	VerificationId         string              `json:"verificationId"`
	VerificationAt         string              `json:"verificationAt"`
	Status                 string              `json:"status"`
	VerificationType       string              `json:"verificationType"`
	ConversationType       string              `json:"conversationType,omitempty"`
	VendingMachineId       string              `json:"vendingMachineId,omitempty"`
	LayoutId               int                 `json:"layoutId,omitempty"`
	LayoutPrefix           string              `json:"layoutPrefix,omitempty"`
	PreviousVerificationId string              `json:"previousVerificationId,omitempty"`
	ReferenceImageUrl      string              `json:"referenceImageUrl"`
	CheckingImageUrl       string              `json:"checkingImageUrl"`
	TurnConfig             *TurnConfig         `json:"turnConfig,omitempty"`
	TurnTimestamps         *TurnTimestamps     `json:"turnTimestamps,omitempty"`
	RequestMetadata        *RequestMetadata    `json:"requestMetadata,omitempty"`
	ResourceValidation     *ResourceValidation `json:"resourceValidation,omitempty"`
	Error                  *ErrorInfo          `json:"error,omitempty"`

	// Enhanced fields for combined function operations
	CurrentStatus     string               `json:"currentStatus,omitempty"`
	LastUpdatedAt     string               `json:"lastUpdatedAt,omitempty"`
	StatusHistory     []StatusHistoryEntry `json:"statusHistory,omitempty"`
	ProcessingMetrics *ProcessingMetrics   `json:"processingMetrics,omitempty"`
	ErrorTracking     *ErrorTracking       `json:"errorTracking,omitempty"`
}

// ErrorInfo provides standardized error reporting
type ErrorInfo struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// TurnConfig defines the conversation turn structure
type TurnConfig struct {
	MaxTurns           int `json:"maxTurns"`
	ReferenceImageTurn int `json:"referenceImageTurn"`
	CheckingImageTurn  int `json:"checkingImageTurn"`
}

// TurnTimestamps tracks timing of each step
type TurnTimestamps struct {
	Initialized    string `json:"initialized"`
	ImagesFetched  string `json:"imagesFetched,omitempty"`
	Turn1Started   string `json:"turn1Started,omitempty"`
	Turn1Completed string `json:"turn1Completed,omitempty"`
	Turn2Started   string `json:"turn2Started,omitempty"`
	Turn2Completed string `json:"turn2Completed,omitempty"`
	Completed      string `json:"completed,omitempty"`
}

// RequestMetadata contains information about the original request
type RequestMetadata struct {
	RequestId         string `json:"requestId"`
	RequestTimestamp  string `json:"requestTimestamp"`
	ProcessingStarted string `json:"processingStarted"`
}

// ResourceValidation tracks resource validation results
type ResourceValidation struct {
	LayoutExists         bool   `json:"layoutExists,omitempty"`
	ReferenceImageExists bool   `json:"referenceImageExists"`
	CheckingImageExists  bool   `json:"checkingImageExists"`
	ValidationTimestamp  string `json:"validationTimestamp"`
}

// S3StorageConfig contains configuration for S3 Base64 storage
type S3StorageConfig struct {
	TempBase64Bucket       string `json:"tempBase64Bucket"`
	Base64RetrievalTimeout int    `json:"base64RetrievalTimeout"`      // in milliseconds
	StateBucketPrefix      string `json:"stateBucketPrefix,omitempty"` // for date-based path structure
}

// ImageData is defined in image_data.go

// ImageInfo is defined in image_info.go

// ConversationState tracks the state of the conversation with Bedrock
type ConversationState struct {
	CurrentTurn       int                    `json:"currentTurn"`
	MaxTurns          int                    `json:"maxTurns"`
	History           []interface{}          `json:"history"`
	ReferenceAnalysis map[string]interface{} `json:"referenceAnalysis,omitempty"`
	CheckingAnalysis  map[string]interface{} `json:"checkingAnalysis,omitempty"`
}

// Bedrock-related types are defined in bedrock.go

// FinalResults represents the final verification results
type FinalResults struct {
	VerificationStatus string                 `json:"verificationStatus"`
	ConfidenceScore    float64                `json:"confidenceScore"`
	DiscrepanciesCount int                    `json:"discrepanciesCount"`
	Discrepancies      []Discrepancy          `json:"discrepancies,omitempty"`
	ResultImageUrl     string                 `json:"resultImageUrl,omitempty"`
	Summary            string                 `json:"summary,omitempty"`
	ComparisonDetails  map[string]interface{} `json:"comparisonDetails,omitempty"`
}

// Discrepancy represents a single discrepancy found during verification
type Discrepancy struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Location    map[string]interface{} `json:"location,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
}

// WorkflowState represents the complete state of the verification workflow
type WorkflowState struct {
	SchemaVersion       string                 `json:"schemaVersion"`
	VerificationContext *VerificationContext   `json:"verificationContext"`
	Images              *ImageData             `json:"images,omitempty"`
	SystemPrompt        *SystemPrompt          `json:"systemPrompt,omitempty"`
	BedrockConfig       *BedrockConfig         `json:"bedrockConfig,omitempty"`
	CurrentPrompt       *CurrentPrompt         `json:"currentPrompt,omitempty"`
	ConversationState   *ConversationState     `json:"conversationState,omitempty"`
	HistoricalContext   map[string]interface{} `json:"historicalContext,omitempty"`
	LayoutMetadata      map[string]interface{} `json:"layoutMetadata,omitempty"`
	Turn1Response       map[string]interface{} `json:"turn1Response,omitempty"`
	Turn2Response       map[string]interface{} `json:"turn2Response,omitempty"`
	ReferenceAnalysis   map[string]interface{} `json:"referenceAnalysis,omitempty"`
	CheckingAnalysis    map[string]interface{} `json:"checkingAnalysis,omitempty"`
	FinalResults        *FinalResults          `json:"finalResults,omitempty"`
	StorageResult       map[string]interface{} `json:"storageResult,omitempty"`
	Error               *ErrorInfo             `json:"error,omitempty"`
}

// -------------------------
// Helper Functions
// -------------------------

// FormatISO8601 returns now in RFC3339
func FormatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GetCurrentTimestamp returns time formatted for IDs
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format("20060102150405")
}

// LayoutMetadata represents layout metadata from DynamoDB
type LayoutMetadata struct {
	LayoutId           int                    `json:"layoutId" dynamodbav:"layoutId"`
	LayoutPrefix       string                 `json:"layoutPrefix" dynamodbav:"layoutPrefix"`
	VendingMachineId   string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	Location           string                 `json:"location" dynamodbav:"location"`
	CreatedAt          string                 `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt          string                 `json:"updatedAt" dynamodbav:"updatedAt"`
	ReferenceImageUrl  string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	SourceJsonUrl      string                 `json:"sourceJsonUrl" dynamodbav:"sourceJsonUrl"`
	MachineStructure   map[string]interface{} `json:"machineStructure" dynamodbav:"machineStructure"`
	ProductPositionMap map[string]interface{} `json:"productPositionMap" dynamodbav:"productPositionMap"`
}

// LayoutKey represents a composite key for layout metadata
type LayoutKey struct {
	LayoutId     int    `json:"layoutId"`
	LayoutPrefix string `json:"layoutPrefix"`
}

// ImageInfo methods are defined in image_info.go

// ImageData methods are defined in image_data.go

// Bedrock message builder functions are defined in bedrock.go
// Add these new structures to support combined function operations

// StatusHistoryEntry represents a single status transition
type StatusHistoryEntry struct {
	Status           string                 `json:"status"`
	Timestamp        string                 `json:"timestamp"`
	FunctionName     string                 `json:"functionName"`
	ProcessingTimeMs int64                  `json:"processingTimeMs"`
	Stage            string                 `json:"stage"`
	Metrics          map[string]interface{} `json:"metrics,omitempty"`
}

// ProcessingMetrics tracks performance across turns
type ProcessingMetrics struct {
	WorkflowTotal *WorkflowMetrics `json:"workflowTotal,omitempty"`
	Turn1         *TurnMetrics     `json:"turn1,omitempty"`
	Turn2         *TurnMetrics     `json:"turn2,omitempty"`
}

// WorkflowMetrics represents overall workflow performance
type WorkflowMetrics struct {
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	TotalTimeMs   int64  `json:"totalTimeMs"`
	FunctionCount int    `json:"functionCount"`
}

// TurnMetrics represents individual turn performance
type TurnMetrics struct {
	StartTime        string      `json:"startTime"`
	EndTime          string      `json:"endTime"`
	TotalTimeMs      int64       `json:"totalTimeMs"`
	BedrockLatencyMs int64       `json:"bedrockLatencyMs"`
	ProcessingTimeMs int64       `json:"processingTimeMs"`
	RetryAttempts    int         `json:"retryAttempts"`
	TokenUsage       *TokenUsage `json:"tokenUsage,omitempty"`
}

// ErrorTracking manages error state throughout workflow
type ErrorTracking struct {
	HasErrors        bool        `json:"hasErrors"`
	CurrentError     *ErrorInfo  `json:"currentError"`
	ErrorHistory     []ErrorInfo `json:"errorHistory"`
	RecoveryAttempts int         `json:"recoveryAttempts"`
	LastErrorAt      string      `json:"lastErrorAt"`
}

// Enhanced VerificationContext now includes fields for combined function operations:
// - CurrentStatus: Current processing status
// - LastUpdatedAt: Last update timestamp
// - StatusHistory: Complete status transition history
// - ProcessingMetrics: Performance metrics tracking
// - ErrorTracking: Error state management

// ADD: Complete DynamoDB table structures
type ConversationHistory struct {
	VerificationId   string                 `json:"verificationId" dynamodbav:"verificationId"`
	ConversationAt   string                 `json:"conversationAt" dynamodbav:"conversationAt"`
	VendingMachineId string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	LayoutId         int                    `json:"layoutId" dynamodbav:"layoutId"`
	LayoutPrefix     string                 `json:"layoutPrefix" dynamodbav:"layoutPrefix"`
	CurrentTurn      int                    `json:"currentTurn" dynamodbav:"currentTurn"`
	MaxTurns         int                    `json:"maxTurns" dynamodbav:"maxTurns"`
	TurnStatus       string                 `json:"turnStatus" dynamodbav:"turnStatus"`
	History          []TurnHistory          `json:"history" dynamodbav:"history"`
	ExpiresAt        int64                  `json:"expiresAt,omitempty" dynamodbav:"expiresAt,omitempty"`
	Metadata         map[string]interface{} `json:"metadata" dynamodbav:"metadata"`
}

type VerificationResults struct {
	VerificationId            string                 `json:"verificationId" dynamodbav:"verificationId"`
	VerificationAt            string                 `json:"verificationAt" dynamodbav:"verificationAt"`
	VerificationType          string                 `json:"verificationType" dynamodbav:"verificationType"`
	PreviousVerificationId    string                 `json:"previousVerificationId,omitempty" dynamodbav:"previousVerificationId"`
	LayoutId                  int                    `json:"layoutId,omitempty" dynamodbav:"layoutId"`
	LayoutPrefix              string                 `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix"`
	VendingMachineId          string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	Location                  string                 `json:"location" dynamodbav:"location"`
	ReferenceImageUrl         string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	CheckingImageUrl          string                 `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
	VerificationStatus        string                 `json:"verificationStatus" dynamodbav:"verificationStatus"`
	MachineStructure          map[string]interface{} `json:"machineStructure" dynamodbav:"machineStructure"`
	InitialConfirmation       string                 `json:"initialConfirmation" dynamodbav:"initialConfirmation"`
	ProcessedTurn1MarkdownRef S3Reference            `json:"processedTurn1MarkdownRef,omitempty" dynamodbav:"processedTurn1MarkdownRef,omitempty"`
	CorrectedRows             []string               `json:"correctedRows" dynamodbav:"correctedRows"`
	EmptySlotReport           map[string]interface{} `json:"emptySlotReport" dynamodbav:"emptySlotReport"`
	ReferenceStatus           map[string]string      `json:"referenceStatus" dynamodbav:"referenceStatus"`
	CheckingStatus            map[string]string      `json:"checkingStatus" dynamodbav:"checkingStatus"`
	Discrepancies             []Discrepancy          `json:"discrepancies" dynamodbav:"discrepancies"`
	VerificationSummary       map[string]interface{} `json:"verificationSummary" dynamodbav:"verificationSummary"`
	Metadata                  map[string]interface{} `json:"metadata" dynamodbav:"metadata"`
}



// ADD: Enhanced error handling
type WorkflowError struct {
	Stage       string                 `json:"stage"`
	Function    string                 `json:"function"`
	ErrorType   string                 `json:"errorType"`
	Recoverable bool                   `json:"recoverable"`
	RetryCount  int                    `json:"retryCount"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}
