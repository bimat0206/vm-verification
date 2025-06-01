# Changelog

## [1.3.0] - 2025-01-03

### Added
- **Thinking Content Infrastructure**: Comprehensive support for Claude thinking content
  - Added `ThinkingTokens` field to `TokenUsage` structure for thinking token tracking
  - Enhanced `Turn1Response` and `Turn2Response` with `Thinking` field for thinking content
  - Implemented `ExtractThinkingFromResponse` utility function for thinking content extraction
  - Added thinking content processing in `ProcessTurn1Response` and `ProcessTurn2Response`
  - Enhanced token usage calculation to include thinking tokens when available

### Enhanced
- **Future-Proof AWS SDK Integration**: Prepared for AWS SDK thinking support
  - Added placeholder configuration for Claude reasoning/thinking mode in client
  - Implemented conditional thinking content block processing (awaiting AWS SDK support)
  - Added proper logging for thinking-related features and token usage
  - Structured code to easily enable full functionality when AWS SDK adds thinking support

### Technical Details
- **Response Processing**: Enhanced response processors to extract and include thinking content
- **Token Usage**: Extended token usage tracking to include thinking tokens in total calculations
- **Content Extraction**: Added robust thinking content extraction with proper error handling
- **Backward Compatibility**: All thinking fields are optional with graceful fallback behavior
- **Infrastructure**: Complete thinking content support infrastructure ready for AWS SDK updates

### Infrastructure Changes
- **Types**: Enhanced response types with thinking content fields
- **Utils**: Added thinking content extraction utilities
- **Client**: Prepared client for thinking/reasoning configuration
- **Response Processing**: Complete thinking content integration in response processing pipeline

## [1.2.0] - 2025-05-17

### Added
- Added support for Turn 2 in conversations
- Added `Turn2Response` type for handling Turn 2 responses
- Added `ProcessTurn2Response` function to process Turn 2 responses
- Added `CreateTurn2ConversationHistory` helper function
- Added `CreateConverseRequestForTurn2` function for Turn 2 requests
- Added `CreateConverseRequestForTurn2WithImages` for Turn 2 with images
- Added validation functions for Turn 1 and Turn 2 responses
- Added constants for analysis stages (TURN1, TURN2)

### Changed
- Removed all S3 URI functionality, now exclusively using base64 encoding
- Removed `parseS3URI` function from client.go
- Removed `S3Location` handling from ImageSource struct
- Removed `CreateImageContentFromS3` function
- Simplified `CreateImageContentBlock` to only accept bytes
- Removed S3 URI validation functions
- Updated constants to support both Turn 1 and Turn 2
- Removed support for legacy InvokeModel API, now Converse API only
- Enhanced error handling in `Converse` method with input validation

### Technical Details
- Multi-turn conversation support with proper history management
- Complete transition to base64-encoded image data
- Simplified API surface focusing only on Converse API
- Added PreviousTurn linking in Turn2Response for conversation context

## [1.1.0] - 2025-05-15

### Fixed
- Fixed Bedrock API validation error "messages.0.content.0.type: Field required" in client.go
- Fixed Bedrock API validation error "messages.0.content.1.image.source: Field required" in client.go
- Ensured proper JSON structure for all Bedrock API requests to comply with API requirements
- Corrected message structure in Converse method to include all required fields

### Added
- Added enhanced debugging with detailed request logging
- Added JSON structure logging for image content blocks
- Added full request logging before API calls
- Created test utilities to validate content structure
- Added image_test.go for message structure validation

### Enhanced
- Improved error handling for Bedrock API validation failures
- Added additional comments explaining required message structure
- Updated documentation with correct message structure examples

## [1.0.0] - 2025-05-01

### Added
- Initial implementation of shared Bedrock package
- Added support for Converse API
- Implemented response processing and parsing
- Created structured types for API interaction
- Added token usage tracking and metrics
- Implemented robust error handling
- Added centralized client configuration
- Implemented response extraction utilities