package main

import (
	"context"
	"log"
	
	"github.com/aws/aws-lambda-go/lambda"
	
	"workflow-function/FetchImages/internal/handler"
)

func main() {
	// Create a handler to process Lambda events
	h, err := handler.NewHandler(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}
	
	// Start the Lambda handler
	lambda.Start(h.Handle)
}