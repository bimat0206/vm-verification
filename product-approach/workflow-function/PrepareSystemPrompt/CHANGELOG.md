# Changelog

PrepareSystemPrompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.0.7] - 2025-05-20

### Fixed
- **Critical**: Fixed reference forwarding issue that was causing "Reference image reference not found" errors in PrepareTurn1Prompt
- Implemented comprehensive reference accumulation pattern to ensure complete data flow between workflow functions
- Enhanced S3 state adapter to preserve all references from previous workflow steps
- Fixed the root cause of missing images_metadata and processing_layout-metadata references
- Ensured all verification data flows properly through the entire workflow pipeline

### Technical Details
- Added AccumulateAllReferences function to S3StateAdapter that properly combines incoming and outgoing references
- Modified both S3 reference and direct JSON handlers to implement reference accumulation
- Added detailed logging to track reference propagation for easier debugging
- Implemented safer reference handling with null-checks and fallback mechanisms
- Preserved backward compatibility with existing code while fixing forward reference management

## [4.0.6] - 2025-05-19

### Fixed
- **Critical**: Fixed duplicate "prompts" directory in system prompt S3 storage paths that was causing empty prompt bucket issue
- Modified S3StateAdapter.StoreSystemPrompt to build explicit path instead of using CategoryPrompts constant
- Resolved S3 path structure from incorrect `prompts/YYYY/MM/DD/verificationId/prompts/system-prompt.json` to correct `YYYY/MM/DD/verificationId/prompts/system-prompt.json`
- System prompts are now properly stored and retrievable at the expected S3 location
- Ensured consistent S3 key generation across all storage operations

### Technical Details
- Changed StoreSystemPrompt method to explicitly construct the full S3 key path
- Replaced `(*s.stateManager).StoreJSON(s3state.CategoryPrompts, key, prompt)` with `(*s.stateManager).StoreJSON("", key, prompt)` where key includes the full path
- Updated path format to `{datePartition}/{verificationId}/prompts/system-prompt.json`
- Maintained backward compatibility with existing path parsing logic

## [4.0.5] - 2025-05-19

### Fixed
- Fixed duplicate "prompts" directory in system prompt S3 paths that was causing empty prompt bucket issue
- Modified S3StateAdapter.StoreSystemPrompt to use CategoryPrompts constant instead of hardcoding the path

## [4.0.4] - 2025-05-19

### Added
- Implemented smart recovery for missing verification contexts:
  - Added auto-recovery in s3state.LoadStateFromEnvelope to create verification context from envelope data
  - Added fallback initialization in handler.processS3ReferenceInput as a safety measure
  - Enhanced logging to identify nil verification context scenarios
  - Added auto-detection of verification type based on reference keys

### Changed
- Improved error handling for state with missing verification context
- Enhanced log messages with additional context for state recovery

## [4.0.3] - 2025-05-19

### Fixed
- Fixed nil pointer dereference by adding nil checks in:
  - TemplateProcessor.BuildTemplateData
  - handler.processS3ReferenceInput
  - handler.processDirectJSONInput
- Improved error handling when state contains invalid or missing verification context
- Added explicit error messages for nil verification contexts to facilitate debugging
- Enhanced logging with additional context for nil verification errors

## [4.0.2] - 2025-05-19

### Fixed
- Fixed S3 date partition extraction in `GetDatePartitionFromReference` by replacing `filepath.SplitList` with `strings.Split`
- Corrected path joining method to ensure forward slashes are used consistently in S3 keys

## [4.0.1] - 2025-05-19

### Fixed
- Corrected logger implementation to use interface type `logger.Logger` instead of pointer type `*logger.Logger`
- Updated logger initialization to match the shared package's New function
- Removed unused import of "workflow-function/shared/s3state" in handler.go
- Fixed timing measurements by adding missing start variable in handler methods
- Removed non-existent Metadata field from SystemPrompt struct in adapters/bedrock.go

## [4.0.0] - 2025-05-19

### Added
- Complete refactoring to modular architecture with separated packages:
  - `internal/config` - Configuration management
  - `internal/models` - Data models and structures
  - `internal/adapters` - External service integrations
  - `internal/processors` - Business logic processing
  - `internal/handlers` - Lambda request handling
- Enhanced S3 state adapter with date-based hierarchical structure
- Support for both direct JSON and S3 reference input types
- Improved error handling with context-rich error types
- Structured logging through shared logger package

### Changed
- Complete reorganization of code structure for better maintainability
- Migrated to shared packages for standard functionality
- Updated handler to support both input types seamlessly
- Improved response structure with S3 references
- Enhanced Bedrock adapter with configurable settings
- Better template processing with structured data models

### Fixed
- Various edge cases in S3 URL validation
- Template version handling issues
- More consistent error reporting
- Improved validation logic for different verification types

## [3.0.0] - 2025-05-30

### Added
- Integration with the shared `s3state` package for standardized state management
- New `state_adapter.go` to bridge between internal models and shared s3state
- Support for both original and shared s3state path formats in validation
- Configuration options for Bedrock parameters (TEMPERATURE, TOP_P, THINKING_TYPE)
- Helper functions to get and set default template versions

### Changed
- Updated `go.mod` to include the shared s3state package
- Refactored `main.go` to use the shared s3state package through an adapter
- Moved hardcoded template versions to a configurable setting
- Updated template manager to support both path formats for s3state

### Fixed
- Removed hardcoded temperature and top-p values
- Improved validation to support both date-based formats
- Fixed potential incompatibilities with shared s3state path patterns

## [2.1.0] - 2025-05-25

### Changed
- Removed dependency on shared/s3utils package
- Added internal S3 URL parsing and validation in validation package
- Simplified dependency graph by reducing external package dependencies
- Enhanced error reporting for S3 URL validation

### Added
- Custom S3 URL parser with support for various S3 URL formats
- Internal image format validation

## [2.0.0] - 2025-05-24

### Added
- S3 State Management integration with date-based hierarchical storage
- Date-based path handling with format: `{year}/{month}/{day}/{verificationId}/...`
- S3 reference-based input/output model with envelope structure
- Robust date extraction from verification IDs and timestamps
- Date-aware error handling with improved context
- Structured logging with date partition information
- Enhanced configuration management with timezone support
- New sample events for testing S3 reference-based input/output
- Detailed REFACTORING-SUMMARY.md document

### Changed
- Complete reorganization into modular directory structure:
  - config/ - Configuration management
  - logging/ - Structured logging
  - models/ - Data models
  - state/ - State management
  - template/ - Template management
  - validation/ - Input validation
- Improved error types with date context
- Enhanced validation with S3 reference support
- More efficient memory usage for large inputs
- Updated build script to support new directory structure
- Enhanced Dockerfile with environment variable defaults

### Fixed
- Issues with timestamp parsing in various formats
- S3 path construction edge cases
- Legacy compatibility issues
- Improved error handling for S3 operations