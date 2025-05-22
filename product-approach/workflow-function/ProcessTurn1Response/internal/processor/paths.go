package processor

import (
	"context"
	"log/slog"
	"strings"

	"workflow-function/ProcessTurn1Response/internal/parser"
	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/ProcessTurn1Response/internal/validator"
)

// PathProcessor defines the interface for specialized path processors
type PathProcessor interface {
	// Process performs specialized processing for a specific path
	Process(ctx context.Context, responseContent string, historicalData *types.HistoricalEnhancement, result *types.Turn1ProcessingResult) error
}

// BasePathProcessor provides common functionality for all path processors
type BasePathProcessor struct {
	logger    *slog.Logger
	parser    parser.Parser
	validator validator.ValidatorInterface
}

// ValidationFlowProcessor specializes in UC1 validation flow
type ValidationFlowProcessor struct {
	BasePathProcessor
}

// NewValidationFlowProcessor creates a new ValidationFlowProcessor
func NewValidationFlowProcessor(logger *slog.Logger, p parser.Parser, v validator.ValidatorInterface) *ValidationFlowProcessor {
	return &ValidationFlowProcessor{
		BasePathProcessor: BasePathProcessor{
			logger:    logger,
			parser:    p,
			validator: v,
		},
	}
}

// Process implements the UC1 validation flow
func (p *ValidationFlowProcessor) Process(
	ctx context.Context,
	responseContent string, 
	_ *types.HistoricalEnhancement,
	result *types.Turn1ProcessingResult,
) error {
	p.logger.Info("Processing validation flow")
	
	// Parse validation response
	validationResults, err := p.parser.ParseValidationResponse(responseContent)
	if err != nil {
		return NewParsingError("Failed to parse validation response", err)
	}
	
	// Extract observations
	observations, err := p.parser.ExtractObservations(responseContent)
	if err != nil {
		p.logger.Warn("Failed to extract observations, continuing without them", "error", err.Error())
		// Continue without observations
		observations = []string{}
	}
	
	// Create context for Turn2
	contextForTurn2 := map[string]interface{}{
		"referenceValidated":       true,
		"useSystemPromptReference": true,
		"validationPassed":         getValidationPassed(validationResults),
		"readyForTurn2":            true,
	}
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":            "VALIDATION_COMPLETE",
		"sourceType":        "REFERENCE_VALIDATION",
		"validationResults": validationResults,
		"basicObservations": observations,
		"contextForTurn2":   contextForTurn2,
	}
	
	// Update result
	result.ReferenceAnalysis = analysisData
	result.ContextForTurn2 = contextForTurn2
	
	// Update metadata
	result.ProcessingMetadata.ExtractedElements = len(validationResults) + len(observations)
	
	return nil
}

// HistoricalEnhancementProcessor specializes in UC2 with historical context
type HistoricalEnhancementProcessor struct {
	BasePathProcessor
}

// NewHistoricalEnhancementProcessor creates a new HistoricalEnhancementProcessor
func NewHistoricalEnhancementProcessor(logger *slog.Logger, p parser.Parser, v validator.ValidatorInterface) *HistoricalEnhancementProcessor {
	return &HistoricalEnhancementProcessor{
		BasePathProcessor: BasePathProcessor{
			logger:    logger,
			parser:    p,
			validator: v,
		},
	}
}

// Process implements the UC2 historical enhancement flow
func (p *HistoricalEnhancementProcessor) Process(
	ctx context.Context,
	responseContent string,
	historicalData *types.HistoricalEnhancement,
	result *types.Turn1ProcessingResult,
) error {
	p.logger.Info("Processing historical enhancement")
	
	// Parse visual analysis from response
	visualAnalysis, err := p.parser.ParseVisualAnalysis(responseContent)
	if err != nil {
		return NewParsingError("Failed to parse visual analysis", err)
	}
	
	if historicalData == nil {
		p.logger.Warn("Historical data is nil, creating empty structure")
		historicalData = &types.HistoricalEnhancement{
			BaselineData:       make(map[string]interface{}),
			VisualConfirmation: make(map[string]interface{}),
			EnrichedBaseline:   make(map[string]interface{}),
		}
	}
	
	// Add visual confirmation to historical data
	historicalData.VisualConfirmation = visualAnalysis
	
	// Combine historical and visual data
	enhancedBaseline := p.enhanceHistoricalBaseline(historicalData)
	
	// Create context for Turn2
	focusAreas := p.identifyFocusAreas(historicalData.BaselineData)
	knownIssues := p.extractKnownIssuesList(historicalData.BaselineData)
	
	contextForTurn2 := map[string]interface{}{
		"baselineEstablished":     true,
		"useHistoricalAsReference": true,
		"focusAreas":              focusAreas,
		"knownIssues":             knownIssues,
		"enrichedBaseline":        enhancedBaseline,
		"readyForTurn2":           true,
	}
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":             "EXTRACTION_COMPLETE",
		"sourceType":         "HISTORICAL_WITH_VISUAL_CONFIRMATION",
		"historicalBaseline": historicalData.BaselineData,
		"visualConfirmation": visualAnalysis,
		"enhancedBaseline":   enhancedBaseline,
		"contextForTurn2":    contextForTurn2,
	}
	
	// Update result
	result.ReferenceAnalysis = analysisData
	result.ContextForTurn2 = contextForTurn2
	
	// Update metadata
	result.ProcessingMetadata.ExtractedElements = len(visualAnalysis) + len(enhancedBaseline)
	
	return nil
}

// FreshExtractionProcessor specializes in UC2 without historical context
type FreshExtractionProcessor struct {
	BasePathProcessor
}

// NewFreshExtractionProcessor creates a new FreshExtractionProcessor
func NewFreshExtractionProcessor(logger *slog.Logger, p parser.Parser, v validator.ValidatorInterface) *FreshExtractionProcessor {
	return &FreshExtractionProcessor{
		BasePathProcessor: BasePathProcessor{
			logger:    logger,
			parser:    p,
			validator: v,
		},
	}
}

// Process implements the UC2 fresh extraction flow
func (p *FreshExtractionProcessor) Process(
	ctx context.Context,
	responseContent string,
	_ *types.HistoricalEnhancement,
	result *types.Turn1ProcessingResult,
) error {
	p.logger.Info("Processing fresh extraction")
	
	// Parse machine structure from response
	machineStructure, err := p.parser.ParseMachineStructure(responseContent)
	if err != nil {
		return NewParsingError("Failed to parse machine structure", err)
	}
	
	
	// Parse machine state from response
	machineState, err := p.parser.ParseMachineState(responseContent)
	if err != nil {
		return NewParsingError("Failed to parse machine state", err)
	}
	
	// Create context for Turn2
	contextForTurn2 := map[string]interface{}{
		"baselineSource":         "EXTRACTED_STATE",
		"useHistoricalData":      false,
		"extractedDataAvailable": true,
		"readyForTurn2":          true,
		"extractedStructure":     machineStructure,
		"extractedState":         machineState,
	}
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":             "EXTRACTION_COMPLETE",
		"sourceType":         "FRESH_VISUAL_ANALYSIS",
		"extractedStructure": machineStructure,
		"extractedState":     machineState,
		"contextForTurn2":    contextForTurn2,
	}
	
	// Update result
	result.ReferenceAnalysis = analysisData
	result.ContextForTurn2 = contextForTurn2
	
	// Update metadata
	result.ProcessingMetadata.ExtractedElements = countStructureElements(machineStructure) + countStateElements(machineState)
	
	return nil
}

// Helper functions for HistoricalEnhancementProcessor

// enhanceHistoricalBaseline combines historical data with visual analysis
func (p *HistoricalEnhancementProcessor) enhanceHistoricalBaseline(historical *types.HistoricalEnhancement) map[string]interface{} {
	enhanced := make(map[string]interface{})
	
	// Copy historical data
	for key, value := range historical.BaselineData {
		enhanced[key] = value
	}
	
	// Add visual confirmations
	if historical.VisualConfirmation != nil {
		enhanced["visualConfirmation"] = historical.VisualConfirmation
		enhanced["enhancementTimestamp"] = FormatTime(TimeNow())
	}
	
	return enhanced
}

// identifyFocusAreas identifies areas that need special attention in Turn 2
func (p *HistoricalEnhancementProcessor) identifyFocusAreas(context map[string]interface{}) []string {
	focusAreas := []string{}
	
	// Extract areas with known issues
	if checkingStatus, ok := context["checkingStatus"].(map[string]interface{}); ok {
		for key, value := range checkingStatus {
			if status, ok := value.(string); ok {
				if contains(status, "empty") ||
				   contains(status, "incorrect") ||
				   contains(status, "changed") {
					focusAreas = append(focusAreas, key)
				}
			}
		}
	}
	
	return focusAreas
}

// extractKnownIssuesList extracts a list of known issue types
func (p *HistoricalEnhancementProcessor) extractKnownIssuesList(context map[string]interface{}) []string {
	issues := []string{}
	
	if summary, ok := context["verificationSummary"].(map[string]interface{}); ok {
		if status, ok := summary["verificationStatus"].(string); ok && status != "CORRECT" {
			// Parse outcome for issue types
			if outcome, ok := summary["verificationOutcome"].(string); ok {
				if contains(outcome, "empty") {
					issues = append(issues, "empty_rows")
				}
				if contains(outcome, "incorrect") {
					issues = append(issues, "incorrect_products")
				}
			}
		}
	}
	
	return issues
}

// Utility functions

// contains checks if a string contains a substring (case-insensitive)
func contains(s string, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

// getValidationPassed determines if validation passed from results
func getValidationPassed(results map[string]interface{}) bool {
	if structureConfirmed, ok := results["structureConfirmed"].(bool); ok {
		return structureConfirmed
	}
	
	// Default to false if not found or not a boolean
	return false
}

// mapSize returns the size of a map safely
func mapSize(m map[string]interface{}) int {
	if m == nil {
		return 0
	}
	return len(m)
}

// sliceSize returns the size of a slice safely
func sliceSize(s []interface{}) int {
	if s == nil {
		return 0
	}
	return len(s)
}

// countStructureElements counts the elements in a MachineStructure
func countStructureElements(ms *types.MachineStructure) int {
	if ms == nil {
		return 0
	}
	// Count important elements: rows, columns, and properties
	elements := 0
	elements += len(ms.RowOrder)
	elements += len(ms.ColumnOrder)
	if ms.RowCount > 0 {
		elements++
	}
	if ms.ColumnsPerRow > 0 {
		elements++
	}
	if ms.TotalPositions > 0 {
		elements++
	}
	return elements
}

// countStateElements counts the elements in an ExtractedState
func countStateElements(state *types.ExtractedState) int {
	if state == nil {
		return 0
	}
	// Count important elements: states, empty/filled positions, and observations
	elements := 0
	if state.RowStates != nil {
		elements += len(state.RowStates)
	}
	if state.EmptyPositions != nil {
		elements += len(state.EmptyPositions)
	}
	if state.FilledPositions != nil {
		elements += len(state.FilledPositions)
	}
	if state.Observations != nil {
		elements += len(state.Observations)
	}
	return elements
}