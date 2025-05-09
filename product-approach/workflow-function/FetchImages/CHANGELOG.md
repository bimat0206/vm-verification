
---

## CHANGELOG.md

```markdown
# Changelog

## [0.2.2] - 2025-05-09

### Fixed
- Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error in InitializeLayoutChecking step
- Updated Step Function state machine definition to properly structure the verificationContext object for the Initialize Lambda

## [0.2.1] - 2025-05-09

### Fixed
- Fixed "verificationId is required" validation error when invoked from Step Functions
- Updated Step Function state machine definition to properly extract fields from verificationContext
- Ensured proper parameter passing between Step Function states

## [0.2.0] - 2025-05-09

### Added
- Enhanced input handling to support multiple invocation types:
  - Direct Step Function invocations
  - Function URL requests
  - Direct struct invocations
  - Fallback for other formats
- Improved error logging with detailed input capture for debugging

### Fixed
- Fixed "Invalid JSON input: unexpected end of JSON input" error when invoked from Step Functions
- Resolved input parsing issues between different invocation methods

## [0.1.0] - 2024-06-01

### Added
- Initial implementation of FetchImages Lambda:
  - Input validation
  - S3 metadata fetch (no image bytes or base64)
  - DynamoDB layout and historical context fetch
  - Parallel/concurrent fetch logic
  - Config via environment variables
  - Structured logging

### Changed
- N/A

### Removed
- Any base64 image handling (S3 URI only)
