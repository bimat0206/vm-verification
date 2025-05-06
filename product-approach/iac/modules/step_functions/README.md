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

The state machine implements a workflow for vending machine verification:

1. **Initialize**: Sets up the verification context and parameters
2. **CheckVerificationType**: Determines the verification type
3. **FetchHistoricalVerification**: (Optional) Retrieves historical verification data
4. **FetchImages**: Gets images from S3
5. **PrepareSystemPrompt**: Prepares the system prompt for Bedrock
6. **InitializeConversationState**: Sets up the conversation state
7. **PrepareTurn1Prompt/ExecuteTurn1**: First turn of the conversation
8. **ProcessTurn1Response**: Processes the first turn response
9. **PrepareTurn2Prompt/ExecuteTurn2**: Second turn of the conversation
10. **ProcessTurn2Response**: Processes the second turn response
11. **FinalizeResults**: Finalizes the verification results
12. **StoreResults**: Stores the results in DynamoDB
13. **Notify**: (Optional) Sends notifications
14. **WorkflowComplete**: Completes the workflow

## Parameters Mapping

The Initialize state includes Parameters mapping to ensure consistent input structure:

```json
"Parameters": {
  "verificationType.$": "$.verificationType",
  "referenceImageUrl.$": "$.referenceImageUrl",
  "checkingImageUrl.$": "$.checkingImageUrl",
  "vendingMachineId.$": "$.vendingMachineId",
  "layoutId.$": "$.layoutId",
  "layoutPrefix.$": "$.layoutPrefix",
  "notificationEnabled.$": "$.notificationEnabled",
  "requestId.$": "$.requestId",
  "requestTimestamp.$": "$.requestTimestamp"
}
```

This mapping ensures that the Lambda function receives a consistent input structure regardless of whether the workflow is invoked directly or via API Gateway.

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
