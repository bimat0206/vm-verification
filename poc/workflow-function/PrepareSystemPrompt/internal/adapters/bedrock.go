package adapters

import (
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	
	"workflow-function/PrepareSystemPrompt/internal/config"
)

// BedrockAdapter handles Bedrock configuration and integration
type BedrockAdapter struct {
	config *config.Config
	logger logger.Logger
}

// NewBedrockAdapter creates a new Bedrock adapter
func NewBedrockAdapter(cfg *config.Config, log logger.Logger) *BedrockAdapter {
	return &BedrockAdapter{
		config: cfg,
		logger: log,
	}
}

// ConfigureBedrockSettings creates a Bedrock configuration
func (b *BedrockAdapter) ConfigureBedrockSettings() *schema.BedrockConfig {
	// Use configuration from the config object
	return &schema.BedrockConfig{
		AnthropicVersion: b.config.AnthropicVersion,
		MaxTokens:        b.config.MaxTokens,
		Thinking: &schema.Thinking{
			Type:         "enabled",
			BudgetTokens: b.config.BudgetTokens,
		},
	}
}

// GeneratePromptID generates a unique prompt ID
func (b *BedrockAdapter) GeneratePromptID(verificationID string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return "prompt-" + verificationID + "-" + timestamp
}

// EstimateTokenUsage estimates the token count for a text
func (b *BedrockAdapter) EstimateTokenUsage(text string) int {
	// Simple estimation: approximately 4 characters per token for English text
	return len(text) / 4
}

// GetModelId returns the configured Bedrock model ID
func (b *BedrockAdapter) GetModelId() string {
	return b.config.BedrockModel
}

// CreateSystemPrompt creates a system prompt object
func (b *BedrockAdapter) CreateSystemPrompt(content, promptVersion string, verificationID string) *schema.SystemPrompt {
	// Configure Bedrock
	bedrockConfig := b.ConfigureBedrockSettings()
	
	// Generate unique prompt ID
	promptID := b.GeneratePromptID(verificationID)
	
	// Create system prompt with metadata
	// Create system prompt without metadata as it's not in the schema
	return &schema.SystemPrompt{
		Content:       content,
		BedrockConfig: bedrockConfig,
		PromptId:      promptID,
		PromptVersion: promptVersion,
	}
}

// CreateCompleteSystemPrompt creates a complete system prompt object with full structure
func (b *BedrockAdapter) CreateCompleteSystemPrompt(content, promptVersion string, verificationContext *schema.VerificationContext) *schema.CompleteSystemPrompt {
	// First create the basic system prompt
	basicPrompt := b.CreateSystemPrompt(content, promptVersion, verificationContext.VerificationId)
	
	// Convert to complete system prompt using the configurable conversion function
	completePrompt := schema.ConvertToCompleteSystemPromptWithConfig(basicPrompt, verificationContext, b.config.BedrockModel)
	
	return completePrompt
}
