# CHANGELOG.md

```markdown
# Changelog

PrepareSystemPrompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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