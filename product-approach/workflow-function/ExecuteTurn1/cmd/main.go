package main

import (
	"context"
	"encoding/json"
	//"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"

	"workflow-function/ExecuteTurn1/internal/config"
	"workflow-function/ExecuteTurn1/internal/dependencies"
	"workflow-function/ExecuteTurn1/internal/handler"
)

// Global handler for re-use between Lambda invocations
var executeTurn1Handler *handler.Handler
var log logger.Logger

func init() {
	// Initialize logger
	log = logger.New("vending-verification", "ExecuteTurn1")

	// Load config
	cfg, err := config.New(log)
	if err != nil {
		log.Error("Failed to load config", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Set up AWS/Bedrock/S3 clients
	clients, err := dependencies.New(context.Background(), cfg, log)
	if err != nil {
		log.Error("Failed to initialize AWS clients", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Create handler
	executeTurn1Handler = handler.NewHandler(clients.BedrockClient, clients.S3Client, clients.HybridConfig, log)
}

// LambdaHandler - main entrypoint for Lambda
func LambdaHandler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	requestID := uuid.NewString()
	ctx = context.WithValue(ctx, "requestID", requestID)
	log := log.WithCorrelationId(requestID)

	log.Info("Starting ExecuteTurn1 Lambda invocation", nil)

	// Parse the input event into schema.WorkflowState
	var inputState schema.WorkflowState
	if err := json.Unmarshal(event, &inputState); err != nil {
		log.Error("Failed to parse input event", map[string]interface{}{"error": err.Error()})
		return nil, errors.NewValidationError("Invalid input format", map[string]interface{}{"error": err.Error()})
	}
	
	// Ensure schema version is set to the latest supported version
	if inputState.SchemaVersion == "" || inputState.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{
			"from": inputState.SchemaVersion, 
			"to": schema.SchemaVersion,
		})
		inputState.SchemaVersion = schema.SchemaVersion
	}

	// Validate workflow state
	if valErrs := schema.ValidateWorkflowState(&inputState); len(valErrs) > 0 {
		log.Error("WorkflowState validation failed", map[string]interface{}{"validationErrors": valErrs.Error()})
		return map[string]interface{}{
			"workflowState": inputState,
			"error": errors.NewValidationError("WorkflowState validation failed", map[string]interface{}{"validationErrors": valErrs.Error()}),
		}, nil
	}

	// Handle the request
	outputState, err := executeTurn1Handler.HandleRequest(ctx, &inputState)
	if err != nil {
		log.Error("Handler error", map[string]interface{}{"error": err.Error()})
		return map[string]interface{}{
			"workflowState": outputState,
			"error":         err,
		}, nil
	}

	log.Info("Successfully processed Turn1", map[string]interface{}{
		"verificationId": outputState.VerificationContext.VerificationId,
		"status":         outputState.VerificationContext.Status,
	})
	return map[string]interface{}{"workflowState": outputState}, nil
}

func main() {
	lambda.Start(LambdaHandler)
}
