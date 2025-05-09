# Changelog

All notable changes to the InitializeFunction will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2025-05-09

### Fixed
- Fixed "failed to parse event: failed to parse event detail: unexpected end of JSON input" error when invoked from Step Functions
- Enhanced input parsing to properly extract top-level requestId and requestTimestamp fields when verificationContext is present
- Improved error logging with more detailed JSON content for debugging

## [1.1.0] - 2025-05-08

### Changed
- Made `previousVerificationId` and `vendingMachineId` optional for `PREVIOUS_VS_CURRENT` verification type
- Updated validation logic in `service.go` to remove requirement for `previousVerificationId` in `PREVIOUS_VS_CURRENT` type
- Ensured alignment between API Gateway model, Step Function state machine, and function validation

### Fixed
- Fixed potential validation error when `PREVIOUS_VS_CURRENT` verification requests don't include `previousVerificationId`

## [1.0.0] - 2025-04-20

### Added
- Initial implementation of InitializeFunction
- Support for two verification types: `LAYOUT_VS_CHECKING` and `PREVIOUS_VS_CURRENT`
- Validation for required fields based on verification type
- Resource validation for images and layouts
- Historical context retrieval for `PREVIOUS_VS_CURRENT` verification type
- DynamoDB integration for storing verification records
- S3 integration for validating image existence
