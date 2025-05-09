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

- `LAYOUT_TABLE_NAME` (default: `LayoutMetadata`)
- `VERIFICATION_TABLE_NAME` (default: `VerificationResults`)

## Input Examples

### Direct Lambda Invocation
```json
{
  "verificationId": "abc123",
  "verificationType": "LAYOUT_VS_CHECKING",
  "referenceImageUrl": "s3://mybucket/reference.jpg",
  "checkingImageUrl": "s3://mybucket/checking.jpg",
  "layoutId": 42,
  "layoutPrefix": "vm-001"
}
```

### Step Function Integration
When used within the Step Function workflow, the Lambda expects parameters to be extracted from the verificationContext object:

```json
// Step Function state definition for FetchImages
{
  "FetchImages": {
    "Type": "Task",
    "Resource": "arn:aws:lambda:region:account:function:FetchImages",
    "Parameters": {
      "verificationId.$": "$.verificationContext.verificationId",
      "verificationType.$": "$.verificationContext.verificationType",
      "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
      "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
      "layoutId.$": "$.verificationContext.layoutId",
      "layoutPrefix.$": "$.verificationContext.layoutPrefix",
      "vendingMachineId.$": "$.verificationContext.vendingMachineId"
    }
  }
}
```

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
    "etag": "\"abc123etag\""
  },
  "checkingImageUrl": "s3://mybucket/checking.jpg",
  "checkingImageMeta": {
    "contentType": "image/jpeg",
    "size": 654321,
    "lastModified": "2024-06-01T12:35:56Z",
    "etag": "\"def456etag\""
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
