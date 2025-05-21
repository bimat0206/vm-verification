# Changelog

All notable changes to the PrepareTurn1Prompt function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.0.19] - 2025-05-21
### Changed
- Standardized on `s3References` field name throughout codebase for consistency
- Modified `Output` struct in `references.go` to use `s3References` instead of `references` 
- Updated `NewOutput` function to work with `s3References` field
- Changed `OutputToEnvelope` function to properly map `output.S3References` to envelope
- Updated all reference handling in `saver.go` to use `s3References` consistently
- Modified `processInput` function in `main.go` to use `input.S3References` instead of `input.References`
- Updated logging statements to reference the correct field names
- Maintained backward compatibility with `input.References` when needed


## [4.0.18] - 2025-05-20
### Fixed
- Fixed duplicated verification IDs in S3 paths for `prompts_turn1-prompt` and `processing_turn1-metrics`
- Modified s3state SaveToEnvelope usage to avoid double-adding verification ID to paths
- Updated BuildS3Key helper function to work correctly with SaveToEnvelope
- Removed unused imports in `bedrock_prep.go` ("context" and "workflow-function/shared/s3state")
- Fixed S3 key structure to follow the standard format: `verificationId/category/filename.json`
- Resolved path duplication issue causing incorrect S3 keys like `verificationId/category/verificationId/category/filename.json`

## [4.0.17] - 2025-05-20
### Added
Enhanced Reference Categorization and Logging:

The ValidateInput method now analyzes and logs reference categories, making it easier to identify missing or incorrect references during validation.
Categories are extracted from reference keys and counted to provide a high-level overview of the reference structure.


New Base64 Reference Validation:

Added a new ValidateImageReferences method that specifically checks for proper Base64 storage references in image data.
This validation ensures that S3 temporary storage has both the bucket and key properly set when Base64 data is generated.
Provides detailed logging about Base64 storage references for troubleshooting.


Reference Structure Validation:

Added a new ValidateReferenceStructure method to ensure that references contain all necessary categories.
This validation checks for required categories such as initialization, images, and prompts to ensure the workflow can proceed correctly.
Provides clear error messages when required categories are missing, with context about what was found versus what was expected.


Reference Accumulation Validation:

Implemented a new ValidateReferenceAccumulation method that compares input and output references to ensure all input references are preserved in the output.
This validation is critical for enforcing the reference accumulation pattern across the workflow.
Provides detailed logging about reference counts and accumulation metrics.
Enhanced NewOutput Call:

Modified the call to state.NewOutput() to include input.References as the fourth parameter, which ensures all existing references are preserved in the new output object.


Updated SaveTurn1Prompt Call:

Changed the call to stateSaver.SaveTurn1Prompt() to include the input parameter, which allows the saver to access and preserve the references from the input.


Improved Logging:

Added detailed logging with reference counts to track the reference accumulation process:

Logs the reference count from the input
Logs both input and output reference counts when processing is completed
Calculates and logs the number of new references added during processing




Additional Context in Log Messages:

Added reference counts to log messages for better debugging and operational visibility
Enhanced existing log messages with more context information

### Fixed
- Fixed reference accumulation pattern to maintain complete state context throughout workflow
- Modified `NewOutput` in references.go to accept and preserve existing references
- Added `CopyReferences` helper function for consistent reference handling
- Enhanced `EnvelopeToInput` to ensure proper copying of all references
- Updated `SaveTurn1Prompt` to accept input and preserve existing references
- Improved image processor to ensure Base64 references are properly set and validated
- Fixed missing Base64 storage reference in Turn 1 prompt imageReference object
- Enhanced imageReference structure with explicit base64StorageReference field
- Added comprehensive logging for reference accumulation and Base64 storage tracking
- Fixed broken chain of state information affecting downstream functions
- Implemented robust error handling for Base64 reference validation

## [4.0.16] - 2025-05-20

### Fixed
- Fixed compilation error in saver.go related to non-existent Base64Reference field
- Updated S3 reference handling to use proper ImageInfo fields (Base64S3Bucket, Base64S3Key)
- Improved compatibility with ImageInfo schema by using GetBase64S3Key() helper method
- Fixed "refImage.Base64Reference undefined" error in state saver

## [4.0.15] - 2025-05-20

### Fixed
- Fixed inconsistent S3 reference key generation 
- Updated the s3state.SaveToEnvelope method to ensure consistent reference keys in output
- Fixed issue where "processing_turn1-metrics" reference was missing from output
- Added proper logging for all stored S3 references
- Ensured metrics files are properly referenced in function output

## [4.0.14] - 2025-05-20

### Fixed
- Fixed incorrect S3 path structure for storing prompts
- Updated SaveToEnvelope method to correctly use date-based path structure
- Fixed path format to use YYYY/MM/DD/verificationID/category/filename pattern
- Resolved issue where prompts were stored in the wrong S3 location
- Ensured proper extraction of date components from verification ID

## [4.0.13] - 2025-05-20

### Fixed
- Fixed unclosed conditional statements in turn1-layout-vs-checking template
- Added missing {{end}} tags for if statements in template
- Fixed empty schemaVersion field in initialization.json
- Updated schemaVersion to "2.0" to match expected format
- Resolved "failed to process template: InternalException: Internal error in component: template-loading" error

## [4.0.12] - 2025-05-20

### Fixed
- Fixed template loading error in turn1-previous-vs-current template
- Updated template to use array indexing with add function instead of split function
- Harmonized template function usage across all templates
- Resolved "internal error in component: template-loading" issue

## [4.0.11] - 2025-05-20

### Fixed
- Fixed template loading issues by following PrepareSystemPrompt design pattern
- Simplified template loader initialization by removing explicit CustomFuncs
- Added template version logging for debugging
- Fixed default TEMPLATE_BASE_PATH value to match Docker container path
- Made template path more robust with fallback to /opt/templates
- Updated templateLoader initialization to avoid function conflicts

## [4.0.10] - 2025-05-20

### Fixed
- Fixed template loading issues and Docker template path inconsistency
- Fixed "template-loading: Internal error" by correcting Dockerfile template paths
- Simplified template processing by using registered template functions
- Fixed template function `sub` issue by replacing with `add` with negative number
- Corrected template path in Dockerfile from `/build/docker_templates` to `/build/templates`
- Removed unnecessary template directory copying in build script
- Added more robust template path handling in the Docker container

## [4.0.9] - 2025-05-20

### Fixed
- Fixed "template: turn1-layout-vs-checking:9:129: executing \"turn1-layout-vs-checking\" at <.RowCount>: invalid value; expected int" error
- Enhanced type handling for machineStructure.rowCount and columnsPerRow by supporting both int and float64 types
- Added robust type conversion for JSON numeric values (which default to float64 in Go)
- Improved TotalPositions calculation to handle different numeric types
- Made template data processing more resilient to different JSON number representations

## [4.0.8] - 2025-05-20

### Fixed
- Fixed "Failed to process template: InternalException: Internal error in component: template-execution" error
- Added custom template functions including "sub" for array indexing arithmetic
- Updated layout-vs-checking template to use correct array indexing
- Modified RowLabels handling to store both string format and array format for template access
- Enhanced error logging for template execution with data key diagnostics
- Added template execution error recovery with detailed debugging

## [4.0.7] - 2025-05-20

### Fixed
- Fixed "Missing required field: verificationContext.verificationType" error by implementing robust loading of nested verification context structures
- Enhanced initialization.json loading to handle both direct and nested verification context structures
- Added fallback to raw JSON parsing with multiple structure handling attempts
- Improved error handling with better logging of field extraction
- Fixed missing verificationType field in nested structure by implementing proper structure traversal
- Added additional validation to ensure verification context is never nil
- Added JSON marshaling/unmarshaling to handle complex nested structures

## [4.0.6] - 2025-05-20

### Fixed
- Fixed "Missing required field: verificationContext.verificationId" and "Missing required field: verificationContext.verificationType" errors by ensuring all required fields are properly set in the verification context
- Added fallback mechanism to use top-level verificationId and verificationType when missing in the verification context
- Added support for extracting verificationType from input.Summary field when available
- Added support for setting missing verification context fields from input (verificationType, status)
- Added automatic generation of verificationAt timestamp when missing
- Enhanced error handling for verification context validation
- Added detailed logging of verification context fields for easier troubleshooting

## [4.0.5] - 2025-05-20

### Fixed
- Fixed "failed to load images from metadata" error by updating the loadImagesFromMetadata method to handle both new and old metadata formats
- Added support for the complex metadata structure produced by FetchImages function
- Implemented backward compatibility for old metadata format
- Added helper functions for extracting values from complex metadata structure
- Enhanced error handling and logging for metadata processing

## [4.0.4] - 2025-05-20

### Fixed
- Fixed duplicate case error in processor.go for "s3-temporary" storage method
- Consolidated image processing logic for S3 temporary storage
- Improved handling of S3 URL processing within the same storage method case

## [4.0.3] - 2025-05-20

### Fixed
- Fixed compilation errors in image handling code
- Updated field names in schema.ImageInfo usage to match shared package
- Replaced deprecated constants with string literals for storage methods
- Simplified image loading from S3 URLs
- Fixed error handling in URL processing
- Fixed "Reference image not found in metadata" error by updating loader.go to use "referenceImage" and "checkingImage" keys instead of "reference" and "checking" to match the actual S3 bucket metadata format

## [4.0.2] - 2025-05-20

### Fixed
- Added fallback mechanism to create image info from URLs in verification context when references are missing
- Improved image processing with support for direct S3 URL processing
- Fixed "Reference image reference not found" error when only URL is available
- Extended the processor to handle cases where image info doesn't have proper storage method set
- Enhanced checking image handling for better compatibility
- Added robust detection and processing of S3 URLs directly from verification context
- Improved error messages for troubleshooting image reference issues

### Added
- New processFromS3URL method in image processor for handling S3 URLs directly
- Support for automatically detecting and setting content type based on image format
- Comprehensive logging for image processing operations

## [4.0.1] - 2025-05-19

### Fixed
- Fixed validation error for turnNumber field in S3 state envelope input
- Enhanced state envelope parsing to handle different schema versions
- Improved input field handling for both direct invocation and Step Functions
- Updated reference key detection with support for multiple naming patterns
- Added robust validation for S3 references with clear error messages
- Fixed main.go handler to properly handle different input formats
- Enhanced compatibility with state machine execution
- Improved error recovery for missing or malformed references

### Added
- Support for alternative reference field names (References/S3References)
- Detailed logging of reference resolution for easier troubleshooting
- Default values for required fields in envelope conversion
- Flexible reference key matching to handle different naming conventions

## [4.0.0] - 2025-05-20

### Added
- Complete refactoring to use shared packages architecture
- Integration with `workflow-function/shared/s3state` package
- Modular codebase with clear separation of concerns
- New directory structure with focused modules
- Enhanced image processing with better format detection
- Updated template handling with improved error recovery
- Comprehensive logging with structured context
- Integrated Bedrock message creation using shared schema

### Changed
- Migrated from custom error types to shared error package
- Replaced custom logging with shared logger package
- Switched from direct S3 operations to s3state package
- Updated prompt generation to use shared template loader
- Improved validation with separate validation module
- Enhanced documentation with clear architecture explanation

### Removed
- Custom S3 client implementation
- Direct AWS SDK usage for S3 operations
- Custom template management code
- Tightly coupled image processing logic

## [2.0.1] - 2025-05-19

### Added
- Enhanced Dockerfile for AWS Lambda ARM64 deployment
- Improved build script with colored output and better error handling
- Automatic shared module resolution and dependency management
- Support for proper Go module resolution in Docker builds

### Fixed
- Docker build issues with shared module dependencies
- Template path resolution in containerized environment
- ARM64 architecture support for AWS Lambda (Graviton)
- Proper directory structure checking in build script

## [2.0.0] - 2025-05-19

### Added
- New S3 state management architecture
- S3 reference-based input/output handling
- Image processing with multiple storage method support
- Validator package for input validation
- Unit tests for validator component
- Comprehensive README documentation
- Dockerfile for containerized deployment
- Build and deployment script with retry logic

### Changed
- Transformed from payload-based to S3 state management architecture
- Refactored code into separate packages with clear responsibilities
- Updated template handling to use existing template structure
- Enhanced error handling with shared error types
- Improved logging with structured context fields

### Removed
- Legacy payload-based input/output handling
- Direct Base64 image processing in main function
- Monolithic function structure

## [1.0.0] - 2025-01-15

### Added
- Initial implementation of PrepareTurn1Prompt function
- Support for layout-vs-checking verification type
- Support for previous-vs-current verification type
- Basic template handling for prompt generation
- Direct Base64 image processing
