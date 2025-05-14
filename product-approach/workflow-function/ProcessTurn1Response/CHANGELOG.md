# Changelog

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
