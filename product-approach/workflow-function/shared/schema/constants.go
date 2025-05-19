// Package schema provides standardized types and constants for the verification workflow
package schema

// SchemaVersion is the current version of the schema
const SchemaVersion = "2.0.0" // Updated to align with JSON schema v2.0.0

// Verification types
const (
	VerificationTypeLayoutVsChecking  = "LAYOUT_VS_CHECKING"
	VerificationTypePreviousVsCurrent = "PREVIOUS_VS_CURRENT"
)

// Storage method constants for S3 Base64 storage
const (
	StorageMethodS3Temporary = "s3-temporary"
	
	// Storage limits
	BedrockMaxImageSize = 20 * 1024 * 1024 // 20MB Bedrock limit
	
	// S3 storage configuration
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
