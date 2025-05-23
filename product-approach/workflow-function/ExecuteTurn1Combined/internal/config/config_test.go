package config

import (
	"os"
	"testing"
	"workflow-function/shared/errors"
)

func TestLoadConfiguration_MissingRequiredEnvVars(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"STATE_BUCKET":               os.Getenv("STATE_BUCKET"),
		"BEDROCK_MODEL":              os.Getenv("BEDROCK_MODEL"),
		"DYNAMODB_VERIFICATION_TABLE": os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		"DYNAMODB_CONVERSATION_TABLE": os.Getenv("DYNAMODB_CONVERSATION_TABLE"),
	}

	// Clear env vars
	os.Clearenv()

	// Test missing STATE_BUCKET
	_, err := LoadConfiguration()
	if err == nil {
		t.Fatal("Expected error for missing STATE_BUCKET, got nil")
	}

	// Verify it's a WorkflowError
	workflowErr, ok := err.(*errors.WorkflowError)
	if !ok {
		t.Fatalf("Expected WorkflowError, got %T", err)
	}

	// Verify error type
	if workflowErr.Type != errors.ErrorTypeConfig {
		t.Errorf("Expected error type Config, got %s", workflowErr.Type)
	}

	// Verify error code
	if workflowErr.Code != "MissingEnv" {
		t.Errorf("Expected error code MissingEnv, got %s", workflowErr.Code)
	}

	// Verify error message
	if workflowErr.Message != "STATE_BUCKET is required" {
		t.Errorf("Expected error message 'STATE_BUCKET is required', got %s", workflowErr.Message)
	}

	// Verify variable in details
	if varName, ok := workflowErr.Details["variable"].(string); !ok || varName != "STATE_BUCKET" {
		t.Errorf("Expected variable STATE_BUCKET in details, got %v", workflowErr.Details["variable"])
	}

	// Verify IsConfigError helper
	if !errors.IsConfigError(err) {
		t.Error("IsConfigError should return true for config errors")
	}

	// Test with STATE_BUCKET set but missing BEDROCK_MODEL
	os.Setenv("STATE_BUCKET", "test-bucket")
	_, err = LoadConfiguration()
	if err == nil {
		t.Fatal("Expected error for missing BEDROCK_MODEL, got nil")
	}

	workflowErr, ok = err.(*errors.WorkflowError)
	if !ok {
		t.Fatalf("Expected WorkflowError, got %T", err)
	}

	if workflowErr.Message != "BEDROCK_MODEL is required" {
		t.Errorf("Expected error message 'BEDROCK_MODEL is required', got %s", workflowErr.Message)
	}

	// Test with all required vars set
	os.Setenv("STATE_BUCKET", "test-bucket")
	os.Setenv("BEDROCK_MODEL", "test-model")
	os.Setenv("DYNAMODB_VERIFICATION_TABLE", "test-verification-table")
	os.Setenv("DYNAMODB_CONVERSATION_TABLE", "test-conversation-table")

	cfg, err := LoadConfiguration()
	if err != nil {
		t.Fatalf("Expected no error with all required vars set, got %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected valid config, got nil")
	}

	// Verify config values
	if cfg.AWS.S3Bucket != "test-bucket" {
		t.Errorf("Expected S3Bucket 'test-bucket', got %s", cfg.AWS.S3Bucket)
	}

	// Restore original env vars
	for key, value := range originalVars {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}

func TestLoadConfiguration_DefaultValues(t *testing.T) {
	// Set required env vars
	os.Setenv("STATE_BUCKET", "test-bucket")
	os.Setenv("BEDROCK_MODEL", "test-model")
	os.Setenv("DYNAMODB_VERIFICATION_TABLE", "test-verification-table")
	os.Setenv("DYNAMODB_CONVERSATION_TABLE", "test-conversation-table")

	cfg, err := LoadConfiguration()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test default values
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("Expected default region 'us-east-1', got %s", cfg.AWS.Region)
	}

	if cfg.AWS.AnthropicVersion != "bedrock-2023-05-31" {
		t.Errorf("Expected default anthropic version 'bedrock-2023-05-31', got %s", cfg.AWS.AnthropicVersion)
	}

	if cfg.Processing.MaxTokens != 24000 {
		t.Errorf("Expected default max tokens 24000, got %d", cfg.Processing.MaxTokens)
	}

	if cfg.Processing.BudgetTokens != 16000 {
		t.Errorf("Expected default budget tokens 16000, got %d", cfg.Processing.BudgetTokens)
	}

	if cfg.Processing.ThinkingType != "enable" {
		t.Errorf("Expected default thinking type 'enable', got %s", cfg.Processing.ThinkingType)
	}

	if cfg.Processing.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", cfg.Processing.MaxRetries)
	}

	if cfg.Logging.Level != "INFO" {
		t.Errorf("Expected default log level 'INFO', got %s", cfg.Logging.Level)
	}
}