# Changelog

All notable changes to the PrepareTurn1Prompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.0.4] - 2025-05-25

### Fixed
- Fixed duplicate case error in processor.go for "s3-temporary" storage method
- Consolidated image processing logic for S3 temporary storage
- Improved handling of S3 URL processing within the same storage method case

## [4.0.3] - 2025-05-20

### Fixed
- Fixed compilation errors in image handling code
- Updated field names in schema.ImageInfo usage to match shared package
- Replaced deprecated constants with string literals for storage methods
- Simplified image loading from S3 URLs
- Fixed error handling in URL processing

## [4.0.2] - 2025-05-20

### Fixed
- Added fallback mechanism to create image info from URLs in verification context when references are missing
- Improved image processing with support for direct S3 URL processing
- Fixed "Reference image reference not found" error when only URL is available
- Extended the processor to handle cases where image info doesn't have proper storage method set
- Enhanced checking image handling for better compatibility
- Added robust detection and processing of S3 URLs directly from verification context
- Improved error messages for troubleshooting image reference issues

### Added
- New processFromS3URL method in image processor for handling S3 URLs directly
- Support for automatically detecting and setting content type based on image format
- Comprehensive logging for image processing operations

## [4.0.1] - 2025-05-19

### Fixed
- Fixed validation error for turnNumber field in S3 state envelope input
- Enhanced state envelope parsing to handle different schema versions
- Improved input field handling for both direct invocation and Step Functions
- Updated reference key detection with support for multiple naming patterns
- Added robust validation for S3 references with clear error messages
- Fixed main.go handler to properly handle different input formats
- Enhanced compatibility with state machine execution
- Improved error recovery for missing or malformed references

### Added
- Support for alternative reference field names (References/S3References)
- Detailed logging of reference resolution for easier troubleshooting
- Default values for required fields in envelope conversion
- Flexible reference key matching to handle different naming conventions

## [4.0.0] - 2025-05-20

### Added
- Complete refactoring to use shared packages architecture
- Integration with `workflow-function/shared/s3state` package
- Modular codebase with clear separation of concerns
- New directory structure with focused modules
- Enhanced image processing with better format detection
- Updated template handling with improved error recovery
- Comprehensive logging with structured context
- Integrated Bedrock message creation using shared schema

### Changed
- Migrated from custom error types to shared error package
- Replaced custom logging with shared logger package
- Switched from direct S3 operations to s3state package
- Updated prompt generation to use shared template loader
- Improved validation with separate validation module
- Enhanced documentation with clear architecture explanation

### Removed
- Custom S3 client implementation
- Direct AWS SDK usage for S3 operations
- Custom template management code
- Tightly coupled image processing logic

## [2.0.1] - 2025-05-19

### Added
- Enhanced Dockerfile for AWS Lambda ARM64 deployment
- Improved build script with colored output and better error handling
- Automatic shared module resolution and dependency management
- Support for proper Go module resolution in Docker builds

### Fixed
- Docker build issues with shared module dependencies
- Template path resolution in containerized environment
- ARM64 architecture support for AWS Lambda (Graviton)
- Proper directory structure checking in build script

## [2.0.0] - 2025-05-19

### Added
- New S3 state management architecture
- S3 reference-based input/output handling
- Image processing with multiple storage method support
- Validator package for input validation
- Unit tests for validator component
- Comprehensive README documentation
- Dockerfile for containerized deployment
- Build and deployment script with retry logic

### Changed
- Transformed from payload-based to S3 state management architecture
- Refactored code into separate packages with clear responsibilities
- Updated template handling to use existing template structure
- Enhanced error handling with shared error types
- Improved logging with structured context fields

### Removed
- Legacy payload-based input/output handling
- Direct Base64 image processing in main function
- Monolithic function structure

## [1.0.0] - 2025-01-15

### Added
- Initial implementation of PrepareTurn1Prompt function
- Support for layout-vs-checking verification type
- Support for previous-vs-current verification type
- Basic template handling for prompt generation
- Direct Base64 image processing