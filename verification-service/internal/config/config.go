package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"verification-service/internal/app/services"
)

// Config is the main configuration structure for the application
type Config struct {
	// Server configuration
	Server struct {
		Port            int `json:"port"`
		ReadTimeoutSecs int `json:"readTimeoutSecs"`
		WriteTimeoutSecs int `json:"writeTimeoutSecs"`
		IdleTimeoutSecs int `json:"idleTimeoutSecs"`
	} `json:"server"`

	// AWS configuration
	AWS struct {
		Region string `json:"region"`
	} `json:"aws"`

	// DynamoDB configuration
	DynamoDB struct {
		VerificationResultsTable string `json:"verificationResultsTable"`
		LayoutMetadataTable      string `json:"layoutMetadataTable"`
	} `json:"dynamodb"`

	// S3 configuration
	S3 struct {
		ReferenceBucket string `json:"referenceBucket"`
		CheckingBucket  string `json:"checkingBucket"`
		ResultsBucket   string `json:"resultsBucket"`
	} `json:"s3"`

	// Bedrock configuration
	Bedrock struct {
		ModelID    string `json:"modelId"`
		MaxRetries int    `json:"maxRetries"`
	} `json:"bedrock"`

	// Notification configuration
	Notification services.NotificationConfig `json:"notification"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = getenvInt("SERVER_PORT", 3000)
	cfg.Server.ReadTimeoutSecs = getenvInt("SERVER_READ_TIMEOUT_SECS", 30)
	cfg.Server.WriteTimeoutSecs = getenvInt("SERVER_WRITE_TIMEOUT_SECS", 30)
	cfg.Server.IdleTimeoutSecs = getenvInt("SERVER_IDLE_TIMEOUT_SECS", 60)

	// AWS configuration
	cfg.AWS.Region = getenv("AWS_REGION", "us-east-1")

	// DynamoDB configuration
	cfg.DynamoDB.VerificationResultsTable = getenv("DYNAMODB_VERIFICATION_TABLE", "VerificationResults")
	cfg.DynamoDB.LayoutMetadataTable = getenv("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata")

	// S3 configuration
	cfg.S3.ReferenceBucket = getenv("S3_REFERENCE_BUCKET", "kootoro-reference-bucket")
	cfg.S3.CheckingBucket = getenv("S3_CHECKING_BUCKET", "kootoro-checking-bucket")
	cfg.S3.ResultsBucket = getenv("S3_RESULTS_BUCKET", "kootoro-results-bucket")

	// Bedrock configuration
	cfg.Bedrock.ModelID = getenv("BEDROCK_MODEL_ID", "anthropic.claude-3-7-sonnet-20250219")
	cfg.Bedrock.MaxRetries = getenvInt("BEDROCK_MAX_RETRIES", 3)

	// Notification configuration
	cfg.Notification.SNSTopicARN = getenv("NOTIFICATION_SNS_TOPIC_ARN", "")
	cfg.Notification.WebhookURL = getenv("NOTIFICATION_WEBHOOK_URL", "")

	// Validate required configuration
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateConfig validates the loaded configuration
func validateConfig(cfg *Config) error {
	// Validate AWS region
	if cfg.AWS.Region == "" {
		return fmt.Errorf("AWS_REGION is required")
	}

	// Validate DynamoDB tables
	if cfg.DynamoDB.VerificationResultsTable == "" {
		return fmt.Errorf("DYNAMODB_VERIFICATION_TABLE is required")
	}
	if cfg.DynamoDB.LayoutMetadataTable == "" {
		return fmt.Errorf("DYNAMODB_LAYOUT_TABLE is required")
	}

	// Validate S3 buckets
	if cfg.S3.ReferenceBucket == "" {
		return fmt.Errorf("S3_REFERENCE_BUCKET is required")
	}
	if cfg.S3.CheckingBucket == "" {
		return fmt.Errorf("S3_CHECKING_BUCKET is required")
	}
	if cfg.S3.ResultsBucket == "" {
		return fmt.Errorf("S3_RESULTS_BUCKET is required")
	}

	// Validate Bedrock
	if cfg.Bedrock.ModelID == "" {
		return fmt.Errorf("BEDROCK_MODEL_ID is required")
	}

	return nil
}

// getenv gets an environment variable or returns a default value
func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getenvInt gets an environment variable as an integer or returns a default value
func getenvInt(key string, defaultValue int) int {
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

// getenvBool gets an environment variable as a boolean or returns a default value
func getenvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}