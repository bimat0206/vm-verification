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
)

// ConfigVars holds environment configuration
type ConfigVars struct {
	LayoutTable        string
	VerificationTable  string
	VerificationPrefix string
	ReferenceBucket    string
	CheckingBucket     string
}

// WrappedRequest represents the structure API Gateway sends to Lambda
// with non-proxy integration
type WrappedRequest struct {
	Body        json.RawMessage   `json:"body"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
	Params      map[string]string `json:"params"`
	Query       map[string]string `json:"query"`
}

// Handler is the Lambda handler function
func Handler(ctx context.Context, event interface{}) (*VerificationContext, error) {
	// Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		log.Printf("Failed to initialize dependencies: %v", err)
		return nil, err
	}
	
	// Get logger
	logger := deps.GetLogger()
	logger.Info("Received event", map[string]interface{}{
		"eventType": fmt.Sprintf("%T", event),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	
	// Extract the request based on event type
	var request InitRequest
	
	// Try to determine if this is coming from API Gateway or Step Functions
	switch eventData := event.(type) {
	case WrappedRequest:
		// API Gateway integration
		if len(eventData.Body) > 0 {
			err := json.Unmarshal(eventData.Body, &request)
			if err != nil {
				logger.Error("Failed to unmarshal body from API Gateway", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, fmt.Errorf("failed to parse API Gateway request: %w", err)
			}
		}
	case map[string]interface{}:
		// Direct JSON input (likely from Step Functions)
		jsonBytes, err := json.Marshal(eventData)
		if err != nil {
			logger.Error("Failed to marshal raw event", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to process raw event: %w", err)
		}
		
		if err := json.Unmarshal(jsonBytes, &request); err != nil {
			logger.Error("Failed to unmarshal direct JSON input", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to parse Step Functions input: %w", err)
		}
	default:
		// Try to unmarshal directly as InitRequest
		jsonBytes, err := json.Marshal(event)
		if err != nil {
			logger.Error("Failed to marshal unknown event type", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("unknown event format: %w", err)
		}
		
		if err := json.Unmarshal(jsonBytes, &request); err != nil {
			logger.Error("Failed to unmarshal as InitRequest", map[string]interface{}{
				"error": err.Error(),
				"eventJson": string(jsonBytes),
			})
			return nil, fmt.Errorf("failed to parse event: %w", err)
		}
	}
	
	// Set default verificationType if not provided
	if request.VerificationType == "" {
		logger.Info("VerificationType not provided, defaulting to LAYOUT_VS_CHECKING", nil)
		request.VerificationType = VerificationTypeLayoutVsChecking
	}
	
	// Log the parsed request with details appropriate for the verification type
	logDetails := map[string]interface{}{
		"verificationType":  request.VerificationType,
		"referenceImageUrl": request.ReferenceImageUrl,
		"checkingImageUrl":  request.CheckingImageUrl,
		"vendingMachineId":  request.VendingMachineId,
	}
	
	// Add type-specific details to log
	if request.VerificationType == VerificationTypeLayoutVsChecking {
		logDetails["layoutId"] = request.LayoutId
		logDetails["layoutPrefix"] = request.LayoutPrefix
	} else if request.VerificationType == VerificationTypePreviousVsCurrent {
		logDetails["previousVerificationId"] = request.PreviousVerificationId
	}
	
	logger.Info("Parsed request", logDetails)
	
	// Get configuration from environment
	config := ConfigVars{
		LayoutTable:        os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		VerificationTable:  os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		VerificationPrefix: getEnvWithDefault("VERIFICATION_PREFIX", "verif-"),
		ReferenceBucket:    os.Getenv("REFERENCE_BUCKET"),
		CheckingBucket:     os.Getenv("CHECKING_BUCKET"),
	}
	
	// Log configuration (excluding sensitive values)
	logger.Info("Using configuration", map[string]interface{}{
		"layoutTable":        config.LayoutTable,
		"verificationTable":  config.VerificationTable,
		"verificationPrefix": config.VerificationPrefix,
	})
	
	// Initialize service
	service := NewInitService(deps, config)
	
	// Process the request
	result, err := service.Process(ctx, request)
	if err != nil {
		logger.Error("Failed to process request", map[string]interface{}{
			"error": err.Error(),
			"verificationType": request.VerificationType,
		})
		return nil, err
	}
	
	logger.Info("Successfully processed request", map[string]interface{}{
		"verificationId": result.VerificationId,
		"verificationType": result.VerificationType,
		"status": result.Status,
	})
	
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
	// Start Lambda handler
	lambda.Start(Handler)
}

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context) (*Dependencies, error) {
	// Load AWS SDK configuration
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize dependencies
	return NewDependencies(awsCfg), nil
}