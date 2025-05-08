# Changelog

All notable changes to the Step Functions module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Renamed `output.tf` to `outputs.tf` to match the expected naming convention in the main.tf file
- Updated module structure documentation in README.md to reflect the file name change
- Updated state machine definition template to pass `previousVerificationId` parameter to InitializePreviousCurrent state
- Made `previousVerificationId` and `vendingMachineId` optional for PREVIOUS_VS_CURRENT verification type
- Updated CheckVerificationType state to use `$.verificationContext.verificationType` path for Choice state
- Updated InitializeLayoutChecking and InitializePreviousCurrent states to use verificationContext-prefixed parameters

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of the Step Functions module
- Support for creating AWS Step Functions state machines
- IAM role configuration for state machine execution
- State machine definition templates
- Integration with Lambda functions
- CloudWatch logging configuration
- X-Ray tracing support
- Tagging support
