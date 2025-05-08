# Changelog

All notable changes to the API Gateway module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Renamed `output.tf` to `outputs.tf` to match the expected naming convention in the main.tf file
- Updated module structure documentation in README.md to reflect the file name change
- Updated VerificationRequest model to make previousVerificationId and vendingMachineId optional for PREVIOUS_VS_CURRENT verification type
- Updated API documentation to clarify that requests must use the verificationContext wrapper structure

## [1.1.0] - 2024-05-XX

### Changed
- **BREAKING**: Modified API base path from `/api/v1` to `/api/` to simplify the final invoke URL structure
- Updated all resource paths in resources.tf to reflect the new base path
- Updated all endpoint documentation in README.md
- Added new section in README.md explaining the API base path structure

### Fixed
- Eliminated redundant path segment in API URLs, which previously resulted in URLs like `https://abc123.execute-api.ap-southeast-1.amazonaws.com/v1/api/v1/...`
- Now URLs follow a cleaner pattern: `https://abc123.execute-api.ap-southeast-1.amazonaws.com/v1/api/...`

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of the API Gateway module
- Support for multiple endpoints with Lambda integrations
- CORS configuration options
- API key support
- Throttling configuration
- CloudWatch logging integration
- Structured module organization for improved maintainability
