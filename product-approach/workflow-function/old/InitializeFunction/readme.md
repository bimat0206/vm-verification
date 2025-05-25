# InitializeFunction Lambda

This Lambda function serves as the entry point for the vending machine verification workflow. It validates inputs, verifies resources, and initializes a new verification process.

## 1. Code Structure

The code is divided into several files, each with a specific responsibility:

1. **main.go**: Entry point with Lambda handler and configuration
2. **models.go**: Data structures, constants, and helper functions 
3. **service.go**: Core business logic for the verification initialization process
4. **dbutils.go**: Adapter for the shared database operations package
5. **s3utils.go**: Adapter for the shared S3 operations package
6. **logger.go**: Adapter for the shared structured logging package
7. **dependencies.go**: Dependency injection and management

## 2. Architecture

The codebase follows a layered architecture with shared packages:

### Clear Separation of Concerns

- **Entry Point Layer**: `main.go` handles Lambda initialization and environment configuration
- **Domain Layer**: `models.go` defines data structures and constants
- **Service Layer**: `service.go` implements the core business logic
- **Infrastructure Layer**: Adapters for shared packages (`dbutils.go`, `s3utils.go`)
- **Utility Layer**: Adapter for shared logging functionality (`logger.go`)
- **Shared Packages**: Common code shared across lambdas:
  - `shared/schema`: Common data models and validation
  - `shared/logger`: Standardized logging
  - `shared/dbutils`: Database operations
  - `shared/s3utils`: S3 operations

### Dependency Injection

The code uses dependency injection for better testability:

- All dependencies are created in `main.go` and passed to the service
- The `Dependencies` struct in `dependencies.go` manages all external dependencies
- Each component only has access to the dependencies it needs

## 3. Key Features

- **Input Validation**: Comprehensive validation of all request parameters
- **Resource Verification**: Parallel checks for images and layouts
- **Idempotency**: Conditional writes to prevent duplicate records
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Error Handling**: Custom error types and comprehensive error handling
- **Shared Components**: Common code in shared packages
- **Configuration Management**: Environment variables centralized in main
- **Backward Compatibility**: Support for both legacy and new input formats

## 4. Verification Types

The function supports two verification types:

### LAYOUT_VS_CHECKING
- **Required fields**: verificationType, referenceImageUrl, checkingImageUrl, layoutId, layoutPrefix, notificationEnabled
- **Optional fields**: vendingMachineId

### PREVIOUS_VS_CURRENT
- **Required fields**: verificationType, referenceImageUrl, checkingImageUrl, notificationEnabled
- **Optional fields**: previousVerificationId, vendingMachineId

## 5. Input Schema

The function supports multiple input formats:

### Standardized Schema Format (with schemaVersion)

```json
{
  "schemaVersion": "1.0.0",
  "verificationContext": {
    "verificationType": "LAYOUT_VS_CHECKING" | "PREVIOUS_VS_CURRENT",
    "referenceImageUrl": "string",
    "checkingImageUrl": "string",
    "vendingMachineId": "string (optional)",
    "layoutId": "integer (required for LAYOUT_VS_CHECKING)",
    "layoutPrefix": "string (required for LAYOUT_VS_CHECKING)",
    "previousVerificationId": "string (optional for PREVIOUS_VS_CURRENT)",
    "notificationEnabled": "boolean"
  },
  "requestId": "string (optional)",
  "requestTimestamp": "string (optional)"
}
```

### Legacy Format (Step Function Input)

```json
{
  "verificationContext": {
    "verificationType": "LAYOUT_VS_CHECKING" | "PREVIOUS_VS_CURRENT",
    "referenceImageUrl": "string",
    "checkingImageUrl": "string",
    "vendingMachineId": "string (optional)",
    "layoutId": "integer (required for LAYOUT_VS_CHECKING)",
    "layoutPrefix": "string (required for LAYOUT_VS_CHECKING)",
    "previousVerificationId": "string (optional for PREVIOUS_VS_CURRENT)",
    "notificationEnabled": "boolean"
  },
  "requestId": "string (optional)",
  "requestTimestamp": "string (optional)"
}
```

### Legacy Format (Direct Lambda Input)

```json
{
  "verificationType": "LAYOUT_VS_CHECKING" | "PREVIOUS_VS_CURRENT",
  "referenceImageUrl": "string",
  "checkingImageUrl": "string",
  "vendingMachineId": "string (optional)",
  "layoutId": "integer (required for LAYOUT_VS_CHECKING)",
  "layoutPrefix": "string (required for LAYOUT_VS_CHECKING)",
  "previousVerificationId": "string (optional for PREVIOUS_VS_CURRENT)",
  "notificationEnabled": "boolean",
  "requestId": "string (optional)",
  "requestTimestamp": "string (optional)"
}
```

## 6. Environment Variables

The Lambda function requires the following environment variables:

- `DYNAMODB_LAYOUT_TABLE` - DynamoDB table for layout metadata
- `DYNAMODB_VERIFICATION_TABLE` - DynamoDB table for verification records
- `VERIFICATION_PREFIX` - Prefix for verification IDs (default: "verif-")
- `REFERENCE_BUCKET` - S3 bucket for reference images
- `CHECKING_BUCKET` - S3 bucket for checking images

## 7. Building and Deploying

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