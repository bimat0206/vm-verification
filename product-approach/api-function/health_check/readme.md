# Health Check Lambda Function

This is a Go-based AWS Lambda function that performs health checks on various AWS resources used in the vending machine verification system.

## Components Checked

- **DynamoDB Tables**
  - Verification results table
  - Conversation history table
  
- **S3 Buckets**
  - Reference bucket
  - Checking bucket
  - Results bucket
  
- **Amazon Bedrock**
  - Model availability

## Environment Variables

The function requires the following environment variables:

| Variable | Description |
|----------|-------------|
| `VERIFICATION_RESULTS_TABLE` | Name of the DynamoDB table for verification results |
| `CONVERSATION_HISTORY_TABLE` | Name of the DynamoDB table for conversation history |
| `REFERENCE_BUCKET` | Name of the S3 bucket for reference images |
| `CHECKING_BUCKET` | Name of the S3 bucket for checking images |
| `RESULTS_BUCKET` | Name of the S3 bucket for results |
| `BEDROCK_MODEL` | ID of the Amazon Bedrock model |
| `LOG_LEVEL` | Logging level (e.g., INFO, DEBUG, ERROR) |

## API Response Format

The health check returns a JSON response with the following structure:

```json
{
  "status": "healthy",  // Overall status: "healthy", "degraded", or "unhealthy"
  "version": "1.0.0",   // Version of the health check
  "timestamp": "2025-05-05T12:34:56Z",  // ISO 8601 timestamp
  "services": {
    "dynamodb": {
      "status": "healthy",  // Status of DynamoDB
      "message": "DynamoDB tables accessible",
      "details": {
        "verification_table": "table-name",
        "conversation_table": "table-name"
      }
    },
    "s3": {
      "status": "healthy",  // Status of S3 buckets
      "message": "S3 buckets accessible",
      "details": {
        "reference_bucket": "bucket-name",
        "checking_bucket": "bucket-name",
        "results_bucket": "bucket-name"
      }
    },
    "bedrock": {
      "status": "healthy",  // Status of Bedrock
      "message": "Bedrock model available",
      "details": {
        "model_id": "model-id"
      }
    }
  }
}
```

## Building and Deployment

### Prerequisites

- Go 1.20 or higher
- Docker
- AWS CLI configured with appropriate permissions
- Access to AWS ECR repository

### Local Build

```bash
# Build the binary
make build

# Run tests
make test
```

### Docker Build and Push

```bash
# Update the DOCKER_REPO in the Makefile to point to your ECR repository

# Build Docker image
make docker-build

# Push to ECR
make docker-push
```

### Deploy to AWS Lambda

```bash
# Deploy Lambda function (updates the function code)
make deploy
```

## Integration with API Gateway

This Lambda function is configured to be triggered by API Gateway at the `/api/v1/health` endpoint. It returns an HTTP 200 response with the health check results in the response body.

## Local Development

To run the function locally (for testing):

```bash
# Set required environment variables
export VERIFICATION_RESULTS_TABLE=your-table-name
export CONVERSATION_HISTORY_TABLE=your-table-name
export REFERENCE_BUCKET=your-bucket-name
export CHECKING_BUCKET=your-bucket-name
export RESULTS_BUCKET=your-bucket-name
export BEDROCK_MODEL=your-model-id
export LOG_LEVEL=INFO

# Run locally
make run
```