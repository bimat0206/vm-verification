# Changelog

All notable changes to the FetchHistoricalVerification Lambda function will be documented in this file.

## [1.1.0] - 2025-05-17

### Changed
- Restructured code to follow standard cmd/internal pattern
- Replaced shared dbutils package with direct DynamoDB calls
- Maintained use of shared schema and logger packages
- Improved error handling for DynamoDB operations
- Updated dependency management

### Added
- DynamoDBRepository for direct DynamoDB interactions
- Exported utility functions for better package integration

### Removed
- Dependency on shared dbutils package

## [1.0.0] - 2025-05-14

### Changed
- Migrated to shared package components
- Replaced custom logger with standardized shared logger
- Replaced DynamoDB client with shared dbutils package
- Added schema version handling with shared schema package
- Improved error handling and standardized error responses
- Updated code structure to match other Lambda functions in the workflow
- Implemented dependencies.go for consistent service initialization

### Added
- ConfigVars struct for standardized environment variable configuration
- DBWrapper for consistent database interface
- Structured logging with correlation IDs

### Removed
- Function-specific DynamoDB client implementation
- Custom logging implementation
- Stand-alone validation module