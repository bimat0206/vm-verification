package handler

import (
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/validation"
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

// ValidateRequest validates the incoming request
func (v *Validator) ValidateRequest(req *models.Turn1Request) error {
	if err := v.validator.ValidateRequest(req); err != nil {
		return errors.NewValidationError("request validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	return nil
}

// ValidateResponse validates the outgoing response
func (v *Validator) ValidateResponse(resp *models.Turn1Response) error {
	if err := v.validator.ValidateResponse(resp); err != nil {
		return errors.NewValidationError("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}
	return nil
}

// GetSchemaVersion returns the current schema version
func (v *Validator) GetSchemaVersion() string {
	return v.validator.GetSchemaVersion()
}