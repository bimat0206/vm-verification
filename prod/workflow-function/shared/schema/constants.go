// Package schema provides standardized types and constants for the verification workflow
package schema

// SchemaVersion is the current version of the schema
const SchemaVersion = "2.1.0" // Updated to use InputTokens/OutputTokens instead of TokenEstimate

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
	StatusVerificationRequested   = "VERIFICATION_REQUESTED"
	StatusVerificationInitialized = "VERIFICATION_INITIALIZED"
	StatusFetchingImages          = "FETCHING_IMAGES"
	StatusImagesFetched           = "IMAGES_FETCHED"
	StatusPromptPrepared          = "PROMPT_PREPARED"
	StatusHistoricalContextLoaded = "HISTORICAL_CONTEXT_LOADED"
	StatusHistoricalContextNotFound = "HISTORICAL_CONTEXT_NOT_FOUND"
	StatusTurn1PromptReady        = "TURN1_PROMPT_READY"
	StatusTurn1Completed          = "TURN1_COMPLETED"
	StatusTurn1Processed          = "TURN1_PROCESSED"
	StatusTurn2PromptReady        = "TURN2_PROMPT_READY"
	StatusTurn2Completed          = "TURN2_COMPLETED"
	StatusTurn2Processed          = "TURN2_PROCESSED"
	StatusResultsFinalized        = "RESULTS_FINALIZED"
	StatusResultsStored           = "RESULTS_STORED"
	StatusCompleted               = "COMPLETED"

	// Error states
	StatusInitializationFailed    = "INITIALIZATION_FAILED"
	StatusHistoricalFetchFailed   = "HISTORICAL_FETCH_FAILED"
	StatusImageFetchFailed        = "IMAGE_FETCH_FAILED"
	StatusBedrockProcessingFailed = "BEDROCK_PROCESSING_FAILED"
	StatusVerificationFailed      = "VERIFICATION_FAILED"
)

// Add these status constants for combined function tracking
const (
	// Existing constants remain...

	// Turn 1 Combined Function Status (ADD THESE)
	StatusTurn1Started            = "TURN1_STARTED"
	StatusTurn1ContextLoaded      = "TURN1_CONTEXT_LOADED"
	StatusTurn1PromptPrepared     = "TURN1_PROMPT_PREPARED"
	StatusTurn1ImageLoaded        = "TURN1_IMAGE_LOADED"
	StatusTurn1BedrockInvoked     = "TURN1_BEDROCK_INVOKED"
	StatusTurn1BedrockCompleted   = "TURN1_BEDROCK_COMPLETED"
	StatusTurn1ResponseProcessing = "TURN1_RESPONSE_PROCESSING"

	// Turn 2 Combined Function Status (ADD THESE)
	StatusTurn2Started            = "TURN2_STARTED"
	StatusTurn2ContextLoaded      = "TURN2_CONTEXT_LOADED"
	StatusTurn2PromptPrepared     = "TURN2_PROMPT_PREPARED"
	StatusTurn2ImageLoaded        = "TURN2_IMAGE_LOADED"
	StatusTurn2BedrockInvoked     = "TURN2_BEDROCK_INVOKED"
	StatusTurn2BedrockCompleted   = "TURN2_BEDROCK_COMPLETED"
	StatusTurn2ResponseProcessing = "TURN2_RESPONSE_PROCESSING"

	// Error handling constants (ADD THESE)
	StatusTurn1Error              = "TURN1_ERROR"
	StatusTurn2Error              = "TURN2_ERROR"
	StatusTemplateProcessingError = "TEMPLATE_PROCESSING_ERROR"
)

// ADD: More specific status constants for DynamoDB operations
const (
	// Existing constants remain...

	// DynamoDB operation statuses
	StatusDynamoDBWriteStarted   = "DYNAMODB_WRITE_STARTED"
	StatusDynamoDBWriteCompleted = "DYNAMODB_WRITE_COMPLETED"
	StatusDynamoDBWriteFailed    = "DYNAMODB_WRITE_FAILED"
)

// ADD: Verification result status constants
const (
	VerificationStatusCorrect   = "CORRECT"
	VerificationStatusIncorrect = "INCORRECT"
	VerificationStatusPartial   = "PARTIAL"
	VerificationStatusFailed    = "FAILED"
)

// ADD: Turn status constants for ConversationHistory
const (
	TurnStatusActive    = "ACTIVE"
	TurnStatusCompleted = "COMPLETED"
	TurnStatusFailed    = "FAILED"
)
