// Package schema provides standardized types and constants for the verification workflow
package schema

import (
	"fmt"
	"time"
)

// SchemaVersion is the current version of the schema
const SchemaVersion = "1.2.0" // Updated for Hybrid Base64 Storage support

// Verification types
const (
	VerificationTypeLayoutVsChecking  = "LAYOUT_VS_CHECKING"
	VerificationTypePreviousVsCurrent = "PREVIOUS_VS_CURRENT"
)

// Storage method constants for hybrid Base64 storage
const (
	StorageMethodInline      = "inline"
	StorageMethodS3Temporary = "s3-temporary"
	
	// Storage thresholds and limits
	DefaultBase64SizeThreshold = 2 * 1024 * 1024 // 2MB threshold
	BedrockMaxImageSize       = 20 * 1024 * 1024 // 20MB Bedrock limit
	LambdaPayloadLimit        = 6 * 1024 * 1024  // 6MB Lambda payload limit
	
	// Temporary storage configuration
	TempBase64KeyPrefix = "temp-base64/"
	TempBase64TTL       = 24 * 60 * 60 // 24 hours in seconds
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
	NotificationEnabled    bool                `json:"notificationEnabled"`
	Error                  *ErrorInfo          `json:"error,omitempty"`
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

// HybridStorageConfig contains configuration for hybrid Base64 storage
type HybridStorageConfig struct {
	TempBase64Bucket       string `json:"tempBase64Bucket"`
	Base64SizeThreshold    int64  `json:"base64SizeThreshold"`
	Base64RetrievalTimeout int    `json:"base64RetrievalTimeout"` // in milliseconds
	EnableHybridStorage    bool   `json:"enableHybridStorage"`
}

// ImageData represents the standardized structure for image references with hybrid Base64 storage
// Supports both new format (Reference/Checking) and legacy format (ReferenceImage/CheckingImage)
type ImageData struct {
	// New format (primary)
	Reference *ImageInfo `json:"reference,omitempty"`
	Checking  *ImageInfo `json:"checking,omitempty"`
	
	// Legacy format support for backward compatibility
	ReferenceImage *ImageInfo `json:"referenceImage,omitempty"`
	CheckingImage  *ImageInfo `json:"checkingImage,omitempty"`
	
	// Processing metadata
	ProcessedAt           string                `json:"processedAt,omitempty"`
	Base64Generated       bool                  `json:"base64Generated"`
	BucketOwner           string                `json:"bucketOwner,omitempty"`
	
	// Hybrid storage metadata (NEW)
	HybridStorageEnabled  bool                  `json:"hybridStorageEnabled"`
	StorageDecisionAt     string                `json:"storageDecisionAt,omitempty"`
	TotalInlineSize       int64                 `json:"totalInlineSize"`        // Total size of inline Base64 data
	TotalS3References     int                   `json:"totalS3References"`      // Count of S3-stored Base64 references
	StorageConfig         *HybridStorageConfig  `json:"storageConfig,omitempty"`
}

// ImageInfo contains details about a single image with hybrid Base64 storage support
// Supports both legacy field names and new unified format
type ImageInfo struct {
	// S3 References (for traceability and logging)
	URL      string `json:"url"`
	S3Key    string `json:"s3Key"`
	S3Bucket string `json:"s3Bucket"`
	
	// Image Properties
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Format      string `json:"format,omitempty"`           // png, jpeg, etc.
	ContentType string `json:"contentType,omitempty"`      // image/png, image/jpeg
	Size        int64  `json:"size,omitempty"`             // Original file size in bytes
	
	// Hybrid Base64 Storage (ENHANCED) - supports both legacy and new field names
	Base64Data           string `json:"base64Data,omitempty"`           // Unified Base64 field
	
	// Legacy field names for backward compatibility with specific naming
	ReferenceImageBase64 string `json:"referenceImageBase64,omitempty"` // For reference images
	CheckingImageBase64  string `json:"checkingImageBase64,omitempty"`  // For checking images
	
	Base64Size           int64  `json:"base64Size,omitempty"`           // Size of Base64 string
	
	// S3 Temporary Storage for Large Base64 (NEW) - supports legacy field names
	Base64S3Bucket       string `json:"base64S3Bucket,omitempty"`       // Temporary bucket for large Base64
	Base64S3Key          string `json:"base64S3Key,omitempty"`          // Temporary S3 key for Base64 data
	
	// Legacy S3 field names for specific image types
	CheckingImageBase64S3Key string `json:"checkingImageBase64S3Key,omitempty"` // Legacy field for checking image S3 key
	ReferenceImageBase64S3Key string `json:"referenceImageBase64S3Key,omitempty"` // Legacy field for reference image S3 key
	
	Base64S3Metadata     map[string]string `json:"base64S3Metadata,omitempty"` // Additional S3 metadata
	
	// Storage Method Indicators (NEW)
	StorageMethod        string `json:"storageMethod"`                  // "inline" or "s3-temporary"
	Base64Generated      bool   `json:"base64Generated"`                // Indicates Base64 conversion completed
	StorageDecisionAt    string `json:"storageDecisionAt,omitempty"`    // When storage method was decided
	
	// Retrieval metadata (NEW)
	LastBase64Access     string `json:"lastBase64Access,omitempty"`     // Last time Base64 was accessed
	Base64AccessCount    int    `json:"base64AccessCount,omitempty"`    // Number of times Base64 was accessed
	
	// Existing Metadata
	LastModified string                 `json:"lastModified,omitempty"`
	ETag         string                 `json:"etag,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationState tracks the state of the conversation with Bedrock
type ConversationState struct {
	CurrentTurn       int                    `json:"currentTurn"`
	MaxTurns          int                    `json:"maxTurns"`
	History           []interface{}          `json:"history"`
	ReferenceAnalysis map[string]interface{} `json:"referenceAnalysis,omitempty"`
	CheckingAnalysis  map[string]interface{} `json:"checkingAnalysis,omitempty"`
}

// SystemPrompt contains the system prompt configuration
type SystemPrompt struct {
	Content       string        `json:"content"`
	BedrockConfig *BedrockConfig `json:"-"` // Marked as ignored in JSON as it will be at top level
	PromptId      string        `json:"promptId,omitempty"`
	PromptVersion string        `json:"promptVersion,omitempty"`
}

// CurrentPrompt contains the user prompt for a specific turn with Bedrock message support
type CurrentPrompt struct {
	// Text prompt (backward compatibility)
	Text       string `json:"text,omitempty"`
	TurnNumber int    `json:"turnNumber"`
	
	// Image configuration
	IncludeImage string `json:"includeImage"`  // "reference", "checking", "both", "none"
	
	// Bedrock-formatted messages (with Base64 images)
	Messages []BedrockMessage `json:"messages,omitempty"`
	
	// Metadata
	PromptId      string                 `json:"promptId,omitempty"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	PromptVersion string                 `json:"promptVersion,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BedrockMessage represents a message in Bedrock API format
type BedrockMessage struct {
	Role    string           `json:"role"`    // "user", "assistant"
	Content []BedrockContent `json:"content"`
}

// BedrockContent represents content within a Bedrock message
type BedrockContent struct {
	Type  string            `json:"type"`            // "text" or "image"
	Text  string            `json:"text,omitempty"`  // For text content
	Image *BedrockImageData `json:"image,omitempty"` // For image content
}

// BedrockImageData represents image data for Bedrock API
type BedrockImageData struct {
	Format string             `json:"format"`         // "png", "jpeg", etc.
	Source BedrockImageSource `json:"source"`
}

// BedrockImageSource represents the source of image data for Bedrock
type BedrockImageSource struct {
	Type  string `json:"type"`           // "bytes" (S3 is deprecated for most models)
	Bytes string `json:"bytes"`          // Base64-encoded image data
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

// TurnResponse represents response from a Bedrock turn with enhanced structure
type TurnResponse struct {
	TurnId      int                    `json:"turnId"`
	Timestamp   string                 `json:"timestamp"`
	Prompt      string                 `json:"prompt"`
	ImageUrls   map[string]string      `json:"imageUrls,omitempty"`  // S3 URLs for reference/traceability
	Response    BedrockApiResponse     `json:"response"`
	LatencyMs   int64                  `json:"latencyMs"`
	TokenUsage  *TokenUsage            `json:"tokenUsage,omitempty"`
	Stage       string                 `json:"analysisStage"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BedrockApiResponse represents the full response from Bedrock API
type BedrockApiResponse struct {
	Content    string `json:"content"`              // Main response text
	Thinking   string `json:"thinking,omitempty"`   // Internal reasoning (if enabled)
	StopReason string `json:"stop_reason,omitempty"`
	ModelId    string `json:"modelId,omitempty"`
	RequestId  string `json:"requestId,omitempty"`
}

// FinalResults represents the final verification results
type FinalResults struct {
	VerificationStatus string                  `json:"verificationStatus"`
	ConfidenceScore    float64                 `json:"confidenceScore"`
	DiscrepanciesCount int                     `json:"discrepanciesCount"`
	Discrepancies      []Discrepancy           `json:"discrepancies,omitempty"`
	ResultImageUrl     string                  `json:"resultImageUrl,omitempty"`
	Summary            string                  `json:"summary,omitempty"`
	ComparisonDetails  map[string]interface{}  `json:"comparisonDetails,omitempty"`
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

// -------------------------
// Enhanced Base64 Image Helper Methods
// -------------------------

// EnsureBase64Generated checks if Base64 data is present and valid
func (img *ImageInfo) EnsureBase64Generated() bool {
	return img.Base64Generated && img.Format != "" && img.HasBase64Data()
}

// HasBase64Data checks if Base64 data is available (inline or S3) supporting legacy field names
func (img *ImageInfo) HasBase64Data() bool {
	// Check unified field
	if img.Base64Data != "" {
		return true
	}
	
	// Check legacy fields
	if img.ReferenceImageBase64 != "" || img.CheckingImageBase64 != "" {
		return true
	}
	
	// Check S3 temporary storage (unified and legacy fields)
	if img.Base64S3Bucket != "" && (img.Base64S3Key != "" || 
		img.CheckingImageBase64S3Key != "" || img.ReferenceImageBase64S3Key != "") {
		return true
	}
	
	return false
}

// GetBase64Data returns the Base64 data, handling legacy field names
func (img *ImageInfo) GetBase64Data() string {
	// Return unified field if available
	if img.Base64Data != "" {
		return img.Base64Data
	}
	
	// Fall back to legacy fields
	if img.ReferenceImageBase64 != "" {
		return img.ReferenceImageBase64
	}
	
	if img.CheckingImageBase64 != "" {
		return img.CheckingImageBase64
	}
	
	return ""
}

// GetBase64S3Key returns the S3 key for Base64 data, handling legacy field names
func (img *ImageInfo) GetBase64S3Key() string {
	// Return unified field if available
	if img.Base64S3Key != "" {
		return img.Base64S3Key
	}
	
	// Fall back to legacy fields
	if img.CheckingImageBase64S3Key != "" {
		return img.CheckingImageBase64S3Key
	}
	
	if img.ReferenceImageBase64S3Key != "" {
		return img.ReferenceImageBase64S3Key
	}
	
	return ""
}

// SetBase64Data sets the Base64 data using the appropriate field based on image type
func (img *ImageInfo) SetBase64Data(base64Data string, imageType string) {
	// Set unified field
	img.Base64Data = base64Data
	
	// Also set legacy field for backward compatibility
	switch imageType {
	case "reference":
		img.ReferenceImageBase64 = base64Data
	case "checking":
		img.CheckingImageBase64 = base64Data
	}
	
	// Update size
	img.Base64Size = int64(len(base64Data))
}

// SetBase64S3Key sets the S3 key for Base64 data using the appropriate field based on image type
func (img *ImageInfo) SetBase64S3Key(s3Key string, imageType string) {
	// Set unified field
	img.Base64S3Key = s3Key
	
	// Also set legacy field for backward compatibility
	switch imageType {
	case "reference":
		img.ReferenceImageBase64S3Key = s3Key
	case "checking":
		img.CheckingImageBase64S3Key = s3Key
	}
}

// GetBase64SizeEstimate estimates the original image size from Base64
func (img *ImageInfo) GetBase64SizeEstimate() int64 {
	if img.Base64Size > 0 {
		return img.Base64Size
	}
	
	// Calculate from available Base64 data
	base64Data := img.GetBase64Data()
	if base64Data != "" {
		return int64(len(base64Data))
	}
	
	// Base64 encoding increases size by ~33%
	return img.Size * 4 / 3
}

// ValidateBase64Size checks if Base64 data is within reasonable limits
func (img *ImageInfo) ValidateBase64Size() error {
	base64Size := img.GetBase64SizeEstimate()
	if base64Size > BedrockMaxImageSize {
		return fmt.Errorf("Base64 data size (%d bytes) exceeds Bedrock limit (%d bytes)", 
			base64Size, BedrockMaxImageSize)
	}
	return nil
}

// IsInlineStorage returns true if image uses inline storage
func (img *ImageInfo) IsInlineStorage() bool {
	return img.StorageMethod == StorageMethodInline
}

// IsS3TemporaryStorage returns true if image uses S3 temporary storage
func (img *ImageInfo) IsS3TemporaryStorage() bool {
	return img.StorageMethod == StorageMethodS3Temporary
}

// GetStorageInfo returns storage method information
func (img *ImageInfo) GetStorageInfo() map[string]interface{} {
	info := map[string]interface{}{
		"storageMethod":    img.StorageMethod,
		"base64Generated":  img.Base64Generated,
		"base64Size":       img.GetBase64SizeEstimate(),
		"hasBase64Data":    img.HasBase64Data(),
	}
	
	// Legacy field information
	if img.ReferenceImageBase64 != "" {
		info["hasReferenceImageBase64"] = true
		info["referenceImageBase64Size"] = len(img.ReferenceImageBase64)
	}
	if img.CheckingImageBase64 != "" {
		info["hasCheckingImageBase64"] = true
		info["checkingImageBase64Size"] = len(img.CheckingImageBase64)
	}
	
	// S3 storage information (supporting legacy fields)
	if img.IsS3TemporaryStorage() {
		info["base64S3Bucket"] = img.Base64S3Bucket
		info["base64S3Key"] = img.GetBase64S3Key() // Uses helper that handles legacy fields
		info["base64S3Metadata"] = img.Base64S3Metadata
		
		// Legacy S3 field information
		if img.CheckingImageBase64S3Key != "" {
			info["checkingImageBase64S3Key"] = img.CheckingImageBase64S3Key
		}
		if img.ReferenceImageBase64S3Key != "" {
			info["referenceImageBase64S3Key"] = img.ReferenceImageBase64S3Key
		}
	}
	
	if img.StorageDecisionAt != "" {
		info["storageDecisionAt"] = img.StorageDecisionAt
	}
	
	return info
}

// UpdateLastAccess updates the last access information for Base64 data
func (img *ImageInfo) UpdateLastAccess() {
	img.LastBase64Access = FormatISO8601()
	img.Base64AccessCount++
}

// -------------------------
// Image Data Helper Methods
// -------------------------

// GetReference returns the reference image, supporting both formats
func (images *ImageData) GetReference() *ImageInfo {
	if images.Reference != nil {
		return images.Reference
	}
	return images.ReferenceImage
}

// GetChecking returns the checking image, supporting both formats
func (images *ImageData) GetChecking() *ImageInfo {
	if images.Checking != nil {
		return images.Checking
	}
	return images.CheckingImage
}

// SetReference sets the reference image in both formats for compatibility
func (images *ImageData) SetReference(img *ImageInfo) {
	images.Reference = img
	images.ReferenceImage = img
}

// SetChecking sets the checking image in both formats for compatibility
func (images *ImageData) SetChecking(img *ImageInfo) {
	images.Checking = img
	images.CheckingImage = img
}

// GetTotalBase64Size returns the total size of all Base64 data
func (images *ImageData) GetTotalBase64Size() int64 {
	var total int64
	
	ref := images.GetReference()
	if ref != nil {
		total += ref.GetBase64SizeEstimate()
	}
	
	checking := images.GetChecking()
	if checking != nil {
		total += checking.GetBase64SizeEstimate()
	}
	
	return total
}

// ValidateForPayloadLimits checks if the total size is within Lambda payload limits
func (images *ImageData) ValidateForPayloadLimits() error {
	if !images.HybridStorageEnabled {
		// For non-hybrid storage, check total payload size
		totalSize := images.GetTotalBase64Size()
		if totalSize > LambdaPayloadLimit {
			return fmt.Errorf("total Base64 size (%d bytes) exceeds Lambda payload limit (%d bytes)",
				totalSize, LambdaPayloadLimit)
		}
	}
	return nil
}

// GetStorageSummary returns a summary of storage methods used
func (images *ImageData) GetStorageSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"hybridStorageEnabled": images.HybridStorageEnabled,
		"base64Generated":      images.Base64Generated,
		"totalInlineSize":      images.TotalInlineSize,
		"totalS3References":    images.TotalS3References,
		"processedAt":          images.ProcessedAt,
	}
	
	ref := images.GetReference()
	if ref != nil {
		summary["referenceStorage"] = ref.GetStorageInfo()
	}
	
	checking := images.GetChecking()
	if checking != nil {
		summary["checkingStorage"] = checking.GetStorageInfo()
	}
	
	return summary
}

// -------------------------
// Bedrock Message Builder Functions
// -------------------------

// BuildBedrockMessage creates a Bedrock message with text and optional image
// Supports both legacy field names and unified format
func BuildBedrockMessage(text string, image *ImageInfo) BedrockMessage {
	content := []BedrockContent{
		{
			Type: "text",
			Text: text,
		},
	}
	
	// Add image if Base64 data is available
	if image != nil && image.EnsureBase64Generated() {
		// Get Base64 data using the helper method that handles legacy fields
		base64Data := image.GetBase64Data()
		if base64Data != "" {
			imageContent := BedrockContent{
				Type: "image",
				Image: &BedrockImageData{
					Format: image.Format,
					Source: BedrockImageSource{
						Type:  "bytes",
						Bytes: base64Data,
					},
				},
			}
			content = append(content, imageContent)
		}
	}
	
	return BedrockMessage{
		Role:    "user",
		Content: content,
	}
}

// BuildBedrockMessages creates the complete messages array for Bedrock API
// Supports both legacy and new field formats
func BuildBedrockMessages(prompt string, includeImage string, images *ImageData) []BedrockMessage {
	var messages []BedrockMessage
	
	var imageToInclude *ImageInfo
	switch includeImage {
	case "reference":
		if images != nil {
			imageToInclude = images.GetReference() // Uses helper that handles both formats
		}
	case "checking":
		if images != nil {
			imageToInclude = images.GetChecking() // Uses helper that handles both formats
		}
	}
	
	message := BuildBedrockMessage(prompt, imageToInclude)
	messages = append(messages, message)
	
	return messages
}