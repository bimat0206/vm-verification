# FetchImages Lambda Function

This AWS Lambda function is designed to fetch image metadata (not the image bytes) from Amazon S3 and relevant context from DynamoDB, for use cases such as image verification workflows. It strictly uses S3 URIs and does not handle or transmit base64-encoded image data.

## Features

- **Validates input** for required parameters and S3 URI format.
- **Fetches S3 metadata** (content type, size, last modified, ETag) for reference and checking images.
- **Retrieves layout metadata** or previous verification context from DynamoDB, depending on the verification type.
- **Runs S3 and DynamoDB queries in parallel** for performance.
- **Structured error handling and logging** for easy debugging.
- **Configurable** via environment variables.
- **Flexible input handling** supporting both direct Step Function invocations and Function URL requests.
- **Dynamic response size management** with hybrid storage options for large images.

## Recent Updates

### Version 3.0.0 (2025-05-16)
- Refactored to remove dependencies on shared packages (s3utils and dbutils)
- Implemented direct AWS SDK interactions for S3 and DynamoDB operations
- Split codebase into multiple specialized files for better maintainability

See the [CHANGELOG.md](CHANGELOG.md) for a complete history of changes.

## Project Structure

```
FetchImages/
├── main.go               # Lambda handler and orchestration
├── models.go             # Function-specific data structures and validation
├── direct_operations.go  # Direct S3 and DynamoDB operations using AWS SDK
├── s3url.go              # S3 URL parsing and validation
├── db_models.go          # Database models and helpers
├── dependencies.go       # Dependency management and configuration
├── parallel.go           # Parallel/concurrent fetch logic
├── response_tracker.go   # Response size tracking for Lambda limits
├── storage_validation.go # Storage integrity validation
├── storage_stats.go      # Storage statistics functions
├── go.mod
└── Dockerfile
```

## Environment Variables

- `DYNAMODB_LAYOUT_TABLE` (default: `LayoutMetadata`) - Name of the DynamoDB table containing layout metadata
- `DYNAMODB_VERIFICATION_TABLE` (default: `VerificationResults`) - Name of the DynamoDB table containing verification results
- `MAX_IMAGE_SIZE` (default: `104857600` - 100MB) - Maximum image size in bytes
- `MAX_INLINE_BASE64_SIZE` (default: `2097152` - 2MB) - Maximum size for inline Base64 storage
- `TEMP_BASE64_BUCKET` - S3 bucket for temporary Base64 storage (for large images)

## Building and Deployment

The function can be built using the included `build.sh` script:

```bash
# Navigate to the FetchImages directory
cd /path/to/workflow-function/FetchImages

# Make the script executable if needed
chmod +x build.sh

# Edit the script to update AWS_REGION and AWS_ACCOUNT_ID variables
# Then run the script
./build.sh
```

### How the Build Process Works

The build script:
1. Creates a temporary directory with the correct module structure
2. Copies the function code with the right paths
3. Sets up proper Go module references
4. Builds the Docker image
5. Pushes the image to ECR
6. Updates the Lambda function

### Automated Build Script

A build script `build.sh` has been provided that handles all the steps automatically:

```bash
# Make the script executable if needed
chmod +x build.sh

# Update the AWS_REGION and AWS_ACCOUNT_ID in the script
# Then run it from the FetchImages directory
./build.sh
```

The script:
1. Creates a temporary build environment
2. Builds the Docker image
3. Logs in to ECR
4. Tags and pushes the image
5. Updates the Lambda function
6. Cleans up temporary files

You'll need to update the `AWS_REGION` and `AWS_ACCOUNT_ID` variables in the script before running it.

## Input Examples

### Direct Lambda Invocation
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

## Response Format Examples

### Use Case 1: LAYOUT_VS_CHECKING
```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z", 
    "status": "IMAGES_FETCHED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "referenceImageUrl": "s3://kootoro-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg"
  },
  "images": {
    "referenceImageMeta": {
      "contentType": "image/png",
      "size": 245678,
      "lastModified": "2025-04-21T14:25:15Z",
      "etag": "\"abc123def456\"",
      "bucketOwner": "879654127886",
      "bucket": "kootoro-reference-bucket",
      "key": "processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png"
    },
    "checkingImageMeta": {
      "contentType": "image/jpeg",
      "size": 356789,
      "lastModified": "2025-04-21T15:30:20Z",
      "etag": "\"def456ghi789\"",
      "bucketOwner": "879654127886",
      "bucket": "kootoro-checking-bucket",
      "key": "2025-04-21/VM-3245/check_15-30-25.jpg"
    }
  },
  "layoutMetadata": {
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e",
    "vendingMachineId": "VM-23591",
    "createdAt": "2025-05-06T05:00:03Z",
    "updatedAt": "2025-05-06T05:00:03Z",
    "location": "Office Building A, Floor 3",
    "machineStructure": {
      "rowCount": 6,
      "columnsPerRow": 10,
      "rowOrder": ["A", "B", "C", "D", "E", "F"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    },
    "productPositionMap": {
      "A01": {
        "productId": 3486,
        "productName": "Mì Hảo Hảo"
      }
    }
  }
}
```

### Use Case 2: PREVIOUS_VS_CURRENT
```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "IMAGES_FETCHED",
    "verificationType": "PREVIOUS_VS_CURRENT",
    "referenceImageUrl": "s3://kootoro-checking-bucket/2025-04-20/VM-3245/check_10-00-00.jpg",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
    "previousVerificationId": "verif-2025042010000000"
  },
  "images": {
    "referenceImageMeta": {
      "contentType": "image/png",
      "size": 245678,
      "lastModified": "2025-04-21T14:25:15Z",
      "etag": "\"abc123def456\"",
      "bucketOwner": "879654127886",
      "bucket": "kootoro-checking-bucket",
      "key": "2025-04-20/VM-3245/check_10-00-00.jpg"
    },
    "checkingImageMeta": {
      "contentType": "image/jpeg",
      "size": 356789,
      "lastModified": "2025-04-21T15:30:20Z",
      "etag": "\"def456ghi789\"",
      "bucketOwner": "879654127886",
      "bucket": "kootoro-checking-bucket",
      "key": "2025-04-21/VM-3245/check_15-30-25.jpg"
    }
  },
  "historicalContext": {
    "previousVerificationId": "verif-2025042010000000",
    "previousVerificationAt": "2025-04-20T10:00:00Z",
    "previousVerificationStatus": "CORRECT",
    "machineStructure": {
      "rowCount": 6,
      "columnsPerRow": 10,
      "rowOrder": ["A", "B", "C", "D", "E", "F"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    },
    "checkingStatus": {
      "A": "Current: 7 pink 'Mi Hảo Hảo' cup noodles visible. Status: No Change.",
      "B": "Current: 7 pink 'Mi Hảo Hảo' cup noodles visible. Status: No Change.",
      "C": "Current: 7 red/white 'Mi modern Lẩu thái' cup noodles visible. Status: No Change.",
      "D": "Current: 7 red/white 'Mi modern Lẩu thái' cup noodles visible. Status: No Change.",
      "E": "Current: 7 **GREEN 'Mi Cung Đình'** cup noodles visible. Status: Changed Product.",
      "F": "Current: 7 **GREEN 'Mi Cung Đình'** cup noodles visible. Status: Filled."
    },
    "verificationSummary": {
      "totalPositionsChecked": 42,
      "correctPositions": 28,
      "discrepantPositions": 14,
      "missingProducts": 7,
      "incorrectProductTypes": 7,
      "unexpectedProducts": 0,
      "emptyPositionsCount": 7,
      "overallAccuracy": 66.7,
      "overallConfidence": 97,
      "verificationStatus": "INCORRECT",
      "verificationOutcome": "Discrepancies Detected - Row E contains incorrect product types and Row F is completely empty"
    }
  }
}
```

## Step Functions Integration

### Important Note on Response Format

The response always includes the following top-level fields to ensure compatibility with Step Functions:

- `verificationContext`: Contains the updated verification context with status
- `images`: Contains metadata for both reference and checking images
- `layoutMetadata`: Contains layout information (may be empty for PREVIOUS_VS_CURRENT)
- `historicalContext`: Contains historical verification data (always present but may be empty for LAYOUT_VS_CHECKING)

This consistent response structure ensures that Step Functions JSONPath expressions can reliably access all fields.

### Data Flow Diagram

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│ Initialize      │────▶│ FetchHistorical │────▶│ FetchImages     │
│ Function        │     │ Verification    │     │ Function        │
│                 │     │                 │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│verificationContext    │historicalContext│     │images           │
│- verificationId │     │- previousVerif..│     │- referenceMeta  │
│- verificationType     │- machineStructure     │- checkingMeta   │
│- referenceImageUrl    │- checkingStatus │     │                 │
│- checkingImageUrl     │- verifSummary   │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Step Function State Definition

When used within the Step Function workflow, the Lambda expects parameters to be extracted from the verificationContext object:

```json
// Step Function state definition for FetchImages
{
  "FetchImages": {
    "Type": "Task",
    "Resource": "arn:aws:lambda:region:account:function:FetchImages",
    "Parameters": {
      "verificationContext": {
        "verificationId.$": "$.verificationContext.verificationId",
        "verificationAt.$": "$.verificationContext.verificationAt",
        "status.$": "$.verificationContext.status",
        "verificationType.$": "$.verificationContext.verificationType",
        "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
        "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
        "layoutId.$": "$.verificationContext.layoutId",
        "layoutPrefix.$": "$.verificationContext.layoutPrefix",
        "vendingMachineId.$": "$.verificationContext.vendingMachineId",
        "previousVerificationId.$": "States.ArrayGetItem(States.Array('', $.verificationContext.previousVerificationId), States.StringEquals($.verificationContext.verificationType, 'PREVIOUS_VS_CURRENT'))",
        "notificationEnabled.$": "$.verificationContext.notificationEnabled"
      }
    },
    "ResultPath": "$"
  }
}
```

### Parameter Handling

The `previousVerificationId` field is conditionally included only when the verificationType is "PREVIOUS_VS_CURRENT". This prevents errors when processing "LAYOUT_VS_CHECKING" verification types.

The conditional logic uses AWS Step Functions intrinsic functions:
- `States.ArrayGetItem(States.Array(null, $.verificationContext.previousVerificationId), States.StringEquals($.verificationContext.verificationType, 'PREVIOUS_VS_CURRENT'))`

This creates an array with `[null, previousVerificationId]` and selects index 1 (the ID) if the condition is true, otherwise index 0 (null).

### FetchHistoricalVerification Integration

For the PREVIOUS_VS_CURRENT verification type, the workflow first calls the FetchHistoricalVerification Lambda, which:

1. Takes the previousVerificationId from the verificationContext
2. Retrieves the historical verification data from DynamoDB
3. Returns a historicalContext object that is passed to the FetchImages function

The FetchHistoricalVerification function expects:
```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationType": "PREVIOUS_VS_CURRENT",
    "referenceImageUrl": "s3://kootoro-checking-bucket/2025-04-20/VM-3245/check_10-00-00.jpg",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
    "previousVerificationId": "verif-2025042010000000",
    "vendingMachineId": "VM-3245"
  }
}
```

## Error Handling

Errors are returned as structured messages with details for debugging:

```json
{
  "error": "BadRequest",
  "message": "Invalid referenceImageUrl: not a valid S3 URI",
  "details": "Expected format: s3://bucket/key"
}
```

## Local Testing

You can test the handler locally using the AWS Lambda Go SDK or by simulating events.

## Extending

- Add more logging or tracing.
- Add unit tests (mocking AWS SDK clients).
- Implement additional validation for input parameters.
