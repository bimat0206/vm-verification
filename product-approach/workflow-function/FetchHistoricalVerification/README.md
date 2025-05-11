# FetchHistoricalVerification Lambda Function

This AWS Lambda function is designed to retrieve historical verification data from DynamoDB for use in the "PREVIOUS_VS_CURRENT" verification workflow. It's specifically used when comparing a current vending machine image against a previous verification.

## Features

- **Retrieves historical verification data** from DynamoDB based on the previous verification ID
- **Calculates time elapsed** since the previous verification
- **Validates input parameters** to ensure proper workflow execution
- **Structured error handling** with detailed error messages
- **Configurable via environment variables**

## Project Structure

```
fetchhistoricalverification/
├── main.go               # Lambda handler and initialization
├── service.go            # Core business logic
├── dynamodb.go           # DynamoDB client and query operations
├── types.go              # Data structures and models
├── validation.go         # Input validation
├── errors.go             # Error handling
├── config.go             # Environment variable management
├── go.mod                # Go module definition
```

## Environment Variables

- `AWS_REGION` - AWS region for the service
- `DYNAMODB_VERIFICATION_TABLE` - Name of the DynamoDB table containing verification results
- `CHECKING_BUCKET` - Name of the S3 bucket containing checking images
- `LOG_LEVEL` - Logging level (INFO, DEBUG, ERROR)

## Input Format

The Lambda function expects input in the following format:

```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "verificationType": "PREVIOUS_VS_CURRENT",
    "referenceImageUrl": "s3://kootoro-checking-bucket/2025-04-20/VM-3245/check_10-00-00.jpg",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
    "previousVerificationId": "verif-2025042010000000",
    "vendingMachineId": "VM-3245"
  }
}
```

### Required Fields

- `verificationId` - Unique identifier for the current verification process
- `verificationType` - Must be "PREVIOUS_VS_CURRENT" for this function
- `referenceImageUrl` - S3 URI of the reference image (previous checking image)
- `checkingImageUrl` - S3 URI of the current checking image
- `previousVerificationId` - ID of the previous verification to retrieve
- `vendingMachineId` - ID of the vending machine being verified

## Output Format

The function returns the historical verification data in the following format:

```json
{
  "historicalContext": {
    "previousVerificationId": "verif-2025042010000000",
    "previousVerificationAt": "2025-04-20T10:00:00Z",
    "previousVerificationStatus": "CORRECT",
    "hoursSinceLastVerification": 29.5,
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
      "E": "Current: 7 GREEN 'Mi Cung Đình' cup noodles visible. Status: Changed Product.",
      "F": "Current: 7 GREEN 'Mi Cung Đình' cup noodles visible. Status: Filled."
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

In the Step Functions state machine, this Lambda is invoked after the InitializePreviousCurrent state and before the FetchImages state:

```json
"FetchHistoricalVerification": {
  "Type": "Task",
  "Resource": "${function_arns["fetch_historical_verification"]}",
  "Parameters": {
    "verificationId.$": "$.verificationContext.verificationId",
    "verificationType.$": "$.verificationContext.verificationType",
    "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
    "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
    "previousVerificationId.$": "$.verificationContext.previousVerificationId",
    "vendingMachineId.$": "$.verificationContext.vendingMachineId"
  },
  "ResultPath": "$.historicalContext",
  "Retry": [
    {
      "ErrorEquals": ["States.TaskFailed"],
      "IntervalSeconds": 2,
      "MaxAttempts": 3,
      "BackoffRate": 2.0
    }
  ],
  "Catch": [
    {
      "ErrorEquals": ["States.ALL"],
      "ResultPath": "$.error",
      "Next": "HandleHistoricalFetchError"
    }
  ],
  "Next": "FetchImages"
}
```

### Integration with FetchImages

The output of this function (historicalContext) is stored at the top level of the Step Functions execution state and is passed to the FetchImages function. The FetchImages function then uses this historical context to provide additional information for the verification process.

## Error Handling

Errors are returned as structured messages with details for debugging:

```json
{
  "code": "ResourceNotFoundError",
  "message": "Verification not found: verif-2025042010000000",
  "details": {
    "resourceType": "Verification",
    "resourceId": "verif-2025042010000000"
  }
}
```

Common error types:
- `ValidationError` - Input validation failed
- `ResourceNotFoundError` - Previous verification not found
- `InternalError` - Internal server error

## Local Testing

You can test the handler locally using the AWS Lambda Go SDK or by simulating events:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	// Set environment variables for local testing
	os.Setenv("DYNAMODB_VERIFICATION_TABLE", "VerificationResults")
	os.Setenv("CHECKING_BUCKET", "kootoro-checking-bucket")

	// Create test event
	event := map[string]interface{}{
		"verificationContext": map[string]interface{}{
			"verificationId":        "verif-2025042115302500",
			"verificationAt":        "2025-04-21T15:30:25Z",
			"verificationType":      "PREVIOUS_VS_CURRENT",
			"referenceImageUrl":     "s3://kootoro-checking-bucket/2025-04-20/VM-3245/check_10-00-00.jpg",
			"checkingImageUrl":      "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
			"previousVerificationId": "verif-2025042010000000",
			"vendingMachineId":      "VM-3245",
		},
	}

	// Convert to JSON
	eventJSON, _ := json.Marshal(event)
	fmt.Println(string(eventJSON))

	// Call handler
	result, err := handler(context.Background(), event)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(resultJSON))
}
```

## Extending

- Add more detailed error handling
- Implement caching for frequently accessed verifications
- Add metrics and tracing
- Add unit tests with mocked AWS SDK clients
