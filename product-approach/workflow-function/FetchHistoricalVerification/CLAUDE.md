# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based AWS Lambda function called `FetchHistoricalVerification` that is part of a vending machine verification workflow. This Lambda specifically retrieves historical verification data from DynamoDB for the "PREVIOUS_VS_CURRENT" verification type. It's used when comparing a current vending machine image against a previous verification.

## Commands

### Building the Docker Image

```bash
# Build Docker image
docker build -t fetchhistoricalverification .
```

### Deploy to AWS

```bash
# Deploy using the retry-docker-build.sh script
./retry-docker-build.sh
```

### Local Go Operations

```bash
# Run tests
go test ./...

# Build binary
go build -o FetchHistoricalVerification

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

### Local Testing

To test the Lambda function locally:

```bash
# Set necessary environment variables
export DYNAMODB_VERIFICATION_TABLE=VerificationResults
export CHECKING_BUCKET=kootoro-checking-bucket
export AWS_REGION=us-east-1

# Then run the Go program
go run *.go
```

## Architecture

This Lambda function is part of a Step Functions workflow for vending machine verification:

1. **Initialize Function** → **FetchHistoricalVerification** → **FetchImages**

The function receives a `VerificationContext` containing details about the current verification, and returns a `HistoricalContext` containing data from the previous verification.

### Data Flow

1. Lambda receives a verification context with the previous verification ID
2. It queries DynamoDB to retrieve the previous verification record
3. It calculates time elapsed since the previous verification
4. It returns structured historical verification data

### Key Components

- **main.go**: Lambda handler and initialization
- **service.go**: Core business logic for retrieving historical data
- **dynamodb.go**: DynamoDB client operations
- **types.go**: Data structures and models
- **validation.go**: Input validation logic
- **errors.go**: Standardized error handling
- **config.go**: Environment variable management

### Environment Variables

The Lambda function requires these environment variables:
- `AWS_REGION`: AWS region for the service
- `DYNAMODB_VERIFICATION_TABLE`: DynamoDB table with verification results
- `CHECKING_BUCKET`: S3 bucket containing checking images
- `LOG_LEVEL`: Logging level (INFO, DEBUG, ERROR)

## Error Handling

The function uses structured error responses:
- `ValidationError`: For input validation failures
- `ResourceNotFoundError`: When previous verification is not found
- `InternalError`: For internal server errors