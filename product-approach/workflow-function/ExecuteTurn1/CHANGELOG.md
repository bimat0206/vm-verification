# Changelog

## [1.0.5] - 2025-05-14

### Fixed
- Fixed validation error "ValidationException: Invalid value for field status: got IMAGES_FETCHED, expected TURN1_PROMPT_READY" when invoked from Step Function
- Improved documentation on expected status values in validation.go
- Added more context to the error message to aid in troubleshooting

## [1.0.4] - 2025-05-13

### Fixed
- Fixed missing main.go file (renamed from main.md)
- Ensured proper main file extension for Go compilation

## [1.0.3] - 2025-05-13

### Fixed
- Fixed Docker build issue by updating the go build command in Dockerfile
- Changed build command to use relative path `./cmd/main.go` instead of `cmd/main.go`

## [1.0.2] - 2025-05-13

### Fixed
- Fixed build errors related to CurrentPromptWrapper structure handling
- Added proper extraction and validation functions for CurrentPrompt
- Modified BedrockClient to use the extraction functions
- Added BucketOwner field to Images struct
- Created helper ExtractBucketOwner function
- Fixed Response processing to handle nested structures
- Updated validation logic to support the wrapper structure
- Improved compatibility with Step Function input format

## [1.0.1] - 2025-05-13

### Fixed
- Fixed input structure handling for ExecuteTurn1 function
- Extended extractCurrentPrompt and extractBedrockConfig functions to better handle nested input structures
- Improved compatibility with the Step Function workflow by accommodating the nested currentPrompt and systemPrompt objects
- Added validation to accept both "enable" and "enabled" values for thinking.type to ensure backward compatibility

## [1.0.0] - 2025-05-10

### Added
- Initial release of ExecuteTurn1 Lambda function
- Integration with Amazon Bedrock to execute Turn 1 of the verification process
- Support for sending reference image to Claude 3.7 model
- Implementation of retry logic for transient failures
- Comprehensive input validation
- Token usage tracking and response processing
- Conversation state management