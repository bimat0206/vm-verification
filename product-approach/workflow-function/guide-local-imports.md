Standardization Guide for Vending Machine Verification Lambda Functions

Follow this guide to standardize all Lambda functions in the vending machine verification workflow following the
standardized approach implemented in InitializeFunction.

Overall Approach

For each Lambda function:
1. Update imports to use the shared schema package with local import paths
2. Remove status management (now handled by Step Functions)
3. Support both standardized and legacy formats
4. Add standardized error handling
5. Update database operations to include schema version

Step-by-Step Instructions

1. Update Dependencies

First, update the go.mod file:

```go
module workflow-function/YourFunctionName

go 1.21

require (
    // existing dependencies...
    workflow-function/shared/schema v0.0.0
)

// Add this replace directive
replace workflow-function/shared/schema => ../shared/schema
```

2. Update Request/Response Models

Change your models to use the standardized schema:

```go
import "workflow-function/shared/schema"

// Legacy format support
type YourFunctionRequest struct {
    // Standard schema fields
    SchemaVersion       string                  `json:"schemaVersion,omitempty"`
    VerificationContext *schema.VerificationContext `json:"verificationContext,omitempty"`

    // Legacy direct fields for backward compatibility
    // Function-specific fields here...
}
```

3. Update Handler Function

Modify the handler to detect and handle both formats:

```go
func Handler(ctx context.Context, event interface{}) (interface{}, error) {
    // Parse event and detect format
    var request YourFunctionRequest

    // Check for schema version to determine format
    if schemaVersion, ok := event["schemaVersion"].(string); ok && schemaVersion != "" {
        // Process standardized format
        // Extract verification context
    } else {
        // Process legacy format
    }

    // Continue with function-specific logic
}
```

4. Remove Status Management

Important: Do NOT set the Status field in the VerificationContext. This is now handled by the Step Functions state
machine.

Remove code like:
```go
// REMOVE: verificationContext.Status = "SOME_STATUS"
```

5. Standardize Error Handling

Use the standardized error structure:

```go
if err != nil {
    errorInfo := &schema.ErrorInfo{
        Code:      "FUNCTION_SPECIFIC_ERROR_CODE",
        Message:   err.Error(),
        Timestamp: schema.FormatISO8601(),
        Details: map[string]interface{}{
            // Function-specific error details
        },
    }

    // Add error to context if returning it
    verificationContext.Error = errorInfo
    return resultWithError, err
}
```

6. Update DynamoDB Operations

If your function interacts with DynamoDB:

```go
// When storing data
item := DynamoDBItem{
    // Other fields...
    SchemaVersion: schema.SchemaVersion,
}

// When retrieving data
// Check for schema version to determine format
var schemaVersion string
if sv, ok := result.Item["schemaVersion"]; ok {
    // Handle based on schema version
}
```

7. Docker Build Considerations

When building Docker images for Lambda functions with local imports:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy only the necessary components
COPY workflow-function/shared/schema/ /app/workflow-function/shared/schema/
COPY workflow-function/YourFunctionName/ /app/workflow-function/YourFunctionName/

# Build from function directory
WORKDIR /app/workflow-function/YourFunctionName
RUN go build -o main

FROM alpine:latest
COPY --from=builder /app/workflow-function/YourFunctionName/main /app/main
CMD ["/app/main"]
```

8. Update CHANGELOG.md

Document your changes:

```markdown
## [1.2.0] - 2025-05-14

### Added
- Integration with shared schema package using local imports
- Support for standardized status transitions
- Schema version handling
- Standardized error handling
- Backward compatibility
```

Function-Specific Standardization Notes

FetchHistoricalVerification
- Use schema.VerificationContext for input/output
- No status field changes - state machine handles them
- Return standardized historical context

FetchImages
- Standardize image structure using schema.ImageData
- Return images in standardized format
- No status updates

PrepareSystemPrompt
- Use schema.SystemPrompt for output
- Standardize BedrockConfig structure

PrepareTurn1Prompt & PrepareTurn2Prompt
- Use schema.CurrentPrompt for output
- Respect standardized image references

ExecuteTurn Functions
- Use standardized BedrockConfig
- Handle errors with ErrorInfo structure
- Return responses in standardized format

ProcessResponse Functions
- Use standardized analysis formats
- No status updates

Testing Approach

For each function:
1. Test with new standardized input format
2. Test with legacy format (backward compatibility)
3. Verify correct output format
4. Test error scenarios with standardized error responses