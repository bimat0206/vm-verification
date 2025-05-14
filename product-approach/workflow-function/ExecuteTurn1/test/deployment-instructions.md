# ExecuteTurn1 Function Fix - Deployment Instructions

## Summary
I've fixed the Step Function definition to correctly structure the input for your `ExecuteTurn1` Lambda function. The main issues were:

1. Nested JSON structure for `currentPrompt` and `systemPrompt`
2. The `thinking.type` value needed to be "enabled" instead of "enable"

## Files Modified

1. Created a complete fixed version of your Step Function definition at:
   `/Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac/modules/step_functions/templates/state_machine_definition_fixed.tftpl`

## Deployment Steps

### 1. Replace the State Machine Definition

```bash
# Backup your current definition
cp /Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac/modules/step_functions/templates/state_machine_definition.tftpl /Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac/modules/step_functions/templates/state_machine_definition.tftpl.bak

# Replace with the fixed version
cp /Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/workflow-function/ExecuteTurn1/test/state_machine_definition_fixed.tftpl /Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac/modules/step_functions/templates/state_machine_definition.tftpl
```

### 2. Deploy the Updated Step Function

Run your Terraform deployment process:

```bash
cd /Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac
terraform apply
```

Alternatively, if you're using AWS Console:
1. Navigate to AWS Step Functions in the console
2. Select your State Machine
3. Click "Edit"
4. Replace the definition with the content from `state_machine_definition_fixed.tftpl`
5. Click "Save"

### 3. Test the Fix

1. Trigger a new verification process through your normal workflow
2. Monitor the Step Function execution in AWS Console
3. Check the CloudWatch Logs for the ExecuteTurn1 function to ensure it's receiving the correctly structured input

## Key Changes Made

```json
"ExecuteTurn1": {
  "Type": "Task",
  "Resource": "${function_arns["execute_turn1"]}",
  "Parameters": {
    "verificationContext.$": "$.verificationContext",
    "images.$": "$.images",
    "systemPrompt": {
      "systemPrompt.$": "$.systemPrompt"
    },
    "currentPrompt": {
      "currentPrompt.$": "$.currentPrompt"
    },
    "conversationState.$": "$.conversationState",
    "historicalContext": {},
    "layoutMetadata.$": "$.layoutMetadata",
    "bedrockConfig": {
      "anthropic_version.$": "$.systemPrompt.bedrockConfig.anthropic_version",
      "max_tokens.$": "$.systemPrompt.bedrockConfig.max_tokens",
      "thinking": {
        "type": "enabled",
        "budget_tokens.$": "$.systemPrompt.bedrockConfig.thinking.budget_tokens"
      }
    }
  },
  "ResultPath": "$.turn1Response",
  "Retry": [
    ...
  ],
  "Catch": [
    ...
  ],
  "Next": "ProcessTurn1Response"
}
```

The changes create the expected nested structure for `systemPrompt` and `currentPrompt`, and explicitly set the `thinking.type` to "enabled" instead of using the input value "enable".