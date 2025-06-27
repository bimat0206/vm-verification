package handler

import (
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/validation"
	"workflow-function/shared/errors"
)

// Validator wraps validation logic
type Validator struct {
	validator *validation.SchemaValidator
}

// NewValidator creates a new instance of Validator
func NewValidator() *Validator {
	return &Validator{
		validator: validation.NewSchemaValidator(),
	}
}

// ValidateTurn2Request validates the incoming Turn2 request
func (v *Validator) ValidateTurn2Request(req *models.Turn2Request) error {
	if err := v.validator.ValidateTurn2Request(req); err != nil {
		return errors.NewValidationError("turn2 request validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	return nil
}

// ValidateTurn2Response validates the outgoing Turn2 response
func (v *Validator) ValidateTurn2Response(resp *models.Turn2Response) error {
	if err := v.validator.ValidateTurn2Response(resp); err != nil {
		return errors.NewValidationError("turn2 response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	return nil
}

// Legacy methods for compatibility (if needed)
// ValidateRequest validates the incoming request (legacy Turn1 method)
func (v *Validator) ValidateRequest(req interface{}) error {
	// This method is kept for compatibility but should not be used for Turn2
	return errors.NewValidationError("legacy validation method not supported for Turn2", nil)
}

// ValidateResponse validates the outgoing response (legacy Turn1 method)
func (v *Validator) ValidateResponse(resp interface{}) error {
	// This method is kept for compatibility but should not be used for Turn2
	return errors.NewValidationError("legacy validation method not supported for Turn2", nil)
}

// GetSchemaVersion returns the current schema version
func (v *Validator) GetSchemaVersion() string {
	return v.validator.GetSchemaVersion()
}
