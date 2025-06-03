# Changelog

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

