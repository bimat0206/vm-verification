# ExecuteTurn1 Lambda Function

This Lambda function executes the first conversation turn with Amazon Bedrock, providing reference image analysis for the vending machine verification workflow.

## Overview

ExecuteTurn1 is a critical component in the verification workflow that:

1. Takes in the workflow state with prepared prompts
2. Retrieves image data (either inline or from S3)
3. Calls the Bedrock Converse API with the image and prompt
4. Processes the response to extract relevant content
5. Updates the workflow state with the results

## Features

- Supports both legacy text prompts and new Bedrock messages format
- Implements hybrid Base64 storage for efficient handling of large images
- Extracts and processes "thinking" content when enabled
- Comprehensive error handling with retryable errors
- Structured logging with correlation IDs

## Directory Structure

```
ExecuteTurn1/
├── cmd/
│   └── main.go                 # Lambda handler entry point
├── internal/
│   ├── handler/
│   │   ├── execute_turn1.go    # Main business logic
│   │   └── response_processor.go # Response processing logic
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── dependencies/
│   │   └── clients.go          # AWS client initialization
│   └── models/
│       └── request.go          # Request/response models
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Configuration

The function uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| BEDROCK_MODEL_ID | Bedrock model identifier | anthropic.claude-3-7-sonnet-20250219-v1:0 |
| AWS_REGION | AWS region for services | us-east-1 |
| ANTHROPIC_VERSION | Anthropic API version | bedrock-2023-05-31 |
| MAX_TOKENS | Maximum tokens for response | 4000 |
| TEMPERATURE | Temperature for model sampling | 0.7 |
| THINKING_TYPE | Type of thinking to extract | thoroughness |
| THINKING_BUDGET_TOKENS | Token budget for thinking | 50000 |
| ENABLE_HYBRID_STORAGE | Enable S3 storage for large Base64 | true |
| TEMP_BASE64_BUCKET | S3 bucket for Base64 storage | temp-base64-bucket |
| BASE64_SIZE_THRESHOLD | Size threshold for S3 storage | 2097152 (2MB) |
| BASE64_RETRIEVAL_TIMEOUT | Timeout for S3 retrieval | 30000ms |
| BEDROCK_TIMEOUT | Timeout for Bedrock API calls | 300000ms (5min) |
| FUNCTION_TIMEOUT | Overall function timeout | 300000ms (5min) |

## Building and Deployment

### Local Build

```bash
go build -o main ./cmd/main.go
```

### Docker Build

```bash
docker build -t execute-turn1 .
```

### AWS Lambda Deployment

The function is designed to be deployed as a container image to AWS Lambda.

1. Build the Docker image
2. Push to Amazon ECR
3. Create or update Lambda function to use the ECR image
4. Configure environment variables and IAM permissions

Required IAM permissions:
- `bedrock:InvokeModel` for the specified model
- `s3:GetObject` for the Base64 storage bucket (if hybrid storage is enabled)

## Input and Output

### Input

```json
{
  "workflowState": {
    "verificationId": "12345",
    "status": "IN_PROGRESS",
    "stage": "REFERENCE_ANALYSIS",
    "correlationId": "corr-123",
    "timestamp": "2023-04-01T12:00:00Z",
    "currentPrompt": {
      "promptId": "prompt-123",
      "messages": [
        {
          "role": "user",
          "content": [
            {
              "type": "text",
              "text": "Analyze this vending machine image..."
            }
          ]
        }
      ]
    },
    "imageData": {
      "imageId": "img-123",
      "imageBase64": "base64data...",
      "contentType": "image/jpeg"
    }
  }
}
```

### Output

```json
{
  "workflowState": {
    "verificationId": "12345",
    "status": "IN_PROGRESS",
    "stage": "TURN1_COMPLETE",
    "correlationId": "corr-123",
    "timestamp": "2023-04-01T12:00:30Z",
    "currentPrompt": { ... },
    "imageData": { ... },
    "turnHistory": [
      {
        "turnId": 1,
        "timestamp": "2023-04-01T12:00:30Z",
        "response": {
          "id": "resp-123",
          "content": [
            {
              "type": "text",
              "text": "Based on the vending machine image..."
            }
          ],
          "role": "assistant",
          "usage": {
            "inputTokens": 1000,
            "outputTokens": 2000,
            "totalTokens": 3000
          }
        },
        "latencyMs": 15000,
        "tokenUsage": {
          "inputTokens": 1000,
          "outputTokens": 2000,
          "totalTokens": 3000
        },
        "stage": "REFERENCE_ANALYSIS",
        "thinking": "I'm analyzing this image step by step..."
      }
    ]
  }
}
```

## Error Handling

The function returns structured errors:

```json
{
  "workflowState": { ... },
  "error": {
    "code": "BEDROCK_API_ERROR",
    "message": "Failed to call Bedrock API",
    "retryable": true,
    "context": {
      "promptId": "prompt-123",
      "error": "Quota exceeded"
    }
  }
}
```

## Monitoring and Logging

All function activity is logged to CloudWatch Logs with structured information.
Key metrics to monitor:

- Bedrock API latency
- Error rates
- Lambda timeout occurrences
- Memory utilization

## Development

### Prerequisites

- Go 1.21+
- Docker
- AWS CLI configured