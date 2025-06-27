# Vending Machine Verification Standardization Implementation Guide

This guide outlines the process for implementing standardization across the vending machine verification workflow. The standardization focuses on creating consistent interfaces between Step Functions and Lambda functions, with explicit status management at the state machine level.

## 1. Overview of Changes

This standardization effort addresses several inconsistencies:

- Inconsistent parameter nesting and structure
- State parameter inconsistencies
- Field naming discrepancies
- Missing fields in state machine parameters
- BedrockConfig handling issues
- Image reference inconsistencies
- Inconsistent error handling
- Historical context discrepancies

## 2. Core Components

The standardization comprises three key components:

1. **JSON Schema Definition**: A comprehensive schema that defines all data structures
2. **Standardized Status Transitions**: Explicit state transitions managed by the state machine
3. **Updated State Machine Definition**: Refactored state machine with consistent parameter passing

## 3. Implementation Steps

### Step 1: Add Schema Definition

1. Add the schema file to your project:
   ```
   /iac/modules/step_functions/templates/schemas/verification_schema.json
   ```

2. Create a shared Go package for schema validation in Lambda functions:
   ```
   /workflow-function/shared/schema/
   ```

### Step 2: Update State Machine Definition

1. Update the state machine definition to use standardized status transitions
2. Add explicit status transition states
3. Ensure consistent parameter passing with JsonPath references
4. Update IAM permissions as needed

### Step 3: Update Lambda Functions

1. For each Lambda function, update the request/response models to use the standardized schema
2. Add validation against the schema
3. Remove status management code from Lambda functions (now handled by state machine)
4. Ensure consistent field naming and structure

### Step 4: Test Integration Points

1. Create integration tests for each state transition
2. Validate schema compliance at each step
3. Test error handling and recovery paths

## 4. Status Transition Map

The state machine now explicitly handles status transitions according to this flow:

```
VERIFICATION_REQUESTED → VERIFICATION_INITIALIZED → IMAGES_FETCHED → 
PROMPT_PREPARED → TURN1_COMPLETED → TURN1_PROCESSED → 
TURN2_COMPLETED → TURN2_PROCESSED → RESULTS_FINALIZED → 
RESULTS_STORED → NOTIFICATION_SENT → COMPLETED
```

Each state in the workflow is responsible for managing its own status transitions.

## 5. Error Management

Error states are now standardized:

- INITIALIZATION_FAILED
- HISTORICAL_FETCH_FAILED
- IMAGE_FETCH_FAILED
- BEDROCK_PROCESSING_FAILED
- VERIFICATION_FAILED

Error information is consistently structured with:
- Error code
- Error message
- Timestamp
- Detailed information

## 6. Lambda Function Changes

### Function Interface Guidelines

1. All Lambda functions should accept the standardized input schema
2. Functions should NOT modify the status field (handled by state machine)
3. Functions should return errors in the standardized format
4. Each function must add the correct structure to its output path only
5. Functions should validate inputs against the schema
6. Functions should NOT assume fields exist - always check for null/empty values

⏺ Based on the workflow sequence and standardization requirements, you should update the Lambda functions in this
  priority order:

  1. InitializeFunction - It's the entry point and sets up basic context
  2. FetchHistoricalVerification - Used early in the workflow for PREVIOUS_VS_CURRENT type
  3. FetchImages - Critical for all verification types and used in most paths
  4. PrepareSystemPrompt - Sets up the foundation for Bedrock interactions
  5. PrepareTurn1Prompt - First step in the conversation sequence
  6. ExecuteTurn1 - Executes the first Bedrock call
  7. ProcessTurn1Response - Creates reference analysis needed by subsequent steps

  For each function, you'll need to:
  1. Update input parameter structure to match the schema
  2. Remove status management code (now handled by state machine)
  3. Ensure consistent error responses
  4. Add schema version handling (with fallback for backward compatibility)
  5. Standardize JSONPath references
### Example Function Changes:

#### Before:
```go
type InitializeRequest struct {
    VerificationType      string `json:"verificationType"`
    ReferenceImageUrl     string `json:"referenceImageUrl"`
    // ...other fields
}
```

#### After:
```go
import "github.com/kootoro/vending-machine-verification/workflow-function/shared/schema"

type InitializeRequest struct {
    VerificationContext schema.VerificationContext `json:"verificationContext"`
    // Other standardized fields
}
```

## 7. Backward Compatibility

To maintain backward compatibility during transition:

1. Add support for both old and new formats in Lambda functions
2. Use a "schemaVersion" field to differentiate
3. Set a deprecation timeline for older formats
4. Deploy and test the new state machine with one Lambda at a time

## 8. Testing Plan

1. Create unit tests for each Lambda function against the schema
2. Create integration tests for each state transition
3. Create end-to-end tests for the entire workflow
4. Validate schema compliance at each step
5. Test error handling and recovery paths

## 9. Future Enhancements

1. Add schema versioning for easier updates
2. Create code generation tools for Lambda handlers
3. Implement a centralized validation service
4. Create visualization tools for status transitions

## 10. Implementation Timeline

1. Phase 1: Schema definition and shared package creation
2. Phase 2: State machine updates
3. Phase 3: Lambda function updates (one at a time)
4. Phase 4: Integration testing
5. Phase 5: Deployment to production

## 11. Troubleshooting

Common issues and solutions:

1. Schema validation errors: 
   - Use the schema validator in the shared package
   - Check for proper field types, especially with JSON numbers and booleans

2. State machine execution failures:
   - Validate JSONPath expressions
   - Check for null values in input
   - Verify IAM permissions

3. Lambda function errors:
   - Use proper error handling
   - Return standardized error structures
   - Log detailed information for debugging