package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"prepare-sys/internal"
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
	
	// Extract verification type
	verificationType := input.VerificationContext.VerificationType
	log.Printf("Processing verification type: %s", verificationType)
	
	// Get appropriate template
	tmpl, err := templateManager.GetTemplate(verificationType)
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
	
	// Generate system prompt
	systemPrompt, err := internal.ProcessTemplate(tmpl, templateData)
	if err != nil {
		log.Printf("Error processing template: %v", err)
		return internal.Response{}, fmt.Errorf("prompt generation failed: %w", err)
	}
	
	// Configure Bedrock
	bedrockConfig := internal.ConfigureBedrockSettings()
	
	// Update verification context status
	input.VerificationContext.Status = "SYSTEM_PROMPT_READY"
	
	// Prepare response
	response := internal.Response{
		VerificationContext: input.VerificationContext,
		SystemPrompt: internal.SystemPrompt{
			Content:       systemPrompt,
			PromptID:      fmt.Sprintf("prompt-%s-system", input.VerificationContext.VerificationID),
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
			PromptVersion: templateManager.GetLatestVersion(verificationType),
		},
		BedrockConfig: bedrockConfig,
	}
	
	// Include appropriate metadata based on verification type
	if verificationType == "LAYOUT_VS_CHECKING" && input.LayoutMetadata != nil {
		response.LayoutMetadata = input.LayoutMetadata
	} else if verificationType == "PREVIOUS_VS_CURRENT" && input.HistoricalContext != nil {
		response.HistoricalContext = input.HistoricalContext
	}
	
	log.Printf("Completed in %v", time.Since(start))
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}