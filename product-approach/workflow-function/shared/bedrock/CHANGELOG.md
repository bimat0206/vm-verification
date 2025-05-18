# Changelog

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