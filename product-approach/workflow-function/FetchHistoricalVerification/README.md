# FetchHistoricalVerification Lambda Function

This Lambda function is part of the vending machine verification workflow. It retrieves historical verification data for "PREVIOUS_VS_CURRENT" verification type. It's used when comparing a current vending machine image against a previous verification.

## Overview

The FetchHistoricalVerification Lambda:
1. Receives a verification context with verification ID and reference image URL
2. Queries DynamoDB to find the most recent verification where the current reference image was used as a checking image
3. Calculates the time elapsed since the previous verification
4. Returns a structured historical context to be used in the subsequent verification steps

## Architecture

This function is part of a Step Functions workflow:

**Initialize Function** → **FetchHistoricalVerification** → **FetchImages**

## Components

- **main.go**: Lambda handler and initialization
- **service.go**: Core business logic for fetching historical verification data
- **dynamodb.go**: DynamoDB client operations wrapper
- **types.go**: Data structures specific to this function
- **errors.go**: Standardized error handling
- **dependencies.go**: Dependency initialization and configuration
- **config.go**: Environment variable management

## Shared Packages

This Lambda uses the following shared packages:
- **schema**: Core data structures and constants
- **logger**: Structured JSON logging
- **dbutils**: DynamoDB operations

## Building and Deploying

### Local Development

```bash
# Build the binary
go build -o FetchHistoricalVerification

# Run tests
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

### Docker Build

```bash
# Build Docker image
docker build -t fetchhistoricalverification .
```

### AWS Deployment

```bash
# Deploy using the retry-docker-build.sh script
./retry-docker-build.sh
```

## Environment Variables

- `AWS_REGION`: AWS region for the service
- `DYNAMODB_VERIFICATION_TABLE`: DynamoDB table containing verification results
- `CHECKING_BUCKET`: S3 bucket containing checking images
- `LOG_LEVEL`: Logging level (INFO, DEBUG, ERROR)

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
      "B": "Current: 7 pink 'Mi Hảo Hảo' cup noodles visible. Status: No Change."
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
      "verificationOutcome": "Discrepancies Detected"
    }
  }
}
```

## Testing

Set necessary environment variables and run the Lambda locally:

```bash
export DYNAMODB_VERIFICATION_TABLE=VerificationResults
export CHECKING_BUCKET=kootoro-checking-bucket
export AWS_REGION=us-east-1

go run *.go
```

## Error Handling

The function uses structured error responses:
- `ValidationError`: For input validation failures
- `ResourceNotFoundError`: When previous verification is not found
- `InternalError`: For internal server errors

## Recent Changes

See [CHANGELOG.md](./CHANGELOG.md) for details about recent updates and changes to this Lambda function.