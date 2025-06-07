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

// ValidateRequest validates the incoming request with enhanced error handling
func (v *Validator) ValidateRequest(req *models.Turn1Request) error {
	if err := v.validator.ValidateRequest(req); err != nil {
		return errors.NewValidationError("request validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		}).
			WithComponent("RequestValidator").
			WithOperation("ValidateRequest").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.ErrorSeverityHigh).
			WithSuggestions(
				"Check request structure matches Turn1Request schema",
				"Verify all required fields are present and valid",
				"Ensure field data types are correct",
				"Validate S3 reference formats",
			).
			WithRecoveryHints(
				"Review API documentation for correct request format",
				"Check field validation requirements",
				"Verify S3 reference structure",
			)
	}
	return nil
}

// ValidateResponse validates the outgoing response with enhanced error handling
func (v *Validator) ValidateResponse(resp *models.Turn1Response) error {
	if err := v.validator.ValidateResponse(resp); err != nil {
		return errors.NewValidationError("response validation failed", map[string]interface{}{
			"validation_error": err.Error(),
		}).
			WithComponent("ResponseValidator").
			WithOperation("ValidateResponse").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.ErrorSeverityMedium).
			WithSuggestions(
				"Check response structure matches Turn1Response schema",
				"Verify all required response fields are populated",
				"Ensure response data types are correct",
				"Validate S3 reference formats in response",
			).
			WithRecoveryHints(
				"Review response building logic",
				"Check field population in response builder",
				"Verify schema compliance",
			)
	}
	return nil
}

// GetSchemaVersion returns the current schema version
func (v *Validator) GetSchemaVersion() string {
	return v.validator.GetSchemaVersion()
}
