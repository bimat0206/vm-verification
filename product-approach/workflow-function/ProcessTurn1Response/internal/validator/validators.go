package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"workflow-function/shared/logger"
)

// Validator provides validation logic for processed responses
type Validator struct {
	log logger.Logger
}

// ValidationResult contains the result of validation
type ValidationResult struct {
	IsValid    bool
	Errors     []ValidationError
	Warnings   []ValidationWarning
	Completeness float64
}

// ValidationError represents a validation error
type ValidationError struct {
	Field       string
	Message     string
	Severity    string
	Suggestions []string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string
	Message string
	Impact  string
}

// NewValidator creates a new validator instance
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		log: log,
	}
}

// ValidateReferenceAnalysis validates the extracted reference analysis
func (v *Validator) ValidateReferenceAnalysis(analysis map[string]interface{}) error {
	if analysis == nil || len(analysis) == 0 {
		return fmt.Errorf("reference analysis is empty or nil")
	}

	result := v.performComprehensiveValidation(analysis)
	
	v.log.Info("Reference analysis validation completed", map[string]interface{}{
		"isValid":      result.IsValid,
		"errorCount":   len(result.Errors),
		"warningCount": len(result.Warnings),
		"completeness": result.Completeness,
	})

	if !result.IsValid {
		return fmt.Errorf("validation failed: %s", v.formatValidationErrors(result.Errors))
	}

	// Log warnings even if validation passes
	if len(result.Warnings) > 0 {
		v.log.Warn("Validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(result.Warnings),
		})
	}

	return nil
}

// ValidateMachineStructure validates machine structure data
func (v *Validator) ValidateMachineStructure(structure map[string]interface{}) error {
	if structure == nil {
		return fmt.Errorf("machine structure is nil")
	}

	errors := []ValidationError{}

	// Validate required fields
	requiredFields := []string{"rowCount", "columnsPerRow"}
	for _, field := range requiredFields {
		if _, exists := structure[field]; !exists {
			errors = append(errors, ValidationError{
				Field:    field,
				Message:  "required field missing",
				Severity: "high",
			})
		}
	}

	// Validate field types and values
	if rowCount, ok := structure["rowCount"]; ok {
		if !v.isPositiveNumber(rowCount) {
			errors = append(errors, ValidationError{
				Field:    "rowCount",
				Message:  "must be a positive number",
				Severity: "high",
			})
		}
	}

	if columnsPerRow, ok := structure["columnsPerRow"]; ok {
		if !v.isPositiveNumber(columnsPerRow) {
			errors = append(errors, ValidationError{
				Field:    "columnsPerRow",
				Message:  "must be a positive number",
				Severity: "high",
			})
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("machine structure validation failed: %s", v.formatValidationErrors(errors))
	}

	return nil
}

// ValidateCompleteness checks data completeness
func (v *Validator) ValidateCompleteness(data map[string]interface{}) ValidationResult {
	if data == nil {
		return ValidationResult{
			IsValid:      false,
			Errors:       []ValidationError{{Field: "data", Message: "data is nil", Severity: "high"}},
			Completeness: 0.0,
		}
	}

	var errors []ValidationError
	var warnings []ValidationWarning
	
	// Check for required fields based on analysis type
	sourceType := v.getSourceType(data)
	requiredFields := v.getRequiredFieldsForType(sourceType)
	
	presentFields := 0
	for _, field := range requiredFields {
		if v.fieldExists(data, field) {
			presentFields++
		} else {
			errors = append(errors, ValidationError{
				Field:    field,
				Message:  fmt.Sprintf("required field missing for %s", sourceType),
				Severity: "medium",
			})
		}
	}

	// Calculate completeness
	completeness := float64(presentFields) / float64(len(requiredFields))

	// Check for unexpected empty fields
	v.checkForEmptyFields(data, &warnings)

	// Validate nested structures
	v.validateNestedStructures(data, &errors, &warnings)

	return ValidationResult{
		IsValid:      len(errors) == 0,
		Errors:       errors,
		Warnings:     warnings,
		Completeness: completeness,
	}
}

// performComprehensiveValidation performs a complete validation of the analysis
func (v *Validator) performComprehensiveValidation(analysis map[string]interface{}) ValidationResult {
	var errors []ValidationError
	var warnings []ValidationWarning

	// Validate basic structure
	if v.isEmptyAnalysis(analysis) {
		errors = append(errors, ValidationError{
			Field:    "analysis",
			Message:  "analysis contains no meaningful data",
			Severity: "high",
		})
	}

	// Get source type and validate accordingly
	sourceType := v.getSourceType(analysis)
	
	// Validate based on source type
	switch sourceType {
	case "REFERENCE_VALIDATION":
		v.validateReferenceValidation(analysis, &errors, &warnings)
	case "HISTORICAL_WITH_VISUAL_CONFIRMATION":
		v.validateHistoricalEnhancement(analysis, &errors, &warnings)
	case "FRESH_VISUAL_ANALYSIS":
		v.validateFreshExtraction(analysis, &errors, &warnings)
	default:
		warnings = append(warnings, ValidationWarning{
			Field:   "sourceType",
			Message: fmt.Sprintf("unknown source type: %s", sourceType),
			Impact:  "medium",
		})
	}

	// Validate context for Turn 2
	v.validateContextForTurn2(analysis, &errors, &warnings)

	// Calculate completeness
	completeness := v.calculateCompleteness(analysis, sourceType)

	return ValidationResult{
		IsValid:      len(errors) == 0,
		Errors:       errors,
		Warnings:     warnings,
		Completeness: completeness,
	}
}

// validateReferenceValidation validates UC1 validation flow results
func (v *Validator) validateReferenceValidation(analysis map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
	// Check for validation results
	if _, exists := analysis["validationResults"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "validationResults",
			Message:  "validation results missing for reference validation",
			Severity: "high",
		})
	}

	// Check for basic observations
	if _, exists := analysis["basicObservations"]; !exists {
		*warnings = append(*warnings, ValidationWarning{
			Field:   "basicObservations",
			Message: "basic observations missing",
			Impact:  "low",
		})
	}
}

// validateHistoricalEnhancement validates UC2 with historical context
func (v *Validator) validateHistoricalEnhancement(analysis map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
	// Check for historical baseline
	if _, exists := analysis["historicalBaseline"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "historicalBaseline",
			Message:  "historical baseline missing for enhanced analysis",
			Severity: "high",
		})
	}

	// Validate machine structure
	if structure, exists := analysis["machineStructure"].(map[string]interface{}); exists {
		if err := v.ValidateMachineStructure(structure); err != nil {
			*warnings = append(*warnings, ValidationWarning{
				Field:   "machineStructure",
				Message: err.Error(),
				Impact:  "medium",
			})
		}
	}
}

// validateFreshExtraction validates UC2 without historical context
func (v *Validator) validateFreshExtraction(analysis map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
	// Check for extracted structure
	if _, exists := analysis["extractedStructure"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "extractedStructure",
			Message:  "extracted structure missing for fresh analysis",
			Severity: "high",
		})
	}

	// Check for extracted state
	if _, exists := analysis["extractedState"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "extractedState",
			Message:  "extracted state missing for fresh analysis",
			Severity: "high",
		})
	}
}

// validateContextForTurn2 validates the context prepared for Turn 2
func (v *Validator) validateContextForTurn2(analysis map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
	context, exists := analysis["contextForTurn2"].(map[string]interface{})
	if !exists {
		*errors = append(*errors, ValidationError{
			Field:    "contextForTurn2",
			Message:  "context for Turn 2 is missing",
			Severity: "high",
		})
		return
	}

	// Check for required context fields
	requiredContextFields := []string{"readyForTurn2"}
	for _, field := range requiredContextFields {
		if _, exists := context[field]; !exists {
			*warnings = append(*warnings, ValidationWarning{
				Field:   fmt.Sprintf("contextForTurn2.%s", field),
				Message: "recommended context field missing",
				Impact:  "low",
			})
		}
	}
}

// Helper methods

// isEmptyAnalysis checks if analysis contains meaningful data
func (v *Validator) isEmptyAnalysis(analysis map[string]interface{}) bool {
	meaningfulFields := []string{"status", "sourceType", "validationResults", "extractedStructure", "extractedState"}
	
	for _, field := range meaningfulFields {
		if v.fieldExists(analysis, field) {
			return false
		}
	}
	
	return true
}

// getSourceType extracts the source type from analysis
func (v *Validator) getSourceType(analysis map[string]interface{}) string {
	if sourceType, ok := analysis["sourceType"].(string); ok {
		return sourceType
	}
	return "UNKNOWN"
}

// getRequiredFieldsForType returns required fields based on source type
func (v *Validator) getRequiredFieldsForType(sourceType string) []string {
	switch sourceType {
	case "REFERENCE_VALIDATION":
		return []string{"status", "sourceType", "validationResults", "contextForTurn2"}
	case "HISTORICAL_WITH_VISUAL_CONFIRMATION":
		return []string{"status", "sourceType", "historicalBaseline", "machineStructure", "contextForTurn2"}
	case "FRESH_VISUAL_ANALYSIS":
		return []string{"status", "sourceType", "extractedStructure", "extractedState", "contextForTurn2"}
	default:
		return []string{"status", "sourceType"}
	}
}

// fieldExists checks if a field exists and is not empty
func (v *Validator) fieldExists(data map[string]interface{}, field string) bool {
	value, exists := data[field]
	if !exists {
		return false
	}

	// Check for various empty types
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case map[string]interface{}:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	case nil:
		return false
	default:
		return true
	}
}

// isPositiveNumber checks if a value is a positive number
func (v *Validator) isPositiveNumber(value interface{}) bool {
	switch v := value.(type) {
	case int:
		return v > 0
	case int64:
		return v > 0
	case float64:
		return v > 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i > 0
		}
		if f, err := v.Float64(); err == nil {
			return f > 0
		}
	}
	return false
}

// checkForEmptyFields identifies unexpectedly empty fields
func (v *Validator) checkForEmptyFields(data map[string]interface{}, warnings *[]ValidationWarning) {
	for key, value := range data {
		if value == nil || value == "" {
			*warnings = append(*warnings, ValidationWarning{
				Field:   key,
				Message: "field is empty or nil",
				Impact:  "low",
			})
		}
	}
}

// validateNestedStructures validates nested map structures
func (v *Validator) validateNestedStructures(data map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
	for key, value := range data {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			if len(nestedMap) == 0 {
				*warnings = append(*warnings, ValidationWarning{
					Field:   key,
					Message: "nested structure is empty",
					Impact:  "low",
				})
			}
		}
	}
}

// calculateCompleteness calculates data completeness percentage
func (v *Validator) calculateCompleteness(analysis map[string]interface{}, sourceType string) float64 {
	requiredFields := v.getRequiredFieldsForType(sourceType)
	presentFields := 0

	for _, field := range requiredFields {
		if v.fieldExists(analysis, field) {
			presentFields++
		}
	}

	if len(requiredFields) == 0 {
		return 1.0
	}

	return float64(presentFields) / float64(len(requiredFields))
}

// formatValidationErrors formats validation errors into a readable string
func (v *Validator) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return strings.Join(messages, "; ")
}

// formatValidationWarnings formats validation warnings into a readable string
func (v *Validator) formatValidationWarnings(warnings []ValidationWarning) string {
	if len(warnings) == 0 {
		return ""
	}

	var messages []string
	for _, warning := range warnings {
		messages = append(messages, fmt.Sprintf("%s: %s (impact: %s)", warning.Field, warning.Message, warning.Impact))
	}

	return strings.Join(messages, "; ")
}
