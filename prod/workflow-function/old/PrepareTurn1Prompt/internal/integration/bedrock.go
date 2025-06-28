package integration

import (
	"fmt"
	"time"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// BedrockIntegrator handles Bedrock message creation and configuration
type BedrockIntegrator struct {
	log logger.Logger
}

// NewBedrockIntegrator creates a new Bedrock integrator with the given logger
func NewBedrockIntegrator(log logger.Logger) *BedrockIntegrator {
	return &BedrockIntegrator{
		log: log,
	}
}

// CreateBedrockMessage creates a Bedrock message with prompt text and reference image
func (b *BedrockIntegrator) CreateBedrockMessage(promptText string, state *schema.WorkflowState) (schema.BedrockMessage, error) {
	// Get reference image
	refImage := state.Images.GetReference()
	if refImage == nil {
		return schema.BedrockMessage{}, fmt.Errorf("reference image is required for Turn 1")
	}

	// Ensure image has Base64 data generated
	if !refImage.Base64Generated {
		return schema.BedrockMessage{}, fmt.Errorf("reference image Base64 data not generated")
	}

	// Use shared schema function to build the message
	message := schema.BuildBedrockMessage(promptText, refImage)

	// Log message creation
	b.log.Info("Created Bedrock message", map[string]interface{}{
		"hasImage":   refImage.Base64Generated,
		"imageFormat": refImage.Format,
		"textLength": len(promptText),
	})

	return message, nil
}

// SetBedrockConfiguration configures the Bedrock settings in the workflow state
func (b *BedrockIntegrator) SetBedrockConfiguration(state *schema.WorkflowState, anthropicVersion string, maxTokens int, thinkingType string, budgetTokens int) {
	// Create or update BedrockConfig
	if state.BedrockConfig == nil {
		state.BedrockConfig = &schema.BedrockConfig{}
	}

	// Set Anthropic version
	if anthropicVersion != "" {
		state.BedrockConfig.AnthropicVersion = anthropicVersion
	}

	// Set max tokens
	if maxTokens > 0 {
		state.BedrockConfig.MaxTokens = maxTokens
	}

	// Set thinking configuration
	if thinkingType != "" {
		if state.BedrockConfig.Thinking == nil {
			state.BedrockConfig.Thinking = &schema.Thinking{}
		}
		state.BedrockConfig.Thinking.Type = thinkingType
		
		if budgetTokens > 0 {
			state.BedrockConfig.Thinking.BudgetTokens = budgetTokens
		}
	}

	// Log configuration
	b.log.Info("Bedrock configuration set", map[string]interface{}{
		"anthropicVersion": state.BedrockConfig.AnthropicVersion,
		"maxTokens":        state.BedrockConfig.MaxTokens,
		"thinkingEnabled":  state.BedrockConfig.Thinking != nil && state.BedrockConfig.Thinking.Type == "enabled",
	})
}

// UpdateCurrentPrompt updates the current prompt in the workflow state
func (b *BedrockIntegrator) UpdateCurrentPrompt(state *schema.WorkflowState, userMessage schema.BedrockMessage, templateVersion string) {
	// Set messages array
	state.CurrentPrompt.Messages = []schema.BedrockMessage{userMessage}
	
	// Set turn number
	state.CurrentPrompt.TurnNumber = 1
	
	// Generate prompt ID
	state.CurrentPrompt.PromptId = GeneratePromptID(state.VerificationContext.VerificationId, 1)
	
	// Set created timestamp
	state.CurrentPrompt.CreatedAt = FormatTimestamp(time.Now())
	
	// Set prompt version
	state.CurrentPrompt.PromptVersion = templateVersion
	
	// Set include image flag
	state.CurrentPrompt.IncludeImage = "reference"
	
	// Log current prompt update
	b.log.Info("Updated current prompt", map[string]interface{}{
		"promptId":      state.CurrentPrompt.PromptId,
		"turnNumber":    state.CurrentPrompt.TurnNumber,
		"promptVersion": state.CurrentPrompt.PromptVersion,
		"messageCount":  len(state.CurrentPrompt.Messages),
	})
}