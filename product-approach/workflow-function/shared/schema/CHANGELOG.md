# Changelog

All notable changes to the shared schema package will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-05-22

### Changed
- Updated schema version to 2.0.0 to align with JSON schema definitions
- Updated BedrockImageSource structure with new field naming convention:
  - Changed `Type` field from "bytes" to "base64"
  - Renamed `Bytes` field to `Data`
  - Added `Media_type` field for content type information (e.g., "image/jpeg")
- Refactored s3_helpers.go to avoid duplication with base64_helpers.go:
  - Made `Base64ImageHelpers` a private type with a public singleton instance
  - Created consistent naming patterns across all helper types
  - Improved implementation of the builder pattern
- Updated validation to enforce stricter JSON schema compliance
- Fixed string comparison in ValidateImageFormat to use strings.EqualFold for case-insensitive matching
- Improved error message formatting for validation errors

### Added
- Enhanced support for metadata in BedrockMessage format
- Added getMediaType method to BedrockMessageBuilder for content type determination
- Added custom type imageProcessor for better backward compatibility
- Added more specific error messages for S3 retrieval failures
- Added documentation for schema version 2.0.0 compatibility

### Fixed
- Unused results in array operations (S3 tags generation)
- Inconsistencies between method signatures and implementations
- Updated ImageData.UpdateStorageSummary to work with refactored code
- Fixed BedrockValidation to properly validate v2.0.0 message format

## [1.3.0] - 2025-05-19

### Changed
- Migrated to S3-only storage for Base64 data
- Removed inline Base64 storage support
- Split large codebase into multiple files for better maintainability:
  - constants.go: All constants
  - core.go: Core types and helper functions
  - image_info.go: ImageInfo struct and methods
  - image_data.go: ImageData struct and methods
  - bedrock.go: Bedrock-related types and functions
  - s3_helpers.go: S3 storage helpers (replacing base64_helpers.go)
  - validation.go: Validation functions
  - types.go: Additional type definitions
- Updated validation to check for S3 storage instead of inline Base64 data
- Renamed HybridStorageConfig to S3StorageConfig
- Renamed HybridBase64Retriever to S3Base64Retriever
- Renamed HybridImageInfoBuilder to S3ImageInfoBuilder
- Renamed HybridImageDataProcessor to S3ImageDataProcessor

### Removed
- Inline Base64 storage support
- Base64Data field from ImageInfo
- StorageMethodInline constant
- DefaultBase64SizeThreshold constant
- TotalInlineSize field from ImageData

## [1.0.0] - 2025-05-14

### Added
- Initial implementation of shared schema package
- Core data models: VerificationContext, ImageData, ConversationState, etc.
- Status constants with explicit state transitions aligned with Step Functions
- Validation functions for verification context and workflow state
- ErrorInfo structure for standardized error reporting
- Helper functions for common operations
- Comprehensive README with usage instructions
- Backward compatibility support
