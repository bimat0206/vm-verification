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
./main < events/initialization.json
./main < events/s3-reference-input.json
```

### Build and Deploy as Container
```bash
# Build the container
docker build -t kootoro-prepare-system-prompt:v2.1.0 .

# Tag for ECR
docker tag kootoro-prepare-system-prompt:v2.1.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v2.1.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v2.1.0

# Alternative build script if the main build fails
./retry-docker-build.sh
```

## Architecture

### Core Components
1. **Lambda Handler**: `cmd/main.go` - Entry point that processes events and returns responses
2. **Configuration**: `internal/config/config.go` - Application configuration
3. **Logging**: `internal/logging/logger.go` - Structured logging
4. **Models**: `internal/models/` - Data models and conversions
5. **State Management**: `internal/state/` - S3 state management
6. **Template Management**: `internal/template/template_manager.go` - Template loading and rendering
7. **Input Validation**: `internal/validation/` - Input validation with S3 helpers

### Data Flow
1. Lambda receives verification context or S3 reference envelope
2. Input is parsed and validated based on verification type
3. State is loaded from S3 if needed
4. Template is loaded based on verification type and version
5. Template data is constructed from input
6. Template is rendered with the data
7. Bedrock configuration is created
8. System prompt is stored in S3
9. Final response with S3 references is assembled and returned

### Verification Types
1. **LAYOUT_VS_CHECKING**: Compares a reference layout image with a real-time checking image
2. **PREVIOUS_VS_CURRENT**: Compares a previous verification image with a current image

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| REFERENCE_BUCKET | S3 bucket for reference layout images | - | Yes |
| CHECKING_BUCKET | S3 bucket for checking images | - | Yes |
| STATE_BUCKET | S3 bucket for state management | - | Yes |
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates | No |
| COMPONENT_NAME | Component name for logging | PrepareSystemPrompt | No |
| DATE_PARTITION_TIMEZONE | Timezone for date partitioning | UTC | No |
| MAX_TOKENS | Maximum tokens for response | 24000 | No |
| BUDGET_TOKENS | Tokens for Claude's thinking process | 16000 | No |
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

## S3 State Management

The function uses a date-based hierarchical storage system for managing state in S3, with the following structure:
```
{STATE_BUCKET}/
└── {YYYY}/
    └── {MM}/
        └── {DD}/
            └── {verificationId}/
                ├── initialization.json     - Initial verification context
                ├── prompts/                - System prompts
                │   └── system-prompt.json  - Generated system prompt
                ├── images/                 - Image data (if stored in S3)
                └── processing/             - Processing artifacts
```

## Dependencies

The application uses:
1. AWS Lambda Go runtime
2. AWS SDK Go v2 for S3 operations
3. Shared packages:
   - schema - Common data structures
   - logger - Logging interface
   - templateloader - Template loading and management

## Best Practices

1. Always validate inputs before processing
2. Use structured logging with verification context
3. Maintain backward compatibility when updating templates
4. Follow the established error handling pattern
5. Ensure image formats are supported by Bedrock (JPEG/PNG only)
6. Template versions should follow semantic versioning