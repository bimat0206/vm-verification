package main

import (
   "context"
   "encoding/json"
   "fmt"
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
   
   // Initialize template loader with base path from environment
   templateBasePath := internal.GetEnvWithDefault("TEMPLATE_BASE_PATH", "")
   if templateBasePath == "" {
   	log.Error("TEMPLATE_BASE_PATH environment variable is required", nil)
   	panic("TEMPLATE_BASE_PATH environment variable not set")
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
   	panic(fmt.Sprintf("Failed to initialize template loader: %v", err))
   }
   
   log.Info("Initialized PrepareTurn1Prompt function", map[string]interface{}{
   	"templateBasePath": templateBasePath,
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
   
   // Set verification ID for panic recovery and logging
   if input.VerificationContext != nil {
   	verificationId = input.VerificationContext.VerificationID
   }
   
   // Validate input
   if err := validateInput(&input); err != nil {
   	log.Error("Validation error", map[string]interface{}{
   		"error":          err.Error(),
   		"verificationId": verificationId,
   	})
   	return nil, err
   }
   
   // Log processing start
   log.Info("Processing Turn 1 prompt", map[string]interface{}{
   	"verificationType": input.VerificationContext.VerificationType,
   	"verificationId":   verificationId,
   	"turnNumber":       input.TurnNumber,
   	"includeImage":     input.IncludeImage,
   })
   
   // Convert input to workflow state
   workflowState := internal.ConvertToWorkflowState(&input)
   
   // Process images for Bedrock before template processing
   if err := processImagesForBedrock(workflowState); err != nil {
   	log.Error("Failed to process images", map[string]interface{}{
   		"error":          err.Error(),
   		"verificationId": verificationId,
   	})
   	return nil, errors.NewInternalError("image-processing", err)
   }
   
   // Get appropriate template
   templateName := buildTemplateName(input.VerificationContext.VerificationType)
   tmpl, err := templateLoader.LoadTemplate(templateName)
   if err != nil {
   	log.Error("Error loading template", map[string]interface{}{
   		"error":          err.Error(),
   		"templateName":   templateName,
   		"verificationId": verificationId,
   	})
   	return nil, errors.NewInternalError("template-loader", err)
   }
   
   // Create template data context
   templateData, err := internal.BuildTemplateData(workflowState)
   if err != nil {
   	log.Error("Error building template data", map[string]interface{}{
   		"error":          err.Error(),
   		"verificationId": verificationId,
   	})
   	return nil, errors.NewInternalError("template-data", err)
   }
   
   // Generate turn 1 prompt text
   promptText, err := internal.ProcessTemplate(tmpl, templateData)
   if err != nil {
   	log.Error("Error processing template", map[string]interface{}{
   		"error":          err.Error(),
   		"verificationId": verificationId,
   	})
   	return nil, errors.NewInternalError("template-processing", err)
   }
   
   // Create Bedrock message with prompt text and reference image
   userMessage, err := createBedrockMessage(promptText, workflowState)
   if err != nil {
   	log.Error("Error creating Bedrock message", map[string]interface{}{
   		"error":          err.Error(),
   		"verificationId": verificationId,
   	})
   	return nil, errors.NewInternalError("bedrock-message", err)
   }
   
   // Update verification context status
   workflowState.VerificationContext.Status = schema.StatusTurn1PromptReady
   
   // Update the current prompt in the workflow state
   updateCurrentPrompt(workflowState, userMessage, templateName)
   
   // Set Bedrock configuration
   setBedrockConfiguration(workflowState, &input)
   
   // Convert workflow state back to response
   response := internal.ConvertToResponse(workflowState)
   
   // Log completion
   duration := time.Since(start)
   log.Info("Completed processing", map[string]interface{}{
   	"duration":       duration.String(),
   	"verificationId": verificationId,
   	"templateUsed":   templateName,
   })
   
   log.LogOutputEvent(response)
   return response, nil
}

// validateInput validates the input parameters
func validateInput(input *internal.Input) error {
   // Basic validation
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
   
   // Turn number validation
   if input.TurnNumber != 1 {
   	return errors.NewInvalidFieldError("turnNumber", input.TurnNumber, "1")
   }
   
   // Include image validation
   if input.IncludeImage != "reference" {
   	return errors.NewInvalidFieldError("includeImage", input.IncludeImage, "reference")
   }
   
   // Verification type validation
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
   
   // Validation specific to verification type
   if input.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
   	return validateLayoutVsChecking(input)
   } else if input.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
   	return validatePreviousVsCurrent(input)
   }
   
   return nil
}

// validateLayoutVsChecking validates layout vs checking specific fields
func validateLayoutVsChecking(input *internal.Input) error {
   ctx := input.VerificationContext
   
   // Layout ID is required
   if ctx.LayoutID <= 0 {
   	return errors.NewValidationError("Layout ID is required for LAYOUT_VS_CHECKING", 
   		map[string]interface{}{"layoutId": ctx.LayoutID})
   }
   
   // Layout prefix is required
   if ctx.LayoutPrefix == "" {
   	return errors.NewMissingFieldError("verificationContext.layoutPrefix")
   }
   
   // Layout metadata should be available (optional check since it might come from upstream)
   if input.LayoutMetadata == nil {
   	log.Warn("Layout metadata not provided", map[string]interface{}{
   		"verificationId": ctx.VerificationID,
   	})
   }
   
   // Check S3 URL is in reference bucket
   if err := validateS3BucketForType(ctx.ReferenceImageURL, "REFERENCE_BUCKET"); err != nil {
   	return err
   }
   
   return nil
}

// validatePreviousVsCurrent validates previous vs current specific fields
func validatePreviousVsCurrent(input *internal.Input) error {
   ctx := input.VerificationContext
   
   // Check S3 URL is in checking bucket (for previous vs current, reference image is in checking bucket)
   if err := validateS3BucketForType(ctx.ReferenceImageURL, "CHECKING_BUCKET"); err != nil {
   	return err
   }
   
   // Historical context is optional for Turn 1
   if input.HistoricalContext != nil {
   	log.Info("Historical context provided", map[string]interface{}{
   		"verificationId":         ctx.VerificationID,
   		"previousVerificationId": input.HistoricalContext.PreviousVerificationID,
   	})
   }
   
   return nil
}

// validateS3BucketForType validates that the S3 URL uses the correct bucket
func validateS3BucketForType(s3URL, bucketEnvVar string) error {
   expectedBucket := internal.GetEnvWithDefault(bucketEnvVar, "")
   if expectedBucket == "" {
   	return errors.NewValidationError(fmt.Sprintf("%s environment variable not set", bucketEnvVar), nil)
   }
   
   bucket, _, err := internal.ExtractS3BucketAndKey(s3URL)
   if err != nil {
   	return errors.NewValidationError("Invalid S3 URL format", 
   		map[string]interface{}{"url": s3URL, "error": err.Error()})
   }
   
   if bucket != expectedBucket {
   	return errors.NewValidationError("S3 URL uses incorrect bucket", 
   		map[string]interface{}{
   			"expectedBucket": expectedBucket,
   			"actualBucket":   bucket,
   			"url":            s3URL,
   		})
   }
   
   return nil
}

// processImagesForBedrock processes images to ensure they have Base64 data
func processImagesForBedrock(workflowState *schema.WorkflowState) error {
   if workflowState.Images == nil {
   	return nil
   }
   
   // Process reference image (required for Turn 1)
   if refImage := workflowState.Images.GetReference(); refImage != nil {
   	if err := internal.ProcessImageForBedrock(refImage); err != nil {
   		return fmt.Errorf("failed to process reference image: %w", err)
   	}
   	log.Info("Reference image processed successfully", map[string]interface{}{
   		"storageMethod": refImage.StorageMethod,
   		"format":        refImage.Format,
   		"hasBase64":     refImage.Base64Data != "",
   	})
   }
   
   // Process checking image if available (not needed for Turn 1 but may be present)
   if checkImage := workflowState.Images.GetChecking(); checkImage != nil {
   	if err := internal.ProcessImageForBedrock(checkImage); err != nil {
   		// Log warning but don't fail for checking image in Turn 1
   		log.Warn("Failed to process checking image", map[string]interface{}{
   			"error": err.Error(),
   		})
   	}
   }
   
   return nil
}

// buildTemplateName constructs the template name based on verification type
func buildTemplateName(verificationType string) string {
   // Convert LAYOUT_VS_CHECKING to layout-vs-checking (replace underscores with hyphens)
   formattedType := strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
   return fmt.Sprintf("turn1-%s", formattedType)
}

// createBedrockMessage creates a Bedrock message with prompt text and reference image
func createBedrockMessage(promptText string, workflowState *schema.WorkflowState) (schema.BedrockMessage, error) {
   // Get reference image
   refImage := workflowState.Images.GetReference()
   if refImage == nil {
   	return schema.BedrockMessage{}, fmt.Errorf("reference image is required for Turn 1")
   }
   
   // Ensure image has Base64 data
   if refImage.Base64Data == "" {
   	return schema.BedrockMessage{}, fmt.Errorf("reference image missing Base64 data")
   }
   
   // Use shared schema function to build the message
   return schema.BuildBedrockMessage(promptText, refImage), nil
}

// updateCurrentPrompt updates the current prompt in the workflow state
func updateCurrentPrompt(workflowState *schema.WorkflowState, userMessage schema.BedrockMessage, templateName string) {
   workflowState.CurrentPrompt.Messages = []schema.BedrockMessage{userMessage}
   workflowState.CurrentPrompt.TurnNumber = 1
   workflowState.CurrentPrompt.PromptId = internal.GeneratePromptID(workflowState.VerificationContext.VerificationId, 1)
   workflowState.CurrentPrompt.CreatedAt = internal.FormatTimestamp(time.Now())
   workflowState.CurrentPrompt.PromptVersion = templateLoader.GetLatestVersion(templateName)
   workflowState.CurrentPrompt.IncludeImage = "reference"
}

// setBedrockConfiguration sets the Bedrock configuration in the workflow state
func setBedrockConfiguration(workflowState *schema.WorkflowState, input *internal.Input) {
   if input.BedrockConfig != nil {
   	// Use the input BedrockConfig if provided
   	workflowState.BedrockConfig = &schema.BedrockConfig{
   		AnthropicVersion: input.BedrockConfig.AnthropicVersion,
   		MaxTokens:        input.BedrockConfig.MaxTokens,
   	}
   	
   	// ThinkingConfig is not a pointer in our internal types
	if input.BedrockConfig.Thinking.Type != "" {
   		workflowState.BedrockConfig.Thinking = &schema.Thinking{
   			Type:         input.BedrockConfig.Thinking.Type,
   			BudgetTokens: input.BedrockConfig.Thinking.BudgetTokens,
   		}
   	}
   } else {
   	// Use environment variables for default configuration
   	workflowState.BedrockConfig = &schema.BedrockConfig{
   		AnthropicVersion: internal.GetEnvWithDefault("ANTHROPIC_VERSION", ""),
   		MaxTokens:        internal.GetIntEnvWithDefault("MAX_TOKENS", 0),
   		Thinking: &schema.Thinking{
   			Type:         internal.GetEnvWithDefault("THINKING_TYPE", ""),
   			BudgetTokens: internal.GetIntEnvWithDefault("BUDGET_TOKENS", 0),
   		},
   	}
   	
   	// Validate that required configuration is provided
   	if workflowState.BedrockConfig.AnthropicVersion == "" {
   		log.Warn("ANTHROPIC_VERSION environment variable not set", map[string]interface{}{
   			"verificationId": workflowState.VerificationContext.VerificationId,
   		})
   	}
   	
   	if workflowState.BedrockConfig.MaxTokens == 0 {
   		log.Warn("MAX_TOKENS environment variable not set", map[string]interface{}{
   			"verificationId": workflowState.VerificationContext.VerificationId,
   		})
   	}
   }
   
   log.Info("Bedrock configuration set", map[string]interface{}{
   	"anthropicVersion": workflowState.BedrockConfig.AnthropicVersion,
   	"maxTokens":        workflowState.BedrockConfig.MaxTokens,
   	"thinkingEnabled":  workflowState.BedrockConfig.Thinking.Type != "",
   })
}

// main is the entry point for the Lambda function
func main() {
   lambda.Start(HandleRequest)
}