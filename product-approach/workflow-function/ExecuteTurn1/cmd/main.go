package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"execute-turn1/internal"
)

// Global configuration loaded from environment variables
var env *internal.Environment

// init loads environment configuration
func init() {
	var err error
	env, err = loadEnvironment()
	if err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}
}

// loadEnvironment loads configuration from environment variables
func loadEnvironment() (*internal.Environment, error) {
	// Get required environment variables
	bedrockModelID := os.Getenv("BEDROCK_MODEL")
	if bedrockModelID == "" {
		return nil, &internal.ExecuteTurn1Error{
			Type:    internal.ErrorTypeValidation,
			Message: "BEDROCK_MODEL environment variable is required",
			Code:    "MISSING_ENV_VAR",
		}
	}

	bedrockRegion := os.Getenv("BEDROCK_REGION")
	if bedrockRegion == "" {
		bedrockRegion = "us-east-1" // Default region
	}

	anthropicVersion := os.Getenv("ANTHROPIC_VERSION")
	if anthropicVersion == "" {
		anthropicVersion = "bedrock-2023-05-31" // Default version
	}

	// Parse integer environment variables
	maxTokens, err := strconv.Atoi(getEnvOrDefault("MAX_TOKENS", "24000"))
	if err != nil {
		return nil, &internal.ExecuteTurn1Error{
			Type:    internal.ErrorTypeValidation,
			Message: "Invalid MAX_TOKENS value",
			Code:    "INVALID_ENV_VAR",
		}
	}

	budgetTokens, err := strconv.Atoi(getEnvOrDefault("BUDGET_TOKENS", "16000"))
	if err != nil {
		return nil, &internal.ExecuteTurn1Error{
			Type:    internal.ErrorTypeValidation,
			Message: "Invalid BUDGET_TOKENS value",
			Code:    "INVALID_ENV_VAR",
		}
	}

	retryMaxAttempts, err := strconv.Atoi(getEnvOrDefault("RETRY_MAX_ATTEMPTS", "3"))
	if err != nil {
		return nil, &internal.ExecuteTurn1Error{
			Type:    internal.ErrorTypeValidation,
			Message: "Invalid RETRY_MAX_ATTEMPTS value",
			Code:    "INVALID_ENV_VAR",
		}
	}

	retryBaseDelay, err := strconv.Atoi(getEnvOrDefault("RETRY_BASE_DELAY", "2000"))
	if err != nil {
		return nil, &internal.ExecuteTurn1Error{
			Type:    internal.ErrorTypeValidation,
			Message: "Invalid RETRY_BASE_DELAY value",
			Code:    "INVALID_ENV_VAR",
		}
	}

	return &internal.Environment{
		BedrockModelID:     bedrockModelID,
		BedrockRegion:      bedrockRegion,
		AnthropicVersion:   anthropicVersion,
		MaxTokens:          maxTokens,
		BudgetTokens:       budgetTokens,
		ThinkingType:       getEnvOrDefault("THINKING_TYPE", "enable"),
		DynamoDBTable:      getEnvOrDefault("DYNAMODB_CONVERSATION_TABLE", ""),
		RetryMaxAttempts:   retryMaxAttempts,
		RetryBaseDelay:     retryBaseDelay,
	}, nil
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Handler is the Lambda handler
func Handler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	// Start timing
	startTime := time.Now()
	
	log.Printf("ExecuteTurn1 function started")

	// Parse input event
	var input internal.ExecuteTurn1Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Printf("Failed to unmarshal input: %v", err)
		return internal.ToLambdaErrorResponse(
			internal.NewParsingError("input JSON", err),
			"",
		), nil
	}

	log.Printf("Processing verification: %s, type: %s", 
		input.VerificationContext.VerificationID, 
		input.VerificationContext.VerificationType)

	// Validate input
	if err := internal.ValidateExecuteTurn1Input(&input); err != nil {
		log.Printf("Input validation failed: %v", err)
		err = internal.SetVerificationID(err, input.VerificationContext.VerificationID)
		return internal.ToLambdaErrorResponse(err, input.VerificationContext.VerificationID), nil
	}

	// Initialize Bedrock client
	bedrockClient, err := internal.NewBedrockClient(env.BedrockRegion, env.BedrockModelID)
	if err != nil {
		log.Printf("Failed to initialize Bedrock client: %v", err)
		err = internal.SetVerificationID(err, input.VerificationContext.VerificationID)
		return internal.ToLambdaErrorResponse(err, input.VerificationContext.VerificationID), nil
	}

	// Initialize retry manager
	retryManager := internal.NewRetryManager(env.RetryMaxAttempts, env.RetryBaseDelay)

	// Initialize response processor
	responseProcessor := internal.NewResponseProcessor()

	// Execute Turn 1 with retry logic
	output, err := executeTurn1WithRetry(
		ctx,
		bedrockClient,
		retryManager,
		responseProcessor,
		&input,
		startTime,
	)

	if err != nil {
		log.Printf("ExecuteTurn1 failed: %v", err)
		err = internal.SetVerificationID(err, input.VerificationContext.VerificationID)
		return internal.ToLambdaErrorResponse(err, input.VerificationContext.VerificationID), nil
	}

	// Log success
	duration := time.Since(startTime)
	log.Printf("ExecuteTurn1 completed successfully in %v", duration)

	return output, nil
}

// executeTurn1WithRetry executes Turn 1 with retry logic
func executeTurn1WithRetry(
	ctx context.Context,
	bedrockClient *internal.BedrockClient,
	retryManager *internal.RetryManager,
	responseProcessor *internal.ResponseProcessor,
	input *internal.ExecuteTurn1Input,
	startTime time.Time,
) (*internal.ExecuteTurn1Output, error) {
	var lastErr error
	var output *internal.ExecuteTurn1Output

	// Execute with retry logic
	success := retryManager.ExecuteWithRetry(ctx, func() error {
		// Call Bedrock API
		bedrockResponse, latency, err := bedrockClient.InvokeModel(ctx, input)
		if err != nil {
			lastErr = err
			return err
		}

		// Process response
		turn1Response, err := responseProcessor.ProcessTurn1Response(
			bedrockResponse,
			input,
			latency,
			time.Now(),
		)
		if err != nil {
			lastErr = err
			return err
		}

		// Create output
		output = &internal.ExecuteTurn1Output{
			VerificationContext: input.VerificationContext,
			Turn1Response:      *turn1Response,
			ConversationState: internal.ConversationState{
				CurrentTurn: 1,
				MaxTurns:    2,
				History: []internal.TurnHistory{
					{
						TurnID:        1,
						Timestamp:     turn1Response.Timestamp,
						Prompt:        turn1Response.Prompt,
						Response:      turn1Response.Response.Content,
						LatencyMs:     turn1Response.LatencyMs,
						TokenUsage:    turn1Response.TokenUsage,
						AnalysisStage: turn1Response.AnalysisStage,
					},
				},
			},
		}

		// Update verification context status
		output.VerificationContext.Status = "TURN1_COMPLETED"
		if output.VerificationContext.TurnTimestamps.Turn1 == nil {
			now := time.Now()
			output.VerificationContext.TurnTimestamps.Turn1 = &now
		}

		return nil
	})

	if !success {
		return nil, lastErr
	}

	return output, nil
}

// main is the entry point for the Lambda function
func main() {
	lambda.Start(Handler)
}