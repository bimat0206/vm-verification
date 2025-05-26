# Changelog

All notable changes to the ExecuteTurn1Combined function will be documented in this file.

## [2.4.6] - 2025-06-02
### Fixed
- Corrected the S3 path for the processed Turn 1 Markdown response. The file is now stored under the `response/` directory instead of `processing/`.

## [2.4.2] - 2025-05-26
### Fixed
- Restored `ParsedTurn1Data` type to resolve compilation errors.
- Removed redundant `nil` check when updating metrics in DynamoDB service.

## [2.4.3] - 2025-05-27
### Fixed
- **Vietnamese Character Encoding:** Ensured UTF-8 preservation when storing the processed Turn 1 Markdown in S3. The `turn1-processed-response.md` object now sets `ContentType` to `text/markdown; charset=utf-8` for proper Unicode display.

## [2.4.4] - 2025-05-31
### Changed
- Layout dimensions (`RowCount`, `ColumnCount`) are now loaded from layout metadata instead of using hard-coded defaults.
- Simplified `VerificationContext.Validate` to only check for presence of `LayoutMetadata` or `HistoricalContext` as required.

## [2.4.5] - 2025-06-01
### Changed
- Modified the structure of the `s3References` field in the Step Function output
  to align with downstream expectations.
  - `initialization` renamed to `processing_initialization`.
  - `images.metadata` flattened to `images_metadata`.
  - `processing.layoutMetadata` flattened to `processing_layout-metadata`.
  - `processing.historicalContext` now outputs as `processing_historical-context`
    only when `verificationType` is `PREVIOUS_VS_CURRENT`.
- Ensured all generated S3 reference keys include the date-partition prefix so
  paths are consistent across artifacts.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.4.0] - 2025-05-25
### Changed
- **Refactor: Adopt "Markdown-Centric" Approach for Turn 1 Bedrock Response Processing (POC Simplification).**
  - Simplified `internal/bedrockparser` to return cleaned Markdown via `ParseBedrockResponseAsMarkdown`.
  - `StorageManager` now stores the Markdown output in `processing/turn1-processed-response.md` with `text/markdown` content type.
  - DynamoDB `VerificationResults` records now reference this Markdown file using `processedTurn1MarkdownRef` instead of storing deeply parsed fields.
  - Added `UpdateTurn1CompletionDetails` in DynamoDB service and updated handler logic accordingly.

## [2.3.0] - 2025-05-25
### Added
- **New `internal/bedrockparser` package:** Introduced parser for Bedrock Turn 1 response.
### Changed
- Turn 1 response processing parses and stores structured data in S3 `turn1-processed-response.json`.
- DynamoDB now stores parsed Turn 1 details for frontend display.

## [2.2.6] - 2025-05-31

### Changed
- **Historical Context Loading Refactor:**
  - Historical context for `PREVIOUS_VS_CURRENT` verifications is now loaded from the provided S3 reference `processing_historical-context`.
  - Removed direct DynamoDB queries for previous verification data in `HistoricalContextLoader`.
  - `EventTransformer` now passes the historical context S3 reference through to the `Turn1Request` structure.
  - `Turn1RequestS3Refs` updated to include `Processing` references.
  - `NewHistoricalContextLoader` accepts an `S3StateManager` instance for S3 operations.

## [2.2.5] - 2025-05-30

### Fixed
- **DynamoDB Update Failures:**
    - Added `verificationAt` sort key to `UpdateVerificationStatusEnhanced` so updates correctly target existing `VerificationResults` records. Requires passing the initial timestamp through `DynamoManager.Update`.
    - Corrected `UpdateConversationTurn` query to use `verificationId` as the partition key for `ConversationHistory`.
    - Corrected `updateExistingConversationHistory` and `CompleteConversation` to use `verificationId` in DynamoDB keys.
    - Modified `RecordConversationHistory` to construct items using `verificationId` as the partition key attribute.
    - `CompleteConversation` now requires `conversationAt` (sort key) to correctly target existing records.

## [2.2.4] - 2025-05-29

### Fixed
- **Prompt Processing Time:** `PromptGenerator` now records actual template processing duration instead of a placeholder value.
- **Historical Context Machine Structure:** `HistoricalContextLoader` dynamically loads machine structure details from previous verifications or associated layout metadata rather than using hardcoded defaults. Warnings are logged when structure details cannot be determined.

### Changed
- Added contextual logging to indicate the source of historical machine structure data.

## [2.2.3] - 2025-05-28

### Fixed
- **DynamoDB Flag Accuracy**: `summary.dynamodbUpdated` now reflects the real
  outcome of DynamoDB writes.

## [2.2.2] - 2025-05-24

### Fixed
- **DynamoDB Conversation History Update**: Fixed error when updating conversation history by using Query operation instead of GetItem
  - Conversation table uses composite key (conversationId + conversationAt) but code was only using partition key
  - Changed from GetItem to Query to find the most recent conversation record
  - Preserves existing conversationAt timestamp when updating existing records
  - Properly handles both new conversation creation and existing conversation updates

## [2.2.1] - 2025-05-24

### Fixed
- **Nil Pointer Dereference**: Fixed runtime panic in `HandleForStepFunction` by properly initializing `ProcessingStagesTracker` and `StatusTracker` before use

## [2.2.0] - 2025-05-24

### Added
- **StepFunctionResponse struct**: New response structure in `models/shared_types.go` that matches the expected Step Function output format
- **HandleForStepFunction method**: New handler method that specifically returns `StepFunctionResponse` for Step Function invocations
- **BuildStepFunctionResponse method**: New response builder method to construct the simplified Step Function output

### Changed
- **Response Format**: Step Functions now receive a simplified output structure with just `verificationId`, `s3References`, `status`, and `summary` fields
- **S3 References Structure**: Aligned S3 references format with expected output, including proper nesting for initialization, images, processing, prompts_system, and responses
- **Bedrock Latency**: Now uses actual Bedrock invocation duration instead of hardcoded 2000ms value

### Fixed
- **Output Structure Mismatch**: Resolved issue where ExecuteTurn1Combined was returning a complex nested structure instead of the simplified format expected by Step Functions
- **Method Signatures**: Fixed compilation errors related to mismatched method signatures and return types

### Technical Details
- Modified `handleStepFunctionEvent` to use the new `HandleForStepFunction` method
- Updated `buildSummary` to accept actual Bedrock latency as a parameter
- Added `GetTemplateUsed` method to PromptGenerator for determining the template based on verification type
- Fixed validator method call from `ValidateInput` to `ValidateRequest`
- Fixed storage manager method call from `StoreResponse` to `StoreResponses`

## [2.1.4] - 2025-05-27

### Fixed - Step Function Output
- **Simplified Response**: `handleStepFunctionEvent` now returns a concise payload
  `{verificationId, s3References, status, summary}` for Step Functions.
- **Impact**: Downstream states receive consistent data across workflow functions.


## [2.1.3] - 2025-05-27

### Fixed - DynamoDB Update Expression
- **Status History Append**: Replaced invalid `ADD` operation with `list_append` using `if_not_exists`.
- **Impact**: Prevents `ValidationException` errors when updating verification status.


## [2.1.2] - 2025-05-26
Reworked StorageManager to build structured prompt data and raw response payloads that include context sources, image references, and Bedrock metadata

Updated handler logic to use the new storage functions and record storage metadata without passing raw size explicitly

## [2.1.1] - 2025-05-26

### Fixed - S3 Path Handling
- **Date-Based Paths**: Added date partition prefix when storing workflow state and reference metadata
- **Reference Tree**: buildS3RefTree now preserves date prefix from raw response key

### Added
- **Timezone Validation**: Configuration validates `DATE_PARTITION_TIMEZONE`


## [2.1.0] - 2025-05-25

### Fixed - Schema v2.1.0 Compliance
- **ResponseBuilder**: Now outputs full v2.1.0 schema
  - Added missing fields: prompt, S3 refs, summary, status, schemaVersion
  - Implemented buildS3RefTree for unified S3 reference structure
  - Added buildSummary for execution metrics
  - Included all required fields for validator compliance

### Fixed - Prompt S3 References
- **StorePrompt Method**: Fixed to use correct key format
  - Now stores JSON: { "prompt": "<renderedPrompt>" } under prompts/turn1-prompt.json
  - Provides S3Reference for response builder
  - Enables prompt retrieval for debugging and analysis
  - Fixed missing S3References.prompts.turn1Prompt in response

### Fixed - Schema Version Drift
- **Schema Version**: Updated all references to use "2.1.0"
  - Fixed schema version in contextEnrichment
  - Updated main.go to use "2.1.0" instead of "2.0"
  - Ensured consistent schema version across all components

### Fixed - DynamoDB Async Errors
- **Enhanced DynamoDB Error Handling**: Added dynamoOK flag
  - Prevents Lambda failure on DynamoDB throttling
  - Updates summary.DynamodbUpdated to reflect actual status
  - Logs warnings but allows function to complete successfully
  - Added tests to verify throttling behavior

### Improved - Code Organization
- **Unified S3 Reference Handling**: Single buildS3RefTree helper
  - Consolidated all S3 reference building logic
  - Implemented consistent reference structure
  - Removed duplicate code paths
  - Added conditional inclusion of promptRef when key is not empty

### Technical Details
- **Schema Compliance**: All responses now pass ValidateTurn1CombinedResponse
- **DynamoDB Integration**: Tables now accept enhanced schema with history arrays
- **Code Quality**: Reduced duplication, improved maintainability
- **Testing**: Added validation for schema version and S3 references


## [2.8.0] - 2025-05-23

### Fixed - Critical Bedrock API Image Encoding Issue
- **Root Cause**: Shared bedrock client was incorrectly converting base64 strings to bytes using `[]byte()` instead of proper base64 decoding
- **Impact**: All image-based verifications failing with "Invalid image input" errors from Bedrock API
- **Resolution**: Fixed shared/bedrock/client.go to properly decode base64 strings using `base64.StdEncoding.DecodeString()`
- **Verification**: ExecuteTurn1Combined now successfully processes images matching ExecuteTurn1 behavior

### Changed - Code Simplification and Cleanup
- **Removed Unnecessary Complexity**: Cleaned up overly verbose architectural comments and jargon
  - Simplified adapter.go, client.go, and services/bedrock.go comments
  - Removed architectural pattern descriptions that added no value
  - Maintained clear, concise documentation focused on functionality
- **Removed Unused Code**: 
  - Deleted empty `types.go` file from internal/bedrock package
  - Removed unused metrics and logging fields
  - Cleaned up unnecessary function like `GetAdapterMetrics()`
- **Improved Code Readability**: 
  - Simplified logging statements by removing redundant fields
  - Streamlined error handling with clearer messages
  - Maintained functionality while improving maintainability

### Technical Details
- **Bug Location**: `/shared/bedrock/client.go` lines 93-96
- **Fix Applied**: Added proper base64 decoding before passing to AWS SDK
- **Testing**: Build successful, ready for integration testing
- **Compatibility**: Maintains full compatibility with ExecuteTurn1 patterns

## [2.7.0] - 2025-05-23

### Changed - Strategic Bedrock Architecture Refactoring
- **Local Bedrock Package Implementation**: Created deterministic control architecture for Bedrock integration
  - Implemented `internal/bedrock` package with adapter pattern for zero-copy translation
  - Created `Client` with `ProcessTurn1` method providing direct base64 pass-through
  - Added `Adapter` for strategic impedance matching between local and shared types
  - Established clear separation between domain control and infrastructure
  
### Added - Adapter Pattern Architecture
- **Zero-Copy Translation Layer**: Implemented adapter pattern for type mapping without data transformation
  - `adapter.go`: Bridges local domain types with shared infrastructure types
  - `client.go`: Orchestrates Bedrock operations with timeout management
  - `config.go`: Defines comprehensive configuration with all Bedrock parameters
  - Direct delegation pattern eliminates preprocessing overhead
  
### Refactored - Service Layer Simplification
- **BedrockService Refactoring**: Migrated to local client pattern with simplified implementation
  - Removed duplicate functions: `buildMessages`, `detectImageFormat`, `transformResponse`
  - Replaced complex error classification with direct error propagation
  - Added `NewBedrockServiceWithLocalClient` for explicit local client usage
  - Eliminated preprocessing logic in favor of direct delegation
  
### Improved - Code Quality and Maintainability
- **Duplicate Code Elimination**: Removed all redundant implementations
  - Consolidated image format detection to adapter layer
  - Removed unused type definitions from `types.go`
  - Eliminated duplicate validation and transformation logic
  - Cleaned up imports and removed unused dependencies
  
### Technical Details
- **Architecture Pattern**: Adapter pattern with deterministic control flow
- **Performance**: Zero-copy data flow from service to Bedrock API
- **Type Safety**: Maintained strong typing with proper interface boundaries
- **Backward Compatibility**: Service interface remains unchanged
- **Testing**: Added unit tests for client validation and adapter functionality

## [2.6.0] - 2025-05-23

### Fixed - Critical Base64 Data Transformation Architecture Alignment
- **Root Cause**: Misalignment between service implementation and shared bedrock package expectations
- **Previous Issue**: Service was decoding Base64 to raw bytes, but shared package expects Base64 strings
- **Resolution**: Corrected data transformation pipeline to preserve Base64 format for shared package
- **Architecture**: Aligned with ExecuteTurn1 reference pattern - Base64 strings flow through to shared package

### Changed - Data Transformation Pipeline
- **Validation-Only Decoding**: Base64 decoding now used solely for validation and format detection
  - Renamed `decodeBase64Image` to `decodeBase64ForValidation` to clarify purpose
  - Decoded bytes used for size validation and image format detection only
  - Original Base64 string passed directly to `bedrock.CreateImageContentFromBytes()`
- **Architectural Clarity**: Established clear data flow pattern:
  - S3 Storage → Base64 String (preserved through service layer)
  - Validation → Temporary decode for size/format checks
  - Shared Package → Receives Base64 string (NOT raw bytes)
  - AWS SDK → Handles final encoding automatically

### Added - Comprehensive Observability
- **Transformation Boundary Logging**: Strategic logging at each data transformation point
  - Pipeline entry with input metrics (Base64 length, prompt sizes)
  - Pre/post validation decode boundaries with metrics
  - Format detection results with detected type
  - Shared package handoff confirmation
  - API invocation start/complete with latency metrics
  - Operation completion with total time and token usage
- **Enhanced Error Context**: Detailed logging for debugging and monitoring
  - Base64 validation failures with input length
  - Decoding errors with transformation metrics
  - Size limit violations with excess calculations
  - API errors with latency and retry recommendations

### Technical Details
- **Performance Metrics**: Added comprehensive latency tracking
  - Decoding operation latency in microseconds
  - API call latency in milliseconds
  - Total operation time tracking
  - Compression ratio calculations (decoded vs encoded size)
- **Logger Integration**: Added structured logging with `workflow-function/shared/logger`
  - Logger instance added to bedrockService struct
  - All operations emit structured JSON logs
  - Correlation ID support for distributed tracing
- **Error Classification**: Enhanced error handling with architectural context
  - Validation phase errors marked as non-retryable
  - Size limit errors include mitigation strategies
  - API errors classified by type (throttling, timeout, content policy)

## [2.5.0] - 2025-05-23

### Fixed - Critical Base64 Encoding Architecture Mismatch
- **Root Cause**: AWS SDK expects raw bytes but was receiving Base64-encoded strings
- **Impact**: All image-based verifications failing with "Invalid image input" errors
- **Resolution**: Implemented Base64 decoding at Bedrock service boundary
- **Architecture**: Established clear data format contracts across service boundaries

## [2.4.1] - 2025-05-23

### Fixed - Template Rendering Production Crash
- **Missing Template Function**: Fixed "function 'printf' not defined" error in production
  - Added `printf` function to shared templateloader's DefaultFunctions map
  - Function was missing despite being used in templates for formatting (e.g., `{{printf "%.1f" .HoursSinceLastVerification}}`)
  - Templates now render successfully in production matching test behavior

- **Context Field Initialization**: Hardened template context building to prevent nil pointer errors
  - Added safe field extraction helpers with proper type checking and defaults
  - Ensure RowCount/ColumnCount default to 6/10 if missing from metadata
  - Generate RowLabels (A, B, C...) automatically if not provided
  - Initialize VerificationSummary with zero values instead of nil for PREVIOUS_VS_CURRENT
  - Added VendingMachineID field (uppercase) for template compatibility

- **VerificationContext Validation**: Added pre-render validation to ensure required fields
  - Created `Validate()` method on VerificationContext model
  - Automatically initializes missing fields based on verification type
  - Prevents template execution errors from missing context data
  - Called before template rendering in prompt service

- **Enhanced Error Logging**: Improved error visibility for template failures
  - Added detailed error classification with original error messages
  - Enhanced logging to include template type, version, and error context
  - Better detection of missing function errors for faster debugging
  - Removed duplicate error handling code in prompt service

### Technical Details
- **Root Cause**: Production templateloader was using default Go template functions only
- **Testing Gap**: Unit tests registered a richer FuncMap that masked the production issue
- **Solution**: Ensured production and test environments use identical template functions
- **Validation**: All template tests pass including new integration tests

## [2.4.0] - 2025-05-23

### Changed - S3 Loader Simplification for Current-Truth Formats
- **System Prompt Loading**: Updated to parse new rich JSON format
  - Now expects `{ "promptContent": { "systemMessage": "..." } }` structure
  - Extracts `systemMessage` field from nested JSON object
  - Returns `WorkflowError{Type:"ValidationException"}` for invalid JSON format
  - Added DEBUG logging showing `"format":"rich"` for operational visibility
  
- **Base64 Image Loading**: Simplified to read plain .base64 files
  - Now validates files must end with `.base64` extension
  - Reads raw base64 text directly without JSON parsing
  - Trims whitespace from content automatically
  - Returns `WorkflowError{Type:"ValidationException"}` for invalid formats
  - Added DEBUG logging showing `"format":".base64"` for tracking
  
### Added - Enhanced Error Handling
- **Format Validation Errors**: Specific error codes for format issues
  - `BadSystemPrompt` - Invalid JSON structure in system prompt
  - `MissingSystemMessage` - Empty or missing systemMessage field
  - `ExpectBase64Ext` - Image file doesn't have .base64 extension
  - `EmptyBase64` - Base64 file is empty after trimming
  - `ReadFailed` - S3 read operation failed (retryable)
  
### Added - Comprehensive Testing
- **Unit Tests**: Added test coverage for new loader implementations
  - System prompt parsing with valid/invalid JSON scenarios
  - Base64 file reading with extension validation
  - Error handling verification for all failure modes
  - Whitespace trimming validation for base64 content
  - All tests pass with proper error classification

### Technical Details
- **Breaking Change**: Removed support for legacy formats (simple strings, `{ "data": ... }`)
- **S3 Operations**: Uses raw byte retrieval (`Retrieve`) instead of JSON unmarshaling for images
- **Error Types**: Format errors use `ValidationException`, S3 errors use `S3Exception`
- **Backward Compatibility**: None - this is a breaking change for current-truth format adoption

## [2.3.0] - 2025-05-23

### Added - Bedrock Timeout Configuration
- **Configurable Timeouts**: Added environment-driven timeout configuration for Bedrock API calls
  - `BEDROCK_CONNECT_TIMEOUT_SEC` - Connection establishment timeout (default: 10 seconds)
  - `BEDROCK_CALL_TIMEOUT_SEC` - API call timeout (default: 30 seconds)
  - Timeouts are applied using `context.WithTimeout` for proper cancellation support
  
- **Configuration Validation**: Added comprehensive timeout validation
  - Created `internal/config/validate.go` with `Validate()` method
  - Ensures both timeouts are greater than 0
  - Ensures call timeout is greater than connect timeout
  - Returns `WorkflowError{Type:"Config", Code:"BedrockTimeoutInvalid"}` for invalid configurations
  
- **Timeout Error Handling**: Enhanced Bedrock error classification for timeout scenarios
  - Timeout errors return `BedrockTimeout` error code (non-retryable)
  - Proper detection of "context deadline exceeded" errors
  - Enriched error context includes timeout duration for debugging
  - Step Functions will route timeout errors to `FinalizeWithError`
  
- **Bootstrap Logging**: Added initialization logging for timeout configuration
  - Logs `connectTimeoutMs` and `callTimeoutMs` during cold start
  - Provides operational visibility into active timeout settings
  - Included in `bedrock_client_init` log entry
  
- **Comprehensive Testing**: Added unit tests for timeout functionality
  - Config validation tests covering all edge cases
  - Service-level timeout error classification tests
  - Integration tests for timeout configuration loading
  - All tests pass with 100% coverage

### Technical Details
- **Implementation Location**: Timeout applied in `internal/services/bedrock.go` Converse method
- **Backward Compatibility**: Maintains existing single timeout field while adding new granular controls
- **Operational Benefits**: Prevents Lambda hanging on stalled Bedrock calls, improves latency metrics
- **Future Consistency**: Pattern ready for adoption by other Lambdas (Turn-2, Finalize)

## [2.2.0] - 2025-05-23

### Changed - Major Code Refactoring for Better Maintainability
- **Handler Refactoring**: Split large handler.go (1000+ lines) into 14 focused, single-responsibility files
  - Reduced main handler.go to just 193 lines - a clean orchestrator
  - Created dedicated components for each major responsibility
  - Improved code organization with clear separation of concerns
  - Enhanced testability through isolated components

### Added - New Component Architecture
- **Core Handler Components**:
  - `handler.go` (193 lines) - Main orchestrator that coordinates all components
  - `handler_helpers.go` - Helper methods for handler operations
  
- **Processing & Tracking Components**:
  - `processing_stages.go` - Tracks workflow processing stages with metadata
  - `status_tracker.go` - Manages status updates and history tracking
  - `response_builder.go` - Builds combined turn responses with all metadata
  
- **Data Loading Components**:
  - `context_loader.go` - Handles concurrent loading of system prompt and base64 image
  - `historical_context_loader.go` - Loads historical verification data for PREVIOUS_VS_CURRENT
  
- **External Service Components**:
  - `bedrock_invoker.go` - Manages Bedrock API invocations with error handling
  - `storage_manager.go` - Handles S3 storage operations for responses
  - `dynamo_manager.go` - Manages DynamoDB operations with async support
  
- **Input/Output Components**:
  - `event_transformer.go` - Transforms Step Functions events to internal format
  - `prompt_generator.go` - Handles prompt generation with template processing
  - `validator.go` - Wraps validation logic for requests and responses
  
- **Utility Components**:
  - `helpers.go` - Utility functions (extractCheckingImageUrl, calculateHoursSince)

### Improved - Code Quality and Maintainability
- **Single Responsibility**: Each component has one clear, focused purpose
- **Better Testability**: Components can be tested in isolation with mock dependencies
- **Enhanced Readability**: Smaller files are easier to understand and navigate
- **Reduced Coupling**: Components interact through well-defined interfaces
- **Easier Collaboration**: Multiple developers can work on different components without conflicts

### Technical Details
- **File Size Reduction**: Main handler reduced from ~1000 lines to 193 lines (80% reduction)
- **Component Count**: 14 focused files replacing 1 monolithic file
- **Backward Compatibility**: All functionality preserved with identical behavior
- **Compilation**: All components compile successfully with no errors
- **Type Safety**: Maintained strong typing across all component interfaces

## [2.1.2] - 2025-05-23

### Fixed - Step Functions Input Format Compatibility
- **Input Format Mismatch**: Fixed validation error when receiving events from Step Functions
  - Function was expecting `Turn1Request` format but receiving Step Functions format with `schemaVersion`
  - Added detection logic to identify input format based on `schemaVersion` field presence
  - Validation errors were causing "INVALID_INPUT" failures with empty verification context

### Added - Step Functions Event Transformation
- **New Input Format Support**: Added support for Step Functions event format
  - Detects events with `schemaVersion` field and `s3References` map structure
  - Loads `processing_initialization` to extract `verificationContext`
  - Loads `images_metadata` to extract reference image S3 location
  - Loads `processing_layout-metadata` for LAYOUT_VS_CHECKING verification types
  - Transforms data into expected `Turn1Request` structure

- **transformStepFunctionEvent Method**: Added comprehensive transformation logic
  - Validates presence of required S3 references
  - Loads and parses initialization data for verification context
  - Loads and parses image metadata for S3 references
  - Handles optional layout metadata loading
  - Provides detailed logging throughout transformation process

### Changed - Handler Input Processing
- **HandleTurn1Combined Enhancement**: Updated to handle dual input formats
  - First attempts to parse as Step Functions format (with `schemaVersion`)
  - Falls back to original `Turn1Request` format for backward compatibility
  - Maintains existing functionality while supporting new input structure
  - Enhanced error messages to include available S3 references for debugging

### Technical Details
- **Backward Compatibility**: Existing Turn1Request format continues to work unchanged
- **S3 Loading Strategy**: Reuses `LoadSystemPrompt` method for JSON loading from S3
- **Error Context**: Enhanced error reporting with verification ID and available references
- **Logging Enhancement**: Added detailed transformation logging for operational visibility

### Input Format Examples
- **New Step Functions Format**:
  ```json
  {
    "schemaVersion": "2.1.0",
    "s3References": {
      "processing_initialization": {...},
      "images_metadata": {...},
      "prompts_system": {...}
    },
    "verificationId": "verif-xxx",
    "status": "PROMPT_PREPARED"
  }

## [2.1.1] - 2025-05-23

### Added - S3 Service Comprehensive Logging
- **Enhanced S3StateManager with Logger Support**: Added comprehensive logging throughout S3 operations
  - Modified `NewS3StateManager` to accept `logger.Logger` parameter
  - Added logger field to `s3Manager` struct for consistent logging across all methods
  - Updated main.go to pass logger during S3 service initialization

- **LoadSystemPrompt Logging**: Added detailed operational logging
  - Logs operation start with bucket, key, and size information
  - Logs validation failures with error context
  - Tracks retrieval duration in milliseconds
  - Logs successful completion with prompt length and preview (truncated to 100 chars)
  - Debug-level logging for JSON retrieval operations

- **LoadBase64Image Logging**: Added comprehensive image loading metrics
  - Logs operation start with expected size
  - Tracks and logs retrieval duration
  - Calculates and logs size ratio (actual/expected) for validation
  - Logs successful completion with actual data length
  - Error logging includes duration and expected size context

- **Store Operations Logging**: Enhanced visibility for storage operations
  - Added logging to `StoreRawResponse` and `StoreProcessedAnalysis`
  - Tracks operation duration from start to completion
  - Logs validation failures and storage errors with full context
  - Logs successful storage with S3 reference details (bucket, key, size)
  - Debug-level logging for category and content type

### Added - Utility Functions
- **truncateForLog Helper**: Added safe string truncation for logging
  - Prevents excessive log sizes by truncating long content
  - Preserves readability with "..." suffix for truncated strings
  - Used for system prompt preview in logs

### Technical Details
- **Performance Monitoring**: All S3 operations now report duration in milliseconds
- **Error Context**: Enhanced error logging with operation type, duration, and S3 metadata
- **Debug Support**: Added debug-level logs for detailed troubleshooting
- **Operational Visibility**: Improved monitoring capabilities for S3 operations

## [2.1.0] - 2025-05-23

### Changed - Configuration Error Handling
- **Removed Panics from Configuration Loading**: Replaced all panic calls with proper error handling
  - Modified `config.LoadConfiguration()` to return `(*Config, error)` instead of panicking
  - Changed `mustGet()` helper to return `(string, error)` and create `WorkflowError{Type:"Config"}`
  - Added structured error handling with `errors.NewConfigError()` for missing environment variables
  - Bootstrap now logs structured JSON errors for configuration failures before exiting
  - Step Functions can now properly route to `FinalizeWithError` branch on config errors

### Added - Error Infrastructure
- **Config Error Type**: Added new error type to shared errors package
  - Added `ErrorTypeConfig` constant for configuration-specific errors
  - Added `IsConfigError()` helper function for easy error type checking
  - Added `NewConfigError()` factory function with variable tracking
  - Config errors are marked as non-retryable with CRITICAL severity

### Added - Testing
- **Configuration Error Tests**: Added comprehensive unit tests for config error handling
  - Tests verify missing environment variables return proper `WorkflowError`
  - Tests verify error type, code, message, and variable details
  - Tests verify `IsConfigError()` helper works correctly
  - Tests verify default values are applied when optional vars are missing
  - All tests pass with 100% coverage of error paths

### Technical Details
- **Operational Improvements**: Cold-start failures now emit structured logs instead of stack traces
- **Error Visibility**: Missing environment variables are clearly identified in error messages
- **Lambda Integration**: Errors are properly propagated to Lambda runtime for Step Functions handling
- **Code Quality**: Removed all panic calls from configuration and bootstrap paths

## [2.0.0] - 2025-05-23

### Changed - BREAKING
- **Schema Constant Cleanup**: Removed duplicate schema constants to prevent drift
  - Deleted entire `const` block from `internal/models/shared_types.go` (lines 99-126)
  - All references now use `schema.*` constants directly from the shared package
  - Updated functions: `ConvertToSchemaStatus`, `ConvertFromSchemaStatus`, `CreateVerificationContext`, `IsEnhancedStatus`, `IsVerificationComplete`, `IsErrorStatus`
  - Single source of truth eliminates risk of constant drift between copies

- **Token Usage Truth Source**: Replaced estimated tokens with actual Bedrock usage
  - **BREAKING**: Removed `TokenEstimate` field from `TemplateProcessor` struct
  - Added `InputTokens` and `OutputTokens` fields to `TemplateProcessor`
  - Schema version bumped from "2.0.0" to "2.1.0" 
  - Token estimation now used only for pre-flight validation (not persisted)
  - Added `getMaxTokenBudget()` method with 16000 token limit
  - Handler now populates actual token counts from Bedrock response
  - Accurate cost tracking based on real usage, not estimates

## [1.3.4] - 2025-05-23

### Fixed
- **Correlation ID Collision Risk**: Implemented collision-resistant correlation ID generation
  - Previous implementation used only millisecond timestamp, risking collisions in high-throughput scenarios
  - New implementation combines multiple components for uniqueness:
    - Millisecond timestamp
    - 4-byte random component (8 hex characters) using crypto/rand
    - Atomic counter to ensure uniqueness within same millisecond
  - Format: `turn1-{timestamp}-{random}-{counter}` (e.g., `turn1-1737654321000-a1b2c3d4-1`)
  - Fallback mechanism if crypto/rand fails

## [1.3.3] - 2025-05-23

### Added
- **Historical Context Loading**: Implemented proper historical verification context loading for PREVIOUS_VS_CURRENT verification type
  - Added Stage 2.5 in handler to query previous verification data using `QueryPreviousVerification`
  - Extracts checking image URL from S3 reference to find matching historical verification
  - Populates HistoricalContext with previous verification data including:
    - PreviousVerificationAt, PreviousVerificationStatus, PreviousVerificationId
    - HoursSinceLastVerification calculation
    - VerificationSummary and layout metadata (RowCount, ColumnCount, RowLabels)
  - Added helper functions: `extractCheckingImageUrl` and `calculateHoursSince`

### Fixed
- **Template Context Flattening**: Updated prompt service to properly flatten HistoricalContext fields
  - Changed from nested `context["HistoricalContext"]` to flattened fields for direct template access
  - Ensures template variables like `{{.PreviousVerificationAt}}` work correctly
- **Template Path Configuration**: Fixed hardcoded template path in prompt service
  - Changed `NewPromptService` to accept full config object instead of just template version
  - Now uses `cfg.Prompts.TemplateBasePath` instead of hardcoded `/opt/templates`
  - Maintains flexibility across different environments (development, staging, production)
  - Updated main.go to pass config to `NewPromptService`
- **Historical Context Field Access**: Fixed compilation errors for missing VerificationContext fields
  - Removed references to non-existent `VerificationSummary` and `LayoutMetadata` fields
  - Updated to use available fields from `schema.VerificationContext` (LayoutId, LayoutPrefix)
  - Added default values for row/column information until proper metadata query is implemented
  - Code now compiles successfully while maintaining template compatibility

## [1.3.2] - 2025-05-23

### Fixed - Critical Production Issues
- **String Matching Bug**: Fixed broken `contains` function in prompt service that always returned true
  - Implemented proper case-insensitive string matching with `strings.Contains`
  - Added missing `strings` import to prompt.go
  - This bug was causing incorrect error classification in template processing
- **DynamoDB Service Initialization**: Replaced panic with proper error handling
  - Changed `NewDynamoDBService` to return `(DynamoDBService, error)` instead of panicking
  - Added proper error wrapping with contextual information
  - Updated main.go to handle initialization errors gracefully
- **Race Condition in Async Updates**: Fixed potential data loss in DynamoDB updates
  - Implemented channel-based completion tracking for asynchronous operations
  - Added 5-second timeout to ensure Lambda doesn't exit before critical updates
  - Prevents loss of verification status and conversation history updates

### Changed - Standardization Improvements
- **Template Management**: Integrated shared `templateloader` package for standardized template handling
  - Replaced undefined `TemplateManager` with `templateloader.TemplateLoader` interface
  - Configured standard template path `/opt/templates` for Lambda layers
  - Maintained backward compatibility with existing template processing metrics
- **Import Organization**: Fixed import aliasing for better clarity
  - Added `goerrors` alias for standard errors package to avoid conflicts
  - Properly distinguished between standard `errors` and shared `errors` packages

### Technical Details
- **Build Verification**: All compilation errors resolved, application builds successfully
- **Error Handling**: Enhanced error context with proper AWS SDK v2 error type checking
- **Code Quality**: Improved reliability and maintainability of critical paths

## [1.3.1] - 2025-05-22

### Fixed - Compilation Issues
- **Handler Cleanup**: Removed unused import `workflow-function/shared/schema` in handler.go
- **Error Handling Fixes**: Fixed `NewValidationError` calls to use proper `map[string]interface{}` parameter type
  - Updated request validation error handling (line 60)
  - Updated response validation error handling (line 289)
- **Prompt Service Fixes**: Resolved field access issues in prompt generation
  - Removed invalid `VerificationID` field references from `VerificationContext` type
  - Fixed string operator logic in `contains` helper function
  - Cleaned up error context building to use available fields only
- **Build Verification**: All Go compilation errors resolved, code now builds successfully

### Technical Details
- **Type Safety**: Ensured proper parameter types for shared error package functions
- **Field Validation**: Corrected field access patterns to match actual model structures
- **Code Quality**: Fixed logical errors in utility functions for better reliability

## [1.3.0] - 2025-01-22

### Added - Schema Package Integration & Standardization
- **Standardized Schema Integration**: Migrated to `workflow-function/shared/schema` v2.2.0
  - Integrated comprehensive type standardization across workflow functions
  - Added schema validation pipeline for request/response validation
  - Implemented enhanced `VerificationContext` with status tracking and metrics
  - Added support for combined function response handling with `CombinedTurnResponse`
- **Enhanced Validation Framework**: Comprehensive schema-based validation system
  - Added `SchemaValidator` for request/response validation pipeline
  - Implemented conversion functions between local and schema types
  - Added validation for S3 references, token usage, and workflow states
  - Enhanced error reporting with standardized `schema.ErrorInfo` structures
- **Template Management Support**: Ready for template-based prompt processing
  - Integrated `PromptTemplate` and `TemplateProcessor` support
  - Added template retrieval capabilities with S3 integration
  - Enhanced context processing with `TemplateContext` structures
- **Advanced Metrics & Tracking**: Comprehensive performance monitoring integration
  - Added `ProcessingMetrics` and `TurnMetrics` for detailed performance tracking
  - Implemented `StatusHistoryEntry` for complete status transition tracking
  - Enhanced error tracking with `ErrorTracking` and recovery attempt monitoring
  - Added granular stage tracking with `ProcessingStage` support
- **Enhanced Status Management**: Detailed status progression tracking
  - Integrated detailed Turn 1 status constants (`StatusTurn1Started`, `StatusTurn1ContextLoaded`, etc.)
  - Added error-specific status tracking (`StatusTurn1Error`, `StatusTemplateProcessingError`)
  - Enhanced status transition monitoring with comprehensive history tracking
- **Service Integration Examples**: Comprehensive schema usage patterns
  - Created `SchemaIntegratedService` demonstrating best practices
  - Added validation examples for Bedrock messages and image data
  - Implemented workflow state creation and management patterns
  - Enhanced error handling with schema standardization

### Enhanced - Architecture & Compatibility
- **Type Standardization**: Consistent types across all workflow functions
  - Enhanced `S3Reference`, `TokenUsage`, and `BedrockResponse` with schema integration
  - Added type conversion utilities for backward compatibility
  - Implemented shared constants from schema package (`VerificationTypeLayoutVsChecking`, etc.)
  - Enhanced model structures with schema-compatible field mappings
- **Validation Pipeline Integration**: Request/response validation in handler workflow
  - Added pre-processing request validation with detailed error reporting
  - Implemented post-processing response validation before return
  - Enhanced token usage validation with comprehensive checks
  - Added schema version tracking in all operations
- **Enhanced Error Handling**: Schema-standardized error management
  - Integrated `schema.ErrorInfo` for consistent error reporting
  - Added schema validation error handling with detailed field-level reporting
  - Enhanced error context with schema-based error creation utilities
  - Improved debugging with schema version tracking in error contexts
- **Backward Compatibility**: Maintained existing functionality while adding enhancements
  - Local type definitions remain functional with schema integration
  - Enhanced validation available as optional upgrade path
  - Consistent API surface with internal improvements
  - Zero breaking changes to existing interfaces

### Technical Improvements
- **Schema Version Management**: Comprehensive version tracking and compatibility
  - Added schema version validation and reporting in all operations
  - Enhanced logging with schema version context
  - Implemented version compatibility checks in validation pipeline
- **Performance Monitoring**: Built-in performance tracking with schema integration
  - Added processing time tracking with schema-standardized metrics
  - Enhanced token usage monitoring with validation
  - Implemented detailed stage-by-stage performance measurement
- **Code Organization**: Enhanced maintainability with schema standardization
  - Created `shared_types.go` for centralized type management
  - Added `schema_validator.go` for comprehensive validation framework
  - Enhanced service layer with schema integration examples
  - Improved code documentation with schema compatibility notes

### Files Added/Modified
- **New Files**:
  - `/internal/models/shared_types.go` - Schema type integration and conversion utilities
  - `/internal/validation/schema_validator.go` - Comprehensive validation framework
  - `/internal/services/schema_integration.go` - Schema usage examples and best practices
  - `SCHEMA_REVIEW.md` - Comprehensive schema compatibility documentation
- **Enhanced Files**:
  - `/go.mod` - Added schema package dependency and version management
  - `/internal/models/request.go` - Schema-compatible type definitions
  - `/internal/models/verification.go` - Enhanced with schema field compatibility
  - `/internal/models/bedrock.go` - Standardized token usage with schema types
  - `/internal/handler/handler.go` - Integrated validation pipeline and schema tracking

### Benefits Achieved
- **Cross-Function Consistency**: Standardized types ensure compatibility with ExecuteTurn2Combined
- **Enhanced Observability**: Comprehensive tracking and metrics with schema standardization
- **Improved Reliability**: Robust validation pipeline with detailed error reporting
- **Future-Proof Architecture**: Template management and advanced features ready for integration
- **Operational Excellence**: Enhanced debugging and monitoring with schema-based error handling

## [1.2.0] - 2025-05-22

### Added - Migration to Shared Packages Architecture
- **Shared Error Handling System**: Migrated from internal error handling to `workflow-function/shared/errors`
  - Implemented intelligent error classification with retry behavior analysis
  - Added contextual error enrichment with operational metadata
  - Enhanced Step Functions integration with structured error types
  - Introduced severity levels and API source tracking for better operational visibility
- **Shared Logger Integration**: Replaced internal logger with `workflow-function/shared/logger`
  - Implemented structured JSON logging with correlation ID support
  - Added fluent interface for logger context enrichment
  - Enhanced distributed tracing capabilities across Lambda functions
  - Improved operational debugging with consistent log format
- **Intelligent Bedrock Error Handling**: Advanced error classification for AI service failures
  - Implemented throttling detection with exponential backoff recommendations
  - Added content policy violation detection with non-retry classification
  - Enhanced token limit error handling with prompt size analysis
  - Improved model availability error detection with infrastructure alerting
- **Enhanced DynamoDB Error Management**: Sophisticated database operation error handling
  - Added context-rich error reporting with table and operation details
  - Implemented retry strategy recommendations based on failure type
  - Enhanced debugging information for attribute marshaling failures
- **Advanced Prompt Service Error Handling**: Deterministic failure detection and classification
  - Template syntax error detection with deployment guidance
  - Missing template file detection with configuration validation
  - Data structure mismatch analysis with debugging context
  - Input validation with proactive error prevention

### Changed - Architectural Improvements
- **Error Handling Philosophy**: Shifted from stage-based to cause-based error classification
  - Replaced `WrapRetryable`/`WrapNonRetryable` with intelligent `WrapError` analysis
  - Enhanced operational decision-making through error type classification
  - Improved system resilience through predictive error management
- **Logging Strategy**: Evolved from simple logging to operational intelligence
  - Enhanced context propagation across service boundaries
  - Improved correlation ID management for distributed tracing
  - Added structured event logging for input/output monitoring
- **Service Layer Architecture**: Unified error handling across all service implementations
  - Standardized error context enrichment patterns
  - Consistent retry behavior classification across services
  - Enhanced operational debugging capabilities
- **Dependency Injection Patterns**: Improved interface-based design
  - Clean separation between internal and shared package concerns
  - Enhanced testability through interface-based logger injection
  - Improved maintainability through consistent dependency patterns

### Technical Details - Implementation Patterns
- **Fluent Interface Implementation**: Leveraged method chaining for error context building
  - Eliminated unnecessary type assertions through proper interface design
  - Enhanced code readability through natural language-like error construction
  - Improved maintainability through consistent API patterns
- **Error Context Enrichment**: Comprehensive contextual information capture
  - Added verification ID propagation across error boundaries
  - Enhanced debugging information with operation-specific metadata
  - Improved incident response through detailed error reporting
- **Step Functions Integration**: Enhanced workflow error handling
  - Intelligent retry classification based on error analysis
  - Improved operational visibility through structured error propagation
  - Enhanced workflow resilience through contextual error information

### Enhanced Operational Excellence
- **Monitoring and Alerting**: Improved observability through structured error classification
  - Enhanced error severity tracking for operational prioritization
  - Improved incident response through detailed error context
  - Better trend analysis through consistent error categorization
- **Debugging and Troubleshooting**: Comprehensive diagnostic information capture
  - Enhanced error context with operation-specific details
  - Improved root cause analysis through structured error information
  - Better operational knowledge transfer through documented error patterns
- **System Resilience**: Intelligent failure handling and recovery strategies
  - Predictive error management with appropriate retry strategies
  - Enhanced fault tolerance through sophisticated error classification
  - Improved system stability through deterministic vs transient failure recognition

### Performance and Reliability Improvements
- **Cold Start Optimization**: Enhanced initialization error handling and monitoring
- **Memory Efficiency**: Improved error object lifecycle management
- **Network Resilience**: Better handling of transient network and service failures
- **Resource Management**: Enhanced error handling for AWS service integration

## [1.1.0] - 2025-05-25 (Updated)

### Planned Additions - Future Enhancements
- Enhanced metrics collection with shared error classification integration
- Support for additional Claude 3.7 features with intelligent error handling
- Improved error recovery mechanisms with predictive failure analysis
- Performance optimizations for large payloads with enhanced error monitoring
- Additional test coverage for shared package integration patterns
- Enhanced operational dashboards with error classification insights

## [1.0.1] - 2025-05-21

### Fixed
- Fixed import issues with shared packages
- Corrected type conflicts in dynamodb.go
- Updated s3.go to use the correct Reference type from s3state
- Added missing module dependencies in go.mod
- Added ExecuteTurn1Combined to go.work workspace

### Changed
- Improved error handling in bedrock.go
- Enhanced logging for better observability
- Updated documentation with accurate configuration options

## [1.0.0] - 2025-05-20

### Added
- Initial release of the ExecuteTurn1Combined Lambda function
- Combined functionality from PrepareSystemPrompt, PrepareTurn1Prompt, and ExecuteTurn1
- Modular service architecture with clean interfaces
- Integration with shared packages (bedrock, s3state, logger, schema)
- Comprehensive error handling and logging
- Support for DynamoDB state tracking
- S3-based state management for large payloads

## [0.9.0] - 2025-05-15

### Added
- Beta implementation with core functionality
- Initial service interfaces and implementations
- Basic error handling and logging
- Configuration from environment variables

### Known Issues
- Missing proper error handling for some edge cases
- Incomplete documentation
- Limited test coverage

## [0.8.0] - 2025-05-10

### Added
- Alpha implementation with basic structure
- Proof of concept for combined workflow
- Initial integration with Bedrock API
- Basic S3 and DynamoDB integration

---

## Migration Guide - Shared Packages Integration

### Breaking Changes
- **Error Handling**: Migrated from internal error stages to shared error types
  - `errors.StageBedrockCall` → `errors.ErrorTypeBedrock`
  - `errors.StageDynamoDB` → `errors.ErrorTypeDynamoDB`
  - `errors.StageStorage` → `errors.ErrorTypeS3`
  - `errors.WrapRetryable` → `errors.WrapError(err, type, msg, true)`
  - `errors.WrapNonRetryable` → `errors.WrapError(err, type, msg, false)`

- **Logger Interface**: Replaced internal logger with shared logger interface
  - Constructor: `logger.New(service, function)` instead of `logger.New(component)`
  - Methods: `logger.Info(msg, details)` instead of `logger.Info(msg, kv...)`
  - Context: `logger.WithCorrelationId(id).WithFields(map)` for enrichment

### Migration Benefits
- **Enhanced Operational Visibility**: Rich error context with severity and retry guidance
- **Improved System Resilience**: Intelligent error classification and handling strategies
- **Better Debugging Experience**: Comprehensive diagnostic information in all errors
- **Consistent Logging**: Structured JSON logging with correlation ID support across all functions
- **Step Functions Integration**: Enhanced workflow error handling with intelligent retry logic

### Configuration Updates
No configuration changes required. All improvements are backward compatible with existing environment variables and configuration patterns.

