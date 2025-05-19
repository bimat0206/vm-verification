package validation

import (
	"fmt"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
	
	"workflow-function/ExecuteTurn1/internal"
)

// Validator provides validation logic for ExecuteTurn1
type Validator struct {
	logger logger.Logger
}

// NewValidator creates a new validator
func NewValidator(logger logger.Logger) *Validator {
	return &Validator{
		logger: logger.WithFields(map[string]interface{}{"component": "Validator"}),
	}
}

// ValidateStateReferences validates incoming state references
func (v *Validator) ValidateStateReferences(refs *internal.StateReferences) error {
	if refs == nil {
		return wferrors.NewValidationError("StateReferences is nil", nil)
	}

	if refs.VerificationId == "" {
		return wferrors.NewValidationError("VerificationId is required", nil)
	}

	var missingRefs []string

	// Required references
	if refs.VerificationContext == nil {
		missingRefs = append(missingRefs, "VerificationContext")
	}

	if refs.SystemPrompt == nil {
		missingRefs = append(missingRefs, "SystemPrompt")
	}

	if refs.BedrockConfig == nil {
		missingRefs = append(missingRefs, "BedrockConfig")
	}

	// Check for images
	if refs.Images == nil {
		v.logger.Warn("No Images reference provided", nil)
	}

	if len(missingRefs) > 0 {
		v.logger.Error("Missing required state references", map[string]interface{}{
			"missingRefs": missingRefs,
		})
		return wferrors.NewValidationError("Missing required state references", map[string]interface{}{
			"missingRefs": missingRefs,
		})
	}

	v.logger.Info("State references validation passed", nil)
	return nil
}

// ValidateWorkflowState validates the complete workflow state
func (v *Validator) ValidateWorkflowState(state *schema.WorkflowState) error {
	if state == nil {
		return wferrors.NewValidationError("WorkflowState is nil", nil)
	}

	// Validate core workflow state using schema validator
	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		v.logger.Error("WorkflowState validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"verificationId":   state.VerificationContext.VerificationId,
		})
		return wferrors.NewValidationError("Invalid WorkflowState", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Additional validation beyond the schema validator
	if err := v.validateRequiredFields(state); err != nil {
		return err
	}

	// Validate current prompt
	if err := v.validateCurrentPrompt(state); err != nil {
		return err
	}

	// Validate Bedrock config
	if err := v.validateBedrockConfig(state); err != nil {
		return err
	}

	v.logger.Info("WorkflowState validation passed", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"status":         state.VerificationContext.Status,
	})

	return nil
}

// validateRequiredFields checks for nil pointers in critical fields
func (v *Validator) validateRequiredFields(state *schema.WorkflowState) error {
	if state.CurrentPrompt == nil {
		return wferrors.NewValidationError("CurrentPrompt is nil", nil)
	}

	if state.BedrockConfig == nil {
		return wferrors.NewValidationError("BedrockConfig is nil", nil)
	}

	if state.VerificationContext == nil {
		return wferrors.NewValidationError("VerificationContext is nil", nil)
	}

	return nil
}

// validateCurrentPrompt validates the current prompt
func (v *Validator) validateCurrentPrompt(state *schema.WorkflowState) error {
	// Standard schema validation
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, false); len(errs) > 0 {
		v.logger.Error("CurrentPrompt validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
		})
		return wferrors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Check that we have at least one message
	if len(state.CurrentPrompt.Messages) == 0 {
		return wferrors.NewValidationError("CurrentPrompt has no messages", map[string]interface{}{
			"promptId": state.CurrentPrompt.PromptId,
		})
	}

	// Check message content
	msg := state.CurrentPrompt.Messages[0]
	if len(msg.Content) == 0 {
		return wferrors.NewValidationError("Message has no content", map[string]interface{}{
			"promptId": state.CurrentPrompt.PromptId,
			"role":     msg.Role,
		})
	}

	// If there's an image in the content, verify we have image data
	if len(msg.Content) > 1 && msg.Content[1].Image != nil {
		if state.Images == nil {
			return wferrors.NewValidationError("Message contains image content but no Images data provided", map[string]interface{}{
				"promptId": state.CurrentPrompt.PromptId,
			})
		}

		// Verify reference image exists
		if state.Images.Reference == nil && state.Images.ReferenceImage == nil {
			return wferrors.NewValidationError("No reference image found in Images data", nil)
		}
	}

	return nil
}

// validateBedrockConfig validates the Bedrock configuration
func (v *Validator) validateBedrockConfig(state *schema.WorkflowState) error {
	// Standard schema validation
	if errs := schema.ValidateBedrockConfig(state.BedrockConfig); len(errs) > 0 {
		v.logger.Error("BedrockConfig validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Ensure reasonable values for MaxTokens
	if state.BedrockConfig.MaxTokens <= 0 {
		v.logger.Warn("BedrockConfig.MaxTokens is invalid, setting to 4096", map[string]interface{}{
			"currentValue": state.BedrockConfig.MaxTokens,
		})
		state.BedrockConfig.MaxTokens = 4096
	}

	// Ensure reasonable values for temperature
	if state.BedrockConfig.Temperature < 0 || state.BedrockConfig.Temperature > 1 {
		v.logger.Warn("BedrockConfig.Temperature is outside valid range, setting to 0.7", map[string]interface{}{
			"currentValue": state.BedrockConfig.Temperature,
		})
		state.BedrockConfig.Temperature = 0.7
	}

	// Ensure AnthropicVersion is set
	if state.BedrockConfig.AnthropicVersion == "" {
		v.logger.Warn("BedrockConfig.AnthropicVersion is empty, setting to bedrock-2023-05-31", nil)
		state.BedrockConfig.AnthropicVersion = "bedrock-2023-05-31"
	}

	return nil
}

// ValidateImageData validates image data after Base64 generation
func (v *Validator) ValidateImageData(images *schema.ImageData) error {
	if images == nil {
		return fmt.Errorf("images is nil")
	}

	// Schema validation
	if errs := schema.ValidateImageData(images, true); len(errs) > 0 {
		v.logger.Error("ImageData validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return wferrors.NewValidationError("Invalid ImageData", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Get the reference image (primary or fallback)
	var imageInfo *schema.ImageInfo
	if images.Reference != nil {
		imageInfo = images.Reference
	} else if images.ReferenceImage != nil {
		imageInfo = images.ReferenceImage
	} else {
		return wferrors.NewValidationError("No reference image found in ImageData", nil)
	}

	// Verify Base64 data
	if !imageInfo.HasBase64Data() {
		return wferrors.NewValidationError("Image has no Base64 data", map[string]interface{}{
			"imageUrl": imageInfo.URL,
		})
	}

	// Verify image format is supported
	if imageInfo.Format != "jpeg" && imageInfo.Format != "png" {
		return wferrors.NewValidationError("Unsupported image format", map[string]interface{}{
			"format": imageInfo.Format,
			"url":    imageInfo.URL,
		})
	}

	v.logger.Info("ImageData validation passed", map[string]interface{}{
		"format": imageInfo.Format,
		"url":    imageInfo.URL,
	})

	return nil
}

// ValidateTurnResponse validates the Turn1 response
func (v *Validator) ValidateTurnResponse(turnResponse *schema.TurnResponse) error {
	if turnResponse == nil {
		return wferrors.NewValidationError("TurnResponse is nil", nil)
	}

	// Validate required fields
	if turnResponse.TurnId != 1 {
		return wferrors.NewValidationError("Invalid TurnId for Turn1", map[string]interface{}{
			"turnId": turnResponse.TurnId,
		})
	}

	if turnResponse.Timestamp == "" {
		return wferrors.NewValidationError("Missing timestamp", nil)
	}

	if turnResponse.Response.Content == "" {
		return wferrors.NewValidationError("Empty response content", nil)
	}

	// Validate token usage
	if turnResponse.TokenUsage != nil {
		if turnResponse.TokenUsage.TotalTokens <= 0 {
			v.logger.Warn("Token usage reporting issue: TotalTokens <= 0", map[string]interface{}{
				"totalTokens": turnResponse.TokenUsage.TotalTokens,
			})
		}

		calculatedTotal := turnResponse.TokenUsage.InputTokens + turnResponse.TokenUsage.OutputTokens + turnResponse.TokenUsage.ThinkingTokens
		if turnResponse.TokenUsage.TotalTokens != calculatedTotal {
			v.logger.Warn("Token usage mismatch", map[string]interface{}{
				"reported":   turnResponse.TokenUsage.TotalTokens,
				"calculated": calculatedTotal,
			})
		}
	}

	v.logger.Info("TurnResponse validation passed", map[string]interface{}{
		"turnId": turnResponse.TurnId,
	})

	return nil
}