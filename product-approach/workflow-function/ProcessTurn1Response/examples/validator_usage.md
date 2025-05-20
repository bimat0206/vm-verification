// Package examples provides usage examples for ProcessTurn1Response components
package examples

import (
	"fmt"

	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/ProcessTurn1Response/internal/validator"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// This example demonstrates how to use the new validator implementation
// to validate different components of the Turn 1 response processing.

func ValidatorUsageExample() {
	// Create a logger
	log := logger.NewLogger()

	// There are multiple ways to create a validator:

	// 1. Direct instantiation
	val := validator.NewValidator(log)

	// 2. Using helper function (recommended)
	// val := validator.GetValidator(log)

	// 3. Using the legacy interface (for backward compatibility)
	// legacyVal := validator.GetLegacyValidator(log)

	// Example 1: Validate Machine Structure
	fmt.Println("Example 1: Validating Machine Structure")
	machineStructure := &schema.MachineStructure{
		RowCount:      3,
		ColumnsPerRow: 4,
		RowOrder:      []string{"A", "B", "C"},
		ColumnOrder:   []string{"1", "2", "3", "4"},
	}

	err := val.ValidateMachineStructure(machineStructure)
	if err != nil {
		fmt.Printf("Machine structure validation failed: %v\n", err)
	} else {
		fmt.Println("Machine structure validation passed")
	}

	// Example 2: Validate Processing Result
	fmt.Println("\nExample 2: Validating Processing Result")
	processingResult := &types.Turn1ProcessingResult{
		Status:     "VALIDATION_COMPLETE",
		SourceType: types.PathValidationFlow,
		ReferenceAnalysis: map[string]interface{}{
			"validationResults": map[string]interface{}{
				"isValid": true,
				"details": "Structure matches expected layout",
			},
			"basicObservations": []string{
				"Machine layout appears to match the reference",
				"All rows are visible and labeled correctly",
			},
		},
		ContextForTurn2: map[string]interface{}{
			"readyForTurn2":           true,
			"referenceAnalysisComplete": true,
			"analysisType":            "VALIDATION",
		},
		ProcessingMetadata: &types.ProcessingMetadata{
			ValidationsPassed:   3,
			ValidationsFailed:   0,
			ProcessingPath:      types.PathValidationFlow,
		},
	}

	err = val.ValidateProcessingResult(processingResult)
	if err != nil {
		fmt.Printf("Processing result validation failed: %v\n", err)
	} else {
		fmt.Println("Processing result validation passed")
	}

	// Example 3: Validate Historical Enhancement
	fmt.Println("\nExample 3: Validating Historical Enhancement")
	historicalEnhancement := map[string]interface{}{
		"baselineData": map[string]interface{}{
			"previousState": "Machine was 80% full during previous verification",
			"lastVerified": "2025-05-15T10:30:00Z",
		},
		"visualConfirmation": map[string]interface{}{
			"currentState": "Machine appears to be 60% full in the current image",
			"changes": []string{
				"Row A is now partially empty",
				"Row C is completely empty",
			},
		},
		"enrichedBaseline": map[string]interface{}{
			"currentState": "60% full",
			"changeDetails": map[string]interface{}{
				"emptiedPositions": []string{"A3", "A4", "C1", "C2", "C3", "C4"},
				"newlyFilledPositions": []string{},
			},
		},
	}

	err = val.ValidateHistoricalEnhancement(historicalEnhancement)
	if err != nil {
		fmt.Printf("Historical enhancement validation failed: %v\n", err)
	} else {
		fmt.Println("Historical enhancement validation passed")
	}

	// Example 4: Validate Context for Turn 2
	fmt.Println("\nExample 4: Validating Context for Turn 2")
	turn2Context := map[string]interface{}{
		"readyForTurn2":           true,
		"referenceAnalysisComplete": true,
		"analysisType":            "VALIDATION",
		"structureValidated":      true,
		"machineState":            "OPERATIONAL",
	}

	err = val.ValidateContextForTurn2(turn2Context)
	if err != nil {
		fmt.Printf("Turn 2 context validation failed: %v\n", err)
	} else {
		fmt.Println("Turn 2 context validation passed")
	}

	// Example 5: Check Completeness
	fmt.Println("\nExample 5: Checking Completeness")
	analysisData := map[string]interface{}{
		"status":     "EXTRACTION_COMPLETE",
		"sourceType": "FRESH_EXTRACTION",
		"extractedStructure": map[string]interface{}{
			"rowCount":      3,
			"columnsPerRow": 4,
			"rowOrder":      []string{"A", "B", "C"},
			"columnOrder":   []string{"1", "2", "3", "4"},
		},
		"extractedState": map[string]interface{}{
			"overallStatus": "PARTIAL",
			"rowStates": map[string]interface{}{
				"A": map[string]interface{}{
					"status":   "FULL",
					"quantity": 4,
				},
				"B": map[string]interface{}{
					"status":   "PARTIAL",
					"quantity": 2,
				},
				"C": map[string]interface{}{
					"status":   "EMPTY",
					"quantity": 0,
				},
			},
		},
		"contextForTurn2": map[string]interface{}{
			"readyForTurn2": true,
		},
	}

	completeness, errors, warnings := val.ValidateCompleteness(analysisData, types.PathFreshExtraction)
	fmt.Printf("Completeness: %.2f\n", completeness)
	fmt.Printf("Errors: %d\n", len(errors))
	fmt.Printf("Warnings: %d\n", len(warnings))

	// Example 6: Validate Extracted State
	fmt.Println("\nExample 6: Validating Extracted State")
	extractedState := &types.ExtractedState{
		MachineStructure: &types.MachineStructure{
			RowCount:      3,
			ColumnsPerRow: 4,
			RowOrder:      []string{"A", "B", "C"},
			ColumnOrder:   []string{"1", "2", "3", "4"},
			TotalPositions: 12,
		},
		RowStates: map[string]*types.RowState{
			"A": {
				Status:          "FULL",
				FilledPositions: []string{"A1", "A2", "A3", "A4"},
				EmptyPositions:  []string{},
				Quantity:        4,
				ProductType:     "Snack",
				ProductColor:    "Red",
			},
			"B": {
				Status:          "PARTIAL",
				FilledPositions: []string{"B1", "B2"},
				EmptyPositions:  []string{"B3", "B4"},
				Quantity:        2,
				ProductType:     "Drink",
				ProductColor:    "Blue",
			},
			"C": {
				Status:          "EMPTY",
				FilledPositions: []string{},
				EmptyPositions:  []string{"C1", "C2", "C3", "C4"},
				Quantity:        0,
			},
		},
		EmptyPositions:   []string{"B3", "B4", "C1", "C2", "C3", "C4"},
		FilledPositions:  []string{"A1", "A2", "A3", "A4", "B1", "B2"},
		TotalEmptyCount:  6,
		TotalFilledCount: 6,
		OverallStatus:    "PARTIAL",
		Observations:     []string{"Machine is partially filled"},
	}

	err = val.ValidateExtractedState(extractedState)
	if err != nil {
		fmt.Printf("Extracted state validation failed: %v\n", err)
	} else {
		fmt.Println("Extracted state validation passed")
	}

	// Example 7: Using Legacy Validator for backward compatibility
	fmt.Println("\nExample 7: Using Legacy Validator")
	legacyVal := validator.NewLegacyValidator(log)

	// Convert the machine structure to a map for the legacy validator
	machineStructureMap := map[string]interface{}{
		"rowCount":      3,
		"columnsPerRow": 4,
		"rowOrder":      []interface{}{"A", "B", "C"},
		"columnOrder":   []interface{}{"1", "2", "3", "4"},
	}

	err = legacyVal.ValidateMachineStructure(machineStructureMap)
	if err != nil {
		fmt.Printf("Legacy machine structure validation failed: %v\n", err)
	} else {
		fmt.Println("Legacy machine structure validation passed")
	}
}