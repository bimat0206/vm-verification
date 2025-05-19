// Package config provides configuration management for the FetchImages function
package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration from environment variables
type Config struct {
	// DynamoDB table names
	LayoutTable       string
	VerificationTable string
	
	// S3 state management
	StateBucket string
	
	// Maximum image size in bytes (optional, for validation)
	MaxImageSize int64
}

// New creates a new Config instance with values from environment variables and defaults
func New() Config {
	// Parse max image size with default of 100MB
	maxImageSizeStr := getEnvWithDefault("MAX_IMAGE_SIZE", "104857600")
	maxImageSize, err := strconv.ParseInt(maxImageSizeStr, 10, 64)
	if err != nil {
		// Use default on parse error
		maxImageSize = 104857600 // 100MB default
	}
	
	// Ensure reasonable limits (minimum 1MB, maximum 1GB)
	if maxImageSize < 1048576 { // 1MB
		maxImageSize = 1048576
	}
	if maxImageSize > 1073741824 { // 1GB
		maxImageSize = 1073741824
	}
	
	return Config{
		LayoutTable:       getEnvWithDefault("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata"),
		VerificationTable: getEnvWithDefault("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
		StateBucket:       getEnvWithDefault("STATE_BUCKET", ""),
		MaxImageSize:      maxImageSize,
	}
}

// Validate checks for required configuration values
func (c *Config) Validate() error {
	if c.StateBucket == "" {
		return NewConfigError("STATE_BUCKET is required but not set")
	}
	return nil
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

// NewConfigError creates a new ConfigError
func NewConfigError(message string) *ConfigError {
	return &ConfigError{Message: message}
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	return e.Message
}

// GetEnvironmentInfo returns a map of all relevant environment variables
func GetEnvironmentInfo() map[string]interface{} {
	return map[string]interface{}{
		"MAX_IMAGE_SIZE":             os.Getenv("MAX_IMAGE_SIZE"),
		"DYNAMODB_LAYOUT_TABLE":      os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		"DYNAMODB_VERIFICATION_TABLE": os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		"STATE_BUCKET":               os.Getenv("STATE_BUCKET"),
	}
}