package handler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// ResponseProcessor handles advanced Bedrock response parsing and workflow state updates.
// It provides specialized functionality for thinking extraction, token analysis, and response processing.
type ResponseProcessor struct {
	logger         logger.Logger
	enableThinking bool
	thinkingBudget int
}

// NewResponseProcessor constructs a new ResponseProcessor with thinking and advanced processing settings.
func NewResponseProcessor(log logger.Logger, enableThinking bool, thinkingBudget int) *ResponseProcessor {
	return &ResponseProcessor{
		logger: log.WithFields(map[string]interface{}{
			"component": "ResponseProcessor",
		}),
		enableThinking: enableThinking,
		thinkingBudget: thinkingBudget,
	}
}

// ProcessBedrockResponse processes raw Bedrock API output into a structured TurnResponse.
// This method handles the advanced parsing logic that was extracted from execute_turn1.go.
func (p *ResponseProcessor) ProcessBedrockResponse(
	rawResponse []byte,
	latencyMs int64,
	modelId string,
	promptText string,
	imageUrls map[string]string,
	turnNumber int,
) (*schema.TurnResponse, error) {
	p.logger.Debug("Processing Bedrock response", map[string]interface{}{
		"responseSize": len(rawResponse),
		"latencyMs":    latencyMs,
		"turnNumber":   turnNumber,
		"modelId":      modelId,
	})

	// Parse the raw JSON response from Bedrock
	bedrockResp, tokenUsage, err := p.parseBedrockJSON(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Bedrock JSON: %w", err)
	}

	// Extract thinking content if enabled
	thinking := p.extractAndProcessThinking(bedrockResp.Content)

	// Create the TurnResponse with all processed data
	turnResp := &schema.TurnResponse{
		TurnId:    turnNumber,
		Timestamp: schema.FormatISO8601(),
		Prompt:    promptText,
		ImageUrls: imageUrls,
		Response:  *bedrockResp,
		LatencyMs: latencyMs,
		TokenUsage: tokenUsage,
		Stage:     p.determineStageFromTurn(turnNumber),
		Metadata:  p.buildResponseMetadata(thinking, modelId),
	}

	p.logger.Info("Bedrock response processed successfully", map[string]interface{}{
		"turnId":         turnResp.TurnId,
		"contentLength":  len(bedrockResp.Content),
		"inputTokens":    tokenUsage.InputTokens,
		"outputTokens":   tokenUsage.OutputTokens,
		"hasThinking":    thinking != "",
		"stopReason":     bedrockResp.StopReason,
	})

	return turnResp, nil
}

// parseBedrockJSON parses the raw Bedrock response JSON and extracts content and usage
func (p *ResponseProcessor) parseBedrockJSON(rawResponse []byte) (*schema.BedrockApiResponse, *schema.TokenUsage, error) {
	var resp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(rawResponse, &resp); err != nil {
		p.logger.Error("Failed to unmarshal Bedrock response", map[string]interface{}{
			"error":        err.Error(),
			"responseSize": len(rawResponse),
		})
		return nil, nil, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	// Concatenate all text content
	var fullText string
	for _, content := range resp.Content {
		if content.Type == "text" {
			fullText += content.Text
		}
	}

	// Build BedrockApiResponse
	bedrockResp := &schema.BedrockApiResponse{
		Content:    fullText,
		StopReason: resp.StopReason,
	}

	// Build TokenUsage
	tokenUsage := &schema.TokenUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	return bedrockResp, tokenUsage, nil
}

// extractAndProcessThinking extracts and processes thinking content from the response
func (p *ResponseProcessor) extractAndProcessThinking(content string) string {
	if !p.enableThinking {
		return ""
	}

	thinking := p.extractThinkingContent(content)
	if thinking == "" {
		return ""
	}

	// Apply thinking budget if configured
	if p.thinkingBudget > 0 && len(thinking) > p.thinkingBudget {
		p.logger.Debug("Truncating thinking content", map[string]interface{}{
			"originalLength": len(thinking),
			"budgetLength":   p.thinkingBudget,
		})
		thinking = thinking[:p.thinkingBudget] + "... [truncated]"
	}

	p.logger.Debug("Thinking content extracted", map[string]interface{}{
		"thinkingLength": len(thinking),
		"hasBudget":      p.thinkingBudget > 0,
	})

	return thinking
}

// extractThinkingContent pulls out content inside <thinking> tags
func (p *ResponseProcessor) extractThinkingContent(text string) string {
	// Match thinking tags with case-insensitive and multiline flags
	regex := regexp.MustCompile(`(?i)(?s)<thinking>(.*?)</thinking>`)
	matches := regex.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		thinking := strings.TrimSpace(matches[1])
		p.logger.Debug("Found thinking content in response", map[string]interface{}{
			"thinkingLength": len(thinking),
		})
		return thinking
	}
	
	// Try alternative patterns
	altPatterns := []string{
		`(?i)(?s)<think>(.*?)</think>`,
		`(?i)(?s)<!-- thinking -->(.*?)<!-- /thinking -->`,
	}
	
	for _, pattern := range altPatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(text)
		if len(matches) > 1 {
			thinking := strings.TrimSpace(matches[1])
			p.logger.Debug("Found thinking content with alternative pattern", map[string]interface{}{
				"pattern":        pattern,
				"thinkingLength": len(thinking),
			})
			return thinking
		}
	}
	
	return ""
}

// buildResponseMetadata creates metadata for the turn response
func (p *ResponseProcessor) buildResponseMetadata(thinking, modelId string) map[string]interface{} {
	metadata := map[string]interface{}{
		"modelId":     modelId,
		"processedAt": schema.FormatISO8601(),
	}

	if thinking != "" {
		metadata["thinking"] = thinking
		metadata["hasThinking"] = true
		metadata["thinkingLength"] = len(thinking)
	} else {
		metadata["hasThinking"] = false
	}

	return metadata
}

// determineStageFromTurn maps turn number to appropriate stage constant
func (p *ResponseProcessor) determineStageFromTurn(turnNumber int) string {
	switch turnNumber {
	case 1:
		return schema.StatusTurn1Completed
	case 2:
		return schema.StatusTurn2Completed
	default:
		p.logger.Warn("Unknown turn number for stage determination", map[string]interface{}{
			"turnNumber": turnNumber,
		})
		// Use a generic status for unknown turn numbers
			return "IN_PROGRESS"
	}
}

// UpdateWorkflowState appends the TurnResponse to history and updates context status.
// This method provides advanced workflow state management beyond basic field updates.
func (p *ResponseProcessor) UpdateWorkflowState(
	state *schema.WorkflowState,
	turnResp *schema.TurnResponse,
) {
	p.logger.Debug("Updating WorkflowState with advanced processing", map[string]interface{}{
		"turnId":         turnResp.TurnId,
		"verificationId": state.VerificationContext.VerificationId,
	})

	// Initialize or update conversation state
	state.ConversationState = p.updateConversationState(state.ConversationState, turnResp)

	// Update turn-specific fields in workflow state
	switch turnResp.TurnId {
	case 1:
		state.Turn1Response = map[string]interface{}{"turnResponse": turnResp}
		state.VerificationContext.Status = schema.StatusTurn1Completed
	case 2:
		state.Turn2Response = map[string]interface{}{"turnResponse": turnResp}
		state.VerificationContext.Status = schema.StatusTurn2Completed
	default:
		p.logger.Warn("Unknown turn ID in response", map[string]interface{}{
			"turnId": turnResp.TurnId,
		})
	}

	// Update verification context metadata
	state.VerificationContext.VerificationAt = schema.FormatISO8601()
	state.VerificationContext.Error = nil

	// Store processing metadata in the appropriate place based on turn
	// WorkflowState doesn't have a Metadata field, so we'll store it in the turn response
	metadataMap := map[string]interface{}{
		"lastProcessedTurn": turnResp.TurnId,
		"lastProcessedAt": schema.FormatISO8601(),
	}
	
	// Add to the appropriate turn response map
	if turnResp.TurnId == 1 && state.Turn1Response != nil {
		state.Turn1Response["processingMetadata"] = metadataMap
	} else if turnResp.TurnId == 2 && state.Turn2Response != nil {
		state.Turn2Response["processingMetadata"] = metadataMap
	}

	p.logger.Info("WorkflowState updated successfully", map[string]interface{}{
		"turnId":      turnResp.TurnId,
		"status":      state.VerificationContext.Status,
		"currentTurn": state.ConversationState.CurrentTurn,
	})
}

// updateConversationState manages the conversation history with advanced logic
func (p *ResponseProcessor) updateConversationState(
	cs *schema.ConversationState,
	turnResp *schema.TurnResponse,
) *schema.ConversationState {
	// Initialize conversation state if needed
	if cs == nil {
		p.logger.Debug("Initializing new conversation state", nil)
		cs = &schema.ConversationState{
			CurrentTurn: 0,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}

	// Add turn to history
	cs.History = append(cs.History, *turnResp)
	cs.CurrentTurn = turnResp.TurnId

	// ConversationState doesn't have a Metadata field in the schema
	// Calculate total latency for logging
	totalLatency := p.calculateTotalLatency(cs.History)

	p.logger.Debug("Conversation state updated", map[string]interface{}{
		"currentTurn":    cs.CurrentTurn,
		"historyLength":  len(cs.History),
		"totalLatencyMs": totalLatency,
		"lastTurnAt":     turnResp.Timestamp,
	})

	return cs
}

// calculateTotalLatency sums up latency across all turns
func (p *ResponseProcessor) calculateTotalLatency(history []interface{}) int64 {
	var total int64
	for _, item := range history {
		if turn, ok := item.(schema.TurnResponse); ok {
			total += turn.LatencyMs
		}
	}
	return total
}

// ExtractBedrockError maps various error types to WorkflowError with enhanced context
func (p *ResponseProcessor) ExtractBedrockError(err error, context map[string]interface{}) *wferrors.WorkflowError {
	p.logger.Debug("Extracting Bedrock error", map[string]interface{}{
		"errorType": fmt.Sprintf("%T", err),
		"error":     err.Error(),
	})

	// If already a WorkflowError, enhance it with additional context
	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
		if context != nil {
			for key, value := range context {
				wfErr = wfErr.WithContext(key, value)
			}
		}
		return wfErr
	}

	errStr := err.Error()
	
	// Enhanced error classification
	switch {
	case strings.Contains(errStr, "ServiceException"), strings.Contains(errStr, "ThrottlingException"):
		return wferrors.NewBedrockError("Bedrock service error", "BEDROCK_SERVICE_ERROR", true).
			WithContext("rawError", errStr).
			WithContext("classification", "service_throttling").
			WithContext("additionalContext", context)

	case strings.Contains(errStr, "ValidationException"), strings.Contains(errStr, "validation"):
		return wferrors.NewBedrockError("Bedrock validation error", "BEDROCK_VALIDATION_ERROR", false).
			WithContext("rawError", errStr).
			WithContext("classification", "validation").
			WithContext("additionalContext", context)

	case strings.Contains(errStr, "timeout"), strings.Contains(errStr, "deadline exceeded"), 
		 strings.Contains(errStr, "context canceled"):
		return wferrors.NewTimeoutError("Bedrock invocation timeout", 0).
			WithContext("rawError", errStr).
			WithContext("classification", "timeout").
			WithContext("additionalContext", context)

	case strings.Contains(errStr, "quota"), strings.Contains(errStr, "limit"), 
		 strings.Contains(errStr, "rate"):
		return wferrors.NewBedrockError("Bedrock quota/rate limit error", "BEDROCK_QUOTA_ERROR", true).
			WithContext("rawError", errStr).
			WithContext("classification", "quota_limit").
			WithContext("additionalContext", context)

	case strings.Contains(errStr, "model"), strings.Contains(errStr, "not found"):
		return wferrors.NewBedrockError("Bedrock model error", "BEDROCK_MODEL_ERROR", false).
			WithContext("rawError", errStr).
			WithContext("classification", "model_not_found").
			WithContext("additionalContext", context)

	default:
		return wferrors.NewBedrockError("Bedrock API error", "BEDROCK_UNKNOWN_ERROR", true).
			WithContext("rawError", errStr).
			WithContext("classification", "unknown").
			WithContext("additionalContext", context)
	}
}

// AnalyzeTokenUsage provides detailed analysis of token consumption
func (p *ResponseProcessor) AnalyzeTokenUsage(turnResp *schema.TurnResponse) map[string]interface{} {
	if turnResp.TokenUsage == nil {
		return map[string]interface{}{"error": "no token usage data"}
	}

	analysis := map[string]interface{}{
		"inputTokens":    turnResp.TokenUsage.InputTokens,
		"outputTokens":   turnResp.TokenUsage.OutputTokens,
		"thinkingTokens": turnResp.TokenUsage.ThinkingTokens,
		"totalTokens":    turnResp.TokenUsage.TotalTokens,
	}

	// Calculate ratios and efficiency metrics
	if turnResp.TokenUsage.TotalTokens > 0 {
		analysis["inputRatio"] = float64(turnResp.TokenUsage.InputTokens) / float64(turnResp.TokenUsage.TotalTokens)
		analysis["outputRatio"] = float64(turnResp.TokenUsage.OutputTokens) / float64(turnResp.TokenUsage.TotalTokens)
		
		if turnResp.TokenUsage.ThinkingTokens > 0 {
			analysis["thinkingRatio"] = float64(turnResp.TokenUsage.ThinkingTokens) / float64(turnResp.TokenUsage.TotalTokens)
		}
	}

	// Performance metrics
	if turnResp.LatencyMs > 0 {
		analysis["tokensPerSecond"] = float64(turnResp.TokenUsage.TotalTokens) / (float64(turnResp.LatencyMs) / 1000.0)
	}

	p.logger.Debug("Token usage analysis completed", analysis)
	return analysis
}

// ValidateResponseStructure performs advanced validation of the TurnResponse
func (p *ResponseProcessor) ValidateResponseStructure(turnResp *schema.TurnResponse) error {
	if turnResp == nil {
		return wferrors.NewValidationError("TurnResponse is nil", nil)
	}

	// Validate required fields
	if turnResp.TurnId <= 0 {
		return wferrors.NewValidationError("Invalid TurnId", map[string]interface{}{
			"turnId": turnResp.TurnId,
		})
	}

	if turnResp.Timestamp == "" {
		return wferrors.NewValidationError("Missing timestamp", nil)
	}

	if turnResp.Response.Content == "" {
		return wferrors.NewValidationError("Empty response content", nil)
	}

	// Validate token usage if present
	if turnResp.TokenUsage != nil {
		if turnResp.TokenUsage.TotalTokens != turnResp.TokenUsage.InputTokens + turnResp.TokenUsage.OutputTokens + turnResp.TokenUsage.ThinkingTokens {
			p.logger.Warn("Token usage totals don't match", map[string]interface{}{
				"calculated": turnResp.TokenUsage.InputTokens + turnResp.TokenUsage.OutputTokens + turnResp.TokenUsage.ThinkingTokens,
				"reported":   turnResp.TokenUsage.TotalTokens,
			})
		}
	}

	p.logger.Debug("Response structure validation passed", map[string]interface{}{
		"turnId": turnResp.TurnId,
	})

	return nil
}

// GetProcessingStats returns statistics about the response processing
func (p *ResponseProcessor) GetProcessingStats() map[string]interface{} {
	return map[string]interface{}{
		"enableThinking": p.enableThinking,
		"thinkingBudget": p.thinkingBudget,
		"component":      "ResponseProcessor",
		"version":        "2.0",
	}
}