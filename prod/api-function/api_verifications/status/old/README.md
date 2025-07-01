# API Verifications Status Lambda Function

This is a Go-based AWS Lambda function that provides a REST API for checking the status of verification processes. It serves as the backend for the `/api/verifications/status/{verificationId}` endpoint, enabling users to poll verification execution status and retrieve results when processing is complete.

## Features

- **Status Polling**: Check verification execution status (RUNNING, COMPLETED, FAILED)
- **Result Retrieval**: Automatically fetch LLM responses and verification summaries when complete
- **S3 Integration**: Retrieve processed markdown content from S3 when verification is finished
- **Error Handling**: Comprehensive error handling for missing verifications and S3 access issues
- **CORS Support**: Full CORS support for web applications
- **Structured Logging**: JSON-formatted logs with contextual information

## API Endpoint

```
GET /api/verifications/status/{verificationId}
```

### Path Parameters
- `verificationId` (required): The verification ID to check status for (e.g., "verif-20250607092759-cc32")

### Response Format

#### Successful Response (200)
```json
{
  "verificationId": "verif-20250607092759-cc32",
  "status": "COMPLETED",
  "currentStatus": "COMPLETED",
  "verificationStatus": "CORRECT",
  "s3References": {
    "turn1Processed": "s3://bucket/path/turn1-processed-response.md",
    "turn2Processed": "s3://bucket/path/turn2-processed-response.md"
  },
  "summary": {
    "message": "Verification completed successfully - No discrepancies found",
    "verificationAt": "2025-06-05T08:52:05Z",
    "verificationStatus": "CORRECT",
    "overallAccuracy": 0.833,
    "correctPositions": 35,
    "discrepantPositions": 7
  },
  "llmResponse": "# Verification Analysis\n\n## Summary\nThe verification process has been completed...",
  "verificationSummary": {
    "overall_confidence": "100%",
    "total_positions_checked": 42,
    "verification_outcome": "CORRECT"
  }
}
```

#### Status Values
- **RUNNING**: Verification is currently in progress
- **COMPLETED**: Verification has finished processing
- **FAILED**: Verification failed during processing

#### Error Response (4xx/5xx)
```json
{
  "error": "Verification not found",
  "message": "No verification found for verificationId: invalid-id",
  "code": "HTTP_404"
}
```

## Architecture

### AWS Services Integration
- **AWS Lambda**: Serverless compute platform for the API handler
- **Amazon DynamoDB**: NoSQL database for verification status and metadata storage
- **Amazon S3**: Object storage for processed LLM responses and verification content
- **AWS API Gateway**: HTTP API gateway for routing and request handling

### Data Flow
1. **Request Processing**: Extract verificationId from path parameters
2. **DynamoDB Query**: Query verification table using verificationId
3. **Status Determination**: Map currentStatus and verificationStatus to overall status
4. **S3 Retrieval**: Fetch LLM response content from S3 if verification is completed
5. **Response Formation**: Return structured JSON response with status and results

## Environment Variables

The function requires the following environment variables:

### Required
- `DYNAMODB_VERIFICATION_TABLE` - Name of the DynamoDB table containing verification records
- `DYNAMODB_CONVERSATION_TABLE` - Name of the DynamoDB table containing conversation metadata

### Optional
- `STATE_BUCKET` - Name of the S3 bucket containing state files (e.g., "kootoro-dev-s3-state-f6d3xl")
- `STEP_FUNCTIONS_STATE_MACHINE_ARN` - ARN of the Step Functions state machine (e.g., "arn:aws:states:us-east-1:879654127886:stateMachine:kootoro-dev-sfn-verification-workflow-f6d3xl")
- `LOG_LEVEL` - Logging level (debug, info, warn, error) - defaults to info

### Example Configuration
```bash
export DYNAMODB_VERIFICATION_TABLE="kootoro-dev-verification-table"
export DYNAMODB_CONVERSATION_TABLE="kootoro-dev-conversation-table"
export STATE_BUCKET="kootoro-dev-s3-state-f6d3xl"
export STEP_FUNCTIONS_STATE_MACHINE_ARN="arn:aws:states:us-east-1:879654127886:stateMachine:kootoro-dev-sfn-verification-workflow-f6d3xl"
export LOG_LEVEL="info"
```

## Building and Deployment

### Quick Start
```bash
# Full deployment (recommended)
./deploy.sh

# Or using make
make deploy
```

### Manual Build
```bash
# Build Go binary
go build -o api-verifications-status *.go

# Build Docker image
docker build -t api-verifications-status .

# Deploy to AWS Lambda
aws lambda update-function-code \
  --function-name api-verifications-status \
  --image-uri your-ecr-repo:latest
```

### Development Commands
```bash
# Build only
./deploy.sh build

# Clean build artifacts
./deploy.sh clean

# Show help
./deploy.sh help
```

## Testing

### Local Testing
```bash
# Build the function
go build -o api-verifications-status *.go

# Test with sample event
echo '{"httpMethod":"GET","pathParameters":{"verificationId":"test-id"}}' | ./api-verifications-status
```

### API Testing
```bash
# Test deployed function
curl -X GET "https://your-api-gateway-url/api/verifications/status/verif-20250607092759-cc32" \
  -H "X-Api-Key: your-api-key" \
  -H "Content-Type: application/json"
```

## Dependencies

- **AWS SDK for Go v2** - Latest DynamoDB and S3 clients
- **AWS Lambda Go runtime** - Native Lambda execution environment
- **Logrus** - Structured JSON logging
- **Go 1.20+** - Modern Go runtime with generics support

## Error Handling

The function handles various error scenarios:

- **Missing verificationId**: Returns 400 Bad Request
- **Verification not found**: Returns 404 Not Found
- **DynamoDB errors**: Returns 500 Internal Server Error with details
- **S3 access errors**: Logs warning but continues (LLM response will be empty)
- **Invalid S3 paths**: Logs warning and skips content retrieval

## Monitoring and Logging

All operations are logged with structured JSON format including:
- Request details (method, path, parameters)
- Processing steps and timing
- Error details with context
- Response status and content summary

## Security

- Uses IAM roles for AWS service access
- Validates input parameters
- Sanitizes error messages in responses
- Supports CORS for web application integration
