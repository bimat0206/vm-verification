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

	// Build request with conversation history using the correct types.
	// The system prompt is provided via the ConverseRequest.System field,
	// so the messages slice must only contain user and assistant roles.
	messages := []sharedBedrock.MessageWrapper{}

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
		System:   systemPrompt,
		Messages: messages,
		InferenceConfig: sharedBedrock.InferenceConfig{
			MaxTokens:   a.cfg.Processing.MaxTokens,
			Temperature: &a.cfg.Processing.Temperature,
		},
	}

	// Add thinking/reasoning configuration if enabled
	if a.cfg.IsThinkingEnabled() {
		request.Reasoning = "enabled"
		request.InferenceConfig.Reasoning = "enabled"
		request.Thinking = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": a.cfg.Processing.BudgetTokens,
		}
		a.log.Info("THINKING_ADAPTER_ENABLED", map[string]interface{}{
			"thinking_type":       a.cfg.Processing.ThinkingType,
			"budget_tokens":       a.cfg.Processing.BudgetTokens,
			"request_reasoning":   request.Reasoning,
			"inference_reasoning": request.InferenceConfig.Reasoning,
			"request_thinking":    request.Thinking,
		})
	} else {
		a.log.Info("THINKING_ADAPTER_DISABLED", map[string]interface{}{
			"thinking_type":       a.cfg.Processing.ThinkingType,
			"request_reasoning":   request.Reasoning,
			"inference_reasoning": request.InferenceConfig.Reasoning,
		})
	}

	// Log the complete request structure for debugging
	a.log.Info("THINKING_REQUEST_STRUCTURE", map[string]interface{}{
		"model_id":            request.ModelId,
		"reasoning_field":     request.Reasoning,
		"inference_reasoning": request.InferenceConfig.Reasoning,
		"thinking_field":      request.Thinking,
		"max_tokens":          request.InferenceConfig.MaxTokens,
	})

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

	// Extract text and thinking content from response
	textContent := sharedBedrock.ExtractTextFromResponse(response)
	thinkingText := sharedBedrock.ExtractThinkingFromResponse(response)

	// Map token usage defensively
	tokenUsage := schema.TokenUsage{
		InputTokens:    0,
		OutputTokens:   0,
		ThinkingTokens: 0,
		TotalTokens:    0,
	}
	if response.Usage != nil {
		tokenUsage.InputTokens = response.Usage.InputTokens
		tokenUsage.OutputTokens = response.Usage.OutputTokens
		tokenUsage.ThinkingTokens = response.Usage.ThinkingTokens
		tokenUsage.TotalTokens = response.Usage.TotalTokens
	}

	// Translate to schema.BedrockResponse
	schemaResponse := &schema.BedrockResponse{
		Content:          textContent,
		Thinking:         thinkingText,
		CompletionReason: response.StopReason,
		InputTokens:      tokenUsage.InputTokens,
		OutputTokens:     tokenUsage.OutputTokens,
		LatencyMs:        latencyMs,
		ModelId:          response.ModelID,
		Timestamp:        time.Now().Format(time.RFC3339),
		Turn:             2,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		TokenUsage:       &tokenUsage,
		Metadata:         map[string]interface{}{},
	}

	// Populate thinking metadata
	thinkingEnabled := a.cfg.IsThinkingEnabled()
	schemaResponse.Metadata["thinking_enabled"] = thinkingEnabled
	if thinkingText != "" {
		blocks := a.generateThinkingBlocks(thinkingText, "response-processing")
		schemaResponse.Metadata["thinking_blocks"] = blocks
		schemaResponse.Metadata["has_thinking"] = true
		schemaResponse.Metadata["thinking_length"] = len(thinkingText)
		a.log.Info("thinking_extracted", map[string]interface{}{
			"enabled":         thinkingEnabled,
			"thinking_tokens": tokenUsage.ThinkingTokens,
			"blocks":          len(blocks),
		})
	} else {
		schemaResponse.Metadata["has_thinking"] = false
		a.log.Info("thinking_not_found", map[string]interface{}{
			"enabled": thinkingEnabled,
		})
	}

	// Log response details
	a.log.Info("bedrock_turn2_response_received", map[string]interface{}{
		"model_id":           response.ModelID,
		"completion_reason":  response.StopReason,
		"input_tokens":       tokenUsage.InputTokens,
		"output_tokens":      tokenUsage.OutputTokens,
		"thinking_tokens":    tokenUsage.ThinkingTokens,
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
		Temperature: a.cfg.Processing.Temperature,
		TopP:        0.9, // Default TopP
		MaxTokens:   a.cfg.Processing.MaxTokens,
	}

	return response, nil
}

// ThinkingBlock represents a structured thinking analysis block
type ThinkingBlock struct {
	Timestamp  string `json:"timestamp"`
	Component  string `json:"component"`
	Stage      string `json:"stage"`
	Decision   string `json:"decision"`
	Reasoning  string `json:"reasoning"`
	Confidence int    `json:"confidence"`
}

// generateThinkingBlocks creates structured thinking blocks for analysis
func (a *AdapterTurn2) generateThinkingBlocks(thinking string, stage string) []ThinkingBlock {
	if thinking == "" {
		return []ThinkingBlock{}
	}

	blocks := []ThinkingBlock{
		{
			Timestamp:  schema.FormatISO8601(),
			Component:  "bedrock-adapter",
			Stage:      stage,
			Decision:   "Extracted thinking content from response",
			Reasoning:  a.summarizeThinking(thinking),
			Confidence: a.calculateThinkingConfidence(thinking),
		},
	}

	if len(thinking) > 500 {
		blocks = append(blocks, ThinkingBlock{
			Timestamp:  schema.FormatISO8601(),
			Component:  "bedrock-adapter",
			Stage:      "content-analysis",
			Decision:   "Analyzed comprehensive thinking content",
			Reasoning:  "Thinking content exceeds 500 characters, indicating detailed reasoning process",
			Confidence: 90,
		})
	}

	return blocks
}

// summarizeThinking creates a summary of thinking content for reasoning field
func (a *AdapterTurn2) summarizeThinking(thinking string) string {
	if len(thinking) <= 200 {
		return thinking
	}

	summary := thinking[:150] + "... [Content extracted using multi-strategy approach: reasoning tags, thinking tags, markdown blocks, and section headers]"
	return summary
}

// calculateThinkingConfidence estimates confidence based on thinking content characteristics
func (a *AdapterTurn2) calculateThinkingConfidence(thinking string) int {
	confidence := 70

	if len(thinking) > 100 {
		confidence += 10
	}
	if len(thinking) > 500 {
		confidence += 10
	}

	indicators := []string{"analysis", "reasoning", "conclusion", "evidence", "assessment"}
	for _, ind := range indicators {
		if contains(thinking, ind) {
			confidence += 2
		}
	}

	if confidence > 95 {
		confidence = 95
	}

	return confidence
}

// truncateForLog safely truncates strings for logging
func (a *AdapterTurn2) truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexContains(s, substr) >= 0))
}

// indexContains finds substring in string
func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
