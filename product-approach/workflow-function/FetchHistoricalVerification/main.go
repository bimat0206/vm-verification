package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	dbClient *DynamoDBClient
	logger   *log.Logger
)

func init() {
	// Initialize logger
	logger = log.New(os.Stdout, "[FetchHistoricalVerification] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getRegion()),
	)
	if err != nil {
		logger.Fatalf("unable to load SDK config: %v", err)
	}

	// Initialize DynamoDB client
	dbClient = NewDynamoDBClient(cfg)
}

func handler(ctx context.Context, event InputEvent) (OutputEvent, error) {
	logger.Printf("Processing event: %+v", event)

	// Create service instance
	service := NewHistoricalVerificationService(dbClient, logger)

	// Validate input
	if err := validateInput(event.VerificationContext); err != nil {
		logger.Printf("Input validation error: %v", err)
		return OutputEvent{}, fmt.Errorf("input validation error: %w", err)
	}

	// Process the request
	result, err := service.FetchHistoricalVerification(ctx, event.VerificationContext)
	if err != nil {
		logger.Printf("Error fetching historical verification: %v", err)
		return OutputEvent{}, fmt.Errorf("error fetching historical verification: %w", err)
	}

	// Return the result
	return OutputEvent{
		HistoricalContext: result,
	}, nil
}

func main() {
	lambda.Start(handler)
}