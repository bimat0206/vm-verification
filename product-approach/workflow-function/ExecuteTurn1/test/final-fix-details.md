# ExecuteTurn1 Function Error Fix Details

## Root Cause Analysis

After examining the Step Function definition and the associated Lambda functions, I've identified the exact cause of the issue:

1. The `PrepareTurn1Prompt` function (lines 233-246 in state machine definition) outputs a flat structure for `CurrentPrompt` in its response:
   ```json
   {
     "CurrentPrompt": {
       "Messages": [...],
       "TurnNumber": 1,
       "PromptID": "prompt-...",
       "CreatedAt": "...",
       "PromptVersion": "...",
       "ImageIncluded": "reference"
     }
   }
   ```

2. When this output is passed to the `ExecuteTurn1` function (lines 249-277), it uses parameter mapping that's causing the nesting:
   ```
   "Parameters": {
     "currentPrompt.$": "$.currentPrompt",
     ...
   }
   ```

3. The `ExecuteTurn1` function is expecting a nested structure with `CurrentPromptWrapper`:
   ```json
   {
     "CurrentPrompt": {
       "CurrentPrompt": {
         "Messages": [...],
         ...
       }
     }
   }
   ```

4. The `ThinkingType` is set to `"enable"` instead of `"enabled"` in `BedrockConfig`.

## Fix Options

### Option 1: Modify the Step Function (Recommended)

Update the `ExecuteTurn1` state in the Step Function (around line 249) to transform the input into the expected structure:

```json
"ExecuteTurn1": {
  "Type": "Task",
  "Resource": "${function_arns["execute_turn1"]}",
  "Parameters": {
    "verificationContext.$": "$.verificationContext",
    "images.$": "$.images",
    "systemPrompt.$": "$.systemPrompt",
    "currentPrompt": {
      "currentPrompt.$": "$.currentPrompt"
    },
    "conversationState.$": "$.conversationState",
    "historicalContext": {},
    "layoutMetadata.$": "$.layoutMetadata",
    "bedrockConfig": {
      "anthropic_version.$": "$.currentPrompt.bedrockConfig.anthropic_version",
      "max_tokens.$": "$.currentPrompt.bedrockConfig.max_tokens",
      "thinking": {
        "type": "enabled",
        "budget_tokens.$": "$.currentPrompt.bedrockConfig.thinking.budget_tokens"
      }
    }
  },
  "ResultPath": "$.turn1Response",
  ...
}
```

### Option 2: Modify the ExecuteTurn1 Lambda (Alternative)

Update the `ExecuteTurn1` function to better handle the input structure:

1. Improve the `extractCurrentPrompt` function to better handle both nested and non-nested structures
2. Update the `extractBedrockConfig` function to normalize the thinking type
3. Ensure consistent use of these extraction functions throughout the code

### Implementation Plan

1. First, try Option 1 (Step Function modification) as it's less invasive and doesn't require redeploying the Lambda function
2. If that doesn't resolve the issue or is difficult to implement, proceed with Option 2
3. Test the changes thoroughly before deploying to production

## Long-term Recommendations

1. **Standardize Input/Output Formats**: Ensure consistent structures between Lambda functions in the workflow
2. **Add Robust Input Validation**: Validate and normalize inputs at each function entry point
3. **Update Types and Documentation**: Ensure type definitions match the actual expected structures
4. **Add Test Suite**: Create comprehensive tests for the workflow to catch these issues early