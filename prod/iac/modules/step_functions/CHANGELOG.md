# Changelog

## [2.2.0] - 2025-06-02

### Removed
- **BREAKING CHANGE**: Completely removed notification functionality from Step Functions workflow
  - **Notify State**: Removed `Notify` task state that invoked the notify lambda function
  - **ShouldNotify Choice State**: Removed `ShouldNotify` choice state that conditionally triggered notifications
  - **Notification Logic**: Eliminated all notification-related conditional logic and parameters

### Changed
- **Simplified Workflow**: Streamlined state machine by removing notification complexity
  - `FinalizeAndStoreResults` now transitions directly to `WorkflowComplete`
  - Removed conditional notification flow based on `verificationContext.notificationEnabled`
  - Eliminated notification result handling and error paths

- **Updated State Machine Definition**: Modified template to remove notification states
  - Removed `ShouldNotify` choice state with notification condition checks
  - Removed `Notify` task state with SNS integration and retry logic
  - Simplified workflow completion path for better performance

### Benefits
- **Reduced Complexity**: Further simplified state machine by removing notification overhead
- **Improved Performance**: Faster workflow completion without notification processing delays
- **Enhanced Reliability**: Fewer potential failure points in the verification workflow
- **Cleaner Architecture**: More focused workflow without notification dependencies

### Technical Details
- **State Count Reduction**: Reduced from 12 to 10 states (17% additional reduction)
- **Simplified Error Handling**: Removed notification-specific error handling paths
- **Direct Completion**: Verification workflow now completes immediately after result storage
- **Maintained Functionality**: All core verification capabilities preserved

### Migration Impact
- **Automatic Update**: State machine definition will be automatically updated during deployment
- **No Data Loss**: Historical verification data and workflow state preserved
- **Backward Compatibility**: Existing verification executions will complete normally
- **Performance Improvement**: New verifications will complete faster without notification overhead

## [2.1.0] - 2025-05-31

### Changed
- **BREAKING CHANGE**: Simplified state machine by removing all Handle...Error states
- Consolidated error handling to use single `FinalizeWithError` state for all errors
- Combined `FinalizeResults` and `StoreResults` into unified `FinalizeAndStoreResults` state
- Updated all Catch blocks to point directly to `FinalizeWithError` instead of specific error handlers
- Removed 7 error handling states: HandleInitializationError, HandleHistoricalFetchError, HandleFetchImagesError, HandlePromptError, HandleBedrockError, HandleFinalizationError, HandleStorageError
- Simplified `FinalizeWithError` state to handle all error types generically without specific error stages

### Benefits
- Reduced state machine complexity from 19 to 12 states (37% reduction)
- Streamlined error handling with single point of failure management
- Improved maintainability with fewer state transitions
- Consolidated finalization and storage operations for better performance
- Maintained all retry logic and error resilience

## [2.0.1] - 2025-05-22

### Fixed
- Removed unreachable `HandleProcessingError` state from Step Functions state machine definition
- Fixed "State 'HandleProcessingError' is not reachable" validation error

## [2.0.0] - 2025-05-22

### Changed
- **BREAKING CHANGE**: Updated state machine workflow to use consolidated Lambda functions
- Simplified state flow: `FetchImages → PrepareSystemPrompt → ExecuteTurn1Combined → ExecuteTurn2Combined → FinalizeResults`
- Removed 6 intermediate states: `PrepareTurn1Prompt`, `ExecuteTurn1`, `ProcessTurn1Response`, `PrepareTurn2Prompt`, `ExecuteTurn2`, `ProcessTurn2Response`
- Added 2 new states: `ExecuteTurn1Combined` and `ExecuteTurn2Combined`
- Maintained `PrepareSystemPrompt` state as shared system prompt generator for both turns
- Enhanced retry logic for combined Bedrock operations (3-5 retries with exponential backoff)
- Streamlined error handling with atomic turn operations

### Benefits
- Reduced state machine complexity with fewer transitions
- Improved workflow reliability with atomic operations
- Enhanced performance with optimized state transitions
- Maintained error resilience with proper retry mechanisms
- Preserved backward compatibility for existing data structures

## [1.2.9] - 2025-05-14

### Fixed
- Removed `historicalContext.$: "$.historicalContext"` parameter from FetchImages state to fix JSONPath error
- This prevents "The JSONPath '$.historicalContext' could not be found in the input" error when processing LAYOUT_VS_CHECKING verification types
- The FetchImages Lambda function already handles the absence of historicalContext by creating an empty object

## [1.2.8] - 2025-05-14

### Changed
- Renamed "GenerateMissingFields" state to "InitializeVerificationContext" to better reflect its purpose
- State name updated to accurately describe its function in initializing the full verification context structure

## [1.2.7] - 2025-05-14

### Fixed
- Fixed "SCHEMA_VALIDATION_FAILED: The value for the field 'previousVerificationId.$' must be a valid JSONPath or a valid intrinsic function call" error
- Replaced complex intrinsic function with a simple default value and a conditional state
- Added CheckPreviousVerificationId and UpdatePreviousVerificationId states to handle PREVIOUS_VS_CURRENT verificationType
- Implemented robust handling of previousVerificationId for different verification types

## [1.2.6] - 2025-05-14

### Fixed
- Fixed "JSONPath '$.verificationContext.previousVerificationId' could not be found" error in GenerateMissingFields state
- Used States.JsonToString and States.Array intrinsic functions to handle missing previousVerificationId field
- Improved the state machine to gracefully handle LAYOUT_VS_CHECKING verification type which doesn't have previousVerificationId
- Eliminated complex multi-state approach in favor of a simpler solution using intrinsic functions

## [1.2.5] - 2025-05-14

### Fixed
- Fixed "JSONPath '$.verificationContext.turnConfig' could not be found" error by adding default turnConfig, turnTimestamps, and requestMetadata structures in the ExecuteTurn1 state
- Eliminated dependency on fields that might not be present in the input
- Added default values for maxTurns, referenceImageTurn, and checkingImageTurn
- Added current timestamp as initialization time in the turnTimestamps structure

## [1.2.4] - 2025-05-14

### Fixed
- Fixed status field mismatch in ExecuteTurn1 state causing "ValidationException: Invalid value for field status: got IMAGES_FETCHED, expected TURN1_PROMPT_READY"
- Modified the ExecuteTurn1 state to explicitly set status to "TURN1_PROMPT_READY" in verificationContext
- Enhanced the Step Function definition to properly pass the status between PrepareTurn1Prompt and ExecuteTurn1
- Improved the structure of verificationContext in ExecuteTurn1 state to meet validation requirements

## [1.2.3] - 2025-05-13

### Fixed
- Fixed ExecuteTurn1 Lambda function error (Runtime.ExitError) by modifying the ExecuteTurn1 state parameters
- Added proper nested structure for currentPrompt and systemPrompt inputs
- Fixed thinking.type value from "enable" to "enabled" in BedrockConfig
- Explicitly structured BedrockConfig parameters to ensure consistent input format
- Modified the Parameters section to create the structure expected by ExecuteTurn1 Lambda

## [1.2.2] - 2025-05-11

### Fixed
- Fixed "JSONPath '$.historicalContext' could not be found" error in the PrepareSystemPrompt state
- Modified all states that reference historicalContext to handle missing field for LAYOUT_VS_CHECKING verification type
- Used an empty object literal for historicalContext instead of intrinsic functions to fix SCHEMA_VALIDATION_FAILED errors
- Simplified approach with static empty JSON object for historicalContext when the field is missing

## [1.2.1] - 2025-05-11

### Fixed
- Fixed issue with `previousVerificationId` field in Step Functions FetchImages state
- Modified Step Functions template to handle missing previousVerificationId field for LAYOUT_VS_CHECKING verification types
- Previously the workflow was trying to access `$.verificationContext.previousVerificationId` which doesn't exist for LAYOUT_VS_CHECKING verification types

## [1.1.2] - 2025-05-11

### Fixed
- Fixed "SCHEMA_VALIDATION_FAILED" error in InitializePreviousCurrent state by properly using JSONPath reference for previousVerificationId parameter instead of hardcoded empty string
- Fixed "SCHEMA_VALIDATION_FAILED" error in FetchImages state by simplifying complex intrinsic function for previousVerificationId

## [1.1.1] - 2025-05-11

### Fixed
- Fixed "JSONPath '$.verificationContext.previousVerificationId' could not be found" error in FetchImages state by implementing conditional parameter inclusion using Step Functions intrinsic functions
- Fixed "SCHEMA_VALIDATION_FAILED" error by simplifying the intrinsic function syntax for conditional previousVerificationId handling
- Used a simpler combination of States.ArrayGetItem and States.Array intrinsic functions to conditionally include previousVerificationId only when verificationType is PREVIOUS_VS_CURRENT

## [1.1.0] - 2025-05-11

### Added
- Added comprehensive documentation for FetchHistoricalVerification Lambda integration
- Added detailed error handling for FetchHistoricalVerification state

### Changed
- Improved documentation for state machine parameter handling and data flow
- Renamed `output.tf` to `outputs.tf` to match the expected naming convention in the main.tf file
- Updated module structure documentation in README.md to reflect the file name change

## [1.0.1] - 2025-05-10

### Changed
- Updated state machine definition template to pass `previousVerificationId` parameter to InitializePreviousCurrent state
- Made `previousVerificationId` and `vendingMachineId` optional for PREVIOUS_VS_CURRENT verification type
- Updated CheckVerificationType state to use `$.verificationContext.verificationType` path for Choice state
- Updated InitializeLayoutChecking and InitializePreviousCurrent states to use verificationContext-prefixed parameters
- Fixed InitializePreviousCurrent state to handle missing previousVerificationId by providing an empty string default
- Added GenerateMissingFields state to generate requestId and requestTimestamp if not present in the input
- Removed `previousVerificationId` parameter from FetchHistoricalVerification state to fix JSONPath error
- Fixed conditional handling of previousVerificationId in FetchImages state by replacing the ternary operator with States.ArrayGetItem and States.Array intrinsic functions

### Fixed
- Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error in InitializeLayoutChecking step by properly nesting the verificationContext object
- Fixed "verificationId is required" validation error in FetchImages step by properly extracting fields from verificationContext
- Improved parameter passing between Step Function states to ensure consistent input structure for Lambda functions
- Fixed "SCHEMA_VALIDATION_FAILED" error in FetchImages state by correcting the intrinsic function syntax for conditional previousVerificationId handling

## [1.0.0] - 2025-05-08

### Added
- Initial release of the Step Functions module
- Support for creating AWS Step Functions state machines
- IAM role configuration for state machine execution
- State machine definition templates
- Integration with Lambda functions
- CloudWatch logging configuration
- X-Ray tracing support
- Tagging support
