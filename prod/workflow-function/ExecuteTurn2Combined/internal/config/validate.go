package config

import (
	"time"
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

	if _, err := time.LoadLocation(c.DatePartitionTimezone); err != nil {
		return errors.NewConfigError(
			"InvalidTimezone",
			"invalid timezone",
			"DATE_PARTITION_TIMEZONE",
		)
	}

	// Validate temperature
	if c.Processing.Temperature < 0 || c.Processing.Temperature > 1 {
		return errors.NewValidationError("temperature must be between 0 and 1",
			map[string]interface{}{"current_value": c.Processing.Temperature})
	}

	return nil
}
