# Changelog

All notable changes to this project are documented here.

---

## [1.3.0] - 2025-05-18

### Changed
- **API Migration:**
  - Migrated from Bedrock Converse API to InvokeModel API for improved compatibility with AWS SDK
  - Restructured request/response handling to work with the newer API contract
  - Fixed type safety issues with error handling and schema compatibility
- **Architecture Support:**
  - Migrated to ARM64/Graviton Lambda for improved cost and performance
  - Updated Go version from 1.21 to 1.22
  - Added platform-specific build flags

### Added
- Enhanced build process with optimized Go compiler flags
- Better error diagnostics in the Docker build script
- Platform detection in the build script for cross-compilation
- Updated documentation to reflect API changes

---

## [1.2.0] - 2025-05-17

### Changed
- **Major migration:**  
  - All workflow and Lambda modules now fully use shared `schema`, `logger`, and `errors` packages.
  - Removed all custom/request-specific type definitions for workflow state, prompt, images, and error contracts.
- **Centralized AWS client initialization and config validation.**
- **Bedrock integration:**  
  - Now uses the Claude 3.7 Sonnet Converse API exclusively.
  - Bedrock messages are always constructed via the shared schema, supporting inline and S3-based Base64 image references.
- **Hybrid Base64 support:**  
  - Large images automatically use S3 temporary storage; retrieval and embedding for Bedrock handled via schema helpers.
- **Logging overhaul:**  
  - Replaced all raw log output with JSON-structured logs via the shared logger.
  - Correlation IDs and context fields now populate all logs for distributed traceability.
- **Error handling overhaul:**  
  - All errors are now `WorkflowError` types (typed, retryable, and with full context).
  - All errors surfaced in both Lambda error and within `VerificationContext.Error` for Step Functions compatibility.

### Added
- Helper utilities in `request.go` and `response_processor.go` for request/response validation and workflow state management.
- Standardized error creation and propagation in all modules.
- Docs/README updated for new architecture.

---

## [1.1.0] - 2025-05-15

### Added
- Initial implementation of shared schema, logger, and error packages.
- Core support for hybrid Base64 image storage (inline/S3).
- Validation and builder helpers for all major workflow state fields.

---

## [1.0.0] - 2025-05-14

### Added
- First release of vending machine verification solution.
- Lambda function skeletons and event-driven Step Functions integration.
