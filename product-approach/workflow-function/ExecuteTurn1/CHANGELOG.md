# Changelog

All notable changes to this project are documented here.

---

## [4.0.6] - 2025-05-19

### Fixed

* **Critical Build & Deployment Fixes:**
  * Resolved compilation issues related to schema package integration
  * Fixed type errors with undefined schema imports:
    * Added missing StateReferences, HybridStorageConfig, StepFunctionInput, and StepFunctionOutput
    * Fixed references to undefined error types in bedrock client
    * Updated image handling to use HasBase64Data() instead of non-existent Base64 field

* **Infrastructure Improvements:**
  * Completely redesigned Dockerfile for more robust builds:
    * Added proper CA certificates installation
    * Improved working directory structure
    * Enhanced build verification with diagnostics
    * Added cache mounting for faster builds
  * Rebuilt docker-build script with better shared module handling:
    * Reliable shared package copying with proper Go module structure
    * Enhanced error detection and reporting
    * Improved temporary build context management
    * Better Lambda deployment integration

* **Bedrock Integration Fixes:**
  * Created proper adapter between shared/bedrock and internal/bedrock
  * Fixed message structure for Bedrock Converse API
  * Improved error handling for Bedrock client interactions
  * Corrected image data handling with proper Base64 processing

### Added

* **Enhanced Architecture:**
  * Added internal types package with properly defined interfaces
  * Created adapter pattern for better integration between shared and internal types
  * Improved type safety throughout codebase
  * Better separation of concerns with clear interfaces

### Technical Details

* Resolved compiler issues by creating local type definitions:
  * HybridStorageConfig with proper field structure
  * StepFunctionInput/Output types with correct field mapping
  * BedrockClient interface with proper adapter implementation
* Fixed dependency injection with proper type conversions
* Enhanced build system for more reliable Lambda deployments

---

## [2.0.0] - 2025-05-21

### Changed

* **MAJOR ARCHITECTURAL OVERHAUL - S3 State Reference Architecture:**
  * Complete transformation from payload-based to reference-based state management
  * Implemented shared S3StateManager pattern for lightweight API contracts
  * Replaced large workflow state payloads with S3 references between Lambda functions
  * Reorganized code into focused modules with single responsibilities

### Added

* **New Directory Structure:**
  * Redesigned with true separation of concerns and modular architecture:
    * `internal/state/` - State loading/saving with S3 reference architecture
    * `internal/bedrock/` - Bedrock client integration with shared package
    * `internal/validation/` - Input validation with schema validation
    * `internal/handler/` - Core business logic coordination
    * `internal/config/` - Enhanced configuration management

* **State Management Components:**
  * `StateLoader` for loading workflow state from S3 references
  * `StateSaver` for saving workflow state to S3 with category-based organization
  * Support for thinking content storage as separate artifacts
  * Hybrid Base64 image processing with improved performance characteristics

* **Enhanced Configuration:**
  * Standardized environment variable handling with sensible defaults
  * Improved categorization by functional area (S3, Bedrock, Images, Timeouts)
  * Comprehensive validation with detailed error messages
  * Better Bedrock configuration with proper regional settings

### Improved

* **Integration with Shared Packages:**
  * Full integration with `shared/s3state` package for state management
  * Complete adoption of `shared/bedrock` for standardized API integration
  * Focused error handling using `shared/errors` workflow error types
  * Streamlined logging with `shared/logger` for structured JSON logs

* **Error Handling:**
  * Consistent error creation and handling across all components
  * Better error classification and propagation with context preservation
  * Enhanced WorkflowError creation with detailed context
  * Improved retry logic with exponential backoff

* **Performance:**
  * Reduced memory usage by eliminating large in-memory payloads
  * Improved Lambda execution efficiency with smaller state transfers
  * Better resource utilization with S3 for large data storage
  * Optimized image processing with hybrid storage model

* **Maintainability:**
  * Clean, focused modules with single responsibilities
  * No file exceeds 300 lines of code
  * Comprehensive interfaces for better testability
  * Standardized error handling and logging patterns

### Technical Details

* Migrated from direct AWS SDK usage to shared Bedrock client
* Replaced custom validation with schema-based validation
* Implemented proper error typing and classification
* Enhanced logging with correlation IDs and structured context
* Streamlined configuration with sensible defaults
* S3 state storage with category-based organization
* Proper handler dependency injection for testability

---

## [1.4.0] - 2025-05-20

### Added

* **Major Refactoring - Models Package Restructure:**
  * Split large `internal/models/request.go` file into focused, single-responsibility modules
  * Created modular architecture with better separation of concerns:
    * `types.go` - Core type definitions and safe field access helpers
    * `parser.go` - JSON parsing with nil pointer protection
    * `validator.go` - Comprehensive validation logic
    * `sanitizer.go` - Data normalization and cleanup
    * `response_builder.go` - Response construction with error handling
    * `models.go` - Package API and convenience functions

### Fixed

* **CRITICAL FIX: Nil Pointer Dereference (Runtime Panic)**
  * Fixed runtime panic in `NewRequestFromJSON` at line 313 where code attempted to access `req.WorkflowState.VerificationContext.VerificationId` when `VerificationContext` was nil
  * Implemented immediate post-unmarshal validation to catch nil pointers before they cause panics
  * Replaced direct field access with safe helper methods throughout codebase
  * Added defensive programming patterns to prevent future nil pointer dereferences

* **Enhanced Error Handling:**
  * Fixed undefined constant `wferrors.ErrorTypeParsing` - replaced with `wferrors.ErrorTypeInternal`
  * Improved error type detection and logging in `determineErrorStatus` function
  * Added graceful handling of malformed JSON that unmarshals with nil nested structures

### Changed

* **Code Organization:**
  * Dramatically improved maintainability by splitting 800+ line file into 6 focused files (~150-200 lines each)
  * Better testability with isolated components
  * Clearer responsibilities and easier debugging
  * Enhanced code readability and navigation

* **Safety Improvements:**
  * All field access now goes through nil-safe helper methods
  * Added comprehensive debug information collection with `GetDebugInfo()`
  * Implemented post-unmarshal structure validation to prevent runtime panics
  * Enhanced logging with safe field access patterns

* **API Enhancements:**
  * Maintained full backward compatibility with existing API
  * Added new convenience functions: `ParseAndValidateRequest()`, `BuildResponse()`
  * Introduced optional `RequestProcessor` struct for stateful processing
  * Improved response building with better error state management

### Improved

* **Error Reporting:**
  * Better error context and structured error information
  * Enhanced debugging capabilities with detailed validation reporting
  * Improved error classification and status mapping
  * More granular error handling for different failure scenarios

* **Request Processing:**
  * Streamlined request processing pipeline with clear stages
  * Better sanitization of input data (schema version updates, timestamp normalization)
  * Automatic initialization of optional structures like `ConversationState`
  * String field cleanup and whitespace trimming

* **Response Handling:**
  * Robust response building that handles edge cases (nil state, nil errors)
  * Intelligent error-to-status mapping based on error types
  * Support for error responses with and without workflow state
  * Consistent error information attachment to state

### Technical Details

* **New File Structure:**
  ```
  internal/models/
  ├── types.go           # Core structs and safe helpers
  ├── parser.go          # JSON parsing with panic prevention
  ├── validator.go       # Request validation logic
  ├── sanitizer.go       # Data normalization
  ├── response_builder.go # Response construction
  └── models.go          # Package API
  ```

* **Safety Features:**
  * Nil-safe field access throughout (`GetVerificationID()`, `GetPromptID()`, etc.)
  * Post-unmarshal validation prevents accessing nil nested structures
  * Defensive programming patterns protect against malformed input
  * Comprehensive error logging with safe operations

* **Backward Compatibility:**
  * All existing function signatures maintained
  * Drop-in replacement for original `request.go`
  * No breaking changes to external API
  * Legacy functions preserved with deprecation notices

---

## [1.3.5] - 2025-05-20

### Fixed

* **Schema Compatibility Fixes:**
  * Fixed numerous field access errors related to undefined fields in schema types
  * Resolved `state.BedrockConfig.ModelId undefined` error by properly accessing model ID from environment
  * Fixed references to non-existent `Metadata` fields in WorkflowState and ConversationState
  * Replaced undefined `schema.StatusInProgress` with a plain string "IN_PROGRESS"
  * Fixed references to undefined status constants (`StatusTurn1Failed`, `StatusTurn2Failed`)
  * Removed unused `time` import from response_processor.go

### Changed

* **Handler Structure:**
  * Modified Handler struct to store modelId from environment configuration
  * Updated NewHandler constructor to accept modelId parameter
  * Added getModelId helper method in bedrock_client.go to ensure proper model selection
  * Replaced direct schema field access with safe alternatives where fields were missing

### Improved

* **Validation Logic:**
  * Enhanced BedrockConfig validation to skip missing ModelId field
  * Improved error state checking with string literals when schema constants not available
  * Better error messages for debugging schema compatibility issues
  * Removed references to non-existent fields in logging statements

---

## [1.3.4] - 2025-05-19

### Changed

* **Major Code Refactoring:**
  * Split large `execute_turn1.go` file (589 lines) into multiple smaller files with clear separation of concerns
  * Created dedicated modules for validation, image processing, Bedrock client, error handling, and state management
  * Improved code organization without changing functionality
  * Enhanced maintainability and troubleshooting capabilities
  
  * **New File Structure:**
    * `handler.go` - Core handler structure and main request handling flow
    * `validation.go` - All validation-related functions
    * `image_processor.go` - Image processing functionality
    * `bedrock_client.go` - Bedrock API interaction
    * `error_handler.go` - Error handling utilities
    * `state_manager.go` - State management functions
    * `response_processor.go` - Response processing functionality (existing)

### Improved

* **Code Organization:**
  * Better file structure with focused responsibilities
  * Simplified navigation and readability
  * Easier to locate and fix issues
  * Improved testability of individual components

---

## [1.3.3] - 2025-05-18

### Fixed

* **CRITICAL FIX**: Resolved "ValidationException: Invalid CurrentPrompt" error by reordering validation steps
* Fixed validation logic to validate prompt structure before requiring images
* Corrected validation sequence: core validation → Base64 generation → complete validation including images
* Fixed `ValidateCurrentPrompt` calls to use `false` for image requirement during initial validation

### Changed

* **Validation Flow Overhaul:**
  * Split validation into two phases: core validation (without images) and complete validation (with images)
  * Enhanced validation error messages with more context and debugging information
  * Added defensive programming with comprehensive nil checks for critical fields
* **Code Organization:**
  * Refactored `execute_turn1.go` into smaller, focused methods for better maintainability
  * Centralized error handling and logging in `createAndLogError` method
  * Improved code readability with clear separation of concerns
* **Response Processing:**
  * Updated `response_processor.go` to work seamlessly with the new structure
  * Enhanced thinking content extraction with alternative pattern matching
  * Added advanced token usage analysis and response structure validation
* **Request Handling:**
  * Complete rewrite of `request.go` with enhanced validation and sanitization
  * Added `RequestValidator` struct for better organization
  * Implemented comprehensive field validation with proper error context
  * Added helper methods for safe field access and debugging

### Added

* Enhanced error logging with structured context for better debugging
* Added `ValidateAndSanitize` method for combined request processing
* Implemented advanced thinking content extraction with multiple pattern support
* Added token usage analysis and efficiency metrics in ResponseProcessor
* Created helper methods: `GetVerificationID()`, `GetPromptID()`, `HasImages()`
* Added `String()` methods for better debugging output
* Implemented comprehensive response structure validation
* Added processing statistics and metadata tracking

### Improved

* **Error Handling:**
  * Better error classification and context preservation
  * Enhanced WorkflowError wrapping with additional context
  * Improved retry logic with exponential backoff and context cancellation support
* **Logging:**
  * Added debug-level logs for detailed execution tracing
  * Structured logging with correlation IDs and relevant context
  * Performance metrics logging (latency, token usage, etc.)
* **Validation:**
  * More comprehensive field validation with specific error messages
  * Better handling of edge cases and malformed input
  * Improved schema version handling and automatic updates

### Technical Details

* Validation order: `validateCoreWorkflowState` → `generateBase64Images` → `validateCompleteWorkflowState`
* Enhanced retry logic with proper AWS SDK error type checking
* Improved Base64 hybrid storage integration with better error handling
* Advanced response parsing with thinking extraction and token analysis

---

## [1.3.2] - 2025-05-18

### Fixed

* Fixed s3.Client type error in NewHandler function by using the correct parameter name
* Added missing errors import and renamed to wferrors to avoid name collision
* Fixed invalid use of respOut.Body as io.Reader by using the bytes directly
* Fixed access to undefined field Identifier in ImageInfo by using URL instead
* Fixed invalid Body field in InvokeModelInput struct literal
* Fixed the mergeTokenUsage function to handle absence of token fields in BedrockApiResponse
* Added missing aws import in dependencies/clients.go
* Fixed code to compile cleanly with Go 1.22

## [1.3.1] - 2025-05-18

### Changed

* Config: removed hardcoded defaults; require environment variables `AWS_REGION`, `BEDROCK_MODEL`, `ANTHROPIC_VERSION`, and validate at startup.
* ExecuteTurn1: generate Base64 images before full validation to prevent INVALID_IMAGE_DATA errors.
* ExecuteTurn1: use `BedrockModelID` from environment for `ModelId` in Bedrock requests.
* ExecuteTurn1: fixed image payload schema to `{ "image": { "format": ..., "source": { "bytes": ... } } }`.
* ExecuteTurn1: added retry logic with exponential back-off for `ServiceException` and `ThrottlingException`.
* clients.go: apply configurable Bedrock timeout from `BEDROCK_TIMEOUT` env var.
* request.go: delay image validation until after Base64 retrieval.
* config.go: removed hardcoded defaults; validate all critical env vars.
* response_processor.go: simplified token usage handling to align with `schema.BedrockApiResponse`.

### Fixed

* Corrected order of Base64 generation and image validation in `execute_turn1.go`.
* Fixed ModelId reference to use `BEDROCK_MODEL` env var.
* Fixed image block construction to comply with Bedrock Converse API schema.
* Ensured `error` import from Go std to properly use `errors.As`.
* Removed misuse of `io.ReadAll` and `Close` on byte slices.

---

## [1.3.0] - 2025-05-18

### Changed

* **API Migration:**
  * Migrated from Bedrock Converse API to InvokeModel API for improved compatibility with AWS SDK
  * Restructured request/response handling to work with the newer API contract
  * Fixed type safety issues with error handling and schema compatibility
* **Architecture Support:**
  * Migrated to ARM64/Graviton Lambda for improved cost and performance
  * Updated Go version from 1.21 to 1.22
  * Added platform-specific build flags

### Added

* Enhanced build process with optimized Go compiler flags
* Better error diagnostics in the Docker build script
* Platform detection in the build script for cross-compilation
* Updated documentation to reflect API changes

---

## [1.2.0] - 2025-05-17

### Changed

* **Major migration:**
  * All workflow and Lambda modules now fully use shared `schema`, `logger`, and `errors` packages.
  * Removed all custom/request-specific type definitions for workflow state, prompt, images, and error contracts.
* **Centralized AWS client initialization and config validation.**
* **Bedrock integration:**
  * Now uses the Claude 3.7 Sonnet Converse API exclusively.
  * Bedrock messages are always constructed via the shared schema, supporting inline and S3-based Base64 image references.
* **Hybrid Base64 support:**
  * Large images automatically use S3 temporary storage; retrieval and embedding for Bedrock handled via schema helpers.
* **Logging overhaul:**
  * Replaced all raw log output with JSON-structured logs via the shared logger.
  * Correlation IDs and context fields now populate all logs for distributed traceability.
* **Error handling overhaul:**
  * All errors are now `WorkflowError` types (typed, retryable, and with full context).
  * All errors surfaced in both Lambda error and within `VerificationContext.Error` for Step Functions compatibility.

### Added

* Helper utilities in `request.go` and `response_processor.go` for request/response validation and workflow state management.
* Standardized error creation and propagation in all modules.
* Docs/README updated for new architecture.

---

## [1.1.0] - 2025-05-15

### Added

* Initial implementation of shared schema, logger, and error packages.
* Core support for hybrid Base64 image storage (inline/S3).
* Validation and builder helpers for all major workflow state fields.

---

## [1.0.0] - 2025-05-14

### Added

* First release of vending machine verification solution.
* Lambda function skeletons and event-driven Step Functions integration.