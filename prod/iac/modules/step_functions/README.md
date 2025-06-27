# Step Functions Module


## Overview
This module creates and configures an AWS Step Functions state machine for the vending machine verification workflow. It includes integration with Lambda functions, DynamoDB, and API Gateway.

## Module Structure

- **main.tf**: Creates the Step Functions state machine and API Gateway integration
- **iam.tf**: Defines IAM roles and policies for Step Functions
- **variables.tf**: Contains input variable definitions
- **outputs.tf**: Contains output definitions (renamed from output.tf)
- **templates/state_machine_definition.tftpl**: Template for the state machine definition

## Usage

```hcl
module "step_functions" {
  source = "./modules/step_functions"
  
  state_machine_name = "verification-workflow"
  log_level          = "ALL"
  
  lambda_function_arns = {
    initialize                    = module.lambda_functions.function_arns["initialize"]
    fetch_historical_verification = module.lambda_functions.function_arns["fetch_historical_verification"]
    # Additional Lambda functions...
  }
  
  # Additional variables as needed
}
```

## State Machine Workflow

The state machine implements a workflow for vending machine verification with different paths based on verification type:

1. **GenerateMissingFields**: Sets initial request ID and timestamp if not provided
2. **CheckVerificationType**: Determines the verification type and routes accordingly

For LAYOUT_VS_CHECKING type:
- **InitializeLayoutChecking**: Sets up the verification context for layout checking
- Continues to FetchImages directly

For PREVIOUS_VS_CURRENT type:
- **InitializePreviousCurrent**: Sets up the verification context for previous vs current
- **FetchHistoricalVerification**: Retrieves historical verification data
- Then continues to FetchImages

Common path for both types:
1. **FetchImages**: Gets images from S3
2. **PrepareSystemPrompt**: Prepares the system prompt for Bedrock
   - Handles missing historicalContext field for LAYOUT_VS_CHECKING type
3. **InitializeConversationState**: Sets up the conversation state
4. **PrepareTurn1Prompt/ExecuteTurn1**: First turn of the conversation
5. **ProcessTurn1Response**: Processes the first turn response
6. **PrepareTurn2Prompt/ExecuteTurn2**: Second turn of the conversation
7. **ProcessTurn2Response**: Processes the second turn response
8. **FinalizeResults**: Finalizes the verification results
9. **StoreResults**: Stores the results in DynamoDB
10. **Notify**: (Optional) Sends notifications
11. **WorkflowComplete**: Completes the workflow

## Parameters Mapping

Different Lambda functions in the workflow expect different input structures. The state machine handles this by using appropriate parameter mapping for each Lambda function:

### Optional Fields Handling

Some fields in the workflow data are only present for certain verification types. The state machine handles these cases:

- **historicalContext**: Only present for PREVIOUS_VS_CURRENT verification type
  ```json
  "historicalContext": {}
  ```
  This approach provides an empty object when historicalContext is missing (as in LAYOUT_VS_CHECKING type), preventing JSONPath errors.

- **previousVerificationId**: Only required for PREVIOUS_VS_CURRENT verification type

### Initialize Lambda

The Initialize Lambda expects a nested verificationContext object:

```json
"Parameters": {
  "verificationContext": {
    "verificationType.$": "$.verificationContext.verificationType",
    "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
    "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
    "vendingMachineId.$": "$.verificationContext.vendingMachineId",
    "layoutId.$": "$.verificationContext.layoutId",
    "layoutPrefix.$": "$.verificationContext.layoutPrefix",
    "notificationEnabled.$": "$.verificationContext.notificationEnabled"
  },
  "requestId.$": "$.requestId",
  "requestTimestamp.$": "$.requestTimestamp"
}
```

### FetchImages Lambda

The FetchImages Lambda expects fields at the top level:

```json
"Parameters": {
  "verificationId.$": "$.verificationContext.verificationId",
  "verificationType.$": "$.verificationContext.verificationType",
  "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
  "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
  "layoutId.$": "$.verificationContext.layoutId",
  "layoutPrefix.$": "$.verificationContext.layoutPrefix",
  "vendingMachineId.$": "$.verificationContext.vendingMachineId"
}
```

This mapping ensures that each Lambda function receives input in the expected format, regardless of whether the workflow is invoked directly or via API Gateway.

## API Gateway Integration

The module can optionally create an API Gateway integration for the state machine. This allows the state machine to be invoked directly from API Gateway without an intermediary Lambda function.

To enable API Gateway integration, set the `create_api_gateway_integration` variable to `true` and provide the necessary API Gateway parameters:

```hcl
module "step_functions" {
  source = "./modules/step_functions"
  
  # Other variables...
  
  create_api_gateway_integration = true
  api_gateway_id                 = module.api_gateway.api_id
  api_gateway_root_resource_id   = module.api_gateway.root_resource_id
  api_gateway_endpoint           = module.api_gateway.api_endpoint
}
```

This creates a new resource at `/api/workflow` that can be used to start a Step Functions execution directly.

## Error Handling

The state machine includes comprehensive error handling:

- **Retry Logic**: Each task state includes retry logic for transient errors
- **Catch Blocks**: Error states handle different types of failures
- **Error States**: Dedicated states for handling different types of errors
- **Result Path**: Error information is stored in the execution state

## Outputs

The module provides the following outputs:

- **state_machine_arn**: ARN of the Step Functions state machine
- **state_machine_name**: Name of the Step Functions state machine
- **api_gateway_role_arn**: ARN of the IAM role for API Gateway to invoke Step Functions
- **workflow_api_endpoint**: API Gateway endpoint for the Step Functions workflow (if enabled)
