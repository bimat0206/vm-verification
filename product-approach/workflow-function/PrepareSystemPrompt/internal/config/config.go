package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Environment variable names
	EnvReferenceBucket      = "REFERENCE_BUCKET"
	EnvCheckingBucket       = "CHECKING_BUCKET"
	EnvStateBucket          = "STATE_BUCKET"
	EnvTemplateBasePath     = "TEMPLATE_BASE_PATH"
	EnvComponentName        = "COMPONENT_NAME"
	EnvDatePartitionTimezone = "DATE_PARTITION_TIMEZONE"
	EnvDebug                 = "DEBUG"
	EnvMaxTokens             = "MAX_TOKENS"
	EnvBudgetTokens          = "BUDGET_TOKENS"
	EnvPromptVersion         = "PROMPT_VERSION"
	
	// Default values
	DefaultTemplateBasePath     = "/opt/templates"
	DefaultComponentName        = "PrepareSystemPrompt"
	DefaultDatePartitionTimezone = "UTC"
	DefaultMaxTokens             = 24000
	DefaultBudgetTokens          = 16000
	DefaultPromptVersion         = "1.0.0"
)

// Config represents the application configuration
type Config struct {
	// S3 Buckets
	StateBucket      string
	ReferenceBucket  string
	CheckingBucket   string
	
	// Template settings
	TemplateBasePath string
	PromptVersion    string
	
	// Bedrock settings
	MaxTokens        int
	BudgetTokens     int
	
	// Application settings
	ComponentName        string
	DatePartitionTimezone string
	Debug                bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		StateBucket:          getEnv(EnvStateBucket, ""),
		ReferenceBucket:      getEnv(EnvReferenceBucket, ""),
		CheckingBucket:       getEnv(EnvCheckingBucket, ""),
		TemplateBasePath:     getEnv(EnvTemplateBasePath, DefaultTemplateBasePath),
		ComponentName:        getEnv(EnvComponentName, DefaultComponentName),
		DatePartitionTimezone: getEnv(EnvDatePartitionTimezone, DefaultDatePartitionTimezone),
		MaxTokens:            getIntEnv(EnvMaxTokens, DefaultMaxTokens),
		BudgetTokens:         getIntEnv(EnvBudgetTokens, DefaultBudgetTokens),
		PromptVersion:        getEnv(EnvPromptVersion, DefaultPromptVersion),
		Debug:                getBoolEnv(EnvDebug, false),
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Required fields
	if c.StateBucket == "" {
		return fmt.Errorf("STATE_BUCKET environment variable is required")
	}
	
	if c.ReferenceBucket == "" {
		return fmt.Errorf("REFERENCE_BUCKET environment variable is required")
	}
	
	if c.CheckingBucket == "" {
		return fmt.Errorf("CHECKING_BUCKET environment variable is required")
	}
	
	// Validate timezone
	_, err := time.LoadLocation(c.DatePartitionTimezone)
	if err != nil {
		return fmt.Errorf("invalid timezone in DATE_PARTITION_TIMEZONE: %s", c.DatePartitionTimezone)
	}
	
	// Check token limits
	if c.MaxTokens <= 0 {
		return fmt.Errorf("MAX_TOKENS must be positive")
	}
	
	if c.BudgetTokens <= 0 {
		return fmt.Errorf("BUDGET_TOKENS must be positive")
	}
	
	if c.BudgetTokens > c.MaxTokens {
		return fmt.Errorf("BUDGET_TOKENS (%d) cannot be greater than MAX_TOKENS (%d)", 
			c.BudgetTokens, c.MaxTokens)
	}
	
	return nil
}

// CurrentDatePartition returns the current date partition in format YYYY/MM/DD
func (c *Config) CurrentDatePartition() string {
	loc, _ := time.LoadLocation(c.DatePartitionTimezone)
	now := time.Now().In(loc)
	return fmt.Sprintf("%04d/%02d/%02d", now.Year(), now.Month(), now.Day())
}

// DatePartitionFromTimestamp returns a date partition from an ISO8601 timestamp
func (c *Config) DatePartitionFromTimestamp(timestamp string) (string, error) {
	loc, _ := time.LoadLocation(c.DatePartitionTimezone)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "", fmt.Errorf("invalid timestamp format: %w", err)
	}
	
	t = t.In(loc)
	return fmt.Sprintf("%04d/%02d/%02d", t.Year(), t.Month(), t.Day()), nil
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv retrieves an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}

// getBoolEnv retrieves a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	// Convert to lowercase for consistent comparison
	value = strings.ToLower(value)
	
	// Check for true values
	if value == "true" || value == "1" || value == "yes" || value == "y" {
		return true
	}
	
	// Check for false values
	if value == "false" || value == "0" || value == "no" || value == "n" {
		return false
	}
	
	// Default to provided default if value doesn't match known patterns
	return defaultValue
}