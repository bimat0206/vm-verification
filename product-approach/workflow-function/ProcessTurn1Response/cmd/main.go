package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

func main() {
	// Initialize logger
	log := logger.New("verification-service", "ProcessTurn1Response")
	
	// Log startup
	log.Info("ProcessTurn1Response Lambda function starting", map[string]interface{}{
		"schemaVersion": schema.SchemaVersion,
		"goVersion":     os.Getenv("GO_VERSION"),
	})

	// Create handler with dependencies
	handler := NewHandler(log)

	// Start Lambda runtime
	lambda.Start(func(ctx context.Context, input schema.WorkflowState) (schema.WorkflowState, error) {
		return handler.Handle(ctx, input)
	})
}