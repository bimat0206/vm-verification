package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"workflow-function/shared/errors"
)

// Config holds environment configuration for ExecuteTurn2Combined
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
	}
	Logging struct {
		Level  string
		Format string
	}
	Prompts struct {
		TemplateVersion  string
		TemplateBasePath string
	}
	DatePartitionTimezone string
}

// LoadConfiguration reads environment variables into Config
func LoadConfiguration() (*Config, error) {
	cfg := &Config{}
	cfg.AWS.Region = getEnv("AWS_REGION", "us-east-1")
	var err error
	if cfg.AWS.S3Bucket, err = mustGet("STATE_BUCKET"); err != nil {
		return nil, err
	}
	if cfg.AWS.BedrockModel, err = mustGet("BEDROCK_MODEL"); err != nil {
		return nil, err
	}
	cfg.AWS.AnthropicVersion = getEnv("ANTHROPIC_VERSION", "bedrock-2023-05-31")
	if cfg.AWS.DynamoDBVerificationTable, err = mustGet("DYNAMODB_VERIFICATION_TABLE"); err != nil {
		return nil, err
	}
	if cfg.AWS.DynamoDBConversationTable, err = mustGet("DYNAMODB_CONVERSATION_TABLE"); err != nil {
		return nil, err
	}
	cfg.Processing.MaxTokens = getInt("MAX_TOKENS", 24000)
	cfg.Processing.BudgetTokens = getInt("BUDGET_TOKENS", 16000)
	cfg.Processing.ThinkingType = getEnv("THINKING_TYPE", "enable")
	cfg.Processing.MaxRetries = getInt("MAX_RETRIES", 3)
	cfg.Processing.BedrockConnectTimeoutSec = getInt("BEDROCK_CONNECT_TIMEOUT_SEC", 10)
	cfg.Processing.BedrockCallTimeoutSec = getInt("BEDROCK_CALL_TIMEOUT_SEC", 30)
	cfg.Logging.Level = getEnv("LOG_LEVEL", "INFO")
	cfg.Logging.Format = getEnv("LOG_FORMAT", "json")
	cfg.Prompts.TemplateVersion = getEnv("TURN2_PROMPT_VERSION", "v1.0")
	cfg.Prompts.TemplateBasePath = getEnv("TEMPLATE_BASE_PATH", "/opt/templates")
	cfg.DatePartitionTimezone = getEnv("DATE_PARTITION_TIMEZONE", "UTC")

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

func (c *Config) Validate() error {
	if c.Processing.BedrockConnectTimeoutSec <= 0 || c.Processing.BedrockCallTimeoutSec <= 0 {
		return errors.NewConfigError("BedrockTimeoutInvalid", "timeouts must be positive", "BEDROCK_CONNECT_TIMEOUT_SEC")
	}
	if c.Processing.BedrockCallTimeoutSec <= c.Processing.BedrockConnectTimeoutSec {
		return errors.NewConfigError("BedrockTimeoutInvalid", "call timeout must be greater than connect timeout", "BEDROCK_CALL_TIMEOUT_SEC")
	}
	return nil
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
