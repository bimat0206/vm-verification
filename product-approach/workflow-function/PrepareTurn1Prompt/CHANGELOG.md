# PrepareTurn1Prompt Changelog

## v1.2.0 - 2025-05-17

### Bug Fixes and Improvements

- Fixed template loading error by correctly handling template naming conventions (replacing underscores with hyphens)
- Fixed Docker build issues by improving shared module dependency handling
- Added panic recovery middleware to gracefully handle unexpected errors
- Improved error logging with detailed stack traces for better debugging
- Updated retry-docker-build.sh script to properly handle module dependencies

## v1.1.0 - 2025-05-17

### Migration to Shared Packages

- Migrated to use shared schema package for standardized types
- Migrated to use shared logger package for consistent logging
- Migrated to use shared templateloader package for template management
- Migrated to use shared bedrock package for Bedrock API interactions
- Migrated to use shared errors package for standardized error handling
- Removed duplicate code and unused functions
- Updated code to follow best practices
- Ensured all AWS dependencies are properly included

### Code Improvements

- Added proper error handling with structured errors
- Improved logging with structured logs
- Standardized template loading and rendering
- Simplified Bedrock message creation using shared utilities
- Removed hardcoded values and replaced with constants from schema
- Added proper validation for input parameters
- Improved code organization and readability
