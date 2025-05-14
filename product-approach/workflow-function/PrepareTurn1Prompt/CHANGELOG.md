# CHANGELOG.md

```markdown
# Changelog

PrepareTurn1Prompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-05-12

### Fixed
- Fixed incompatible assignment when assigning BedrockConfig pointer to value type
- Removed unused mediaType variable in bedrock.go
- Removed unused imports in utils.go (bytes, encoding/base64, image, image/jpeg, image/png, regexp)

### Changed
- Removed hardcoded AWS account ID (111122223333)
- Updated CreateTurn1Message to accept bucket owner from input rather than using default values
- Added ImageMetadata struct to properly represent image metadata from PrepareSystemPrompt
- Modified S3 location handling to use bucket owner from previous function output
- Removed GetAWSAccountID function in favor of using data passed from earlier steps

### Security
- Eliminated hardcoded AWS account IDs to prevent potential misuse
- Improved S3 location handling by using metadata from previous function

## [1.0.0] - 2025-05-12

### Added
- Initial release of PrepareTurn1Prompt Lambda function
- Support for generating Turn 1 prompts for Amazon Bedrock (Claude 3.7 Sonnet)
- Template system with support for two verification types:
  - LAYOUT_VS_CHECKING: Compare layout vs checking image
  - PREVIOUS_VS_CURRENT: Compare previous state vs current state
- Structured validation of input data
- Proper handling of S3 image references for Bedrock
- Configuration via environment variables
- Comprehensive logging and error handling
- Documentation with README.md and CLAUDE.md

### Security
- Validation of input parameters against expected format
- Image format validation to ensure compatibility with Bedrock
- Structured error messages that avoid sensitive information exposure
- Support for IAM role-based access control
```