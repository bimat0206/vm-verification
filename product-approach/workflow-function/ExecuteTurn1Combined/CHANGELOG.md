# Changelog

All notable changes to the ExecuteTurn1Combined function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-05-20

### Added
- Initial release of the ExecuteTurn1Combined Lambda function
- Combined functionality from PrepareSystemPrompt, PrepareTurn1Prompt, and ExecuteTurn1
- Modular service architecture with clean interfaces
- Integration with shared packages (bedrock, s3state, logger, schema)
- Comprehensive error handling and logging
- Support for DynamoDB state tracking
- S3-based state management for large payloads

## [0.9.0] - 2025-05-15

### Added
- Beta implementation with core functionality
- Initial service interfaces and implementations
- Basic error handling and logging
- Configuration from environment variables

### Known Issues
- Missing proper error handling for some edge cases
- Incomplete documentation
- Limited test coverage

## [0.8.0] - 2025-05-10

### Added
- Alpha implementation with basic structure
- Proof of concept for combined workflow
- Initial integration with Bedrock API
- Basic S3 and DynamoDB integration

## [1.0.1] - 2025-05-21

### Fixed
- Fixed import issues with shared packages
- Corrected type conflicts in dynamodb.go
- Updated s3.go to use the correct Reference type from s3state
- Added missing module dependencies in go.mod
- Added ExecuteTurn1Combined to go.work workspace

### Changed
- Improved error handling in bedrock.go
- Enhanced logging for better observability
- Updated documentation with accurate configuration options

## [1.1.0] - 2025-05-25 (Planned)

### Planned Additions
- Enhanced metrics collection
- Support for additional Claude 3.7 features
- Improved error recovery mechanisms
- Performance optimizations for large payloads
- Additional test coverage