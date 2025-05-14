# Step Function Update Instructions

## Overview

This document provides instructions for updating the `ExecuteTurn1` state in your Step Function to fix the nested structure issue causing the Lambda error.

## What to Change

Replace the current `ExecuteTurn1` state in your Step Function definition with the updated configuration.

### File to Edit
`/Users/mac/Library/CloudStorage/OneDrive-Personal/1.WORK/Git/programing/vending-machine-verification/product-approach/iac/modules/step_functions/templates/state_machine_definition.tftpl`

### Changes Needed

Find the existing `ExecuteTurn1` state (around line 249) that looks like:

```json
"ExecuteTurn1": {
  "Type": "Task",
  "Resource": "${function_arns[\"execute_turn1\"]}",
  "Parameters": {
    "verificationContext.$": "$.verificationContext",
    "images.$": "$.images",
    "systemPrompt.$": "$.systemPrompt",
    "currentPrompt.$": "$.currentPrompt",
    "conversationState.$": "$.conversationState",
    "historicalContext": {},
    "layoutMetadata.$": "$.layoutMetadata"
  },
  "ResultPath": "$.turn1Response",
  "Retry": [
    {
      "ErrorEquals": ["ServiceException", "ThrottlingException"],
      "IntervalSeconds": 3,
      "MaxAttempts": 5,
      "BackoffRate": 2.0
    }
  ],
  "Catch": [
    {
      "ErrorEquals": ["States.ALL"],
      "ResultPath": "$.error",
      "Next": "HandleBedrockError"
    }
  ],
  "Next": "ProcessTurn1Response"
}
```

Replace it with:

```json
"ExecuteTurn1": {
  "Type": "Task",
  "Resource": "${function_arns[\"execute_turn1\"]}",
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
    {
      "ErrorEquals": ["ServiceException", "ThrottlingException"],
      "IntervalSeconds": 3,
      "MaxAttempts": 5,
      "BackoffRate": 2.0
    }
  ],
  "Catch": [
    {
      "ErrorEquals": ["States.ALL"],
      "ResultPath": "$.error",
      "Next": "HandleBedrockError"
    }
  ],
  "Next": "ProcessTurn1Response"
}
```

## Key Changes

1. **Added Nested Structure for systemPrompt:**
   ```json
   "systemPrompt": {
     "systemPrompt.$": "$.systemPrompt"
   }
   ```

2. **Added Nested Structure for currentPrompt:**
   ```json
   "currentPrompt": {
     "currentPrompt.$": "$.currentPrompt"
   }
   ```

3. **Added Explicit bedrockConfig with Fixed Thinking Type:**
   ```json
   "bedrockConfig": {
     "anthropic_version.$": "$.systemPrompt.bedrockConfig.anthropic_version",
     "max_tokens.$": "$.systemPrompt.bedrockConfig.max_tokens",
     "thinking": {
       "type": "enabled",
       "budget_tokens.$": "$.systemPrompt.bedrockConfig.thinking.budget_tokens"
     }
   }
   ```

## After Making Changes

1. Deploy the updated Step Function using your Terraform workflow
2. Test with a sample verification request
3. Monitor for successful execution without the previous error

## Verification

To verify the fix is working:
1. Check CloudWatch logs for the ExecuteTurn1 Lambda function
2. Confirm there are no more "Runtime.ExitError" messages
3. Ensure the overall workflow completes successfully