# InitializeFunction Lambda

This Lambda function serves as the entry point for the vending machine verification workflow. It validates inputs, verifies resources, and initializes a new verification process.

## 1. Code Structure

The code follows a modular architecture with a clear separation of concerns:

1. **cmd/initialize/main.go**: Entry point with Lambda handler
2. **internal/**: Core implementation
   - **config.go**: Configuration structures and defaults
   - **initialize_service.go**: Core business logic
   - **s3_client.go**: S3 client wrapper
   - **s3_validator.go**: S3 resource validation
   - **s3_url_parser.go**: S3 URL parsing and validation
   - **dynamodb_client.go**: DynamoDB client wrapper
   - **verification_repo.go**: Verification record operations
   - **layout_repo.go**: Layout metadata operations

## 2. Architecture

The codebase follows a clean architecture with direct AWS SDK interactions:

### Clean Architecture

- **Entry Point Layer**: `cmd/initialize/main.go` handles Lambda initialization
- **Domain Layer**: Configuration and models in the internal package
- **Service Layer**: `initialize_service.go` implements the core business logic
- **Repository Layer**: Direct implementations for verification and layout operations
- **Infrastructure Layer**: Client wrappers for S3 and DynamoDB
- **Shared Packages**: Common code shared across lambdas:
  - `shared/schema`: Common data models and validation
  - `shared/logger`: Standardized logging

### Direct AWS SDK Implementations

- Direct implementation of S3 and DynamoDB operations
- No shared utility packages for AWS services
- All S3 and DynamoDB operations are implemented directly using the AWS SDK
- Repository pattern for data access

## 3. Key Features

- **Input Validation**: Comprehensive validation of all request parameters
- **Resource Verification**: Parallel checks for images and layouts
- **Idempotency**: Conditional writes to prevent duplicate records
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Error Handling**: Custom error types and comprehensive error handling
- **Configuration Management**: Environment variables handled in service initialization
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