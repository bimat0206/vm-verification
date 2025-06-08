# Changelog

All notable changes to the FetchHistoricalVerification Lambda function will be documented in this file.

## [1.2.2] - 2025-06-07

### Added
- **VerificationContext Output**: Added `verificationContext` field to the function output to support Step Function integration
  - Includes complete verification context with `verificationId`, `verificationAt`, `status`, `verificationType`
  - Contains `referenceImageUrl` and `checkingImageUrl` for downstream processing
  - Provides `resourceValidation` with validation timestamp and image existence flags
  - Maintains compatibility with existing S3 state management while providing direct context access

### Changed
- **OutputEvent Structure**: Enhanced `OutputEvent` type to include optional `VerificationContext` field
- **Response Building**: Modified main handler to construct and include verification context in output
- **Status Management**: Verification context status is set to `HISTORICAL_CONTEXT_LOADED` upon successful processing

### Fixed
- **Step Function Integration**: Resolved JSONPath error where `$.verificationContext` was not found in function output
- **FetchImages Compatibility**: Ensured downstream FetchImages function receives expected verification context structure

### Technical Details
- Added `createVerificationContext()` function to build standardized verification context from input
- Resource validation assumes images exist during processing (validation performed by upstream functions)
- Maintains backward compatibility with existing S3 state-based context loading

## [1.2.1] - 2025-06-05

### Fixed
- Now queries DynamoDB using `ReferenceImageIndex` and `referenceImageUrl` to
  correctly locate historical verification records.

## [1.2.0] - 2025-06-04

### Changed
- **Refactored to use shared components**: Migrated from custom implementations to standardized shared packages
- **Error Handling**: Replaced custom error types with `workflow-function/shared/errors` package
  - Updated all error creation to use `errors.NewMissingFieldError()` and `errors.NewValidationError()`
  - Consistent error formatting across all Lambda functions
- **Type System**: Migrated to shared schema types
  - `schema.MachineStructure` instead of custom `MachineStructure`
  - `schema.VerificationSummary` instead of custom `VerificationSummary`
- **Configuration Management**: Simplified environment variable handling
  - Removed custom config abstraction layer
  - Direct `os.Getenv()` usage for better transparency
- **Docker Build**: Updated Dockerfile to include new shared packages (`errors`, `s3state`)

### Added
- **Shared Package Dependencies**: Added support for `workflow-function/shared/errors` and `workflow-function/shared/s3state`
- **Future-Ready Architecture**: Prepared for s3state integration when S3 state management is needed
- **Enhanced Validation**: Improved input validation using shared error types

### Removed
- **Custom Error Implementation**: Removed `internal/errors.go` in favor of shared errors package
- **Custom Configuration**: Removed `internal/config.go` abstraction layer
- **Code Duplication**: Eliminated duplicate type definitions now available in shared schema

### Technical Improvements
- **Consistency**: Now follows established patterns from other Lambda functions in the codebase
- **Maintainability**: Reduced code duplication and improved standardization
- **AWS Compatibility**: Maintained full AWS Bedrock API compatibility and Lambda handler structure
- **Build Process**: Preserved Docker containerization and build automation patterns

## [1.1.0] - 2025-05-17

### Changed
- Restructured code to follow standard cmd/internal pattern
- Replaced shared dbutils package with direct DynamoDB calls
- Maintained use of shared schema and logger packages
- Improved error handling for DynamoDB operations
- Updated dependency management

### Added
- DynamoDBRepository for direct DynamoDB interactions
- Exported utility functions for better package integration

### Removed
- Dependency on shared dbutils package

## [1.0.0] - 2025-05-14

### Changed
- Migrated to shared package components
- Replaced custom logger with standardized shared logger
- Replaced DynamoDB client with shared dbutils package
- Added schema version handling with shared schema package
- Improved error handling and standardized error responses
- Updated code structure to match other Lambda functions in the workflow
- Implemented dependencies.go for consistent service initialization

### Added
- ConfigVars struct for standardized environment variable configuration
- DBWrapper for consistent database interface
- Structured logging with correlation IDs

### Removed
- Function-specific DynamoDB client implementation
- Custom logging implementation
- Stand-alone validation module