# Fix for Step Function to Lambda Input Structure

## Issues Found

Analyzing the actual Step Function input showed two critical issues:

1. **Nested Structure Issue**: The `currentPrompt` structure is correct but you have another `currentPrompt` object inside it with different properties
2. **Thinking Type**: The value is set to `"enable"` instead of `"enabled"` in `bedrockConfig.thinking.type`

## Fix Instructions

### For Step Function

Update your Step Function definition to ensure it passes the correct structure to the ExecuteTurn1 Lambda:

```json
{
  "currentPrompt": {
    "currentPrompt": { ... } // This nesting is expected by the code
  },
  "bedrockConfig": {
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 24000,
    "thinking": {
      "type": "enabled", // Change from "enable" to "enabled"
      "budget_tokens": 16000
    }
  }
}
```

### For Lambda Code

The Lambda code has both `extractCurrentPrompt()` and `extractBedrockConfig()` functions to handle these issues, but they may not be consistently used throughout the code. Update the Lambda code to:

1. Change all references to `input.CurrentPrompt.Messages` to get messages from the extracted prompt
2. Update all validation logic to properly validate the nested structure
3. Ensure all parameters pass through the normalized structure 

## Immediate Stop-Gap Fix

If changing the Step Function is difficult right now, you can modify your Lambda to better handle the input:

```go
// Update Line 135 in main.go to also accept "enable" as a valid value
if config.Thinking.Type == "enable" || config.Thinking.Type == "enabled" {
    config.Thinking.Type = "enabled"
}

// Update validation in internal/validation.go
if config.Thinking.Type != "enabled" && config.Thinking.Type != "enable" {
    return NewInvalidFieldError("thinking.type", config.Thinking.Type, "enabled or enable")
}

// Fix the root cause of the problem by ensuring you process the extracted prompt
currentPrompt, err := extractCurrentPrompt(input)
if err != nil {
    return err
}
// Use currentPrompt throughout the code instead of input.CurrentPrompt
```

## Sample Input Structure for Testing

See the `test-input-correct.json` file for a sample structure that matches what the Lambda code expects. The key difference is that it includes the expected nesting pattern but uses "enabled" instead of "enable" for thinking type.