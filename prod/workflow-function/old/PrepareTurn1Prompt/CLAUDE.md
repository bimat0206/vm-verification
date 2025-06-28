# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains the PrepareTurn1Prompt Lambda function for the Kootoro GenAI Vending Machine Verification solution. The function generates properly formatted first turn prompts for Amazon Bedrock (Claude 3.7 Sonnet) to analyze vending machine reference images.

## Key Commands

### Building and Testing

```bash
# Download dependencies
go mod download

# Build locally
go build -o main cmd/main.go

# Run tests
go test ./...

# Build Docker container
docker build -t kootoro-prepare-turn1-prompt:v1.0.0 .

# Tag for ECR (replace with actual AWS details)
docker tag kootoro-prepare-turn1-prompt:v1.0.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-turn1-prompt:v1.0.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-turn1-prompt:v1.0.0
```

## Architecture and Workflow

The function follows this workflow with S3 state management:
1. Loads state data from S3 references (verification context, images)
2. Loads the appropriate template based on verification type
3. Processes images to ensure they have Base64 data for Bedrock
4. Builds template data context from the loaded state
5. Processes the template to generate the Turn 1 prompt text
6. Creates a Bedrock message with both prompt text and reference image
7. Stores the prompt data in S3 and updates verification status
8. Returns a lightweight S3 reference envelope to the next function

### Key Components

- **main.go**: Lambda handler and core execution logic
- **types.go**: Data structures and type definitions
- **validator.go**: Input validation logic
- **templates.go**: Template loading and management
- **bedrock.go**: Bedrock message construction
- **processor.go**: Core business logic for prompt creation
- **utils.go**: Helper functions
- **s3client.go**: S3 operations for state management

### Verification Types

The function supports two verification types:
- **LAYOUT_VS_CHECKING**: Compares a layout reference image with a checking image
- **PREVIOUS_VS_CURRENT**: Compares previous and current state images

### S3 State Management

The function uses the S3 state management pattern:
1. **Input**: Receives S3 references envelope
2. **Load**: Loads state data from S3 using the references
3. **Process**: Generates Turn 1 prompt and Bedrock message
4. **Store**: Saves prompt data and metadata to S3
5. **Output**: Returns updated S3 references for the next function

#### S3 Categories
- **initialization**: Initial verification context and metadata
- **images**: Image data including Base64 and metadata
- **processing**: Intermediate processing results and metrics
- **prompts**: Generated Turn 1 prompt with Bedrock messages
- **responses**: Bedrock responses (used by downstream functions)

#### S3 Files
- **initialization.json**: Initial verification context
- **metadata.json**: Image metadata and information
- **turn1-prompt.json**: Generated Turn 1 prompt
- **turn1-metrics.json**: Processing metrics and status

#### Backward Compatibility
The function maintains backward compatibility with the old payload-based approach through a feature flag.

### Template System

The function uses Go templates stored in the `/opt/templates` directory:
- Templates are organized by verification type (`layout-vs-checking/` or `previous-vs-current/`)
- Templates use semantic versioning (x.y.z)
- Template placeholders are filled with structured data from the verification context

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| STATE_BUCKET | S3 bucket for state management | *required* |
| ENABLE_S3_STATE | Enable S3 state management | true |
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates |
| ANTHROPIC_VERSION | Anthropic API version | bedrock-2023-05-31 |
| MAX_TOKENS | Maximum tokens for response | 24000 |
| BUDGET_TOKENS | Tokens for Claude's thinking | 16000 |
| THINKING_TYPE | Claude's thinking mode | enabled |
| DEBUG | Enable debug logging | false |

## Important Development Notes

1. Always validate input data before processing
2. Use the S3StateManager for all state operations
3. Handle image processing with the enhanced image utilities
4. Ensure template versions match the expected format
5. Handle S3 image references properly as Bedrock requires specific formatting
6. Format Bedrock requests according to the Anthropic API specifications
7. Use structured logging for better observability
8. Store metrics and processing status in S3 for monitoring