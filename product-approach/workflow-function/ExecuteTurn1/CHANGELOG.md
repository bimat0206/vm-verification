# Changelog

All notable changes to this project are documented here.

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
  
  * **New File Structure
* handler.go

Core handler structure and main request handling flow
Contains the main HandleRequest method that orchestrates the workflow

* validation.go

All validation-related functions
Handles workflow state validation at different stages
* image_processor.go

Image processing functionality
Manages Base64 encoding and image content preparation
* bedrock_client.go

Bedrock API interaction
Handles request building, API calls with retry logic, and response parsing
* error_handler.go

Error handling utilities
Centralizes error creation, logging, and state updates
* state_manager.go

State management functions
Handles workflow state updates and conversation state management
* response_processor.go (existing)

Response processing functionality
Already had good separation of concerns

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