# FetchImages Lambda Function

This AWS Lambda function is part of the vending machine verification system. It fetches image metadata from S3 and retrieves relevant context from DynamoDB for verification workflows.

## Recent Updates

### Version 4.0.0 (2025-05-19)
- **Major Refactoring**: Completely restructured to use S3 State Manager pattern
- **Architectural Change**: Eliminated Base64 encoding of images in favor of S3 references
- **Project Structure**: Reorganized into a modern Go project structure with proper separation of concerns
- **Shared Packages**: Integrated with shared schema, logger, and s3state packages

See the [CHANGELOG.md](CHANGELOG.md) for a complete history of changes.

## Architecture

The FetchImages function has been refactored to use the S3 State Manager pattern:

- **S3 State Management**: All state is stored in S3, with references passed between Step Functions
- **Reference-Based Architecture**: No more Base64 encoding of images in responses
- **Clean Code Structure**: Follows Go best practices with a clear separation of concerns

## Project Structure

```
FetchImages/
├── cmd/
│   └── fetchimages/
│       └── main.go              # Lambda entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handler/
│   │   └── handler.go           # Lambda event handler
│   ├── models/
│   │   └── models.go            # Data structures and validation
│   ├── repository/
│   │   ├── s3_repository.go     # S3 operations
│   │   └── dynamodb_repository.go # DynamoDB operations
│   └── service/
│       ├── fetch_service.go     # Business logic
│       └── s3state_manager.go   # S3 state management
├── go.mod
└── Dockerfile
```

## Environment Variables

- `DYNAMODB_LAYOUT_TABLE` (default: `LayoutMetadata`): DynamoDB table for layout metadata
- `DYNAMODB_VERIFICATION_TABLE` (default: `VerificationResults`): DynamoDB table for verification results
- `STATE_BUCKET` (required): S3 bucket for state management
- `MAX_IMAGE_SIZE` (default: `104857600`): Maximum image size in bytes (100MB)

## Verification Types

The function supports two verification types:

1. **LAYOUT_VS_CHECKING**: Compares layout reference image with current checking image
   - Retrieves layout metadata from DynamoDB
   
2. **PREVIOUS_VS_CURRENT**: Compares previous checking image with current checking image
   - Retrieves historical verification data from DynamoDB

## S3 State Structure

```
s3://state-management-bucket/
├── {YYYY}/{MM}/{DD}/
    └── {verificationId}/
        ├── processing/
        │   ├── initialization.json       - Initial verification state
        │   ├── layout-metadata.json      # For UC1
        │   └── historical-context.json   # For UC2
        ├── images/
        │   ├── metadata.json             # Image metadata only
        │   ├── reference-base64.base64   # reference base64 encode
        │   └── checking-base64.base64    # checking base64 encode
        ├── prompts/
        │   └── system-prompt.json        - Generated system prompt
        └── responses/
```

## Building and Deployment

```bash
# Build Docker image
docker build -t fetchimages .

# Push to ECR and update Lambda
./deploy.sh
```

## Input/Output

### Input
```json
{
    "verificationId": "string",
    "s3References": {
        "processing_initialization": {
            "bucket": "string",
            "key": "string"
        }
    },
    "status": "string"
}
```

### Legacy Input (for backward compatibility)
```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationType": "LAYOUT_VS_CHECKING",
    "referenceImageUrl": "s3://kootoro-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e"
  }
}
```

### Output
```json
{
    "verificationId": "string",
    "s3References": {
        "processing_initialization": {
            "bucket": "string",
            "key": "string"
        },
        "images_metadata": {
            "bucket": "string",
            "key": "string"
        },
        "processing_layout-metadata": {
            "bucket": "string",
            "key": "string"
        },
        "processing_historical_context": {
            "bucket": "string",
            "key": "string"
        }
    },
    "status": "IMAGES_FETCHED",
    "summary": {
        "imagesFetched": true,
        "verificationType": "string",
        "layoutId": 123,
        "previousVerificationId": "string"
    }
}
```

## Error Handling

Errors are returned as structured messages with details for debugging:

```json
{
  "error": "ValidationError",
  "message": "Request validation failed: missing initialization reference",
  "details": "processing_initialization reference not found in s3References"
}
```

## Step Functions Integration

When used within the Step Function workflow, the Lambda expects S3 references from previous steps:

```json
{
  "FetchImages": {
    "Type": "Task",
    "Resource": "arn:aws:lambda:region:account:function:FetchImages",
    "Parameters": {
      "verificationId.$": "$.verificationId",
      "s3References.$": "$.s3References",
      "status.$": "$.status"
    },
    "ResultPath": "$"
  }
}
```