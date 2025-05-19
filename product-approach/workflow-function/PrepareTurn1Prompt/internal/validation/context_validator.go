package validation

import (
	//"fmt"
	"strings"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// ValidateVerificationContext validates the verification context
func (v *Validator) ValidateVerificationContext(ctx *schema.VerificationContext) error {
	if ctx == nil {
		return errors.NewValidationError("Verification context is nil", nil)
	}

	// Required fields
	if ctx.VerificationId == "" {
		return errors.NewMissingFieldError("verificationContext.verificationId")
	}

	if ctx.VerificationType == "" {
		return errors.NewMissingFieldError("verificationContext.verificationType")
	}

	if ctx.VerificationAt == "" {
		return errors.NewMissingFieldError("verificationContext.verificationAt")
	}

	if ctx.Status == "" {
		return errors.NewMissingFieldError("verificationContext.status")
	}

	// Validate specific fields based on verification type
	if ctx.VerificationType == schema.VerificationTypeLayoutVsChecking {
		return v.validateLayoutVsCheckingContext(ctx)
	} else if ctx.VerificationType == schema.VerificationTypePreviousVsCurrent {
		return v.validatePreviousVsCurrentContext(ctx)
	}

	return nil
}

// validateLayoutVsCheckingContext validates layout vs checking specific fields
func (v *Validator) validateLayoutVsCheckingContext(ctx *schema.VerificationContext) error {
	// Layout ID is required
	if ctx.LayoutId <= 0 {
		return errors.NewValidationError("Layout ID is required for LAYOUT_VS_CHECKING", 
			map[string]interface{}{"layoutId": ctx.LayoutId})
	}

	// Layout prefix is required
	if ctx.LayoutPrefix == "" {
		return errors.NewMissingFieldError("verificationContext.layoutPrefix")
	}

	// Vending machine ID is required
	if ctx.VendingMachineId == "" {
		return errors.NewMissingFieldError("verificationContext.vendingMachineId")
	}

	// Reference image URL is required
	if ctx.ReferenceImageUrl == "" {
		return errors.NewMissingFieldError("verificationContext.referenceImageUrl")
	}

	// Validate reference image URL is from S3
	if !strings.HasPrefix(ctx.ReferenceImageUrl, "s3://") {
		return errors.NewValidationError("Reference image URL must be an S3 URL", 
			map[string]interface{}{"url": ctx.ReferenceImageUrl})
	}

	return nil
}

// validatePreviousVsCurrentContext validates previous vs current specific fields
func (v *Validator) validatePreviousVsCurrentContext(ctx *schema.VerificationContext) error {
	// Vending machine ID is required
	if ctx.VendingMachineId == "" {
		return errors.NewMissingFieldError("verificationContext.vendingMachineId")
	}

	// Reference image URL is required (in this case, it's the previous state image)
	if ctx.ReferenceImageUrl == "" {
		return errors.NewMissingFieldError("verificationContext.referenceImageUrl")
	}

	// Validate reference image URL is from S3
	if !strings.HasPrefix(ctx.ReferenceImageUrl, "s3://") {
		return errors.NewValidationError("Reference image URL must be an S3 URL", 
			map[string]interface{}{"url": ctx.ReferenceImageUrl})
	}

	return nil
}

// ValidateWorkflowState validates the complete workflow state
func (v *Validator) ValidateWorkflowState(state *schema.WorkflowState) error {
	if state == nil {
		return errors.NewValidationError("Workflow state is nil", nil)
	}

	// Validate verification context
	if err := v.ValidateVerificationContext(state.VerificationContext); err != nil {
		return err
	}

	// Validate current prompt
	if state.CurrentPrompt == nil {
		return errors.NewValidationError("Current prompt is nil", nil)
	}

	if state.CurrentPrompt.TurnNumber != 1 {
		return errors.NewInvalidFieldError("currentPrompt.turnNumber", state.CurrentPrompt.TurnNumber, "1")
	}

	// Validate images
	if state.Images == nil {
		return errors.NewValidationError("Images data is nil", nil)
	}

	// Reference image is required for Turn 1
	if refImage := state.Images.GetReference(); refImage == nil {
		return errors.NewValidationError("Reference image is required for Turn 1", nil)
	}

	// Verify that at least some data is present
	switch state.VerificationContext.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		if state.LayoutMetadata == nil {
			v.log.Warn("Layout metadata is missing", map[string]interface{}{
				"verificationId": state.VerificationContext.VerificationId,
			})
		}
	case schema.VerificationTypePreviousVsCurrent:
		if state.HistoricalContext == nil {
			v.log.Warn("Historical context is missing", map[string]interface{}{
				"verificationId": state.VerificationContext.VerificationId,
			})
		}
	}

	return nil
}