# Shared Package Components for Lambda Functions

This directory contains shared packages that provide standardized functionality across all Lambda functions in the vending machine verification workflow.

## Overview

The shared components provide:
- Consistent data structures and validation
- Standardized logging
- Common AWS service utilities (S3, DynamoDB)
- Error handling patterns
- Helper functions for common operations

## Packages

### 1. Schema (`/schema`)

Core data structures and status constants that flow through the Step Functions workflow.

```go
import "workflow-function/shared/schema"

// Use schema components
context := &schema.VerificationContext{
    VerificationId: "some-id",
    Status: schema.StatusVerificationRequested,
    // ...
}
```

### 2. Logger (`/logger`)

Structured JSON logging with correlation ID tracking and log levels.

```go
import "workflow-function/shared/logger"

// Create a logger instance
log := logger.New("verification-service", "InitializeFunction")

// Use structured logging
log.Info("Processing request", map[string]interface{}{
    "verificationId": id,
    "timestamp": time.Now().String(),
})

// Add correlation ID for tracing
logWithId := log.WithCorrelationId(verificationId)
logWithId.Info("Operation completed", nil)
```

### 3. S3 Utilities (`/s3utils`)

Common operations for S3 resources, especially image handling.

```go
import "workflow-function/shared/s3utils"

// Create S3Utils instance
s3Utils := s3utils.New(s3Client, log)

// Check if image exists
exists, err := s3Utils.ValidateImageExists(ctx, "s3://bucket/key.png", 10*1024*1024)

// Parse S3 URLs
refUrl, checkUrl, err := s3Utils.ParseS3URLs(referenceImageUrl, checkingImageUrl)
```

### 4. DynamoDB Utilities (`/dbutils`)

DynamoDB operations for verification data and conversation history.

```go
import "workflow-function/shared/dbutils"

// Create config
config := dbutils.Config{
    VerificationTable: "verifications-table",
    LayoutTable: "layouts-table",
    ConversationTable: "conversations-table",
}

// Create DynamoDBUtils instance
dbUtils := dbutils.New(dynamoDBClient, log, config)

// Store verification record
err := dbUtils.StoreVerificationRecord(ctx, verificationContext)

// Retrieve verification record
context, err := dbUtils.GetVerificationRecord(ctx, verificationId)
```

## Usage in Lambda Functions

1. Update `go.mod` with local import paths:

```
module workflow-function/YourFunctionName

go 1.21

require (
    // ... other dependencies
    workflow-function/shared/schema v0.0.0
    workflow-function/shared/logger v0.0.0
    workflow-function/shared/s3utils v0.0.0
    workflow-function/shared/dbutils v0.0.0
)

replace workflow-function/shared/schema => ../shared/schema
replace workflow-function/shared/logger => ../shared/logger
replace workflow-function/shared/s3utils => ../shared/s3utils
replace workflow-function/shared/dbutils => ../shared/dbutils
```

2. Update Dockerfile to include shared packages:

```dockerfile
# Copy shared packages
COPY workflow-function/shared/ /app/workflow-function/shared/
COPY workflow-function/YourFunctionName/ /app/workflow-function/YourFunctionName/
```

3. Initialize components in your Lambda handler:

```go
// Create logger
log := logger.New("verification-service", "YourFunctionName")

// Set up AWS clients (from dependencies)
s3Client := dependencies.GetS3Client()
dynamoDBClient := dependencies.GetDynamoDBClient()

// Create utils
s3Utils := s3utils.New(s3Client, log)
dbConfig := dbutils.Config{
    VerificationTable: os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
    // Other config...
}
dbUtils := dbutils.New(dynamoDBClient, log, dbConfig)

// Use shared components in your function
```

## Docker Build Notes

When building Lambda functions that use these shared packages, make sure to include all shared directories in your Docker build context. The function's `retry-docker-build.sh` script has been updated to properly handle this setup.

## Adding New Shared Components

When adding new shared functionality:

1. Create a new directory under `/shared`
2. Define the module with `go.mod`
3. Implement the core functionality
4. Update the README to document usage
5. Add appropriate tests
6. Add replace directives to Lambda function `go.mod` files