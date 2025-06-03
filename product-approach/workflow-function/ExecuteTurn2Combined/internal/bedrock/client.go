package bedrock

import (
	"context"
	"strings"

	"workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// Client handles Bedrock API interactions
type Client struct {
	adapter *Adapter
	config  *Config
	logger  logger.Logger
}

// NewClient creates a new Bedrock client
func NewClient(sharedClient *bedrock.BedrockClient, config *Config, logger logger.Logger) *Client {
	adapter := NewAdapter(sharedClient, logger)
	return &Client{
		adapter: adapter,
		config:  config,
		logger: logger.WithFields(map[string]interface{}{
			"component": "BedrockClient",
		}),
	}
}

// ProcessTurn1 handles Turn 1 processing
func (c *Client) ProcessTurn1(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*BedrockResponse, error) {
	// Apply operational timeout boundary
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	// Direct delegation to adapter - no preprocessing
	response, latencyMs, err := c.adapter.Converse(ctx, systemPrompt, turnPrompt, base64Image, c.config)
	if err != nil {
		return nil, err
	}

	// Enrich with metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["latency_ms"] = latencyMs
	response.Metadata["model_id"] = c.config.ModelID

	return response, nil
}

// ValidateConfiguration ensures operational parameters are within bounds
func (c *Client) ValidateConfiguration() error {
	if c.config.ModelID == "" {
		return errors.NewValidationError("model ID cannot be empty", nil)
	}

	if c.config.MaxTokens <= 0 {
		return errors.NewValidationError("max tokens must be positive",
			map[string]interface{}{"current_value": c.config.MaxTokens})
	}

	if c.config.Temperature < 0 || c.config.Temperature > 1 {
		return errors.NewValidationError("temperature must be between 0 and 1",
			map[string]interface{}{"current_value": c.config.Temperature})
	}

	if c.config.Temperature == 1 && !strings.EqualFold(c.config.ThinkingType, "enable") && !strings.EqualFold(c.config.ThinkingType, "enabled") {
		return errors.NewValidationError("temperature may only be set to 1 when thinking is enabled",
			map[string]interface{}{
				"current_value": c.config.Temperature,
				"thinking_type": c.config.ThinkingType,
			})
	}

	return nil
}

// GetOperationalMetrics returns current operational parameters
func (c *Client) GetOperationalMetrics() map[string]interface{} {
	return map[string]interface{}{
		"model_id":          c.config.ModelID,
		"max_tokens":        c.config.MaxTokens,
		"temperature":       c.config.Temperature,
		"timeout_seconds":   c.config.Timeout.Seconds(),
		"region":            c.config.Region,
		"anthropic_version": c.config.AnthropicVersion,
		"architecture":      "adapter_pattern",
	}
}
