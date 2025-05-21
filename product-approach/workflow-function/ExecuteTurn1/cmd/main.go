package main

import (
   "context"
   "encoding/json"
   "os"

   "github.com/aws/aws-lambda-go/lambda"
   "github.com/google/uuid"

   "workflow-function/shared/schema"
   "workflow-function/shared/logger"
   wferrors "workflow-function/shared/errors"

   "workflow-function/ExecuteTurn1/internal"
   "workflow-function/ExecuteTurn1/internal/config"
   "workflow-function/ExecuteTurn1/internal/dependencies"
)

// Global handler and dependencies for re-use between Lambda invocations
var clients *dependencies.Clients
var log logger.Logger

func init() {
   // Initialize logger
   log = logger.New("vending-verification", "ExecuteTurn1")
   log.Info("Initializing ExecuteTurn1 Lambda function", nil)

   // Load config
   cfg, err := config.New(log)
   if err != nil {
   	log.Error("Failed to load config", map[string]interface{}{"error": err.Error()})
   	os.Exit(1)
   }

   // Set up all dependencies
   clients, err = dependencies.New(context.Background(), cfg, log)
   if err != nil {
   	log.Error("Failed to initialize dependencies", map[string]interface{}{"error": err.Error()})
   	os.Exit(1)
   }

   log.Info("Lambda initialization completed successfully", map[string]interface{}{
   	"bedrockModel": cfg.BedrockModel, // Changed from BedrockModelID
   	"stateBucket":  cfg.StateBucket,
   })
}

// LambdaHandler - main entrypoint for Lambda
func LambdaHandler(ctx context.Context, event json.RawMessage) (interface{}, error) {
   // Generate request ID and configure context and logging
   requestID := uuid.NewString()
   ctx = context.WithValue(ctx, "requestID", requestID)
   log := log.WithCorrelationId(requestID)

   log.Info("Starting ExecuteTurn1 Lambda invocation", nil)

   // Parse the input event into StepFunctionInput
   var input internal.StepFunctionInput
   if err := json.Unmarshal(event, &input); err != nil {
   	log.Error("Failed to parse input event", map[string]interface{}{"error": err.Error()})
   	return createErrorResponse("invalid_input", "Invalid input format", map[string]interface{}{
   		"error": err.Error(),
   	}), nil
   }

   // Handle the request using the main Handler
   output, err := clients.Handler.HandleRequest(ctx, &input)
   if err != nil {
   	log.Error("Handler error", map[string]interface{}{"error": err.Error()})
   	
   	// Return a proper error response for Step Functions
   	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
   		return createErrorResponseFromWFError(wfErr), nil
   	}
   	
   	// Wrap unexpected errors
   	wfErr := wferrors.WrapError(err, wferrors.ErrorTypeInternal, "unexpected handler error", false)
   	return createErrorResponseFromWFError(wfErr), nil
   }

   log.Info("Successfully completed ExecuteTurn1", map[string]interface{}{
   	"verificationId": output.StateReferences.VerificationId,
   	"status":         output.Status,
   })

   // Ensure S3References is populated for step function compatibility
   if output.StateReferences != nil {
   	output.S3References = output.StateReferences
   }

   return output, nil
}

// createErrorResponse creates a StepFunctionOutput with error information
func createErrorResponse(code, message string, details map[string]interface{}) *internal.StepFunctionOutput {
   output := &internal.StepFunctionOutput{
   	Status: schema.StatusBedrockProcessingFailed,
   	Error: &schema.ErrorInfo{
   		Code:      code,
   		Message:   message,
   		Timestamp: schema.FormatISO8601(),
   		Details:   details,
   	},
   	Summary: map[string]interface{}{
   		"error":  message,
   		"status": schema.StatusBedrockProcessingFailed,
   	},
   }
   
   // For step function compatibility
   if output.StateReferences != nil {
   	output.S3References = output.StateReferences
   }
   
   return output
}

// createErrorResponseFromWFError creates a StepFunctionOutput from a WorkflowError
func createErrorResponseFromWFError(wfErr *wferrors.WorkflowError) *internal.StepFunctionOutput {
   output := &internal.StepFunctionOutput{
   	Status: schema.StatusBedrockProcessingFailed,
   	Error: &schema.ErrorInfo{
   		Code:      wfErr.Code,
   		Message:   wfErr.Message,
   		Timestamp: schema.FormatISO8601(),
   		Details:   wfErr.Context,
   	},
   	Summary: map[string]interface{}{
   		"error":     wfErr.Message,
   		"status":    schema.StatusBedrockProcessingFailed,
   		"retryable": wfErr.Retryable,
   		"errorType": wfErr.Type,
   	},
   }
   
   // For step function compatibility
   if output.StateReferences != nil {
   	output.S3References = output.StateReferences
   }
   
   return output
}

func main() {
   lambda.Start(LambdaHandler)
}