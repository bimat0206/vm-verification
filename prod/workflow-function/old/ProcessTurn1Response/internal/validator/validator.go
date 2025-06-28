// Package validator provides validation functionality for ProcessTurn1Response Lambda
package validator

import (
	"fmt"
	"strings"

	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// ValidatorInterface defines the validator contract
type ValidatorInterface interface {
	// ValidateReferenceAnalysis validates the extracted reference analysis
	ValidateReferenceAnalysis(analysis map[string]interface{}) error
	
	// ValidateMachineStructure validates machine structure data
	ValidateMachineStructure(structure *schema.MachineStructure) error
	
	// ValidateProcessingResult validates the complete processing result
	ValidateProcessingResult(result *types.Turn1ProcessingResult) error
	
	// ValidateHistoricalEnhancement validates historical enhancement data
	ValidateHistoricalEnhancement(enhancement map[string]interface{}) error
	
	// ValidateCompleteness checks data completeness for a given processing path
	ValidateCompleteness(data map[string]interface{}, path types.ProcessingPath) (float64, []ValidationError, []ValidationWarning)
	
	// ValidateContextForTurn2 validates the context prepared for Turn 2
	ValidateContextForTurn2(context map[string]interface{}) error
	
	// ValidateExtractedState validates the extracted state data
	ValidateExtractedState(state *types.ExtractedState) error
}

// Validator provides validation logic for processed responses
// It implements the ValidatorInterface
type Validator struct {
	log logger.Logger
}

// Ensure Validator implements ValidatorInterface
var _ ValidatorInterface = (*Validator)(nil)

// NewValidator creates a new validator instance
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		log: log,
	}
}

// ValidationResult contains the result of validation
type ValidationResult struct {
	IsValid      bool
	Errors       []ValidationError
	Warnings     []ValidationWarning
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

// ValidateReferenceAnalysis validates the extracted reference analysis
func (v *Validator) ValidateReferenceAnalysis(analysis map[string]interface{}) error {
	if analysis == nil || len(analysis) == 0 {
		return errors.NewValidationError("reference analysis is empty or nil", nil)
	}

	result := v.performComprehensiveValidation(analysis)
	
	v.log.Info("Reference analysis validation completed", map[string]interface{}{
		"isValid":      result.IsValid,
		"errorCount":   len(result.Errors),
		"warningCount": len(result.Warnings),
		"completeness": result.Completeness,
	})

	if !result.IsValid {
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"errors": v.formatValidationErrors(result.Errors),
		})
	}

	// Log warnings even if validation passes
	if len(result.Warnings) > 0 {
		v.log.Warn("Validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(result.Warnings),
		})
	}

	return nil
}

// ValidateMachineStructure validates machine structure data using the shared schema type
func (v *Validator) ValidateMachineStructure(structure *schema.MachineStructure) error {
	if structure == nil {
		return errors.NewValidationError("machine structure is nil", nil)
	}

	validationErrors := []ValidationError{}

	// Validate required fields
	if structure.RowCount <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "rowCount",
			Message:  "must be a positive number",
			Severity: "high",
		})
	}

	if structure.ColumnsPerRow <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "columnsPerRow",
			Message:  "must be a positive number",
			Severity: "high",
		})
	}

	// Validate row and column orders
	if len(structure.RowOrder) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "rowOrder",
			Message:  "cannot be empty",
			Severity: "high",
		})
	} else if len(structure.RowOrder) != structure.RowCount {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "rowOrder",
			Message:  fmt.Sprintf("length (%d) does not match rowCount (%d)", len(structure.RowOrder), structure.RowCount),
			Severity: "high",
		})
	}

	if len(structure.ColumnOrder) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "columnOrder",
			Message:  "cannot be empty",
			Severity: "high",
		})
	} else if len(structure.ColumnOrder) != structure.ColumnsPerRow {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "columnOrder",
			Message:  fmt.Sprintf("length (%d) does not match columnsPerRow (%d)", len(structure.ColumnOrder), structure.ColumnsPerRow),
			Severity: "high",
		})
	}

	if len(validationErrors) > 0 {
		errDetails := map[string]interface{}{
			"errors": v.formatValidationErrors(validationErrors),
		}
		return errors.NewValidationError("machine structure validation failed", errDetails)
	}

	return nil
}

// ValidateProcessingResult validates the complete processing result
func (v *Validator) ValidateProcessingResult(result *types.Turn1ProcessingResult) error {
	if result == nil {
		return errors.NewValidationError("processing result is nil", nil)
	}

	var validationErrors []ValidationError
	var validationWarnings []ValidationWarning

	// Validate required fields
	if result.Status == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "status",
			Message:  "required field missing",
			Severity: "high",
		})
	} else {
		// Validate status is a known value
		validStatuses := []string{"EXTRACTION_COMPLETE", "VALIDATION_COMPLETE", "EXTRACTION_FAILED", "VALIDATION_FAILED"}
		validStatus := false
		for _, s := range validStatuses {
			if result.Status == s {
				validStatus = true
				break
			}
		}
		if !validStatus {
			validationErrors = append(validationErrors, ValidationError{
				Field:       "status",
				Message:     fmt.Sprintf("invalid status: %s", result.Status),
				Severity:    "high",
				Suggestions: validStatuses,
			})
		}
	}

	// Validate processing path
	if !result.SourceType.IsValid() {
		validationErrors = append(validationErrors, ValidationError{
			Field:       "sourceType",
			Message:     fmt.Sprintf("invalid processing path: %s", result.SourceType),
			Severity:    "high",
			Suggestions: []string{string(types.PathValidationFlow), string(types.PathHistoricalEnhancement), string(types.PathFreshExtraction)},
		})
	}

	// Different validations based on processing path
	switch result.SourceType {
	case types.PathValidationFlow:
		if len(result.ReferenceAnalysis) == 0 {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "referenceAnalysis",
				Message:  "required for validation flow",
				Severity: "high",
			})
		}

		// Optional fields for validation flow

	case types.PathHistoricalEnhancement:
		if len(result.ReferenceAnalysis) == 0 {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "referenceAnalysis",
				Message:  "required for historical enhancement",
				Severity: "high",
			})
		}

		// Check for historical context
		if _, exists := result.ReferenceAnalysis["historicalBaseline"]; !exists {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   "referenceAnalysis.historicalBaseline",
				Message: "historical baseline missing for historical enhancement path",
				Impact:  "medium",
			})
		}

	case types.PathFreshExtraction:
		if result.ExtractedStructure == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "extractedStructure",
				Message:  "required for fresh extraction",
				Severity: "high",
			})
		}

		if _, exists := result.ReferenceAnalysis["extractedState"]; !exists {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "referenceAnalysis.extractedState",
				Message:  "extracted state is required for fresh extraction",
				Severity: "high",
			})
		}
	}

	// Always validate context for Turn 2
	if len(result.ContextForTurn2) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "contextForTurn2",
			Message:  "context for Turn 2 is missing",
			Severity: "high",
		})
	} else {
		if err := v.ValidateContextForTurn2(result.ContextForTurn2); err != nil {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   "contextForTurn2",
				Message: err.Error(),
				Impact:  "medium",
			})
		}
	}

	// Validate processing metadata
	if result.ProcessingMetadata == nil {
		validationWarnings = append(validationWarnings, ValidationWarning{
			Field:   "processingMetadata",
			Message: "processing metadata is missing",
			Impact:  "low",
		})
	} else {
		// Validate processing metadata fields
		if result.ProcessingMetadata.ProcessingDuration <= 0 {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   "processingMetadata.processingDuration",
				Message: "processing duration should be positive",
				Impact:  "low",
			})
		}
	}

	if len(validationErrors) > 0 {
		errDetails := map[string]interface{}{
			"errors": v.formatValidationErrors(validationErrors),
		}
		return errors.NewValidationError("processing result validation failed", errDetails)
	}

	// Log warnings
	if len(validationWarnings) > 0 {
		v.log.Warn("Processing result validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(validationWarnings),
		})
	}

	return nil
}

// ValidateHistoricalEnhancement validates historical enhancement data
func (v *Validator) ValidateHistoricalEnhancement(enhancement map[string]interface{}) error {
	if enhancement == nil || len(enhancement) == 0 {
		return errors.NewValidationError("historical enhancement is empty or nil", nil)
	}

	var validationErrors []ValidationError
	var validationWarnings []ValidationWarning

	// Check for required fields
	requiredFields := []string{"baselineData", "visualConfirmation", "enrichedBaseline"}
	for _, field := range requiredFields {
		if _, exists := enhancement[field]; !exists {
			validationErrors = append(validationErrors, ValidationError{
				Field:    field,
				Message:  "required field missing",
				Severity: "high",
			})
		}
	}

	// Validate baselineData
	if baselineData, ok := enhancement["baselineData"].(map[string]interface{}); ok {
		if len(baselineData) == 0 {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   "baselineData",
				Message: "baseline data is empty",
				Impact:  "medium",
			})
		}
	}

	// Validate visualConfirmation
	if visualConfirmation, ok := enhancement["visualConfirmation"].(map[string]interface{}); ok {
		if len(visualConfirmation) == 0 {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   "visualConfirmation",
				Message: "visual confirmation data is empty",
				Impact:  "medium",
			})
		}
	}

	// Validate enrichedBaseline
	if enrichedBaseline, ok := enhancement["enrichedBaseline"].(map[string]interface{}); ok {
		if len(enrichedBaseline) == 0 {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "enrichedBaseline",
				Message:  "enriched baseline is empty",
				Severity: "high",
			})
		}
	}

	// Validate discrepancies field, if present
	if discrepancies, ok := enhancement["discrepancies"].([]interface{}); ok {
		if len(discrepancies) > 0 {
			// Detailed validation of discrepancies could be added here
		}
	}

	if len(validationErrors) > 0 {
		errDetails := map[string]interface{}{
			"errors": v.formatValidationErrors(validationErrors),
		}
		return errors.NewValidationError("historical enhancement validation failed", errDetails)
	}

	// Log warnings
	if len(validationWarnings) > 0 {
		v.log.Warn("Historical enhancement validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(validationWarnings),
		})
	}

	return nil
}

// ValidateCompleteness checks data completeness for a given processing path
func (v *Validator) ValidateCompleteness(data map[string]interface{}, path types.ProcessingPath) (float64, []ValidationError, []ValidationWarning) {
	if data == nil {
		return 0.0, []ValidationError{{
			Field:    "data",
			Message:  "data is nil",
			Severity: "high",
		}}, nil
	}

	var errors []ValidationError
	var warnings []ValidationWarning
	
	// Get required fields based on processing path
	requiredFields := v.getRequiredFieldsForPath(path)
	
	presentFields := 0
	for _, field := range requiredFields {
		if v.fieldExists(data, field) {
			presentFields++
		} else {
			errors = append(errors, ValidationError{
				Field:    field,
				Message:  fmt.Sprintf("required field missing for %s", path),
				Severity: "medium",
			})
		}
	}

	// Calculate completeness
	completeness := 0.0
	if len(requiredFields) > 0 {
		completeness = float64(presentFields) / float64(len(requiredFields))
	}

	// Check for unexpected empty fields
	v.checkForEmptyFields(data, &warnings)

	return completeness, errors, warnings
}

// ValidateContextForTurn2 validates the context prepared for Turn 2
func (v *Validator) ValidateContextForTurn2(context map[string]interface{}) error {
	if context == nil || len(context) == 0 {
		return errors.NewValidationError("context for Turn 2 is empty or nil", nil)
	}

	var validationErrors []ValidationError

	// Check for required fields
	requiredFields := []string{"readyForTurn2"}
	for _, field := range requiredFields {
		if _, exists := context[field]; !exists {
			validationErrors = append(validationErrors, ValidationError{
				Field:    field,
				Message:  "required field missing",
				Severity: "high",
			})
		}
	}

	// Check readyForTurn2 value if it exists
	if readyForTurn2, ok := context["readyForTurn2"].(bool); ok {
		if !readyForTurn2 {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "readyForTurn2",
				Message:  "readyForTurn2 must be true",
				Severity: "high",
			})
		}
	}

	// Recommended fields (warnings only)
	recommendedFields := []string{"referenceAnalysisComplete", "analysisType"}
	var validationWarnings []ValidationWarning
	for _, field := range recommendedFields {
		if _, exists := context[field]; !exists {
			validationWarnings = append(validationWarnings, ValidationWarning{
				Field:   field,
				Message: "recommended field missing",
				Impact:  "low",
			})
		}
	}

	if len(validationErrors) > 0 {
		errDetails := map[string]interface{}{
			"errors": v.formatValidationErrors(validationErrors),
		}
		return errors.NewValidationError("Turn 2 context validation failed", errDetails)
	}

	// Log warnings
	if len(validationWarnings) > 0 {
		v.log.Warn("Turn 2 context validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(validationWarnings),
		})
	}

	return nil
}

// ValidateExtractedState validates the extracted state data
func (v *Validator) ValidateExtractedState(state *types.ExtractedState) error {
	if state == nil {
		return errors.NewValidationError("extracted state is nil", nil)
	}

	var validationErrors []ValidationError
	var validationWarnings []ValidationWarning

	// Check if machine structure is present
	if state.MachineStructure == nil {
		validationWarnings = append(validationWarnings, ValidationWarning{
			Field:   "machineStructure",
			Message: "machine structure is missing",
			Impact:  "medium",
		})
	} else {
		// Convert internal MachineStructure to schema.MachineStructure for validation
		schemaStructure := &schema.MachineStructure{
			RowCount:      state.MachineStructure.RowCount,
			ColumnsPerRow: state.MachineStructure.ColumnsPerRow,
			RowOrder:      state.MachineStructure.RowOrder,
			ColumnOrder:   state.MachineStructure.ColumnOrder,
		}
		
		if err := v.ValidateMachineStructure(schemaStructure); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:    "machineStructure",
				Message:  err.Error(),
				Severity: "high",
			})
		}
	}

	// Check row states
	if state.RowStates == nil || len(state.RowStates) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:    "rowStates",
			Message:  "row states are missing",
			Severity: "high",
		})
	} else if state.MachineStructure != nil {
		// Check if all rows defined in machine structure have a state
		for _, row := range state.MachineStructure.RowOrder {
			if _, exists := state.RowStates[row]; !exists {
				validationErrors = append(validationErrors, ValidationError{
					Field:    fmt.Sprintf("rowStates.%s", row),
					Message:  "row state missing for defined row",
					Severity: "medium",
				})
			}
		}

		// Check each row state
		for rowId, rowState := range state.RowStates {
			if rowState == nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:    fmt.Sprintf("rowStates.%s", rowId),
					Message:  "row state is nil",
					Severity: "high",
				})
				continue
			}

			// Validate status
			if rowState.Status == "" {
				validationWarnings = append(validationWarnings, ValidationWarning{
					Field:   fmt.Sprintf("rowStates.%s.status", rowId),
					Message: "status is empty",
					Impact:  "low",
				})
			}

			// Validate consistency of filled/empty positions with quantity
			filledCount := len(rowState.FilledPositions)
			emptyCount := len(rowState.EmptyPositions)
			
			if rowState.Quantity != filledCount {
				validationWarnings = append(validationWarnings, ValidationWarning{
					Field:   fmt.Sprintf("rowStates.%s.quantity", rowId),
					Message: fmt.Sprintf("quantity (%d) does not match filled positions count (%d)", rowState.Quantity, filledCount),
					Impact:  "medium",
				})
			}

			// Check for position consistency if machine structure is available
			if state.MachineStructure != nil {
				expectedTotal := state.MachineStructure.ColumnsPerRow
				if filledCount+emptyCount != expectedTotal {
					validationWarnings = append(validationWarnings, ValidationWarning{
						Field:   fmt.Sprintf("rowStates.%s", rowId),
						Message: fmt.Sprintf("total positions (%d) does not match expected column count (%d)", filledCount+emptyCount, expectedTotal),
						Impact:  "medium",
					})
				}
			}
		}
	}

	// Validate total counts
	if state.TotalEmptyCount != len(state.EmptyPositions) {
		validationWarnings = append(validationWarnings, ValidationWarning{
			Field:   "totalEmptyCount",
			Message: fmt.Sprintf("total empty count (%d) does not match empty positions length (%d)", state.TotalEmptyCount, len(state.EmptyPositions)),
			Impact:  "low",
		})
	}

	if state.TotalFilledCount != len(state.FilledPositions) {
		validationWarnings = append(validationWarnings, ValidationWarning{
			Field:   "totalFilledCount",
			Message: fmt.Sprintf("total filled count (%d) does not match filled positions length (%d)", state.TotalFilledCount, len(state.FilledPositions)),
			Impact:  "low",
		})
	}

	// Validate overall status
	if state.OverallStatus == "" {
		validationWarnings = append(validationWarnings, ValidationWarning{
			Field:   "overallStatus",
			Message: "overall status is empty",
			Impact:  "low",
		})
	}

	if len(validationErrors) > 0 {
		errDetails := map[string]interface{}{
			"errors": v.formatValidationErrors(validationErrors),
		}
		return errors.NewValidationError("extracted state validation failed", errDetails)
	}

	// Log warnings
	if len(validationWarnings) > 0 {
		v.log.Warn("Extracted state validation completed with warnings", map[string]interface{}{
			"warnings": v.formatValidationWarnings(validationWarnings),
		})
	}

	return nil
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
	case "VALIDATION_FLOW":
		v.validateReferenceValidation(analysis, &errors, &warnings)
	case "HISTORICAL_ENHANCEMENT":
		v.validateHistoricalEnhancement(analysis, &errors, &warnings)
	case "FRESH_EXTRACTION":
		v.validateFreshExtraction(analysis, &errors, &warnings)
	default:
		warnings = append(warnings, ValidationWarning{
			Field:   "sourceType",
			Message: fmt.Sprintf("unknown source type: %s", sourceType),
			Impact:  "medium",
		})
	}

	// Validate context for Turn 2
	v.validateContextForTurn2Fields(analysis, &errors, &warnings)

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

	// Check for visual confirmation
	if _, exists := analysis["visualConfirmation"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "visualConfirmation",
			Message:  "visual confirmation missing for enhanced analysis",
			Severity: "high",
		})
	}

	// Check for enriched baseline
	if _, exists := analysis["enrichedBaseline"]; !exists {
		*errors = append(*errors, ValidationError{
			Field:    "enrichedBaseline",
			Message:  "enriched baseline missing for enhanced analysis",
			Severity: "high",
		})
	}

	// Validate machine structure if present
	if machineStructureMap, exists := analysis["machineStructure"].(map[string]interface{}); exists {
		// Convert to schema.MachineStructure for validation
		structure := &schema.MachineStructure{}
		
		if rowCount, ok := machineStructureMap["rowCount"].(float64); ok {
			structure.RowCount = int(rowCount)
		}
		
		if columnsPerRow, ok := machineStructureMap["columnsPerRow"].(float64); ok {
			structure.ColumnsPerRow = int(columnsPerRow)
		}
		
		if rowOrder, ok := machineStructureMap["rowOrder"].([]interface{}); ok {
			for _, row := range rowOrder {
				if rowStr, ok := row.(string); ok {
					structure.RowOrder = append(structure.RowOrder, rowStr)
				}
			}
		}
		
		if columnOrder, ok := machineStructureMap["columnOrder"].([]interface{}); ok {
			for _, col := range columnOrder {
				if colStr, ok := col.(string); ok {
					structure.ColumnOrder = append(structure.ColumnOrder, colStr)
				}
			}
		}
		
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

// validateContextForTurn2Fields validates the context prepared for Turn 2 fields
func (v *Validator) validateContextForTurn2Fields(analysis map[string]interface{}, errors *[]ValidationError, warnings *[]ValidationWarning) {
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
				Message: "required context field missing",
				Impact:  "medium",
			})
		}
	}

	// Check readyForTurn2 value if it exists
	if readyForTurn2, ok := context["readyForTurn2"].(bool); ok {
		if !readyForTurn2 {
			*errors = append(*errors, ValidationError{
				Field:    "contextForTurn2.readyForTurn2",
				Message:  "readyForTurn2 must be true",
				Severity: "high",
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

// getRequiredFieldsForPath returns required fields based on processing path
func (v *Validator) getRequiredFieldsForPath(path types.ProcessingPath) []string {
	switch path {
	case types.PathValidationFlow:
		return []string{"status", "sourceType", "validationResults", "contextForTurn2"}
	case types.PathHistoricalEnhancement:
		return []string{"status", "sourceType", "historicalBaseline", "visualConfirmation", "enrichedBaseline", "contextForTurn2"}
	case types.PathFreshExtraction:
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

// calculateCompleteness calculates data completeness percentage
func (v *Validator) calculateCompleteness(analysis map[string]interface{}, sourceType string) float64 {
	var requiredFields []string
	
	switch sourceType {
	case "VALIDATION_FLOW":
		requiredFields = v.getRequiredFieldsForPath(types.PathValidationFlow)
	case "HISTORICAL_ENHANCEMENT":
		requiredFields = v.getRequiredFieldsForPath(types.PathHistoricalEnhancement)
	case "FRESH_EXTRACTION":
		requiredFields = v.getRequiredFieldsForPath(types.PathFreshExtraction)
	default:
		requiredFields = []string{"status", "sourceType"}
	}
	
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