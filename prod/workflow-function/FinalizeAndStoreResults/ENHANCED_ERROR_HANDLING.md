# Enhanced DynamoDB Error Handling

This document describes the enhanced DynamoDB error handling implemented in the FinalizeAndStoreResults Lambda function to resolve the generic "WRAPPED_ERROR" issues and provide detailed diagnostic information.

## Problem Solved

**Before**: Generic error messages like:
```json
{
  "errorType": "DYNAMODB",
  "errorCode": "WRAPPED_ERROR", 
  "errorMessage": "failed to put item to DynamoDB",
  "details": {
    "originalError": "ValidationException: One or more parameter values were invalid"
  }
}
```

**After**: Detailed, actionable error information:
```json
{
  "errorType": "DYNAMODB",
  "errorCode": "VALIDATION_EXCEPTION",
  "errorMessage": "DynamoDB validation error during PutItem",
  "details": {
    "operation": "PutItem",
    "table": "kootoro-dev-dynamodb-verification-results-f6d3xl",
    "awsError": "ValidationException: One or more parameter values were invalid",
    "errorType": "ValidationException",
    "troubleshooting": "Check item structure, required fields, and data types",
    "itemStructure": {
      "verificationId": "verif-20250605035206-97d7",
      "verificationAt": "2025-01-06T10:30:00Z",
      "verificationType": "layout",
      "currentStatus": "COMPLETED",
      "referenceImageUrl": "[URL_PROVIDED]",
      "checkingImageUrl": "[URL_PROVIDED]",
      "hasStatusHistory": false,
      "hasProcessingMetrics": false
    }
  },
  "retryable": false,
  "severity": "HIGH"
}
```

## Enhanced Features

### 1. AWS Error Type Detection
The system now detects and classifies specific AWS DynamoDB error types:

- **ValidationException**: Schema/data validation errors (not retryable)
- **ConditionalCheckFailedException**: Conditional write failures (not retryable)  
- **ProvisionedThroughputExceededException**: Capacity exceeded (retryable)
- **ResourceNotFoundException**: Table not found (not retryable)
- **InternalServerError**: AWS service issues (retryable)
- **ServiceUnavailable**: Temporary service issues (retryable)
- **ThrottlingException**: Rate limiting (retryable)

### 2. Data Validation
Pre-storage validation prevents common DynamoDB errors:

```go
// Validates required fields before attempting DynamoDB operations
func validateVerificationResultItem(item models.VerificationResultItem) error {
    // Checks: verificationId, verificationAt, verificationType, currentStatus
}
```

### 3. Sanitized Logging
Sensitive data is sanitized while preserving debugging information:

```go
// Example sanitized output
{
  "verificationId": "verif-20250605035206-97d7",
  "verificationType": "layout", 
  "referenceImageUrl": "[URL_PROVIDED]",  // Actual URL hidden
  "vendingMachineId": "[PROVIDED]",       // Indicates field is populated
  "layoutId": "[42]",                     // Shows actual value for IDs
  "hasStatusHistory": false               // Summary information
}
```

### 4. Troubleshooting Guidance
Each error type includes specific troubleshooting information:

- **ValidationException**: "Check item structure, required fields, and data types"
- **ResourceNotFoundException**: "Verify table name and AWS region configuration"
- **ThrottlingException**: "Implement exponential backoff retry strategy"

## Usage Examples

### Successful Operation
```
INFO storing_verification_result verificationId=verif-123 table=results operation=PutItem
INFO verification_result_stored verificationId=verif-123 status=COMPLETED
```

### Validation Error
```
ERROR validation_failed error="verificationId is required" field=verificationId table=results
```

### AWS DynamoDB Error
```
ERROR dynamodb_put_failed 
  error="DynamoDB validation error during PutItem"
  verificationId=verif-123
  table=results
  awsErrorCode=VALIDATION_EXCEPTION
  retryable=false
  details={
    "operation": "PutItem",
    "errorType": "ValidationException", 
    "troubleshooting": "Check item structure, required fields, and data types"
  }
```

## Integration

### Function Signatures
The enhanced functions require a logger parameter:

```go
// Before
func StoreVerificationResult(ctx, client, tableName, item) error

// After  
func StoreVerificationResult(ctx, client, tableName, item, log) error
```

### Error Handling
Enhanced errors provide detailed context:

```go
err := dynamodbhelper.StoreVerificationResult(ctx, client, table, item, log)
if err != nil {
    // Error already logged with full details
    // Enhanced error includes AWS error codes, troubleshooting info
    return err
}
```

## Benefits

1. **Root Cause Analysis**: Specific AWS error codes and messages
2. **Faster Debugging**: Sanitized data structure logging
3. **Operational Guidance**: Built-in troubleshooting recommendations  
4. **Retry Logic**: Clear indication of retryable vs non-retryable errors
5. **Security**: Sensitive data sanitization in logs
6. **Monitoring**: Structured error data for alerting and metrics

## Testing

Run the test suite to verify enhanced error handling:

```bash
go test ./internal/dynamodbhelper -v
```

Tests cover:
- AWS error type detection and classification
- Data validation for required fields
- Sanitization of sensitive logging data
- Error message formatting and context
