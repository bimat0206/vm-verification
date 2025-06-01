# Thinking Blocks Implementation for ExecuteTurn1Combined

## Overview

This document describes the implementation of thinking blocks (reasoning mode) in the ExecuteTurn1Combined Lambda function. Thinking blocks provide detailed insights into the AI model's decision-making process during vending machine layout verification.

## Environment Variable Configuration

The thinking blocks feature is controlled by the `THINKING_TYPE` environment variable:

- **`THINKING_TYPE=enable`**: Enables thinking/reasoning mode
- **`THINKING_TYPE=disable`**: Disables thinking/reasoning mode  
- **`THINKING_TYPE` unset**: Disables thinking/reasoning mode (default behavior)

## Implementation Details

### 1. Configuration Layer (`internal/config/config.go`)

```go
// IsThinkingEnabled returns true if thinking/reasoning mode is enabled
// Thinking is enabled only when THINKING_TYPE is explicitly set to "enable"
// Thinking is disabled when THINKING_TYPE is "disable" or unset (empty string)
func (c *Config) IsThinkingEnabled() bool {
    return c.Processing.ThinkingType == "enable"
}
```

The configuration reads directly from the Lambda environment variable without hardcoded defaults, ensuring thinking is only enabled when explicitly set to "enable".

### 2. Service Layer (`internal/services/bedrock.go`)

The service layer checks the configuration and passes the appropriate thinking type to the shared Bedrock client:

```go
// Create shared bedrock client first
thinkingType := "disable"
if cfg.IsThinkingEnabled() {
    thinkingType = "enable"
}
clientConfig := sharedBedrock.CreateClientConfig(
    cfg.AWS.Region,
    cfg.AWS.AnthropicVersion,
    cfg.Processing.MaxTokens,
    thinkingType,
    cfg.Processing.BudgetTokens,
)
```

### 3. Adapter Layer (`internal/bedrock/adapter.go`)

The adapter layer handles the actual API request configuration:

```go
// Add thinking/reasoning configuration if enabled
if config.ThinkingType == "enable" {
    request.Reasoning = "enable"
    request.InferenceConfig.Reasoning = "enable"
    a.logger.Debug("thinking_enabled", map[string]interface{}{
        "thinking_type":   config.ThinkingType,
        "budget_tokens":   config.ThinkingBudget,
    })
} else {
    a.logger.Debug("thinking_disabled", map[string]interface{}{
        "thinking_type": config.ThinkingType,
    })
}
```

The adapter also extracts thinking content from responses and includes it in metadata:

```go
// Add thinking content to metadata if available
if thinking != "" {
    metadata["thinking"] = thinking
    metadata["has_thinking"] = true
    metadata["thinking_length"] = len(thinking)
    a.logger.Debug("thinking_extracted", map[string]interface{}{
        "thinking_length": len(thinking),
        "thinking_tokens": tokenUsage.ThinkingTokens,
    })
} else {
    metadata["has_thinking"] = false
}
```

## JSON Response Structure

When thinking blocks are enabled, the JSON responses include additional fields:

### Raw Response (`turn1-raw-response.json`)

```json
{
  "bedrockMetadata": {
    "hasThinking": true,
    "thinkingEnabled": true
  },
  "tokenUsage": {
    "thinking": 245,
    "total": 6142
  },
  "thinkingBlocks": [
    {
      "timestamp": "2025-05-31T16:11:33Z",
      "component": "bedrock-adapter",
      "stage": "api-initialization",
      "decision": "Selected Claude-3.7-Sonnet model for visual analysis",
      "reasoning": "This model provides excellent multimodal capabilities...",
      "confidence": 95
    }
  ]
}
```

### Conversation Response (`turn1-conversation.json`)

```json
{
  "bedrockMetadata": {
    "hasThinking": true,
    "thinkingEnabled": true
  },
  "tokenUsage": {
    "thinking": 245,
    "total": 6142
  },
  "thinkingBlocks": [
    {
      "timestamp": "2025-05-31T16:11:33Z",
      "component": "bedrock-service",
      "stage": "conversation-initialization",
      "decision": "Initiated multimodal conversation with system prompt",
      "reasoning": "System prompt provides comprehensive vending machine analysis framework...",
      "confidence": 99
    }
  ]
}
```

## Thinking Block Structure

Each thinking block contains:

- **`timestamp`**: ISO 8601 timestamp of when the decision was made
- **`component`**: The system component making the decision (e.g., "bedrock-adapter", "bedrock-service")
- **`stage`**: The processing stage (e.g., "api-initialization", "image-processing", "response-generation")
- **`decision`**: Brief description of the decision made
- **`reasoning`**: Detailed explanation of why the decision was made
- **`confidence`**: Confidence level (0-100) in the decision

## Benefits

1. **Transparency**: Provides insight into the AI model's decision-making process
2. **Debugging**: Helps identify issues in the reasoning chain
3. **Audit Trail**: Creates a detailed log of processing decisions
4. **Performance Analysis**: Token usage tracking for thinking vs. output tokens
5. **Quality Assurance**: Confidence levels help assess decision reliability

## Usage Examples

### Enable Thinking Blocks
Set the Lambda environment variable:
```bash
THINKING_TYPE=enable
```

### Disable Thinking Blocks
Set the Lambda environment variable:
```bash
THINKING_TYPE=disable
```

Or leave it unset (default behavior).

## Integration with Shared Bedrock Client

The implementation leverages the shared Bedrock client (`workflow-function/shared/bedrock`) which handles the low-level API interactions. The thinking configuration is passed through the client configuration and the thinking content is extracted using shared utilities:

- `sharedBedrock.ExtractThinkingFromResponse(response)`
- `sharedBedrock.ExtractTextFromResponse(response)`

## Future Enhancements

1. **Granular Control**: Add more specific thinking types (e.g., "detailed", "minimal")
2. **Filtering**: Allow filtering of thinking blocks by component or stage
3. **Metrics**: Add performance metrics for thinking vs. non-thinking requests
4. **Storage Optimization**: Compress thinking blocks for large responses

## Testing

To test the thinking blocks implementation:

1. Deploy the Lambda with `THINKING_TYPE=enable`
2. Trigger a vending machine verification
3. Check the S3 response files for thinking blocks
4. Verify token usage includes thinking tokens
5. Test with `THINKING_TYPE=disable` to ensure thinking is properly disabled

## Troubleshooting

### Common Issues

1. **Thinking blocks not appearing**: Check that `THINKING_TYPE=enable` is set
2. **High token usage**: Thinking tokens are additional to regular tokens
3. **Missing thinking content**: Ensure the shared Bedrock client supports reasoning
4. **JSON parsing errors**: Validate thinking block structure matches schema

### Debug Logging

The implementation includes debug logging at key points:
- Configuration loading
- Thinking enablement/disablement
- Thinking content extraction
- Token usage tracking

Check CloudWatch logs for debug messages when troubleshooting.
