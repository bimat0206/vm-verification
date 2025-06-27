# Vending Machine Verification System - Changelog

All notable changes to the Vending Machine Verification System will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.3] - 2025-01-10

### Fixed
- **ExecuteTurn1Combined Validation**: Fixed validation error for PREVIOUS_VS_CURRENT verification type when no historical data exists
  - **Root Cause**: Validation logic was too strict, requiring `HistoricalContext` to be non-nil for `PREVIOUS_VS_CURRENT` verification type
  - **Solution**: Updated validation logic to allow `HistoricalContext` to be optional during initial validation
  - **Impact**: System can now proceed with verification when `sourceType` is `NO_HISTORICAL_DATA`
  - **Error Resolved**: `ValidationException: invalid verification context (Code: INVALID_INPUT, Source: Unknown)`
  - **Files Modified**: `product-approach/workflow-function/ExecuteTurn1Combined/internal/models/verification.go`

- **ExecuteTurn2Combined Conversation History**: Fixed missing Turn1 context in Turn2 Bedrock API calls
  - **Root Cause**: Turn1 prompt was being stored as `null` in raw response, causing Turn2 to skip Turn1 conversation history
  - **Issue**: Turn2 conversations were missing Turn1 context, making them appear as separate API calls instead of combined conversations
  - **Solution**: Fixed Turn1 prompt storage in both DynamoDB updates and S3 raw response storage
  - **Impact**: Turn2 now correctly includes full conversation history (Turn1 + Turn2) in single Bedrock API call
  - **Conversation Flow**: System prompt → Turn1 user message → Turn1 assistant response → Turn2 user message → Turn2 assistant response
  - **Files Modified**:
    - `product-approach/workflow-function/ExecuteTurn1Combined/internal/handler/handler.go` - Fixed prompt field in TurnResponse creation
    - `product-approach/workflow-function/ExecuteTurn1Combined/internal/handler/storage_manager.go` - Added prompt to raw data storage

### Technical Details
- **Validation Fix**: Modified `VerificationContext.Validate()` method to handle `NO_HISTORICAL_DATA` scenarios gracefully
- **Conversation History Fix**: Ensured Turn1 prompt is properly stored and available for Turn2 conversation building
- **Testing**: Both ExecuteTurn1Combined and ExecuteTurn2Combined build successfully with fixes applied
- **Verification**: Fixes resolve the core issues preventing proper two-turn conversation flow in verification workflow

## [2.0.2] - 2025-01-03

### Fixed
- **API Images Upload Render JSON Function**: Fixed critical image rendering issues by aligning with proven render-layout-go-lambda implementation
  - **Image Drawing Logic**: Fixed improper image scaling and positioning that was causing distorted or incorrectly sized images
    - Replaced `DrawImageAnchored()` with proper scaling sequence using `Push()`, `Translate()`, `Scale()`, `DrawImage()`, `Pop()`
    - Images now correctly scale and position within vending machine layout cells
  - **Text Rendering**: Added missing `splitTextToLines` function that was causing compilation errors
    - Product names now wrap intelligently across maximum 2 lines with ellipsis for overflow
    - Fixed positioning to display below product images instead of at cell bottom
    - Added fallback to "Sản phẩm" for empty product names
  - **Image Loading**: Enhanced image loading reliability and caching
    - Increased HTTP timeout from 10s to 20s for better image loading success rates
    - Added proper cache directory creation to prevent caching failures
    - Improved error handling and cache file management
  - **Placeholder Handling**: Improved image unavailable placeholders with proper styling and positioning
  - **Footer Rendering**: Fixed footer font styling to use bold font for consistency

### Changed
- **Code Modernization**: Updated deprecated Go functions for better compatibility
  - Replaced deprecated `ioutil` functions with modern `io` and `os` package equivalents
  - Updated import statements to remove deprecated packages

### Technical Details
- **Files Modified**:
  - `product-approach/api-function/api_images/upload-render-json/renderer/renderer.go` - Complete image rendering engine overhaul
- **Testing**: All compilation, unit tests, and build processes verified successful

## [2.0.1] - 2025-06-08

### Fixed
- **FetchImages Lambda Function**: Fixed compilation error in `fetch_service.go`
  - Removed unnecessary type assertion `s.dynamoDBRepo.(*repository.DynamoDBRepository).GetTableName()`
  - Changed to direct method call `s.dynamoDBRepo.GetTableName()` since the field is already typed as `*repository.DynamoDBRepository`
  - Resolved Docker build failure that was preventing Lambda deployment

- **PrepareSystemPrompt Function**: Fixed historical context requirement for PREVIOUS_VS_CURRENT verification type
  - Removed hard requirement for historical context in `processPreviousVsCurrentData` function
  - Historical context is now optional for PREVIOUS_VS_CURRENT verification type
  - Historical data comparison will be handled in the final step instead of being injected into system prompts
  - Added informational logging when historical context is not provided
  - Prevents "historical context is required for PREVIOUS_VS_CURRENT" error during prompt preparation

### Changed
- **Verification Logic**: Updated PREVIOUS_VS_CURRENT verification workflow
  - Historical data is no longer injected into system prompts during PrepareSystemPrompt step
  - Historical data will be loaded and used in the final comparison step for user comparison of previous vs current results
  - Simplified prompt preparation process for PREVIOUS_VS_CURRENT verification type

### Technical Details
- **Files Modified**:
  - `product-approach/workflow-function/FetchImages/internal/service/fetch_service.go` - Fixed type assertion error
  - `product-approach/workflow-function/PrepareSystemPrompt/internal/processors/template.go` - Made historical context optional for PREVIOUS_VS_CURRENT

## [2.0.0] - 2025-06-02

### Removed
- **BREAKING CHANGE**: Completely removed notification functionality from the verification system
  - **Notify Lambda Function**: Removed `notify` lambda function and all its related components
  - **ECR Repository**: Removed `kootoro-dev-ecr-notify-*` ECR repository
  - **Step Function Integration**: Removed `ShouldNotify` choice state and `Notify` task state from workflow
  - **CloudWatch Resources**: Removed CloudWatch logs and alarms for notify function
  - **IAM Permissions**: Removed SNS publish permissions and notify-related IAM policies
  - **API Gateway Integration**: Removed API Gateway lambda permissions for notify function
  - **Configuration**: Removed all notify-related configurations from terraform.tfvars and locals.tf

### Changed
- **Simplified Workflow**: Verification workflow now goes directly from `FinalizeAndStoreResults` to `WorkflowComplete`
  - Eliminated notification step entirely from the verification process
  - Reduced workflow complexity and resource overhead
  - Maintained all core verification functionality without notification dependencies

- **API Model Updates**: Updated API Gateway verification request model
  - Removed `notificationEnabled` property from verification request schema
  - Simplified verification initiation by removing notification preferences
  - Streamlined API payload structure for better performance

### Infrastructure Impact
- **Resource Reduction**: Eliminated 6 AWS resources (Lambda function, ECR repository, CloudWatch logs, alarms, permissions)
- **Cost Optimization**: Reduced infrastructure costs by removing unnecessary notification components
- **Simplified Deployment**: Reduced deployment complexity with fewer resources to manage
- **Enhanced Security**: Removed unused SNS permissions and simplified IAM policies

### Benefits
- **Simplified Architecture**: Cleaner, more focused verification system without notification overhead
- **Improved Performance**: Faster verification completion without notification processing delays
- **Reduced Maintenance**: Fewer components to monitor, update, and troubleshoot
- **Cost Efficiency**: Lower AWS costs with reduced resource footprint

### Migration Notes
- **No User Impact**: End users will not notice any functional changes in verification capabilities
- **Automatic Cleanup**: Existing notify-related resources will be automatically removed during deployment
- **Backward Compatibility**: All existing verification functionality remains intact
- **No Data Loss**: Historical verification data and results are preserved

### Technical Details
- **Files Modified**: 
  - `product-approach/iac/locals.tf` - Removed notify ECR repository and lambda function definitions
  - `product-approach/iac/terraform.tfvars` - Removed notify configurations
  - `product-approach/iac/main.tf` - Removed notify function ARN reference
  - `product-approach/iac/modules/step_functions/templates/state_machine_definition.tftpl` - Removed notification states
  - `product-approach/iac/test_template.tf` - Removed notify function reference

- **Terraform Plan Impact**: 
  - 6 resources to be destroyed (notify-related components)
  - 16 resources to be updated (IAM policies, Step Functions, API Gateway)
  - 1 resource to be added (updated configurations)

### Verification
- ✅ Terraform configuration validated successfully
- ✅ All remaining infrastructure components maintain proper dependencies
- ✅ Step Function workflow simplified and optimized
- ✅ No breaking changes to core verification functionality
- ✅ All IAM policies automatically updated to remove notify permissions

---

## Previous Versions

For detailed changelogs of individual components, see:
- [Infrastructure (IAC) Changelog](iac/CHANGELOG.md)
- [Frontend (Streamlit) Changelog](fe-1/CHANGELOG.md)
- [API Images Upload Render JSON Changelog](api-function/api_images/upload-render-json/CHANGELOG.md)

## Component Overview

This system consists of:
- **Infrastructure as Code (IAC)**: Terraform configurations for AWS resources
- **Frontend Application**: Streamlit-based web interface
- **Lambda Functions**: Serverless verification processing
- **API Gateway**: RESTful API endpoints
- **Step Functions**: Workflow orchestration
- **Storage**: S3 buckets for images and results
- **Database**: DynamoDB for verification records
