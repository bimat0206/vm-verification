package handler

import (
	"context"
	"time"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// BedrockInvoker handles Bedrock API invocations
type BedrockInvoker struct {
	bedrock services.BedrockService
	cfg     config.Config
	log     logger.Logger
}

// NewBedrockInvoker creates a new instance of BedrockInvoker
func NewBedrockInvoker(bedrock services.BedrockService, cfg config.Config, log logger.Logger) *BedrockInvoker {
	return &BedrockInvoker{
		bedrock: bedrock,
		cfg:     cfg,
		log:     log,
	}
}

// InvokeResult contains the results of Bedrock invocation
type InvokeResult struct {
	Response *models.BedrockResponse
	Duration time.Duration
	Error    error
}

// InvokeBedrock handles the Bedrock API call with proper error handling
func (b *BedrockInvoker) InvokeBedrock(ctx context.Context, systemPrompt, turnPrompt, base64Img string, verificationID string) *InvokeResult {
	startTime := time.Now()
	result := &InvokeResult{}

	resp, err := b.bedrock.Converse(ctx, systemPrompt, turnPrompt, base64Img)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Error = b.handleBedrockError(err, turnPrompt, base64Img, verificationID)
		return result
	}

	result.Response = resp
	return result
}

// handleBedrockError processes and enriches Bedrock errors
func (b *BedrockInvoker) handleBedrockError(err error, turnPrompt, base64Img, verificationID string) error {
	contextLogger := b.log.WithCorrelationId(verificationID)

	if workflowErr, ok := err.(*errors.WorkflowError); ok {
		enrichedErr := workflowErr.WithContext("model_id", b.cfg.AWS.BedrockModel).
			WithContext("max_tokens", b.cfg.Processing.MaxTokens).
			WithContext("prompt_size", len(turnPrompt)).
			WithContext("image_size", len(base64Img))

		finalErr := errors.SetVerificationID(enrichedErr, verificationID)

		if workflowErr.Retryable {
			contextLogger.Warn("bedrock retryable error", map[string]interface{}{
				"error_code":    workflowErr.Code,
				"api_source":    string(workflowErr.APISource),
				"retry_attempt": "will_be_retried_by_step_functions",
			})
		} else {
			contextLogger.Error("bedrock non-retryable error", map[string]interface{}{
				"error_code": workflowErr.Code,
				"api_source": string(workflowErr.APISource),
				"severity":   string(workflowErr.Severity),
			})
		}

		return finalErr
	}

	// Handle unexpected errors
	wrappedErr := errors.WrapError(err, errors.ErrorTypeBedrock,
		"bedrock invocation failed", false).
		WithAPISource(errors.APISourceConverse)

	enrichedErr := errors.SetVerificationID(wrappedErr, verificationID)

	contextLogger.Error("bedrock unexpected error", map[string]interface{}{
		"original_error": err.Error(),
	})

	return enrichedErr
}

// GetInvocationMetadata returns metadata for tracking the Bedrock invocation
func (b *BedrockInvoker) GetInvocationMetadata(resp *models.BedrockResponse, duration time.Duration) map[string]interface{} {
	return map[string]interface{}{
		"model_id":           b.cfg.AWS.BedrockModel,
		"input_tokens":       resp.TokenUsage.InputTokens,
		"output_tokens":      resp.TokenUsage.OutputTokens,
		"thinking_tokens":    resp.TokenUsage.ThinkingTokens,
		"total_tokens":       resp.TokenUsage.TotalTokens,
		"bedrock_request_id": resp.RequestID,
		"latency_ms":         duration.Milliseconds(),
	}
}
