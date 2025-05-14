package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"shared/promptutils"
)

var (
	promptProcessor *promptutils.PromptProcessor
)

func init() {
	// Set environment variables for templates
	os.Setenv("COMPONENT_NAME", "ExampleFunction")
	
	// Initialize prompt processor with template path
	templateBasePath := "./templates" // For local testing
	promptProcessor = promptutils.NewPromptProcessor(templateBasePath)
	
	log.Printf("Initialized prompt processor with base path: %s", templateBasePath)
	log.Printf("Available templates: %v", promptProcessor.ListAvailableTemplates())
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (promptutils.Response, error) {
	start := time.Now()
	log.Printf("Received event: %s", string(event))
	
	// Process the input using the shared promptutils package
	response, err := promptProcessor.ProcessInput(event)
	if err != nil {
		log.Printf("Error processing input: %v", err)
		return promptutils.Response{}, err
	}
	
	log.Printf("Completed in %v", time.Since(start))
	return response, nil
}

func main() {
	// Check if we're running in Lambda or locally
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda, use Lambda handler
		lambda.Start(HandleRequest)
	} else {
		// Running locally, read from file
		filePath := "test-input.json"
		if len(os.Args) > 1 {
			filePath = os.Args[1]
		}
		
		// Read input file
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read input file %s: %v", filePath, err)
		}
		
		// Process input
		response, err := promptProcessor.ProcessInput(data)
		if err != nil {
			log.Fatalf("Error processing input: %v", err)
		}
		
		// Output result
		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(responseJSON))
	}
}