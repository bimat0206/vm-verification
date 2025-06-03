# Changelog

## [1.0.8] - 2025-06-03
### Fixed
- **Output References**: Always return `processing_initialization` reference when provided in the input
  - Falls back to the input S3 reference if initialization data could not be loaded
  - Ensures Step Functions receive complete S3 reference data even on failures

## [1.0.7] - 2025-06-03
### Fixed
- **Critical Bug**: Fixed "failed to get processing initialization reference" error
  - Added conditional logic to only retrieve processing reference when initialization data is available
  - Function now handles cases where initialization data cannot be loaded from S3
  - Added proper fallback to error-only output when processing reference is unavailable
  - Enhanced logging to warn when initialization data is missing

### Changed
- **Output Structure**: Modified output generation to conditionally include processing_initialization reference
  - When initialization data is available: includes both processing_initialization and error references
  - When initialization data is missing: includes only error reference with empty processing_initialization
  - Maintains consistent output structure while preventing null pointer errors

### Technical Details
- Added `hasProcessingRef` flag to track when initialization data is successfully saved
- Conditional reference retrieval prevents attempting to get non-existent references
- Enhanced error handling ensures function completes successfully even when initialization S3 data is unavailable
- Maintains backward compatibility with existing workflow expectations

## [1.0.6] - 2025-06-03
### Added
- **Initialization Data Update**: Function now updates the initialization.json file with error information
  - Added error tracking fields to InitializationData model: `Status`, `ErrorStage`, `ErrorMessage`, `FailedAt`
  - Updates initialization.json with FAILED status and error details when errors occur
  - Saves updated initialization data back to S3 using proper date-based path structure

### Changed
- **Enhanced Output Format**: Updated LambdaOutput to include both S3 references
  - Added `ProcessingInitialization` reference to updated initialization.json
  - Maintained existing `Error` reference to error.json
  - Output now matches expected format with dual S3 references
- **S3 State Management**: Improved S3 storage integration
  - Uses `SaveToEnvelope` method for consistent date-based path structure
  - Proper reference handling for both error and initialization data
  - Fixed config field name from `StateManagementBucket` to `StateBucket` for consistency

### Fixed
- **Import Paths**: Corrected module import paths to use `FinalizeWithErrorFunction` module name
- **Configuration**: Fixed environment variable handling for `STATE_BUCKET`
- **Error Stage Mapping**: Enhanced error stage inference with more specific stage names
  - `IMAGE_FETCH` for image-related errors
  - `TURN1_PROCESSING` and `TURN2_PROCESSING` for turn-specific errors
  - `INITIALIZATION` for initialization errors
  - `PROMPT_PREPARATION` for prompt-related errors
  - `BEDROCK_PROCESSING` for Bedrock service errors

### Technical Details
- Function now produces output with both `processing_initialization` and `error` S3 references
- Initialization.json is updated with failure status and persisted with proper S3 key structure
- Enhanced error handling ensures both error storage and initialization update succeed
- Maintains backward compatibility while adding new functionality

## [1.0.5] - 2025-06-03
### Fixed
- **Critical**: Fixed empty errorStage issue causing "ERROR_" status and empty stage messages
- **Error Stage Inference**: Added intelligent error stage detection from error messages when errorStage is not provided by Step Functions
- **Status Generation**: Fixed status generation to use inferred error stage instead of empty values

### Added
- **inferErrorStage Function**: New function that analyzes error messages to determine the workflow stage where the error occurred
  - Detects "Turn2", "Turn1", "Initialize", "FetchImages", "PrepareSystemPrompt" stages from error content
  - Falls back to "Turn1" for generic Bedrock errors
  - Uses "Unknown" as final fallback
- **Enhanced Logging**: Added INFO-level logging when error stage is inferred from error message

### Changed
- Updated HandleRequest to use inferred errorStage when event.ErrorStage is empty
- Modified output generation to use the determined errorStage for consistent status and message formatting
- Enhanced error stage detection logic with comprehensive pattern matching

### Technical Details
- Error stage inference analyzes error message content using case-insensitive string matching
- For the reported error "BedrockException: failed to invoke Bedrock for Turn2", the function now correctly identifies "Turn2" as the error stage
- Output status changes from "ERROR_" to "ERROR_TURN2" and message from "Verification failed at stage ''" to "Verification failed at stage 'Turn2'"

## [1.0.4] - 2025-06-03
### Fixed
- **Critical**: Fixed "missing verificationAt from initialization data" error when S3 initialization data is unavailable
- **Resilience**: Added fallback logic to extract verificationAt from verificationID when initialization data is missing
- **Field Name**: Corrected field name from `partialS3References` to `s3References` in LambdaInput struct
- **Error Handling**: Enhanced error handling to ensure FinalizeWithError can complete workflow even when initialization S3 references are empty

### Changed
- Updated `UpdateVerificationResultOnError` to parse timestamp from verificationID format (verif-YYYYMMDDHHMMSS-XXXX)
- Updated `UpdateConversationHistoryOnError` to use same fallback logic for conversationAt field
- Modified LambdaInput struct to use correct field name `S3References` instead of `PartialS3References`

### Technical Details
- Added timestamp parsing logic that extracts date/time from verificationID when S3 initialization data is unavailable
- Implemented graceful fallback to current timestamp if all other methods fail
- Ensures DynamoDB updates can proceed even when workflow errors occur early in the process

## [1.0.3] - 2025-06-03
### Fixed
- **DynamoDB Schema Validation Error**: Fixed "The provided key element does not match the schema" error
  - Updated `UpdateVerificationResultOnError` to use composite primary key with both `verificationId` and `verificationAt`
  - Updated `UpdateConversationHistoryOnError` to use composite primary key with both `verificationId` and `conversationAt`
  - Added proper extraction of `verificationAt` from initialization data for both table updates
  - Enhanced error handling when initialization data is missing or incomplete

### Technical Details
- **VerificationResults Table**: Now correctly uses composite key `{verificationId, verificationAt}` instead of just `verificationId`
- **ConversationHistory Table**: Now correctly uses composite key `{verificationId, conversationAt}` instead of just `verificationId`
- **Data Validation**: Added validation to ensure `verificationAt` is available from initialization data before attempting DynamoDB updates
- **Function Signature**: Updated `UpdateConversationHistoryOnError` to accept `initData` parameter for accessing sort key values

### Impact
- Resolves DynamoDB ValidationException errors that were preventing error state updates
- Ensures proper error tracking and status updates in both verification results and conversation history tables
- Maintains data consistency with the established table schema design

## [1.0.2] - 2025-01-03
### Added
- **Docker Containerization**: Added complete Docker containerization setup
  - Multi-stage Dockerfile optimized for AWS Lambda ARM64 deployment
  - Automated build script with ECR integration and Lambda deployment
  - Makefile for comprehensive testing and development workflows
- **Build Automation**: Standardized build process consistent with other workflow functions
  - Cross-compilation support for AWS Lambda (linux/arm64)
  - Shared module handling via temporary build contexts
  - Environment variable configuration support
- **Development Tools**: Enhanced development experience
  - Test automation with coverage reporting
  - Linting and code quality checks
  - CI/CD pipeline support

## [1.0.1] - 2025-01-03
### Fixed
- **Schema Compilation Errors**: Fixed undefined schema types that were causing compilation failures
  - Replaced `schema.InitializationData` with local `models.InitializationData` type definition
  - Replaced `schema.ErrorDetails` with `schema.ErrorInfo` to match actual shared schema
  - Replaced `schema.InputS3References` with local `models.InputS3References` type definition
  - Fixed `StatusHistoryEntry` struct literal to use correct field names (`FunctionName`, `ProcessingTimeMs`, `Stage`, `Metrics`) instead of non-existent `ErrorStage` and `ErrorMessage` fields
- **Type Consistency**: Ensured all schema types align with the shared schema package definitions
- **Build Process**: Added module to workspace to resolve build configuration issues

### Technical Details
- Added local type definitions for `InitializationData`, `InputS3References`, and `S3Reference` in `models.go`
- Updated `dynamodbhelper.go` to use `schema.ErrorInfo` instead of undefined `schema.ErrorDetails`
- Modified `StatusHistoryEntry` creation to include error information in the `Metrics` field as a map
- Updated function signature in `dynamodbhelper.go` to accept `*models.InitializationData` instead of `*schema.InitializationData`

## [1.0.0] - 2025-06-03
- Initial implementation of FinalizeWithError Lambda.
- Handles Step Functions errors and marks verification as failed.
