# Changelog

## [1.3.0] - 2025-01-04 - Input Structure Support & S3 Error Resolution

### Fixed
- **S3Exception Resolution**: Fixed "failed to load initialization data" S3Exception error
  - Root cause: Function couldn't parse the expected nested s3References input structure
  - Added comprehensive input parsing to handle Step Functions state format correctly
  - Enhanced reference extraction for both flat and nested s3References structures

- **Input Structure Compatibility**: Complete rewrite of input handling logic
  - Implemented `parseInputAndExtractReferences()` function for robust input parsing
  - Added support for multiple input types: map[string]interface{}, JSON string, byte array
  - Enhanced `extractNestedReference()` to handle `s3References.responses.turn2Processed` structure
  - Added `extractReferenceFromMap()` helper for top-level reference extraction

### Enhanced
- **Error Handling & Logging**: Improved debugging capabilities
  - Added detailed logging of extracted references (bucket, key, size)
  - Enhanced error messages with expected structure information
  - Improved validation error reporting for missing references
  - Added comprehensive input validation before processing

- **Documentation**: Updated README with clear input structure examples
  - Added JSON example showing expected nested s3References format
  - Documented processing workflow and input requirements
  - Provided clear structure documentation for integration

### Technical Details
- **Expected Input Format**: Now fully supports the standard Step Functions state format:
  ```json
  {
    "s3References": {
      "processing_initialization": {"bucket": "...", "key": "..."},
      "responses": {
        "turn2Processed": {"bucket": "...", "key": "...", "size": 151}
      }
    },
    "verificationId": "verif-20250605025241-3145",
    "status": "TURN2_COMPLETED"
  }
  ```
- **Backward Compatibility**: Maintains compatibility with existing error handling patterns
- **Robust Parsing**: Handles edge cases and provides clear error messages for debugging

## [1.2.1] - 2025-06-01 - Reference Lookup Fix

### Fixed
- **Turn2Processed Reference Lookup**: Fixed "Missing required field: turn2Processed reference" error
  - Implemented `extractNestedReference()` function to handle nested JSON structures from Step Functions state
  - Now correctly extracts references from nested `s3References.responses.turn2Processed` structure
  - Enhanced error logging to show expected structure for better debugging

### Technical Details
- **Root Cause**: Function was looking for flat `turn2Processed` key but Step Functions provides nested structure `s3References.responses.turn2Processed`
- **Solution**: Added nested structure extraction to handle standard Step Functions state format
- **Standard Format**: `{"s3References": {"responses": {"turn2Processed": {...}}}}`

## [1.2.0] - 2025-01-03 - Maintenance & Shared Component Updates

### Changed
- **Shared Component Compatibility**: Updated to maintain compatibility with latest shared component versions
- **Dependency Alignment**: Ensured compatibility with updated logger, s3state, and schema packages

### Technical Details
- **Shared Dependencies**: Compatible with shared components v2.2.0 with enhanced error handling and logging
- **No Functional Changes**: Core finalization functionality remains unchanged, updates are for compatibility only

### Notes
- This release maintains backward compatibility while ensuring integration with updated shared components
- No configuration changes required for existing deployments

## [1.1.0] - 2025-01-03
### Changed
- **Major Refactoring**: Migrated to shared packages for standardization and consistency
  - Replaced custom input/output handling with s3state envelope pattern
  - Updated function signature to use `interface{}` input with envelope loading
  - Implemented envelope-based input validation using `s3state.LoadEnvelope()`
  - Added proper S3 reference validation for initialization and turn2 processed data

- **S3 Operations Modernization**: Replaced custom S3 operations with shared s3state manager
  - Removed `internal/s3helper` package entirely (~50 lines of code eliminated)
  - Updated S3 data loading to use `stateManager.RetrieveJSON()` and `stateManager.Retrieve()`
  - Added s3state manager initialization in init function
  - Implemented envelope-based S3 state management for better traceability

- **Error Handling Standardization**: Replaced simple error returns with structured error handling
  - Integrated `shared/errors.WorkflowError` with proper error categorization
  - Added error types: `ErrorTypeS3`, `ErrorTypeDynamoDB`, validation, and parsing errors
  - Enhanced error context with verification IDs and operation details
  - Updated DynamoDB helper to use shared error handling patterns

- **Schema Integration**: Adopted shared schema types and constants
  - Replaced custom status constants with `schema.StatusCompleted`
  - Updated timestamp formatting to use `schema.FormatISO8601()`
  - Implemented envelope-based output instead of custom output structures
  - Added summary information to envelope for better traceability

### Added
- **Dependencies**: Added shared package dependencies to go.mod
  - `workflow-function/shared/schema`
  - `workflow-function/shared/s3state`
  - `workflow-function/shared/errors`

### Removed
- **Legacy Code**: Eliminated custom implementations in favor of shared packages
  - Removed `internal/s3helper` package
  - Removed custom error handling patterns
  - Removed custom input/output structures

### Fixed
- **Docker Build**: Fixed Go version compatibility and dependency management
  - Removed `toolchain go1.24.0` directive from go.mod files
  - Updated build script to copy existing go.mod instead of hardcoding dependencies
  - Fixed shared module path resolution in Docker build context
  - Verified successful compilation with shared packages

## [1.0.1] - 2025-01-03
### Added
- **Docker Containerization**: Added complete Docker containerization setup
  - Multi-stage Dockerfile optimized for AWS Lambda ARM64 deployment
  - Automated build script with ECR integration and Lambda deployment
  - Standardized build process consistent with other workflow functions
- **Build Automation**: Enhanced deployment capabilities
  - Cross-compilation support for AWS Lambda (linux/arm64)
  - Shared module handling via temporary build contexts
  - Environment variable configuration support
  - Command-line argument parsing for flexible deployment

## [0.1.0] - 2025-06-03
### Added
- Initial implementation of `FinalizeAndStoreResults` Lambda function.
- Parses Turn 2 processed results from S3 and writes final verification record to DynamoDB.
- Updates conversation history status to `WORKFLOW_COMPLETED`.

