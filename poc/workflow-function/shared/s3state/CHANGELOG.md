# Changelog

All notable changes to the s3state package will be documented in this file.

## [1.0.3] - 2025-05-19

### Fixed
- Enhanced SaveToEnvelope to detect and prevent verification ID duplication in complex paths with date components
- Added additional validation to handle year/month/day path components correctly
- Fixed incorrect key generation when using deep hierarchical paths like "2025/05/19/verif-ID/images"

## [1.0.2] - 2025-05-19

### Fixed
- Fixed nested verification ID issue in S3 paths by modifying SaveToEnvelope to avoid duplicating verification ID when it's already in the category path

## [1.0.1] - 2025-05-19

### Fixed
- Fixed broken dependencies in go.mod file
- Updated imports to use correct AWS SDK v2 packages
- Updated NoSuchKey error handling to use strings.Contains instead of aws.ErrorContains
- Fixed unused variable warning in ValidateKeyStructure function
- Added CLAUDE.md file with package documentation

### Changed
- Updated module name to github.com/kootoro/s3state
- Set Go version requirement to 1.19 for better compatibility

## [1.0.0] - 2025-04-21

### Added
- Initial release of the s3state package
- Implemented Manager interface for S3 state operations
- Added category-based organization system
- Added Reference and Envelope types for state tracking
- Implemented comprehensive error handling
- Added examples demonstrating package usage
