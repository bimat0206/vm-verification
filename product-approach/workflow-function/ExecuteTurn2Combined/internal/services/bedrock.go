package services

import (
	"context"
	"strings"
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
		TopP:             cfg.Processing.TopP,
		ThinkingType:     "",
		ThinkingBudget:   0,
		Timeout:          time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
		Region:           cfg.AWS.Region,
	}

	// Initialize local bedrock client with adapter pattern
	localClient := localBedrock.NewClient(sharedClient, localConfig, log)

	log.Info("bedrock_service_initialized", map[string]interface{}{
		"model_id":    cfg.AWS.BedrockModel,
		"region":      cfg.AWS.Region,
		"max_tokens":  cfg.Processing.MaxTokens,
		"temperature": cfg.Processing.Temperature,
		"top_p":       cfg.Processing.TopP,
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
		// Determine error category and retry strategy based on error type
		category := errors.CategoryServer
		retryStrategy := errors.RetryExponential
		severity := errors.ErrorSeverityHigh
		maxRetries := 3
		
		// Check for specific Bedrock error patterns
		errorStr := err.Error()
		if strings.Contains(errorStr, "throttling") || strings.Contains(errorStr, "rate limit") {
			category = errors.CategoryCapacity
			retryStrategy = errors.RetryJittered
			severity = errors.ErrorSeverityMedium
			maxRetries = 5
		} else if strings.Contains(errorStr, "validation") || strings.Contains(errorStr, "invalid") {
			category = errors.CategoryClient
			retryStrategy = errors.RetryNone
			severity = errors.ErrorSeverityCritical
			maxRetries = 0
		} else if strings.Contains(errorStr, "timeout") {
			category = errors.CategoryNetwork
			retryStrategy = errors.RetryLinear
			severity = errors.ErrorSeverityHigh
			maxRetries = 2
		}
		
		wfErr := errors.WrapError(err, errors.ErrorTypeBedrock,
			"Bedrock API invocation failed", maxRetries > 0).
			WithContext("model_id", s.config.AWS.BedrockModel).
			WithContext("system_prompt_size", len(systemPrompt)).
			WithContext("turn_prompt_size", len(turnPrompt)).
			WithContext("image_size", len(base64Image)).
			WithComponent("BedrockClient").
			WithOperation("ProcessTurn1").
			WithCategory(category).
			WithRetryStrategy(retryStrategy).
			SetMaxRetries(maxRetries).
			WithSeverity(severity).
			WithSuggestions(
				"Check Bedrock service availability and quotas",
				"Verify model permissions and access policies",
				"Ensure prompt and image sizes are within limits",
				"Check for service throttling or rate limits",
			).
			WithRecoveryHints(
				"Retry with exponential backoff for transient errors",
				"Review and optimize prompt size if too large",
				"Check AWS service health dashboard",
				"Verify Bedrock model availability in region",
			)
		
		s.logger.Error("bedrock_api_error", map[string]interface{}{
			"error_type":         string(wfErr.Type),
			"error_code":         wfErr.Code,
			"message":            wfErr.Message,
			"retryable":          wfErr.Retryable,
			"severity":           string(wfErr.Severity),
			"category":           string(wfErr.Category),
			"retry_strategy":     string(wfErr.RetryStrategy),
			"max_retries":        wfErr.MaxRetries,
			"component":          wfErr.Component,
			"operation":          wfErr.Operation,
			"model_id":           s.config.AWS.BedrockModel,
			"system_prompt_size": len(systemPrompt),
			"turn_prompt_size":   len(turnPrompt),
			"image_size":         len(base64Image),
			"suggestions":        wfErr.Suggestions,
			"recovery_hints":     wfErr.RecoveryHints,
		})
		return nil, wfErr
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
