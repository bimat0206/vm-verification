# ExecuteTurn1 Lambda Function

This Lambda function is responsible for executing the first turn of a conversation with Amazon Bedrock's Claude model for vending machine verification. It processes the input prompt, sends it to the Bedrock Converse API, and returns the model's response.

## Overview

The ExecuteTurn1 function is part of a larger workflow for vending machine verification. It takes a prepared prompt (including reference image) and sends it to Amazon Bedrock's Claude model using the Converse API. The response is then processed and returned for further analysis.

## Architecture

The function uses the following shared packages:
- `shared/bedrock`: For interacting with Amazon Bedrock's Converse API
- `shared/errors`: For standardized error handling
- `shared/schema`: For common data structures and validation

## Building and Deployment

### Prerequisites
- Docker installed
- AWS CLI configured with appropriate permissions
- Access to the ECR repository

### Building the Docker Image
The function is packaged as a Docker container and deployed to AWS Lambda. To build and deploy:

1. Use the provided build script:
```bash
./retry-docker-build.sh
```

This script:
- Creates a temporary build context with the shared packages
- Sets up the correct directory structure for Go modules
- Creates a go.work file for local module resolution
- Builds the Docker image
- Pushes the image to ECR
- Updates the Lambda function

### Manual Build
If you need to build manually:

1. Create a temporary build directory with the shared packages:
```bash
mkdir -p temp_build/workflow-function/shared
cp -r ../shared/* temp_build/workflow-function/shared/
cp -r ./* temp_build/
```

2. Create a go.work file in the temporary directory:
```bash
cat > temp_build/go.work << EOF
go 1.24

use (
  .
  ./workflow-function/shared/bedrock
  ./workflow-function/shared/errors
  ./workflow-function/shared/schema
)
EOF
```

3. Build the Docker image:
```bash
docker build -t your-ecr-repo:latest temp_build
```

4. Push to ECR:
```bash
docker push your-ecr-repo:latest
```

### Troubleshooting
If you encounter build errors:

1. Verify the shared packages are correctly copied to the temporary build directory
2. Check that the directory structure matches the replace directives in go.mod
3. Ensure the go.work file is correctly configured
4. Try building locally first to identify any Go module issues

## Input

The function expects an input with the following structure:
```json
{
  "verificationContext": {
    "verificationId": "string",
    "verificationAt": "string",
    "status": "string",
    "verificationType": "string",
    "vendingMachineId": "string",
    "layoutId": 123,
    "layoutPrefix": "string",
    "referenceImageUrl": "string",
    "checkingImageUrl": "string",
    "turnConfig": {
      "maxTurns": 2,
      "referenceImageTurn": 1,
      "checkingImageTurn": 2
    },
    "turnTimestamps": {},
    "requestMetadata": {},
    "notificationEnabled": true,
    "resourceValidation": {}
  },
  "currentPrompt": {
    "currentPrompt": {
      "messages": [],
      "turnNumber": 1,
      "promptId": "string",
      "createdAt": "2025-05-15T00:00:00Z",
      "promptVersion": "string",
      "imageIncluded": "reference"
    }
  },
  "bedrockConfig": {
    "anthropic_version": "string",
    "max_tokens": 24000,
    "temperature": 0.7,
    "top_p": 0.9,
    "thinking": {
      "type": "enable",
      "budget_tokens": 16000
    }
  },
  "systemPrompt": {
    "systemPrompt": {
      "content": "string",
      "promptId": "string",
      "createdAt": "string",
      "promptVersion": "string"
    }
  }
}
```

## Output

The function returns an output with the following structure:
```json
{
  "verificationContext": {
    "verificationId": "string",
    "verificationAt": "string",
    "status": "TURN1_COMPLETED",
    "verificationType": "string",
    "vendingMachineId": "string",
    "layoutId": 123,
    "layoutPrefix": "string",
    "referenceImageUrl": "string",
    "checkingImageUrl": "string",
    "turnConfig": {
      "maxTurns": 2,
      "referenceImageTurn": 1,
      "checkingImageTurn": 2
    },
    "turnTimestamps": {
      "turn1Completed": "2025-05-15T00:00:00Z"
    },
    "requestMetadata": {},
    "notificationEnabled": true,
    "resourceValidation": {}
  },
  "turn1Response": {
    "turnId": 1,
    "timestamp": "2025-05-15T00:00:00Z",
    "prompt": "string",
    "response": {
      "content": "string",
      "stop_reason": "string"
    },
    "latencyMs": 1234,
    "tokenUsage": {
      "inputTokens": 123,
      "outputTokens": 456,
      "totalTokens": 579
    },
    "analysisStage": "REFERENCE_ANALYSIS",
    "bedrockMetadata": {
      "modelId": "string",
      "requestId": "string",
      "invokeLatencyMs": 1234,
      "apiType": "Converse"
    },
    "apiType": "Converse"
  },
  "conversationState": {
    "currentTurn": 1,
    "maxTurns": 2,
    "history": [
      {
        "turnId": 1,
        "timestamp": "2025-05-15T00:00:00Z",
        "prompt": "string",
        "response": "string",
        "latencyMs": 1234,
        "tokenUsage": {
          "inputTokens": 123,
          "outputTokens": 456,
          "totalTokens": 579
        },
        "analysisStage": "REFERENCE_ANALYSIS"
      }
    ]
  }
}
```

## Environment Variables

The function uses the following environment variables:
- `BEDROCK_MODEL`: The Bedrock model ID to use (required)
- `BEDROCK_REGION`: The AWS region for Bedrock (default: "us-east-1")
- `ANTHROPIC_VERSION`: The Anthropic version to use (default: "bedrock-2023-05-31")
- `MAX_TOKENS`: Maximum tokens for the response (default: 24000)
- `BUDGET_TOKENS`: Budget tokens for thinking (default: 16000)
- `THINKING_TYPE`: Type of thinking to use (default: "enable")
- `DYNAMODB_CONVERSATION_TABLE`: DynamoDB table for conversation history
- `RETRY_MAX_ATTEMPTS`: Maximum retry attempts (default: 3)
- `RETRY_BASE_DELAY`: Base delay for retries in milliseconds (default: 2000)

## Error Handling

The function uses the shared errors package for standardized error handling. Errors are categorized by type and severity, and some errors are retryable. The function includes a retry mechanism with exponential backoff for retryable errors.

## Dependencies

- AWS Lambda Go Runtime
- AWS SDK for Go v2
- Shared packages:
  - `workflow-function/shared/bedrock`
  - `workflow-function/shared/errors`
  - `workflow-function/shared/schema`

## Recent Changes

See [CHANGELOG.md](./CHANGELOG.md) for recent changes.
