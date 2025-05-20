// Package config provides configuration management for the ProcessTurn1Response Lambda
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	// AWS configuration
	AWSRegion string

	// State management configuration
	S3StateBucket string
	ConversationTable string
	
	// Processing configuration
	MaxResponseSize    int64
	ProcessingTimeout  time.Duration
	StrictValidation   bool
	FallbackParsing    bool
	
	// Logging configuration
	LogLevel     string
	ServiceName  string
	FunctionName string
}

// New creates a new configuration with values from environment variables
func New() *Config {
	return &Config{
		// AWS configuration
		AWSRegion: getEnvOrDefault("AWS_REGION", "us-east-1"),
		
		// State management configuration
		S3StateBucket:    getEnvOrDefault("S3_STATE_BUCKET", ""),
		ConversationTable: getEnvOrDefault("DYNAMODB_CONVERSATION_TABLE", ""),
		
		// Processing configuration
		MaxResponseSize:   getEnvAsInt64("MAX_RESPONSE_SIZE", 1024*1024), // 1MB default
		ProcessingTimeout: getEnvAsDuration("PROCESSING_TIMEOUT", "60s"),
		StrictValidation:  getEnvAsBool("STRICT_VALIDATION", false),
		FallbackParsing:   getEnvAsBool("FALLBACK_PARSING", true),
		
		// Logging configuration
		LogLevel:     getEnvOrDefault("LOG_LEVEL", "INFO"),
		ServiceName:  getEnvOrDefault("SERVICE_NAME", "verification-service"),
		FunctionName: getEnvOrDefault("FUNCTION_NAME", "ProcessTurn1Response"),
	}
}

// Validate checks if the configuration has all required values
func (c *Config) Validate() error {
	// Required configuration values
	if c.S3StateBucket == "" {
		return &ConfigError{
			Field:   "S3_STATE_BUCKET",
			Message: "S3 state bucket is required",
		}
	}
	
	return nil
}

// LogConfig logs the current configuration
func (c *Config) LogConfig(logger *slog.Logger) {
	logger.Info("Configuration loaded",
		"awsRegion", c.AWSRegion,
		"s3StateBucket", c.S3StateBucket,
		"conversationTable", c.ConversationTable,
		"maxResponseSize", c.MaxResponseSize,
		"processingTimeout", c.ProcessingTimeout,
		"strictValidation", c.StrictValidation,
		"fallbackParsing", c.FallbackParsing,
		"logLevel", c.LogLevel,
		"serviceName", c.ServiceName,
		"functionName", c.FunctionName,
	)
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

// Error returns a string representation of the error
func (e *ConfigError) Error() string {
	return e.Message + " (field: " + e.Field + ")"
}

// Helper functions for environment variable parsing

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 gets an environment variable as an int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsedVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsedVal
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as a duration
func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

// getEnvAsBool gets an environment variable as a boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "yes", "1", "t", "y":
			return true
		case "false", "no", "0", "f", "n":
			return false
		}
	}
	return defaultValue
}