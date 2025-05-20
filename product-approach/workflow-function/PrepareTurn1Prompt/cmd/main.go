package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/templateloader"

	"prepare-turn1/internal/core"
	"prepare-turn1/internal/images"
	"prepare-turn1/internal/integration"
	"prepare-turn1/internal/state"
	"prepare-turn1/internal/validation"
)

var (
	log            logger.Logger
	templateLoader templateloader.TemplateLoader
	s3StateManager s3state.Manager
)

func init() {
	// Initialize logger
	log = logger.New("vending-machine-verification", "PrepareTurn1Prompt")
	
	// Initialize template loader with base path from environment
	templateBasePath := integration.GetEnvWithDefault("TEMPLATE_BASE_PATH", "/opt/templates")
	
	// Create template loader config - don't set CustomFuncs explicitly as they're set in the templateloader by default
	config := templateloader.Config{
		BasePath:     templateBasePath,
		CacheEnabled: true,
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
	
	// Log discovered templates
	log.Info("Template loader initialized", map[string]interface{}{
		"basePath": templateBasePath,
		"versions": templateLoader.(*templateloader.Loader).ListVersions("turn1-layout-vs-checking"),
	})

	// Initialize S3 state manager
	s3Bucket := integration.GetEnvWithDefault("STATE_BUCKET", "")
	if s3Bucket == "" {
		log.Error("STATE_BUCKET environment variable is required", nil)
		panic("STATE_BUCKET environment variable not set")
	}

	s3StateManager, err = s3state.New(s3Bucket)
	if err != nil {
		log.Error("Failed to initialize S3 state manager", map[string]interface{}{
			"error":  err.Error(),
			"bucket": s3Bucket,
		})
		panic(fmt.Sprintf("Failed to initialize S3 state manager: %v", err))
	}
	
	log.Info("Initialized PrepareTurn1Prompt function", map[string]interface{}{
		"templateBasePath": templateBasePath,
		"s3Bucket":         s3Bucket,
	})
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (*state.Output, error) {
	start := time.Now()
	log.LogReceivedEvent(event)
	
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Error("Panic recovered", map[string]interface{}{
				"panic": fmt.Sprintf("%v", r),
			})
		}
	}()
	
	// First, try to parse as S3 state envelope
	var envelope s3state.Envelope
	
	// Also try to parse as a JSON object to check for schemaVersion field
	var rawEvent map[string]interface{}
	_ = json.Unmarshal(event, &rawEvent)
	
	hasSchemaVersion := false
	schemaVersionValue := ""
	if schemaVersion, ok := rawEvent["schemaVersion"].(string); ok && schemaVersion != "" {
		hasSchemaVersion = true
		schemaVersionValue = schemaVersion
	}
	
	if err := json.Unmarshal(event, &envelope); err == nil && hasSchemaVersion && len(envelope.References) > 0 {
		log.Info("Detected S3 state envelope format", map[string]interface{}{
			"schemaVersion":  schemaVersionValue,
			"verificationId": envelope.VerificationID,
			"referenceCount": len(envelope.References),
		})
		
		// Convert envelope to input
		var err error
		inputFromEnvelope, err := state.EnvelopeToInput(&envelope)
		if err != nil {
			log.Error("Failed to convert envelope to input", map[string]interface{}{
				"error":          err.Error(),
				"verificationId": envelope.VerificationID,
			})
			return nil, err
		}
		
		// Create validator
		validator := validation.NewValidator(log)
		
		// Validate input
		if err := validator.ValidateInput(inputFromEnvelope); err != nil {
			log.Error("Validation error", map[string]interface{}{
				"error":          err.Error(),
				"verificationId": inputFromEnvelope.VerificationID,
			})
			return nil, err
		}
		
		// Use the converted input
		log.Info("Successfully parsed S3 state envelope", map[string]interface{}{
			"verificationId": inputFromEnvelope.VerificationID,
			"turnNumber":     inputFromEnvelope.TurnNumber,
			"includeImage":   inputFromEnvelope.IncludeImage,
		})
		
		// Continue with the converted input
		return processInput(ctx, inputFromEnvelope, start)
	}
	
	// If not an envelope, try to parse as direct input
	var input state.Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Error("Error parsing input", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.NewParsingError("JSON", err)
	}
	
	// Set default values for Turn 1 if not set
	if input.TurnNumber == 0 {
		input.TurnNumber = 1
	}
	
	if input.IncludeImage == "" {
		input.IncludeImage = "reference"
	}
	
	// Create validator
	validator := validation.NewValidator(log)
	
	// Validate input
	if err := validator.ValidateInput(&input); err != nil {
		log.Error("Validation error", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, err
	}
	
	// Continue with the input
	return processInput(ctx, &input, start)
}

// processInput processes the input and returns the output
func processInput(ctx context.Context, input *state.Input, start time.Time) (*state.Output, error) {
	// Log processing start
	log.Info("Processing Turn 1 prompt", map[string]interface{}{
		"verificationType": input.VerificationType,
		"verificationId":   input.VerificationID,
		"turnNumber":       input.TurnNumber,
		"includeImage":     input.IncludeImage,
		"referenceCount":   len(input.References),
	})
	
	// Create validator
	validator := validation.NewValidator(log)
	
	// Create state loader
	stateLoader := state.NewLoader(s3StateManager, log)
	
	// Load workflow state from S3
	workflowState, err := stateLoader.LoadWorkflowState(input)
	if err != nil {
		log.Error("Failed to load workflow state", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, errors.NewInternalError("state-loading", err)
	}
	
	// Log reference count for monitoring
	log.Info("Loaded workflow state with input references", map[string]interface{}{
		"verificationId": input.VerificationID,
		"referenceCount": len(input.References),
	})
	
	// Validate workflow state
	if err := validator.ValidateWorkflowState(workflowState); err != nil {
		log.Error("Workflow state validation error", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, err
	}
	
	// Create image processor
	imageProcessor := images.NewProcessor(s3StateManager, log)
	
	// Process images for Bedrock
	if err := imageProcessor.ProcessImagesForBedrock(ctx, workflowState); err != nil {
		log.Error("Failed to process images", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, errors.NewInternalError("image-processing", err)
	}
	
	// Create template processor
	templateProcessor := core.NewTemplateProcessor(templateLoader, log)
	
	// Create prompt generator
	promptGenerator := core.NewPromptGenerator(log, templateProcessor)
	
	// Generate prompt text
	promptText, err := promptGenerator.GeneratePrompt(workflowState)
	if err != nil {
		log.Error("Failed to generate prompt", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, errors.NewInternalError("prompt-generation", err)
	}
	
	// Create Bedrock integrator
	bedrockIntegrator := integration.NewBedrockIntegrator(log)
	
	// Create response builder
	responseBuilder := core.NewResponseBuilder(log, bedrockIntegrator)
	
	// Configure Bedrock
	responseBuilder.ConfigureBedrock(workflowState, input)
	
	// Build response
	templateName := promptGenerator.BuildTemplateName(workflowState.VerificationContext.VerificationType)
	if err := responseBuilder.BuildResponse(workflowState, promptText, templateProcessor.GetTemplateVersion(templateName)); err != nil {
		log.Error("Failed to build response", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, errors.NewInternalError("response-building", err)
	}
	
	// Create state saver
	stateSaver := state.NewSaver(s3StateManager, log)
	
	// Create output with input references - passing input.References to preserve them
	output := state.NewOutput(
		workflowState.VerificationContext.VerificationId,
		workflowState.VerificationContext.VerificationType,
		workflowState.VerificationContext.Status,
		input.References, // Pass input references for accumulation
	)
	
	// Save results to S3 - passing input for reference preservation
	if err := stateSaver.SaveTurn1Prompt(workflowState, input, output); err != nil {
		log.Error("Failed to save Turn 1 prompt", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": input.VerificationID,
		})
		return nil, errors.NewInternalError("state-saving", err)
	}
	
	// Log completion with reference counts
	duration := time.Since(start)
	log.Info("Completed processing with accumulated references", map[string]interface{}{
		"duration":              duration.String(),
		"verificationId":        input.VerificationID,
		"templateUsed":          templateName,
		"inputReferenceCount":   len(input.References),
		"outputReferenceCount":  len(output.References),
		"newReferencesAdded":    len(output.References) - len(input.References),
	})
	
	log.LogOutputEvent(output)
	return output, nil
}

// main is the entry point for the Lambda function
func main() {
	// Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(HandleRequest)
	} else {
		// Local execution for testing
		log.Info("Running in local mode", nil)
		
		// Read input from environment or file
		inputPath := os.Getenv("INPUT_FILE")
		if inputPath == "" {
			log.Error("INPUT_FILE environment variable is required for local execution", nil)
			os.Exit(1)
		}
		
		// Read input file
		inputData, err := os.ReadFile(inputPath)
		if err != nil {
			log.Error("Failed to read input file", map[string]interface{}{
				"error": err.Error(),
				"path":  inputPath,
			})
			os.Exit(1)
		}
		
		// Process input
		output, err := HandleRequest(context.Background(), inputData)
		if err != nil {
			log.Error("Error processing input", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
		
		// Write output to stdout
		outputData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(outputData))
	}
}