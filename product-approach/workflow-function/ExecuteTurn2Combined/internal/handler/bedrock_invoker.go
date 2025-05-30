package handler

import (
	"context"
	"encoding/json"
	"time"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// BedrockInvoker handles Bedrock API invocations
type BedrockInvoker struct {
	bedrock services.BedrockServiceTurn2
	cfg     config.Config
	log     logger.Logger
}

// NewBedrockInvoker creates a new instance of BedrockInvoker
func NewBedrockInvoker(bedrock services.BedrockServiceTurn2, cfg config.Config, log logger.Logger) *BedrockInvoker {
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

	// Use a fixed image format - jpeg is the most common format
	imageFormat := "jpeg" // Default format for all images

	// Use ConverseWithHistory instead of Converse, passing nil for Turn1Response
	// This aligns with v2.1.2 changes that removed Turn1 dependencies
	schemaResp, err := b.bedrock.ConverseWithHistory(ctx, systemPrompt, turnPrompt, base64Img, imageFormat, nil)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Error = b.handleBedrockError(err, turnPrompt, base64Img, verificationID)
		return result
	}

	// Convert schema.BedrockResponse to models.BedrockResponse
	// First, create the raw JSON representation of the schema response
	rawJSON, _ := json.Marshal(schemaResp)

	// Create the models.BedrockResponse with the correct fields
	resp := &models.BedrockResponse{
		Raw: rawJSON,
		TokenUsage: schema.TokenUsage{
			InputTokens:    schemaResp.InputTokens,
			OutputTokens:   schemaResp.OutputTokens,
			ThinkingTokens: 0, // Not used in schema response
			TotalTokens:    schemaResp.InputTokens + schemaResp.OutputTokens,
		},
		RequestID: schemaResp.ModelId, // Use ModelId as RequestID if needed
	}

	// Set the processed field to the content from the schema response
	resp.Processed = schemaResp.Content

	result.Response = resp
	return result
}

// handleBedrockError processes and enriches Bedrock errors
func (b *BedrockInvoker) handleBedrockError(err error, turnPrompt, base64Img, verificationID string) error {
	contextLogger := b.log.WithCorrelationId(verificationID)

	// Use a fixed image format - jpeg is the most common format
	imageFormat := "jpeg" // Default format for all images

	if workflowErr, ok := err.(*errors.WorkflowError); ok {
		enrichedErr := workflowErr.WithContext("model_id", b.cfg.AWS.BedrockModel).
			WithContext("max_tokens", b.cfg.Processing.MaxTokens).
			WithContext("prompt_size", len(turnPrompt)).
			WithContext("image_size", len(base64Img)).
			WithContext("image_format", imageFormat).
			WithContext("operation", "bedrock_converse_with_history")

		finalErr := errors.SetVerificationID(enrichedErr, verificationID)

		if workflowErr.Retryable {
			contextLogger.Warn("bedrock_retryable_error", map[string]interface{}{
				"error_code":    workflowErr.Code,
				"api_source":    string(workflowErr.APISource),
				"retry_attempt": "will_be_retried_by_step_functions",
				"image_format":  imageFormat,
				"operation":     "bedrock_converse_with_history",
			})
		} else {
			contextLogger.Error("bedrock_non_retryable_error", map[string]interface{}{
				"error_code":    workflowErr.Code,
				"api_source":    string(workflowErr.APISource),
				"severity":      string(workflowErr.Severity),
				"image_format":  imageFormat,
				"operation":     "bedrock_converse_with_history",
			})
		}

		return finalErr
	}

	// Handle unexpected errors
	wrappedErr := errors.WrapError(err, errors.ErrorTypeBedrock,
		"bedrock invocation failed", true). // Mark as retryable
		WithAPISource(errors.APISourceConverse).
		WithContext("image_format", imageFormat).
		WithContext("operation", "bedrock_converse_with_history").
		WithContext("model_id", b.cfg.AWS.BedrockModel)

	enrichedErr := errors.SetVerificationID(wrappedErr, verificationID)

	contextLogger.Error("bedrock_unexpected_error", map[string]interface{}{
		"original_error": err.Error(),
		"image_format":   imageFormat,
		"operation":      "bedrock_converse_with_history",
		"model_id":       b.cfg.AWS.BedrockModel,
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
