package services

import (
	"context"
	"time"

	localBedrock "workflow-function/ExecuteTurn2Combined/internal/bedrock"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	sharedBedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// BedrockService defines the interface for AI model integration
type BedrockService interface {
	Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error)
}

// bedrockService implements AI model integration
type bedrockService struct {
	client *localBedrock.Client
	config *config.Config
	logger logger.Logger
}

// NewBedrockService creates a new BedrockService
func NewBedrockService(ctx context.Context, cfg config.Config) (BedrockService, error) {
	// Initialize structured logger for comprehensive observability
	log := logger.New("ExecuteTurn2Combined", "BedrockService")

	// Create shared bedrock client first
	clientConfig := sharedBedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		"",
		0,
	)

	sharedClient, err := sharedBedrock.NewBedrockClient(ctx, cfg.AWS.BedrockModel, clientConfig)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to initialize shared bedrock client", false).
			WithContext("model_id", cfg.AWS.BedrockModel).
			WithContext("region", cfg.AWS.Region)
	}

	// Create local bedrock configuration
	localConfig := &localBedrock.Config{
		ModelID:          cfg.AWS.BedrockModel,
		AnthropicVersion: cfg.AWS.AnthropicVersion,
		MaxTokens:        cfg.Processing.MaxTokens,
		Temperature:      cfg.Processing.Temperature,
		ThinkingType:     "",
		ThinkingBudget:   0,
		Timeout:          time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
		Region:           cfg.AWS.Region,
	}

	// Initialize local bedrock client with adapter pattern
	localClient := localBedrock.NewClient(sharedClient, localConfig, log)

	log.Info("bedrock_service_initialized", map[string]interface{}{
		"model_id":   cfg.AWS.BedrockModel,
		"region":     cfg.AWS.Region,
		"max_tokens": cfg.Processing.MaxTokens,
	})

	return &bedrockService{
		client: localClient,
		config: &cfg,
		logger: log,
	}, nil
}

// NewBedrockServiceWithLocalClient creates service with local control
// This constructor is used when the local client is already initialized
func NewBedrockServiceWithLocalClient(client *localBedrock.Client, cfg config.Config, logger logger.Logger) BedrockService {
	return &bedrockService{
		client: client,
		config: &cfg,
		logger: logger,
	}
}

// Converse implements multimodal AI inference
func (s *bedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
	operationStart := time.Now()

	s.logger.Info("bedrock_converse_initiated", map[string]interface{}{
		"system_prompt_size": len(systemPrompt),
		"turn_prompt_size":   len(turnPrompt),
		"base64_image_size":  len(base64Image),
	})

	response, err := s.client.ProcessTurn1(ctx, systemPrompt, turnPrompt, base64Image)
	if err != nil {
		s.logger.Error("bedrock_api_error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Convert to models.BedrockResponse
	bedrockResponse := &models.BedrockResponse{
		Raw:        response.Raw,
		Processed:  map[string]interface{}{"content": response.Content},
		TokenUsage: response.TokenUsage,
		RequestID:  response.RequestID,
	}

	totalDuration := time.Since(operationStart)

	s.logger.Info("bedrock_converse_completed", map[string]interface{}{
		"total_duration_ms": totalDuration.Milliseconds(),
		"token_usage":       bedrockResponse.TokenUsage,
		"request_id":        "", // RequestID not available in schema.BedrockResponse
		"success":           true,
	})

	return bedrockResponse, nil
}
