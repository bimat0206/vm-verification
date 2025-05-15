# Changelog

## [1.2.0] - 2025-05-15

### Fixed
- Fixed Bedrock API validation error "messages.0.content.0.type: Field required" by adding required type field
- Fixed Bedrock API validation error "messages.0.content.1.image.source: Field required" by ensuring proper image structure
- Updated shared/bedrock/client.go to properly structure both text and image content blocks

### Added
- Added detailed request logging for easier troubleshooting
- Added JSON structure logging for image content blocks
- Added full request logging before Bedrock API calls
- Created test-image-structure.go utility to verify content structure
- Added comprehensive test for image content structure validation

### Enhanced
- Enhanced error handling for Bedrock API request failures
- Improved debugging capabilities for request structure issues
- Added additional comments explaining required message structure

## [1.1.0] - 2025-05-15

### Changed
- Refactored to use shared bedrock package exclusively
- Removed internal bedrock.go, types.go, validation.go, and response.go files
- Updated to use Bedrock Converse API only (removed legacy InvokeModel support)
- Simplified code structure and reduced duplication

### Added
- Enhanced validation using shared validation utilities
- Added support for standardized schema package
- Created validation_fix.go to provide compatibility with shared validation
- Added comprehensive README.md with usage instructions

### Fixed
- Improved error handling with shared errors package
- Updated error handling to use shared error types
- Updated go.mod to include all required shared packages
- Maintained backward compatibility with existing code
- Updated Dockerfile to properly handle shared packages
- Enhanced retry-docker-build.sh with better error handling and shared package support

## [1.0.0] - 2025-05-01

### Added
- Initial implementation of ExecuteTurn1 Lambda function
- Bedrock client integration for Claude 3.7 Sonnet model
- Support for two verification types: LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT
- Environment-based configuration for flexible deployment
- Comprehensive error handling with retry logic
- Docker-based deployment for AWS Lambda
