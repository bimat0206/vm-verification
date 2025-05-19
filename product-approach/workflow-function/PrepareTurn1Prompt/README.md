# PrepareTurn1Prompt Lambda Function

## Overview

The `PrepareTurn1Prompt` Lambda function is responsible for generating the first turn prompt for the vending machine verification workflow. It takes S3 references as input, loads the necessary data, processes images, and generates a prompt for the Bedrock model. This function has been refactored to use the new S3 State Management architecture with shared packages integration.

## Version

Current version: 4.0.0 (See [CHANGELOG.md](./CHANGELOG.md) for version history)

## Architecture

### Input/Output Architecture

- **Input**: S3 references to initialization data, images, and processing categories
- **Output**: S3 references to stored prompt data and updated context

#### Input Example:
```json
{
  "references": {
    "initialization_initialization": {
      "bucket": "vending-machine-verification-dev",
      "key": "initialization/verification-123/initialization.json",
      "size": 512
    },
    "images_reference": {
      "bucket": "vending-machine-verification-dev",
      "key": "images/verification-123/reference-image.json",
      "size": 2048
    },
    "processing_layout_metadata": {
      "bucket": "vending-machine-verification-dev",
      "key": "processing/verification-123/layout-metadata.json",
      "size": 1024
    }
  },
  "verificationId": "verification-123",
  "verificationType": "LAYOUT_VS_CHECKING",
  "turnNumber": 1,
  "includeImage": "reference",
  "enableS3StateManager": true
}
```

#### Output Example:
```json
{
  "references": {
    "initialization_initialization": {
      "bucket": "vending-machine-verification-dev",
      "key": "initialization/verification-123/initialization.json",
      "size": 512
    },
    "prompts_turn1-prompt": {
      "bucket": "vending-machine-verification-dev",
      "key": "prompts/verification-123/turn1-prompt.json",
      "size": 1024
    },
    "processing_turn1-metrics": {
      "bucket": "vending-machine-verification-dev",
      "key": "processing/verification-123/turn1-metrics.json",
      "size": 256
    }
  },
  "verificationId": "verification-123",
  "verificationType": "LAYOUT_VS_CHECKING",
  "status": "TURN1_PROMPT_READY"
}
```

### S3 State Management

- Uses `shared/s3state` package for all state operations
- Loads state data from S3 references provided in input
- Returns lightweight S3 references pointing to stored prompt data
- Implements automatic retry and error handling for S3 operations
- Supports concurrent operations for improved performance

### Image Processing

- Supports multiple storage methods for images:
  - Inline Base64 data
  - S3-temporary storage (pre-encoded Base64 in S3)
  - S3-direct storage (download and encode)
- Uses transparent processing with automatic format detection and validation
- Validates image size against Bedrock limits
- Implements efficient Base64 encoding with minimal memory overhead
- Handles various image formats (PNG, JPEG)

### Template and Prompt Management

- Templates are organized by verification type and version
- Prompt text is generated from templates and stored in S3
- Complete Bedrock message structure is created and stored
- Supports template versioning for backward compatibility
- Includes metadata for tracking and debugging

## Directory Structure

```
PrepareTurn1Prompt/
├── cmd/
│   └── main.go                 # Lambda entry point
├── internal/
│   ├── core/                   # Core business logic
│   │   ├── prompt_generator.go    # Generate Turn 1 prompt text
│   │   ├── response_builder.go    # Build response for downstream functions
│   │   └── template_processor.go  # Process templates with data
│   ├── images/                 # Image processing
│   │   ├── processor.go           # Main image processing coordination
│   │   ├── bedrock_prep.go        # Bedrock-specific image preparation
│   │   └── format_detector.go     # Image format detection and validation
│   ├── integration/            # External systems integration
│   │   ├── bedrock.go             # Bedrock message creation
│   │   └── workflow.go            # Workflow helpers
│   ├── state/                  # S3 state management
│   │   ├── loader.go              # Load state from S3
│   │   ├── references.go          # S3 reference handling
│   │   └── saver.go               # Save state to S3
│   └── validation/             # Input validation
│       ├── input_validator.go     # Validate input parameters
│       ├── context_validator.go   # Verify verification context
│       └── type_validators.go     # Type-specific validation
├── templates/                  # Prompt templates
│   ├── turn1-layout-vs-checking/
│   │   └── v1.0.0.tmpl            # Template for LAYOUT_VS_CHECKING
│   └── turn1-previous-vs-current/
│       └── v1.0.0.tmpl            # Template for PREVIOUS_VS_CURRENT
├── Dockerfile                  # Container definition
└── retry-docker-build.sh       # Build and deployment script
```

## Workflow

The function follows this workflow:

1. Receive S3 references as input
2. Validate the input structure and references
3. Load verification context, images, and metadata from S3
4. Process images to ensure Base64 encoding for Bedrock
5. Generate Turn 1 prompt using appropriate template based on verification type
6. Create Bedrock message with prompt text and reference image
7. Update verification status to TURN1_PROMPT_READY
8. Store prompt data and metrics in S3
9. Return S3 references envelope to next function

## Environment Variables

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

## Deployment

Use the `retry-docker-build.sh` script to build and deploy the function:

```bash
./retry-docker-build.sh
```

This will:
1. Build the Docker image
2. Push the image to ECR
3. Update the Lambda function

## Error Handling

The function uses structured error types from the shared errors package:
- `errors.NewValidationError()`: For input validation errors
- `errors.NewInternalError()`: For internal processing errors
- `errors.NewMissingFieldError()`: For required fields that are missing
- `errors.NewInvalidFieldError()`: For fields with invalid values

### Error Scenarios

| Error Type | Description | Recovery Strategy |
|------------|-------------|-------------------|
| Validation Error | Invalid input format or missing required fields | Fix input format and retry |
| S3 State Access | Unable to access S3 objects or permissions issues | Check IAM permissions and S3 bucket configuration |
| Base64 Processing | Issues with image encoding or size limits | Verify image format and size, consider compression |
| Template Processing | Template not found or execution error | Check template path and syntax |
| Internal Error | Unexpected errors during processing | Check logs for detailed error context |

## Logging

The function uses structured logging from the shared logger package, with consistent log levels and context fields. Log entries include:

- Verification ID
- Processing timestamps
- Operation durations
- Error contexts
- S3 reference details
- Template information

## Integration with Workflow

### Upstream Functions

The PrepareTurn1Prompt function expects input from:
- `Initialize`: Provides initial verification context
- `FetchImages`: Provides image references

### Downstream Functions

The PrepareTurn1Prompt function outputs to:
- `ProcessTurn1Response`: Consumes the generated prompt and Bedrock response

## Verification Types

The function supports two verification types:
- `LAYOUT_VS_CHECKING`: Compares a layout reference image with a checking image
- `PREVIOUS_VS_CURRENT`: Compares previous and current state images

## S3 State Categories

The function organizes data in S3 using these categories:

- `initialization`: Initial verification context and metadata
- `images`: Image data including Base64 and metadata
- `processing`: Intermediate processing results and metrics
- `prompts`: Generated Turn 1 prompt with Bedrock messages
- `responses`: Bedrock responses (used by downstream functions)

## Performance Considerations

- **Memory Usage**: The function is optimized to handle Base64 encoding efficiently
- **Concurrency**: S3 operations use AWS SDK's built-in concurrency for improved performance
- **Caching**: Templates are cached for improved performance
- **Timeouts**: Default Lambda timeout is set to 30 seconds
- **Error Handling**: Includes automatic retries for transient S3 errors

## Testing

Local testing can be performed with:

```bash
# Set environment variables
export TEMPLATE_BASE_PATH="./templates"
export STATE_BUCKET="vending-machine-verification-dev"

# Run with test input
export INPUT_FILE="test-input.json"
go run cmd/main.go
```

## Contributing

When making changes to this function:

1. Follow the modular architecture with clear separation of concerns
2. Ensure all shared packages are properly integrated
3. Test all changes with sample inputs
4. Update documentation for any API changes
