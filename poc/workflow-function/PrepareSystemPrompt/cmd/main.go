package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/PrepareSystemPrompt/internal/config"
	"workflow-function/PrepareSystemPrompt/internal/handlers"
)

var handler *handlers.Handler

func init() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize handler
	handler, err = handlers.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	log.Printf("Successfully initialized PrepareSystemPrompt Lambda")
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (json.RawMessage, error) {
	start := time.Now()
	log.Printf("Received event: %s", string(event))

	// Handle request using the handler
	response, err := handler.HandleRequest(ctx, event)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		return nil, err
	}

	log.Printf("Successfully processed request in %v", time.Since(start))
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}