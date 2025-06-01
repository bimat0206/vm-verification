package bedrock

import (
	"context"
	"encoding/json"
	"time"

	sharedBedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// BedrockResponse represents the strategic response structure
type BedrockResponse struct {
	Content    string                 `json:"content"`
	TokenUsage schema.TokenUsage      `json:"tokenUsage"`
	RequestID  string                 `json:"requestId"`
	Raw        json.RawMessage        `json:"raw,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Adapter bridges local domain types with shared infrastructure types
type Adapter struct {
	sharedClient *sharedBedrock.BedrockClient
	logger       logger.Logger
}

// NewAdapter creates a new adapter
func NewAdapter(sharedClient *sharedBedrock.BedrockClient, logger logger.Logger) *Adapter {
	return &Adapter{
		sharedClient: sharedClient,
		logger: logger.WithFields(map[string]interface{}{
			"component": "BedrockAdapter",
		}),
	}
}

// Converse handles the Bedrock API call
func (a *Adapter) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string, config *Config) (*BedrockResponse, int64, error) {
	startTime := time.Now()

	a.logger.Debug("adapter_translation_initiated", map[string]interface{}{
		"model_id":    config.ModelID,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
	})

	request, err := a.buildConverseRequest(systemPrompt, turnPrompt, base64Image, config)
	if err != nil {
		return nil, 0, errors.WrapError(err, errors.ErrorTypeInternal,
			"adapter request construction failed", false)
	}

	apiStart := time.Now()
	response, latencyMs, err := a.sharedClient.Converse(ctx, request)
	apiDuration := time.Since(apiStart)

	if err != nil {
		a.logger.Debug("shared_client_invocation_failed", map[string]interface{}{
			"error":      err.Error(),
			"latency_ms": apiDuration.Milliseconds(),
		})
		return nil, latencyMs, err
	}

	bedrockResponse := a.translateResponse(response)

	translationDuration := time.Since(startTime)
	a.logger.Debug("adapter_translation_completed", map[string]interface{}{
		"translation_time_ms": translationDuration.Milliseconds(),
		"api_latency_ms":      latencyMs,
		"token_usage":         bedrockResponse.TokenUsage.TotalTokens,
	})

	return bedrockResponse, latencyMs, nil
}

func (a *Adapter) buildConverseRequest(systemPrompt, turnPrompt, base64Image string, config *Config) (*sharedBedrock.ConverseRequest, error) {
	// Detect image format using lightweight heuristic
	imageFormat := a.detectImageFormat(base64Image)

	imageBlock := sharedBedrock.CreateImageContentFromBytes(imageFormat, base64Image)

	// Construct user message with multimodal content
	userMessage := sharedBedrock.CreateUserMessageWithContent(
		turnPrompt,
		[]sharedBedrock.ContentBlock{imageBlock},
	)

	// Build request using shared package constructor
	temperature := config.Temperature
	request := sharedBedrock.CreateConverseRequest(
		config.ModelID,
		[]sharedBedrock.MessageWrapper{userMessage},
		systemPrompt,
		config.MaxTokens,
		&temperature,
		nil, // TopP - defer to model defaults
	)

	// Add thinking/reasoning configuration if enabled
	if config.ThinkingType == "enable" {
		request.Reasoning = "enable"
		request.InferenceConfig.Reasoning = "enable"
		request.Thinking = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": config.ThinkingBudget,
		}
		a.logger.Info("THINKING_ADAPTER_ENABLED", map[string]interface{}{
			"thinking_type":         config.ThinkingType,
			"budget_tokens":         config.ThinkingBudget,
			"request_reasoning":     request.Reasoning,
			"inference_reasoning":   request.InferenceConfig.Reasoning,
			"request_thinking":      request.Thinking,
		})
	} else {
		a.logger.Info("THINKING_ADAPTER_DISABLED", map[string]interface{}{
			"thinking_type":       config.ThinkingType,
			"request_reasoning":   request.Reasoning,
			"inference_reasoning": request.InferenceConfig.Reasoning,
		})
	}

	// Log the complete request structure for debugging
	a.logger.Info("THINKING_REQUEST_STRUCTURE", map[string]interface{}{
		"model_id":            request.ModelId,
		"reasoning_field":     request.Reasoning,
		"inference_reasoning": request.InferenceConfig.Reasoning,
		"thinking_field":      request.Thinking,
		"max_tokens":          request.InferenceConfig.MaxTokens,
	})

	// Log the final JSON payload for verification
	if payload, err := json.Marshal(request); err == nil {
		a.logger.Info("BEDROCK_REQUEST_PAYLOAD", map[string]interface{}{
			"payload_json": string(payload),
			"payload_size": len(payload),
		})
	} else {
		a.logger.Warn("FAILED_TO_MARSHAL_REQUEST", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return request, nil
}

// translateResponse converts shared response to local type
func (a *Adapter) translateResponse(response *sharedBedrock.ConverseResponse) *BedrockResponse {
	// Extract text content using shared utility
	content := sharedBedrock.ExtractTextFromResponse(response)

	// Map token usage with defensive programming
	tokenUsage := schema.TokenUsage{
		InputTokens:    0,
		OutputTokens:   0,
		ThinkingTokens: 0,
		TotalTokens:    0,
	}

	if response.Usage != nil {
		tokenUsage.InputTokens = response.Usage.InputTokens
		tokenUsage.OutputTokens = response.Usage.OutputTokens
		tokenUsage.TotalTokens = response.Usage.TotalTokens
	}

	// Marshal raw response for audit trail
	raw, _ := json.Marshal(response)

	// Construct response with strategic metadata preservation
	return &BedrockResponse{
		Content:    content,
		TokenUsage: tokenUsage,
		RequestID:  response.RequestID,
		Raw:        raw,
		Metadata: map[string]interface{}{
			"model_id":    response.ModelID,
			"stop_reason": response.StopReason,
		},
	}
}

// detectImageFormat detects image format from base64 data
func (a *Adapter) detectImageFormat(base64Data string) string {
	if len(base64Data) < 20 {
		return "png" // Safe default for edge cases
	}

	// Extract prefix for pattern matching
	prefix := base64Data[:20]

	// PNG: iVBORw0KGgo
	if len(prefix) >= 11 && prefix[:11] == "iVBORw0KGgo" {
		return "png"
	}

	// JPEG: /9j/
	if len(prefix) >= 4 && prefix[:4] == "/9j/" {
		return "jpeg"
	}

	// Default to PNG for maximum compatibility
	return "png"
}

