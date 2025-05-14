package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/shared/schema"

	"workflow-function/PrepareSystemPrompt/internal"
)

func init() {
	// Initialize template loader with base path from environment or default
	templateBasePath := os.Getenv("TEMPLATE_BASE_PATH")
	if templateBasePath == "" {
		templateBasePath = "/opt/templates" // Default in container
	}
	
	// Initialize shared template loader
	err := internal.InitializeTemplateLoader(templateBasePath)
	if err != nil {
		log.Fatalf("Failed to initialize template loader: %v", err)
	}
	
	log.Printf("Initialized template loader with base path: %s", templateBasePath)
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (json.RawMessage, error) {
	start := time.Now()
	log.Printf("Received event: %s", string(event))
	
	// Parse and validate input
	var input internal.Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Printf("Error parsing input: %v", err)
		return nil, fmt.Errorf("invalid input format: %w", err)
	}
	
	// Validate input
	if err := internal.ValidateInput(&input); err != nil {
		log.Printf("Validation error: %v", err)
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	
	// Extract verification type
	verificationType := input.VerificationContext.VerificationType
	log.Printf("Processing verification type: %s", verificationType)
	
	// Get appropriate template
	tmpl, err := internal.GetTemplate(verificationType)
	if err != nil {
		log.Printf("Error getting template: %v", err)
		return nil, fmt.Errorf("template error: %w", err)
	}
	
	// Create template data context
	templateData, err := internal.BuildTemplateData(&input)
	if err != nil {
		log.Printf("Error building template data: %v", err)
		return nil, fmt.Errorf("context preparation failed: %w", err)
	}
	
	// Generate system prompt
	systemPrompt, err := internal.ProcessTemplate(tmpl, templateData)
	if err != nil {
		log.Printf("Error processing template: %v", err)
		return nil, fmt.Errorf("prompt generation failed: %w", err)
	}
	
	// Configure Bedrock
	bedrockConfig := internal.ConfigureBedrockSettings()
	
	// Update verification context status
	input.VerificationContext.Status = schema.StatusPromptPrepared
	
	// Create shared schema SystemPrompt
	sysPrompt := &schema.SystemPrompt{
		SystemPrompt: systemPrompt,
		BedrockConfig: bedrockConfig,
	}
	
	// Update workflow state
	if input.State != nil {
		input.State.SystemPrompt = sysPrompt
		input.State.VerificationContext.Status = schema.StatusPromptPrepared
	}
	
	// Create response
	response := internal.Response{
		VerificationContext: input.VerificationContext,
		SystemPrompt: sysPrompt,
		BedrockConfig: bedrockConfig,
	}
	
	// Include appropriate metadata based on verification type
	if verificationType == schema.VerificationTypeLayoutVsChecking && input.LayoutMetadata != nil {
		response.LayoutMetadata = input.LayoutMetadata
	} else if verificationType == schema.VerificationTypePreviousVsCurrent && input.HistoricalContext != nil {
		response.HistoricalContext = input.HistoricalContext
	}
	
	// Convert to JSON for return
	respJson, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	
	log.Printf("Completed in %v", time.Since(start))
	return respJson, nil
}

func main() {
	lambda.Start(HandleRequest)
}