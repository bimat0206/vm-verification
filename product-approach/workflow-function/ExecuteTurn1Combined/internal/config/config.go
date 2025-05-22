// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

// Config bundles all environment settings required by the Lambda.
type Config struct {
	AWSRegion                   string
	StateBucket                 string
	BedrockModel                string
	AnthropicVersion            string
	MaxTokens                   int
	BudgetTokens                int
	ThinkingType                string
	DynamoDBVerificationTable   string
	DynamoDBConversationTable   string
	LogLevel                    string
	TemplateBasePath            string
	TURN1PromptVersion          string
}

// Load parses environment variables into a Config structure.
// It applies sensible defaults where optional values are absent.
func Load() Config {
	return Config{
		AWSRegion:                 getEnv("AWS_REGION", "us-east-1"),
		StateBucket:               mustGet("STATE_BUCKET"),
		BedrockModel:              mustGet("BEDROCK_MODEL"),
		AnthropicVersion:          getEnv("ANTHROPIC_VERSION", "bedrock-2023-05-31"),
		MaxTokens:                 getInt("MAX_TOKENS", 24000),
		BudgetTokens:              getInt("BUDGET_TOKENS", 16000),
		ThinkingType:              getEnv("THINKING_TYPE", "enable"),
		DynamoDBVerificationTable: mustGet("DYNAMODB_VERIFICATION_TABLE"),
		DynamoDBConversationTable: mustGet("DYNAMODB_CONVERSATION_TABLE"),
		LogLevel:                  getEnv("LOG_LEVEL", "INFO"),
		TemplateBasePath:          getEnv("TEMPLATE_BASE_PATH", "/opt/templates"),
		TURN1PromptVersion:        getEnv("TURN1_PROMPT_VERSION", "v1.0"),
	}
}

// mustGet retrieves the env-var value or panics (fail-fast) if missing.
func mustGet(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic("missing required env var: " + key)
}

// getEnv returns the env-var value or the provided default when not present.
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getInt returns an integer env-var value or the supplied default.
func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
