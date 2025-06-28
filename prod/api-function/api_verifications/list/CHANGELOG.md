# Changelog

All notable changes to the API Verifications List Lambda Function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-06-05

### Changed
- **Environment Variables**: Updated to use standardized environment variable names
  - Changed from `VERIFICATION_TABLE` to `DYNAMODB_VERIFICATION_TABLE`
  - Added support for `DYNAMODB_CONVERSATION_TABLE` (loaded but not currently used)
  - Both environment variables are now required and validated at startup
  - Added comprehensive logging of loaded environment variables for better debugging

### Added
- **Enhanced Logging**: Added detailed logging of environment variables during initialization
  - Logs both verification and conversation table names
  - Improves debugging and monitoring capabilities
  - Consistent with other functions in the codebase

### Technical
- **Consistency**: Aligned environment variable naming with other Lambda functions
- **Future-proofing**: Added conversation table support for potential future features
- **Validation**: Both DynamoDB table environment variables are now required and validated

## [1.0.0] - 2025-01-02

### Added
- **Initial Implementation**: Complete Lambda function for listing verification results
  - GET `/api/verifications` endpoint with comprehensive filtering and pagination
  - DynamoDB integration with efficient query patterns using Global Secondary Indexes
  - Support for filtering by verification status, vending machine ID, and date ranges
  - Pagination with configurable limits (1-100) and offset-based navigation
  - Sorting capabilities by verification date and overall accuracy
  - CORS support for web application integration
  - Structured logging with configurable log levels (DEBUG, INFO, WARN, ERROR)

### Features
- **Advanced Filtering System**:
  - `verificationStatus`: Filter by CORRECT or INCORRECT status
  - `vendingMachineId`: Filter by specific vending machine identifier
  - `fromDate` and `toDate`: Date range filtering with RFC3339 format support
  - Combined filtering support for complex queries

- **Pagination and Sorting**:
  - Configurable page size with `limit` parameter (default: 20, max: 100)
  - Offset-based pagination with `offset` parameter
  - Multiple sorting options: `verificationAt:desc/asc`, `overallAccuracy:desc/asc`
  - Pagination metadata in response including total count and next offset

- **Performance Optimization**:
  - Uses DynamoDB VerificationStatusIndex GSI for efficient status-based queries
  - Falls back to table scan for general queries when no status filter is provided
  - Optimized query patterns to minimize DynamoDB read capacity consumption
  - Efficient data marshaling/unmarshaling with AWS SDK v2

- **Error Handling and Validation**:
  - Comprehensive input validation for all query parameters
  - Structured error responses with appropriate HTTP status codes
  - Graceful handling of DynamoDB errors and timeouts
  - Detailed error messages for debugging and troubleshooting

### Technical Implementation
- **Go 1.20**: Modern Go implementation with latest AWS SDK v2
- **AWS SDK v2**: Latest AWS SDK for Go with improved performance and features
- **DynamoDB Integration**: 
  - Support for multiple GSI queries (VerificationStatusIndex, VerificationTypeIndex, etc.)
  - Efficient attribute value marshaling/unmarshaling
  - Proper handling of DynamoDB data types and optional fields
- **Docker Support**: Multi-stage Docker build for optimized container images
- **Lambda Runtime**: AWS Lambda Go runtime with proper event handling

### Infrastructure
- **Environment Variables**:
  - `DYNAMODB_VERIFICATION_TABLE`: DynamoDB table name for verification records
  - `DYNAMODB_CONVERSATION_TABLE`: DynamoDB table name for conversation records (loaded but not used)
  - `LOG_LEVEL`: Configurable logging level for debugging and monitoring

- **IAM Permissions**: Minimal required permissions for DynamoDB operations
  - `dynamodb:Query`: For GSI queries
  - `dynamodb:Scan`: For table scans when needed
  - `dynamodb:GetItem`: For individual record retrieval

- **DynamoDB Table Structure**: Compatible with existing verification results table
  - Primary key: `verificationId` (Hash) + `verificationAt` (Range)
  - GSI support: VerificationStatusIndex, VerificationTypeIndex, and others
  - Proper handling of optional fields and nested objects

### API Specification
- **Endpoint**: `GET /api/verifications`
- **Response Format**: JSON with results array and pagination metadata
- **Query Parameters**: Comprehensive set of filtering and pagination options
- **CORS Headers**: Proper CORS configuration for web application integration
- **Error Responses**: Structured error format with error codes and messages

### Development Tools
- **Deployment Script**: Comprehensive `deploy.sh` script with multiple deployment options
  - Full deployment pipeline (build, test, push, update)
  - Individual operations (build-only, push-only, update-only)
  - Local development commands (go-build, go-test, go-run)
  - Automatic ECR repository discovery and Lambda function detection

- **Docker Configuration**: Optimized Dockerfile with multi-stage build
  - Alpine-based final image for minimal size
  - Proper dependency caching for faster builds
  - Security best practices with non-root user

- **Go Module**: Proper dependency management with go.mod
  - AWS Lambda Go runtime
  - AWS SDK v2 for DynamoDB
  - Logrus for structured logging
  - All necessary indirect dependencies

### Documentation
- **Comprehensive README**: Detailed documentation covering all aspects
  - API endpoint documentation with examples
  - Environment variable configuration
  - DynamoDB table structure requirements
  - IAM permission specifications
  - Building and deployment instructions
  - Performance considerations and optimization tips
  - Troubleshooting guide with common issues

- **Code Documentation**: Well-documented Go code with clear function descriptions
  - Struct definitions for all data types
  - Clear separation of concerns with dedicated functions
  - Proper error handling and logging throughout

### Testing and Quality
- **Input Validation**: Comprehensive validation for all query parameters
- **Error Handling**: Graceful error handling with proper HTTP status codes
- **Logging**: Structured logging for monitoring and debugging
- **Performance**: Optimized query patterns for efficient DynamoDB usage

### Security
- **Input Sanitization**: Proper validation and sanitization of all inputs
- **IAM Permissions**: Minimal required permissions following principle of least privilege
- **CORS Configuration**: Secure CORS headers for web application integration
- **Error Messages**: Informative error messages without exposing sensitive information

### Monitoring and Observability
- **CloudWatch Integration**: Compatible with AWS CloudWatch for monitoring
- **Structured Logging**: JSON-formatted logs for easy parsing and analysis
- **Performance Metrics**: Built-in logging of query performance and result counts
- **Error Tracking**: Detailed error logging for troubleshooting and debugging

### Future Enhancements
- **Caching**: Consider implementing caching for frequently accessed data
- **Advanced Sorting**: Enhanced sorting capabilities for complex use cases
- **Batch Operations**: Support for batch queries and operations
- **Real-time Updates**: Integration with DynamoDB Streams for real-time updates
- **Analytics**: Enhanced analytics and reporting capabilities
