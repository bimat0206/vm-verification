package models

import (
	//"fmt"
	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// Validator handles request validation with detailed error reporting
type Validator struct {
	logger logger.Logger
}

// NewValidator creates a new validator instance
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		logger: log.WithFields(map[string]interface{}{
			"component": "RequestValidator",
		}),
	}
}

// ValidateRequest performs comprehensive validation of the ExecuteTurn1Request
func (v *Validator) ValidateRequest(req *ExecuteTurn1Request) error {
	v.logger.Debug("Starting request validation", map[string]interface{}{
		"verificationId":    req.GetVerificationID(),
		"hasValidStructure": req.HasValidStructure(),
	})

	// Step 1: Validate basic structure and required fields
	if err := v.validateBasicStructure(req); err != nil {
		return err
	}

	// Step 2: Validate workflow state structure using schema validation
	if err := v.validateWorkflowState(&req.WorkflowState); err != nil {
		return err
	}

	// Step 3: Validate prompt structure (without requiring images yet)
	if err := v.validatePromptStructure(req.WorkflowState.CurrentPrompt); err != nil {
		return err
	}

	// Step 4: Validate Bedrock configuration
	if err := v.validateBedrockConfig(req.WorkflowState.BedrockConfig); err != nil {
		return err
	}

	// Step 5: Validate verification context
	if err := v.validateVerificationContext(req.WorkflowState.VerificationContext); err != nil {
		return err
	}

	v.logger.Debug("Request validation completed successfully", map[string]interface{}{
		"verificationId": req.GetVerificationID(),
	})

	return nil
}

// validateBasicStructure performs basic nil checks and structure validation
func (v *Validator) validateBasicStructure(req *ExecuteTurn1Request) error {
	if req == nil {
		v.logger.Error("Request is nil", nil)
		return wferrors.NewValidationError("Request cannot be nil", nil)
	}

	// Check for required nested structures
	if req.WorkflowState.VerificationContext == nil {
		v.logger.Error("VerificationContext is required but is nil", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		return wferrors.NewValidationError("VerificationContext is required", map[string]interface{}{
			"field": "VerificationContext",
		})
	}

	if req.WorkflowState.CurrentPrompt == nil {
		v.logger.Error("CurrentPrompt is required but is nil", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		return wferrors.NewValidationError("CurrentPrompt is required", map[string]interface{}{
			"field": "CurrentPrompt",
		})
	}

	if req.WorkflowState.BedrockConfig == nil {
		v.logger.Error("BedrockConfig is required but is nil", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		return wferrors.NewValidationError("BedrockConfig is required", map[string]interface{}{
			"field": "BedrockConfig",
		})
	}

	return nil
}

// validateWorkflowState validates the overall workflow state using schema validation
func (v *Validator) validateWorkflowState(state *schema.WorkflowState) error {
	v.logger.Debug("Validating workflow state structure", map[string]interface{}{
		"schemaVersion": state.SchemaVersion,
	})

	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		v.logger.Error("WorkflowState schema validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"schemaVersion":    state.SchemaVersion,
		})
		return wferrors.NewValidationError("Invalid WorkflowState structure", map[string]interface{}{
			"validationErrors": errs.Error(),
			"schemaVersion":    state.SchemaVersion,
		})
	}

	return nil
}

// validatePromptStructure validates the current prompt structure
// Note: This validation does NOT require images - they will be validated later in the workflow
func (v *Validator) validatePromptStructure(prompt *schema.CurrentPrompt) error {
	v.logger.Debug("Validating prompt structure", map[string]interface{}{
		"promptId":     prompt.PromptId,
		"messageCount": len(prompt.Messages),
	})

	// Use schema validation without requiring images
	if errs := schema.ValidateCurrentPrompt(prompt, false); len(errs) > 0 {
		v.logger.Error("CurrentPrompt schema validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         prompt.PromptId,
			"messageCount":     len(prompt.Messages),
		})
		return wferrors.NewValidationError("Invalid CurrentPrompt structure", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         prompt.PromptId,
		})
	}

	// Additional business logic validation
	if err := v.validatePromptMessages(prompt); err != nil {
		return err
	}

	return nil
}

// validatePromptMessages performs additional validation of prompt messages
func (v *Validator) validatePromptMessages(prompt *schema.CurrentPrompt) error {
	if len(prompt.Messages) == 0 {
		v.logger.Error("Prompt has no messages", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
		return wferrors.NewValidationError("CurrentPrompt must have at least one message", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
	}

	// Validate first message (required for ExecuteTurn1)
	firstMsg := prompt.Messages[0]
	if firstMsg.Role == "" {
		v.logger.Error("First message missing role", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
		return wferrors.NewValidationError("First message must have a role", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
	}

	// Check if message has content
	if len(firstMsg.Content) == 0 {
		v.logger.Error("First message has no content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
		return wferrors.NewValidationError("First message must have content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
	}

	// Validate text content exists
	if firstMsg.Content[0].Text == "" {
		v.logger.Error("First message has empty text content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
		return wferrors.NewValidationError("First message must have text content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
	}

	return nil
}

// validateBedrockConfig validates the Bedrock configuration
func (v *Validator) validateBedrockConfig(config *schema.BedrockConfig) error {
	v.logger.Debug("Validating Bedrock configuration", map[string]interface{}{
		"anthropicVersion": config.AnthropicVersion,
		"maxTokens":        config.MaxTokens,
	})

	// Use schema validation
	if errs := schema.ValidateBedrockConfig(config); len(errs) > 0 {
		v.logger.Error("BedrockConfig schema validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"anthropicVersion": config.AnthropicVersion,
			"maxTokens":        config.MaxTokens,
		})
		return wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Additional business logic validation
	if config.AnthropicVersion == "" {
		v.logger.Error("BedrockConfig missing AnthropicVersion", nil)
		return wferrors.NewValidationError("BedrockConfig.AnthropicVersion is required", nil)
	}

	if config.MaxTokens <= 0 {
		v.logger.Error("BedrockConfig has invalid MaxTokens", map[string]interface{}{
			"maxTokens": config.MaxTokens,
		})
		return wferrors.NewValidationError("BedrockConfig.MaxTokens must be greater than 0", map[string]interface{}{
			"maxTokens": config.MaxTokens,
		})
	}

	return nil
}

// validateVerificationContext validates the verification context
func (v *Validator) validateVerificationContext(context *schema.VerificationContext) error {
	v.logger.Debug("Validating verification context", map[string]interface{}{
		"verificationId": context.VerificationId,
		"status":         context.Status,
	})

	// Check required fields
	if context.VerificationId == "" {
		v.logger.Error("VerificationContext missing VerificationId", nil)
		return wferrors.NewValidationError("VerificationContext.VerificationId is required", nil)
	}

	// Validate status is not in a terminal error state
	errorStates := []string{
		schema.StatusBedrockProcessingFailed,
		schema.StatusVerificationFailed,
		"TURN1_FAILED", // TODO: Use schema constants when available
		"TURN2_FAILED", // TODO: Use schema constants when available
	}

	for _, errorState := range errorStates {
		if context.Status == errorState {
			v.logger.Error("VerificationContext is in error state", map[string]interface{}{
				"status":         context.Status,
				"verificationId": context.VerificationId,
			})
			return wferrors.NewValidationError("VerificationContext is in error state", map[string]interface{}{
				"status":         context.Status,
				"verificationId": context.VerificationId,
			})
		}
	}

	return nil
}

// Backward compatibility method - maintains existing API
func (req *ExecuteTurn1Request) Validate(log logger.Logger) error {
	validator := NewValidator(log)
	return validator.ValidateRequest(req)
}