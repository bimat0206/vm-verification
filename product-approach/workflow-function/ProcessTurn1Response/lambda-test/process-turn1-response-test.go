package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"workflow-function/ProcessTurn1Response/internal/handler"
)

func main() {
	// Set the environment variable for local testing
	os.Setenv("STATE_BUCKET", "test-verification-bucket")

	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "ProcessTurn1Response",
	)

	// Log startup
	logger.Info("ProcessTurn1Response Lambda function test starting",
		"timestamp", time.Now().Format(time.RFC3339),
	)

	// Create handler
	h, err := handler.New(logger)
	if err != nil {
		logger.Error("Failed to initialize handler", "error", err)
		os.Exit(1)
	}

	// Read input file
	inputFile := "./test-input.json"
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		logger.Error("Failed to read input file", "error", err, "file", inputFile)
		os.Exit(1)
	}

	// Parse input JSON
	var input interface{}
	if err := json.Unmarshal(inputData, &input); err != nil {
		logger.Error("Failed to parse input JSON", "error", err)
		os.Exit(1)
	}

	logger.Info("Processing input", "file", inputFile)

	// Mock S3 state bucket for testing
	// In a real environment, this would be a valid S3 bucket
	// For testing, we're using environment variable, but the S3 operations will fail
	// The code is structured to handle this by returning mock references

	// Process input with handler
	result, err := h.Handle(context.Background(), input)
	if err != nil {
		logger.Error("Handler processing failed", "error", err)
		os.Exit(1)
	}

	// Output result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("---- Processing Result ----")
	fmt.Println(string(resultJSON))
	fmt.Println("--------------------------")

	// Save result to file
	outputFile := "./process-result.json"
	if err := os.WriteFile(outputFile, resultJSON, 0644); err != nil {
		logger.Error("Failed to write result to file", "error", err, "file", outputFile)
	} else {
		logger.Info("Result saved to file", "file", outputFile)
	}

	// Log completion
	logger.Info("ProcessTurn1Response Lambda function test completed successfully")
}