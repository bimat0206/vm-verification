package config

import (
	"os"
	"strconv"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
)

// Config holds all environment variables for ExecuteTurn1.
type Config struct {
	AWSRegion              string
	BedrockModelID         string
	AnthropicVersion       string
	MaxTokens              int
	Temperature            float64
	ThinkingType           string
	ThinkingBudgetTokens   int

	EnableHybridStorage    bool
	TempBase64Bucket       string
	Base64SizeThreshold    int64
	Base64RetrievalTimeout time.Duration

	BedrockTimeout         time.Duration
	FunctionTimeout        time.Duration
}

// New loads and validates config from environment.
// All values must be provided via Lambda environment variables; no defaults are hardcoded.
func New(log logger.Logger) (*Config, error) {
	cfg := &Config{
		AWSRegion:              getenv("AWS_REGION", ""),
		BedrockModelID:         getenv("BEDROCK_MODEL", ""),
		AnthropicVersion:       getenv("ANTHROPIC_VERSION", ""),
		MaxTokens:              getenvInt("MAX_TOKENS", 0),
		Temperature:            getenvFloat("TEMPERATURE", 0),
		ThinkingType:           getenv("THINKING_TYPE", ""),
		ThinkingBudgetTokens:   getenvInt("BUDGET_TOKENS", 0),

		EnableHybridStorage:    getenvBool("ENABLE_HYBRID_STORAGE", false),
		TempBase64Bucket:       getenv("TEMP_BASE64_BUCKET", ""),
		Base64SizeThreshold:    getenvInt64("BASE64_SIZE_THRESHOLD", 0),
		Base64RetrievalTimeout: time.Duration(getenvInt("BASE64_RETRIEVAL_TIMEOUT", 0)) * time.Millisecond,

		BedrockTimeout:         time.Duration(getenvInt("BEDROCK_TIMEOUT", 0)) * time.Millisecond,
		FunctionTimeout:        time.Duration(getenvInt("FUNCTION_TIMEOUT", 0)) * time.Millisecond,
	}

	if err := cfg.Validate(log); err != nil {
		return nil, err
	}

	log.Info("Loaded configuration", map[string]interface{}{ 
		"AWSRegion":        cfg.AWSRegion, 
		"BedrockModelID":   cfg.BedrockModelID,
		"AnthropicVersion": cfg.AnthropicVersion,
	})
	return cfg, nil
}

// Validate checks for required config fields.
func (c *Config) Validate(log logger.Logger) error {
	var missing []string
	if c.AWSRegion == "" {
		missing = append(missing, "AWS_REGION")
	}
	if c.BedrockModelID == "" {
		missing = append(missing, "BEDROCK_MODEL")
	}
	if c.AnthropicVersion == "" {
		missing = append(missing, "ANTHROPIC_VERSION")
	}
	if c.EnableHybridStorage && c.TempBase64Bucket == "" {
		missing = append(missing, "TEMP_BASE64_BUCKET")
	}
	if len(missing) > 0 {
		log.Error("Missing required environment variables", map[string]interface{}{"vars": missing})
		return errors.NewValidationError("Missing environment variables", map[string]interface{}{"vars": missing})
	}
	return nil
}

// --- Helper functions for environment parsing ---
func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getenvInt64(key string, def int64) int64 {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return def
}

func getenvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
