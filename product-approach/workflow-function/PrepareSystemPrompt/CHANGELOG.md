# Changelog

PrepareSystemPrompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.1] - 2025-05-22

### Changed
- Moved bedrockConfig field to top level in the response structure
- Updated shared schema package to include bedrockConfig at top level
- Fixed systemPrompt structure to use "content" field instead of nested "systemPrompt"
- Removed duplicate BedrockConfig field in Response struct to prevent JSON duplication

### Breaking Changes
- SystemPrompt structure in shared schema now separates bedrockConfig to top level

## [1.3.0] - 2025-05-20

### Changed
- Updated to Go 1.24
- Fixed compatibility issues with AWS SDK Go v2
- Added validation package to schema module
- Updated status constants to use schema.StatusPromptPrepared
- Fixed template loader interface compatibility issues
- Updated Dockerfile to include internal package

### Added
- New validation utilities in schema/validation package
- Improved error handling for template loading
- Better compatibility with AWS SDK Go v2

### Fixed
- Fixed "undefined: schema.VerificationStatusSystemPromptReady" error
- Fixed template loader method compatibility issues
- Fixed unused imports in main.go

## [1.2.0] - 2025-05-15

### Changed
- Complete refactoring to use granular shared packages instead of monolithic promptutils
- Migrated to shared/schema for type definitions
- Migrated to shared/templateloader for template management
- Migrated to shared/s3utils for S3 operations
- Added MIGRATION.md with detailed migration approach documentation
- Updated interfaces to match shared package standards
- Modified internal types to adapt shared schema to function-specific needs

### Added
- Support for schema.WorkflowState for consistent data handling
- Enhanced error handling through shared validation library
- Better type safety with shared schema types
- Improved maintainability through shared code

## [1.1.0] - 2025-05-14

### Changed
- Migrated to shared package structure for code reuse across lambda functions
- Refactored main.go to use the shared/promptutils package
- Deprecated internal package in favor of the shared/promptutils package
- Modified build process to exclude internal package
- Updated build process to handle shared package dependencies
- Enhanced documentation to reflect architectural changes
- Added COMPONENT_NAME environment variable for standardized logging

### Security
- Improved security through standardized code patterns
- Better error handling with shared validation logic

## [1.0.1] - 2025-05-11

### Fixed
- Updated S3 URL validator to support spaces and additional special characters in file paths
- Fixed "invalid S3 URL format for checking image" error by making the S3 URL regex pattern more permissive
- Added support for common file name characters like spaces, parentheses, and percent signs in S3 URLs

## [1.0.0] - 2025-05-09

### Added
- Initial release of PrepareSystemPrompt Lambda function
- Support for two verification types: LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT
- Local template loading with versioning support
- Comprehensive input validation for all parameters
- Support for JPEG and PNG images (Bedrock requirement)
- Structured error handling and logging
- Dynamic template data extraction and formatting
- Bedrock configuration optimization
- Environment-based configuration for bucket names and other settings
- Documentation including README and deployment instructions

### Security
- Validation of S3 bucket names against environment variables
- Image format validation to prevent unsupported formats
- Structured error messages that avoid exposing sensitive information
- Support for IAM role-based access control

## [0.2.0] - 2025-04-25

### Added
- Template versioning support
- Dynamic machine structure extraction
- Product mapping formatting
- Bedrock parameter configuration
- Support for historical context in PREVIOUS_VS_CURRENT mode

### Changed
- Switched from hardcoded bucket names to environment variables
- Improved error handling with more detailed messages
- Enhanced logging with structured JSON format

### Fixed
- S3 URL parsing edge cases
- Template path resolution issues

## [0.1.0] - 2025-04-15

### Added
- Initial prototype with basic functionality
- Support for LAYOUT_VS_CHECKING verification type
- Simple template loading
- Basic input validation
- Preliminary documentation

### Known Issues
- Hardcoded bucket names
- Limited error handling
- No support for template versioning
- Incomplete validation for PREVIOUS_VS_CURRENT verification type