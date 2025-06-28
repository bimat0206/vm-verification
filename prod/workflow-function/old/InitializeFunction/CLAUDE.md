# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the `InitializeFunction` AWS Lambda for a vending machine verification system. It serves as the entry point for the verification workflow that analyzes product placement in vending machines using image processing.

The function handles two verification types:
1. `LAYOUT_VS_CHECKING` - Compares current vending machine state against a predefined layout
2. `PREVIOUS_VS_CURRENT` - Compares current vending machine state against a previous state

## Architecture

The codebase is organized into distinct layers:

1. **Entry Point Layer** (`main.go`)
   - Lambda handler and configuration setup

2. **Domain Layer** (`models.go`)
   - Data structures, constants, and helper methods

3. **Service Layer** (`service.go`)
   - Core verification workflow logic

4. **Infrastructure Layer**
   - `dbutils.go` - DynamoDB operations
   - `s3utils.go` - S3 operations and URL validation

5. **Utility Layer** (`logger.go`)
   - Structured JSON logging

6. **Dependency Management** (`dependencies.go`)
   - Dependency injection for AWS clients

## Key Features

- Input validation for different verification types
- Parallel resource verification using goroutines
- Structured JSON logging with correlation IDs
- Idempotent database operations
- S3 URL parsing and validation
- Support for multiple input formats (Lambda, API Gateway, Step Functions)

## Common Commands

### Building and Deploying

Build Docker image locally:
```bash
docker build -t initialize-function .
```

Deploy to AWS using the script:
```bash
./retry-docker-build.sh
```

The script handles:
1. AWS ECR login
2. Docker image build
3. Pushing image to ECR
4. Updating Lambda function code

### Environment Variables

The Lambda function requires the following environment variables:
- `DYNAMODB_LAYOUT_TABLE` - DynamoDB table for layout metadata
- `DYNAMODB_VERIFICATION_TABLE` - DynamoDB table for verification records
- `VERIFICATION_PREFIX` - Prefix for verification IDs (default: "verif-")
- `REFERENCE_BUCKET` - S3 bucket for reference images
- `CHECKING_BUCKET` - S3 bucket for checking images

## Development Guidelines

1. **Error Handling**
   - All errors should be properly wrapped and propagated with context
   - Use structured logging with correlation IDs for traceability

2. **Concurrency**
   - Resource verification operations should run in parallel using goroutines
   - Use error channels for collecting errors from concurrent operations

3. **Input Validation**
   - Validate all input parameters based on verification type
   - Check for required fields and ensure proper formats

4. **S3 Operations**
   - Always validate S3 URLs before processing
   - Use parallel checks to verify resource existence

5. **DynamoDB Operations**
   - Use conditional writes to ensure idempotency
   - Apply proper error handling for DynamoDB-specific errors