# Changelog

All notable changes to the Health Check Lambda Function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-06-05

### Changed
- **Environment Variables**: Updated to use standardized environment variable names
  - Changed from `VERIFICATION_RESULTS_TABLE` to `DYNAMODB_VERIFICATION_TABLE`
  - Changed from `CONVERSATION_HISTORY_TABLE` to `DYNAMODB_CONVERSATION_TABLE`
  - Aligned with other Lambda functions in the codebase for consistency

### Added
- **Enhanced Logging**: Added comprehensive logging of all environment variables during initialization
  - Logs DynamoDB table names (verification and conversation)
  - Logs S3 bucket names (reference, checking, results)
  - Logs Bedrock model configuration
  - Improves debugging and monitoring capabilities

### Technical
- **Consistency**: Standardized environment variable naming across all Lambda functions
- **Monitoring**: Enhanced observability with detailed environment variable logging
- **Compatibility**: Maintains backward compatibility in functionality while updating configuration

## [1.0.0] - 2025-01-02

### Added
- **Initial Implementation**: Complete Lambda function for health checking system components
  - Health check endpoint for monitoring system status
  - DynamoDB table connectivity verification
  - S3 bucket accessibility checks
  - Bedrock model availability validation
  - Structured health status reporting with service-level details

### Features
- **Multi-Service Health Checks**:
  - DynamoDB: Verifies access to verification and conversation tables
  - S3: Checks accessibility of reference, checking, and results buckets
  - Bedrock: Validates model configuration and availability
  - Overall system status aggregation (healthy/degraded/unhealthy)

- **Comprehensive Status Reporting**:
  - Service-level health status with detailed error messages
  - Timestamp information for monitoring and alerting
  - Version information for deployment tracking
  - Structured JSON response format for easy parsing

- **Error Handling and Resilience**:
  - Graceful degradation when individual services are unavailable
  - Detailed error messages for troubleshooting
  - Proper HTTP status codes and CORS headers
  - Non-blocking health checks that don't fail the entire system

### Technical Implementation
- **Go 1.20**: Modern Go implementation with AWS SDK v2
- **AWS SDK v2**: Latest AWS SDK for Go with improved performance
- **Multi-Service Integration**:
  - DynamoDB client for table health checks
  - S3 client for bucket accessibility verification
  - Bedrock client for model availability checks
- **Docker Support**: Containerized deployment with optimized Docker image
- **Lambda Runtime**: AWS Lambda Go runtime with proper event handling

### Infrastructure
- **Environment Variables**:
  - `DYNAMODB_VERIFICATION_TABLE`: DynamoDB verification table name
  - `DYNAMODB_CONVERSATION_TABLE`: DynamoDB conversation table name
  - `REFERENCE_BUCKET`: S3 bucket for reference images
  - `CHECKING_BUCKET`: S3 bucket for checking images
  - `RESULTS_BUCKET`: S3 bucket for processed results
  - `BEDROCK_MODEL`: Bedrock model identifier
  - `LOG_LEVEL`: Configurable logging level

- **IAM Permissions**: Minimal required permissions for health checks
  - `dynamodb:DescribeTable`: For table existence verification
  - `s3:HeadBucket`: For bucket accessibility checks
  - `bedrock:GetFoundationModel`: For model availability verification

### API Specification
- **Endpoint**: Health check endpoint for system monitoring
- **Response Format**: JSON with overall status and service-level details
- **CORS Headers**: Proper CORS configuration for web application integration
- **Status Codes**: Appropriate HTTP status codes for different health states

### Monitoring and Observability
- **CloudWatch Integration**: Compatible with AWS CloudWatch for monitoring
- **Structured Logging**: JSON-formatted logs for easy parsing and analysis
- **Health Metrics**: Built-in health status reporting for each service component
- **Error Tracking**: Detailed error logging for troubleshooting

### Security
- **IAM Permissions**: Minimal required permissions following principle of least privilege
- **CORS Configuration**: Secure CORS headers for web application integration
- **Error Messages**: Informative error messages without exposing sensitive information

### Development Tools
- **Makefile**: Build and deployment automation
- **Docker Configuration**: Optimized Dockerfile for containerized deployment
- **Go Module**: Proper dependency management with go.mod

### Documentation
- **README**: Comprehensive documentation covering configuration and deployment
- **Code Documentation**: Well-documented Go code with clear function descriptions
- **Health Check Specification**: Clear definition of health check behavior and responses
