// Package schema provides standardized types and constants for the verification workflow
package schema

import "time"

// SchemaVersion is the current version of the schema
const SchemaVersion = "1.0.0"

// Verification types
const (
	VerificationTypeLayoutVsChecking  = "LAYOUT_VS_CHECKING"
	VerificationTypePreviousVsCurrent = "PREVIOUS_VS_CURRENT"
)

// Status constants aligned with state machine
const (
	StatusVerificationRequested  = "VERIFICATION_REQUESTED"
	StatusVerificationInitialized = "VERIFICATION_INITIALIZED"
	StatusFetchingImages         = "FETCHING_IMAGES"
	StatusImagesFetched          = "IMAGES_FETCHED"
	StatusPromptPrepared         = "PROMPT_PREPARED"
	StatusTurn1PromptReady       = "TURN1_PROMPT_READY"
	StatusTurn1Completed         = "TURN1_COMPLETED"
	StatusTurn1Processed         = "TURN1_PROCESSED"
	StatusTurn2PromptReady       = "TURN2_PROMPT_READY"
	StatusTurn2Completed         = "TURN2_COMPLETED"
	StatusTurn2Processed         = "TURN2_PROCESSED"
	StatusResultsFinalized       = "RESULTS_FINALIZED"
	StatusResultsStored          = "RESULTS_STORED"
	StatusNotificationSent       = "NOTIFICATION_SENT"
	StatusCompleted              = "COMPLETED"

	// Error states
	StatusInitializationFailed    = "INITIALIZATION_FAILED"
	StatusHistoricalFetchFailed   = "HISTORICAL_FETCH_FAILED"
	StatusImageFetchFailed        = "IMAGE_FETCH_FAILED"
	StatusBedrockProcessingFailed = "BEDROCK_PROCESSING_FAILED"
	StatusVerificationFailed      = "VERIFICATION_FAILED"
)

// VerificationContext represents the core context that flows through the Step Functions
type VerificationContext struct {
	VerificationId         string            `json:"verificationId"`
	VerificationAt         string            `json:"verificationAt"`
	Status                 string            `json:"status"`
	VerificationType       string            `json:"verificationType"`
	ConversationType       string            `json:"conversationType,omitempty"`
	VendingMachineId       string            `json:"vendingMachineId,omitempty"`
	LayoutId               int               `json:"layoutId,omitempty"`
	LayoutPrefix           string            `json:"layoutPrefix,omitempty"`
	PreviousVerificationId string            `json:"previousVerificationId,omitempty"`
	ReferenceImageUrl      string            `json:"referenceImageUrl"`
	CheckingImageUrl       string            `json:"checkingImageUrl"`
	TurnConfig             *TurnConfig       `json:"turnConfig,omitempty"`
	TurnTimestamps         *TurnTimestamps   `json:"turnTimestamps,omitempty"`
	RequestMetadata        *RequestMetadata  `json:"requestMetadata,omitempty"`
	ResourceValidation     *ResourceValidation `json:"resourceValidation,omitempty"`
	NotificationEnabled    bool              `json:"notificationEnabled"`
	Error                  *ErrorInfo        `json:"error,omitempty"`
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

// ImageData represents the standardized structure for image references
type ImageData struct {
	Reference *ImageInfo `json:"reference"`
	Checking  *ImageInfo `json:"checking"`
}

// ImageInfo contains details about a single image
type ImageInfo struct {
	URL       string                 `json:"url"`
	S3Key     string                 `json:"s3Key"`
	S3Bucket  string                 `json:"s3Bucket"`
	Width     int                    `json:"width,omitempty"`
	Height    int                    `json:"height,omitempty"`
	Format    string                 `json:"format,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationState tracks the state of the conversation with Bedrock
type ConversationState struct {
	CurrentTurn       int                      `json:"currentTurn"`
	MaxTurns          int                      `json:"maxTurns"`
	History           []interface{}            `json:"history"`
	ReferenceAnalysis map[string]interface{}   `json:"referenceAnalysis,omitempty"`
	CheckingAnalysis  map[string]interface{}   `json:"checkingAnalysis,omitempty"`
}

// SystemPrompt contains the system prompt configuration
// SystemPrompt contains the system prompt configuration
type SystemPrompt struct {
	SystemPrompt  string        `json:"systemPrompt"`
	BedrockConfig *BedrockConfig `json:"bedrockConfig"`
	PromptId      string        `json:"promptId,omitempty"`
	PromptVersion string        `json:"promptVersion,omitempty"`
}

// CurrentPrompt contains the user prompt for a specific turn
type CurrentPrompt struct {
	CurrentPrompt string                 `json:"currentPrompt"`
	TurnNumber    int                    `json:"turnNumber"`
	IncludeImage  string                 `json:"includeImage"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BedrockConfig contains configuration for Bedrock API calls
type BedrockConfig struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	Thinking         *Thinking `json:"thinking,omitempty"`
}

// Thinking configures the thinking feature
type Thinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// FinalResults represents the final verification results
type FinalResults struct {
	VerificationStatus  string                  `json:"verificationStatus"`
	ConfidenceScore     float64                 `json:"confidenceScore"`
	DiscrepanciesCount  int                     `json:"discrepanciesCount"`
	Discrepancies       []Discrepancy           `json:"discrepancies,omitempty"`
	ResultImageUrl      string                  `json:"resultImageUrl,omitempty"`
	Summary             string                  `json:"summary,omitempty"`
	ComparisonDetails   map[string]interface{}  `json:"comparisonDetails,omitempty"`
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
	NotificationResult  map[string]interface{} `json:"notificationResult,omitempty"`
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