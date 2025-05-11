# Testing the Lambda Function

This directory contains sample event payloads for testing the Lambda function.

## Environment Variables

The Lambda function requires the following environment variables to be set:

```bash
# Required environment variables
REFERENCE_BUCKET=kootoro-dev-s3-reference-x1y2z3  # S3 bucket for reference layout images
CHECKING_BUCKET=kootoro-dev-s3-checking-f6d3xl    # S3 bucket for checking images

# Optional environment variables with defaults
TEMPLATE_BASE_PATH=/opt/templates            # Path to template directory
ANTHROPIC_VERSION=bedrock-2023-05-31         # Anthropic API version for Bedrock
MAX_TOKENS=24000                             # Maximum tokens for response
BUDGET_TOKENS=16000                          # Tokens for Claude's thinking process
THINKING_TYPE=enabled                        # Claude's thinking mode
PROMPT_VERSION=1.0.0                         # Default prompt version
DEBUG=false                                  # Enable debug logging
```

## Testing Locally

To test the function locally with the sample events:

```bash
# Set the environment variables
export REFERENCE_BUCKET=kootoro-dev-s3-reference-x1y2z3
export CHECKING_BUCKET=kootoro-dev-s3-checking-f6d3xl

# Build and run
go build -o main ../cmd/main.go
./main < layout-vs-checking.json
```

## Common Errors

### Bucket Validation

The function validates that the S3 URLs for images point to the correct buckets:

1. For `LAYOUT_VS_CHECKING` verification type:
   - `referenceImageUrl` must be in the `REFERENCE_BUCKET`
   - `checkingImageUrl` must be in the `CHECKING_BUCKET`

2. For `PREVIOUS_VS_CURRENT` verification type:
   - Both `referenceImageUrl` and `checkingImageUrl` must be in the `CHECKING_BUCKET`

Example of correct S3 URLs:
```json
"referenceImageUrl": "s3://kootoro-dev-s3-reference-x1y2z3/path/to/reference.jpg",
"checkingImageUrl": "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg"
```

If you get a bucket validation error, make sure your S3 URLs use the correct bucket names from the environment variables.