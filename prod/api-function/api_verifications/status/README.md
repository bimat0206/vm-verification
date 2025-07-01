# Verification Status API (Python)

This is a Python implementation of the verification status API Lambda function, refactored from Go to avoid AWS SDK v2 endpoint resolution issues.

## Overview

The API provides a GET endpoint to retrieve verification status from DynamoDB:
- `GET /api/verifications/status/{verificationId}`

## Why Python?

The original Go implementation faced persistent "ResolveEndpointV2" errors due to AWS SDK Go v2 module incompatibilities. Python's boto3 SDK provides more stable endpoint resolution in Lambda environments.

## Features

- Simple and reliable DynamoDB queries using boto3
- Automatic retry logic built into boto3
- Clear error handling and logging
- S3 content retrieval for completed verifications
- CORS support for web applications
- Lightweight Python implementation
- Proper handling of DynamoDB Decimal types for JSON serialization

## Deployment

Deploy using the provided script:

```bash
./deploy.sh
```

Available commands:
- `./deploy.sh deploy` - Deploy to AWS Lambda (default)
- `./deploy.sh test` - Test locally
- `./deploy.sh clean` - Clean up artifacts
- `./deploy.sh help` - Show help

## Environment Variables

Required:
- `DYNAMODB_VERIFICATION_TABLE` - DynamoDB table for verification records
- `DYNAMODB_CONVERSATION_TABLE` - DynamoDB table for conversation metadata

Optional:
- `STATE_BUCKET` - S3 bucket for state files
- `STEP_FUNCTIONS_STATE_MACHINE_ARN` - Step Functions state machine ARN
- `AWS_REGION` or `REGION` - AWS region (defaults to us-east-1)
- `LOG_LEVEL` - Logging level (DEBUG, INFO, WARNING, ERROR)

## Response Format

```json
{
  "verificationId": "string",
  "status": "RUNNING|COMPLETED|FAILED",
  "currentStatus": "string",
  "verificationStatus": "CORRECT|INCORRECT|PENDING",
  "s3References": {
    "turn1Processed": "s3://bucket/path",
    "turn2Processed": "s3://bucket/path"
  },
  "summary": {
    "message": "string",
    "verificationAt": "string",
    "verificationStatus": "string",
    "overallAccuracy": 0.0,
    "correctPositions": 0,
    "discrepantPositions": 0
  },
  "llmResponse": "string",
  "verificationSummary": {}
}
```

## Error Responses

```json
{
  "error": "Error type",
  "message": "Detailed error message",
  "code": "HTTP_XXX"
}
```

## Testing

Test locally:
```bash
python3 -m pytest test_lambda_function.py
```

Or use the deployment script:
```bash
./deploy.sh test
```

## Migration from Go

The Go implementation is preserved in the `old/` directory for reference. This Python implementation maintains the same API interface while providing better reliability in AWS Lambda environments.