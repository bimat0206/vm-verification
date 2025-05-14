# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ExecuteTurn1 is a Go-based AWS Lambda function that integrates with Amazon Bedrock's Claude 3.7 Sonnet model to perform intelligent analysis of vending machine images. It handles the first turn of AI conversation in the verification process, supporting two scenarios:

1. **LAYOUT_VS_CHECKING**: Compares a reference layout image with a checking image
2. **PREVIOUS_VS_CURRENT**: Uses a previous checking image as reference for comparison

## Build & Deployment Commands

```bash
# Build the Docker image locally
docker build -t execute-turn1 .

# Tag for ECR (replace with actual account ID and region)
docker tag execute-turn1:latest <account-id>.dkr.ecr.<region>.amazonaws.com/execute-turn1:latest

# Push to ECR
docker push <account-id>.dkr.ecr.<region>.amazonaws.com/execute-turn1:latest
```

## Development Commands

```bash
# Run tests (from project root)
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Architecture

The codebase follows a clean separation of concerns with the following structure:

1. **Lambda Handler** (`cmd/main.go`):
   - Manages environment variable configuration
   - Handles input validation
   - Initializes Bedrock client
   - Implements retry logic for robust error handling

2. **Core Modules** (`internal/`):
   - `types.go`: Data structure definitions for all components
   - `bedrock.go`: Amazon Bedrock API integration
   - `validation.go`: Input validation logic
   - `response.go`: Response processing
   - `retry.go`: Intelligent retry mechanisms
   - `errors.go`: Error handling patterns

3. **Request Flow**:
   - Lambda receives input with verification context and prompt
   - Input validation ensures required fields and formats
   - Bedrock client prepares and sends request to Claude 3.7
   - Response is processed and structured for downstream components
   - Error handling with retry logic for transient issues

## Configuration

The function uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `BEDROCK_MODEL` | Claude 3.7 Sonnet model ID | required |
| `BEDROCK_REGION` | AWS region for Bedrock | `us-east-1` |
| `ANTHROPIC_VERSION` | Anthropic API version | `bedrock-2023-05-31` |
| `MAX_TOKENS` | Maximum output tokens | `24000` |
| `BUDGET_TOKENS` | Thinking tokens budget | `16000` |
| `THINKING_TYPE` | Enable step-by-step reasoning | `enable` |
| `RETRY_MAX_ATTEMPTS` | Maximum retry attempts | `3` |
| `RETRY_BASE_DELAY` | Base delay for retries (ms) | `2000` |
| `DYNAMODB_CONVERSATION_TABLE` | Conversation history table | `` |

## Error Handling

The function implements comprehensive error handling with specific error types:

- `ValidationException`: Input validation errors
- `BedrockException`: Bedrock API errors (throttling, model issues)
- `S3Exception`: S3 access or format errors
- `TimeoutException`: Operation timeouts
- `InternalException`: Internal processing errors

Retry logic uses exponential backoff with jitter to handle transient failures.