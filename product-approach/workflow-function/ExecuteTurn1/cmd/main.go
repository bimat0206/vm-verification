package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"workflow-function/ExecuteTurn1/internal/config"
	"workflow-function/ExecuteTurn1/internal/dependencies"
	"workflow-function/ExecuteTurn1/internal/handler"
	"workflow-function/ExecuteTurn1/internal/models"
)

// Initialize handler outside of the handler function to reuse across invocations
var executeTurn1Handler *handler.Handler

func init() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize clients
	clients, err := dependencies.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	
	// Create handler
	executeTurn1Handler = handler.NewHandler(clients)
}

// LambdaHandler is the Lambda handler function
func LambdaHandler(ctx context.Context, request json.RawMessage) (interface{}, error) {
	// Generate a request ID for correlation
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, "requestID", requestID)
	
	log.Printf("Starting ExecuteTurn1 function, RequestID: %s", requestID)
	
	// Parse the input
	var executeTurn1Request models.ExecuteTurn1Request
	if err := json.Unmarshal(request, &executeTurn1Request); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return nil, fmt.Errorf("invalid request format: %w", err)
	}
	
	// Log key information
	log.Printf("Processing verification: %s, Stage: %s, Status: %s", 
		executeTurn1Request.WorkflowState.VerificationID,
		executeTurn1Request.WorkflowState.Stage,
		executeTurn1Request.WorkflowState.Status)
	
	// Handle the request
	response, err := executeTurn1Handler.HandleRequest(ctx, executeTurn1Request)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		return nil, err
	}
	
	// Check for error in response
	if response.Error != nil {
		log.Printf("Function returned error: %s: %s (Retryable: %t)", 
			response.Error.Code, 
			response.Error.Message, 
			response.Error.Retryable)
		
		// Lambda Step Functions integration expects errors to be returned
		// as JSON objects with certain fields
		return response, nil
	}
	
	// Log success
	log.Printf("Successfully processed Turn1 for verification: %s", 
		response.WorkflowState.VerificationID)
	
	return response, nil
}

func main() {
	lambda.Start(LambdaHandler)
}