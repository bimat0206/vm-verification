# Changelog

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
