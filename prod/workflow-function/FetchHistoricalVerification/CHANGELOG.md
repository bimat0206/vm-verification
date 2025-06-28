# Changelog

All notable changes to the FetchHistoricalVerification Lambda function will be documented in this file.

## [1.2.4] - 2025-06-28

### Fixed
- **Status Handling for Fresh Verifications**: Fixed logical contradiction where `StatusHistoricalContextLoaded` was set even when no historical data was found
  - Now sets `StatusHistoricalContextNotFound` when `historicalDataFound` is false
  - Sets `StatusHistoricalContextLoaded` only when historical data actually exists
  - Status is properly propagated through the enhanced verification context
  - Ensures consistent workflow state with downstream FetchImages function

### Changed
- **Status Management**: Enhanced status setting logic to differentiate between historical data found vs not found scenarios
- **Enhanced Verification Context**: Updated `createEnhancedVerificationContext` to set appropriate status based on `historicalDataFound` flag
- **Historical Context JSON**: Added `status` field to HistoricalContext struct to include status in historical-context.json file
  - Sets `StatusHistoricalContextNotFound` in fallback context for fresh verifications
  - Sets `StatusHistoricalContextLoaded` when historical data is successfully retrieved

### Technical Details
- Added conditional status setting based on `result.HistoricalDataFound` in main handler
- Modified `createEnhancedVerificationContext` to determine status dynamically
- Added `Status` field to `HistoricalContext` struct in types.go
- Updated both `createFallbackContext` and `createHistoricalContext` functions to set appropriate status
- Maintains backward compatibility while improving workflow state consistency

## [1.2.3] - 2025-06-20

### Fixed
- **Fresh Verification Error Handling**: Resolved ValidationException error for fresh verifications (verifications without previous historical data)
  - Modified `FindPreviousVerification` to return `(nil, nil)` instead of ValidationError when no previous verification is found
  - Updated service logic to handle `nil` verification as a normal case rather than an error condition
  - Changed logging level from WARN to INFO for fresh verifications to reduce noise in logs
  - Updated `SourceType` from `"NO_HISTORICAL_DATA"` to `"FRESH_VERIFICATION"` for better clarity
  - Removed unused `workflow-function/shared/errors` import from dynamodb.go

### Changed
- **Error Classification**: Fresh verifications (no previous data) are now treated as normal operations instead of error conditions
- **Logging Improvements**: 
  - Fresh verifications now log as INFO: "No previous verification found - this is normal for fresh verifications"
  - Fallback context creation logs as INFO: "Creating fallback context for fresh verification"
  - Only actual DynamoDB query errors are logged as ERROR level

### Technical Details
- Enhanced `createFallbackContext()` function with better logging and clearer field values
- Improved error handling in `FetchHistoricalVerification()` to distinguish between query errors and no-data scenarios
- Maintained backward compatibility for historical verifications (verifications with previous data)
- Function now properly supports both fresh and historical verification workflows without generating false error warnings

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