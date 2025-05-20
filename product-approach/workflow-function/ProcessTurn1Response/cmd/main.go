package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"workflow-function/ProcessTurn1Response/internal/handler"
	"workflow-function/shared/schema"
)

func main() {
	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "ProcessTurn1Response",
	)
	
	// Log startup
	logger.Info("ProcessTurn1Response Lambda function starting",
		"schemaVersion", schema.SchemaVersion,
		"goVersion", os.Getenv("GO_VERSION"),
	)

	// Global dependencies have been moved to state-based dependency management

	// Create handler with dependencies
	h, err := handler.New(logger)
	if err != nil {
		logger.Error("Failed to initialize handler", "error", err)
		os.Exit(1)
	}

	// Start Lambda runtime
	lambda.Start(func(ctx context.Context, input interface{}) (interface{}, error) {
		return h.Handle(ctx, input)
	})
}
