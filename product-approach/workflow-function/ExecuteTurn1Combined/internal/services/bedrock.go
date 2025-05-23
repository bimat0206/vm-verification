// internal/services/bedrock.go - FIXED WITH INTELLIGENT ERROR HANDLING
package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/bedrock"
	
	// FIXED: Using shared errors package
	"workflow-function/shared/errors"
)

// BedrockService defines the Converse API integration.
type BedrockService interface {
	Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error)
}

type bedrockService struct {
	client         *bedrock.BedrockClient
	modelID        string
	maxTokens      int
	connectTimeout time.Duration
	callTimeout    time.Duration
}

// NewBedrockService constructs a BedrockService using the provided configuration.
func NewBedrockService(ctx context.Context, cfg config.Config) (BedrockService, error) {
	clientCfg := bedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		cfg.Processing.ThinkingType,
		cfg.Processing.BudgetTokens,
	)
	c, err := bedrock.NewBedrockClient(ctx, cfg.AWS.BedrockModel, clientCfg)
	if err != nil {
		// Enhanced error context for client creation failures
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock, 
			"failed to create Bedrock client", false). // false = non-retryable, this is a config issue
			WithContext("model_id", cfg.AWS.BedrockModel).
			WithContext("region", cfg.AWS.Region).
			WithContext("max_tokens", cfg.Processing.MaxTokens)
	}
	return &bedrockService{
		client:         c,
		modelID:        cfg.AWS.BedrockModel,
		maxTokens:      cfg.Processing.MaxTokens,
		connectTimeout: time.Duration(cfg.Processing.BedrockConnectTimeoutSec) * time.Second,
		callTimeout:    time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
	}, nil
}

// Converse performs the multimodal Converse call to Bedrock with intelligent error handling.
func (s *bedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
	// Apply call timeout to the context
	ctx, cancel := context.WithTimeout(ctx, s.callTimeout)
	defer cancel()

	// Build the request components
	img := bedrock.CreateImageContentFromBytes("jpeg", base64Image)
	userMsg := bedrock.CreateUserMessageWithContent(turnPrompt, []bedrock.ContentBlock{img})
	req := bedrock.CreateConverseRequest(s.modelID, []bedrock.MessageWrapper{userMsg}, systemPrompt, s.maxTokens, nil, nil)

	// Make the Bedrock API call
	resp, _, err := s.client.Converse(ctx, req)
	if err != nil {
		// FIXED: This is where we apply intelligent error classification
		// Instead of just wrapping with a generic stage, we analyze the error
		// to determine the appropriate retry behavior and error context
		return nil, s.classifyAndWrapBedrockError(err, systemPrompt, turnPrompt, base64Image)
	}

	// Process successful response
	raw, _ := json.Marshal(resp)
	usage := models.TokenUsage{
		InputTokens:    resp.Usage.InputTokens,
		OutputTokens:   resp.Usage.OutputTokens,
		ThinkingTokens: 0,
		TotalTokens:    resp.Usage.TotalTokens,
	}

	return &models.BedrockResponse{
		Raw:        raw,
		Processed:  resp,
		TokenUsage: usage,
		RequestID:  resp.RequestID,
	}, nil
}

// classifyAndWrapBedrockError analyzes Bedrock errors and applies appropriate
// error handling strategies based on the specific failure type.
// This demonstrates the power of the shared error system's contextual approach.
func (s *bedrockService) classifyAndWrapBedrockError(err error, systemPrompt, turnPrompt, base64Image string) error {
	errMsg := strings.ToLower(err.Error())
	
	// Base error context that we'll enrich based on the specific error type
	baseError := errors.WrapError(err, errors.ErrorTypeBedrock, 
		"Bedrock Converse API call failed", true). // Default to retryable
		WithAPISource(errors.APISourceConverse).
		WithContext("model_id", s.modelID).
		WithContext("max_tokens", s.maxTokens).
		WithContext("system_prompt_length", len(systemPrompt)).
		WithContext("turn_prompt_length", len(turnPrompt)).
		WithContext("image_data_length", len(base64Image))

	// Now we apply intelligent classification based on error patterns
	// This is much more sophisticated than the old stage-based approach
	
	if strings.Contains(errMsg, "throttl") || strings.Contains(errMsg, "rate") {
		// Throttling errors should be retried with backoff
		return baseError.
			WithContext("error_category", "throttling").
			WithContext("retry_strategy", "exponential_backoff").
			WithContext("severity", "low") // Not a serious problem, just need to wait
			
	} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") || strings.Contains(errMsg, "context deadline exceeded") {
		// Timeout errors should use BedrockTimeout error type for proper Step Functions handling
		return errors.NewBedrockError(
			"Bedrock call exceeded timeout limit",
			"BedrockTimeout",
			false, // non-retryable as per requirements
		).WithAPISource(errors.APISourceConverse).
			WithContext("model_id", s.modelID).
			WithContext("timeout_seconds", s.callTimeout.Seconds()).
			WithContext("error_category", "timeout").
			WithContext("original_error", err.Error())
			
	} else if strings.Contains(errMsg, "content") || strings.Contains(errMsg, "policy") || strings.Contains(errMsg, "safety") {
		// Content policy violations should not be retried - they're deterministic failures
		return errors.WrapError(err, errors.ErrorTypeBedrock, 
			"content policy violation", false). // false = non-retryable
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "content_policy").
			WithContext("model_id", s.modelID).
			WithContext("system_prompt_length", len(systemPrompt)).
			WithContext("turn_prompt_length", len(turnPrompt)).
			WithContext("severity", "high") // This indicates a problem with our prompt design
			
	} else if strings.Contains(errMsg, "token") && (strings.Contains(errMsg, "limit") || strings.Contains(errMsg, "exceeded")) {
		// Token limit errors are non-retryable without changing the request
		return errors.WrapError(err, errors.ErrorTypeBedrock, 
			"token limit exceeded", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "token_limit").
			WithContext("model_id", s.modelID).
			WithContext("configured_max_tokens", s.maxTokens).
			WithContext("system_prompt_length", len(systemPrompt)).
			WithContext("turn_prompt_length", len(turnPrompt)).
			WithContext("severity", "high") // This indicates our prompt is too long
			
	} else if strings.Contains(errMsg, "model") && (strings.Contains(errMsg, "unavailable") || strings.Contains(errMsg, "not found")) {
		// Model availability issues are infrastructure problems, potentially retryable
		return baseError.
			WithContext("error_category", "model_availability").
			WithContext("retry_strategy", "limited_retry"). // Don't retry indefinitely
			WithContext("severity", "critical") // This could indicate a service outage
			
	} else if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "authorization") || strings.Contains(errMsg, "permission") {
		// Auth errors are non-retryable configuration issues
		return errors.WrapError(err, errors.ErrorTypeBedrock, 
			"authentication/authorization failure", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "authorization").
			WithContext("model_id", s.modelID).
			WithContext("severity", "critical") // This indicates a serious config problem
			
	} else {
		// Unknown error types - be conservative but provide rich context for debugging
		return baseError.
			WithContext("error_category", "unknown").
			WithContext("retry_strategy", "cautious").
			WithContext("severity", "medium").
			WithContext("original_error", err.Error()) // Preserve the full original error for analysis
	}
}