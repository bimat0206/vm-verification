# Changelog

All notable changes to the FinalizeWithErrorFunction will be documented in this file.

## [1.0.0] - 2025-05-20

### Added
- Initialized `FinalizeWithErrorFunction` Lambda to handle errors from the Step Functions workflow.
- Implements parsing of Step Functions error payloads.
- Loads existing `VerificationContext` from S3 (`initialization.json` at `.../processing/initialization.json`) or creates a minimal context if loading fails.
- Updates `VerificationContext` with standardized `schema.ErrorInfo`, failure status, and appends to status and error history.
- Passes through existing `partialS3References` and optionally stores a new `error-summary.json` to S3.
- Returns a standardized output payload (`FinalizeWithErrorOutput`) for subsequent workflow steps.
- Includes structured logging and configuration management.
