# Changelog

All notable changes to the InitializeFunction will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.2.1] - 2025-06-02

### Fixed
- **CRITICAL**: Fixed system prompt structure mismatch causing "unexpected end of JSON input" error in ExecuteTurn1Combined
- **Root Cause**: InitializeService was creating flat systemPrompt structure but ExecuteTurn1Combined expected nested promptContent structure
- **Solution**: Updated `createInitializationData` function to generate correct nested structure:
  ```json
  "systemPrompt": {
    "promptContent": {
      "systemMessage": ""
    },
    "promptId": "...",
    "promptVersion": "1.0.0"
  }
  ```
- **ExecuteTurn1Combined Compatibility**: Fixed LoadSystemPrompt parsing by matching expected JSON schema
- **Error Prevention**: Eliminates JSON parsing failures when ExecuteTurn1Combined loads system prompt from S3

### Technical Details
- **Issue**: ExecuteTurn1Combined's LoadSystemPrompt expected `promptContent.systemMessage` but received flat `content` field
- **Fix**: Modified systemPrompt structure in initialization.json to match ExecuteTurn1Combined's parser expectations
- **Impact**: Resolves "failed to parse event detail: unexpected end of JSON input" error completely
- **Validation**: Ensures proper JSON structure compatibility between Initialize and ExecuteTurn1Combined functions

## [3.2.0] - 2025-06-02

### Fixed
- **CRITICAL**: Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error in ExecuteTurn1Combined function
- **Root Cause**: Initialize function was saving raw VerificationContext to S3 instead of proper InitializationData structure
- **Solution**: Modified Initialize function to create proper InitializationData structure with required schemaVersion field
- Added `createInitializationData` function to wrap VerificationContext in expected format
- Set schemaVersion to "2.1.0" as required by ExecuteTurn1Combined function
- Fixed structure mismatch between Initialize output and ExecuteTurn1Combined input expectations
- **Layout Metadata Integration**: Fixed layoutMetadata field to contain actual layout data instead of null

### Added
- **InitializationData Structure**: Added proper initialization data structure creation with:
  - `schemaVersion` field set to "2.1.0"
  - `verificationContext` field containing the verification context
  - `systemPrompt` field with placeholder structure
  - `layoutMetadata` field populated with actual layout data for LAYOUT_VS_CHECKING verification type
- **Layout Metadata Fetching**: Added automatic fetching of layout metadata from DynamoDB during initialization
- **Enhanced Error Prevention**: Prevents JSON parsing errors in downstream functions
- **Schema Compliance**: Ensures compatibility with ExecuteTurn1Combined function expectations

### Changed
- **S3 Storage Format**: Modified `createStateStructure` method to save InitializationData instead of raw VerificationContext
- **Data Structure**: Updated saved initialization data to match ExecuteTurn1Combined input schema exactly
- **Layout Metadata Handling**: Changed from storing null to fetching and storing actual layout metadata
- **Error Handling**: Improved error handling for initialization data creation and layout metadata fetching

### Technical Details
- **Issue**: ExecuteTurn1Combined expected InitializationData with schemaVersion and populated layoutMetadata but received raw VerificationContext
- **Fix**: Wrapped VerificationContext in proper InitializationData structure with actual layout metadata before S3 storage
- **Layout Integration**: Added DynamoDB query to fetch complete layout metadata including productPositionMap and machineStructure
- **Impact**: Eliminates JSON parsing errors and ensures proper workflow execution with complete layout data
- **Compatibility**: Maintains backward compatibility while fixing schema mismatch and providing complete initialization data

### Structure Compliance
The initialization.json now matches the expected format:
```json
{
  "schemaVersion": "2.1.0",
  "verificationContext": { ... },
  "systemPrompt": { ... },
  "layoutMetadata": {
    "layoutId": 23591,
    "layoutPrefix": "5560c9c9",
    "vendingMachineId": "VM-23591",
    "productPositionMap": { ... },
    "machineStructure": { ... }
  }
}
```

## [3.1.0] - 2025-01-02

### Fixed
- **CRITICAL**: Fixed root cause of "failed to parse event: failed to parse event detail: unexpected end of JSON input" error
- Enhanced JSON validation with comprehensive pre-parsing checks for size, structure, and completeness
- Improved error handling for truncated or malformed JSON input
- Added specific error messages for different types of JSON parsing failures

### Added
- **Enhanced JSON Validation**: Added pre-parsing validation checks including:
  - Empty JSON input detection
  - Minimum valid JSON size validation
  - Basic JSON structure validation (proper opening/closing braces)
  - JSON content preview logging for debugging
- **Improved Error Messages**: Added specific error messages for truncated JSON vs. other parsing errors
- **Enhanced Logging**: Added comprehensive logging throughout the event parsing pipeline
- **Input Size Tracking**: Added logging of JSON input size and content preview for debugging
- **Event Type Detection**: Enhanced logging to track different event types and their processing paths
- **Helper Functions**: Added utility functions for JSON validation and error handling

### Changed
- **Error Handling Strategy**: Improved error messages to provide more specific information about JSON parsing failures
- **Logging Enhancement**: Added detailed logging for each step of the event parsing process
- **Input Validation**: Strengthened input validation to catch issues before they cause parsing errors
- **Event Processing**: Enhanced event marshaling and unmarshaling with better error context

### Technical Details
- **Root Cause**: JSON input was being truncated or incomplete, causing "unexpected end of JSON input" errors
- **Solution**: Added comprehensive validation pipeline that checks JSON integrity before parsing
- **Error Prevention**: Implemented early detection of malformed JSON to provide better error messages
- **Debugging Support**: Enhanced logging provides detailed information for troubleshooting JSON parsing issues

### Impact
- Eliminates "unexpected end of JSON input" errors at the source
- Provides clear error messages for different types of JSON parsing failures
- Improves debugging capabilities with detailed logging
- Prevents error propagation to downstream workflow steps

## [3.0.6] - 2025-12-19

### Enhanced
- Enhanced `StoreMinimalRecord` method to include essential verification fields in DynamoDB storage
- Added `referenceImageUrl` and `checkingImageUrl` fields to MinimalVerificationRecord struct
- Added `layoutId` and `layoutPrefix` fields to MinimalVerificationRecord struct for LAYOUT_VS_CHECKING verification type support
- Updated record initialization to populate all essential verification parameters from VerificationContext
- Enhanced logging to include image URLs and layout information for better debugging and monitoring

### Changed
- Modified MinimalVerificationRecord struct to store comprehensive verification metadata while maintaining lightweight approach
- Updated DynamoDB storage to include all key verification parameters for improved query capability and workflow continuity
- Improved field mapping to ensure proper JSON and DynamoDB attribute value serialization

### Benefits
- Complete verification context now stored in DynamoDB for efficient querying and filtering
- Subsequent workflow steps can access all necessary verification information directly from DynamoDB
- Enhanced debugging capabilities with comprehensive logging of stored fields
- Maintained backward compatibility while expanding stored data scope

## [3.0.5] - 2025-05-31

### Fixed
- Fixed RESOURCE_VALIDATION_FAILED error by implementing mandatory layout lookup for LAYOUT_VS_CHECKING verification type
- Enhanced layout lookup logic to automatically retrieve layoutId and layoutPrefix from DynamoDB using referenceImageUrl when missing from request
- Added comprehensive error handling for layout lookup failures with specific error codes:
  - LAYOUT_LOOKUP_FAILED: When DynamoDB query fails
  - LAYOUT_NOT_FOUND: When no layout found for referenceImageUrl
  - INVALID_LAYOUT_METADATA: When retrieved layout has missing fields
- Improved validation to ensure both layoutId and layoutPrefix are properly populated after successful lookup
- Enhanced state management to properly save error information for all failure scenarios
- Added detailed logging throughout layout lookup process for better observability

### Changed
- Modified Process method to enforce successful layout resolution when layoutId/layoutPrefix are missing
- Updated error handling to fail initialization immediately on layout lookup failure instead of proceeding to downstream validation
- Enhanced verification context validation to ensure complete layout metadata before resource verification

## [3.0.4] - 2025-05-19

### Changed
- Reorganized S3 state structure to use date-based hierarchical path: `{year}/{month}/{day}/{verificationId}/`
- Implemented helper method `getDateBasedPath` to standardize the path generation
- Modified folder creation and S3 storage to use the date-based structure
- Improved S3 performance by distributing objects across date-based prefixes
- Enhanced logging to show the date-based paths being used

## [3.0.3] - 2025-05-19

### Fixed
- Fixed Step Functions integration error by ensuring `verificationContext` field is included in the output
- Added `ExtendedEnvelope` type to maintain compatibility with Step Functions Choice state requirements
- Updated Process method to return the extended envelope format with both S3 references and verification context
- Improved logging to detect when verification context is included in the output

## [3.0.2] - 2025-05-19

### Fixed
- Fixed critical bug where `VerificationAt` field was not set when using standardized schema format
- Added validation and auto-setting of `VerificationAt` field to prevent DynamoDB errors
- Improved logging to track when `VerificationAt` is missing and automatically set

## [3.0.1] - 2025-05-19

### Fixed
- Fixed dependency management for shared s3state package
- Resolved compilation issues with SchemaVersion field not found in VerificationContext
- Fixed undefined variable references in previous verification lookup
- Corrected import statements for proper module resolution
- Added go.mod replacements for all shared packages

## [3.0.0] - 2025-05-19

### Added
- Integration with shared S3 state management package
- S3 state folder structure creation for verification
- S3StateManagerWrapper for managing state operations
- STATE_BUCKET environment variable support
- S3 reference-based output format instead of full data payload
- Minimal DynamoDB record storage with S3 references

### Changed
- Refactored to use S3 state management architecture
- Transformed function from "data creator" to "state initializer"
- Updated `Process` method to return S3 state envelope
- Simplified `createVerificationContext` to use shared schema
- Enhanced error handling to store errors in S3 state
- Modified lambda handler to work with S3 references
- Updated resource validation to store results in S3

### Removed
- Historical context retrieval from Initialize function (moved to FetchHistoricalVerification)
- Complex conversation configuration handling (simplified)

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
