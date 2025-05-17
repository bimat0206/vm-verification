package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"workflow-function/Initialize/internal"
	"workflow-function/shared/schema"
	 
)

// WrappedRequest represents the structure API Gateway sends to Lambda
// with non-proxy integration
type WrappedRequest struct {
	Body    json.RawMessage   `json:"body"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
}

// InitRequest represents the input payload to the Lambda function
// This struct continues to support legacy format for backwards compatibility
type InitRequest struct {
	// Either coming from a wrapper or direct fields
	SchemaVersion         string                  `json:"schemaVersion,omitempty"`
	VerificationContext   *schema.VerificationContext `json:"verificationContext,omitempty"`
	
	// Direct fields (legacy format)
	VerificationType      string              `json:"verificationType"`
	ReferenceImageUrl     string              `json:"referenceImageUrl"`
	CheckingImageUrl      string              `json:"checkingImageUrl"`
	VendingMachineId      string              `json:"vendingMachineId,omitempty"`
	LayoutId              int                 `json:"layoutId,omitempty"`
	LayoutPrefix          string              `json:"layoutPrefix,omitempty"`
	PreviousVerificationId string             `json:"previousVerificationId,omitempty"`
	ConversationConfig    *ConversationConfig `json:"conversationConfig,omitempty"`
	RequestId             string              `json:"requestId,omitempty"`
	RequestTimestamp      string              `json:"requestTimestamp,omitempty"`
	NotificationEnabled   bool                `json:"notificationEnabled"`
}

// ConversationConfig defines configuration for the conversation
type ConversationConfig struct {
	Type     string `json:"type"`
	MaxTurns int    `json:"maxTurns"`
}

// Handler is the Lambda handler function
func Handler(ctx context.Context, event interface{}) (interface{}, error) {
	// 1) Initialize AWS SDK config
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return nil, err
	}

	// 2) Initialize service with configuration
	cfg := internal.Config{
		LayoutTable:        os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		VerificationTable:  os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		VerificationPrefix: getEnvWithDefault("VERIFICATION_PREFIX", "verif-"),
		ReferenceBucket:    os.Getenv("REFERENCE_BUCKET"),
		CheckingBucket:     os.Getenv("CHECKING_BUCKET"),
	}
	
	// Create the service
	svc, err := internal.NewInitializeService(ctx, awsCfg, cfg)
	if err != nil {
		log.Printf("Failed to initialize service: %v", err)
		return nil, err
	}
	
	logger := svc.Logger()
	logger.Info("Received event", map[string]interface{}{
		"eventType": fmt.Sprintf("%T", event),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	// 3) Marshal the incoming event to JSON bytes
	var jsonBytes []byte
	switch e := event.(type) {
	case WrappedRequest:
		// API Gateway
		if len(e.Body) > 0 {
			jsonBytes = e.Body
		} else {
			jsonBytes = []byte("{}")
		}
	case map[string]interface{}:
		// Step Functions / direct map
		jsonBytes, err = json.Marshal(e)
		if err != nil {
			logger.Error("Failed to marshal raw event", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to process raw event: %w", err)
		}
	default:
		// Fallback: try to marshal entire event
		jsonBytes, err = json.Marshal(e)
		if err != nil {
			logger.Error("Failed to marshal unknown event type", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("unknown event format: %w", err)
		}
	}

	// 4) Parse the incoming event
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		logger.Error("Failed to unmarshal event to map", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	var request InitRequest
	
	// Check for schema version to determine format
	if schemaVersion, ok := raw["schemaVersion"].(string); ok && schemaVersion != "" {
		// This is the new standardized format
		logger.Info("Detected standardized schema format", map[string]interface{}{
			"schemaVersion": schemaVersion,
		})
		
		// Extract verification context
		if vc, exist := raw["verificationContext"]; exist {
			vcBytes, err := json.Marshal(vc)
			if err != nil {
				logger.Error("Failed to marshal verificationContext", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, fmt.Errorf("failed to parse verificationContext: %w", err)
			}
			
			// Unmarshal into the appropriate struct
			var verificationContext schema.VerificationContext
			if err := json.Unmarshal(vcBytes, &verificationContext); err != nil {
				logger.Error("Failed to unmarshal verificationContext", map[string]interface{}{
					"error": err.Error(),
					"jsonBytes": string(vcBytes),
				})
				return nil, fmt.Errorf("failed to parse verificationContext detail: %w", err)
			}
			
			// Populate request with verification context
			request.SchemaVersion = schemaVersion
			request.VerificationContext = &verificationContext
			
			// Extract request metadata
			if requestId, ok := raw["requestId"].(string); ok && requestId != "" {
				request.RequestId = requestId
			}
			if requestTimestamp, ok := raw["requestTimestamp"].(string); ok && requestTimestamp != "" {
				request.RequestTimestamp = requestTimestamp
			}
		}
	} else {
		// First check if this is a Step Functions invocation (by checking for specific parameters pattern)
		_, hasVerificationType := raw["verificationType"]
		_, hasReferenceImageUrl := raw["referenceImageUrl"]
		_, hasCheckingImageUrl := raw["checkingImageUrl"]
		
		// Step Functions invocation will have these parameters directly
		isStepFunctions := hasVerificationType && hasReferenceImageUrl && hasCheckingImageUrl
		
		if vc, exist := raw["verificationContext"]; exist {
			// If there's a verificationContext wrapper (legacy format), extract fields from it
			vcBytes, err := json.Marshal(vc)
			if err != nil {
				logger.Error("Failed to marshal verificationContext", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, fmt.Errorf("failed to parse verificationContext: %w", err)
			}
			if err := json.Unmarshal(vcBytes, &request); err != nil {
				logger.Error("Failed to unmarshal verificationContext", map[string]interface{}{
					"error": err.Error(),
					"jsonBytes": string(vcBytes),
				})
				return nil, fmt.Errorf("failed to parse verificationContext detail: %w", err)
			}
			
			// Extract top-level requestId and requestTimestamp if present
			if requestId, ok := raw["requestId"].(string); ok && requestId != "" {
				request.RequestId = requestId
			}
			if requestTimestamp, ok := raw["requestTimestamp"].(string); ok && requestTimestamp != "" {
				request.RequestTimestamp = requestTimestamp
			}
			
			logger.Info("Parsed request with verificationContext wrapper", map[string]interface{}{
				"requestId": request.RequestId,
				"requestTimestamp": request.RequestTimestamp,
			})
		} else if isStepFunctions {
			// Direct parameters from Step Functions (no wrapper)
			if err := json.Unmarshal(jsonBytes, &request); err != nil {
				logger.Error("Failed to unmarshal direct Step Functions input", map[string]interface{}{
					"error": err.Error(),
					"eventJson": string(jsonBytes),
				})
				return nil, fmt.Errorf("failed to parse Step Functions input: %w", err)
			}
			logger.Info("Parsed direct Step Functions request", map[string]interface{}{
				"verificationType": request.VerificationType,
			})
		} else {
			// API Gateway or direct Lambda invocation without wrapper
			if err := json.Unmarshal(jsonBytes, &request); err != nil {
				logger.Error("Failed to unmarshal direct JSON input", map[string]interface{}{
					"error": err.Error(),
					"eventJson": string(jsonBytes),
				})
				return nil, fmt.Errorf("failed to parse event detail: %w", err)
			}
		}
	}

	// 5) Default verificationType if missing
	if request.VerificationContext == nil &&
	   request.VerificationType == "" {
		logger.Info("VerificationType not provided, defaulting to LAYOUT_VS_CHECKING", nil)
		request.VerificationType = schema.VerificationTypeLayoutVsChecking
	}

	// 6) Log parsed request
	logDetails := map[string]interface{}{}
	
	if request.VerificationContext != nil {
		// Log from verification context
		logDetails["verificationType"] = request.VerificationContext.VerificationType
		logDetails["referenceImageUrl"] = request.VerificationContext.ReferenceImageUrl
		logDetails["checkingImageUrl"] = request.VerificationContext.CheckingImageUrl
		logDetails["vendingMachineId"] = request.VerificationContext.VendingMachineId
		logDetails["notificationEnabled"] = request.VerificationContext.NotificationEnabled
		
		if request.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
			logDetails["layoutId"] = request.VerificationContext.LayoutId
			logDetails["layoutPrefix"] = request.VerificationContext.LayoutPrefix
		} else {
			logDetails["previousVerificationId"] = request.VerificationContext.PreviousVerificationId
		}
	} else {
		// Log from direct fields
		logDetails["verificationType"] = request.VerificationType
		logDetails["referenceImageUrl"] = request.ReferenceImageUrl
		logDetails["checkingImageUrl"] = request.CheckingImageUrl
		logDetails["vendingMachineId"] = request.VendingMachineId
		logDetails["notificationEnabled"] = request.NotificationEnabled
		
		if request.VerificationType == schema.VerificationTypeLayoutVsChecking {
			logDetails["layoutId"] = request.LayoutId
			logDetails["layoutPrefix"] = request.LayoutPrefix
		} else {
			logDetails["previousVerificationId"] = request.PreviousVerificationId
		}
	}
	
	logger.Info("Parsed request", logDetails)

	// Log configuration
	logger.Info("Using configuration", map[string]interface{}{
		"layoutTable":        cfg.LayoutTable,
		"verificationTable":  cfg.VerificationTable,
		"verificationPrefix": cfg.VerificationPrefix,
	})

	// 7) Process the request with our internal service
	result, err := svc.Process(ctx, internal.ProcessRequest{
		SchemaVersion:         request.SchemaVersion,
		VerificationContext:   request.VerificationContext,
		VerificationType:      request.VerificationType,
		ReferenceImageUrl:     request.ReferenceImageUrl,
		CheckingImageUrl:      request.CheckingImageUrl,
		VendingMachineId:      request.VendingMachineId,
		LayoutId:              request.LayoutId,
		LayoutPrefix:          request.LayoutPrefix,
		PreviousVerificationId: request.PreviousVerificationId,
		RequestId:             request.RequestId,
		RequestTimestamp:      request.RequestTimestamp,
		NotificationEnabled:   request.NotificationEnabled,
		ConversationConfig: internal.ConversationConfig{
			Type:     request.ConversationConfig.Type,
			MaxTurns: request.ConversationConfig.MaxTurns,
		},
	})
	
	if err != nil {
		logger.Error("Failed to process request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	
	// 8) Return result
	// Always return just the VerificationContext for Step Functions
	// This prevents nesting and follows the state machine design
	return result, nil
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	lambda.Start(Handler)
}