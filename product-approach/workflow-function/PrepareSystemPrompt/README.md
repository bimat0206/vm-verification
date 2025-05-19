# PrepareSystemPrompt Lambda Function

This Lambda function generates system prompts for the Kootoro GenAI Vending Machine Verification solution. It processes verification context and generates system prompts for Amazon Bedrock (Claude 3.7 Sonnet).

## Architecture

### Core Components

1. **Lambda Handler**: `cmd/main.go` - Entry point that handles events and returns responses
2. **Configuration**: `internal/config/config.go` - Application configuration from environment variables
3. **Models**: `internal/models/` - Data models and conversions
   - `input.go` - Input handling and processing
   - `output.go` - Response structures
   - `template_data.go` - Template data structures
4. **Adapters**: `internal/adapters/` - External service integrations
   - `s3state.go` - S3 state management
   - `bedrock.go` - Bedrock configuration and settings
5. **Processors**: `internal/processors/` - Business logic
   - `template.go` - Template loading and rendering
   - `validation.go` - Input validation
6. **Handlers**: `internal/handlers/` - Lambda request handling
   - `handler.go` - Main request handler

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

## Date-Based S3 Storage Structure

State is organized in a date-based hierarchical structure:

```
{STATE_BUCKET}/
└── {YYYY}/
    └── {MM}/
        └── {DD}/
            └── {verificationId}/
                ├── processing/
                │   └── initialization.json     - Initial verification context
                ├── prompts/
                │   └── system-prompt.json      - Generated system prompt
                └── images/                     - Image data (if stored in S3)
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| STATE_BUCKET | S3 bucket for state management | - | Yes |
| REFERENCE_BUCKET | S3 bucket for reference layout images | - | Yes |
| CHECKING_BUCKET | S3 bucket for checking images | - | Yes |
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates | No |
| COMPONENT_NAME | Component name for logging | PrepareSystemPrompt | No |
| DATE_PARTITION_TIMEZONE | Timezone for date partitioning | UTC | No |
| MAX_TOKENS | Maximum tokens for response | 24000 | No |
| BUDGET_TOKENS | Tokens for Claude's thinking process | 16000 | No |
| PROMPT_VERSION | Default prompt version | 1.0.0 | No |
| DEBUG | Enable debug logging | false | No |

## Build and Run

```bash
# Download dependencies
go mod download

# Build locally
go build -o main cmd/main.go

# Test locally (using sample input)
./main < test/test_input.json
```

## Docker Build

```bash
# Build container
docker build -t kootoro-prepare-system-prompt:v2.1.0 .

# Tag for ECR
docker tag kootoro-prepare-system-prompt:v2.1.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v2.1.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v2.1.0
```

## Dependencies

The application uses:
1. AWS Lambda Go runtime
2. AWS SDK Go v2 for S3 operations
3. Shared packages:
   - `shared/schema` - Common data structures
   - `shared/logger` - Logging interface
   - `shared/templateloader` - Template loading and management
   - `shared/s3state` - S3 state management
EOF < /dev/null