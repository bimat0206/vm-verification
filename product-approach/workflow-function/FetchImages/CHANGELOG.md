# Changelog

## [4.2.1] - 2025-06-02

### Fixed
- **CRITICAL**: Fixed missing storage metadata in metadata.json output
- Fixed S3 reference key lookup for Base64 files (dashes are replaced with underscores)
- Corrected reference keys from `"images_reference-base64"` to `"images_reference_base64"`
- Corrected reference keys from `"images_checking-base64"` to `"images_checking_base64"`
- Added comprehensive debug logging for S3 reference lookup

### Technical Details
- **Root Cause**: The s3state `BuildReferenceKey` function replaces dashes with underscores in filenames
- **Fix**: Updated reference lookup to use correct naming convention: `images_reference_base64` and `images_checking_base64`
- **Impact**: Storage metadata now properly populated with bucket, key, storedSize, storageClass, and encryption info

### Expected Output
- Storage metadata now includes complete S3 information:
  ```json
  "storageMetadata": {
    "bucket": "kootoro-dev-s3-state-f6d3xl",
    "key": "2025/06/02/verif-xxx/images/reference-base64.base64",
    "storedSize": 1656692,
    "storageClass": "STANDARD",
    "encryption": {"method": "SSE-S3"}
  }
  ```

## [4.2.0] - 2025-06-02

### Fixed
- **CRITICAL**: Restored Base64 functionality that was removed in version 4.0.0
- Fixed metadata.json format to match expected comprehensive structure
- Restored image download and Base64 conversion capabilities
- Fixed integration with downstream functions that expect Base64 data

### Added
- **Base64 Processing**: Restored full image download and Base64 conversion functionality
- **Comprehensive Metadata**: Added detailed metadata structure matching expected format with:
  - `originalMetadata` with source information and image dimensions
  - `base64Metadata` with encoding details and compression ratios
  - `storageMetadata` with S3 storage information
  - `validation` with Bedrock compatibility checks
  - `processingMetadata` with processing steps and timing
- **S3 Base64 Storage**: Added storage of Base64 data in S3 state bucket using s3state files:
  - `reference-base64.base64` for reference image Base64 data
  - `checking-base64.base64` for checking image Base64 data
- **Image Analysis**: Added image dimension extraction (width, height, aspect ratio)
- **Enhanced S3Repository**: Added `DownloadAndConvertToBase64` method for full image processing
- **Detailed Models**: Added comprehensive data models matching expected API format

### Changed
- **Processing Logic**: Updated parallel fetch to download and convert both images to Base64
- **Metadata Structure**: Enhanced metadata to include all required fields for downstream compatibility
- **Storage Strategy**: Now stores both metadata and Base64 data in S3 state bucket
- **Error Handling**: Improved error handling for image download and conversion failures

### Technical Details
- **Root Cause**: Version 4.0.0 removed Base64 functionality but downstream functions still expected it
- **Architecture**: Maintained s3state integration while restoring Base64 processing
- **Performance**: Parallel processing of image download and conversion
- **Compatibility**: Maintains backward compatibility with existing s3state architecture
- **Validation**: Added Bedrock size limit validation for Base64 data

### Impact
- Restores full functionality expected by PrepareSystemPrompt and other downstream functions
- Generates comprehensive metadata.json matching expected format
- Provides Base64 data for AI/ML processing workflows
- Maintains high performance through parallel processing
- Ensures Bedrock compatibility validation

## [4.1.0] - 2025-01-02

### Fixed
- **CRITICAL**: Fixed root cause of "failed to parse event: failed to parse event detail: unexpected end of JSON input" error
- Enhanced Initialize function with comprehensive JSON validation and error handling
- Improved error isolation between workflow steps to prevent error propagation
- Added error categorization to distinguish between inherited errors and current processing errors
- Enhanced S3 state manager to detect and catalog inherited errors from previous workflow steps

### Added
- **Enhanced JSON Validation**: Added pre-parsing validation checks for JSON size, structure, and completeness
- **Error Source Tracking**: Implemented error categorization system to identify JSON parsing, inherited, and processing errors
- **Inherited Error Detection**: Added automatic detection of errors from previous workflow steps
- **Improved Error Responses**: Enhanced error response structure with detailed error categorization and source information
- **Enhanced Logging**: Added comprehensive logging for error diagnosis and debugging
- **Error Recovery**: Implemented graceful error handling that preserves workflow state while isolating current step errors
- **S3 JSON Validation**: Enhanced S3 RetrieveJSON method with comprehensive validation and error reporting
- **Request Context Prioritization**: Modified verification context loading to prioritize request data over S3 fallback
- **Detailed Request Logging**: Added comprehensive logging of parsed requests and verification context details

### Changed
- **Error Handling Strategy**: Modified FetchImages to return errors in response structure rather than as Go errors
- **State Management**: Enhanced S3 state manager to track error inheritance and provide better error context
- **Logging Enhancement**: Improved logging throughout the error handling pipeline for better debugging
- **Input Validation**: Strengthened input validation in Initialize function to prevent truncated JSON issues

### Technical Details
- **Root Cause**: The error was originating from the Initialize function's JSON parsing logic, not FetchImages
- **Error Propagation**: AWS Step Functions was wrapping Initialize errors and passing them through the workflow
- **Solution**: Implemented three-phase fix:
  1. **Phase 1**: Enhanced Initialize function with robust JSON validation and error handling
  2. **Phase 2**: Improved error isolation in FetchImages to separate inherited vs. current errors
  3. **Phase 3**: Added workflow-level error categorization and enhanced monitoring

### Impact
- Eliminates "unexpected end of JSON input" errors in FetchImages function
- Provides clear error source identification for debugging
- Maintains workflow continuity even when inheriting errors from previous steps
- Improves overall system reliability and error transparency

## [4.0.1] - 2025-05-20

### Fixed
- Fixed import naming collision in handler.go by aliasing internal/config to localConfig
- Fixed ContentLength handling in S3Repository to properly dereference pointer value
- Removed unused imports (strings, aws, schema) to fix compiler warnings
- Fixed proper initialization of ImageInfo struct with nil-safe property assignments
- Fixed "unsupported input type: *models.FetchImagesRequest" error by enhancing S3StateManager.LoadEnvelope to handle FetchImagesRequest directly

## [4.0.0] - 2025-05-19

### Added
- Integrated with shared/s3state package for state management
- Implemented S3 state management with category-based organization
- Added support for reference-based image handling (no Base64 encoding)
- Added a new structured project layout with clear separation of concerns

### Changed
- Complete architectural redesign to use the S3 State Manager pattern
- Refactored to a modern Go project structure (cmd and internal packages)
- Split code into separate packages: config, models, repository, service, and handler
- Updated handler to support both direct invocation and S3 reference-based inputs
- Changed response format to use S3 references instead of inline Base64 data

### Removed
- Removed all Base64 encoding/decoding logic
- Removed response size tracking and hybrid storage approach
- Removed unnecessary storage method validation

## [3.0.0] - 2025-05-16

### Changed
- Refactored to remove dependencies on shared packages (s3utils and dbutils)
- Implemented direct AWS SDK interactions for S3 and DynamoDB operations
- Split codebase into multiple specialized files for better maintainability:
  - s3url.go - S3 URL parsing and validation
  - response_tracker.go - Response size tracking for Lambda limits
  - storage_validation.go - Storage integrity validation
  - storage_stats.go - Storage statistics functions
  - db_models.go - Database models and helpers

### Removed
- Removed dependencies on workflow-function/shared/s3utils
- Removed dependencies on workflow-function/shared/dbutils
- Removed s3wrapper.go and dbwrapper.go files
- Removed wrapper initialization in dependencies.go

## [2.0.3] - 2025-05-14

### Fixed
- Fixed type mismatch error in parallel.go where s3Client (*s3.Client) was incorrectly passed to NewS3Utils function that expects aws.Config
- Updated parallel fetch operations to directly pass AWS config to S3Utils constructor
- Removed unused s3 package import in parallel.go

## [2.0.2] - 2025-05-14

### Fixed
- Fixed Step Functions integration with missing historicalContext field
- Made historicalContext parameter optional in Step Functions state machine definition
- Updated both the local state_machine_definition.json and terraform template 
- Preserved empty historicalContext object generation in Lambda function for backward compatibility

## [2.0.1] - 2025-05-14

### Fixed
- Fixed Step Functions integration issue with missing historicalContext field
- Ensured response always includes historicalContext field even when empty (LAYOUT_VS_CHECKING type)
- Initialized map fields in ParallelFetchResults to prevent nil maps in response
- Updated models to always include historicalContext in JSON output (removed omitempty tag)
- Added documentation on response structure compatibility with Step Functions

## [2.0.0] - 2025-05-14

### Changed
- Migrated to use shared package structure for better consistency across lambda functions
- Moved common code to shared modules:
  - logger: Standardized logging across all lambda functions
  - s3utils: Common S3 operations and utilities
  - dbutils: Common DynamoDB operations and utilities
  - schema: Standardized data models and constants

### Refactored
- Reorganized code structure to leverage dependency injection pattern
- Replaced individual utility implementations with shared package usage
- Updated models to use schema package for standardized types
- Enhanced parallel execution to maintain compatibility with shared packages

### Added
- Added S3UtilsWrapper to adapt shared s3utils package for function-specific use
- Added DBUtilsWrapper to adapt shared dbutils package for function-specific use
- Improved dependency management through ConfigVars struct
- Created specialized build script for handling shared package dependencies
- Updated Docker build process to work with shared package structure

## [1.0.6] - 2025-05-11

### Fixed
- Improved bucketOwner retrieval in S3 metadata using STS GetCallerIdentity
- Added support for AWS_ACCOUNT_ID environment variable for bucket owner
- Made the code more portable across different AWS accounts
- Implemented proper fallback mechanism for bucket owner retrieval
- Added detailed logging for bucket owner determination

## [1.0.5] - 2025-05-11

### Fixed
- Fixed compilation errors in S3 code (GetObjectAttributes API was incompatible)
- Simplified S3 bucketOwner retrieval by using GetBucketAcl API only
- Added error handling for S3 owner retrieval with proper logging
- Made code more resilient to missing bucket owner information

## [1.0.4] - 2025-05-11

### Fixed
- Fixed DynamoDB access issue by properly using table names from environment variables
- Added detailed logging for DynamoDB operations to help with troubleshooting
- Fixed hardcoded table names in dynamodb.go to use config values

## [1.0.3] - 2025-05-11

### Fixed
- Fixed issue with `previousVerificationId` field in Step Functions integration for LAYOUT_VS_CHECKING verification type
- Modified ParallelFetch to only use previousVerificationId when verificationType is PREVIOUS_VS_CURRENT
- Updated validation to explicitly note that previousVerificationId is not required for LAYOUT_VS_CHECKING
- Updated Step Functions example configuration in README.md to use empty string instead of null for conditional previousVerificationId

## [1.0.2] - 2025-05-11

### Added
- Enhanced validation for verification types in models.go
- Added specific validation for referenceImageUrl in PREVIOUS_VS_CURRENT verification type
- Improved error messages for historical verification data fetching
- Added detailed logging for historical verification data

### Changed
- Updated documentation with clearer examples and integration details
- Added data flow diagram to README.md for better visualization
- Enhanced error handling in parallel.go for historical context fetching

## [1.0.1] - 2025-05-10

### Fixed
- Fixed "JSONPath '$.verificationContext.previousVerificationId' could not be found" error in LAYOUT_VS_CHECKING workflow by improving the conditional handling in the state machine definition
- Fixed schema validation error in Step Functions state machine by replacing the ternary operator with States.ArrayGetItem and States.Array intrinsic functions for conditional previousVerificationId handling

## [1.0.0] - 2025-05-10

### Added
- Added bucket owner information (AWS account ID) to image metadata
- Added Bucket and Key fields to ImageMetadata struct for better traceability
- Added S3 GetBucketAcl call to retrieve bucket owner information

### Changed
- Updated state machine definition to ensure historicalContext is at the top level
- Modified ResultPath in FetchImages task to "$" to promote all fields to root level
- Made previousVerificationId field conditional based on verificationType in state machine definition
- Updated validation logic to only require previousVerificationId for PREVIOUS_VS_CURRENT verification type

### Fixed
- Fixed "JSONPath '$.verificationContext.previousVerificationId' could not be found" error in LAYOUT_VS_CHECKING workflow

## [0.3.0] - 2025-05-10

### Added
- Added MachineStructure struct definition to models.go
- Added comprehensive field definitions to LayoutMetadata struct
- Added comprehensive field definitions to HistoricalContext struct
- Added ImagesData struct for better image metadata organization

### Fixed
- Fixed "layoutMeta.ReferenceImageUrl undefined" error by adding missing fields
- Fixed "layoutMeta.SourceJsonUrl undefined" error by adding missing field
- Fixed "undefined: MachineStructure" error by defining the struct
- Fixed "layoutMeta.RowProductMapping undefined" error by adding field
- Fixed "historicalCtx.Summary undefined" error by adding field
- Removed unused imports to eliminate compiler warnings
- Fixed structure of verificationContext in response payload
- Improved metadata handling for DynamoDB records

### Changed
- Updated FetchImagesResponse to include structured verificationContext
- Enhanced DynamoDB attribute parsing for better error handling
- Restructured data models for more consistent API responses

## [0.2.2] - 2025-05-09

### Fixed
- Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error in InitializeLayoutChecking step
- Updated Step Function state machine definition to properly structure the verificationContext object for the Initialize Lambda

## [0.2.1] - 2025-05-09

### Fixed
- Fixed "verificationId is required" validation error when invoked from Step Functions
- Updated Step Function state machine definition to properly extract fields from verificationContext
- Ensured proper parameter passing between Step Function states

## [0.2.0] - 2025-05-09

### Added
- Enhanced input handling to support multiple invocation types:
  - Direct Step Function invocations
  - Function URL requests
  - Direct struct invocations
  - Fallback for other formats
- Improved error logging with detailed input capture for debugging

### Fixed
- Fixed "Invalid JSON input: unexpected end of JSON input" error when invoked from Step Functions
- Resolved input parsing issues between different invocation methods

## [0.1.0] - 2024-06-01

### Added
- Initial implementation of FetchImages Lambda:
  - Input validation
  - S3 metadata fetch (no image bytes or base64)
  - DynamoDB layout and historical context fetch
  - Parallel/concurrent fetch logic
  - Config via environment variables
  - Structured logging

### Changed
- N/A

### Removed
- Any base64 image handling (S3 URI only)
