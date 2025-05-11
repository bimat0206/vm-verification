# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This repository contains the PrepareSystemPrompt Lambda function, a component of the Kootoro GenAI Vending Machine Verification solution. The function generates system prompts for Amazon Bedrock (Claude 3.7 Sonnet) based on verification type, formatting data for the vending machine verification workflow.

## Common Commands

### Build and Run Locally
```bash
# Download dependencies
go mod download

# Run tests
go test ./...

# Build locally
go build -o main cmd/main.go

# Local testing with sample events
./main < events/layout-vs-checking.json
```

### Build and Deploy as Container
```bash
# Build the container
docker build -t kootoro-prepare-system-prompt:v1.0.0 .

# Tag for ECR
docker tag kootoro-prepare-system-prompt:v1.0.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v1.0.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v1.0.0

# Alternative build script if the main build fails
./retry-docker-build.sh
```

## Architecture

### Core Components
1. **Lambda Handler**: `cmd/main.go` - Entry point that processes events and returns responses
2. **Template Management**: `internal/templates.go` - Handles loading and rendering Go templates with version support
3. **Input Validation**: `internal/validator.go` - Comprehensive validation for input data
4. **Bedrock Integration**: `internal/bedrock.go` - Amazon Bedrock API configuration
5. **Data Processing**: `internal/processor.go` - Prepares data for template rendering
6. **Utility Functions**: `internal/utils.go` - Helper functions for various operations
7. **Type Definitions**: `internal/types.go` - Data structures used throughout the application

### Data Flow
1. Lambda receives verification context and metadata
2. Input is validated based on verification type
3. Template is loaded based on verification type and version
4. Template data is constructed from input
5. Template is rendered with the data
6. Bedrock configuration is created
7. Final response is assembled and returned

### Verification Types
1. **LAYOUT_VS_CHECKING**: Compares a reference layout image with a real-time checking image
2. **PREVIOUS_VS_CURRENT**: Compares a previous verification image with a current image

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| REFERENCE_BUCKET | S3 bucket for reference layout images | - | Yes |
| CHECKING_BUCKET | S3 bucket for checking images | - | Yes |
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates | No |
| ANTHROPIC_VERSION | Anthropic API version for Bedrock | bedrock-2023-05-31 | No |
| MAX_TOKENS | Maximum tokens for response | 24000 | No |
| BUDGET_TOKENS | Tokens for Claude's thinking process | 16000 | No |
| THINKING_TYPE | Claude's thinking mode | enabled | No |
| PROMPT_VERSION | Default prompt version | 1.0.0 | No |
| DEBUG | Enable debug logging | false | No |

## Template System

Templates use Go's text/template format and are organized in the following structure:
```
/opt/templates/
├── layout-vs-checking/
│   ├── v1.0.0.tmpl
│   ├── v1.1.0.tmpl
│   └── v1.2.3.tmpl
└── previous-vs-current/
    ├── v1.0.0.tmpl
    └── v1.1.0.tmpl
```

To add new templates:
1. Create a new template file in the appropriate directory
2. Use Go's text/template syntax
3. The template will be automatically discovered

## Best Practices

1. Always validate inputs before processing
2. Use environment variables for configuration
3. Maintain backward compatibility when updating templates
4. Follow the established error handling pattern
5. Ensure image formats are supported by Bedrock (JPEG/PNG only)
6. Template versions should follow semantic versioning