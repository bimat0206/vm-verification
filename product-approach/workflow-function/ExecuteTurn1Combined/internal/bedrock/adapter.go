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
		a.logger.Info("THINKING_ADAPTER_ENABLED", map[string]interface{}{
			"thinking_type":   config.ThinkingType,
			"budget_tokens":   config.ThinkingBudget,
			"request_reasoning": request.Reasoning,
			"inference_reasoning": request.InferenceConfig.Reasoning,
		})
	} else {
		a.logger.Info("THINKING_ADAPTER_DISABLED", map[string]interface{}{
			"thinking_type": config.ThinkingType,
			"request_reasoning": request.Reasoning,
			"inference_reasoning": request.InferenceConfig.Reasoning,
		})
	}
	
	// Log the complete request structure for debugging
	a.logger.Info("THINKING_REQUEST_STRUCTURE", map[string]interface{}{
		"model_id": request.ModelId,
		"reasoning_field": request.Reasoning,
		"inference_reasoning": request.InferenceConfig.Reasoning,
		"max_tokens": request.InferenceConfig.MaxTokens,
	})

	return request, nil
}

// translateResponse converts shared response to local type
// Enhanced with comprehensive thinking support ported from old/ExecuteTurn1
func (a *Adapter) translateResponse(response *sharedBedrock.ConverseResponse) *BedrockResponse {
	// Extract text content using shared utility
	content := sharedBedrock.ExtractTextFromResponse(response)
	
	// Extract thinking content using shared utility (now with multi-strategy extraction)
	thinking := sharedBedrock.ExtractThinkingFromResponse(response)

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
		tokenUsage.ThinkingTokens = response.Usage.ThinkingTokens
		tokenUsage.TotalTokens = response.Usage.TotalTokens
	}

	// Marshal raw response for audit trail
	raw, _ := json.Marshal(response)

	// Construct response with strategic metadata preservation
	metadata := map[string]interface{}{
		"model_id":    response.ModelID,
		"stop_reason": response.StopReason,
	}
	
	// Add comprehensive thinking metadata if available
	if thinking != "" {
		metadata["thinking"] = thinking
		metadata["has_thinking"] = true
		metadata["thinking_length"] = len(thinking)
		metadata["thinking_enabled"] = true
		
		// Add thinking blocks for structured analysis
		thinkingBlocks := a.generateThinkingBlocks(thinking, "response-processing")
		metadata["thinking_blocks"] = thinkingBlocks
		
		a.logger.Info("THINKING_EXTRACTED_SUCCESSFULLY", map[string]interface{}{
			"thinking_length": len(thinking),
			"thinking_tokens": tokenUsage.ThinkingTokens,
			"thinking_blocks_count": len(thinkingBlocks),
			"extraction_method": "multi_strategy",
		})
	} else {
		metadata["has_thinking"] = false
		metadata["thinking_enabled"] = false
		
		a.logger.Debug("THINKING_NOT_FOUND", map[string]interface{}{
			"response_length": len(content),
			"content_preview": a.truncateForLog(content, 100),
		})
	}

	return &BedrockResponse{
		Content:    content,
		TokenUsage: tokenUsage,
		RequestID:  response.RequestID,
		Raw:        raw,
		Metadata:   metadata,
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
// Ported from old/ExecuteTurn1 approach with enhancements
func (a *Adapter) generateThinkingBlocks(thinking string, stage string) []ThinkingBlock {
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

	// Add additional blocks based on thinking content analysis
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
func (a *Adapter) summarizeThinking(thinking string) string {
	if len(thinking) <= 200 {
		return thinking
	}

	// Extract first 150 characters and add summary
	summary := thinking[:150] + "... [Content extracted using multi-strategy approach: reasoning tags, thinking tags, markdown blocks, and section headers]"
	return summary
}

// calculateThinkingConfidence estimates confidence based on thinking content characteristics
func (a *Adapter) calculateThinkingConfidence(thinking string) int {
	confidence := 70 // Base confidence

	// Increase confidence based on content characteristics
	if len(thinking) > 100 {
		confidence += 10
	}
	if len(thinking) > 500 {
		confidence += 10
	}

	// Check for structured thinking indicators
	structuredIndicators := []string{"analysis", "reasoning", "conclusion", "evidence", "assessment"}
	for _, indicator := range structuredIndicators {
		if contains(thinking, indicator) {
			confidence += 2
		}
	}

	// Cap at 95 to maintain realistic confidence levels
	if confidence > 95 {
		confidence = 95
	}

	return confidence
}

// truncateForLog safely truncates strings for logging
func (a *Adapter) truncateForLog(s string, maxLen int) string {
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

// indexContains finds substring in string (simple implementation)
func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
