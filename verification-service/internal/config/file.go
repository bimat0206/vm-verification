package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// loadFileConfig loads configuration from a JSON file
func loadFileConfig(filePath string) (*Config, error) {
	// Expand file path if it starts with ~
	if filePath[:1] == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to expand home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, filePath[1:])
	}

	// Read file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfigToFile saves the configuration to a JSON file
func SaveConfigToFile(cfg *Config, filePath string) error {
	// Expand file path if it starts with ~
	if filePath[:1] == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, filePath[1:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal config to JSON
	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfigFromFile loads configuration from environment variables and optionally overrides with values from a JSON file
func LoadConfigFromFile(filePath string) (*Config, error) {
	// First load from environment variables
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, just return env config
		return cfg, nil
	}

	// Load from file
	fileCfg, err := loadFileConfig(filePath)
	if err != nil {
		return nil, err
	}

	// Merge configurations (file takes precedence)
	mergedCfg := mergeConfigs(cfg, fileCfg)

	return mergedCfg, nil
}

// mergeConfigs merges two configurations, with the second taking precedence
func mergeConfigs(base, override *Config) *Config {
	// Create a new config with base values
	merged := *base

	// Override server settings if provided
	if override.Server.Port != 0 {
		merged.Server.Port = override.Server.Port
	}
	if override.Server.ReadTimeoutSecs != 0 {
		merged.Server.ReadTimeoutSecs = override.Server.ReadTimeoutSecs
	}
	if override.Server.WriteTimeoutSecs != 0 {
		merged.Server.WriteTimeoutSecs = override.Server.WriteTimeoutSecs
	}
	if override.Server.IdleTimeoutSecs != 0 {
		merged.Server.IdleTimeoutSecs = override.Server.IdleTimeoutSecs
	}

	// Override AWS settings if provided
	if override.AWS.Region != "" {
		merged.AWS.Region = override.AWS.Region
	}

	// Override DynamoDB settings if provided
	if override.DynamoDB.VerificationResultsTable != "" {
		merged.DynamoDB.VerificationResultsTable = override.DynamoDB.VerificationResultsTable
	}
	if override.DynamoDB.LayoutMetadataTable != "" {
		merged.DynamoDB.LayoutMetadataTable = override.DynamoDB.LayoutMetadataTable
	}

	// Override S3 settings if provided
	if override.S3.ReferenceBucket != "" {
		merged.S3.ReferenceBucket = override.S3.ReferenceBucket
	}
	if override.S3.CheckingBucket != "" {
		merged.S3.CheckingBucket = override.S3.CheckingBucket
	}
	if override.S3.ResultsBucket != "" {
		merged.S3.ResultsBucket = override.S3.ResultsBucket
	}

	// Override Bedrock settings if provided
	if override.Bedrock.ModelID != "" {
		merged.Bedrock.ModelID = override.Bedrock.ModelID
	}
	if override.Bedrock.MaxRetries != 0 {
		merged.Bedrock.MaxRetries = override.Bedrock.MaxRetries
	}

	// Override Notification settings if provided
	if override.Notification.SNSTopicARN != "" {
		merged.Notification.SNSTopicARN = override.Notification.SNSTopicARN
	}
	if override.Notification.WebhookURL != "" {
		merged.Notification.WebhookURL = override.Notification.WebhookURL
	}

	return &merged
}