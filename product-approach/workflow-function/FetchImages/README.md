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
Use case 1:
Response Format
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

Use case 2:
{
    "verificationContext": {
      "verificationId": "verif-2025042115302500",
      "verificationAt": "2025-04-21T15:30:25Z",
      "status": "IMAGES_FETCHED",
      "verificationType": "PREVIOUS_VS_CURRENT",
      "referenceImageUrl": "s3://kootoro-checking-bucket/2025-04-20/VM-3245/check_10-00-00.jpg",
      "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg"
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
        "checkingStatus": { // Describes state of CURRENT check image
        "A": "Current: 7 pink 'Mi Hảo Hảo' cup noodles visible. Status: No Change.",
        "B": "Current: 7 pink 'Mi Hảo Hảo' cup noodles visible. Status: No Change.",
        "C": "Current: 7 red/white 'Mi modern Lẩu thái' cup noodles visible. Status: No Change.",
        "D": "Current: 7 red/white 'Mi modern Lẩu thái' cup noodles visible. Status: No Change.",
        "E": "Current: 7 **GREEN 'Mi Cung Đình'** cup noodles visible. Status: Changed Product.", // Example change
        "F": "Current: 7 **GREEN 'Mi Cung Đình'** cup noodles visible. Status: Filled." // Example change
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
        },
      }
  }


### Step Function Integration
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
          "previousVerificationId.$": "States.ArrayGetItem(States.Array(null, $.verificationContext.previousVerificationId), States.StringEquals($.verificationContext.verificationType, 'PREVIOUS_VS_CURRENT'))",
        "notificationEnabled.$": "$.verificationContext.notificationEnabled"
      }
    },
    "ResultPath": "$"
  }
}
```

Note: The `previousVerificationId` field is conditionally included only when the verificationType is "PREVIOUS_VS_CURRENT". This prevents errors when processing "LAYOUT_VS_CHECKING" verification types.

Error Handling
{
  "error": "BadRequest",
  "message": "Invalid referenceImageUrl: not a valid S3 URI",
  "details": "Expected format: s3://bucket/key"
}


For the Initialize Lambda, the verificationContext object needs to be nested:

```json
// Step Function state definition for Initialize
{
  "InitializeLayoutChecking": {
    "Type": "Task",
    "Resource": "arn:aws:lambda:region:account:function:Initialize",
    "Parameters": {
      "verificationContext": {
        "verificationType.$": "$.verificationContext.verificationType",
        "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
        "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
        "vendingMachineId.$": "$.verificationContext.vendingMachineId",
        "layoutId.$": "$.verificationContext.layoutId",
        "layoutPrefix.$": "$.verificationContext.layoutPrefix",
        "notificationEnabled.$": "$.verificationContext.notificationEnabled"
      },
      "requestId.$": "$.requestId",
      "requestTimestamp.$": "$.requestTimestamp"
    }
  }
}
```
Output Example
{
  "verificationId": "abc123",
  "referenceImageUrl": "s3://mybucket/reference.jpg",
  "referenceImageMeta": {
    "contentType": "image/jpeg",
    "size": 123456,
    "lastModified": "2024-06-01T12:34:56Z",
    "etag": "\"abc123etag\"",
    "bucketOwner": "123456789012",
    "bucket": "mybucket",
    "key": "reference.jpg"
  },
  "checkingImageUrl": "s3://mybucket/checking.jpg",
  "checkingImageMeta": {
    "contentType": "image/jpeg",
    "size": 654321,
    "lastModified": "2024-06-01T12:35:56Z",
    "etag": "\"def456etag\"",
    "bucketOwner": "123456789012",
    "bucket": "mybucket",
    "key": "checking.jpg"
  }
  // Optionally: layoutMetadata, historicalContext
}
Error Handling
Errors are returned as structured messages with details for debugging.

Local Testing
You can test the handler locally using the AWS Lambda Go SDK or by simulating events.

Extending
Implement more detailed DynamoDB unmarshalling as needed.
Add more logging or tracing.
Add unit tests (mocking AWS SDK clients).
