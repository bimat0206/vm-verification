# Changelog

All notable changes to the Lambda module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2025-05-22

### Changed
- **BREAKING CHANGE**: Consolidated 6 turn-specific Lambda functions into 2 combined functions
- Removed `prepare_turn1_prompt`, `execute_turn1`, `process_turn1_response`, `prepare_turn2_prompt`, `execute_turn2`, `process_turn2_response` functions
- Added `execute_turn1_combined` function (1024MB memory, 120s timeout) - combines Turn1 prompt preparation, Bedrock execution, and response processing
- Added `execute_turn2_combined` function (1536MB memory, 150s timeout) - combines Turn2 prompt preparation, Bedrock execution, and response processing
- Retained `prepare_system_prompt` function as shared system-level prompt logic
- Simplified Lambda function architecture from 7 turn-related functions to 4 functions (43% reduction)

### Benefits
- Reduced complexity with fewer state transitions in Step Functions workflow
- Atomic turn operations reduce failure points
- Enhanced error handling with single point per turn
- Improved resource efficiency with optimized memory allocation
- Maintains modularity while reducing operational overhead

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of the Lambda module
- Support for creating Lambda functions from various sources (S3, local files, container images)
- Environment variable configuration
- Memory and timeout settings
- VPC configuration
- Dead letter queue configuration
- CloudWatch log group creation
- Lambda function URL configuration
- Lambda permissions management
- Tagging support