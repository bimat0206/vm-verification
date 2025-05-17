# Changelog

All notable changes to the InitializeFunction will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.1] - 2025-05-17

### Fixed
- Fixed DynamoDB `ValidationException` error: "Missing the key verificationId in the item"
- Added validation in verification_repo.go to ensure verificationId is present before DynamoDB operations
- Added safeguard to set a default Status value when missing to satisfy DynamoDB schema requirements
- Enhanced input validation in initialize_service.go to verify required fields exist before processing
- Improved error handling for missing primary key fields in DynamoDB operations

## [2.0.0] - 2025-05-17

### Added
- New modular directory structure with cmd/initialize and internal packages
- Direct AWS SDK implementations for DynamoDB and S3 operations
- Repository pattern for data access
- Client wrappers for S3 and DynamoDB

### Changed
- Refactored codebase to use clean architecture principles
- Restructured project to follow Go best practices
- Removed dependency on shared dbutils and s3utils packages
- Maintained compatibility with shared schema and logger packages
- Updated Docker build process to compile from cmd/initialize
- Improved error handling with custom error types
- Enhanced configuration management
- Updated README to reflect new architecture

### Removed
- Removed dependency on shared dbutils package
- Removed dependency on shared s3utils package
- Removed adapter wrappers for shared packages

## [1.5.0] - 2025-05-15

### Changed
- Refactored codebase to use shared packages
- Moved logger implementation to shared/logger package
- Moved dbutils implementation to shared/dbutils package
- Moved s3utils implementation to shared/s3utils package
- Created wrapper classes for backward compatibility
- Updated imports and dependencies to use shared modules
- Improved type safety and code organization

### Fixed
- Fixed type compatibility issues with shared packages
- Resolved dependency conflicts with shared modules
- Added proper wrapper types for shared package interfaces

## [1.4.1] - 2025-05-14

### Fixed
- Removed duplicate type declarations for `HistoricalContext`, `MachineStructure`, and `VerificationSummary` that were defined in both models.go and service.go
- Added clear documentation to indicate canonical type definitions in service.go
- Added section heading comments to improve code organization

## [1.4.0] - 2025-05-14

### Changed
- Migrated to using shared package components with local imports
- Updated to use shared logger package for standardized logging
- Integrated shared s3utils package for S3 operations
- Integrated shared dbutils package for DynamoDB operations
- Removed custom logger implementation in favor of shared version
- Updated Docker build process to work with local imports

## [1.3.0] - 2025-05-14

### Changed
- Updated imports to use local module paths instead of GitHub dependencies
- Changed schema import from GitHub to `workflow-function/shared/schema`
- Updated Dockerfile to support local module imports
- Simplified build process by only copying necessary directories

## [1.2.0] - 2025-05-14

### Added
- Integration with shared schema package for standardized data models
- Support for standardized status transitions managed by Step Functions
- Explicit schema version handling (1.0.0)
- Standardized error handling with error info structure
- Backward compatibility with legacy format requests

### Changed
- Updated DynamoDBUtils to work with the standardized schema
- Modified validation logic to use schema validation functions
- Refactored code to support both new and legacy formats
- Improved error messages with standardized codes and formats
- Added schema version to DynamoDB records

### Removed
- Status management code (now handled by Step Functions)
- Manual verification context creation (now uses schema package)

## [1.1.3] - 2025-05-11

### Fixed
- Added explicit handling in the `Process` method to ensure that `previousVerificationId` field is always set for `PREVIOUS_VS_CURRENT` verification type, even if not provided in the request
- Fixed Step Function error: `JSONPath '$.verificationContext.previousVerificationId' could not be found in the input`
- Enhanced logging to track state of `previousVerificationId` field throughout processing

## [1.1.2] - 2025-05-11

### Fixed
- Removed `omitempty` JSON tag from `PreviousVerificationId` field in the `VerificationContext` struct to ensure it's always included in the JSON output
- Added enhanced logging to track the serialization and presence of the `previousVerificationId` field
- Fixed Step Function error: `JSONPath '$.verificationContext.previousVerificationId' could not be found in the input`

## [1.1.1] - 2025-05-09

### Fixed
- Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error when invoked from Step Functions
- Enhanced input parsing to properly extract top-level requestId and requestTimestamp fields when verificationContext is present
- Improved error logging with more detailed JSON content for debugging

## [1.1.0] - 2025-05-08

### Changed
- Made `previousVerificationId` and `vendingMachineId` optional for `PREVIOUS_VS_CURRENT` verification type
- Updated validation logic in `service.go` to remove requirement for `previousVerificationId` in `PREVIOUS_VS_CURRENT` type
- Ensured alignment between API Gateway model, Step Function state machine, and function validation

### Fixed
- Fixed potential validation error when `PREVIOUS_VS_CURRENT` verification requests don't include `previousVerificationId`

## [1.0.0] - 2025-04-20

### Added
- Initial implementation of InitializeFunction
- Support for two verification types: `LAYOUT_VS_CHECKING` and `PREVIOUS_VS_CURRENT`
- Validation for required fields based on verification type
- Resource validation for images and layouts
- Historical context retrieval for `PREVIOUS_VS_CURRENT` verification type
- DynamoDB integration for storing verification records
- S3 integration for validating image existence