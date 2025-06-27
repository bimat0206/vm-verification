package validation

import (
	//"fmt"
	"strings"
	"net/url"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// ValidateS3URL validates that a URL is a properly formatted S3 URL
func (v *Validator) ValidateS3URL(s3URL string) (string, string, error) {
	if s3URL == "" {
		return "", "", errors.NewValidationError("S3 URL is empty", nil)
	}

	// Parse S3 URL (format: s3://bucket/key)
	if !strings.HasPrefix(s3URL, "s3://") {
		return "", "", errors.NewValidationError("Invalid S3 URL format, must start with s3://", 
			map[string]interface{}{"url": s3URL})
	}

	// Remove the s3:// prefix
	s3Path := strings.TrimPrefix(s3URL, "s3://")
	
	// Split into bucket and key parts
	parts := strings.SplitN(s3Path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.NewValidationError("Invalid S3 URL format, missing key", 
			map[string]interface{}{"url": s3URL})
	}
	
	bucket := parts[0]
	key := parts[1]
	
	// Basic bucket validation
	if bucket == "" {
		return "", "", errors.NewValidationError("S3 bucket name is empty", 
			map[string]interface{}{"url": s3URL})
	}
	
	// Basic key validation
	if key == "" {
		return "", "", errors.NewValidationError("S3 key is empty", 
			map[string]interface{}{"url": s3URL})
	}
	
	return bucket, key, nil
}

// ValidateHTTPURL validates that a URL is a properly formatted HTTP URL
func (v *Validator) ValidateHTTPURL(httpURL string) error {
	if httpURL == "" {
		return errors.NewValidationError("HTTP URL is empty", nil)
	}

	// Parse the URL
	parsedURL, err := url.Parse(httpURL)
	if err != nil {
		return errors.NewValidationError("Invalid URL format", 
			map[string]interface{}{
				"url":   httpURL,
				"error": err.Error(),
			})
	}

	// Check that it's HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.NewValidationError("URL must use HTTP or HTTPS scheme", 
			map[string]interface{}{
				"url":    httpURL,
				"scheme": parsedURL.Scheme,
			})
	}

	// Check that there's a host
	if parsedURL.Host == "" {
		return errors.NewValidationError("URL is missing host", 
			map[string]interface{}{"url": httpURL})
	}

	return nil
}

// ValidateBedrockConfig validates the Bedrock configuration
func (v *Validator) ValidateBedrockConfig(config *schema.BedrockConfig) error {
	if config == nil {
		return errors.NewValidationError("Bedrock configuration is nil", nil)
	}

	// Validate Anthropic version
	if config.AnthropicVersion == "" {
		return errors.NewValidationError("Anthropic version is required", nil)
	}

	// Validate max tokens
	if config.MaxTokens <= 0 {
		return errors.NewValidationError("Max tokens must be positive", 
			map[string]interface{}{"maxTokens": config.MaxTokens})
	}

	// Validate thinking configuration if provided
	if config.Thinking != nil {
		if config.Thinking.Type != "enabled" && config.Thinking.Type != "disabled" {
			return errors.NewValidationError("Thinking type must be 'enabled' or 'disabled'", 
				map[string]interface{}{"type": config.Thinking.Type})
		}

		if config.Thinking.Type == "enabled" && config.Thinking.BudgetTokens <= 0 {
			return errors.NewValidationError("Budget tokens must be positive when thinking is enabled", 
				map[string]interface{}{"budgetTokens": config.Thinking.BudgetTokens})
		}
	}

	return nil
}

// ValidateCurrentPrompt validates the current prompt
func (v *Validator) ValidateCurrentPrompt(prompt *schema.CurrentPrompt) error {
	if prompt == nil {
		return errors.NewValidationError("Current prompt is nil", nil)
	}

	// Validate turn number (must be 1 for Turn 1)
	if prompt.TurnNumber != 1 {
		return errors.NewInvalidFieldError("currentPrompt.turnNumber", prompt.TurnNumber, "1")
	}

	// Validate include image (must be "reference" for Turn 1)
	if prompt.IncludeImage != "reference" {
		return errors.NewInvalidFieldError("currentPrompt.includeImage", prompt.IncludeImage, "reference")
	}

	// Validate messages (must have at least one for Turn 1)
	if len(prompt.Messages) == 0 {
		return errors.NewValidationError("Current prompt has no messages", nil)
	}

	return nil
}