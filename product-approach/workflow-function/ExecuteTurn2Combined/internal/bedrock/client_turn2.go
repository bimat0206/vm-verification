package bedrock

import (
	"context"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	sharedBedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// ClientTurn2 extends Client with Turn2-specific functionality
type ClientTurn2 struct {
	*Client
	adapterTurn2 *AdapterTurn2
}

// NewClientTurn2 creates a new ClientTurn2 instance
func NewClientTurn2(cfg config.Config, log logger.Logger) (*ClientTurn2, error) {
	// Create shared Bedrock client with proper configuration
	clientConfig := sharedBedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		cfg.Processing.ThinkingType,
		cfg.Processing.BudgetTokens,
	)

	sharedClient, err := sharedBedrock.NewBedrockClient(
		context.Background(),
		cfg.AWS.BedrockModel,
		clientConfig,
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to create shared Bedrock client", false).
			WithContext("region", cfg.AWS.Region).
			WithContext("model", cfg.AWS.BedrockModel)
	}

	// Create bedrock Config for the base client
	bedrockConfig := &Config{
		ModelID:          cfg.AWS.BedrockModel,
		AnthropicVersion: cfg.AWS.AnthropicVersion,
		MaxTokens:        cfg.Processing.MaxTokens,
		Temperature:      cfg.Processing.Temperature,
		ThinkingType:     cfg.Processing.ThinkingType,
		ThinkingBudget:   cfg.Processing.BudgetTokens,
		Timeout:          time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
		Region:           cfg.AWS.Region,
	}

	// Create base client with correct parameters
	baseClient := NewClient(sharedClient, bedrockConfig, log)

	// Create Turn2 adapter with correct parameter order
	adapterTurn2 := NewAdapterTurn2(sharedClient, &cfg, log)

	return &ClientTurn2{
		Client:       baseClient,
		adapterTurn2: adapterTurn2,
	}, nil
}

// ProcessTurn2 handles the complete Turn2 processing
// MODIFICATION START: added imageFormat parameter
func (c *ClientTurn2) ProcessTurn2(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error) {
	startTime := time.Now()

	// Validate configuration
	if err := c.ValidateConfiguration(); err != nil {
		return nil, err
	}

	// Process Turn2 using adapter
	response, err := c.adapterTurn2.ProcessTurn2(ctx, systemPrompt, turn2Prompt, base64Image, imageFormat, turn1Response)
	if err != nil {
		return nil, err
	}

	// Add processing time to response
	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	return response, nil
}
