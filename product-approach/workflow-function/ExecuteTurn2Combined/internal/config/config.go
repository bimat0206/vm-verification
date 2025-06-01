package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"workflow-function/shared/errors"
)

// Config bundles all environment settings required by the Lambda.
type Config struct {
	AWS struct {
		Region                    string
		S3Bucket                  string
		BedrockModel              string
		AnthropicVersion          string
		DynamoDBVerificationTable string
		DynamoDBConversationTable string
	}
	Processing struct {
		MaxTokens                int
		BudgetTokens             int
		ThinkingType             string
		MaxRetries               int
		BedrockConnectTimeoutSec int
		BedrockCallTimeoutSec    int
		DiscrepancyThreshold     int
	}
	Logging struct {
		Level  string
		Format string
	}
	Prompts struct {
		TemplateVersion      string
		TemplateBasePath     string
		Turn2TemplateVersion string
	}
	DatePartitionTimezone string
}

// LoadConfiguration parses environment variables into a Config structure.
func LoadConfiguration() (*Config, error) {
	cfg := &Config{}
	cfg.AWS.Region = getEnv("AWS_REGION", "us-east-1")

	var err error
	cfg.AWS.S3Bucket, err = mustGet("STATE_BUCKET")
	if err != nil {
		return nil, err
	}

	cfg.AWS.BedrockModel, err = mustGet("BEDROCK_MODEL")
	if err != nil {
		return nil, err
	}

	cfg.AWS.AnthropicVersion = getEnv("ANTHROPIC_VERSION", "bedrock-2023-05-31")

	cfg.AWS.DynamoDBVerificationTable, err = mustGet("DYNAMODB_VERIFICATION_TABLE")
	if err != nil {
		return nil, err
	}

	cfg.AWS.DynamoDBConversationTable, err = mustGet("DYNAMODB_CONVERSATION_TABLE")
	if err != nil {
		return nil, err
	}

	cfg.Processing.MaxTokens = getInt("MAX_TOKENS", 24000)
	cfg.Processing.BudgetTokens = getInt("BUDGET_TOKENS", 16000)
	cfg.Processing.ThinkingType = getEnv("THINKING_TYPE", "enable")
	cfg.Processing.MaxRetries = getInt("MAX_RETRIES", 3)
	cfg.Processing.BedrockConnectTimeoutSec = getInt("BEDROCK_CONNECT_TIMEOUT_SEC", 10)
	cfg.Processing.BedrockCallTimeoutSec = getInt("BEDROCK_CALL_TIMEOUT_SEC", 30)
	cfg.Processing.DiscrepancyThreshold = getInt("DISCREPANCY_THRESHOLD", 5)

	cfg.Logging.Level = getEnv("LOG_LEVEL", "INFO")
	cfg.Logging.Format = getEnv("LOG_FORMAT", "json")

	cfg.Prompts.TemplateVersion = getEnv("TURN1_PROMPT_VERSION", "v1.0")
	cfg.Prompts.TemplateBasePath = getEnv("TEMPLATE_BASE_PATH", "/opt/templates")
	cfg.Prompts.Turn2TemplateVersion = getEnv("TURN2_PROMPT_VERSION", "v1.0")
	cfg.DatePartitionTimezone = getEnv("DATE_PARTITION_TIMEZONE", "UTC")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func mustGet(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", errors.NewConfigError("MissingEnv", key+" is required", key)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// CurrentDatePartition returns the current date partition in YYYY/MM/DD format
func (c *Config) CurrentDatePartition() string {
	loc, _ := time.LoadLocation(c.DatePartitionTimezone)
	now := time.Now().In(loc)
	return fmt.Sprintf("%04d/%02d/%02d", now.Year(), now.Month(), now.Day())
}

// DatePartitionFromTimestamp returns the date partition for a given timestamp
func (c *Config) DatePartitionFromTimestamp(ts string) (string, error) {
	loc, _ := time.LoadLocation(c.DatePartitionTimezone)
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "", err
	}
	t = t.In(loc)
	return fmt.Sprintf("%04d/%02d/%02d", t.Year(), t.Month(), t.Day()), nil
}

// IsThinkingEnabled returns true if thinking/reasoning mode is enabled
// Thinking is enabled only when THINKING_TYPE is explicitly set to "enable"
// Thinking is disabled when THINKING_TYPE is "disable" or unset (empty string)
func (c *Config) IsThinkingEnabled() bool {
	return strings.EqualFold(c.Processing.ThinkingType, "enable")
}
