# Changelog

All notable changes to this project are documented here.

---
# Updated Changelog Entry for ExecuteTurn1


## [4.0.22] - 2025-05-22

### Fixed

* **Enhanced Schema Compatibility with New Format:**
  * Updated state loader to properly handle nested "verificationContext" field in initialization.json
  * Fixed image Base64 reference extraction from turn1-prompt.json imageReference structure
  * Implemented proper handling of bedrockConfiguration from system-prompt.json
  * Added support for messageStructure format in turn1-prompt.json
  * Resolved issue with reference image Base64 data access that was causing image processing failures

* **Turn-Specific Image Handling:**
  * Modified image processing to be more lenient about checking image Base64 data in Turn1
  * Changed strict validation to warning for missing checking image data since it's only needed in Turn2
  * Ensured reference image validation remains strict as it's critical for Turn1 processing
  * Added more resilient handling of image data across both turns

* **Input Validation Improvements:**
  * Enhanced input handling when StateReferences or S3References are missing but verificationId is available
  * Added fallback mechanisms to create minimal references when possible
  * Improved error messages to be more specific about missing information

### Improved

* **Validation Logic:**
  * Removed redundant format validation in validator since it's now handled by the loader
  * Enhanced image data validation to focus on required Base64 data presence
  * Added turn-aware validation that understands which images are required at each stage
  * Removed hardcoded default values in validator for better separation of concerns
  * Simplified token usage validation to focus on essential structural elements

* **Error Handling:**
  * Added more descriptive error messages for easier troubleshooting
  * Improved logging of image processing steps and decisions
  * Enhanced fallback mechanisms when loading from different schema formats
  * Transformed critical errors into warnings when appropriate for the current turn

### Code Quality

* **Architecture and Maintainability:**
  * Created more robust loader functions with better error handling
  * Implemented turn-aware processing that distinguishes between Turn1 and Turn2 requirements
  * Added graceful fallbacks for missing fields rather than failing completely
  * Improved logging for clearer debugging of schema-related issues
  * Enhanced function interfaces for better type safety

### Technical Details

* Modified image processing to continue when checking image lacks Base64 data in Turn1:
  ```go
  // For checking image, only log a warning since it's not needed in Turn1
  if images.GetChecking() != nil && !images.GetChecking().HasBase64Data() {
      log.Warn("Checking image missing Base64 data, but continuing since it's not needed for Turn1", nil)
      // No error return here - we continue processing
  }
  ```

* Enhanced LoadImages to set default values for checking image:
  ```go
  // Create a minimal checking image if needed
  if images.Checking == nil && images.CheckingImage == nil {
      checkingImage := &schema.ImageInfo{
          Format: "png", // Default format
          StorageMethod: "s3-temporary",
      }
      
      images.Checking = checkingImage
      images.CheckingImage = checkingImage
  }
  ```

* Modified LoadVerificationContext to properly extract from nested structure:
  ```go
  var wrapper struct {
      VerificationContext *schema.VerificationContext `json:"verificationContext"`
      SchemaVersion       string                      `json:"schemaVersion"`
  }
  ```

* Enhanced LoadTurn1Prompt to extract Base64 references from the new structure:
  ```go
  var promptWrapper struct {
      ImageReference struct {
          ImageType string `json:"imageType"`
          Base64StorageReference struct {
              Bucket string `json:"bucket"`
              Key    string `json:"key"`
          } `json:"base64StorageReference"`
      } `json:"imageReference"`
  }
  ```

This update ensures proper compatibility with the new schema format while making the function more resilient to variations in input data, particularly understanding the different image requirements for Turn1 versus Turn2 processing.
## [4.0.21] - 2025-05-22

### Fixed

* **Schema Compatibility with New Format:**
  * Updated state loader to support nested "verificationContext" field in initialization.json
  * Fixed handling of the new turn1-prompt.json format with "messageStructure" instead of direct text field
  * Added support for extracting Bedrock configuration from "bedrockConfiguration" top-level field
  * Resolved validation issues with machine structure inconsistencies between different files

### Changed

* **Response Format Updated for New Schema:**
  * Changed FileTurn1Response from "turn1-response.json" to "turn1-raw-response.json"
  * Removed FileTurn1Thinking as thinking content is now included in the Turn1 response
  * Updated SaveTurn1Response to combine response and thinking content in a single file
  * Adjusted response format to conform to the new schema definition

### Improved

* **Error Handling and Default Values:**
  * Enhanced fallback mechanisms when loading from different schema formats
  * Added better warnings for missing fields rather than failing completely
  * Improved field validation to handle different data types more flexibly
  * Enhanced log messages to provide more context about schema-related issues

### Technical Details

* Modified `state/loader.go` to handle the new schema structure:
  ```go
  // Load verification context from the nested structure
  var wrapper struct {
      VerificationContext *schema.VerificationContext `json:"verificationContext"`
      SchemaVersion       string                      `json:"schemaVersion"`
  }
  ```

* Updated prompt loading logic to extract from new structure:
  ```go
  var newFormatPrompt struct {
      MessageStructure struct {
          Content []struct {
              Type string `json:"type"`
              Text string `json:"text"`
          } `json:"content"`
          Role string `json:"role"`
      } `json:"messageStructure"`
      // ...other fields...
  }
  ```

* Modified Bedrock configuration loading to handle the new location:
  ```go
  var newSystemPrompt struct {
      BedrockConfiguration struct {
          AnthropicVersion string  `json:"anthropicVersion"`
          MaxTokens        int     `json:"maxTokens"`
          // ...other fields...
      } `json:"bedrockConfiguration"`
  }
  ```

* Updated response saving to include thinking content directly in the same file:
  ```go
  turn1ResponseWithThinking := map[string]interface{}{
      // ...other fields...
      "response": map[string]interface{}{
          "content": ...,
          "thinking": thinkingContent
      }
  }
  ```

This update ensures compatibility with the updated schema format while maintaining backward compatibility with existing workflows.

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

