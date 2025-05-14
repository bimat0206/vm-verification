# Using the PromptUtils Shared Package

This guide explains how to integrate the `promptutils` shared package into your Lambda functions for the Vending Machine Verification system.

## Integration Steps

### 1. Update your go.mod file

Add the local replacement for the shared module:

```go
module your-function-name

go 1.22

require (
    github.com/aws/aws-lambda-go v1.46.0
    github.com/aws/aws-sdk-go-v2 v1.24.1
    // Other dependencies...
    shared/promptutils v0.0.0
)

replace shared/promptutils => ../shared/promptutils
```

### 2. Adapt your main.go file

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/aws/aws-lambda-go/lambda"
    "shared/promptutils"
)

var (
    promptProcessor *promptutils.PromptProcessor
)

func init() {
    // Initialize template manager with base path from environment or default
    templateBasePath := os.Getenv("TEMPLATE_BASE_PATH")
    if templateBasePath == "" {
        templateBasePath = "/opt/templates" // Default in container
    }
    
    // Create prompt processor
    promptProcessor = promptutils.NewPromptProcessor(templateBasePath)
    log.Printf("Initialized prompt processor with base path: %s", templateBasePath)
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (promptutils.Response, error) {
    start := time.Now()
    log.Printf("Received event: %s", string(event))
    
    // Process input and generate system prompt
    response, err := promptProcessor.ProcessInput(event)
    if err != nil {
        log.Printf("Error processing input: %v", err)
        return promptutils.Response{}, err
    }
    
    // Additional custom processing can be done here
    
    log.Printf("Completed in %v", time.Since(start))
    return response, nil
}

func main() {
    lambda.Start(HandleRequest)
}
```

### 3. Update your Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod first for better caching
COPY go.mod go.sum ./
COPY ../shared/promptutils /shared/promptutils

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build
RUN go build -o main cmd/main.go

# Create runtime container
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy templates directory
COPY templates /opt/templates

# Set environment variables
ENV TEMPLATE_BASE_PATH=/opt/templates
ENV COMPONENT_NAME=YourFunctionName

# Run the binary
ENTRYPOINT ["/app/main"]
```

## Function-Specific Templates

Each function can have its own set of templates while reusing the shared prompt generation code. Just place your templates in the appropriate directory structure:

```
/opt/templates/
├── layout-vs-checking/
│   └── v1.0.0.tmpl
└── previous-vs-current/
    └── v1.0.0.tmpl
```

## Working with Bedrock

The shared package includes helpers for Bedrock API integration:

```go
// Prepare a Bedrock request with your system prompt
bedrockConfig := promptutils.BedrockConfig{
    AnthropicVersion: "bedrock-2023-05-31",
    MaxTokens: 24000,
    Thinking: promptutils.ThinkingConfig{
        Type: "enabled",
        BudgetTokens: 16000,
    },
}

// Create a standard request
bedrockRequest, err := promptProcessor.PrepareBedrockRequest(
    response.SystemPrompt.Content, 
    bedrockConfig,
)

// Or for turn-based conversations:
userMessage := promptutils.CreateUserMessageWithImage(
    "Please verify this image",
    base64ImageData,
    "jpeg",
)

bedrockTurnRequest, err := promptProcessor.PrepareBedrockTurnRequest(
    response.SystemPrompt.Content,
    userMessage,
    bedrockConfig,
)
```

## Versioning Templates

Template versions follow semantic versioning and are automatically discovered by the template manager. To use a specific version:

1. Set an environment variable: `TEMPLATE_VERSION_LAYOUT_VS_CHECKING=1.2.3`
2. Or place it in the correct directory structure: `/opt/templates/layout-vs-checking/v1.2.3.tmpl`

## Validation Rules

The shared package includes comprehensive validation for:

- Verification context (ID, type, timestamp)
- Machine structure (rows, columns, mappings)
- Image URLs (S3 paths, image types)
- Historical data (when applicable)

You can extend the validation with custom rules for specific functions if needed.

## Logging

The package includes structured logging with different log levels:

```go
import "shared/promptutils/utils"

// Log at different levels
utils.LogDebug("Processing template", verificationID, verificationType, details)
utils.LogInfo("Generated system prompt", verificationID, verificationType, nil)
utils.LogWarn("Template not found, using default", verificationID, verificationType, details, err)
utils.LogError("Failed to generate prompt", verificationID, verificationType, details, err)
```

## Testing Locally

Use the example in `examples/main.go` as a starting point for local testing:

```bash
# Set required environment variables
export REFERENCE_BUCKET=reference-bucket
export CHECKING_BUCKET=checking-bucket
export COMPONENT_NAME=TestFunction

# Run with test input
cd examples
go run main.go test-input.json
```