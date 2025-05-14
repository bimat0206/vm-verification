# Shared Schema Package

This package provides standardized data models and validation for the vending machine verification workflow. It ensures consistency across all Lambda functions and Step Functions states.

## Overview

The shared schema package defines:

1. Core data structures used throughout the workflow
2. Status constants for explicit state transitions
3. Validation functions to ensure data integrity
4. Helper functions for common operations

## Key Components

### Data Structures

- `VerificationContext`: The central context object that flows through the entire workflow
- `ImageData`: Standardized image reference structure
- `ConversationState`: Tracks conversation state with Bedrock
- `SystemPrompt` and `CurrentPrompt`: Prompt structures for Bedrock API
- `BedrockConfig`: Configuration for Bedrock API calls
- `FinalResults`: Structure for verification results
- `WorkflowState`: Comprehensive state representation

### Status Constants

The package defines a comprehensive set of status constants that model the state machine's explicit status transitions:

```go
// Status constants aligned with state machine
const (
    StatusVerificationRequested  = "VERIFICATION_REQUESTED"
    StatusVerificationInitialized = "VERIFICATION_INITIALIZED"
    StatusFetchingImages         = "FETCHING_IMAGES"
    StatusImagesFetched          = "IMAGES_FETCHED"
    // ... and many more
)
```

### Validation Functions

- `ValidateVerificationContext`: Validates the verification context structure
- `ValidateWorkflowState`: Validates the complete workflow state

## Using the Package

### Importing

```go
import "github.com/kootoro/vending-machine-verification/workflow-function/shared/schema"
```

### Creating a Standardized Context

```go
context := &schema.VerificationContext{
    VerificationId:    uuid.New().String(), 
    VerificationType:  schema.VerificationTypeLayoutVsChecking,
    Status:            schema.StatusVerificationRequested,
    // ... other fields
}
```

### Validating Input

```go
errors := schema.ValidateVerificationContext(context)
if len(errors) > 0 {
    return fmt.Errorf("validation failed: %s", errors.Error())
}
```

### Working with Status Transitions

The Step Functions state machine manages status transitions. Lambda functions should:

1. Read the current status (if needed)
2. Perform their specific task
3. Return the result without modifying the status

The Step Functions state machine will:
1. Set the appropriate status before invoking a Lambda
2. Update the status after the Lambda completes

## Backward Compatibility

For backward compatibility, Lambda functions should:

1. Check for the `schemaVersion` field
2. Support both old and new formats
3. Always return the standardized format
4. Set the schema version to the current version

## Best Practices

1. Always use constants from the shared package for status and verification types
2. Validate input using the provided validation functions
3. Don't modify the status field in Lambda functions (handled by state machine)
4. Use standardized error structures for consistent error handling
5. Check for schema version to determine input format

## Schema Evolution

When evolving the schema:

1. Create a new schema version
2. Maintain backward compatibility with older versions
3. Document all changes in the CHANGELOG.md
4. Update Step Functions state machine to work with the new version
5. Update Lambda functions one at a time