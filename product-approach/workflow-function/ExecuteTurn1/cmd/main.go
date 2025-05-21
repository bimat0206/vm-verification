package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"

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
   	"bedrockModel": cfg.BedrockModel,
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
   	log.Error("Failed to parse input event", map[string]interface{}{
   		"error": err.Error(),
   		"event": string(event),
   	})
   	return createErrorResponse("invalid_input", "Invalid input format", map[string]interface{}{
   		"error": err.Error(),
   	}, ""), nil
   }

   // Ensure we have a verification ID for later reference
   verificationId := extractVerificationId(input)

   log.Info("Input parsed successfully", map[string]interface{}{
   	"hasStateReferences": input.StateReferences != nil,
   	"hasS3References": input.S3References != nil && len(input.S3References) > 0,
   	"verificationId": verificationId,
   })

   // Map S3References to StateReferences if needed
   if input.StateReferences == nil && input.S3References != nil {
   	mappedRefs := input.MapS3References()
   	input.StateReferences = mappedRefs
   	
   	// If we have verification ID from input but not in the mapped refs, add it
   	if mappedRefs != nil && mappedRefs.VerificationId == "" && verificationId != "" {
   		mappedRefs.VerificationId = verificationId
   	}
   	
   	log.Info("Mapped S3References to StateReferences", map[string]interface{}{
   		"mappedReferenceFields": countNonNilFields(input.StateReferences),
   		"originalS3ReferenceCount": len(input.S3References),
   		"verificationId": verificationId,
   	})
   }

   // Handle the request using the main Handler
   output, err := clients.Handler.HandleRequest(ctx, &input)
   if err != nil {
   	log.Error("Handler error", map[string]interface{}{
   		"error": err.Error(), 
   		"verificationId": verificationId,
   	})
   	
   	// Return a proper error response for Step Functions
   	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
   		return createErrorResponseFromWFError(wfErr, verificationId), nil
   	}
   	
   	// Wrap unexpected errors
   	wfErr := wferrors.WrapError(err, wferrors.ErrorTypeInternal, "unexpected handler error", false)
   	return createErrorResponseFromWFError(wfErr, verificationId), nil
   }

   // If we got here, we have a successful output
   // Use safe logging to avoid nil pointer dereferences
   outputVerificationId := getVerificationIdSafely(output, verificationId)
   outputStatus := getSafeString(output.Status, "UNKNOWN")

   log.Info("Successfully completed ExecuteTurn1", map[string]interface{}{
   	"verificationId": outputVerificationId,
   	"status":        outputStatus,
   })

   // Ensure S3References is populated for step function compatibility
   if output != nil && output.StateReferences != nil && output.S3References == nil {
   	output.S3References = output.StateReferences
   }

   return output, nil
}

// extractVerificationId gets verification ID from input in a safe manner
func extractVerificationId(input internal.StepFunctionInput) string {
   // Try from direct field first
   if input.VerificationId != "" {
      return input.VerificationId
   }
   
   // Try from StateReferences if available
   if input.StateReferences != nil && input.StateReferences.VerificationId != "" {
      return input.StateReferences.VerificationId
   }
   
   // Try from S3References prefixes if available
   if input.S3References != nil {
      // Look for common reference pattern that might contain verification ID
      for _, ref := range input.S3References {
         if ref != nil {
            // Access the "key" field from the map
            if keyVal, ok := ref["key"].(string); ok && keyVal != "" {
               // Extract verification ID from key pattern like "2025/05/21/verif-12345/..."
               parts := extractPathParts(keyVal)
               if len(parts) >= 4 && parts[3] != "" && isLikelyVerificationId(parts[3]) {
                  return parts[3]
               }
            }
         }
      }
   }
   
   return ""
}

// isLikelyVerificationId checks if a string looks like a verification ID
func isLikelyVerificationId(s string) bool {
   return len(s) >= 5 && (s[:5] == "verif-" || s[:3] == "vrf")
}

// extractPathParts splits a path into components
func extractPathParts(path string) []string {
   // Split by forward slash and return all parts
   parts := make([]string, 0)
   for _, part := range strings.Split(path, "/") {
      if part != "" {
         parts = append(parts, part)
      }
   }
   return parts
}

// createErrorResponse creates a StepFunctionOutput with error information
func createErrorResponse(code, message string, details map[string]interface{}, verificationId string) *internal.StepFunctionOutput {
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
   
   // Add verification ID if available
   if verificationId != "" {
      output.StateReferences = &internal.StateReferences{
         VerificationId: verificationId,
      }
      output.S3References = output.StateReferences
      
      // Add to summary for easier access
      output.Summary["verificationId"] = verificationId
   }
   
   return output
}

// createErrorResponseFromWFError creates a StepFunctionOutput from a WorkflowError
func createErrorResponseFromWFError(wfErr *wferrors.WorkflowError, verificationId string) *internal.StepFunctionOutput {
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
   
   // Add verification ID if available
   if verificationId != "" {
      output.StateReferences = &internal.StateReferences{
         VerificationId: verificationId,
      }
      output.S3References = output.StateReferences
      
      // Add to summary for easier access
      output.Summary["verificationId"] = verificationId
   }
   
   return output
}

// countNonNilFields counts how many non-nil fields are in the StateReferences
func countNonNilFields(refs *internal.StateReferences) int {
   if refs == nil {
      return 0
   }
   
   count := 0
   
   // Count each non-nil reference field
   if refs.Initialization != nil { count++ }
   if refs.ImageMetadata != nil { count++ }
   if refs.SystemPrompt != nil { count++ }
   if refs.LayoutMetadata != nil { count++ }
   if refs.HistoricalContext != nil { count++ }
   if refs.ConversationState != nil { count++ }
   if refs.ReferenceBase64 != nil { count++ }
   if refs.CheckingBase64 != nil { count++ }
   if refs.Turn1Prompt != nil { count++ }
   if refs.Turn1Response != nil { count++ }
   if refs.Turn1Thinking != nil { count++ }
   if refs.Turn2Response != nil { count++ }
   if refs.Turn2Thinking != nil { count++ }
   
   return count
}

// getVerificationIdSafely gets verification ID from output without causing nil pointer issues
func getVerificationIdSafely(output *internal.StepFunctionOutput, fallback string) string {
   if output == nil || output.StateReferences == nil || output.StateReferences.VerificationId == "" {
      return fallback
   }
   return output.StateReferences.VerificationId
}

// getSafeString returns the string value or a fallback if empty
func getSafeString(value string, fallback string) string {
   if value == "" {
      return fallback
   }
   return value
}

func main() {
   lambda.Start(LambdaHandler)
}