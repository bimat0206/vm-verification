# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ExecuteTurn1 is a Lambda function for a vending machine verification system. It handles the first conversation turn with Amazon Bedrock (Claude 3.7 Sonnet), analyzing the reference image for verification.

Key capabilities:
- Accepts workflow state and prompts per shared schema
- Retrieves image data via hybrid Base64 logic (inline or S3)
- Invokes the Bedrock Converse API with the Claude model
- Processes structured responses, including "thinking" content
- Updates workflow state for downstream step function processing

## Directory Structure

```
ExecuteTurn1/
├── cmd/
│   └── main.go              # Lambda handler entry point
├── internal/
│   ├── handler/
│   │   ├── execute_turn1.go  # Main business logic
│   │   └── response_processor.go   # Response processing
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── dependencies/
│   │   └── clients.go       # AWS client initialization
│   └── models/
│       └── request.go       # Lambda input/output models
├── Dockerfile               # Container build definition
├── go.mod                   # Go module definition
└── retry-docker-build.sh    # Build/deployment script
```

## Development Commands

### Building the Code

Build locally:
```bash
go build -o main ./cmd/main.go
```

### Docker Build and Deployment

Build Docker image:
```bash
docker build -t execute-turn1 .
```

For complete build, push to ECR, and Lambda deployment:
```bash
./retry-docker-build.sh
```

With custom options:
```bash
./retry-docker-build.sh --repo=<ECR_REPO_URI> --function=<LAMBDA_FUNCTION_NAME> --region=<AWS_REGION>
```

### Environment Configuration

The Lambda function is configured using environment variables as described in the configuration section of the README. All parameters have sensible defaults but can be overridden.

Key environment variables:
- `BEDROCK_MODEL_ID`: Claude model identifier
- `AWS_REGION`: AWS region for services
- `ANTHROPIC_VERSION`: Anthropic API version
- `MAX_TOKENS`: Max tokens for Bedrock response
- `TEMPERATURE`: Model sampling temperature
- `THINKING_TYPE`: Type of thinking to extract
- `ENABLE_HYBRID_STORAGE`: Enable S3 storage for large Base64

## Code Structure and Flow

1. The Lambda handler is the entry point (`cmd/main.go`), which:
   - Initializes the logger, config, and clients
   - Validates the input workflow state
   - Calls the handler to process the request

2. The handler (`internal/handler/execute_turn1.go`):
   - Validates the input state, images, and prompt
   - Retrieves Base64 image data (using hybrid storage if needed)
   - Builds the Bedrock Converse input with the prompt and image
   - Calls Bedrock and processes the response
   - Updates the workflow state with the results
   - Returns the updated state or error

3. The response processor (`internal/handler/response_processor.go`):
   - Parses Bedrock API responses
   - Extracts "thinking" content if available
   - Updates the conversation state/history

## IAM Permissions

The Lambda function requires:
- `bedrock:InvokeModel` for the specified Claude model
- `s3:GetObject` for the Base64 storage bucket (if hybrid storage is enabled)

## Error Handling

The function returns structured errors within the response:
- Validation errors for invalid input
- Bedrock API errors (with retryable flag)
- S3/AWS service errors
- Timeout errors

## Dependencies

Key dependencies in go.mod:
- AWS Lambda Go Runtime
- AWS SDK v2 (for S3, Bedrock)
- Shared schema/error/logger packages