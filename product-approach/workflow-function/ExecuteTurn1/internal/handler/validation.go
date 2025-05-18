package handler

import (
	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// validateCoreWorkflowState performs initial validation before image processing
func (h *Handler) validateCoreWorkflowState(state *schema.WorkflowState, log logger.Logger) error {
	log.Debug("Starting core workflow state validation", nil)

	// Check for nil pointers early
	if err := h.validateRequiredFields(state); err != nil {
		return h.createAndLogError(state, err, log, schema.StatusBedrockProcessingFailed)
	}

	// Validate workflow state structure
	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		log.Error("WorkflowState validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"verificationId":   state.VerificationContext.VerificationId,
		})
		wfErr := wferrors.NewValidationError("Invalid WorkflowState", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	// Validate current prompt WITHOUT requiring images (images will be validated later)
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, false); len(errs) > 0 {
		log.Error("CurrentPrompt initial validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
			"messagesCount":    len(state.CurrentPrompt.Messages),
		})
		wfErr := wferrors.NewValidationError("Invalid CurrentPrompt structure", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
		})
		return h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	// Validate Bedrock configuration
	if errs := schema.ValidateBedrockConfig(state.BedrockConfig); len(errs) > 0 {
		log.Error("BedrockConfig validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			// ModelId removed as it's not in schema.BedrockConfig
				"anthropicVersion": state.BedrockConfig.AnthropicVersion,
		})
		wfErr := wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	log.Debug("Core workflow state validation completed successfully", nil)
	return nil
}

// validateRequiredFields checks for nil pointers in critical fields
func (h *Handler) validateRequiredFields(state *schema.WorkflowState) error {
	if state.CurrentPrompt == nil {
		return wferrors.NewValidationError("CurrentPrompt is nil", nil)
	}

	if state.BedrockConfig == nil {
		return wferrors.NewValidationError("BedrockConfig is nil", nil)
	}

	if state.VerificationContext == nil {
		return wferrors.NewValidationError("VerificationContext is nil", nil)
	}

	if len(state.CurrentPrompt.Messages) == 0 {
		return wferrors.NewValidationError("CurrentPrompt has no messages", map[string]interface{}{
			"promptId": state.CurrentPrompt.PromptId,
		})
	}

	return nil
}

// validateCompleteWorkflowState validates the workflow state after images are processed
func (h *Handler) validateCompleteWorkflowState(state *schema.WorkflowState, log logger.Logger) error {
	log.Debug("Starting complete workflow state validation", nil)

	// Validate image data now that Base64 is available
	if errs := schema.ValidateImageData(state.Images, true); len(errs) > 0 {
		log.Error("ImageData validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		wfErr := wferrors.NewValidationError("Invalid ImageData", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	// Now validate current prompt WITH images
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, true); len(errs) > 0 {
		log.Error("CurrentPrompt full validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
		})
		wfErr := wferrors.NewValidationError("Invalid CurrentPrompt with images", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
		})
		return h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	log.Debug("Complete workflow state validation completed successfully", nil)
	return nil
}