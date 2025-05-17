package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration settings for the ExecuteTurn1 function
type Config struct {
	// Bedrock configuration
	BedrockModelID     string
	AWSRegion          string
	AnthropicVersion   string
	MaxTokens          int
	Temperature        float64
	ThinkingType       string
	ThinkingBudgetTokens int

	// Hybrid storage configuration
	EnableHybridStorage     bool
	TempBase64Bucket        string
	Base64SizeThreshold     int64
	Base64RetrievalTimeout  time.Duration

	// Timeouts
	BedrockTimeout          time.Duration
	FunctionTimeout         time.Duration
}

// New creates a new Config with values loaded from environment variables
func New() (*Config, error) {
	config := &Config{}

	// Required configurations
	config.BedrockModelID = getEnvOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-3-7-sonnet-20250219-v1:0")
	config.AWSRegion = getEnvOrDefault("AWS_REGION", "us-east-1")

	// Bedrock configuration
	config.AnthropicVersion = getEnvOrDefault("ANTHROPIC_VERSION", "bedrock-2023-05-31")
	config.MaxTokens = getEnvAsIntOrDefault("MAX_TOKENS", 4000)
	config.Temperature = getEnvAsFloatOrDefault("TEMPERATURE", 0.7)
	config.ThinkingType = getEnvOrDefault("THINKING_TYPE", "thoroughness")
	config.ThinkingBudgetTokens = getEnvAsIntOrDefault("THINKING_BUDGET_TOKENS", 50000)

	// Hybrid storage configuration
	config.EnableHybridStorage = getEnvAsBoolOrDefault("ENABLE_HYBRID_STORAGE", true)
	config.TempBase64Bucket = getEnvOrDefault("TEMP_BASE64_BUCKET", "temp-base64-bucket")
	config.Base64SizeThreshold = getEnvAsInt64OrDefault("BASE64_SIZE_THRESHOLD", 2*1024*1024) // 2MB
	config.Base64RetrievalTimeout = time.Duration(getEnvAsIntOrDefault("BASE64_RETRIEVAL_TIMEOUT", 30000)) * time.Millisecond

	// Timeouts
	config.BedrockTimeout = time.Duration(getEnvAsIntOrDefault("BEDROCK_TIMEOUT", 300000)) * time.Millisecond // 5 minutes
	config.FunctionTimeout = time.Duration(getEnvAsIntOrDefault("FUNCTION_TIMEOUT", 300000)) * time.Millisecond // 5 minutes

	return config, config.Validate()
}

// Validate checks that all required configuration is present and valid
func (c *Config) Validate() error {
	if c.BedrockModelID == "" {
		return fmt.Errorf("BEDROCK_MODEL_ID must be set")
	}
	if c.AWSRegion == "" {
		return fmt.Errorf("AWS_REGION must be set")
	}
	if c.EnableHybridStorage && c.TempBase64Bucket == "" {
		return fmt.Errorf("TEMP_BASE64_BUCKET must be set when ENABLE_HYBRID_STORAGE is true")
	}
	return nil
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64OrDefault(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}