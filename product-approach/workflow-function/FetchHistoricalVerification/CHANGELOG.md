# Changelog

All notable changes to the FetchHistoricalVerification Lambda function will be documented in this file.

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