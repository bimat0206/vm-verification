# Changelog

All notable changes to this project are documented here.

---

## [4.0.20] - 2025-05-21

### Fixed

* **Critical Fix: S3 References Mapping Between Step Function and Lambda:**
  * Fixed structural mismatch between dynamic S3 references and expected StateReferences format
  * Added proper mapping function to convert from flat map structure to structured references
  * Enhanced validation to handle partially mapped references without failing
  * Implemented better error reporting with clear indication of missing references

### Added

* **Enhanced Reference Handling:**
  * Added `MapS3References()` method in StepFunctionInput to handle dynamic reference keys
  * Implemented reference format detection and conversion
  * Added detailed logging for reference mapping process
  * Created more robust validation that distinguishes between critical and optional references

### Improved

* **Error Handling and Debugging:**
  * Enhanced error messages to include both field names and expected key patterns
  * Added verification ID consistency checks throughout the workflow
  * Improved state loader to continue with partial state when possible
  * Added fallback mechanism for creating default Bedrock configuration

### Technical Details

* Modified `StepFunctionInput` in `internal/types.go` to support map-based dynamic references:
  ```go
  type StepFunctionInput struct {
      StateReferences *StateReferences               `json:"stateReferences"`
      S3References    map[string]*s3state.Reference  `json:"s3References"` // Changed to map structure
      VerificationId  string                         `json:"verificationId"`
      // ...
  }
  ```

* Added mapping function to convert reference formats:
  ```go
  func (input *StepFunctionInput) MapS3References() *StateReferences {
      // Maps dynamic keys like "processing_initialization" to structured fields
      // ...
  }
  ```

* Updated handler to use new mapping approach:
  ```go
  if input.StateReferences == nil {
      input.StateReferences = input.MapS3References()
  }
  ```

* Enhanced state loader to handle partially mapped references and create default values when needed

* Updated validation to distinguish between critical and optional references:
  ```go
  // Required references - critical for functionality
  if refs.Initialization == nil {
      missingRefs = append(missingRefs, "Initialization (processing_initialization)")
  }
  
  // Optional references - log warnings but don't fail validation
  if refs.ImageMetadata == nil {
      optionalMissing = append(optionalMissing, "ImageMetadata (images_metadata)")
  }
  ```

This update resolves the incompatibility issue between the Step Function input format and the Lambda's expected reference structure, ensuring proper handling of state references throughout the verification workflow.

## [4.0.19] - 2025-05-21

### Fixed

* **Field Mismatch Between Step Function and Lambda:**
  * Fixed issue where the Step Function expected `s3References` but the Lambda was using `stateReferences`
  * Added support for both field names in StepFunctionInput and StepFunctionOutput structs
  * Updated handler to use either `s3References` or `stateReferences` field for compatibility
  * Ensured all output responses include both `s3References` and `stateReferences` fields with identical content
  * Also updated error handling to ensure proper field population in error cases

### Technical Details

* Added `S3References` field to both input and output structs in `types.go`
* Modified the handler to check and use either field:
  ```go
  if input.StateReferences == nil && input.S3References == nil {
      return nil, wferrors.NewValidationError("Neither StateReferences nor S3References is provided", nil)
  }
  if input.StateReferences == nil {
      input.StateReferences = input.S3References
  }
  ```
* Updated response creation to populate both fields consistently:
  ```go
  if output.StateReferences != nil {
      output.S3References = output.StateReferences
  }
  ```
* Enhanced error response creation to maintain field consistency
* Added documentation in `changes-summary.md` explaining the field compatibility issue and solution

---

## [4.0.6] - 2025-05-19

### Fixed

* **Critical Build & Deployment Fixes:**
  * Resolved compilation issues related to schema package integration
  * Fixed type errors with undefined schema imports:
    * Added missing StateReferences, HybridStorageConfig, StepFunctionInput, and StepFunctionOutput
    * Fixed references to undefined error types in bedrock client
    * Updated image handling to use HasBase64Data() instead of non-existent Base64 field

* **Infrastructure Improvements:**
  * Completely redesigned Dockerfile for more robust builds:
    * Added proper CA certificates installation
    * Improved working directory structure
    * Enhanced build verification with diagnostics
    * Added cache mounting for faster builds
  * Rebuilt docker-build script with better shared module handling:
    * Reliable shared package copying with proper Go module structure
    * Enhanced error detection and reporting
    * Improved temporary build context management
    * Better Lambda deployment integration

* **Bedrock Integration Fixes:**
  * Created proper adapter between shared/bedrock and internal/bedrock
  * Fixed message structure for Bedrock Converse API
  * Improved error handling for Bedrock client interactions
  * Corrected image data handling with proper Base64 processing

### Added

* **Enhanced Architecture:**
  * Added internal types package with properly defined interfaces
  * Created adapter pattern for better integration between shared and internal types
  * Improved type safety throughout codebase
  * Better separation of concerns with clear interfaces

### Technical Details

* Resolved compiler issues by creating local type definitions:
  * HybridStorageConfig with proper field structure
  * StepFunctionInput/Output types with correct field mapping
  * BedrockClient interface with proper adapter implementation
* Fixed dependency injection with proper type conversions
* Enhanced build system for more reliable Lambda deployments

---

## [2.0.0] - 2025-05-21

### Changed

* **MAJOR ARCHITECTURAL OVERHAUL - S3 State Reference Architecture:**
  * Complete transformation from payload-based to reference-based state management
  * Implemented shared S3StateManager pattern for lightweight API contracts
  * Replaced large workflow state payloads with S3 references between Lambda functions
  * Reorganized code into focused modules with single responsibilities

### Added

* **New Directory Structure:**
  * Redesigned with true separation of concerns and modular architecture:
    * `internal/state/` - State loading/saving with S3 reference architecture
    * `internal/bedrock/` - Bedrock client integration with shared package
    * `internal/validation/` - Input validation with schema validation
    * `internal/handler/` - Core business logic coordination
    * `internal/config/` - Enhanced configuration management

* **State Management Components:**
  * `StateLoader` for loading workflow state from S3 references
  * `StateSaver` for saving workflow state to S3 with category-based organization
  * Support for thinking content storage as separate artifacts
  * Hybrid Base64 image processing with improved performance characteristics

* **Enhanced Configuration:**
  * Standardized environment variable handling with sensible defaults
  * Improved categorization by functional area (S3, Bedrock, Images, Timeouts)
  * Comprehensive validation with detailed error messages
  * Better Bedrock configuration with proper regional settings

### Improved

* **Integration with Shared Packages:**
  * Full integration with `shared/s3state` package for state management
  * Complete adoption of `shared/bedrock` for standardized API integration
  * Focused error handling using `shared/errors` workflow error types
  * Streamlined logging with `shared/logger` for structured JSON logs

* **Error Handling:**
  * Consistent error creation and handling across all components
  * Better error classification and propagation with context preservation
  * Enhanced WorkflowError creation with detailed context
  * Improved retry logic with exponential backoff

* **Performance:**
  * Reduced memory usage by eliminating large in-memory payloads
  * Improved Lambda execution efficiency with smaller state transfers
  * Better resource utilization with S3 for large data storage
  * Optimized image processing with hybrid storage model

* **Maintainability:**
  * Clean, focused modules with single responsibilities
  * No file exceeds 300 lines of code
  * Comprehensive interfaces for better testability
  * Standardized error handling and logging patterns

### Technical Details

* Migrated from direct AWS SDK usage to shared Bedrock client
* Replaced custom validation with schema-based validation
* Implemented proper error typing and classification
* Enhanced logging with correlation IDs and structured context
* Streamlined configuration with sensible defaults
* S3 state storage with category-based organization
* Proper handler dependency injection for testability

