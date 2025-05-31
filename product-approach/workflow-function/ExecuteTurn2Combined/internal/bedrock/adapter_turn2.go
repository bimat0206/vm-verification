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

// AdapterTurn2 provides Turn2-specific Bedrock functionality
type AdapterTurn2 struct {
	cfg    *config.Config
	log    logger.Logger
	client *sharedBedrock.BedrockClient
}

// NewAdapterTurn2 creates a new AdapterTurn2 instance
func NewAdapterTurn2(client *sharedBedrock.BedrockClient, cfg *config.Config, log logger.Logger) *AdapterTurn2 {
	return &AdapterTurn2{
		client: client,
		cfg:    cfg,
		log:    log,
	}
}

// ConverseWithHistory handles Turn2 conversation with history from Turn1
func (a *AdapterTurn2) ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error) {
	startTime := time.Now()

	// Validate inputs with detailed error context
	if systemPrompt == "" {
		a.log.Error("bedrock_validation_error", map[string]interface{}{
			"error":     "system prompt cannot be empty",
			"operation": "bedrock_converse_with_history",
		})
		return nil, errors.NewValidationError(
			"system prompt cannot be empty",
			map[string]interface{}{
				"operation": "bedrock_converse_with_history",
				"component": "adapter_turn2",
			})
	}

	if turn2Prompt == "" {
		a.log.Error("bedrock_validation_error", map[string]interface{}{
			"error":     "turn2 prompt cannot be empty",
			"operation": "bedrock_converse_with_history",
		})
		return nil, errors.NewValidationError(
			"turn2 prompt cannot be empty",
			map[string]interface{}{
				"operation": "bedrock_converse_with_history",
				"component": "adapter_turn2",
			})
	}

	if base64Image == "" {
		a.log.Error("bedrock_validation_error", map[string]interface{}{
			"error":     "base64 image cannot be empty",
			"operation": "bedrock_converse_with_history",
		})
		return nil, errors.NewValidationError(
			"base64 image cannot be empty",
			map[string]interface{}{
				"operation": "bedrock_converse_with_history",
				"component": "adapter_turn2",
			})
	}

	if imageFormat == "" {
		a.log.Error("bedrock_validation_error", map[string]interface{}{
			"error":     "image format cannot be empty",
			"operation": "bedrock_converse_with_history",
		})
		return nil, errors.NewValidationError(
			"image format cannot be empty",
			map[string]interface{}{
				"operation": "bedrock_converse_with_history",
				"component": "adapter_turn2",
			})
	}

	// Normalize image format dynamically
	format := sharedBedrock.NormalizeImageFormat(imageFormat)
	a.log.Debug("image_format_normalized", map[string]interface{}{
		"original_format":   imageFormat,
		"normalized_format": format,
		"operation":         "bedrock_converse_with_history",
	})

	// Build request with conversation history using the correct types
	// Start with the system prompt as the first message
	messages := []sharedBedrock.MessageWrapper{
		{
			Role:    "system",
			Content: []sharedBedrock.ContentBlock{{Type: "text", Text: systemPrompt}},
		},
	}

	// If Turn1 response is available, include the original user prompt and assistant answer
	if turn1Response != nil {
		a.log.Debug("turn1_response_present", map[string]interface{}{
			"operation": "bedrock_converse_with_history",
		})

		turn1 := &sharedBedrock.Turn1Response{
			TurnID:    turn1Response.TurnId,
			Timestamp: turn1Response.Timestamp,
			Prompt:    turn1Response.Prompt,
			Response: sharedBedrock.TextResponse{
				Content:    turn1Response.Response.Content,
				StopReason: turn1Response.Response.StopReason,
			},
		}
		messages = append(messages, sharedBedrock.CreateTurn2ConversationHistory(turn1)...)
	} else {
		a.log.Debug("turn1_response_nil", map[string]interface{}{
			"operation": "bedrock_converse_with_history",
			"message":   "missing turn1 context",
		})
	}

	// Add final user message with Turn 2 prompt and image
	messages = append(messages, sharedBedrock.MessageWrapper{
		Role: "user",
		Content: []sharedBedrock.ContentBlock{
			{
				Type: "text",
				Text: "[Turn 2] " + turn2Prompt,
			},
			{
				Type: "image",
				// MODIFICATION START: use dynamic image format
				Image: &sharedBedrock.ImageBlock{
					Format: format,
					Source: sharedBedrock.ImageSource{
						Type:  "bytes",
						Bytes: base64Image,
					},
				},
				// MODIFICATION END
			},
		},
	})

	request := &sharedBedrock.ConverseRequest{
		ModelId:  a.cfg.AWS.BedrockModel,
		System:   "",
		Messages: messages,
		InferenceConfig: sharedBedrock.InferenceConfig{
			MaxTokens:   a.cfg.Processing.MaxTokens,
			Temperature: &[]float64{0.7}[0], // Default temperature
		},
	}

	// Validate request before sending to Bedrock
	if err := sharedBedrock.ValidateConverseRequest(request); err != nil {
		a.log.Error("bedrock_request_validation_failed", map[string]interface{}{
			"error":              err.Error(),
			"model_id":           a.cfg.AWS.BedrockModel,
			"system_prompt_size": len(systemPrompt),
			"turn2_prompt_size":  len(turn2Prompt),
			"image_size":         len(base64Image),
			"image_format":       format,
			"message_count":      len(request.Messages),
		})

		return nil, errors.WrapError(err, errors.ErrorTypeValidation,
			"Bedrock request validation failed", false).
			WithContext("model_id", a.cfg.AWS.BedrockModel).
			WithContext("operation", "bedrock_request_validation").
			WithContext("image_format", format)
	}

	// Log request details
	a.log.Info("bedrock_turn2_request_prepared", map[string]interface{}{
		"model_id":           a.cfg.AWS.BedrockModel,
		"max_tokens":         a.cfg.Processing.MaxTokens,
		"system_prompt_size": len(systemPrompt),
		"turn2_prompt_size":  len(turn2Prompt),
		"image_size":         len(base64Image),
		"image_format":       format,
		"message_count":      len(request.Messages),
	})

	// Invoke Bedrock (note: returns 3 values)
	a.log.Info("bedrock_converse_api_call_start", map[string]interface{}{
		"model_id":           a.cfg.AWS.BedrockModel,
		"system_prompt_size": len(systemPrompt),
		"turn2_prompt_size":  len(turn2Prompt),
		"image_size":         len(base64Image),
		"image_format":       format,
		"message_count":      len(request.Messages),
		"max_tokens":         a.cfg.Processing.MaxTokens,
		"operation":          "bedrock_converse_with_history",
	})

	response, latencyMs, err := a.client.Converse(ctx, request)

	if err != nil {
		// Enhanced error logging for debugging
		a.log.Error("bedrock_converse_api_error", map[string]interface{}{
			"error":              err.Error(),
			"model_id":           a.cfg.AWS.BedrockModel,
			"system_prompt_size": len(systemPrompt),
			"turn2_prompt_size":  len(turn2Prompt),
			"image_size":         len(base64Image),
			"image_format":       format,
			"message_count":      len(request.Messages),
			"max_tokens":         a.cfg.Processing.MaxTokens,
			"operation":          "bedrock_converse_with_history",
		})

		// Check for context cancellation or deadline exceeded
		if ctx.Err() != nil {
			a.log.Error("bedrock_context_error", map[string]interface{}{
				"context_error": ctx.Err().Error(),
				"operation":     "bedrock_converse_with_history",
			})

			return nil, errors.WrapError(ctx.Err(), errors.ErrorTypeBedrock,
				"Bedrock API call context error: "+ctx.Err().Error(), true).
				WithContext("model_id", a.cfg.AWS.BedrockModel).
				WithContext("operation", "bedrock_converse_with_history").
				WithContext("image_format", format).
				WithContext("message_count", len(request.Messages))
		}

		// Mark as retryable for most Bedrock errors
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to invoke Bedrock for Turn2", true).
			WithContext("model_id", a.cfg.AWS.BedrockModel).
			WithContext("operation", "bedrock_converse_with_history").
			WithContext("image_format", format).
			WithContext("message_count", len(request.Messages)).
			WithContext("max_tokens", a.cfg.Processing.MaxTokens)
	}

	a.log.Info("bedrock_converse_api_call_success", map[string]interface{}{
		"model_id":   a.cfg.AWS.BedrockModel,
		"latency_ms": latencyMs,
		"operation":  "bedrock_converse_with_history",
	})

	// Extract text content from response
	textContent := sharedBedrock.ExtractTextFromResponse(response)

	// Translate to schema.BedrockResponse
	schemaResponse := &schema.BedrockResponse{
		Content:          textContent,
		CompletionReason: response.StopReason,
		InputTokens:      response.Usage.InputTokens,
		OutputTokens:     response.Usage.OutputTokens,
		LatencyMs:        latencyMs,
		ModelId:          response.ModelID,
		Timestamp:        time.Now().Format(time.RFC3339),
		Turn:             2,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
	}

	// Log response details
	a.log.Info("bedrock_turn2_response_received", map[string]interface{}{
		"model_id":           response.ModelID,
		"completion_reason":  response.StopReason,
		"input_tokens":       response.Usage.InputTokens,
		"output_tokens":      response.Usage.OutputTokens,
		"latency_ms":         latencyMs,
		"content_length":     len(textContent),
		"processing_time_ms": schemaResponse.ProcessingTimeMs,
	})

	return schemaResponse, nil
}

// ProcessTurn2 handles the complete Turn2 processing
func (a *AdapterTurn2) ProcessTurn2(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error) {
	// Apply operational timeout using Processing config
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.cfg.Processing.BedrockCallTimeoutSec)*time.Second)
	defer cancel()

	// Invoke Bedrock with conversation history
	response, err := a.ConverseWithHistory(timeoutCtx, systemPrompt, turn2Prompt, base64Image, imageFormat, turn1Response)
	if err != nil {
		return nil, err
	}

	// Enrich response with metadata using available config
	response.ModelConfig = &schema.ModelConfig{
		ModelId:     a.cfg.AWS.BedrockModel,
		Temperature: 0.7, // Default temperature
		TopP:        0.9, // Default TopP
		MaxTokens:   a.cfg.Processing.MaxTokens,
	}

	return response, nil
}
