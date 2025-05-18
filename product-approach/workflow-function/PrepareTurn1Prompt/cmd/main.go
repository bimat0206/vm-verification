package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/templateloader"

	"prepare-turn1/internal"
)

var (
	log            logger.Logger
	templateLoader templateloader.TemplateLoader
)

func init() {
	// Initialize logger
	log = logger.New("vending-machine-verification", "PrepareTurn1Prompt")
	
	// Initialize template loader with base path from environment or default
	templateBasePath := os.Getenv("TEMPLATE_BASE_PATH")
	if templateBasePath == "" {
		templateBasePath = "/opt/templates" // Default in container
	}
	
	// Create template loader config
	config := templateloader.Config{
		BasePath:     templateBasePath,
		CacheEnabled: true,
		CustomFuncs:  templateloader.DefaultFunctions,
	}
	
	// Initialize template loader
	var err error
	templateLoader, err = templateloader.New(config)
	if err != nil {
		log.Error("Failed to initialize template loader", map[string]interface{}{
			"error":    err.Error(),
			"basePath": templateBasePath,
		})
		return
	}
	
	log.Info("Initialized template loader", map[string]interface{}{
		"basePath": templateBasePath,
	})
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (*internal.Response, error) {
	start := time.Now()
	log.LogReceivedEvent(event)
	
	// Add panic recovery middleware
	var verificationId string
	defer func() {
		internal.RecoverFromPanic(log, verificationId)
	}()
	
	// Parse and validate input
	var input internal.Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Error("Error parsing input", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.NewParsingError("JSON", err)
	}
	
	// Set verification ID for panic recovery
	if input.VerificationContext != nil {
		verificationId = input.VerificationContext.VerificationID
	}
	
	// Validate input
	if err := validateInput(&input); err != nil {
		log.Error("Validation error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	
	// Ensure Turn Number is 1
	if input.TurnNumber != 1 {
		log.Error("Invalid turn number", map[string]interface{}{
			"turnNumber": input.TurnNumber,
			"expected":   1,
		})
		return nil, errors.NewValidationError("This function only processes Turn 1 prompts", 
			map[string]interface{}{"turnNumber": input.TurnNumber})
	}
	
	// Extract verification type and ensure includeImage is reference
	verificationType := input.VerificationContext.VerificationType
	if input.IncludeImage != "reference" {
		log.Error("Invalid includeImage value", map[string]interface{}{
			"includeImage": input.IncludeImage,
			"expected":     "reference",
		})
		return nil, errors.NewValidationError("Turn 1 must include reference image only", 
			map[string]interface{}{"includeImage": input.IncludeImage})
	}
	
	log.Info("Processing Turn 1 prompt", map[string]interface{}{
		"verificationType": verificationType,
		"verificationId":   input.VerificationContext.VerificationID,
	})
	
	// Convert input to workflow state
	workflowState := internal.ConvertToWorkflowState(&input)
	
	// Get appropriate template
	// Convert LAYOUT_VS_CHECKING to layout-vs-checking (replace underscores with hyphens)
	formattedType := strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
	tmplName := fmt.Sprintf("turn1-%s", formattedType)
	tmpl, err := templateLoader.LoadTemplate(tmplName)
	if err != nil {
		log.Error("Error getting template", map[string]interface{}{
			"error":       err.Error(),
			"templateKey": tmplName,
		})
		return nil, errors.NewInternalError("template-loader", err)
	}
	
	// Create template data context
	templateData, err := internal.BuildTemplateData(workflowState)
	if err != nil {
		log.Error("Error building template data", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.NewInternalError("template-data", err)
	}
	
	// Generate turn 1 prompt text
	promptText, err := internal.ProcessTemplate(tmpl, templateData)
	if err != nil {
		log.Error("Error processing template", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.NewInternalError("template-processing", err)
	}
	
	// Create message using schema.BuildBedrockMessage
	userMessage := schema.BuildBedrockMessage(promptText, workflowState.Images.GetReference())
	
	// Update verification context status
	input.VerificationContext.Status = schema.StatusTurn1PromptReady
	workflowState.VerificationContext.Status = schema.StatusTurn1PromptReady
	
	// Update the current prompt in the workflow state
	workflowState.CurrentPrompt.Messages = []schema.BedrockMessage{userMessage}
	workflowState.CurrentPrompt.TurnNumber = 1
	workflowState.CurrentPrompt.PromptId = fmt.Sprintf("prompt-%s-turn1", input.VerificationContext.VerificationID)
	workflowState.CurrentPrompt.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	workflowState.CurrentPrompt.PromptVersion = templateLoader.GetLatestVersion(tmplName)
	workflowState.CurrentPrompt.IncludeImage = "reference"
	
	// Include Bedrock configuration if available
	if input.BedrockConfig != nil {
		// Use the input BedrockConfig
		workflowState.BedrockConfig = &schema.BedrockConfig{
			AnthropicVersion: input.BedrockConfig.AnthropicVersion,
			MaxTokens:        input.BedrockConfig.MaxTokens,
			Thinking: &schema.Thinking{
				Type:         input.BedrockConfig.Thinking.Type,
				BudgetTokens: input.BedrockConfig.Thinking.BudgetTokens,
			},
		}
	} else {
		// Use default Bedrock configuration
		workflowState.BedrockConfig = &schema.BedrockConfig{
			AnthropicVersion: "bedrock-2023-05-31",
			MaxTokens:        24000,
			Thinking: &schema.Thinking{
				Type:         "enabled",
				BudgetTokens: 16000,
			},
		}
	}
	
	// Convert workflow state back to response
	response := internal.ConvertToResponse(workflowState)
	
	log.Info("Completed processing", map[string]interface{}{
		"duration": time.Since(start).String(),
	})
	
	log.LogOutputEvent(response)
	return response, nil
}

// validateInput validates the input
func validateInput(input *internal.Input) error {
	if input == nil {
		return errors.NewValidationError("Input cannot be nil", nil)
	}
	
	if input.VerificationContext == nil {
		return errors.NewMissingFieldError("verificationContext")
	}
	
	if input.VerificationContext.VerificationID == "" {
		return errors.NewMissingFieldError("verificationContext.verificationId")
	}
	
	if input.VerificationContext.VerificationType == "" {
		return errors.NewMissingFieldError("verificationContext.verificationType")
	}
	
	if input.VerificationContext.ReferenceImageURL == "" {
		return errors.NewMissingFieldError("verificationContext.referenceImageUrl")
	}
	
	if input.TurnNumber != 1 {
		return errors.NewInvalidFieldError("turnNumber", input.TurnNumber, "1")
	}
	
	if input.IncludeImage != "reference" {
		return errors.NewInvalidFieldError("includeImage", input.IncludeImage, "reference")
	}
	
	// Validate verification type
	validTypes := []string{schema.VerificationTypeLayoutVsChecking, schema.VerificationTypePreviousVsCurrent}
	isValidType := false
	for _, vt := range validTypes {
		if input.VerificationContext.VerificationType == vt {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return errors.NewInvalidFieldError("verificationType", 
			input.VerificationContext.VerificationType, 
			fmt.Sprintf("one of %v", validTypes))
	}
	
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
