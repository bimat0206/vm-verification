# InitializeFunction Lambda

This Lambda function serves as the entry point for the vending machine verification workflow. It initializes the S3 state structure, validates input parameters, verifies resources, and creates the foundation for the verification process.

## 1. Code Structure

The code follows a modular architecture with a clear separation of concerns:

1. **cmd/initialize/main.go**: Entry point with Lambda handler
2. **internal/**: Core implementation
   - **config.go**: Configuration structures and defaults
   - **initialize_service.go**: Core business logic
   - **s3_client.go**: S3 client wrapper
   - **s3_validator.go**: S3 resource validation
   - **s3_url_parser.go**: S3 URL parsing and validation
   - **s3_state_manager.go**: S3 state management wrapper
   - **dynamodb_client.go**: DynamoDB client wrapper
   - **verification_repo.go**: Verification record operations
   - **layout_repo.go**: Layout metadata operations

## 2. Architecture

The codebase follows the S3 State Management pattern with reference-based workflow:

### S3 State Management Architecture

- **Entry Point Layer**: `cmd/initialize/main.go` handles Lambda initialization
- **Domain Layer**: Configuration and models in the internal package
- **Service Layer**: `initialize_service.go` implements the core business logic
- **State Management Layer**: S3 state structure creation and reference management
- **Repository Layer**: Minimal DynamoDB records with S3 references
- **Infrastructure Layer**: Client wrappers for S3 and DynamoDB
- **Shared Packages**: Common code shared across lambdas:
  - `shared/schema`: Common data models and validation
  - `shared/logger`: Standardized logging
  - `shared/s3state`: S3 state management utilities

### S3 State Structure

The function creates the following S3 state structure:

```
s3://state-management-bucket/
├── {YYYY}/{MM}/{DD}/
    └── {verificationId}/
        ├── processing/
        │   ├── initialization.json       - Initial verification state
        │   ├── layout-metadata.json      # For UC1
        │   └── historical-context.json   # For UC2
        ├── images/
        │   ├── metadata.json             # Image metadata only
        │   ├── reference-base64.base64   # reference base64 encode
        │   └── checking-base64.base64    # checking base64 encode
        ├── prompts/
        │   └── system-prompt.json        - Generated system prompt
        └── responses/
```

## 3. Key Features

- **S3 State Initialization**: Creates the foundational folder structure for verification
- **Input Validation**: Comprehensive validation of all request parameters
- **Resource Verification**: Parallel checks for images and layouts
- **Reference-Based Workflow**: Returns S3 references instead of full data payloads
- **Minimal DynamoDB Storage**: Stores only essential metadata with S3 references
- **Idempotency**: Conditional writes to prevent duplicate records
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Error Handling**: Custom error types and comprehensive error handling with error storage in S3
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
  "schemaVersion": "2.0.0",
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

## 6. Output Format

The function returns an S3 state envelope with references:

```json
{
  "verificationId": "verif-2025042115302500",
  "s3References": {
    "processing_initialization": {
      "bucket": "state-management-bucket",
      "key": "verif-2025042115302500/processing/initialization.json",
      "size": 1247
    }
  },
  "status": "VERIFICATION_INITIALIZED",
  "summary": {
    "verificationType": "LAYOUT_VS_CHECKING",
    "resourcesValidated": ["referenceImage", "checkingImage", "layoutMetadata"],
    "contextEstablished": true,
    "stateStructureCreated": true
  }
}
```

## 7. Environment Variables

The Lambda function requires the following environment variables:

- `DYNAMODB_LAYOUT_TABLE` - DynamoDB table for layout metadata
- `DYNAMODB_VERIFICATION_TABLE` - DynamoDB table for verification records
- `VERIFICATION_PREFIX` - Prefix for verification IDs (default: "verif-")
- `REFERENCE_BUCKET` - S3 bucket for reference images
- `CHECKING_BUCKET` - S3 bucket for checking images
- `STATE_BUCKET` - S3 bucket for state management storage

## 8. Building and Deploying

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