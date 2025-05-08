package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
	Body    json.RawMessage   `json:"body"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
}



// Handler is the Lambda handler function
func Handler(ctx context.Context, event interface{}) (*InitResponse, error) {
	// 1) Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		log.Printf("Failed to initialize dependencies: %v", err)
		return nil, err
	}
	logger := deps.GetLogger()
	logger.Info("Received event", map[string]interface{}{
		"eventType": fmt.Sprintf("%T", event),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	// 2) Marshal the incoming event to JSON bytes
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

	// 3) Unwrap if there is a top-level "verificationContext" field
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		logger.Error("Failed to unmarshal event to map", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	var request InitRequest
	if vc, ok := raw["verificationContext"]; ok {
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
			})
			return nil, fmt.Errorf("failed to parse verificationContext: %w", err)
		}
	} else {
		// No wrapper: parse the full payload directly
		if err := json.Unmarshal(jsonBytes, &request); err != nil {
			logger.Error("Failed to unmarshal direct JSON input", map[string]interface{}{
				"error":     err.Error(),
				"eventJson": string(jsonBytes),
			})
			return nil, fmt.Errorf("failed to parse event: %w", err)
		}
	}

	// 4) Default verificationType if missing
	if request.VerificationType == "" {
		logger.Info("VerificationType not provided, defaulting to LAYOUT_VS_CHECKING", nil)
		request.VerificationType = VerificationTypeLayoutVsChecking
	}

	// 5) Log parsed request
	logDetails := map[string]interface{}{
		"verificationType":    request.VerificationType,
		"referenceImageUrl":   request.ReferenceImageUrl,
		"checkingImageUrl":    request.CheckingImageUrl,
		"vendingMachineId":    request.VendingMachineId,
		"notificationEnabled": request.NotificationEnabled,
	}
	if request.VerificationType == VerificationTypeLayoutVsChecking {
		logDetails["layoutId"] = request.LayoutId
		logDetails["layoutPrefix"] = request.LayoutPrefix
	} else {
		logDetails["previousVerificationId"] = request.PreviousVerificationId
	}
	logger.Info("Parsed request", logDetails)

	// 6) Load config from environment
	config := ConfigVars{
		LayoutTable:        os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		VerificationTable:  os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		VerificationPrefix: getEnvWithDefault("VERIFICATION_PREFIX", "verif-"),
		ReferenceBucket:    os.Getenv("REFERENCE_BUCKET"),
		CheckingBucket:     os.Getenv("CHECKING_BUCKET"),
	}
	logger.Info("Using configuration", map[string]interface{}{
		"layoutTable":        config.LayoutTable,
		"verificationTable":  config.VerificationTable,
		"verificationPrefix": config.VerificationPrefix,
	})

	// 7) Run business logic
	service := NewInitService(deps, config)
	result, err := service.Process(ctx, request)
	if err != nil {
		logger.Error("Failed to process request", map[string]interface{}{
			"error":            err.Error(),
			"verificationType": request.VerificationType,
		})
		return nil, err
	}
	logger.Info("Successfully processed request", map[string]interface{}{
		"verificationId":   result.VerificationId,
		"verificationType": result.VerificationType,
		"status":           result.Status,
	})

	// 8) Wrap and return response
	message := fmt.Sprintf(
		"Verification initialized successfully. Will perform %s verification with two-turn approach.",
		strings.ToLower(request.VerificationType),
	)
	return &InitResponse{
		VerificationContext: result,
		Message:             message,
	}, nil
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

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context) (*Dependencies, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return NewDependencies(awsCfg), nil
}
