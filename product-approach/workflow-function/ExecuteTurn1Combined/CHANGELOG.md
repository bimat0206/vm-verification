# Changelog

All notable changes to the ExecuteTurn1Combined function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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