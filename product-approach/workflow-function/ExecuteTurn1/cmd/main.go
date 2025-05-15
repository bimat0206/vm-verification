package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// Global configuration loaded from environment variables
var env *Environment

// Environment contains configuration from environment variables
type Environment struct {
	BedrockModelID     string
	BedrockRegion      string
	AnthropicVersion   string
	MaxTokens          int
	BudgetTokens       int
	ThinkingType       string
	DynamoDBTable      string
	RetryMaxAttempts   int
	RetryBaseDelay     int
}

// ExecuteTurn1Input represents the input to ExecuteTurn1 function
type ExecuteTurn1Input struct {
	VerificationContext schema.VerificationContext `json:"verificationContext"`
	CurrentPrompt       CurrentPromptWrapper       `json:"currentPrompt"`
	BedrockConfig       schema.BedrockConfig       `json:"bedrockConfig"`
	Images              *schema.Images             `json:"images,omitempty"`
	LayoutMetadata      *schema.LayoutMetadata     `json:"layoutMetadata,omitempty"`
	SystemPrompt        *SystemPromptWrapper       `json:"systemPrompt,omitempty"`
	HistoricalContext   map[string]interface{}     `json:"historicalContext,omitempty"`
	ConversationState   map[string]interface{}     `json:"conversationState,omitempty"`
}

// CurrentPromptWrapper wraps the nested currentPrompt structure
type CurrentPromptWrapper struct {
	CurrentPrompt CurrentPrompt            `json:"currentPrompt"`
	Messages      []bedrock.MessageWrapper `json:"messages,omitempty"`
	TurnNumber    int                      `json:"turnNumber,omitempty"`
	PromptID      string                   `json:"promptId,omitempty"`
	CreatedAt     time.Time                `json:"createdAt,omitempty"`
	PromptVersion string                   `json:"promptVersion,omitempty"`
	ImageIncluded string                   `json:"imageIncluded,omitempty"`
}

// CurrentPrompt represents the prompt structure for Turn 1
type CurrentPrompt struct {
	Messages       []bedrock.MessageWrapper `json:"messages"`
	TurnNumber     int                      `json:"turnNumber"`
	PromptID       string                   `json:"promptId"`
	CreatedAt      time.Time                `json:"createdAt"`
	PromptVersion  string                   `json:"promptVersion"`
	ImageIncluded  string                   `json:"imageIncluded"`
}

// SystemPromptWrapper wraps the nested systemPrompt structure
type SystemPromptWrapper struct {
	SystemPrompt SystemPrompt `json:"systemPrompt"`
}

// SystemPrompt represents the system prompt data
type SystemPrompt struct {
	Content       string `json:"content"`
	PromptID      string `json:"promptId"`
	CreatedAt     string `json:"createdAt"`
	PromptVersion string `json:"promptVersion"`
}

// ExecuteTurn1Output represents the output from ExecuteTurn1 function
type ExecuteTurn1Output struct {
	VerificationContext schema.VerificationContext `json:"verificationContext"`
	Turn1Response       bedrock.Turn1Response      `json:"turn1Response"`
	ConversationState   ConversationState          `json:"conversationState"`
}

// ConversationState represents the conversation state
type ConversationState struct {
	CurrentTurn int           `json:"currentTurn"`
	MaxTurns    int           `json:"maxTurns"`
	History     []TurnHistory `json:"history"`
}

// TurnHistory represents a single turn in conversation history
type TurnHistory struct {
	TurnID        int                `json:"turnId"`
	Timestamp     time.Time          `json:"timestamp"`
	Prompt        string             `json:"prompt"`
	Response      string             `json:"response"`
	LatencyMs     int64              `json:"latencyMs"`
	TokenUsage    bedrock.TokenUsage `json:"tokenUsage"`
	AnalysisStage string             `json:"analysisStage"`
}

// init loads environment configuration
func init() {
	var err error
	env, err = loadEnvironment()
	if err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}
}

// loadEnvironment loads configuration from environment variables
func loadEnvironment() (*Environment, error) {
	// Get required environment variables
	bedrockModelID := os.Getenv("BEDROCK_MODEL")
	if bedrockModelID == "" {
		return nil, fmt.Errorf("BEDROCK_MODEL environment variable is required")
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
		return nil, fmt.Errorf("invalid MAX_TOKENS value")
	}

	budgetTokens, err := strconv.Atoi(getEnvOrDefault("BUDGET_TOKENS", "16000"))
	if err != nil {
		return nil, fmt.Errorf("invalid BUDGET_TOKENS value")
	}

	retryMaxAttempts, err := strconv.Atoi(getEnvOrDefault("RETRY_MAX_ATTEMPTS", "3"))
	if err != nil {
		return nil, fmt.Errorf("invalid RETRY_MAX_ATTEMPTS value")
	}

	retryBaseDelay, err := strconv.Atoi(getEnvOrDefault("RETRY_BASE_DELAY", "2000"))
	if err != nil {
		return nil, fmt.Errorf("invalid RETRY_BASE_DELAY value")
	}

	return &Environment{
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
	var input ExecuteTurn1Input
	if err := json.Unmarshal(event, &input); err != nil {
		log.Printf("Failed to unmarshal input: %v", err)
		return nil, errors.NewParsingError("input JSON", err)
	}

	log.Printf("Processing verification: %s, type: %s", 
		input.VerificationContext.VerificationId, 
		input.VerificationContext.VerificationType)

	// Create Bedrock client configuration
	clientConfig := bedrock.CreateClientConfig(
		env.BedrockRegion,
		env.AnthropicVersion,
		env.MaxTokens,
		env.ThinkingType,
		env.BudgetTokens,
	)

	// Initialize Bedrock client
	bedrockClient, err := bedrock.NewBedrockClient(ctx, env.BedrockModelID, clientConfig)
	if err != nil {
		log.Printf("Failed to initialize Bedrock client: %v", err)
		return nil, errors.NewBedrockError("Failed to initialize Bedrock client", "CLIENT_INIT_ERROR", false)
	}

	// Initialize response processor
	responseProcessor := bedrock.NewResponseProcessor()

	// Execute Turn 1
	output, err := executeTurn1(
		ctx,
		bedrockClient,
		responseProcessor,
		&input,
		startTime,
	)

	if err != nil {
		log.Printf("ExecuteTurn1 failed: %v", err)
		return nil, err
	}

	// Log success
	duration := time.Since(startTime)
	log.Printf("ExecuteTurn1 completed successfully in %v", duration)

	return output, nil
}

// extractPromptText extracts the text portion from the current prompt
func extractPromptText(messages []bedrock.MessageWrapper) (string, error) {
	if len(messages) == 0 {
		return "", errors.NewValidationError("No messages in current prompt", nil)
	}

	var textParts []string

	// Extract text from all messages
	for _, message := range messages {
		for _, content := range message.Content {
			if content.Type == "text" {
				textParts = append(textParts, content.Text)
			}
		}
	}

	if len(textParts) == 0 {
		return "", errors.NewValidationError("No text content found in prompt", nil)
	}

	if len(textParts) > 0 {
		return textParts[0], nil
	}
	
	return "", errors.NewValidationError("No text content found in prompt", nil)
}

// executeTurn1 executes Turn 1
func executeTurn1(
	ctx context.Context,
	bedrockClient *bedrock.BedrockClient,
	responseProcessor *bedrock.ResponseProcessor,
	input *ExecuteTurn1Input,
	startTime time.Time,
) (*ExecuteTurn1Output, error) {
	// Extract current prompt
	currentPrompt := input.CurrentPrompt.CurrentPrompt
	if len(currentPrompt.Messages) == 0 {
		return nil, errors.NewValidationError("No messages in current prompt", nil)
	}

	// Extract system prompt if available
	var systemPromptText string
	if input.SystemPrompt != nil {
		systemPromptText = input.SystemPrompt.SystemPrompt.Content
	}

	// Set temperature and topP if provided
	var temperature, topP *float64
	if input.BedrockConfig.Temperature > 0 {
		temp := input.BedrockConfig.Temperature
		temperature = &temp
	}
	
	if input.BedrockConfig.TopP > 0 {
		tp := input.BedrockConfig.TopP
		topP = &tp
	}

	// Create Converse request
	converseRequest := bedrock.CreateConverseRequest(
		env.BedrockModelID,
		currentPrompt.Messages,
		systemPromptText,
		env.MaxTokens,
	)
	
	// Set temperature and topP if provided
	if temperature != nil {
		converseRequest.InferenceConfig.Temperature = temperature
	}
	
	if topP != nil {
		converseRequest.InferenceConfig.TopP = topP
	}

	// Call Bedrock Converse API
	converseResponse, latencyMs, err := bedrockClient.Converse(ctx, converseRequest)
	if err != nil {
		return nil, errors.NewBedrockError("Bedrock Converse API call failed", "API_ERROR", true)
	}

	// Extract prompt text
	promptText, err := extractPromptText(currentPrompt.Messages)
	if err != nil {
		return nil, err
	}

	// Process response
	turn1Response, err := responseProcessor.ProcessTurn1Response(
		converseResponse,
		promptText,
		latencyMs,
		time.Now(),
	)
	if err != nil {
		return nil, errors.NewBedrockError("Failed to process Turn1 response", "PROCESSING_ERROR", false)
	}

	// Create conversation history
	history := TurnHistory{
		TurnID:        turn1Response.TurnID,
		Timestamp:     time.Now(),
		Prompt:        turn1Response.Prompt,
		Response:      turn1Response.Response.Content,
		LatencyMs:     turn1Response.LatencyMs,
		TokenUsage:    turn1Response.TokenUsage,
		AnalysisStage: turn1Response.AnalysisStage,
	}

	// Create output
	output := &ExecuteTurn1Output{
		VerificationContext: input.VerificationContext,
		Turn1Response:      *turn1Response,
		ConversationState: ConversationState{
			CurrentTurn: 1,
			MaxTurns:    input.VerificationContext.TurnConfig.MaxTurns,
			History:     []TurnHistory{history},
		},
	}

	// Update verification context status
	output.VerificationContext.Status = schema.StatusTurn1Completed
	
	// Update turn timestamps
	now := time.Now().UTC().Format(time.RFC3339)
	if output.VerificationContext.TurnTimestamps == nil {
		output.VerificationContext.TurnTimestamps = &schema.TurnTimestamps{}
	}
	output.VerificationContext.TurnTimestamps.Turn1Completed = now

	return output, nil
}

// main is the entry point for the Lambda function
func main() {
	lambda.Start(Handler)
}
