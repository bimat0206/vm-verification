package config

import (
	//"fmt"
	"os"
	"strconv"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
	//"workflow-function/shared/schema"
)

// Config holds all config/environment variables for ExecuteTurn1.
type Config struct {
	AWSRegion             string
	BedrockModelID        string
	AnthropicVersion      string
	MaxTokens             int
	Temperature           float64
	ThinkingType          string
	ThinkingBudgetTokens  int

	EnableHybridStorage   bool
	TempBase64Bucket      string
	Base64SizeThreshold   int64
	Base64RetrievalTimeout time.Duration

	BedrockTimeout        time.Duration
	FunctionTimeout       time.Duration
}

// New loads and validates config from environment, logging all issues.
func New(log logger.Logger) (*Config, error) {
	cfg := &Config{
		AWSRegion:              getenv("AWS_REGION", "us-east-1"),
		BedrockModelID:         getenv("BEDROCK_MODEL_ID", "anthropic.claude-3-7-sonnet-20250219-v1:0"),
		AnthropicVersion:       getenv("ANTHROPIC_VERSION", "bedrock-2023-05-31"),
		MaxTokens:              getenvInt("MAX_TOKENS", 4000),
		Temperature:            getenvFloat("TEMPERATURE", 0.7),
		ThinkingType:           getenv("THINKING_TYPE", "thoroughness"),
		ThinkingBudgetTokens:   getenvInt("THINKING_BUDGET_TOKENS", 16000),
		EnableHybridStorage:    getenvBool("ENABLE_HYBRID_STORAGE", true),
		TempBase64Bucket:       getenv("TEMP_BASE64_BUCKET", "temp-base64-bucket"),
		Base64SizeThreshold:    getenvInt64("BASE64_SIZE_THRESHOLD", 2*1024*1024),
		Base64RetrievalTimeout: time.Duration(getenvInt("BASE64_RETRIEVAL_TIMEOUT", 30000)) * time.Millisecond,
		BedrockTimeout:         time.Duration(getenvInt("BEDROCK_TIMEOUT", 120000)) * time.Millisecond,
		FunctionTimeout:        time.Duration(getenvInt("FUNCTION_TIMEOUT", 120000)) * time.Millisecond,
	}

	if err := cfg.Validate(log); err != nil {
		return nil, err
	}
	log.Info("Loaded and validated configuration", map[string]interface{}{
		"AWSRegion": cfg.AWSRegion, "BedrockModelID": cfg.BedrockModelID,
		"HybridBase64": cfg.EnableHybridStorage, "TempBase64Bucket": cfg.TempBase64Bucket,
	})
	return cfg, nil
}

// Validate checks for required/invalid config fields and logs via the provided logger.
func (c *Config) Validate(log logger.Logger) error {
	if c.BedrockModelID == "" {
		log.Error("BEDROCK_MODEL_ID must be set", nil)
		return errors.NewValidationError("BEDROCK_MODEL_ID missing", nil)
	}
	if c.AWSRegion == "" {
		log.Error("AWS_REGION must be set", nil)
		return errors.NewValidationError("AWS_REGION missing", nil)
	}
	if c.EnableHybridStorage && c.TempBase64Bucket == "" {
		log.Error("TEMP_BASE64_BUCKET must be set when hybrid storage is enabled", nil)
		return errors.NewValidationError("TEMP_BASE64_BUCKET missing for hybrid storage", nil)
	}
	return nil
}

// --- Helper functions for env parsing ---

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
func getenvInt(key string, def int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return def
}
func getenvInt64(key string, def int64) int64 {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return def
}
func getenvFloat(key string, def float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return def
}
func getenvBool(key string, def bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return def
}
