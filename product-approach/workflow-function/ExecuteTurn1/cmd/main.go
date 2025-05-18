package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"

	"workflow-function/ExecuteTurn1/internal/config"
	"workflow-function/ExecuteTurn1/internal/dependencies"
	"workflow-function/ExecuteTurn1/internal/handler"
	"workflow-function/ExecuteTurn1/internal/models"  // Add this import
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

	// Create handler with model ID from config
	executeTurn1Handler = handler.NewHandler(clients.BedrockClient, clients.S3Client, clients.HybridConfig, log, cfg.BedrockModelID)
}

// LambdaHandler - main entrypoint for Lambda
func LambdaHandler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	requestID := uuid.NewString()
	ctx = context.WithValue(ctx, "requestID", requestID)
	log := log.WithCorrelationId(requestID)

	log.Info("Starting ExecuteTurn1 Lambda invocation", nil)

	// Parse and validate the input event
	request, err := models.NewRequestFromJSON(event, log)
	if err != nil {
		log.Error("Failed to parse input event", map[string]interface{}{"error": err.Error()})
		return nil, errors.NewValidationError("Invalid input format", map[string]interface{}{"error": err.Error()})
	}

	// Validate and sanitize the request
	if err := request.ValidateAndSanitize(log); err != nil {
		log.Error("Request validation failed", map[string]interface{}{"error": err.Error()})
		return models.NewErrorResponse(&request.WorkflowState, err.(*errors.WorkflowError), log, schema.StatusBedrockProcessingFailed), nil
	}

	log.Info("Request validation passed", map[string]interface{}{
		"verificationId": request.GetVerificationID(),
		"promptId":       request.GetPromptID(),
	})

	// Handle the request
	outputState, err := executeTurn1Handler.HandleRequest(ctx, &request.WorkflowState)
	if err != nil {
		log.Error("Handler error", map[string]interface{}{"error": err.Error()})
		if wfErr, ok := err.(*errors.WorkflowError); ok {
			return models.NewErrorResponse(outputState, wfErr, log, schema.StatusBedrockProcessingFailed), nil
		}
		// Wrap unexpected errors
		wfErr := errors.WrapError(err, errors.ErrorTypeInternal, "unexpected handler error", false)
		return models.NewErrorResponse(outputState, wfErr, log, schema.StatusBedrockProcessingFailed), nil
	}

	log.Info("Successfully processed Turn1", map[string]interface{}{
		"verificationId": outputState.VerificationContext.VerificationId,
		"status":         outputState.VerificationContext.Status,
	})

	return models.NewResponse(outputState, log), nil
}

func main() {
	lambda.Start(LambdaHandler)
}