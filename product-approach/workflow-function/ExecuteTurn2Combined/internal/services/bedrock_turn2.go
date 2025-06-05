package services

import (
	"context"
	"strings"

	"workflow-function/ExecuteTurn2Combined/internal/bedrock"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// BedrockServiceTurn2 extends BedrockService with Turn2-specific functionality
type BedrockServiceTurn2 interface {
	BedrockService

	// ConverseWithHistory handles Turn2 conversation with history from Turn1
	ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error)
	// MODIFICATION END
}

// bedrockServiceTurn2 implements BedrockServiceTurn2
type bedrockServiceTurn2 struct {
	*bedrockService
	clientTurn2 *bedrock.ClientTurn2
}

// NewBedrockServiceTurn2 creates a new BedrockServiceTurn2 instance
func NewBedrockServiceTurn2(cfg config.Config, log logger.Logger) (BedrockServiceTurn2, error) {
	// Create base service
	baseService, err := NewBedrockService(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	// Create Turn2 client
	clientTurn2, err := bedrock.NewClientTurn2(cfg, log)
	if err != nil {
		return nil, err
	}

	return &bedrockServiceTurn2{
		bedrockService: baseService.(*bedrockService),
		clientTurn2:    clientTurn2,
	}, nil
}

// ConverseWithHistory handles Turn2 conversation with history from Turn1
func (s *bedrockServiceTurn2) ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error) {
	response, err := s.clientTurn2.ProcessTurn2(ctx, systemPrompt, turn2Prompt, base64Image, imageFormat, turn1Response)
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
			"Turn2 Bedrock conversation with history failed", maxRetries > 0).
			WithContext("system_prompt_size", len(systemPrompt)).
			WithContext("turn2_prompt_size", len(turn2Prompt)).
			WithContext("image_size", len(base64Image)).
			WithContext("image_format", imageFormat).
			WithContext("has_turn1_response", turn1Response != nil).
			WithComponent("BedrockClientTurn2").
			WithOperation("ProcessTurn2").
			WithCategory(category).
			WithRetryStrategy(retryStrategy).
			SetMaxRetries(maxRetries).
			WithSeverity(severity).
			WithSuggestions(
				"Check Bedrock service availability and quotas",
				"Verify model permissions and access policies",
				"Ensure prompt and image sizes are within limits",
				"Check for service throttling or rate limits",
				"Validate Turn1 response format compatibility",
			).
			WithRecoveryHints(
				"Retry with exponential backoff for transient errors",
				"Review and optimize prompt size if too large",
				"Check AWS service health dashboard",
				"Verify Bedrock model availability in region",
				"Ensure Turn1 response is properly formatted",
			)
		
		s.logger.Error("turn2_bedrock_conversation_failed", map[string]interface{}{
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
			"system_prompt_size": len(systemPrompt),
			"turn2_prompt_size":  len(turn2Prompt),
			"image_size":         len(base64Image),
			"image_format":       imageFormat,
			"has_turn1_response": turn1Response != nil,
			"suggestions":        wfErr.Suggestions,
			"recovery_hints":     wfErr.RecoveryHints,
		})
		return nil, wfErr
	}
	
	return response, nil
}
