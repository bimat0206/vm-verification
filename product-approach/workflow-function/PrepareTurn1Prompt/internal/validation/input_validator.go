package validation

import (
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"prepare-turn1/internal/state"
)

// Validator handles input validation
type Validator struct {
	log logger.Logger
}

// NewValidator creates a new validator with the given logger
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		log: log,
	}
}

// ValidateInput validates the input parameters
func (v *Validator) ValidateInput(input *state.Input) error {
	// Basic validation
	if input == nil {
		return errors.NewValidationError("Input cannot be nil", nil)
	}

	// Validate required fields
	if input.VerificationID == "" {
		return errors.NewMissingFieldError("verificationId")
	}

	// Turn number validation (must be 1 for Turn 1)
	if input.TurnNumber != 1 {
		return errors.NewInvalidFieldError("turnNumber", input.TurnNumber, "1")
	}

	// Include image validation (must be "reference" for Turn 1)
	if input.IncludeImage != "reference" {
		return errors.NewInvalidFieldError("includeImage", input.IncludeImage, "reference")
	}

	// Validate verification type if available
	if input.VerificationType != "" {
		if err := v.validateVerificationType(input.VerificationType); err != nil {
			return err
		}
	}

	// Validate S3 references
	if err := state.ValidateReferences(input); err != nil {
		return err
	}

	return nil
}

// validateVerificationType validates that the verification type is supported
func (v *Validator) validateVerificationType(verificationType string) error {
	validTypes := []string{schema.VerificationTypeLayoutVsChecking, schema.VerificationTypePreviousVsCurrent}
	
	isValidType := false
	for _, vt := range validTypes {
		if verificationType == vt {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return errors.NewInvalidFieldError("verificationType", 
			verificationType, 
			"one of LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT")
	}
	
	return nil
}