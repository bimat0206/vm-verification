# Changelog

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
