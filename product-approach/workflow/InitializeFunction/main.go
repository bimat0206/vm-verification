package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

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
func Handler(ctx context.Context, wrappedRequest WrappedRequest) (*VerificationContext, error) {
	// Extract the actual request from the wrapped body field
	var request InitRequest
	
	// If Body is a JSON string (happens with some API Gateway configurations)
	if len(wrappedRequest.Body) > 0 {
		// Try unmarshaling directly
		err := json.Unmarshal(wrappedRequest.Body, &request)
		if err != nil {
			// If direct unmarshal fails, it might be a string that needs another unmarshal
			var bodyString string
			if err := json.Unmarshal(wrappedRequest.Body, &bodyString); err == nil {
				// Successfully got a string, now try to unmarshal that
				if err := json.Unmarshal([]byte(bodyString), &request); err != nil {
					log.Printf("Failed to unmarshal body string: %v", err)
					return nil, err
				}
			} else {
				log.Printf("Failed to unmarshal body: %v", err)
				return nil, err
			}
		}
	}

	// Get configuration from environment
	config := ConfigVars{
		LayoutTable:        os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		VerificationTable:  os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		VerificationPrefix: os.Getenv("VERIFICATION_PREFIX"),
		ReferenceBucket:    os.Getenv("REFERENCE_BUCKET"),
		CheckingBucket:     os.Getenv("CHECKING_BUCKET"),
	}

	// Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		log.Printf("Failed to initialize dependencies: %v", err)
		return nil, err
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