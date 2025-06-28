# Changelog

## [2.5.0] - 2025-06-27

### Removed
- **BREAKING CHANGE**: Completely removed ECS Streamlit frontend module and all related AWS components
  - **Streamlit ECS Module**: Removed `modules/ecs-streamlit/` directory and all associated Terraform configurations
  - **Module References**: Removed `ecs_streamlit` module block from main.tf and all dependencies
  - **Variables**: Removed `streamlit_frontend` variable block from variables.tf and terraform.tfvars
  - **Outputs**: Removed all Streamlit frontend outputs (URLs, ECS cluster/service names, ALB DNS)
  - **VPC Security Groups**: Removed `ecs_streamlit` security group and updated outputs
  - **ECR Repository**: Removed Streamlit frontend ECR repository configuration from locals.tf
  - **Secrets Management**: Removed `ecs_config_secret` module for Streamlit configuration
  - **Monitoring**: Updated monitoring module to remove ECS and ALB monitoring for Streamlit

- **SNS Email Alerts**: Completely removed SNS email notification system and all related components
  - **SNS Resources**: Removed `aws_sns_topic.alarms` and `aws_sns_topic_subscription.email_alerts` resources
  - **CloudWatch Alarms**: Updated all alarms to disable SNS notifications (`actions_enabled = false`)
  - **Variables**: Removed `alarm_email_endpoints` variable from monitoring module and main variables
  - **Outputs**: Removed `alarm_sns_topic_arn` output from monitoring module
  - **IAM Policies**: Removed SNS publish permissions from Lambda execution role
  - **Configuration**: Removed email alert configuration from terraform.tfvars

### Changed
- **VPC Module**: Updated VPC configuration to support only React frontend
  - Changed VPC creation condition from dual frontend support to React-only
  - Updated container port default from 8501 (Streamlit) to 3000 (React)
  - Updated legacy ECS security group output to reference React security group
  - Removed Streamlit-specific security group and associated outputs

- **Monitoring Module**: Simplified CloudWatch alarm configuration
  - All alarms now use `actions_enabled = false` with empty action arrays
  - Removed email endpoint conditional logic from all alarm resources
  - Updated alarm creation conditions to remove email endpoint dependencies
  - Streamlined monitoring configuration without notification requirements

- **API Gateway Module**: Cleaned up unused variables
  - Removed `streamlit_service_url` variable that was no longer referenced
  - Simplified CORS configuration by removing Streamlit frontend URL references

### Infrastructure Impact
- **Resource Cleanup**: Significant AWS resource reduction with removal of Streamlit infrastructure
  - ECS cluster, service, task definition, and related networking components for Streamlit
  - Application Load Balancer and target groups for Streamlit frontend
  - SNS topic and email subscriptions for alarm notifications
  - Secrets Manager secret for Streamlit configuration
  - CloudWatch alarms continue to function but without email notifications

- **Simplified Architecture**: Infrastructure now supports single React frontend approach
  - Reduced complexity with single frontend service
  - Streamlined VPC configuration for React-only deployment
  - Simplified monitoring without notification requirements

### Benefits
- **Cost Optimization**: Significant cost reduction by removing duplicate frontend infrastructure
- **Simplified Maintenance**: Single frontend reduces operational complexity
- **Security Enhancement**: Reduced attack surface with fewer running services
- **Infrastructure Efficiency**: Streamlined resource usage and management

### Migration Notes
- **Automatic Cleanup**: All Streamlit and SNS resources will be automatically removed during terraform apply
- **React Frontend Preserved**: React frontend infrastructure remains fully functional
- **No API Changes**: Backend API functionality and verification workflows unchanged
- **Monitoring Continues**: CloudWatch alarms continue monitoring without email notifications

### Technical Details
- **Terraform Validation**: ✅ All configurations validated successfully
- **Dependency Check**: ✅ No broken dependencies after resource removal
- **File Cleanup**: Comprehensive removal of references across all Terraform files
- **Configuration Consistency**: ✅ Updated terraform.tfvars and all variable references

## [2.4.0] - 2025-06-09

### Added
- **Enhanced Upload API with JSON Rendering**: Upgraded `/api/images/upload` endpoint with automatic JSON layout rendering capability
  - **New Lambda Function**: Added `api_images_upload_render` Lambda function combining upload and render functionality
  - **Automatic JSON Processing**: JSON layout files uploaded to configured paths trigger automatic image rendering
  - **S3 URI Configuration**: Support for full S3 URI format in `JSON_RENDER_PATH` environment variable
  - **Cross-Bucket Rendering**: Ability to render JSON files from different S3 buckets
  - **Enhanced Response Format**: Upload responses now include render results when applicable

### Infrastructure
- **ECR Repository**: Added `api_images_upload_render` ECR repository for combined functionality
  - Configured with AES256 encryption and vulnerability scanning
  - Mutable image tags for development flexibility
  - Standardized naming convention: `*-ecr-api-images-upload-render-*`

- **Lambda Configuration**:
  - Memory: 1024 MB (optimized for image rendering)
  - Timeout: 300 seconds (5 minutes for complex layouts)
  - Environment variables: `REFERENCE_BUCKET`, `CHECKING_BUCKET`, `JSON_RENDER_PATH`, `DYNAMODB_LAYOUT_TABLE`, `LOG_LEVEL`
  - Full S3 URI support: `s3://bucket-name/path/` format

- **API Gateway Integration**: Updated existing `/api/images/upload` endpoint
  - **No Breaking Changes**: Same endpoint path maintained for frontend compatibility
  - Connected to new `api_images_upload_render` Lambda function
  - Enhanced functionality while preserving backward compatibility
  - Updated deployment configuration to reference new Lambda function

### Enhanced Features
- **JSON Layout Rendering**: Automatic detection and rendering of vending machine layout JSON files
  - Validates JSON structure for layout compatibility
  - Generates high-quality PNG images with product information
  - Organized output structure: `processed/{year}/{month}/{date}/{layoutId}_{prefix}_reference_image.png`
  - Font rendering support with custom typography
  - Product image loading with caching and fallback handling

- **Environment-Driven Configuration**: All settings configurable via Lambda environment variables
  - `JSON_RENDER_PATH`: Supports both full S3 URI (`s3://bucket/path/`) and simple path formats
  - `DYNAMODB_LAYOUT_TABLE`: Optional metadata storage in DynamoDB
  - Flexible configuration without code changes

- **Enhanced Response Format**: Upload API now returns comprehensive information
  ```json
  {
    "success": true,
    "message": "File uploaded and rendered successfully",
    "files": [...],
    "renderResult": {
      "rendered": true,
      "layoutId": 12345,
      "layoutPrefix": "20240115-143022-XYZ89",
      "processedKey": "processed/2024/01/15/12345_20240115-143022-XYZ89_reference_image.png",
      "message": "Layout rendered successfully"
    }
  }
  ```

### Technical Improvements
- **Modular Architecture**: Clean separation of concerns with dedicated packages
  - `config/`: Render configuration management
  - `renderer/`: Layout to image conversion
  - `s3utils/`: S3 operations and validation
  - `dynamodb/`: Metadata storage and retrieval
  - `logger/`: Structured logging
  - `prefix/`: Unique identifier generation

- **Advanced S3 URI Parsing**: Flexible path configuration support
  - Full S3 URI: `s3://kootoro-dev-s3-reference-f6d3xl/raw/`
  - Simple path: `raw` (backward compatibility)
  - Cross-bucket rendering capability

- **Comprehensive Error Handling**: Robust error management and logging
  - Detailed error messages for troubleshooting
  - Structured logging with JSON format
  - Graceful fallback for rendering failures

### Security & Performance
- **File Validation**: Enhanced security with comprehensive file type checking
- **Size Limits**: 10MB file size limit for Lambda compatibility
- **Memory Optimization**: Efficient image processing with 1024MB memory allocation
- **Timeout Management**: 5-minute timeout for complex layout rendering
- **IAM Permissions**: Least-privilege access to required AWS services

### Backward Compatibility
- **Zero Breaking Changes**: Existing frontend applications continue to work unchanged
- **Same API Endpoint**: `/api/images/upload` path preserved
- **Enhanced Functionality**: Additional features without disrupting existing workflows
- **Graceful Degradation**: Upload functionality works even if rendering fails

### Migration Notes
- **Automatic Enhancement**: Existing upload functionality enhanced automatically
- **No Frontend Changes**: No modifications required to frontend applications
- **Optional Features**: JSON rendering only activates for files in configured paths
- **Seamless Deployment**: Can be deployed alongside existing infrastructure

### Configuration Example
```hcl
environment_variables = {
  REFERENCE_BUCKET      = "kootoro-dev-s3-reference-f6d3xl"
  CHECKING_BUCKET       = "kootoro-dev-s3-checking-f6d3xl"
  JSON_RENDER_PATH      = "s3://kootoro-dev-s3-reference-f6d3xl/raw/"
  DYNAMODB_LAYOUT_TABLE = "VendingMachineLayoutMetadata"
  LOG_LEVEL             = "INFO"
}
```

## [2.3.0] - 2025-06-07

### Added
- **New Lambda Function**: Added `api_images_upload` Lambda function for file upload functionality
  - Go-based Lambda function with S3 upload capabilities
  - Support for multiple file types: images (.jpg, .jpeg, .png, .gif, .bmp, .webp, .tiff, .tif) and documents (.pdf, .txt, .json, .csv, .xml)
  - File validation with type whitelist and 10MB size limit
  - Base64 file content processing with proper decoding
  - Configurable bucket selection via `bucketType` query parameter (reference/checking)
  - Optional path organization with `path` query parameter
  - CORS support for cross-origin requests
  - Comprehensive error handling and structured logging

### Infrastructure
- **ECR Repository**: Added `api_images_upload` ECR repository for container image storage
  - Configured with AES256 encryption and vulnerability scanning
  - Mutable image tags for development flexibility
  - Standardized naming convention: `*-ecr-api-images-upload-*`

- **Lambda Configuration**:
  - Memory: 256 MB (optimized for file processing)
  - Timeout: 30 seconds
  - Environment variables: `REFERENCE_BUCKET`, `CHECKING_BUCKET`, `LOG_LEVEL`
  - Integrated with existing IAM roles with S3 write permissions

- **API Gateway Integration**: Added new `/api/images/upload` POST endpoint
  - Connected to `api_images_upload` Lambda function
  - AWS_PROXY integration for full request/response handling
  - CORS preflight support with OPTIONS method
  - Proper CORS headers configuration

### API Specification
- **Endpoint**: `POST /api/images/upload`
- **Query Parameters**:
  - `bucketType` (required): "reference" or "checking"
  - `fileName` (required): Name of the file to upload
  - `path` (optional): Path prefix for organizing uploads
- **Request Body**: JSON with `fileContent` field containing base64-encoded file data
- **Response Format**: JSON with upload status, S3 key, and metadata
- **CORS Support**: Full CORS configuration for web application integration

### Security Features
- **File Type Validation**: Whitelist-based file type checking
- **File Size Limits**: Maximum 10MB file size enforcement
- **Path Sanitization**: Secure path handling to prevent directory traversal
- **S3 Permissions**: Least-privilege access to specific S3 buckets only

### Enhanced Terraform Configuration
- **Updated `locals.tf`**: Added ECR repository and Lambda function configuration
- **Updated API Gateway Resources**: Added `/api/images/upload` resource definition
- **Updated API Gateway Methods**: Added POST and OPTIONS methods with proper integrations
- **Maintained Consistency**: Follows existing naming conventions and patterns

### Development Tools
- **Docker Support**: Multi-stage Docker build with Go 1.22
- **Deployment Script**: Comprehensive deployment automation
- **Testing Support**: Test payload configuration for function validation
- **Documentation**: Complete README and CHANGELOG documentation

### Benefits
- **File Upload Capability**: Enables direct file uploads to S3 buckets via REST API
- **Flexible Organization**: Support for custom path organization within buckets
- **Type Safety**: Comprehensive file type validation and size limits
- **Performance Optimized**: Efficient base64 processing and S3 upload handling
- **Security Focused**: Multiple layers of validation and sanitization

### Backward Compatibility
- **Non-Breaking Changes**: All changes are additive, no existing functionality modified
- **Existing Endpoints**: No modifications to existing API endpoints
- **Infrastructure**: Additive changes only, no resource modifications

## [2.2.0] - 2025-01-04

### Changed
- **Lambda Function Naming Standardization**: Renamed `get_conversation` Lambda function to `api_get_conversation`
  - **Function Name**: Updated from `*-lambda-get-conversation-*` to `*-lambda-api-get-conversation-*`
  - **ECR Repository**: Updated from `*-ecr-get-conversation-*` to `*-ecr-api-get-conversation-*`
  - **CloudWatch Log Group**: Updated from `/aws/lambda/*-lambda-get-conversation-*` to `/aws/lambda/*-lambda-api-get-conversation-*`
  - **API Gateway Integration**: Updated Lambda function ARN reference in `/api/verifications/{verificationId}/conversation` endpoint
  - **Terraform Configuration**: Updated all references in `locals.tf`, `terraform.tfvars`, and API Gateway modules

### Infrastructure Impact
- **Resource Rename**: 3 AWS resources will be renamed during deployment
  - Lambda function: `*-lambda-get-conversation-*` → `*-lambda-api-get-conversation-*`
  - ECR repository: `*-ecr-get-conversation-*` → `*-ecr-api-get-conversation-*`
  - CloudWatch log group: `/aws/lambda/*-lambda-get-conversation-*` → `/aws/lambda/*-lambda-api-get-conversation-*`

- **Updated Resources**: 2 AWS resources will be updated
  - API Gateway integration for conversation endpoint
  - API Gateway deployment stage variables

### Benefits
- **Consistent Naming Convention**: Aligns with existing API function naming pattern (`api_*`)
- **Improved Organization**: Clear distinction between API functions and internal workflow functions
- **Enhanced Maintainability**: Standardized naming makes infrastructure management easier
- **Better Documentation**: Function names now clearly indicate their API-facing purpose

### Migration Notes
- **Automatic Rename**: All resources will be automatically renamed during terraform apply
- **No Functional Impact**: API endpoint `/api/verifications/{verificationId}/conversation` continues to work unchanged
- **No Data Loss**: Function configuration and environment variables are preserved
- **Backward Compatibility**: No changes required to API clients or application code

### Technical Details
- **Configuration Files Updated**:
  - `locals.tf`: ECR repository and Lambda function configuration
  - `terraform.tfvars`: Memory size and timeout configuration
  - `modules/api_gateway/methods.tf`: Lambda function ARN reference
  - `modules/api_gateway/deployment.tf`: Stage variable reference
  - All backup files updated for consistency
- **Terraform Validation**: ✅ All references updated and validated
- **Naming Consistency**: ✅ Function now follows `api_*` naming convention
- **Zero Downtime**: ✅ Rename operation preserves all functionality

## [2.1.0] - 2025-01-03

### Removed
- **Lambda Functions Cleanup**: Removed 4 deprecated Lambda functions and their associated ECR repositories
  - **`store_results`**: Removed Lambda function and ECR repository configuration
    - Function was redundant with existing `finalize_results` functionality
    - ECR repository: `kootoro-dev-ecr-store-results-*` configuration removed
  - **`handle_bedrock_error`**: Removed Lambda function and ECR repository configuration
    - Error handling consolidated into `finalize_with_error` function
    - ECR repository: `kootoro-dev-ecr-handle-bedrock-error-*` configuration removed
  - **`list_verifications`**: Removed Lambda function and ECR repository configuration
    - Functionality replaced by enhanced `api_verifications_list` function
    - ECR repository: `kootoro-dev-ecr-list-verifications-*` configuration removed
  - **`get_verification`**: Removed Lambda function and ECR repository configuration
    - Basic functionality maintained through existing API endpoints
    - ECR repository: `kootoro-dev-ecr-get-verification-*` configuration removed

### Changed
- **API Gateway Integration**: Updated API Gateway to use remaining Lambda functions
  - `/api/verifications/{verificationId}` endpoint now uses `api_verifications_list` function
  - Updated deployment stage variables to reference correct function names
  - Maintained all existing API endpoints and functionality

- **Step Functions Template**: Fixed Step Functions state machine definition
  - Updated `FinalizeAndStoreResults` state to use `finalize_results` function
  - Corrected function reference in test template

### Infrastructure Impact
- **Resource Cleanup**: 8 AWS resources will be destroyed during deployment
  - 4 Lambda functions: `store_results`, `handle_bedrock_error`, `list_verifications`, `get_verification`
  - 4 ECR repositories: corresponding container image repositories
  - Associated CloudWatch log groups and IAM permissions will be automatically cleaned up

### Benefits
- **Simplified Architecture**: Reduced system complexity by eliminating redundant Lambda functions
- **Cost Optimization**: Lower AWS costs with fewer Lambda functions and ECR repositories
- **Improved Maintainability**: Fewer components to monitor, update, and troubleshoot
- **Enhanced Performance**: Streamlined workflow with consolidated functionality

### Migration Notes
- **Automatic Cleanup**: All deprecated Lambda functions and ECR repositories will be automatically removed during terraform apply
- **No Functional Impact**: Core verification functionality remains unchanged
- **Enhanced API**: Users should continue using the enhanced `api_verifications_list` endpoint for listing verifications
- **Backward Compatibility**: Existing API endpoints and workflows are preserved

### Technical Details
- **Terraform Validation**: ✅ Configuration validated successfully
- **Syntax Check**: ✅ All Terraform files properly formatted and validated
- **Dependency Check**: ✅ No broken dependencies after function removal
- **File Size Reduction**: locals.tf reduced from 508 to 427 lines (16% reduction)
- **API Gateway Updates**: ✅ Updated API Gateway integrations to use remaining functions
- **Step Functions**: ✅ Fixed Step Functions template to use correct function references
- **Test Template**: ✅ Updated test template to match new function names

## [2.0.0] - 2025-06-02

### Removed
- **BREAKING CHANGE**: Completely removed notify lambda function and all related AWS components
  - **Notify Lambda Function**: Removed `notify` lambda function definition from locals.tf
  - **Notify ECR Repository**: Removed `kootoro-dev-ecr-notify-*` ECR repository configuration
  - **Step Function Integration**: Removed `ShouldNotify` choice state and `Notify` task state from state machine
  - **CloudWatch Resources**: Removed CloudWatch log group and error rate alarm for notify function
  - **IAM Permissions**: Removed SNS publish permissions and notify function references from IAM policies
  - **API Gateway Integration**: Removed lambda permission for notify function in API Gateway
  - **Configuration Files**: Removed notify configurations from terraform.tfvars and test_template.tf

### Changed
- **Simplified Step Functions Workflow**: Updated state machine definition to eliminate notification logic
  - `FinalizeAndStoreResults` now transitions directly to `WorkflowComplete`
  - Removed conditional notification flow based on `notificationEnabled` flag
  - Streamlined workflow reduces complexity and improves performance

- **Updated API Gateway Model**: Modified verification request model
  - Removed `notificationEnabled` property from `VerificationRequest` schema
  - Simplified API payload structure for verification initiation
  - Updated API Gateway stage deployment to reflect model changes

- **IAM Policy Updates**: Automatically updated IAM policies to remove notify-related permissions
  - ECR access policy no longer includes notify repository ARN
  - Step Functions lambda invoke policy no longer includes notify function ARN
  - Removed unused SNS publish permissions

### Infrastructure Impact
- **Resource Cleanup**: 6 AWS resources will be destroyed during deployment
  - ECR repository: `kootoro-dev-ecr-notify-f6d3xl`
  - Lambda function: `kootoro-dev-lambda-notify-f6d3xl`
  - CloudWatch log group: `/aws/lambda/kootoro-dev-lambda-notify-f6d3xl`
  - CloudWatch alarm: `kootoro-dev-lambda-notify-f6d3xl-error-rate-alarm`
  - API Gateway lambda permission for notify function
  - Step Functions lambda invoke permission for notify function

- **Updated Resources**: 16 AWS resources will be updated
  - Step Functions state machine definition
  - API Gateway verification request model
  - IAM policies (ECR access, Step Functions lambda invoke)
  - Lambda function image URIs (unrelated updates)
  - ECS task definition and related resources

### Benefits
- **Simplified Architecture**: Reduced system complexity by eliminating notification components
- **Cost Optimization**: Lower AWS costs with fewer Lambda executions and reduced resource footprint
- **Improved Performance**: Faster verification completion without notification processing overhead
- **Enhanced Maintainability**: Fewer components to monitor, update, and troubleshoot

### Migration Notes
- **Automatic Cleanup**: All notify-related resources will be automatically removed during terraform apply
- **No Data Loss**: Historical verification data and results are preserved
- **Backward Compatibility**: Core verification functionality remains unchanged
- **No User Impact**: End users will not experience any functional changes

### Technical Details
- **Terraform Validation**: ✅ Configuration validated successfully
- **Plan Verification**: ✅ Terraform plan shows expected resource changes
- **Dependency Check**: ✅ All remaining resources maintain proper dependencies
- **State Machine**: ✅ Simplified workflow tested and validated

## [1.9.0] - 2025-01-02

### Added
- **New Lambda Function**: Added `api_verifications_list` Lambda function for `/api/verifications` GET endpoint
  - Go-based Lambda function with advanced filtering, pagination, and sorting capabilities
  - DynamoDB integration with efficient query patterns using Global Secondary Indexes
  - Support for filtering by verification status, vending machine ID, and date ranges
  - Pagination with configurable limits (1-100) and offset-based navigation
  - Sorting capabilities by verification date and overall accuracy
  - CORS support and comprehensive error handling

### Infrastructure
- **ECR Repository**: Added `api_verifications_list` ECR repository for container image storage
  - Configured with AES256 encryption and vulnerability scanning
  - Mutable image tags for development flexibility
  - Standardized naming convention following existing patterns

- **Lambda Configuration**:
  - Memory: 512 MB (optimized for DynamoDB operations)
  - Timeout: 30 seconds
  - Environment variables: `VERIFICATION_TABLE`, `LOG_LEVEL`
  - Integrated with existing IAM roles and permissions

- **API Gateway Integration**: Updated existing `/api/verifications` GET endpoint
  - Connected to new `api_verifications_list` Lambda function
  - Enhanced request parameter validation for query parameters
  - Updated response model to match Go Lambda function output structure

### Enhanced API Models
- **Updated Verification List Model**: Enhanced `verification_list` API Gateway model
  - Added support for all verification record fields (layoutId, overallAccuracy, etc.)
  - Proper handling of nullable fields with type unions
  - Comprehensive pagination metadata structure
  - Validation for required fields and enum values

### Performance Optimizations
- **DynamoDB Query Strategy**: Implemented intelligent query routing
  - Uses `VerificationStatusIndex` GSI for status-based queries
  - Falls back to table scan for general queries when needed
  - Application-level filtering for complex criteria
  - Optimized for minimal read capacity consumption

- **Response Optimization**:
  - Efficient JSON marshaling/unmarshaling with AWS SDK v2
  - Proper handling of optional fields and nested objects
  - Structured error responses with appropriate HTTP status codes

### Security & Monitoring
- **IAM Permissions**: Leverages existing Lambda execution role
  - DynamoDB permissions: Query, Scan, GetItem
  - CloudWatch Logs permissions for structured logging
  - Principle of least privilege access

- **Structured Logging**: Comprehensive logging with configurable levels
  - JSON-formatted logs for easy parsing and analysis
  - Request/response logging with performance metrics
  - Error tracking with detailed context information

### Documentation
- **Comprehensive Documentation**: Added `API_VERIFICATIONS_LIST_TERRAFORM.md`
  - Complete Terraform resource documentation
  - Deployment procedures and troubleshooting guide
  - Performance considerations and monitoring setup
  - Security best practices and IAM configuration

### API Specification
- **Endpoint**: `GET /api/verifications`
- **Query Parameters**:
  - `verificationStatus` (CORRECT, INCORRECT)
  - `vendingMachineId` (string)
  - `fromDate`, `toDate` (RFC3339 format)
  - `limit` (1-100, default: 20)
  - `offset` (default: 0)
  - `sortBy` (verificationAt:desc/asc, overallAccuracy:desc/asc)

- **Response Format**: JSON with `results` array and `pagination` metadata
- **Error Handling**: Structured error responses with appropriate HTTP status codes

### Development Tools
- **Deployment Script**: Comprehensive `deploy.sh` with multiple operation modes
- **Testing Suite**: Unit tests covering parameter validation and sorting logic
- **Docker Support**: Multi-stage Docker build for optimized container images

### Backward Compatibility
- **Non-Breaking Changes**: All changes maintain backward compatibility
- **Existing Endpoints**: No modifications to existing API endpoints
- **Infrastructure**: Additive changes only, no resource modifications

## [1.8.0] - 2025-05-31

### Changed
- **BREAKING CHANGE**: Simplified Step Functions state machine by removing all Handle...Error states
- Updated Step Functions module to use consolidated error handling with single `FinalizeWithError` state
- Combined `FinalizeResults` and `StoreResults` into unified `FinalizeAndStoreResults` state
- Reduced Step Functions state machine complexity from 19 to 12 states (37% reduction)
- Updated Step Functions module to version 2.1.0

### Benefits
- Streamlined error handling with single point of failure management
- Improved maintainability with fewer state transitions
- Consolidated finalization and storage operations for better performance
- Maintained all retry logic and error resilience

### Infrastructure
- Updated Step Functions state machine definition template
- Simplified error flow routing to single error handler
- Enhanced state machine performance with reduced complexity

## [1.7.0] - 2024-12-20

### Added
- Enhanced ECS Streamlit IAM policies to support dual AWS Secrets Manager access
- Added support for both CONFIG_SECRET and API_KEY_SECRET_NAME in IAM permissions
- Implemented comprehensive Secrets Manager resource ARN management for ECS tasks

### Changed
- **BREAKING**: Updated ECS Streamlit IAM policy to require both CONFIG_SECRET and API_KEY_SECRET_NAME environment variables
- Enhanced IAM policy logic to dynamically build secret ARNs based on configured environment variables
- Improved IAM policy conditions to create Secrets Manager access only when secrets are configured
- Updated IAM policy resource targeting to include multiple specific secret ARNs instead of wildcard access

### Security
- **MAJOR SECURITY IMPROVEMENT**: Implemented least-privilege access for Secrets Manager
- Restricted IAM permissions to only the specific secrets required by the application
- Enhanced security by removing wildcard Secrets Manager access in favor of targeted resource permissions
- Improved IAM policy structure for better security auditing and compliance

### Infrastructure
- Updated ECS task role IAM policies to support the new secrets-based configuration approach
- Enhanced Terraform locals to dynamically calculate required secret ARNs
- Improved IAM policy attachment logic based on configured environment variables
- Added backward compatibility for deployments without secrets configuration

### Migration Required
- **ACTION REQUIRED**: Apply Terraform changes to update ECS task role IAM permissions
- **ACTION REQUIRED**: Ensure both CONFIG_SECRET and API_KEY_SECRET_NAME are properly configured in ECS Task Definition
- **ACTION REQUIRED**: Verify that both secrets exist in AWS Secrets Manager before deployment
- **ACTION REQUIRED**: Restart ECS service after applying IAM policy changes

### Deployment Notes
- ECS service will automatically restart when IAM policies are updated
- Application will gain access to both configuration and API key secrets
- No downtime expected during IAM policy updates
- Monitor CloudWatch logs to verify successful secret access after deployment

## [1.6.0] - 2024-12-19

### Added
- Added new ECS Streamlit configuration secret management using AWS Secrets Manager
- Created `ecs_config_secret` module for centralized configuration storage
- Implemented CONFIG_SECRET environment variable approach for enhanced security

### Changed
- Replaced hardcoded environment variables (REGION, DYNAMODB_TABLE, S3_BUCKET, AWS_DEFAULT_REGION, API_ENDPOINT) with CONFIG_SECRET
- Updated ECS Streamlit module to use single CONFIG_SECRET environment variable
- Enhanced Secrets Manager module to support both simple strings and JSON configuration objects
- Updated IAM policies in ECS Streamlit module to support configuration secret access

### Security
- Improved security posture by moving sensitive configuration from environment variables to encrypted AWS Secrets Manager
- Centralized configuration management for better security and maintainability
- Reduced environment variable exposure in ECS task definitions

## [1.5.0] - 2025-05-19

### Fixed
- Fixed IAM policy error in Step Functions module: "MalformedPolicyDocument: Policy statement must contain resources"
- Removed S3 state management policy when s3_state_bucket_arn is null
- Added state lifecycle rules to terraform.tfvars for proper S3 bucket configuration
- Removed default values from variables.tf for s3_buckets to ensure values are defined in terraform.tfvars

### Changed
- Updated main.tf to properly handle Step Functions module integration with S3 state bucket
- Improved error handling in Step Functions module for S3 state management policy

## [1.4.0] - 2025-05-18

### Added
- Added S3 state management bucket configuration
- Added lifecycle rules for state bucket data management
- Added integration between Step Functions and S3 state bucket

### Changed
- Updated S3 module to support state management bucket
- Updated Step Functions module to access S3 state bucket
