# ExecuteTurn1 Lambda Function

This Lambda executes the first conversation turn with Amazon Bedrock, analyzing the reference image for the vending machine verification workflow.  
**Now fully standardized on a shared schema, logger, and error package for robust, auditable, and maintainable production use.**  
**Refactored with modular code organization for improved maintainability and easier troubleshooting.**

---

## Overview

**ExecuteTurn1** is a core component in the AI-powered vending machine verification solution. It:

1. Accepts the canonical workflow state and system/user prompts (per shared schema).
2. Retrieves image data via hybrid Base64 logic (inline or S3 as needed).
3. Invokes the Bedrock InvokeModel API (Claude 3.7 Sonnet) using a standardized JSON payload.
4. Processes and structures the response, including "thinking" content if enabled.
5. Updates the workflow state, including full turn history, verification context, and standardized error surfaces.

---

## Key Features

- **Modular Code Structure**:  
  - Codebase is organized into focused modules with clear separation of concerns for better maintainability and troubleshooting.
- **Shared Schema**:  
  - All requests, responses, prompts, and images use a common Go struct package, enabling strong validation and seamless Step Functions integration.
- **Hybrid Base64 Storage**:  
  - Images are handled as Base64 inline or via S3, based on size and configuration. S3 retrieval and embedding is automatic and fully abstracted.
- **Bedrock InvokeModel API**:  
  - Supports the Claude 3.7 Sonnet multimodal model via standardized JSON format.
- **Centralized Logging**:  
  - All logs are structured JSON, with service/function, severity, and correlation ID for distributed traceability.
- **Comprehensive Error Handling**:  
  - Every error (validation, AWS, Bedrock, etc.) is captured as a `WorkflowError` (typed, retryable, with context), and surfaced both in Lambda and in `VerificationContext.Error` for Step Functions.
- **Step Functions Ready**:  
  - All state transitions and output match the canonical shared schema for downstream workflow compatibility.

---

## Directory Structure



ExecuteTurn1/
├── cmd/
│ └── main.go # Lambda handler entry point
├── internal/
│ ├── handler/
│ │ ├── handler.go # Core handler structure and main request handling
│ │ ├── validation.go # Validation-related functions
│ │ ├── image_processor.go # Image processing functionality
│ │ ├── bedrock_client.go # Bedrock API interaction
│ │ ├── error_handler.go # Error handling utilities
│ │ ├── state_manager.go # State management functions
│ │ └── response_processor.go # Response processing logic
│ ├── config/
│ │ └── config.go # Centralized configuration
│ ├── dependencies/
│ │ └── clients.go # AWS client initialization
│ └── request/
│ └── request.go # Lambda input/output and validation helpers
├── shared/ # Shared schema/logger/error packages
│ ├── schema/
│ ├── logger/
│ └── errors/
├── Dockerfile
├── go.mod
├── go.sum
└── README.md


---

## Configuration

The function uses environment variables for all settings.  
**All values are strongly validated at startup.**

| Variable                | Description                                | Default                                     |
|-------------------------|--------------------------------------------|---------------------------------------------|
| BEDROCK_MODEL_ID        | Bedrock model identifier                   | anthropic.claude-3-7-sonnet-20250219-v1:0   |
| AWS_REGION              | AWS region for services                    | us-east-1                                   |
| ANTHROPIC_VERSION       | Anthropic API version                      | bedrock-2023-05-31                          |
| MAX_TOKENS              | Max tokens for Bedrock response            | 4000                                        |
| TEMPERATURE             | Model sampling temperature                 | 0.7                                         |
| THINKING_TYPE           | Type of thinking to extract                | thoroughness                                |
| THINKING_BUDGET_TOKENS  | Token budget for thinking                  | 16000                                       |
| ENABLE_HYBRID_STORAGE   | Enable S3 storage for large Base64         | true                                        |
| TEMP_BASE64_BUCKET      | S3 bucket for Base64 storage               | temp-base64-bucket                          |
| BASE64_SIZE_THRESHOLD   | Size threshold for S3 storage (bytes)      | 2097152 (2MB)                               |
| BASE64_RETRIEVAL_TIMEOUT| Timeout for S3 retrieval (ms)              | 30000                                       |
| BEDROCK_TIMEOUT         | Timeout for Bedrock API calls (ms)         | 120000                                      |
| FUNCTION_TIMEOUT        | Overall function timeout (ms)              | 120000                                      |

---

## Building and Deployment

### Local Build

```bash
go build -o main ./cmd/main.go



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