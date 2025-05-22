# Changelog

## [2.1.1] - 2025-05-22

### Enhanced
- Specialized vending machine parsing patterns for improved accuracy and dynamic structure detection
- Enhanced PatternTypeRowColumn to match "examining each row from top to bottom" format
- Updated PatternTypeRowStatus to handle markdown headers like "## Row A ... **Status: FULL**"
- Improved PatternTypePosition to extract "- A1: Product Name" format correctly
- Made patterns generic to handle any machine structure size (not limited to specific dimensions)

### Fixed
- Resolved incorrect machine structure extraction for vending machine data
- Fixed row status pattern matching issues with markdown formatted responses
- Corrected position extraction from bullet-point formatted content
- Eliminated false positives in structure detection that caused incorrect row/column counts
- Removed hardcoded structure assumptions to ensure dynamic data handling

## [2.1.0] - 2025-05-21

### Added
- New logger adapter for improved compatibility between slog and shared logger interfaces
- Additional helper functions for counting elements in complex data structures
- Better error detection and handling in state management operations

### Changed
- Improved Dockerfile with more efficient multi-stage build based on FetchImages approach
- Enhanced build script (retry-docker-build.sh) with better error handling and temporary build context
- Consolidated duplicate operation constants in state package
- Refactored validator integration to use ValidatorInterface consistently
- Simplified processor components by removing duplicate utility functions
- Removed redundant handler in processor package to align with new architecture

### Fixed
- Resolved compiler errors in type signatures for handler.handleError function
- Fixed incorrect return type in LoadReferenceImage method
- Corrected type mismatch in observations assignment (now properly using []string)
- Fixed invalid len() operations on struct pointers with helper functions
- Resolved duplicate method declarations across multiple packages
- Fixed import errors in multiple packages
- Ensured proper return types for all methods in the state manager

## [2.0.0] - 2025-05-21

### Added
- Complete architectural transformation to reference-based S3 state management
- New state management layer in internal/state for S3-based workflow state
- Comprehensive error handling system in internal/errors with custom error types
- Specialized error handling for each component (state, handler, processor)
- Error categorization by type (Input, Process, State, System) with standardized codes
- Test examples demonstrating all aspects of error handling
- Storage-specific error types for S3 and DynamoDB operations

### Changed
- Reorganized folder structure following architecture transformation guide
- Transformed handler layer into a workflow coordinator in internal/handler
- Evolved processor layer to use strategy pattern with specialized path processors
- Refined parser layer to focus purely on data extraction
- All components now use the S3StateManager for state operations
- Updated README.md with detailed architecture documentation
- Moved from in-memory data processing to reference-based state management
- Error propagation now includes operation context and severity levels
- API responses use standardized error formats with detailed information

### Removed
- Direct DynamoDB operations from processor and handler layers
- In-memory complete state passing between components
- Custom state storage logic in favor of shared S3StateManager
- Redundant validation code replaced by comprehensive validator

## [1.4.0] - 2025-05-20

### Added
- New comprehensive validator layer implementation in validator/validator.go
- Validator interface for improved testability and integration
- Enhanced validation for each processing path (Validation Flow, Historical Enhancement, Fresh Extraction)
- Integration with shared schema types for consistent validation
- Unit tests for all validator functionality
- Detailed documentation in validator/README.md
- Backward compatibility layer for existing code

### Changed
- Refactored validation logic to use schema types where appropriate
- Enhanced error reporting with more detailed validation messages
- Improved validation completeness measurements
- Better integration with shared schema types

## [1.3.0] - 2025-05-17

### Changed
- Streamlined function to focus solely on processing Bedrock responses
- Removed unnecessary S3 operations and dependencies
- Simplified DynamoDB interactions to only include conversation history updates
- Streamlined handler implementation for better maintainability
- Reduced overall codebase complexity and dependency footprint
- Improved performance by removing extraneous service calls
- Updated Dockerfile to remove unneeded dependencies
- Enhanced build script with better error handling and verification
- Updated go.mod to remove unused module replacements

### Fixed
- Resolved compiler errors in parser components
- Fixed missing pattern field references in extractor methods
- Added proper implementation of machine structure parsing functions
- Resolved duplicate method declarations in parser package
- Renamed conflicting methods to avoid compilation errors
- Removed unused imports in storage/db_manager.go and validator/validators.go
- Fixed validator error handling in validators.go

## [1.2.0] - 2025-05-14

### Changed
- Updated Dockerfile to use multi-stage build for smaller image size
- Upgraded to Go 1.24 for improved performance and security
- Enhanced build process with optimized compilation flags
- Improved dependency management with Go workspace integration
- Added environment variable configuration for better runtime flexibility
- Updated documentation with detailed build and troubleshooting instructions

## [1.1.0] - 2025-05-14

### Changed
- Refactored parser package into multiple specialized files for better maintainability
- Improved separation of concerns with dedicated files for each parsing responsibility
- Enhanced documentation of parser components
- Added parser test suite

## [1.0.0] - 2025-05-14

### Added
- Initial implementation of ProcessTurn1Response function
- Support for both LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT use cases
- Response parsing for JSON and structured text formats
- Machine structure extraction logic
- Row state parsing and validation
- Integration with shared packages
- Comprehensive error handling and fallbacks
- Unit and integration test suite