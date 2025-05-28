# Changelog

All notable changes to the ExecuteTurn2Combined function will be documented in this file.

## [Unreleased] – 2025-05-27
### Added
- `DynamoManager.UpdateTurn2Completion`
- `S3Manager.StoreTurn2RawResponse` and `StoreTurn2ProcessedResponse`
### Changed
- Removed stubbed `handler.go`
- Wired `PromptServiceTurn2` into `turn2_handler.go`
### Fixed
- Ensured DynamoDB statuses now include `TURN2_COMPLETED`
- Resolved compilation error by renaming helper method receivers to `Turn2Handler`

## [1.3.4] - 2025-05-28
### Changed
- Modified `internal/bedrock/adapter_turn2.go` in `ConverseWithHistory` to use the actual checking image format rather than a hardcoded `"jpeg"`.
- Updated method signatures throughout the call chain (`client_turn2.go`, `services/bedrock_turn2.go`, `handler/turn2_handler.go`) to pass the checking image format from `ContextLoader`.
- `ContextLoader` now loads image metadata to determine the checking image format.

## [1.3.1] - 2025-06-06
### Fixed
- Resolved compilation errors due to outdated struct fields and renamed parser functions.
- Updated `S3StateManager` interface to include Turn 2 storage helpers.
- Adjusted `Turn2Handler` to use `VerificationContext` fields and updated processing metrics structure.

## [1.3.2] - 2025-06-07
### Fixed
- Added `LoadTurn1ProcessedResponse` and `LoadTurn1RawResponse` to `S3StateManager` interface.
- Corrected unused variables and token usage fields in `Turn2Handler`.
- Defined `ErrorTypeTemplate` for template rendering errors.

## [1.3.3] - 2025-06-08
### Fixed
- Resolved missing field errors in `handler_helpers.go` by expanding `Turn2Handler` with
  validator, tracking utilities, and service references.
- Added stub implementations for `Handle` and `HandleForStepFunction` to satisfy
  compiler requirements.

## [1.3.0] - 2025-06-05
### Added
- **Business Logic for Discrepancy Interpretation**:
  - Implemented a new method in `Turn2Handler` to apply business rules to the discrepancies and verification outcome parsed from Bedrock's Turn 2 response.
  - This allows for refinement of the AI's output, such as adjusting the final `verificationStatus` based on the severity or quantity of discrepancies, overriding Bedrock's initial assessment if specific critical conditions are met.
  - Introduced configurable thresholds (e.g., `DiscrepancyThreshold` in `config.Config`) for rule-based decision making.
- **`PromptServiceTurn2` Implementation**:
  - Provided a concrete implementation for the `PromptServiceTurn2` interface in `internal/services/prompt_turn2.go`.
  - This service now dynamically generates Turn 2 comparison prompts by selecting templates based on `VerificationContext.VerificationType` and building rich context for template rendering.
  - Returns the rendered prompt and a `*schema.TemplateProcessor` object with processing metrics.
### Changed
- Modified `turn2_handler.go` to utilize the new discrepancy interpretation logic and updated configuration values.
- Added new configuration fields `DiscrepancyThreshold` and `Turn2TemplateVersion`.
### Purpose
- These changes improve the robustness, accuracy, and configurability of the ExecuteTurn2Combined function, enabling more reliable verification results and prompt generation.

## [1.2.0] - 2025-05-28
### Added
- **DynamoDB Integration for Turn 2**:
  - Implemented updates to the `VerificationResults` table upon successful completion or critical failure of Turn 2 processing. This includes persisting `currentStatus`, `processingMetrics.turn2`, final `verificationStatus`, `discrepancies`, `verificationSummary`, and error tracking information.
  - Implemented updates to the `ConversationHistory` table to append Turn 2 interaction details and finalize `currentTurn` and `turnStatus`.
- **Enhanced Error Handling**:
  - Added granular error logging and persistence using `shared/errors.WorkflowError` and structured logger.
  - Critical failures now update DynamoDB error tracking before the error is returned.

### Fixed
- Addressed gap where Turn 2 processing outcomes were not persisted to DynamoDB.
- Improved robustness by ensuring errors are categorized and logged consistently.

### Purpose
- These changes ensure comprehensive data persistence for Turn 2 and better observability for debugging and workflow management.

## [1.1.0] - 2025-05-27
### Added
- **Complete architectural transformation from ExecuteTurn1Combined to ExecuteTurn2Combined**
- **Core Turn2 Processing Architecture:**
  - Turn2Request/Turn2Response data models with discrepancy tracking and verification outcomes
  - Turn2-specific handler with proper JSON event parsing and response formatting
  - Concurrent context loading for checking images, Turn1 results, and system prompts
  - Turn2 conversation flow with Turn1 history integration for Bedrock invocation

### Enhanced
- **Bedrock Integration:**
  - Turn2-specific adapter (`AdapterTurn2`) with conversation history from Turn1
  - Proper conversation structure: Turn1 analysis → Turn2 comparison prompt
  - Enhanced client initialization with correct parameter handling and configuration mapping
  - Turn2 processing with timeout handling and comprehensive error management

- **Service Layer:**
  - Updated all import paths from ExecuteTurn1Combined to ExecuteTurn2Combined
  - Fixed Bedrock service configuration to use ExecuteTurn2Combined config structure
  - Enhanced storage manager with Turn2-specific artifact handling
  - Proper S3 state management integration for Turn2 workflows

- **Schema Support:**
  - Added `Turn1ProcessedResponse` type for Turn1 result integration
  - Enhanced `BedrockResponse` with Turn2-specific metadata
  - Added `ModelConfig` for Bedrock configuration tracking
  - Turn2 template processor integration for comparison prompts

### Fixed
- **Import Path Corrections:**
  - Resolved all ExecuteTurn1Combined → ExecuteTurn2Combined import references
  - Fixed internal package access violations across service layers
  - Corrected shared package imports for bedrock, errors, logger, and schema

- **Configuration Handling:**
  - Fixed config structure references (cfg.AWS.BedrockModel vs cfg.Bedrock.ModelId)
  - Proper timeout configuration using Processing.BedrockCallTimeoutSec
  - Correct parameter mapping for Bedrock client initialization

- **Client Integration:**
  - Fixed NewClient parameter order and types
  - Resolved assignment mismatches in client initialization
  - Proper shared Bedrock client integration with local adapter pattern

### Technical Implementation
- **Turn2 Conversation Flow:**
  ```go
  Messages: []MessageWrapper{
      {Role: "user", Content: "[Turn 1] Please analyze this image."},
      {Role: "assistant", Content: turn1Message}, // From Turn1 results
      {Role: "user", Content: "[Turn 2] " + turn2Prompt}, // Comparison prompt
  }
  ```

- **Context Loading Architecture:**
  - Concurrent loading of 4 resources: system prompt, checking image, Turn1 processed response, Turn1 raw response
  - Enhanced error handling with proper context enrichment
  - Turn2-specific S3 reference handling

- **Processing Pipeline:**
  - Turn2Request → LoadContext → GeneratePrompt → InvokeBedrock → ParseResponse → Turn2Response
  - Proper verification outcome determination (CORRECT/INCORRECT)
  - Discrepancy tracking and comparison summary generation

### Infrastructure
- **Lambda Function Setup:**
  - Updated main entry point with Turn2-specific routing
  - Enhanced logging with "turn2_comparison" architecture pattern
  - Proper correlation ID generation with "turn2-" prefix
  - Turn2-specific error handling and status tracking

- **Template Integration:**
  - Turn2 comparison templates: `turn2-layout-vs-checking`, `turn2-previous-vs-current`
  - Template processor integration for Turn2 prompt generation
  - Enhanced template metadata tracking

### Dependencies
- **Shared Package Integration:**
  - Enhanced shared/bedrock client with Turn2 conversation support
  - Proper shared/errors integration with Turn2-specific error types
  - Enhanced shared/logger with Turn2 context tracking
  - Updated shared/schema with Turn2 data structures

## [1.0.0] - 2025-05-26
### Added
- Initial development of ExecuteTurn2Combined Lambda function.
- Implements Turn 2 processing for vending machine verification:
  - Consumes output from ExecuteTurn1Combined (or equivalent state).
  - Loads checking image and Turn 1 analysis.
  - Generates Turn 2 comparison prompts using shared/templateloader.
  - Invokes Amazon Bedrock (Claude 3.7 Sonnet) via shared/bedrock client, maintaining conversation history.
  - Parses Bedrock response to identify discrepancies.
  - Stores Turn 2 artifacts (raw response, processed analysis) to S3 using shared/s3state and date-partitioned paths.
  - Updates VerificationResults and ConversationHistory DynamoDB tables.
  - Updates the input initialization.json S3 object with its completion status.
  - Leverages shared/logger and shared/errors for observability and error handling.

## [0.1.0] - 2025-06-04
### Added
- Initial skeleton implementation.
