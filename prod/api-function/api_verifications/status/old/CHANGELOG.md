# Changelog

All notable changes to the API Verifications Status Lambda Function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-01-30

### Fixed
- Fixed DynamoDB "ResolveEndpointV2" error caused by AWS SDK v2 module incompatibility (ref: [GitHub issue #2370](https://github.com/aws/aws-sdk-go-v2/issues/2370))
- Resolved endpoint resolution issues by using BaseEndpoint option to explicitly set DynamoDB endpoint
- Fixed timeout issues with proper context-based timeouts

### Changed
- Simplified AWS client initialization by removing custom endpoint resolvers
- Used BaseEndpoint option for DynamoDB client to avoid endpoint resolution issues
- Improved region configuration with multiple fallback options (AWS_REGION → REGION → AWS_DEFAULT_REGION → us-east-1)
- Enhanced error handling with retry logic for transient errors
- Better separation of concerns with dedicated initialization functions
- Improved logging with more structured fields for debugging

### Added
- Custom HTTP client configuration with 30-second timeout
- Exponential backoff retry logic for DynamoDB operations
- Helper function to identify retryable errors
- Context-based timeouts for all AWS operations
- Comprehensive error classification and better error messages
- Documentation for refactoring changes (REFACTORING_NOTES.md)

### Technical Improvements
- Simplified AWS configuration to use minimal options
- Used BaseEndpoint instead of complex custom resolvers
- Improved resource management with defer statements
- Better initialization flow with error propagation

## [1.0.0] - 2025-01-XX

### Added
- Initial implementation of verification status API endpoint
- GET `/api/verifications/status/{verificationId}` endpoint for polling verification status
- Support for status checking: RUNNING, COMPLETED, FAILED
- Automatic retrieval of LLM responses from S3 when verification is completed
- Integration with DynamoDB for verification record lookup
- Integration with S3 for processed content retrieval
- Comprehensive error handling for missing verifications and S3 access issues
- CORS support for web applications
- Structured JSON logging with contextual information
- Docker-based deployment with ECR integration
- Automated deployment script with dependency checking
- Test payload files for development and testing
- Comprehensive documentation and README

### Features
- **Status Polling**: Real-time verification status checking
- **Result Retrieval**: Automatic fetching of verification results when complete
- **S3 Integration**: Seamless retrieval of processed markdown content
- **Error Handling**: Robust error handling with appropriate HTTP status codes
- **Logging**: Structured logging for monitoring and debugging
- **Security**: IAM-based access control and input validation

### Technical Details
- Built with Go 1.20+ for optimal performance
- Uses AWS SDK for Go v2 for modern AWS service integration
- Implements AWS Lambda runtime for serverless execution
- Supports container-based deployment via ECR
- Includes comprehensive test scenarios and documentation

### Environment Variables
- `DYNAMODB_VERIFICATION_TABLE`: DynamoDB table for verification records (required)
- `DYNAMODB_CONVERSATION_TABLE`: DynamoDB table for conversation metadata (required)
- `STATE_BUCKET`: S3 bucket for state files (optional)
- `STEP_FUNCTIONS_STATE_MACHINE_ARN`: Step Functions state machine ARN (optional)
- `LOG_LEVEL`: Configurable logging level (debug, info, warn, error)

### API Response Format
```json
{
  "verificationId": "string",
  "status": "RUNNING|COMPLETED|FAILED",
  "currentStatus": "string",
  "verificationStatus": "CORRECT|INCORRECT|PENDING",
  "s3References": {
    "turn1Processed": "string",
    "turn2Processed": "string"
  },
  "summary": {
    "message": "string",
    "verificationAt": "string",
    "verificationStatus": "string",
    "overallAccuracy": "number",
    "correctPositions": "number",
    "discrepantPositions": "number"
  },
  "llmResponse": "string",
  "verificationSummary": "object"
}
```

### Deployment
- Automated deployment via `deploy.sh` script
- Docker containerization for consistent deployment
- ECR integration for image storage
- AWS Lambda function updates with zero downtime

### Testing
- Comprehensive test event collection
- Local testing support
- API integration testing examples
- Error scenario coverage
