package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go"

	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// executeBedrockRequest builds the request and calls Bedrock with retry logic
func (h *Handler) executeBedrockRequest(
	ctx context.Context,
	state *schema.WorkflowState,
	retriever *schema.HybridBase64Retriever,
	log logger.Logger,
) (*schema.TurnResponse, error) {
	// Build Bedrock request
	modelInput, err := h.buildBedrockInput(ctx, state, retriever, log)
	if err != nil {
		wfErr := wferrors.NewBedrockError("Failed to build Bedrock input", "BEDROCK_BUILD_INPUT", false).
			WithContext("detail", err.Error())
		return nil, h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	// Execute Bedrock call with retry logic
	response, latencyMs, err := h.callBedrockWithRetry(ctx, modelInput, log)
	if err != nil {
		return nil, h.createAndLogError(state, err, log, schema.StatusBedrockProcessingFailed)
	}

	// Parse response and build TurnResponse
	turn1Response, err := h.parseBedrockResponse(state, response, latencyMs, aws.ToString(modelInput.ModelId), log)
	if err != nil {
		wfErr := wferrors.NewParsingError("Bedrock response parsing", err)
		return nil, h.createAndLogError(state, wfErr, log, schema.StatusBedrockProcessingFailed)
	}

	return turn1Response, nil
}

// buildBedrockInput prepares the Bedrock payload according to InvokeModel API spec
func (h *Handler) buildBedrockInput(
	ctx context.Context,
	state *schema.WorkflowState,
	retriever *schema.HybridBase64Retriever,
	log logger.Logger,
) (*bedrockruntime.InvokeModelInput, error) {
	log.Debug("Building Bedrock input payload", nil)

	if len(state.CurrentPrompt.Messages) == 0 {
		return nil, fmt.Errorf("no messages in CurrentPrompt")
	}

	msg := state.CurrentPrompt.Messages[0]

	// Build content array: text then image
	var content []interface{}

	// Add text content
	if len(msg.Content) > 0 && msg.Content[0].Text != "" {
		content = append(content, msg.Content[0].Text)
	}

	// Add image content if present
	if len(msg.Content) > 1 && msg.Content[1].Image != nil {
		imageContent, err := h.buildImageContent(state.Images, retriever, log)
		if err != nil {
			return nil, fmt.Errorf("failed to build image content: %w", err)
		}
		content = append(content, imageContent)
	}

	// Assemble payload
	payload := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": msg.Role, "content": content},
		},
		"max_tokens":        state.BedrockConfig.MaxTokens,
		"anthropic_version": state.BedrockConfig.AnthropicVersion,
	}

	// Add optional parameters if present
	if state.BedrockConfig.Temperature > 0 {
		payload["temperature"] = state.BedrockConfig.Temperature
	}

	// Marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	log.Debug("Bedrock input payload built successfully", map[string]interface{}{
		"payloadSize": len(bodyBytes),
		"hasImage":    len(content) > 1,
	})

	// Use default model ID from environment since BedrockConfig does not have a ModelId field
	// Get model ID from configuration if available, or use a default
	modelId := h.getModelId(state)
	
	return &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	}, nil
}

// callBedrockWithRetry executes the Bedrock API call with exponential backoff retry
func (h *Handler) callBedrockWithRetry(
	ctx context.Context,
	input *bedrockruntime.InvokeModelInput,
	log logger.Logger,
) (*bedrockruntime.InvokeModelOutput, int64, error) {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		log.Info("Invoking Bedrock API", map[string]interface{}{
			"modelId": aws.ToString(input.ModelId),
			"attempt": attempt,
		})

		start := time.Now()
		response, err := h.bedrockClient.InvokeModel(ctx, input)
		latencyMs := time.Since(start).Milliseconds()

		if err == nil {
			log.Info("Bedrock API call successful", map[string]interface{}{
				"latencyMs": latencyMs,
				"attempt":   attempt,
			})
			return response, latencyMs, nil
		}

		lastErr = err
		
		// Check if error is retryable
		if !h.isRetryableError(err) || attempt == maxRetryAttempts {
			break
		}

		// Calculate retry delay with exponential backoff
		delay := time.Duration(attempt*attempt) * baseRetryDelay
		log.Warn("Bedrock API call failed, retrying", map[string]interface{}{
			"attempt":    attempt,
			"error":      err.Error(),
			"retryDelay": delay.String(),
		})

		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		}
	}

	// All retries exhausted
	log.Error("Bedrock API call failed after all retries", map[string]interface{}{
		"attempts": maxRetryAttempts,
		"error":    lastErr.Error(),
	})

	return nil, 0, wferrors.NewBedrockError("Bedrock API failure after retries", "BEDROCK_API_FAILURE", true).
		WithContext("bedrockError", lastErr.Error()).
		WithContext("attempts", maxRetryAttempts)
}

// isRetryableError determines if an error is worth retrying
func (h *Handler) isRetryableError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		// Retry on service exceptions and throttling
		return code == "ServiceException" || code == "ThrottlingException"
	}
	return false
}

// getModelId gets the model ID for Bedrock API call
// Since schema.BedrockConfig doesn't have a ModelId field, we need to get it from elsewhere
func (h *Handler) getModelId(state *schema.WorkflowState) string {
	// First, check if we have a modelId in our handler configuration that came from environment
	if h.modelId != "" {
		return h.modelId
	}
	
	// Default to widely used model ID for Claude 3 Sonnet if not set in handler
	defaultModelId := "anthropic.claude-3-sonnet-20240229-v1:0"
	
	// Log that we're using a default model ID
	h.logger.Warn("Using default model ID because BedrockConfig.ModelId is undefined and handler modelId is not set", map[string]interface{}{
		"defaultModelId": defaultModelId,
	})
	
	return defaultModelId
}

// parseBedrockResponse parses the Bedrock response and constructs a TurnResponse
func (h *Handler) parseBedrockResponse(
	state *schema.WorkflowState,
	output *bedrockruntime.InvokeModelOutput,
	latencyMs int64,
	modelId string,
	log logger.Logger,
) (*schema.TurnResponse, error) {
	// Create ResponseProcessor with thinking enabled if configured
	processor := NewResponseProcessor(log, true, 16000) // Enable thinking with 16k budget
	
	// Get prompt text
	promptText := ""
	if len(state.CurrentPrompt.Messages) > 0 && len(state.CurrentPrompt.Messages[0].Content) > 0 {
		promptText = state.CurrentPrompt.Messages[0].Content[0].Text
	}
	
	// Get image URLs for metadata
	imageUrls := make(map[string]string)
	if imageInfo := h.getReferenceImageInfo(state.Images); imageInfo != nil {
		imageUrls["reference"] = imageInfo.URL
	}
	
	// Use ResponseProcessor for advanced parsing
	turnResponse, err := processor.ProcessBedrockResponse(
		output.Body,
		latencyMs,
		modelId,
		promptText,
		imageUrls,
		1, // Turn 1
	)
	if err != nil {
		return nil, err
	}
	
	// Validate the response structure
	if err := processor.ValidateResponseStructure(turnResponse); err != nil {
		log.Warn("Response structure validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue processing despite validation warnings
	}
	
	return turnResponse, nil
}