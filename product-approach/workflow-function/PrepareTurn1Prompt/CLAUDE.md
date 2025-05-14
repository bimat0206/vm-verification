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

The function follows this workflow:
1. Validates input data (verification context, image references)
2. Loads the appropriate template based on verification type
3. Builds template data context from input
4. Processes the template to generate the Turn 1 prompt text
5. Creates a Bedrock message with both prompt text and reference image
6. Returns a properly formatted response with the Bedrock request

### Key Components

- **main.go**: Lambda handler and core execution logic
- **types.go**: Data structures and type definitions
- **validator.go**: Input validation logic
- **templates.go**: Template loading and management
- **bedrock.go**: Bedrock message construction
- **processor.go**: Core business logic for prompt creation
- **utils.go**: Helper functions

### Verification Types

The function supports two verification types:
- **LAYOUT_VS_CHECKING**: Compares a layout reference image with a checking image
- **PREVIOUS_VS_CURRENT**: Compares previous and current state images

### Template System

The function uses Go templates stored in the `/opt/templates` directory:
- Templates are organized by verification type (`layout-vs-checking/` or `previous-vs-current/`)
- Templates use semantic versioning (x.y.z)
- Template placeholders are filled with structured data from the verification context

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates |
| ANTHROPIC_VERSION | Anthropic API version | bedrock-2023-05-31 |
| MAX_TOKENS | Maximum tokens for response | 24000 |
| BUDGET_TOKENS | Tokens for Claude's thinking | 16000 |
| THINKING_TYPE | Claude's thinking mode | enabled |
| DEBUG | Enable debug logging | false |

## Important Development Notes

1. Always validate input data before processing
2. Ensure template versions match the expected format
3. Handle S3 image references properly as Bedrock requires specific formatting
4. Format Bedrock requests according to the Anthropic API specifications