# Changelog

All notable changes to the API Verifications Status Lambda Function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.1] - 2025-01-30

### Fixed
- Fixed "Object of type Decimal is not JSON serializable" error
- Added custom JSON encoder to handle DynamoDB Decimal types
- Added convert_decimals helper function for nested structures
- Updated all JSON serialization to use DecimalEncoder

### Added
- Unit test for Decimal type conversion

## [2.0.0] - 2025-01-30

### Changed
- **BREAKING**: Completely refactored from Go to Python to resolve persistent AWS SDK v2 endpoint resolution issues
- Replaced AWS SDK Go v2 with boto3 for more reliable DynamoDB operations
- Simplified codebase while maintaining the same API interface
- Improved error handling with Python's exception model

### Added
- Unit tests for Lambda function
- Python-specific deployment script
- Comprehensive documentation for Python implementation

### Fixed
- Eliminated "ResolveEndpointV2" errors that plagued the Go implementation
- Resolved AWS SDK module version incompatibility issues
- Fixed endpoint resolution problems in Lambda environment

### Technical Details
- Runtime: Python 3.11
- SDK: boto3 1.34.0
- Deployment: Docker container on AWS Lambda
- No external dependencies beyond boto3

### Migration Notes
- The API interface remains unchanged - no client modifications needed
- Environment variables remain the same
- Response format is identical to the Go implementation
- Go implementation preserved in `old/` directory for reference

## Previous Go Implementation

See `old/CHANGELOG.md` for the history of the Go implementation (versions 1.0.0 - 1.1.0).