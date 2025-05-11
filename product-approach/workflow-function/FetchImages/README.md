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

## Project Structure

fetchimages/
├── main.go               # Lambda handler and orchestration
├── models.go             # Data structures and validation
├── s3_utils.go           # S3 URI parsing and metadata retrieval
├── dynamodb_utils.go     # DynamoDB fetch helpers
├── parallel.go           # Parallel/concurrent fetch logic
├── config.go             # Environment variable management
├── logger.go             # Structured logging
├── go.mod


## Environment Variables

- `DYNAMODB_LAYOUT_TABLE` (default: `LayoutMetadata`) - Name of the DynamoDB table containing layout metadata
- `DYNAMODB_VERIFICATION_TABLE` (default: `VerificationResults`) - Name of the DynamoDB table containing verification results

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

- Implement more detailed DynamoDB unmarshalling as needed.
- Add more logging or tracing.
- Add unit tests (mocking AWS SDK clients).
