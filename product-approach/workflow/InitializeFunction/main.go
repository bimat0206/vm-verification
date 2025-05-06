package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
"fmt"
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
	Body        json.RawMessage         `json:"body"`
	Headers     map[string]string       `json:"headers"`
	Method      string                  `json:"method"`
	Params      map[string]string       `json:"params"`
	Query       map[string]string       `json:"query"`
}

// Handler is the Lambda handler function
// main.go - update Handler function
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
    
    // Log the parsed request for debugging
    logger.Info("Parsed request", map[string]interface{}{
        "referenceImageUrl": request.ReferenceImageUrl,
        "checkingImageUrl": request.CheckingImageUrl,
        "layoutId": request.LayoutId,
        "layoutPrefix": request.LayoutPrefix,
    })
    
    // Get configuration from environment
    config := ConfigVars{
        LayoutTable:        os.Getenv("DYNAMODB_LAYOUT_TABLE"),
        VerificationTable:  os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
        VerificationPrefix: os.Getenv("VERIFICATION_PREFIX"),
        ReferenceBucket:    os.Getenv("REFERENCE_BUCKET"),
        CheckingBucket:     os.Getenv("CHECKING_BUCKET"),
    }
    
    // Initialize service
    service := NewInitService(deps, config)
    
    // Process the request
    return service.Process(ctx, request)
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