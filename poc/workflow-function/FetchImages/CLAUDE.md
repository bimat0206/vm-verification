# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FetchImages is an AWS Lambda function in a vending machine verification system. It fetches image metadata from S3 and retrieves relevant context from DynamoDB for verification workflows. The function is part of a Step Functions workflow for product layout verification in vending machines.

## Key Components

- **main.go**: Lambda handler and core logic
- **models.go**: Data structures for verification contexts and responses
- **s3.go**: S3 URI parsing and metadata retrieval
- **dynamodb.go**: DynamoDB interactions 
- **parallel.go**: Concurrent execution of data fetching operations
- **config.go**: Environment variable management
- **logger.go**: Structured logging utilities

## Development Commands

### Building and Deploying

**Build Docker Image and Deploy to Lambda**
```bash
./retry-docker-build.sh
```

This script handles:
1. Logging into AWS ECR
2. Building the Docker image
3. Pushing to the ECR repository
4. Updating the Lambda function with the new image

### Local Testing

You can use the AWS Lambda Go runtime API for local testing:

```bash
go run .
```

## Architecture Notes

### Data Flow

This Lambda function is part of a Step Function workflow with the following flow:
1. Initialize Function
2. FetchHistoricalVerification (for PREVIOUS_VS_CURRENT verification type)
3. FetchImages (this function)
4. Further processing steps...

### Input/Output

- **Input**: A request containing verification context with S3 image URLs
- **Output**: Response with S3 metadata and context from DynamoDB (either layout metadata or historical verification data)

### Verification Types

1. **LAYOUT_VS_CHECKING**: Compares layout reference image with current checking image
2. **PREVIOUS_VS_CURRENT**: Compares previous checking image with current checking image

### Environment Variables

- `DYNAMODB_LAYOUT_TABLE` (default: `LayoutMetadata`)
- `DYNAMODB_VERIFICATION_TABLE` (default: `VerificationResults`)

## AWS Resource Dependencies

- **AWS Lambda**: Execution environment
- **Amazon S3**: For storing and retrieving images
- **Amazon DynamoDB**: For storing layout metadata and verification results
- **AWS Step Functions**: Orchestrates the verification workflow