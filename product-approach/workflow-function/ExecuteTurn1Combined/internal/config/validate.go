package config

import (
	"workflow-function/shared/errors"
)

// Validate performs validation on the loaded configuration
func (c *Config) Validate() error {
	// Validate Bedrock timeouts
	if c.Processing.BedrockConnectTimeoutSec <= 0 {
		return errors.NewConfigError(
			"BedrockTimeoutInvalid",
			"Bedrock connect timeout must be greater than 0",
			"BEDROCK_CONNECT_TIMEOUT_SEC",
		)
	}

	if c.Processing.BedrockCallTimeoutSec <= 0 {
		return errors.NewConfigError(
			"BedrockTimeoutInvalid",
			"Bedrock call timeout must be greater than 0",
			"BEDROCK_CALL_TIMEOUT_SEC",
		)
	}

	// Additional validation: call timeout should be greater than connect timeout
	if c.Processing.BedrockCallTimeoutSec <= c.Processing.BedrockConnectTimeoutSec {
		return errors.NewConfigError(
			"BedrockTimeoutInvalid",
			"Bedrock call timeout must be greater than connect timeout",
			"BEDROCK_CALL_TIMEOUT_SEC",
		)
	}

	return nil
}
