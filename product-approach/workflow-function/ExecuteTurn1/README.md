# ExecuteTurn1 Function

The ExecuteTurn1 function is a critical component of the Kootoro GenAI Vending Machine Verification Solution. It handles the first turn of the AI conversation with Amazon Bedrock to analyze reference images and establish baseline understanding for vending machine verification.

## Overview

This AWS Lambda function integrates with Amazon Bedrock's Claude 3.7 Sonnet model to perform intelligent analysis of vending machine images. It supports two verification scenarios:

1. **LAYOUT_VS_CHECKING**: Compares a reference layout image with a checking image
2. **PREVIOUS_VS_CURRENT**: Uses a previous checking image as reference for comparison

## Features

- 🤖 **Bedrock Integration**: Direct integration with Claude 3.7 Sonnet for multimodal AI analysis
- 🔄 **Intelligent Retry Logic**: Exponential backoff with jitter and circuit breaker patterns
- 🖼️ **Multi-format Support**: Handles JPEG and PNG images
- 📊 **Comprehensive Logging**: Structured JSON logging for monitoring and debugging
- ⚡ **Optimized Performance**: Built for ARM64 Lambda runtime with efficient memory usage
- 🔒 **Secure**: IAM role-based access, input validation, and secure error handling
- 📈 **Metrics Tracking**: Built-in performance and token usage monitoring

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │ -> │  ExecuteTurn1   │ -> │ Amazon Bedrock  │
│   (Trigger)     │    │   Lambda        │    │ Claude 3.7      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              |
                              v
                       ┌─────────────────┐    ┌─────────────────┐
                       │  Conversation   │    │     S3 Bucket   │
                       │   History       │    │   (Images)      │
                       │  (DynamoDB)     │    │                 │
                       └─────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites

- AWS CLI configured
- Docker (for building and deployment)
- Go 1.21+ (for local development)

### Building

```bash
# Build the Docker image
docker build -t execute-turn1 .

# Tag for ECR
docker tag execute-turn1:latest <account-id>.dkr.ecr.<region>.amazonaws.com/execute-turn1:latest

# Push to ECR
docker push <account-id>.dkr.ecr.<region>.amazonaws.com/execute-turn1:latest
```

### Deployment

Deploy using your preferred infrastructure as code tool (e.g., CloudFormation, Terraform, CDK):

```yaml
# Example CloudFormation snippet
ExecuteTurn1Function:
  Type: AWS::Lambda::Function
  Properties:
    PackageType: Image
    Code:
      ImageUri: <account-id>.dkr.ecr.<region>.amazonaws.com/execute-turn1:latest
    MemorySize: 1024
    Timeout: 120
    Architectures:
      - arm64
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `BEDROCK_MODEL` | Claude 3.7 Sonnet model identifier | - | Yes |
| `BEDROCK_REGION` | AWS region for Bedrock | us-east-1 | No |
| `ANTHROPIC_VERSION` | Anthropic API version | bedrock-2023-05-31 | No |
| `MAX_TOKENS` | Maximum output tokens | 24000 | No |
| `BUDGET_TOKENS` | Thinking tokens budget | 16000 | No |
| `THINKING_TYPE` | Enable step-by-step reasoning | enable | No |
| `RETRY_MAX_ATTEMPTS` | Maximum retry attempts | 3 | No |
| `RETRY_BASE_DELAY` | Base delay for retries (ms) | 2000 | No |
| `DYNAMODB_CONVERSATION_TABLE` | Conversation history table | - | No |

### IAM Permissions

The Lambda function requires the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "arn:aws:bedrock:*:*:foundation-model/anthropic.claude-3-7-sonnet-*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::kootoro-reference-bucket/*",
        "arn:aws:s3:::kootoro-checking-bucket/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem"
      ],
      "Resource": "arn:aws:dynamodb:*:*:table/ConversationHistory"
    }
  ]
}
```

## Usage

### Input Format

```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "TURN1_PROMPT_READY",
    "verificationType": "LAYOUT_VS_CHECKING",
    "referenceImageUrl": "s3://bucket/reference.jpg",
    "checkingImageUrl": "s3://bucket/checking.jpg",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e"
  },
  "currentPrompt": {
    "messages": [...],
    "turnNumber": 1,
    "promptId": "prompt-id",
    "promptVersion": "1.0"
  },
  "bedrockConfig": {
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 24000,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 16000
    }
  }
}
```

### Output Format

```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "status": "TURN1_COMPLETED",
    "verificationType": "LAYOUT_VS_CHECKING"
  },
  "turn1Response": {
    "turnId": 1,
    "timestamp": "2025-04-21T15:30:30Z",
    "prompt": "Analyze the reference image...",
    "response": {
      "content": "I've analyzed the reference layout...",
      "thinking": "Let me analyze this systematically...",
      "stopReason": "end_turn"
    },
    "latencyMs": 2840,
    "tokenUsage": {
      "input": 4252,
      "output": 1837,
      "thinking": 876,
      "total": 6965
    },
    "analysisStage": "REFERENCE_ANALYSIS"
  },
  "conversationState": {
    "currentTurn": 1,
    "maxTurns": 2,
    "history": [...]
  }
}
```

## Error Handling

The function implements comprehensive error handling with specific error types:

- **ValidationException**: Input validation errors
- **BedrockException**: Bedrock API errors (throttling, model issues)
- **S3Exception**: S3 access or format errors
- **TimeoutException**: Operation timeouts
- **InternalException**: Internal processing errors

Errors include detailed information for debugging while avoiding exposure of sensitive data.

## Monitoring

### CloudWatch Metrics

Key metrics to monitor:

- `Duration`: Function execution time
- `Errors`: Error rate and types
- `TokenUsage`: Bedrock token consumption
- `RetryRate`: Frequency of retries
- `MemoryUtilization`: Memory usage patterns

### Logs

All logs are in structured JSON format:

```json
{
  "timestamp": "2025-04-21T15:30:25Z",
  "level": "INFO",
  "verificationId": "verif-2025042115302500",
  "component": "bedrock",
  "message": "Invoking Bedrock model",
  "metadata": {
    "modelId": "anthropic.claude-3-7-sonnet-20250219",
    "tokenCount": 4252
  }
}
```

## Development

### Local Testing

```bash
# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Structure

```
execute-turn1/
├── cmd/
│   └── main.go          # Lambda handler entry point
├── internal/
│   ├── types.go         # Data structures and types
│   ├── validation.go    # Input validation logic
│   ├── bedrock.go       # Bedrock API client
│   ├── response.go      # Response processing
│   ├── retry.go         # Retry logic and patterns
│   └── errors.go        # Error handling and types
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## Troubleshooting

### Common Issues

**High token usage**
- Check prompt length and optimize
- Verify thinking tokens budget
- Review image resolution

**Timeout errors**
- Increase Lambda timeout
- Optimize retry configuration
- Check Bedrock API limits

**S3 access errors**
- Verify IAM permissions
- Check bucket policy
- Ensure correct URI format

### Debug Mode

Enable debug logging by setting log level to DEBUG in CloudWatch.

## Contributing

1. Follow Go conventions and best practices
2. Add tests for new features
3. Update documentation
4. Follow the existing error handling patterns
5. Ensure all environment variables are documented

## License

Copyright © 2025 Kootoro. All rights reserved.

## Support

For technical support, please contact the development team or create an issue in the project repository.