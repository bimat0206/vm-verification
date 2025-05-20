# Error Handling System

This package provides a comprehensive error handling system for the ProcessTurn1Response Lambda function, focusing on consistent error creation, propagation, and reporting.

## Key Components

- **FunctionError**: A structured error type that includes operation, category, code, message, and severity
- **ErrorList**: A utility for collecting and managing multiple errors
- **Error Categories**: Input, Process, State, System
- **Error Severities**: Debug, Info, Warning, Error, Critical
- **Error Converters**: Utilities for converting between error types

## Usage Guidelines

### Creating Errors

Use the provided factory functions to create errors with appropriate categories:

```go
// For input validation errors
err := errors.InputError("ValidateRequest", "missing required field", nil)

// For processing errors
err := errors.ProcessingError("ProcessResponse", "failed to extract data", innerErr)

// For state management errors
err := errors.StateError("LoadState", errors.ErrStateLoadFailed, "failed to load state", innerErr)

// For system errors
err := errors.SystemError("Initialize", "dependency initialization failed", innerErr)

// For errors with additional context
details := map[string]interface{}{"field": "username", "value": "invalid"}
err := errors.ValidationError("ValidateUser", "invalid username format", details)
```

### Error Propagation

Wrap errors with additional context as they propagate up the call stack:

```go
func ProcessData(data []byte) error {
    result, err := parseData(data)
    if err != nil {
        return errors.Wrap("ProcessData", err)
    }
    // ...
}
```

### Handling Multiple Errors

Use ErrorList to collect and process multiple errors:

```go
func ValidateInput(input Input) error {
    errList := errors.NewErrorList()
    
    if input.Name == "" {
        errList.Add(errors.InputError("ValidateInput", "name is required", nil))
    }
    
    if input.Age < 0 {
        errList.Add(errors.InputError("ValidateInput", "age cannot be negative", nil))
    }
    
    return errList.ToError()
}
```

### Error Logging

Log errors with appropriate context:

```go
func HandleRequest(input Input) error {
    err := ProcessInput(input)
    if err != nil {
        errors.LogError(logger, err)
        return err
    }
    // ...
}
```

### API Error Responses

Convert errors to API-friendly responses:

```go
func CreateErrorResponse(err error) map[string]interface{} {
    return errors.ConvertToErrorInfo(err)
}
```

## Integration Points

The error handling system is integrated with several key components:

1. **State Manager**: Converts S3 state errors to function errors
2. **Handler**: Creates standardized error responses for API clients
3. **Processor**: Handles processing-specific errors with context
4. **Parser**: Reports parsing errors with detailed information

## Error Categories and Codes

### Input Errors (CategoryInput)
- `INVALID_INPUT`: General input validation failure
- `MISSING_FIELD`: Required field is missing
- `INVALID_FORMAT`: Input format is invalid

### Processing Errors (CategoryProcess)
- `PROCESSING_FAILED`: General processing failure
- `PARSING_FAILED`: Failed to parse data
- `EXTRACTION_FAILED`: Failed to extract data
- `VALIDATION_FAILED`: Validation check failed

### State Errors (CategoryState)
- `STATE_LOAD_FAILED`: Failed to load state
- `STATE_STORE_FAILED`: Failed to store state
- `REFERENCE_INVALID`: Invalid S3 reference
- `ENVELOPE_INVALID`: Invalid state envelope

### System Errors (CategorySystem)
- `INTERNAL_ERROR`: Unexpected internal error
- `DEPENDENCY_FAILED`: External dependency failed
- `TIMEOUT`: Operation timed out