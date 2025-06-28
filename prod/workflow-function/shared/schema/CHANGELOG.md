# Changelog

## [2.3.2] - 2025-06-28

### Changed
- **BREAKING**: Removed `ConversationId` field from `ConversationTracker` struct
  - **Rationale**: Field was redundant and caused DynamoDB mapping issues since `verificationId` is passed as parameter
  - **Impact**: Eliminates confusion between struct field names and DynamoDB table schema
  - **Migration**: All consumers now pass `verificationID` explicitly as function parameter

### Enhanced
- **DynamoDB Attribute Mapping**: Added proper `dynamodbav` tags to remaining ConversationTracker fields
  - `CurrentTurn`, `MaxTurns`, `TurnStatus`, `ConversationAt`, etc. now have correct DynamoDB mapping
  - Improved marshalling/unmarshalling consistency with DynamoDB operations
- **Validation Cleanup**: Removed redundant ConversationId validation since verification ID is validated at function level

### Technical Details
- **Files Modified**:
  - `types.go` - Removed ConversationId field from ConversationTracker struct
  - `validation.go` - Removed ConversationId validation logic
- **Field Mapping**: ConversationTracker fields now properly map to DynamoDB table schema without intermediate struct fields
- **Backward Compatibility**: Breaking change - consumers must update to pass verificationID as parameter

### Impact
- ✅ Eliminates DynamoDB ValidationException errors for empty key attributes
- ✅ Simplifies data model by removing redundant field
- ✅ Makes verification ID handling explicit and clear
- ✅ Improves code maintainability across all consumers

## [2.3.1] - 2025-06-28

### Added
- Added `StatusHistoricalContextNotFound` constant to handle cases where no historical verification data exists
- This new status is used when FetchHistoricalVerification function creates a fallback context for fresh verifications

### Fixed
- Fixed logical contradiction where `StatusHistoricalContextLoaded` was set even when `historicalDataFound` was false
- Improved workflow state consistency between FetchHistoricalVerification and FetchImages functions

## [2.3.0] - 2025-01-03

### Added
#### Thinking Content Support
- Added `ThinkingTokens` field to `TokenUsage` structure for Claude thinking token tracking
- Enhanced token usage calculation to include thinking tokens in total token counts
- Added thinking content support infrastructure for future Claude reasoning features

### Enhanced
#### Token Usage Tracking
- Extended `TokenUsage` structure with thinking token support
- Maintained backward compatibility with existing token usage implementations
- Added proper JSON serialization for thinking tokens with `omitempty` tag

### Technical Details
- **Breaking Change**: `TokenUsage` structure enhanced with `ThinkingTokens` field
- **Backward Compatibility**: All existing code continues to work without modification
- **Future-Proof**: Infrastructure ready for AWS SDK thinking token support
- **Schema Compliance**: All token usage tracking now includes thinking token capability

## [2.2.0] - 2025-01-22

### Added
#### Combined Function Support
- Added `CombinedTurnResponse` struct with embedded `TurnResponse` for enhanced response handling
- Added `ProcessingStage` struct for granular stage tracking within combined functions
- Added `TemplateContext` struct for template processing context
- Added fields to `CombinedTurnResponse`: `ProcessingStages`, `InternalPrompt`, `TemplateUsed`, `ContextEnrichment`

#### Template Management System
- Added `PromptTemplate` struct for template management with versioning support
- Added `TemplateProcessor` struct for template processing context and metrics
- Added `TemplateRetriever` class for S3-based template loading
- Added template processing functions in s3_helpers.go

#### Enhanced Metrics & Performance Tracking
- Added `ProcessingMetrics` struct for comprehensive workflow performance tracking
- Added `TurnMetrics` struct for individual turn performance metrics
- Added `WorkflowMetrics` struct for overall workflow timing and statistics
- Added `StatusHistoryEntry` struct for detailed status transition tracking
- Added `ErrorTracking` struct for comprehensive error state management

#### Enhanced VerificationContext
- Added `CurrentStatus` field for real-time status tracking
- Added `LastUpdatedAt` field for timestamp tracking
- Added `StatusHistory` field for complete status transition history
- Added `ProcessingMetrics` field for embedded performance metrics
- Added `ErrorTracking` field for error state management

#### Conversation Management
- Added `ConversationTracker` struct for conversation progress and state tracking
- Enhanced conversation history with metadata support

#### Enhanced Status Constants
- Added detailed Turn 1 Combined Function status constants:
  - `StatusTurn1Started`, `StatusTurn1ContextLoaded`, `StatusTurn1PromptPrepared`
  - `StatusTurn1ImageLoaded`, `StatusTurn1BedrockInvoked`, `StatusTurn1BedrockCompleted`
  - `StatusTurn1ResponseProcessing`
- Added detailed Turn 2 Combined Function status constants:
  - `StatusTurn2Started`, `StatusTurn2ContextLoaded`, `StatusTurn2PromptPrepared`
  - `StatusTurn2ImageLoaded`, `StatusTurn2BedrockInvoked`, `StatusTurn2BedrockCompleted`
  - `StatusTurn2ResponseProcessing`
- Added error handling constants:
  - `StatusTurn1Error`, `StatusTurn2Error`, `StatusTemplateProcessingError`

#### Comprehensive Validation Functions
- Added `ValidateTemplateProcessor()` for template processing validation
- Added `ValidateCombinedTurnResponse()` for combined response validation
- Added `ValidateConversationTracker()` for conversation state validation
- Added `ValidateVerificationContextEnhanced()` for enhanced context validation
- Added `ValidateStatusHistoryEntry()` for status transition validation
- Added `ValidateProcessingMetrics()` for performance metrics validation
- Added `ValidateErrorTracking()` for error state validation

### Enhanced
- Enhanced VerificationContext struct with new fields for combined function operations
- Enhanced validation coverage for all new structures
- Enhanced error handling with recovery attempt tracking
- Enhanced template retrieval with S3 integration

### Fixed
- Fixed missing VerificationContext fields that were commented out but not implemented
- Fixed validation functions to include comprehensive error checking
- Fixed status constants organization and naming consistency

### Compatibility
- Full compatibility with ExecuteTurn1Combined function
- Full compatibility with ExecuteTurn2Combined function
- Backward compatibility maintained with existing implementations
- Enhanced observability and monitoring capabilities

## [2.1.0] - 2025-05-20

### Added
- Added `CompleteSystemPrompt` struct that matches the expected JSON schema format
- Added related structs for system prompt components:
  - `PromptContent`
  - `BedrockConfiguration`
  - `ContextInformation`
  - `LayoutInformation`
  - `HistoricalContext`
  - `OutputSpecification`
  - `ProcessingMetadata`
- Added `ConvertToCompleteSystemPrompt` function to convert from simple SystemPrompt to CompleteSystemPrompt
- Added support for the complete system prompt format with all required fields

### Fixed
- Fixed issue where PrepareTurn1Prompt function was failing due to missing fields in system-prompt.json

## [2.0.1] - 2025-05-20

### Fixed
- Fixed S3 path issue where "unknown" was used as a fallback for verification ID in `generateTempS3Key` method
- Modified `generateTempS3Key` to use the provided `StateBucketPrefix` directly when available
- Changed fallback value from "unknown" to "ERROR-MISSING-VERIFICATION-ID" to make issues more visible
- Enhanced `getVerificationId` method to validate that verification IDs are not empty
- Improved error handling for missing verification IDs in S3 path generation

## [2.0.0] - 2025-05-15

### Added
- Added support for Bedrock API v2.0.0 message format
- Added `BedrockMessageBuilder` for creating Bedrock messages with S3 retrieval
- Added `S3Base64Retriever` for retrieving Base64 data from S3
- Added `S3ImageInfoBuilder` for creating image info with S3 storage
- Added `S3ImageDataProcessor` for processing image data with S3 storage

### Changed
- Updated Bedrock message format to match JSON schema v2.0.0
- Changed `BedrockImageSource.Type` from "bytes" to "base64"
- Renamed `BedrockImageSource.Bytes` to `BedrockImageSource.Data`
- Added `BedrockImageSource.Media_type` field for content type
- Enhanced validation for Bedrock message format
- Updated schema version to "2.0.0"

### Fixed
- Fixed compatibility issues with Bedrock API v2.0.0
- Improved error handling for S3 operations
- Enhanced validation for image formats and sizes

## [1.3.0] - 2025-05-10

### Added
- Added S3-based Base64 storage for Bedrock API integration
- Added `S3StorageConfig` for configuring S3 storage
- Added helper functions for S3 operations
- Added support for date-based hierarchical path structure
- Added validation for S3-based Base64 storage

### Changed
- Split package into multiple files for better maintainability
- Enhanced ImageInfo struct with S3 storage fields
- Updated schema version to "1.3.0"

### Fixed
- Fixed memory usage issues with large Base64 data
- Improved error handling for Base64 operations
- Enhanced validation for image formats and sizes

## [1.0.0] - 2025-05-01

### Added
- Initial release of the schema package
- Core data structures for verification workflow
- Status constants for state transitions
- Validation functions for data integrity
- Helper functions for common operations
- Basic Bedrock message format support
