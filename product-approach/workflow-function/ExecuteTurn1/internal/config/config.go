package config

import (
	"os"
	"strconv"
	"time"

	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// Config holds all environment variables for ExecuteTurn1.
// Enhanced with S3 state management and improved categorization of configuration
type Config struct {
	// S3 State Management
	StateBucket         string        `env:"STATE_BUCKET,required"`
	StateBasePrefix     string        `env:"STATE_BASE_PREFIX" envDefault:"verification-states"`
	StateTimeout        time.Duration `env:"STATE_TIMEOUT" envDefault:"30s"`
	
	// Bedrock Configuration  
	BedrockModel        string        `env:"BEDROCK_MODEL,required"` // Changed from BEDROCK_MODEL_ID
	BedrockRegion       string        `env:"BEDROCK_REGION,required"`
	AnthropicVersion    string        `env:"ANTHROPIC_VERSION,required"`
	MaxTokens           int           `env:"MAX_TOKENS" envDefault:"4096"`
	Temperature         float64       `env:"TEMPERATURE" envDefault:"0.7"`
	ThinkingType        string        `env:"THINKING_TYPE" envDefault:"thinking"`
	ThinkingBudgetTokens int          `env:"THINKING_BUDGET_TOKENS" envDefault:"16000"`
	
	// Image Processing Configuration
	Base64SizeThreshold int64         `env:"BASE64_SIZE_THRESHOLD" envDefault:"1048576"` // 1MB default
	
	// Timeouts and Performance Settings
	BedrockTimeout      time.Duration `env:"BEDROCK_TIMEOUT" envDefault:"120s"`
	FunctionTimeout     time.Duration `env:"FUNCTION_TIMEOUT" envDefault:"240s"`
	RetryMaxAttempts    int           `env:"RETRY_MAX_ATTEMPTS" envDefault:"3"`
	RetryBaseDelay      time.Duration `env:"RETRY_BASE_DELAY" envDefault:"1s"`
}

// New loads and validates config from environment with sensible defaults.
func New(log logger.Logger) (*Config, error) {
	cfg := &Config{
		// S3 State Management
		StateBucket:         getenv("STATE_BUCKET", ""),
		StateBasePrefix:     getenv("STATE_BASE_PREFIX", "verification-states"),
		StateTimeout:        time.Duration(getenvInt("STATE_TIMEOUT", 30)) * time.Second,
		
		// Bedrock Configuration
		BedrockModel:        getenv("BEDROCK_MODEL", ""), // Changed from BEDROCK_MODEL_ID
		BedrockRegion:       getenv("BEDROCK_REGION", getenv("AWS_REGION", "")),
		AnthropicVersion:    getenv("ANTHROPIC_VERSION", "bedrock-2023-05-31"),
		MaxTokens:           getenvInt("MAX_TOKENS", 4096),
		Temperature:         getenvFloat("TEMPERATURE", 0.7),
		ThinkingType:        getenv("THINKING_TYPE", "thinking"),
		ThinkingBudgetTokens: getenvInt("THINKING_BUDGET_TOKENS", 16000),
		
		// Image Processing Configuration
		Base64SizeThreshold: getenvInt64("BASE64_SIZE_THRESHOLD", 1048576), // 1MB default
		
		// Timeouts and Performance Settings
		BedrockTimeout:      time.Duration(getenvInt("BEDROCK_TIMEOUT", 120)) * time.Second,
		FunctionTimeout:     time.Duration(getenvInt("FUNCTION_TIMEOUT", 240)) * time.Second,
		RetryMaxAttempts:    getenvInt("RETRY_MAX_ATTEMPTS", 3),
		RetryBaseDelay:      time.Duration(getenvInt("RETRY_BASE_DELAY", 1)) * time.Second,
	}

	if err := cfg.Validate(log); err != nil {
		return nil, err
	}

	log.Info("Loaded configuration", map[string]interface{}{ 
		"stateBucket":       cfg.StateBucket,
		"bedrockModel":      cfg.BedrockModel, // Changed from bedrockModelID
		"bedrockRegion":     cfg.BedrockRegion,
		"anthropicVersion":  cfg.AnthropicVersion,
		"thinkingEnabled":   cfg.ThinkingType != "",
	})
	
	return cfg, nil
}

// Validate checks for required config fields.
func (c *Config) Validate(log logger.Logger) error {
	var missing []string
	
	// Critical S3 state bucket
	if c.StateBucket == "" {
		missing = append(missing, "STATE_BUCKET")
	}
	
	// Required Bedrock configuration
	if c.BedrockModel == "" {
		missing = append(missing, "BEDROCK_MODEL") // Changed from BEDROCK_MODEL_ID
	}
	
	if c.BedrockRegion == "" {
		missing = append(missing, "BEDROCK_REGION or AWS_REGION")
	}
	
	if c.AnthropicVersion == "" {
		missing = append(missing, "ANTHROPIC_VERSION")
	}
	
	if len(missing) > 0 {
		log.Error("Missing required environment variables", map[string]interface{}{"vars": missing})
		return wferrors.NewValidationError("Missing environment variables", map[string]interface{}{"vars": missing})
	}
	
	// Validate reasonable values
	if c.MaxTokens <= 0 {
		log.Warn("Setting default MaxTokens=4096 as configured value was invalid", 
			map[string]interface{}{"configuredValue": c.MaxTokens})
		c.MaxTokens = 4096
	}
	
	if c.BedrockTimeout < 10*time.Second {
		log.Warn("BedrockTimeout is very short, setting to minimum of 30 seconds", nil)
		c.BedrockTimeout = 30 * time.Second
	}
	
	return nil
}

// GetHybridStorageConfig returns the hybrid storage configuration for S3/Base64 handling
func (c *Config) GetHybridStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":          false, // Set to false since we removed the config option
		"tempBucket":       "", // Empty since we removed the config option  
		"sizeThreshold":    c.Base64SizeThreshold,
		"retrievalTimeout": c.StateTimeout.Milliseconds(),
	}
}

// GetBedrockConfig returns the Bedrock API configuration
func (c *Config) GetBedrockConfig() map[string]interface{} {
	return map[string]interface{}{
		"modelId":          c.BedrockModel, // Changed from c.BedrockModelID
		"region":           c.BedrockRegion,
		"anthropicVersion": c.AnthropicVersion,
		"maxTokens":        c.MaxTokens,
		"temperature":      c.Temperature,
		"thinkingEnabled":  c.ThinkingType != "",
		"budgetTokens":     c.ThinkingBudgetTokens,
		"timeout":          c.BedrockTimeout.Milliseconds(),
	}
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