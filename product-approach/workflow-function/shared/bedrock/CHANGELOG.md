# Changelog

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