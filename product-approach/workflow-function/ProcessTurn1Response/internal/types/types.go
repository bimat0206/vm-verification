// Package types provides the ProcessTurn1Response Lambda function types
package types

import (
	"time"
	"workflow-function/shared/schema"
)

// ProcessingConfig defines configuration for Turn 1 response processing
type ProcessingConfig struct {
	// ExtractMachineStructure enables machine structure extraction
	ExtractMachineStructure bool `json:"extractMachineStructure"`
	
	// ValidateCompleteness enables completeness validation
	ValidateCompleteness bool `json:"validateCompleteness"`
	
	// RequireProductMapping indicates if product mapping is required
	RequireProductMapping bool `json:"requireProductMapping"`
	
	// FallbackToTextParsing enables text parsing as fallback
	FallbackToTextParsing bool `json:"fallbackToTextParsing"`
	
	// StrictValidation enforces strict validation rules
	StrictValidation bool `json:"strictValidation"`
	
	// MaxResponseSize limits the size of response content to process
	MaxResponseSize int64 `json:"maxResponseSize"`
}

// ProcessingPath represents the processing path for Turn 1 response
type ProcessingPath string

const (
	// PathValidationFlow - UC1: Simple validation flow for LAYOUT_VS_CHECKING
	PathValidationFlow ProcessingPath = "VALIDATION_FLOW"
	
	// PathHistoricalEnhancement - UC2: Enhancement with historical context
	PathHistoricalEnhancement ProcessingPath = "HISTORICAL_ENHANCEMENT"
	
	// PathFreshExtraction - UC2: Fresh extraction without historical context
	PathFreshExtraction ProcessingPath = "FRESH_EXTRACTION"
)

// Turn1ProcessingResult represents the result of Turn 1 processing
type Turn1ProcessingResult struct {
	// Status of the processing operation
	Status string `json:"status"`
	
	// SourceType indicates the processing path used
	SourceType ProcessingPath `json:"sourceType"`
	
	// ExtractedStructure contains machine structure information
	ExtractedStructure *MachineStructure `json:"extractedStructure,omitempty"`
	
	// ReferenceAnalysis contains the processed reference analysis
	ReferenceAnalysis map[string]interface{} `json:"referenceAnalysis"`
	
	// ContextForTurn2 contains prepared context for Turn 2
	ContextForTurn2 map[string]interface{} `json:"contextForTurn2"`
	
	// ProcessingMetadata contains metadata about the processing
	ProcessingMetadata *ProcessingMetadata `json:"processingMetadata"`
	
	// Warnings contains any warnings during processing
	Warnings []string `json:"warnings,omitempty"`
	
	// FallbackUsed indicates if fallback parsing was used
	FallbackUsed bool `json:"fallbackUsed"`
}

// MachineStructure represents the vending machine structure
type MachineStructure struct {
	// RowCount is the number of rows in the machine
	RowCount int `json:"rowCount"`
	
	// ColumnsPerRow is the number of columns per row
	ColumnsPerRow int `json:"columnsPerRow"`
	
	// RowOrder defines the order of rows (e.g., ["A", "B", "C"])
	RowOrder []string `json:"rowOrder"`
	
	// ColumnOrder defines the order of columns (e.g., ["1", "2", "3"])
	ColumnOrder []string `json:"columnOrder"`
	
	// TotalPositions is the total number of positions
	TotalPositions int `json:"totalPositions"`
	
	// StructureConfirmed indicates if the structure was confirmed
	StructureConfirmed bool `json:"structureConfirmed"`
}

// RowState represents the state of a single row
type RowState struct {
	// Status of the row (Full/Partial/Empty)
	Status string `json:"status"`
	
	// FilledPositions lists positions that contain products
	FilledPositions []string `json:"filledPositions"`
	
	// EmptyPositions lists positions that are empty
	EmptyPositions []string `json:"emptyPositions"`
	
	// ProductType describes the type of product in the row
	ProductType string `json:"productType,omitempty"`
	
	// ProductColor describes the color/appearance of products
	ProductColor string `json:"productColor,omitempty"`
	
	// Quantity is the number of products in the row
	Quantity int `json:"quantity"`
	
	// Notes contains additional observations about the row
	Notes string `json:"notes,omitempty"`
}

// ParsedResponse represents a parsed Bedrock response
type ParsedResponse struct {
	// MainContent is the primary response content
	MainContent string `json:"mainContent"`
	
	// ThinkingContent is the reasoning/thinking section
	ThinkingContent string `json:"thinkingContent,omitempty"`
	
	// IsStructured indicates if the response is in structured format
	IsStructured bool `json:"isStructured"`
	
	// ParsedSections contains identified sections of the response
	ParsedSections map[string]string `json:"parsedSections"`
	
	// ExtractedData contains any structured data found
	ExtractedData map[string]interface{} `json:"extractedData"`
	
	// ParsingErrors contains any errors during parsing
	ParsingErrors []string `json:"parsingErrors,omitempty"`
}

// ProcessingMetadata contains metadata about the processing operation
type ProcessingMetadata struct {
	// ProcessingStartTime when processing started
	ProcessingStartTime time.Time `json:"processingStartTime"`
	
	// ProcessingEndTime when processing completed
	ProcessingEndTime time.Time `json:"processingEndTime"`
	
	// ProcessingDuration total time taken
	ProcessingDuration time.Duration `json:"processingDuration"`
	
	// ResponseSize size of the response content
	ResponseSize int64 `json:"responseSize"`
	
	// ExtractedElements number of elements extracted
	ExtractedElements int `json:"extractedElements"`
	
	// ValidationsPassed number of validations that passed
	ValidationsPassed int `json:"validationsPassed"`
	
	// ValidationsFailed number of validations that failed
	ValidationsFailed int `json:"validationsFailed"`
	
	// ProcessingPath the path used for processing
	ProcessingPath ProcessingPath `json:"processingPath"`
	
	// FallbackReason reason for using fallback parsing
	FallbackReason string `json:"fallbackReason,omitempty"`
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	// Valid indicates if the validation passed
	Valid bool `json:"valid"`
	
	// ErrorMessage contains error details if validation failed
	ErrorMessage string `json:"errorMessage,omitempty"`
	
	// WarningMessage contains warning details
	WarningMessage string `json:"warningMessage,omitempty"`
	
	// FieldName is the name of the field being validated
	FieldName string `json:"fieldName"`
	
	// ExpectedValue is what was expected
	ExpectedValue interface{} `json:"expectedValue,omitempty"`
	
	// ActualValue is what was found
	ActualValue interface{} `json:"actualValue,omitempty"`
}

// ParsingContext holds context information for parsing operations
type ParsingContext struct {
	// VerificationType indicates the type of verification
	VerificationType string `json:"verificationType"`
	
	// HasHistoricalContext indicates if historical context is available
	HasHistoricalContext bool `json:"hasHistoricalContext"`
	
	// LayoutMetadata contains layout information
	LayoutMetadata map[string]interface{} `json:"layoutMetadata,omitempty"`
	
	// HistoricalContext contains historical verification data
	HistoricalContext map[string]interface{} `json:"historicalContext,omitempty"`
	
	// ExpectedStructure contains expected machine structure
	ExpectedStructure *MachineStructure `json:"expectedStructure,omitempty"`
	
	// ParsingConfig configuration for parsing
	ParsingConfig *ProcessingConfig `json:"parsingConfig"`
}

// ExtractedState represents the state extracted from the response
type ExtractedState struct {
	// MachineStructure contains the machine layout structure
	MachineStructure *MachineStructure `json:"machineStructure,omitempty"`
	
	// RowStates maps row identifiers to their states
	RowStates map[string]*RowState `json:"rowStates,omitempty"`
	
	// EmptyPositions lists all empty positions found
	EmptyPositions []string `json:"emptyPositions,omitempty"`
	
	// FilledPositions lists all filled positions found
	FilledPositions []string `json:"filledPositions,omitempty"`
	
	// TotalEmptyCount total number of empty positions
	TotalEmptyCount int `json:"totalEmptyCount"`
	
	// TotalFilledCount total number of filled positions
	TotalFilledCount int `json:"totalFilledCount"`
	
	// OverallStatus overall status of the machine
	OverallStatus string `json:"overallStatus"`
	
	// Observations general observations about the state
	Observations []string `json:"observations,omitempty"`
}

// HistoricalEnhancement represents enhanced analysis with historical data
type HistoricalEnhancement struct {
	// BaselineData from historical context
	BaselineData map[string]interface{} `json:"baselineData"`
	
	// VisualConfirmation confirmations from visual analysis
	VisualConfirmation map[string]interface{} `json:"visualConfirmation"`
	
	// NewObservations observations not in historical data
	NewObservations []string `json:"newObservations,omitempty"`
	
	// Discrepancies between historical and visual analysis
	Discrepancies []string `json:"discrepancies,omitempty"`
	
	// EnrichedBaseline combined historical and visual data
	EnrichedBaseline map[string]interface{} `json:"enrichedBaseline"`
}

// ValidationType represents different types of validation
type ValidationType string

const (
	// ValidationTypeStructure validates machine structure
	ValidationTypeStructure ValidationType = "structure"
	
	// ValidationTypeCompleteness validates data completeness
	ValidationTypeCompleteness ValidationType = "completeness"
	
	// ValidationTypeConsistency validates data consistency
	ValidationTypeConsistency ValidationType = "consistency"
	
	// ValidationTypeFormat validates data format
	ValidationTypeFormat ValidationType = "format"
	
	// ValidationTypeRequired validates required fields
	ValidationTypeRequired ValidationType = "required"
)

// ResponsePattern represents patterns for parsing responses
type ResponsePattern struct {
	// Name of the pattern
	Name string `json:"name"`
	
	// Regex pattern for matching
	Pattern string `json:"pattern"`
	
	// Description of what the pattern matches
	Description string `json:"description"`
	
	// Required indicates if this pattern must be found
	Required bool `json:"required"`
	
	// ExtractGroups specifies which regex groups to extract
	ExtractGroups []string `json:"extractGroups"`
}

// DefaultProcessingConfig returns default processing configuration
func DefaultProcessingConfig() *ProcessingConfig {
	return &ProcessingConfig{
		ExtractMachineStructure: true,
		ValidateCompleteness:    true,
		RequireProductMapping:   false,
		FallbackToTextParsing:   true,
		StrictValidation:        false,
		MaxResponseSize:         1024 * 1024, // 1MB
	}
}

// GetProcessingConfigForUseCase returns configuration for specific use case
func GetProcessingConfigForUseCase(verificationType string, hasHistorical bool) *ProcessingConfig {
	config := DefaultProcessingConfig()
	
	switch verificationType {
	case schema.VerificationTypeLayoutVsChecking:
		// UC1: Simpler validation flow
		config.ExtractMachineStructure = false
		config.RequireProductMapping = true
		config.StrictValidation = true
		
	case schema.VerificationTypePreviousVsCurrent:
		if hasHistorical {
			// UC2 with historical: Focus on enhancement
			config.ExtractMachineStructure = false
			config.ValidateCompleteness = false
		} else {
			// UC2 without historical: Full extraction
			config.ExtractMachineStructure = true
			config.ValidateCompleteness = true
		}
	}
	
	return config
}

// String returns string representation of ProcessingPath
func (p ProcessingPath) String() string {
	return string(p)
}

// IsValid checks if the ProcessingPath is valid
func (p ProcessingPath) IsValid() bool {
	return p == PathValidationFlow || 
		   p == PathHistoricalEnhancement || 
		   p == PathFreshExtraction
}

// GetExpectedRowStates returns expected row states for validation
func (ms *MachineStructure) GetExpectedRowStates() map[string]*RowState {
	states := make(map[string]*RowState)
	
	for _, row := range ms.RowOrder {
		states[row] = &RowState{
			Status:          "",
			FilledPositions: []string{},
			EmptyPositions:  []string{},
			Quantity:        0,
		}
	}
	
	return states
}

// Validate validates the machine structure
func (ms *MachineStructure) Validate() []ValidationResult {
	results := []ValidationResult{}
	
	// Validate row count
	if ms.RowCount <= 0 {
		results = append(results, ValidationResult{
			Valid:        false,
			ErrorMessage: "Row count must be positive",
			FieldName:    "RowCount",
			ActualValue:  ms.RowCount,
		})
	}
	
	// Validate columns per row
	if ms.ColumnsPerRow <= 0 {
		results = append(results, ValidationResult{
			Valid:        false,
			ErrorMessage: "Columns per row must be positive",
			FieldName:    "ColumnsPerRow",
			ActualValue:  ms.ColumnsPerRow,
		})
	}
	
	// Validate row order length matches row count
	if len(ms.RowOrder) != ms.RowCount {
		results = append(results, ValidationResult{
			Valid:        false,
			ErrorMessage: "Row order length must match row count",
			FieldName:    "RowOrder",
			ExpectedValue: ms.RowCount,
			ActualValue:   len(ms.RowOrder),
		})
	}
	
	// Validate column order length matches columns per row
	if len(ms.ColumnOrder) != ms.ColumnsPerRow {
		results = append(results, ValidationResult{
			Valid:        false,
			ErrorMessage: "Column order length must match columns per row",
			FieldName:    "ColumnOrder",
			ExpectedValue: ms.ColumnsPerRow,
			ActualValue:   len(ms.ColumnOrder),
		})
	}
	
	// Validate total positions
	expectedTotal := ms.RowCount * ms.ColumnsPerRow
	if ms.TotalPositions != 0 && ms.TotalPositions != expectedTotal {
		results = append(results, ValidationResult{
			Valid:        false,
			ErrorMessage: "Total positions doesn't match row count * columns per row",
			FieldName:    "TotalPositions",
			ExpectedValue: expectedTotal,
			ActualValue:   ms.TotalPositions,
		})
	}
	
	return results
}

// CalculateTotalPositions calculates and sets the total positions
func (ms *MachineStructure) CalculateTotalPositions() {
	ms.TotalPositions = ms.RowCount * ms.ColumnsPerRow
}

// GetPosition returns the position identifier for given row and column indices
func (ms *MachineStructure) GetPosition(rowIndex, colIndex int) string {
	if rowIndex < 0 || rowIndex >= len(ms.RowOrder) ||
	   colIndex < 0 || colIndex >= len(ms.ColumnOrder) {
		return ""
	}
	
	return ms.RowOrder[rowIndex] + ms.ColumnOrder[colIndex]
}

// AddWarning adds a warning to the processing result
func (r *Turn1ProcessingResult) AddWarning(warning string) {
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	r.Warnings = append(r.Warnings, warning)
}

// HasWarnings checks if there are any warnings
func (r *Turn1ProcessingResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// IsSuccessful checks if the processing was successful
func (r *Turn1ProcessingResult) IsSuccessful() bool {
	return r.Status == "EXTRACTION_COMPLETE" || r.Status == "VALIDATION_COMPLETE"
}
