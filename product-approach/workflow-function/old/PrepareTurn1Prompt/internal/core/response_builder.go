package core

import (
	//"time"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"prepare-turn1/internal/integration"
	"prepare-turn1/internal/state"
)

// ResponseBuilder handles building the response from the workflow state
type ResponseBuilder struct {
	log             logger.Logger
	bedrockIntegrator *integration.BedrockIntegrator
}

// NewResponseBuilder creates a new response builder with the given logger and Bedrock integrator
func NewResponseBuilder(log logger.Logger, bedrockIntegrator *integration.BedrockIntegrator) *ResponseBuilder {
	return &ResponseBuilder{
		log:              log,
		bedrockIntegrator: bedrockIntegrator,
	}
}

// BuildResponse creates a response from the workflow state
func (r *ResponseBuilder) BuildResponse(workflowState *schema.WorkflowState, promptText string, templateName string) error {
	if workflowState == nil {
		return errors.NewValidationError("Workflow state is nil", nil)
	}

	// Create Bedrock message with prompt text and reference image
	userMessage, err := r.bedrockIntegrator.CreateBedrockMessage(promptText, workflowState)
	if err != nil {
		return errors.NewInternalError("bedrock-message-creation", err)
	}

	// Update verification context status
	integration.UpdateVerificationStatus(workflowState, schema.StatusTurn1PromptReady)

	// Update the current prompt in the workflow state
	templateVersion := templateName // This should be replaced with the actual version from the template loader
	r.bedrockIntegrator.UpdateCurrentPrompt(workflowState, userMessage, templateVersion)

	// Log response building
	r.log.Info("Response built successfully", map[string]interface{}{
		"verificationId": workflowState.VerificationContext.VerificationId,
		"status":         workflowState.VerificationContext.Status,
		"promptId":       workflowState.CurrentPrompt.PromptId,
	})

	return nil
}

// ConfigureBedrock sets the Bedrock configuration in the workflow state
func (r *ResponseBuilder) ConfigureBedrock(workflowState *schema.WorkflowState, input *state.Input) {
	// Get Bedrock configuration from environment variables
	anthropicVersion := integration.GetEnvWithDefault("ANTHROPIC_VERSION", "bedrock-2023-05-31")
	maxTokens := integration.GetIntEnvWithDefault("MAX_TOKENS", 24000)
	thinkingType := integration.GetEnvWithDefault("THINKING_TYPE", "enabled")
	budgetTokens := integration.GetIntEnvWithDefault("BUDGET_TOKENS", 16000)

	// Set Bedrock configuration
	r.bedrockIntegrator.SetBedrockConfiguration(
		workflowState,
		anthropicVersion,
		maxTokens,
		thinkingType,
		budgetTokens,
	)
}