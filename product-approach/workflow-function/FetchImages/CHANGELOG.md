# Changelog

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