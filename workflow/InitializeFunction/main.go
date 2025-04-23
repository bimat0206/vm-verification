package main

import (
	"context"
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

// Handler is the Lambda handler function
func Handler(ctx context.Context, request InitRequest) (*VerificationContext, error) {
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