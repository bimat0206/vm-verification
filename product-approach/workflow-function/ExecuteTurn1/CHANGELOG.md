# Changelog

All notable changes to this project are documented here.

---

## \[1.3.2] - 2025-05-18

### Fixed

* Fixed s3.Client type error in NewHandler function by using the correct parameter name
* Added missing errors import and renamed to wferrors to avoid name collision
* Fixed invalid use of respOut.Body as io.Reader by using the bytes directly
* Fixed access to undefined field Identifier in ImageInfo by using URL instead
* Fixed invalid Body field in InvokeModelInput struct literal
* Fixed the mergeTokenUsage function to handle absence of token fields in BedrockApiResponse
* Added missing aws import in dependencies/clients.go
* Fixed code to compile cleanly with Go 1.22

## \[1.3.1] - 2025-05-18

### Changed

* Config: removed hardcoded defaults; require environment variables `AWS_REGION`, `BEDROCK_MODEL`, `ANTHROPIC_VERSION`, and validate at startup.
* ExecuteTurn1: generate Base64 images before full validation to prevent INVALID\_IMAGE\_DATA errors.
* ExecuteTurn1: use `BedrockModelID` from environment for `ModelId` in Bedrock requests.
* ExecuteTurn1: fixed image payload schema to `{ "image": { "format": ..., "source": { "bytes": ... } } }`.
* ExecuteTurn1: added retry logic with exponential back-off for `ServiceException` and `ThrottlingException`.
* clients.go: apply configurable Bedrock timeout from `BEDROCK_TIMEOUT` env var.
* request.go: delay image validation until after Base64 retrieval.
* config.go: removed hardcoded defaults; validate all critical env vars.
* response\_processor.go: simplified token usage handling to align with `schema.BedrockApiResponse`.

### Fixed

* Corrected order of Base64 generation and image validation in `execute_turn1.go`.
* Fixed ModelId reference to use `BEDROCK_MODEL` env var.
* Fixed image block construction to comply with Bedrock Converse API schema.
* Ensured `error` import from Go std to properly use `errors.As`.
* Removed misuse of `io.ReadAll` and `Close` on byte slices.

---

## \[1.3.0] - 2025-05-18

### Changed

* **API Migration:**

  * Migrated from Bedrock Converse API to InvokeModel API for improved compatibility with AWS SDK
  * Restructured request/response handling to work with the newer API contract
  * Fixed type safety issues with error handling and schema compatibility
* **Architecture Support:**

  * Migrated to ARM64/Graviton Lambda for improved cost and performance
  * Updated Go version from 1.21 to 1.22
  * Added platform-specific build flags

### Added

* Enhanced build process with optimized Go compiler flags
* Better error diagnostics in the Docker build script
* Platform detection in the build script for cross-compilation
* Updated documentation to reflect API changes

---

## \[1.2.0] - 2025-05-17

### Changed

* **Major migration:**

  * All workflow and Lambda modules now fully use shared `schema`, `logger`, and `errors` packages.
  * Removed all custom/request-specific type definitions for workflow state, prompt, images, and error contracts.
* **Centralized AWS client initialization and config validation.**
* **Bedrock integration:**

  * Now uses the Claude 3.7 Sonnet Converse API exclusively.
  * Bedrock messages are always constructed via the shared schema, supporting inline and S3-based Base64 image references.
* **Hybrid Base64 support:**

  * Large images automatically use S3 temporary storage; retrieval and embedding for Bedrock handled via schema helpers.
* **Logging overhaul:**

  * Replaced all raw log output with JSON-structured logs via the shared logger.
  * Correlation IDs and context fields now populate all logs for distributed traceability.
* **Error handling overhaul:**

  * All errors are now `WorkflowError` types (typed, retryable, and with full context).
  * All errors surfaced in both Lambda error and within `VerificationContext.Error` for Step Functions compatibility.

### Added

* Helper utilities in `request.go` and `response_processor.go` for request/response validation and workflow state management.
* Standardized error creation and propagation in all modules.
* Docs/README updated for new architecture.

---

## \[1.1.0] - 2025-05-15

### Added

* Initial implementation of shared schema, logger, and error packages.
* Core support for hybrid Base64 image storage (inline/S3).
* Validation and builder helpers for all major workflow state fields.

---

## \[1.0.0] - 2025-05-14

### Added

* First release of vending machine verification solution.
* Lambda function skeletons and event-driven Step Functions integration.
