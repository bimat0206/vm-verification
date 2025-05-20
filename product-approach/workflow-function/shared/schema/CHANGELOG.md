# Changelog

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
