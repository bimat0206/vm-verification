package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
	"workflow-function/ProcessTurn1Response/internal/parser"
	"workflow-function/ProcessTurn1Response/internal/processor"
	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/ProcessTurn1Response/internal/validator"
)

// This test specifically tests the parsing functionality for vending machine data
func main() {
	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "ProcessTurn1Response-ImproveMock",
	)

	// Log startup
	logger.Info("ProcessTurn1Response Lambda improved mock test starting",
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

	// Create parser with direct pattern overrides for vending machine data
	p := parser.New(logger)
	patternProvider := parser.NewPatternProvider(logger)
	
	// Register custom patterns for vending machine structure
	patternProvider.RegisterCustomPattern("row_column", `(?i)(?:VM-)?(\d+)[^.]*?(\d+)`)
	patternProvider.RegisterCustomPattern("row_status", `(?i)^## Row ([A-F])(?:[^*]+)\*\*Status: ([A-Za-z]+)\*\*`)
	
	// Create our FreshExtractionProcessor directly
	logAdapter := processor.NewSlogLoggerAdapter(logger)
	validatorObj := validator.NewValidator(logAdapter)
	freshProcessor := processor.NewFreshExtractionProcessor(logger, p, validatorObj)
	
	// Initialize result structure
	result := &types.Turn1ProcessingResult{
		Status:     "PROCESSING",
		SourceType: types.PathFreshExtraction,
		ProcessingMetadata: &types.ProcessingMetadata{
			ProcessingStartTime: time.Now(),
			ProcessingPath:      types.PathFreshExtraction,
			ResponseSize:        int64(len(responseContent)),
		},
	}
	
	// Process using the fresh extraction processor directly
	err = freshProcessor.Process(context.Background(), responseContent, nil, result)
	if err != nil {
		logger.Error("Processing failed", "error", err)
		os.Exit(1)
	}
	
	// Update metadata
	result.ProcessingMetadata.ProcessingEndTime = time.Now()
	result.ProcessingMetadata.ProcessingDuration = result.ProcessingMetadata.ProcessingEndTime.Sub(result.ProcessingMetadata.ProcessingStartTime)
	result.Status = "EXTRACTION_COMPLETE"

	// Convert the result to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal result to JSON", "error", err)
		os.Exit(1)
	}

	// Save result to file
	outputFile := "./improved-result.json"
	if err := os.WriteFile(outputFile, resultJSON, 0644); err != nil {
		logger.Error("Failed to write result to file", "error", err, "file", outputFile)
	} else {
		logger.Info("Result saved to file", "file", outputFile)
	}

	// Print partial result
	fmt.Println("---- Processing Result (Success) ----")
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Source Type: %s\n", result.SourceType)
	
	if result.ExtractedStructure != nil {
		fmt.Printf("Extracted Structure: Row Count=%d, Columns Per Row=%d\n", 
			result.ExtractedStructure.RowCount, 
			result.ExtractedStructure.ColumnsPerRow)
		fmt.Printf("Row Order: %v\n", result.ExtractedStructure.RowOrder)
	} else {
		fmt.Println("No structure extracted")
	}
	
	// Check for extracted state
	if state, ok := result.ReferenceAnalysis["extractedState"].(*types.ExtractedState); ok && state != nil {
		fmt.Println("Extracted Row States:")
		for row, rowState := range state.RowStates {
			fmt.Printf("  Row %s: Status=%s, Quantity=%d\n", 
				row, rowState.Status, rowState.Quantity)
		}
	}
	
	fmt.Printf("Processing Duration: %s\n", result.ProcessingMetadata.ProcessingDuration)
	fmt.Println("------------------------------------")

	logger.Info("Improved mock test completed successfully",
		"status", result.Status,
		"durationMs", result.ProcessingMetadata.ProcessingDuration.Milliseconds(),
	)
}