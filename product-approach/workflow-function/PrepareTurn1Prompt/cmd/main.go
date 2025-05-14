package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"prepare-turn1/internal"
)

var (
	templateManager *internal.TemplateManager
)

func init() {
	// Initialize template manager with base path from environment or default
	templateBasePath := os.Getenv("TEMPLATE_BASE_PATH")
	if templateBasePath == "" {
		templateBasePath = "/opt/templates" // Default in container
	}
	
	templateManager = internal.NewTemplateManager(templateBasePath)
	log.Printf("Initialized template manager with base path: %s", templateBasePath)
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (internal.Response, error) {
	start := time.Now()
	log.Printf("Received event: %s", string(event))
	
	// Parse and validate input
	var input internal.Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Printf("Error parsing input: %v", err)
		return internal.Response{}, fmt.Errorf("invalid input format: %w", err)
	}
	
	// Validate input
	if err := internal.ValidateInput(&input); err != nil {
		log.Printf("Validation error: %v", err)
		return internal.Response{}, fmt.Errorf("input validation failed: %w", err)
	}
	
	// Ensure Turn Number is 1
	if input.TurnNumber != 1 {
		log.Printf("Invalid turn number: %d, expected 1", input.TurnNumber)
		return internal.Response{}, fmt.Errorf("this function only processes Turn 1 prompts")
	}
	
	// Extract verification type and ensure includeImage is reference
	verificationType := input.VerificationContext.VerificationType
	if input.IncludeImage != "reference" {
		log.Printf("Invalid includeImage value: %s, expected 'reference'", input.IncludeImage)
		return internal.Response{}, fmt.Errorf("Turn 1 must include reference image only")
	}
	
	log.Printf("Processing Turn 1 prompt for verification type: %s", verificationType)
	
	// Get appropriate template
	tmpl, err := templateManager.GetTurn1Template(verificationType)
	if err != nil {
		log.Printf("Error getting template: %v", err)
		return internal.Response{}, fmt.Errorf("template error: %w", err)
	}
	
	// Create template data context
	templateData, err := internal.BuildTemplateData(&input)
	if err != nil {
		log.Printf("Error building template data: %v", err)
		return internal.Response{}, fmt.Errorf("context preparation failed: %w", err)
	}
	
	// Generate turn 1 prompt text
	promptText, err := internal.ProcessTemplate(tmpl, templateData)
	if err != nil {
		log.Printf("Error processing template: %v", err)
		return internal.Response{}, fmt.Errorf("prompt generation failed: %w", err)
	}
	
	// Create the Bedrock message with image
	var bucketOwner string
	if input.Images != nil && input.Images.ReferenceImageMeta != nil {
		bucketOwner = input.Images.ReferenceImageMeta.BucketOwner
	}

	userMessage, err := internal.CreateTurn1Message(promptText, input.VerificationContext.ReferenceImageURL, bucketOwner)
	if err != nil {
		log.Printf("Error creating user message: %v", err)
		return internal.Response{}, fmt.Errorf("message creation failed: %w", err)
	}
	
	// Update verification context status
	input.VerificationContext.Status = "TURN1_PROMPT_READY"
	
	// Prepare response
	response := internal.Response{
		VerificationContext: input.VerificationContext,
		CurrentPrompt: internal.CurrentPrompt{
			Messages:      []internal.BedrockMessage{userMessage},
			TurnNumber:    1,
			PromptID:      fmt.Sprintf("prompt-%s-turn1", input.VerificationContext.VerificationID),
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
			PromptVersion: templateManager.GetLatestTurn1Version(verificationType),
			ImageIncluded: "reference",
		},
	}
	
	// Include appropriate metadata based on verification type
	if verificationType == "LAYOUT_VS_CHECKING" && input.LayoutMetadata != nil {
		response.LayoutMetadata = input.LayoutMetadata
	} else if verificationType == "PREVIOUS_VS_CURRENT" && input.HistoricalContext != nil {
		response.HistoricalContext = input.HistoricalContext
	}
	
	// Include Bedrock configuration if available
	if input.BedrockConfig != nil {
		response.BedrockConfig = *input.BedrockConfig // Dereference the pointer
	} else {
		// Default Bedrock configuration
		response.BedrockConfig = internal.ConfigureBedrockSettings()
	}
	
	log.Printf("Completed in %v", time.Since(start))
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}