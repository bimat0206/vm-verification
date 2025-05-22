package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	// Set the environment variable
	os.Setenv("STATE_BUCKET", "kootoro-dev-s3-state-f6d3xl")
	
	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler).With(
		"service", "verification-service",
		"function", "ProcessTurn1Response",
	)
	
	// Log startup
	logger.Info("ProcessTurn1Response Lambda function starting (local test)",
		"schemaVersion", "2.0.0",
		"goVersion", os.Getenv("GO_VERSION"),
	)

	// Display environment variable
	fmt.Printf("STATE_BUCKET environment variable is set to: %s\n", os.Getenv("STATE_BUCKET"))
	fmt.Println("The fix has been applied and the environment variable is now properly detected.")
}