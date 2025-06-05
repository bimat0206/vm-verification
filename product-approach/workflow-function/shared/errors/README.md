# Enhanced Error Package

This package provides comprehensive error handling for workflow functions with specialized support for DynamoDB operations.

## Features

### 1. Enhanced Error Types
- **Comprehensive Error Categories**: Validation, API, DynamoDB, Transaction, Batch, etc.
- **Error Severity Levels**: Low, Medium, High, Critical
- **Error Categories**: Transient, Permanent, Client, Server, Network, Capacity, Auth, Validation
- **Retry Strategies**: None, Immediate, Linear, Exponential, Jittered

### 2. DynamoDB-Specific Error Handling
- **Specific Error Constructors** for all common DynamoDB errors
- **Automatic Error Analysis** from AWS error messages
- **Retry Logic** with appropriate strategies for each error type
- **Recovery Suggestions** and troubleshooting hints

### 3. Enhanced Error Context
- **Stack Traces** for debugging
- **Correlation IDs** for distributed tracing
- **Component and Operation** tracking
- **Table and Index** information for DynamoDB errors
- **Retry Count** and limits tracking

## Usage Examples

### Basic DynamoDB Error Creation

```go
// Create a validation error
err := errors.NewDynamoDBValidationError("PutItem", "UserTable", "Missing required field: userId")

// Create a throughput exceeded error
err := errors.NewDynamoDBThroughputExceededError("Query", "UserTable")

// Create a conditional check failed error
err := errors.NewDynamoDBConditionalCheckFailedError("PutItem", "UserTable", "item_exists = false")
```

### Enhanced Error with Context

```go
err := errors.NewDynamoDBError("PutItem", "UserTable", originalError).
    WithCategory(errors.CategoryTransient).
    WithRetryStrategy(errors.RetryExponential).
    WithComponent("UserService").
    WithOperation("CreateUser").
    WithCorrelationID("req-123-456").
    WithVerificationID("verify-789").
    WithSuggestions("Check item structure", "Verify required fields").
    WithRecoveryHints("Retry with exponential backoff").
    SetMaxRetries(3)
```

### Automatic Error Analysis

```go
// Analyze any DynamoDB error and get enhanced error information
enhancedErr := errors.AnalyzeDynamoDBError("PutItem", "UserTable", awsError)

// Check if error is retryable
if errors.IsDynamoDBRetryableError(enhancedErr) {
    strategy := errors.GetDynamoDBRetryStrategy(enhancedErr)
    // Implement retry logic based on strategy
}
```

### Retry Logic

```go
err := errors.NewDynamoDBThroughputExceededError("Query", "UserTable")

for !err.IsRetryLimitExceeded() {
    // Attempt operation
    if operationErr := performOperation(); operationErr != nil {
        err.IncrementRetryCount()
        
        // Calculate delay based on retry strategy
        delay := err.GetRetryDelay(time.Second)
        time.Sleep(delay)
        continue
    }
    break
}
```

### Error Classification

```go
// Check error characteristics
if errors.IsTransientError(err) {
    // Handle temporary errors
}

if errors.IsPermanentError(err) {
    // Handle permanent errors that won't resolve with retry
}

// Get suggestions and recovery hints
suggestions := errors.GetErrorSuggestions(err)
hints := errors.GetRecoveryHints(err)
```

### Error Monitoring and Metrics

```go
// Get metrics for monitoring
metrics := errors.GetErrorMetrics(err)

// Log metrics to your monitoring system
logger.Info("Error occurred", metrics)

// Aggregate multiple errors for analysis
errorList := []error{err1, err2, err3}
summary := errors.AggregateErrors(errorList)

fmt.Printf("Total errors: %d, Retryable: %d, Critical: %d\n", 
    summary.TotalErrors, summary.RetryableCount, summary.CriticalCount)
```

## DynamoDB Error Types

### Validation Errors
- **ValidationException**: Invalid item structure or data types
- **ConditionalCheckFailedException**: Condition expression failed
- **ItemCollectionSizeLimitExceededException**: Partition too large

### Capacity Errors
- **ProvisionedThroughputExceededException**: Read/write capacity exceeded
- **ThrottlingException**: Request rate too high
- **LimitExceededException**: Service limits exceeded

### Resource Errors
- **ResourceNotFoundException**: Table or index not found
- **TableNotFoundException**: Specific table not found
- **IndexNotFoundException**: GSI/LSI not found

### Server Errors
- **InternalServerError**: AWS service issue
- **ServiceUnavailableException**: Service temporarily unavailable

### Transaction Errors
- **TransactionConflictException**: Concurrent transaction conflict
- **TransactionCanceledException**: Transaction cancelled due to condition failure
- **TransactionInProgressException**: Another transaction in progress

## Best Practices

### 1. Use Specific Error Constructors
```go
// Good: Use specific constructor
err := errors.NewDynamoDBValidationError("PutItem", "UserTable", "Invalid data")

// Avoid: Generic error
err := errors.NewDynamoDBError("PutItem", "UserTable", genericError)
```

### 2. Add Context Information
```go
err := errors.NewDynamoDBThroughputExceededError("Query", "UserTable").
    WithCorrelationID(requestID).
    WithComponent("UserService").
    WithOperation("GetUsersByStatus")
```

### 3. Implement Proper Retry Logic
```go
if errors.IsDynamoDBRetryableError(err) {
    strategy := errors.GetDynamoDBRetryStrategy(err)
    // Implement retry with appropriate strategy
}
```

### 4. Use Error Analysis for Unknown Errors
```go
// When you receive an error from AWS SDK
enhancedErr := errors.AnalyzeDynamoDBError(operation, tableName, awsError)
```

### 5. Monitor Error Patterns
```go
metrics := errors.GetErrorMetrics(err)
// Send to your monitoring system (CloudWatch, DataDog, etc.)
```

## Error Categories and Retry Strategies

| Error Type | Category | Retry Strategy | Max Retries |
|------------|----------|----------------|-------------|
| ValidationException | Validation | None | 0 |
| ConditionalCheckFailedException | Client | None | 0 |
| ProvisionedThroughputExceededException | Capacity | Exponential | 5 |
| ThrottlingException | Capacity | Jittered | 5 |
| InternalServerError | Server | Exponential | 3 |
| TransactionConflictException | Transient | Exponential | 3 |
| ResourceNotFoundException | Client | None | 0 |

## Integration with Existing Code

The enhanced error package is backward compatible with existing error handling code. You can gradually migrate to use the new features:

```go
// Existing code continues to work
if errors.IsRetryable(err) {
    // retry logic
}

// Enhanced features available when needed
if workflowErr, ok := err.(*errors.WorkflowError); ok {
    if workflowErr.Category == errors.CategoryCapacity {
        // Handle capacity issues specifically
    }
}
```
