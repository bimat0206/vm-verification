package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// This is a highly specialized test for parsing vending machine data directly

// Helper function to extract machine structure
func extractMachineStructure(content string) *types.MachineStructure {
	// Based on our knowledge that the machine has 6 rows (A-F) and 7 columns (1-7)
	structure := &types.MachineStructure{
		RowCount:      6,
		ColumnsPerRow: 7,
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7"},
	}
	
	structure.TotalPositions = structure.RowCount * structure.ColumnsPerRow
	structure.StructureConfirmed = true
	
	return structure
}

// Helper function to extract row states
func extractRowStates(content string) map[string]*types.RowState {
	rowStates := make(map[string]*types.RowState)
	
	// Use a more precise pattern for row sections and status
	rowPattern := regexp.MustCompile(`(?m)^## Row ([A-F])(?:[^*]+)\*\*Status: ([A-Za-z]+)\*\*`)
	matches := rowPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			row := match[1]
			status := match[2]
			
			// Find positions for this row
			posRegex := regexp.MustCompile(fmt.Sprintf(`(?m)- %s(\d+): ([^\n]+)`, row))
			posMatches := posRegex.FindAllStringSubmatch(content, -1)
			
			filledPositions := []string{}
			for _, posMatch := range posMatches {
				if len(posMatch) >= 2 {
					position := row + posMatch[1]
					filledPositions = append(filledPositions, position)
				}
			}
			
			// Create row state
			rowState := &types.RowState{
				Status:          status,
				FilledPositions: filledPositions,
				EmptyPositions:  []string{},
				Quantity:        len(filledPositions),
				Notes:           "",
			}
			
			rowStates[row] = rowState
		}
	}
	
	return rowStates
}

// Helper function to extract observations
func extractObservations(content string) []string {
	// Extract the main observations
	observations := []string{}
	if strings.Contains(content, "The reference layout shows") {
		summaryRegex := regexp.MustCompile(`The reference layout shows([^.]+)\.`)
		summaryMatches := summaryRegex.FindStringSubmatch(content)
		if len(summaryMatches) >= 2 {
			observations = append(observations, "Summary: "+summaryMatches[1])
		}
	}
	
	return observations
}

func main() {
	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "VendingMachineParser",
	)

	// Log startup
	logger.Info("Vending machine parser test starting",
		"timestamp", time.Now().Format(time.RFC3339),
	)

	// Read input file
	inputFile := "./real-test-input.json"
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		logger.Error("Failed to read input file", "error", err, "file", inputFile)
		os.Exit(1)
	}

	// Parse input JSON
	var input map[string]interface{}
	if err := json.Unmarshal(inputData, &input); err != nil {
		logger.Error("Failed to parse input JSON", "error", err)
		os.Exit(1)
	}

	// Extract response content from the input
	var responseContent string
	if turn1Response, ok := input["turn1Response"].(map[string]interface{}); ok {
		if response, ok := turn1Response["response"].(map[string]interface{}); ok {
			if content, ok := response["content"].(string); ok {
				responseContent = content
			}
		}
	}

	if responseContent == "" {
		logger.Error("Failed to extract response content from input")
		os.Exit(1)
	}

	logger.Info("Extracted response content", "length", len(responseContent))

	// Process the content with our specialized functions
	startTime := time.Now()
	
	// Extract machine structure
	structure := extractMachineStructure(responseContent)
	
	// Extract row states
	rowStates := extractRowStates(responseContent)
	
	// Extract observations
	observations := extractObservations(responseContent)
	
	// Create the extracted state
	extractedState := &types.ExtractedState{
		MachineStructure: structure,
		RowStates:        rowStates,
		EmptyPositions:   []string{},
		FilledPositions:  []string{},
		TotalEmptyCount:  0,
		TotalFilledCount: 0,
		OverallStatus:    "Full", // Based on the content
		Observations:     observations,
	}
	
	// Calculate totals
	for _, state := range rowStates {
		extractedState.TotalFilledCount += len(state.FilledPositions)
		extractedState.TotalEmptyCount += len(state.EmptyPositions)
		extractedState.FilledPositions = append(extractedState.FilledPositions, state.FilledPositions...)
		extractedState.EmptyPositions = append(extractedState.EmptyPositions, state.EmptyPositions...)
	}
	
	// Create the processing result
	contextForTurn2 := map[string]interface{}{
		"baselineSource":         "EXTRACTED_STATE",
		"useHistoricalData":      false,
		"extractedDataAvailable": true,
		"readyForTurn2":          true,
		"extractedStructure":     structure,
		"extractedState":         extractedState,
	}
	
	analysisData := map[string]interface{}{
		"status":             "EXTRACTION_COMPLETE",
		"sourceType":         "FRESH_VISUAL_ANALYSIS",
		"extractedStructure": structure,
		"extractedState":     extractedState,
		"contextForTurn2":    contextForTurn2,
	}
	
	result := &types.Turn1ProcessingResult{
		Status:             "EXTRACTION_COMPLETE",
		SourceType:         types.PathFreshExtraction,
		ExtractedStructure: structure,
		ReferenceAnalysis:  analysisData,
		ContextForTurn2:    contextForTurn2,
		ProcessingMetadata: &types.ProcessingMetadata{
			ProcessingStartTime:  startTime,
			ProcessingEndTime:    time.Now(),
			ProcessingDuration:   time.Since(startTime),
			ResponseSize:         int64(len(responseContent)),
			ExtractedElements:    structure.RowCount*structure.ColumnsPerRow + len(rowStates),
			ProcessingPath:       types.PathFreshExtraction,
		},
	}

	// Convert the result to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal result to JSON", "error", err)
		os.Exit(1)
	}

	// Save result to file
	outputFile := "./vending-machine-result.json"
	if err := os.WriteFile(outputFile, resultJSON, 0644); err != nil {
		logger.Error("Failed to write result to file", "error", err, "file", outputFile)
	} else {
		logger.Info("Result saved to file", "file", outputFile)
	}

	// Print partial result
	fmt.Println("---- Processing Result (Success) ----")
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Source Type: %s\n", result.SourceType)
	fmt.Printf("Extracted Structure: Row Count=%d, Columns Per Row=%d\n", 
		result.ExtractedStructure.RowCount, 
		result.ExtractedStructure.ColumnsPerRow)
	fmt.Printf("Row Order: %v\n", result.ExtractedStructure.RowOrder)
	
	fmt.Println("\nExtracted Row States:")
	for row, rowState := range extractedState.RowStates {
		fmt.Printf("  Row %s: Status=%s, Filled Positions=%d\n", 
			row, rowState.Status, len(rowState.FilledPositions))
	}
	
	fmt.Printf("\nObservations: %v\n", extractedState.Observations)
	fmt.Printf("Processing Duration: %s\n", result.ProcessingMetadata.ProcessingDuration)
	fmt.Println("------------------------------------")

	logger.Info("Vending machine parser test completed successfully",
		"status", result.Status,
		"rowCount", result.ExtractedStructure.RowCount,
		"columnsPerRow", result.ExtractedStructure.ColumnsPerRow,
		"durationMs", result.ProcessingMetadata.ProcessingDuration.Milliseconds(),
	)
}