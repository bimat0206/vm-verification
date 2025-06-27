# Changelog

All notable changes to the API Gateway module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-06-02

### Removed
- **BREAKING CHANGE**: Removed notification functionality from API Gateway integration
  - **Lambda Permission**: Removed lambda permission for notify function invocation
  - **Model Property**: Removed `notificationEnabled` property from `VerificationRequest` model schema

### Changed
- **Simplified API Model**: Updated verification request model to remove notification support
  - Removed `notificationEnabled` boolean property from verification context schema
  - Updated required fields list to exclude notification preferences
  - Simplified API payload structure for verification initiation

- **Updated Integration Responses**: Modified integration responses for verification endpoints
  - Updated response parameter handling for simplified verification workflow
  - Removed notification-related response headers and templates
  - Streamlined response structure for better performance

### Benefits
- **Simplified API**: Cleaner API interface without notification complexity
- **Reduced Payload Size**: Smaller request payloads without notification preferences
- **Enhanced Performance**: Faster API processing without notification parameter validation
- **Cleaner Documentation**: Simplified API documentation without notification options

### Migration Impact
- **Automatic Update**: API Gateway model will be automatically updated during deployment
- **Client Compatibility**: Existing clients should remove `notificationEnabled` from requests
- **Backward Compatibility**: API will ignore any notification-related fields if present
- **No Functional Impact**: Core verification functionality remains unchanged

## [Unreleased]

### Fixed
- **Binary Media Type Support**: Added binary media types configuration to API Gateway REST API to properly handle image uploads
  - Added `multipart/form-data`, `image/*`, and `application/octet-stream` to `binary_media_types`
  - Added `content_handling = "CONVERT_TO_BINARY"` to image upload integration
  - Added missing integration response for image upload POST method
  - Updated deployment triggers to include image upload methods
  - This fixes image corruption issues during upload by ensuring binary data is properly base64 encoded

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
