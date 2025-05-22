package main

import (
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

// This test bypasses the S3 integration and directly tests the parsing functionality
func main() {
	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "ProcessTurn1Response-Mock",
	)

	// Log startup
	logger.Info("ProcessTurn1Response Lambda function mock test starting",
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

	// Create parser
	p := parser.New(logger)

	// Create processor adapter
	logAdapter := processor.NewSlogLoggerAdapter(logger)
	validatorObj := validator.NewValidator(logAdapter)
	proc := processor.NewWithOptions(logger, 
		processor.WithParser(p),
		processor.WithValidator(validatorObj),
	)

	// Process the response - use PathFreshExtraction since it's PREVIOUS_VS_CURRENT without historical data
	result, err := proc.ProcessTurn1Response(nil, responseContent, types.PathFreshExtraction, nil)
	if err != nil {
		logger.Error("Processing failed", "error", err)
		os.Exit(1)
	}

	// Convert the result to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal result to JSON", "error", err)
		os.Exit(1)
	}

	// Save result to file
	outputFile := "./real-process-result.json"
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
	
	fmt.Printf("Number of Reference Analysis Entries: %d\n", len(result.ReferenceAnalysis))
	fmt.Printf("Processing Duration: %s\n", result.ProcessingMetadata.ProcessingDuration)
	fmt.Println("------------------------------------")

	logger.Info("ProcessTurn1Response Lambda function mock test completed successfully",
		"status", result.Status,
		"durationMs", result.ProcessingMetadata.ProcessingDuration.Milliseconds(),
	)
}