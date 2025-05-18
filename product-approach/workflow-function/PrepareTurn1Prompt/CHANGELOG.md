# PrepareTurn1Prompt Changelog

## v1.3.1 - 2025-05-18

### Bug Fixes

- Fixed struct field issues in bedrock.go to match Bedrock API requirements:
  - Updated BedrockImageSource field references from `Data` to `Bytes`
  - Changed image source type from `base64` to `bytes` to match Bedrock schema
  - Fixed S3 references to use the `bytes` field instead of deprecated `URI`
- Fixed undefined type references:
  - Changed `schema.ContentBlock` to `schema.BedrockContent`
  - Changed `schema.ImageBlock` to `schema.BedrockImageData`
- Fixed S3 temporary storage field references:
  - Updated `S3TempBucket` to `Base64S3Bucket`
  - Updated `S3TempKey` to `Base64S3Key`
- Fixed struct nil comparison issues with BedrockImageSource and Thinking
- Removed unused imports from s3client.go
- Fixed function redeclaration conflicts between utils.go and s3client.go

## v1.3.0 - 2025-05-18

### Major Features Added

- **Base64 Image Processing**: Implemented hybrid Base64 storage approach for Bedrock compatibility
 - Added automatic image retrieval and Base64 encoding for S3 images
 - Support for inline Base64 data and S3-temporary storage methods
 - Automatic image format detection and validation (JPEG/PNG only)
 - Size validation (10MB limit for Bedrock compatibility)

- **Enhanced S3 Client**: New S3ImageProcessor with comprehensive image handling
 - Smart storage method detection (inline, S3-temporary, direct S3)
 - Automatic Base64 encoding with validation
 - Configurable timeouts and size limits via environment variables
 - Proper error handling for S3 access issues

- **Improved Error Handling**: Added specific error types for different failure scenarios
 - ImageProcessingError for image-related failures
 - StorageMethodError for storage handling issues
 - S3AccessError for S3 operation failures
 - Base64Error for encoding/decoding issues
 - ConfigurationError for missing/invalid configuration

- **Comprehensive Validation**: Enhanced input validation with detailed error messages
 - Environment variable validation (REFERENCE_BUCKET, CHECKING_BUCKET)
 - S3 URL format and bucket validation
 - Machine structure validation with reasonable limits
 - Improved verification type-specific validation

### Code Improvements

- **Removed Hardcoded Values**: All configuration now sourced from environment variables
 - Template paths, bucket names, Bedrock settings
 - Configurable timeouts and limits
 - Flexible configuration with validation

- **Enhanced Logging**: Added structured logging with verification ID context
 - Image processing status logging
 - Configuration validation warnings
 - Performance metrics (processing duration)
 - Detailed error context for troubleshooting

- **Modular Architecture**: Separated concerns into focused modules
 - `processor.go`: Core image processing and template data building
 - `s3client.go`: S3 operations and image handling
 - `bedrock.go`: Bedrock message creation and validation
 - `validator.go`: Comprehensive input validation
 - `errors.go`: Specific error types and handling
 - `utils.go`: Common utility functions

### Environment Variables

- **Required Variables**:
 - `TEMPLATE_BASE_PATH`: Path to template directory
 - `REFERENCE_BUCKET`: S3 bucket for reference images
 - `CHECKING_BUCKET`: S3 bucket for checking images

- **Optional Variables** (with validation warnings):
 - `ANTHROPIC_VERSION`: Bedrock API version
 - `MAX_TOKENS`: Maximum tokens for Bedrock response
 - `BUDGET_TOKENS`: Tokens for Claude's thinking process
 - `THINKING_TYPE`: Claude's thinking mode
 - `BASE64_RETRIEVAL_TIMEOUT`: Timeout for S3 operations (ms)
 - `MAX_IMAGE_SIZE_MB`: Maximum image size limit

### Bug Fixes

- Fixed missing Base64 data issue that prevented proper Bedrock integration
- Resolved template loading errors with proper naming convention handling
- Improved Docker build reliability with enhanced shared module dependencies
- Added proper panic recovery with detailed stack traces

### Performance Improvements

- Parallel image processing where possible
- Optimized S3 operations with configurable timeouts
- Efficient Base64 encoding for large images
- Template caching for improved performance

### Breaking Changes

- Function now requires Base64 image processing before sending to ExecuteTurn1
- Environment variables are now required (no hardcoded defaults)
- Updated response structure to include Base64 data in ImageInfo

### Dependencies Updated

- Added AWS SDK v2 for S3 operations
- Enhanced integration with shared packages (schema, logger, errors)
- Improved templateloader integration

## v1.2.0 - 2025-05-17

### Bug Fixes and Improvements

- Fixed template loading error by correctly handling template naming conventions (replacing underscores with hyphens)
- Fixed Docker build issues by improving shared module dependency handling
- Added panic recovery middleware to gracefully handle unexpected errors
- Improved error logging with detailed stack traces for better debugging
- Updated retry-docker-build.sh script to properly handle module dependencies

## v1.1.0 - 2025-05-17

### Migration to Shared Packages

- Migrated to use shared schema package for standardized types
- Migrated to use shared logger package for consistent logging
- Migrated to use shared templateloader package for template management
- Migrated to use shared bedrock package for Bedrock API interactions
- Migrated to use shared errors package for standardized error handling
- Removed duplicate code and unused functions
- Updated code to follow best practices
- Ensured all AWS dependencies are properly included

### Code Improvements

- Added proper error handling with structured errors
- Improved logging with structured logs
- Standardized template loading and rendering
- Simplified Bedrock message creation using shared utilities
- Removed hardcoded values and replaced with constants from schema
- Added proper validation for input parameters
- Improved code organization and readability