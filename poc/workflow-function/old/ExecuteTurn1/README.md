# ExecuteTurn1 Lambda Function

This Lambda handles the first conversation turn with Bedrock (Claude 3.7) for vending machine verification, now completely redesigned with a **S3 state reference architecture** for better performance, scalability, and maintainability.

---

## Overview

**ExecuteTurn1** is a critical component in the vending machine verification workflow, reimagined with a modern state reference architecture:

1. Accepts S3 references to workflow state components (no large payloads)
2. Loads state components from S3 using category-based organization
3. Processes the Turn1 conversation with Bedrock using shared client package
4. Saves results back to S3 with optimized storage patterns
5. Returns lightweight references for downstream Step Functions processing

---

## Key Architectural Improvements

- **S3 State Reference Architecture**:  
  - Replaced large workflow state payloads with lightweight S3 references
  - Implemented category-based state storage with shared S3StateManager
  - Improved performance by eliminating serialization/deserialization overhead
  - Enhanced scalability by removing API Gateway and Lambda payload size limitations

- **Modular Component Design**:  
  - Complete separation of concerns with focused, single-responsibility modules
  - Clear boundaries between state management, validation, and business logic
  - Improved testability with proper dependency injection
  - No file exceeds 300 lines for better maintainability

- **Shared Package Integration**:  
  - Fully leverages shared/s3state for standardized state management
  - Uses shared/bedrock for consistent Bedrock API interactions
  - Standardized error handling with shared/errors WorkflowError types
  - Enhanced logging with shared/logger structured JSON logging

- **Enhanced Configuration**:  
  - Improved environment variable organization by functional area
  - Sensible defaults with comprehensive validation
  - Better configuration error reporting
  - Support for hybrid storage modes and timeouts

---

## New Directory Structure

```
ExecuteTurn1/
├── cmd/
│   └── main.go              # Lambda handler entry point
├── internal/
│   ├── state/               # S3 state management
│   │   ├── loader.go        # State loading operations
│   │   └── saver.go         # State saving operations
│   ├── bedrock/             # Bedrock integration
│   │   └── client.go        # Bedrock API wrapper
│   ├── validation/          # Input validation
│   │   └── validator.go     # Schema validation
│   ├── handler/             # Core business logic
│   │   └── handler.go       # Request coordination
│   ├── config/              # Configuration
│   │   └── config.go        # Environment config
│   └── dependencies/        # Dependencies
│       └── clients.go       # Client initialization
├── Dockerfile               # Container build
├── go.mod                   # Go module definition
└── README.md                # Documentation
```

---

## Configuration

The function uses these environment variables with sensible defaults:

| Variable                | Description                           | Default                     |
|-------------------------|---------------------------------------|----------------------------|
| **S3 State Management** |                                       |                            |
| STATE_BUCKET            | Bucket for S3 state storage           | (required)                  |
| STATE_BASE_PREFIX       | Base prefix for S3 state objects      | verification-states         |
| STATE_TIMEOUT           | Timeout for state operations          | 30s                         |
| **Bedrock Configuration** |                                     |                            |
| BEDROCK_MODEL_ID        | Claude model identifier               | (required)                  |
| BEDROCK_REGION          | AWS region for Bedrock                | (required)                  |
| ANTHROPIC_VERSION       | Anthropic API version                 | bedrock-2023-05-31          |
| MAX_TOKENS              | Max tokens for response               | 4096                        |
| TEMPERATURE             | Model sampling temperature            | 0.7                         |
| THINKING_TYPE           | Thinking extraction pattern           | thinking                    |
| THINKING_BUDGET_TOKENS  | Max tokens for thinking               | 16000                       |
| **Image Processing**    |                                       |                            |
| ENABLE_HYBRID_STORAGE   | Use S3 for large Base64               | true                        |
| TEMP_BASE64_BUCKET      | Bucket for Base64 storage             | (required if hybrid=true)   |
| BASE64_SIZE_THRESHOLD   | Size threshold for S3 (bytes)         | 1048576 (1MB)               |
| **Timeouts**            |                                       |                            |
| BEDROCK_TIMEOUT         | Timeout for Bedrock calls             | 120s                        |
| FUNCTION_TIMEOUT        | Overall function timeout              | 240s                        |
| RETRY_MAX_ATTEMPTS      | Max retry attempts                    | 3                           |
| RETRY_BASE_DELAY        | Base delay for retries                | 1s                          |

---

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

```bash
./retry-docker-build.sh --repo=<ECR_REPO_URI> --function=<LAMBDA_FUNCTION_NAME> --region=<AWS_REGION>
```

Required IAM permissions:
- `bedrock:InvokeModel` for the specified model
- `s3:GetObject` and `s3:PutObject` for the state bucket and Base64 bucket

---

## Input and Output

### Input (S3 References)

```json
{
  "stateReferences": {
    "verificationId": "verif-12345",
    "verificationContext": {
      "bucket": "state-bucket",
      "key": "contexts/verif-12345/verification-context.json"
    },
    "systemPrompt": {
      "bucket": "state-bucket",
      "key": "prompts/verif-12345/system-prompt.json"
    },
    "images": {
      "bucket": "state-bucket",
      "key": "images/verif-12345/image-data.json"
    },
    "bedrockConfig": {
      "bucket": "state-bucket",
      "key": "configs/verif-12345/bedrock-config.json"
    }
  }
}
```

### Output (Updated References)

```json
{
  "stateReferences": {
    "verificationId": "verif-12345",
    "verificationContext": {
      "bucket": "state-bucket",
      "key": "contexts/verif-12345/verification-context.json"
    },
    "systemPrompt": {
      "bucket": "state-bucket",
      "key": "prompts/verif-12345/system-prompt.json"
    },
    "images": {
      "bucket": "state-bucket",
      "key": "images/verif-12345/image-data.json"
    },
    "bedrockConfig": {
      "bucket": "state-bucket",
      "key": "configs/verif-12345/bedrock-config.json"
    },
    "conversationState": {
      "bucket": "state-bucket",
      "key": "conversations/verif-12345/conversation-state.json"
    },
    "turn1Response": {
      "bucket": "state-bucket",
      "key": "responses/verif-12345/turn1-response.json"
    },
    "turn1Thinking": {
      "bucket": "state-bucket",
      "key": "thinking/verif-12345/turn1-thinking.json"
    }
  },
  "status": "TURN1_COMPLETED",
  "summary": {
    "tokenUsage": 3500,
    "latencyMs": 12500,
    "status": "TURN1_COMPLETED"
  }
}
```

## Error Handling

The function returns structured errors with the same reference pattern:

```json
{
  "status": "BEDROCK_PROCESSING_FAILED",
  "error": {
    "code": "BEDROCK_API_ERROR",
    "message": "Failed to call Bedrock API",
    "timestamp": "2023-05-19T08:12:30Z",
    "details": {
      "error": "Quota exceeded",
      "retryable": true
    }
  },
  "summary": {
    "error": "Failed to call Bedrock API",
    "status": "BEDROCK_PROCESSING_FAILED",
    "retryable": true
  }
}
```

## Monitoring and Logging

All function activity is logged to CloudWatch Logs with structured JSON format:

```json
{
  "level": "info",
  "timestamp": "2025-05-19T12:34:56Z",
  "service": "vending-verification",
  "function": "ExecuteTurn1",
  "correlationId": "req-abc123",
  "verificationId": "verif-12345",
  "step": "Turn1Processing",
  "message": "Bedrock API call successful",
  "data": {
    "latencyMs": 12500,
    "tokenUsage": 3500,
    "modelId": "anthropic.claude-3-sonnet-20250219-v1:0"
  }
}
```

Key metrics to monitor:
- Bedrock API latency (average, p95, p99)
- S3 operation latency
- Error rates by error type
- Lambda duration and memory utilization
- State size metrics

---

## Development

### Prerequisites

- Go 1.22+
- Docker
- AWS CLI configured with appropriate permissions
- S3 bucket for state storage

### Local Testing

```bash
# Set up environment variables
export STATE_BUCKET=my-state-bucket
export BEDROCK_MODEL_ID=anthropic.claude-3-sonnet-20250219-v1:0
export BEDROCK_REGION=us-east-1

# Build and run locally
go build -o main ./cmd/main.go
./main
```