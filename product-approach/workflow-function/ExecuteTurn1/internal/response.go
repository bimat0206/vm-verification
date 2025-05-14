package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ResponseProcessor handles processing of Bedrock responses
type ResponseProcessor struct{}

// NewResponseProcessor creates a new response processor
func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{}
}

// ProcessTurn1Response processes the Bedrock response for Turn 1
func (rp *ResponseProcessor) ProcessTurn1Response(
	bedrockResponse *BedrockResponse,
	input *ExecuteTurn1Input,
	latencyMs int64,
	timestamp time.Time,
) (*Turn1Response, error) {
	if bedrockResponse == nil {
		return nil, NewValidationError("Bedrock response cannot be nil", nil)
	}

	// Extract text content from response
	responseText := ExtractTextFromResponse(bedrockResponse)
	if responseText == "" {
		return nil, NewBedrockError(
			"No text content found in Bedrock response",
			"EMPTY_RESPONSE",
			false,
		)
	}

	// Extract prompt text from input
	promptText, err := rp.extractPromptText(input)
	if err != nil {
		return nil, err
	}

	// Create Turn1Response object
	turn1Response := &Turn1Response{
		TurnID:        1,
		Timestamp:     timestamp,
		Prompt:        promptText,
		Response:      rp.formatBedrockTextResponse(bedrockResponse),
		LatencyMs:     latencyMs,
		TokenUsage:    bedrockResponse.Usage,
		AnalysisStage: "REFERENCE_ANALYSIS",
		BedrockMetadata: BedrockMetadata{
			ModelID:        bedrockResponse.Model,
			RequestID:      bedrockResponse.ID,
			InvokeLatencyMs: latencyMs,
		},
	}

	// Validate the response
	if err := rp.validateTurn1Response(turn1Response); err != nil {
		return nil, err
	}

	return turn1Response, nil
}

// extractPromptText extracts the text portion from the current prompt
func (rp *ResponseProcessor) extractPromptText(input *ExecuteTurn1Input) (string, error) {
	// Extract current prompt
	currentPrompt, err := ExtractAndValidateCurrentPrompt(input)
	if err != nil {
		return "", err
	}

	if len(currentPrompt.Messages) == 0 {
		return "", NewValidationError("No messages in current prompt", nil)
	}

	var textParts []string

	// Extract text from all messages
	for _, message := range currentPrompt.Messages {
		for _, content := range message.Content {
			if content.Type == "text" && content.Text != nil {
				textParts = append(textParts, *content.Text)
			}
		}
	}

	if len(textParts) == 0 {
		return "", NewValidationError("No text content found in prompt", nil)
	}

	return strings.Join(textParts, "\n"), nil
}

// formatBedrockTextResponse formats the Bedrock response into our internal format
func (rp *ResponseProcessor) formatBedrockTextResponse(response *BedrockResponse) BedrockTextResponse {
	formatted := BedrockTextResponse{
		Content:    ExtractTextFromResponse(response),
		StopReason: response.StopReason,
	}

	// Include thinking if present
	if response.Thinking != nil {
		formatted.Thinking = response.Thinking
	}

	return formatted
}

// validateTurn1Response validates the Turn1Response object
func (rp *ResponseProcessor) validateTurn1Response(response *Turn1Response) error {
	if response == nil {
		return NewValidationError("Turn1Response cannot be nil", nil)
	}

	if response.TurnID != 1 {
		return NewInvalidFieldError("turnId", response.TurnID, "1")
	}

	if response.Response.Content == "" {
		return NewValidationError("Response content cannot be empty", nil)
	}

	if response.LatencyMs < 0 {
		return NewInvalidFieldError("latencyMs", response.LatencyMs, "non-negative")
	}

	if response.TokenUsage.Total <= 0 {
		return NewInvalidFieldError("tokenUsage.total", response.TokenUsage.Total, "positive")
	}

	if response.AnalysisStage != "REFERENCE_ANALYSIS" {
		return NewInvalidFieldError("analysisStage", response.AnalysisStage, "REFERENCE_ANALYSIS")
	}

	return nil
}

// CreateConversationHistory creates conversation history from Turn1Response
func (rp *ResponseProcessor) CreateConversationHistory(
	verificationContext *VerificationContext,
	turn1Response *Turn1Response,
) *ConversationState {
	history := TurnHistory{
		TurnID:        turn1Response.TurnID,
		Timestamp:     turn1Response.Timestamp,
		Prompt:        turn1Response.Prompt,
		Response:      turn1Response.Response.Content,
		LatencyMs:     turn1Response.LatencyMs,
		TokenUsage:    turn1Response.TokenUsage,
		AnalysisStage: turn1Response.AnalysisStage,
	}

	return &ConversationState{
		CurrentTurn: 1,
		MaxTurns:    verificationContext.TurnConfig.MaxTurns,
		History:     []TurnHistory{history},
	}
}

// GenerateRequestID generates a unique request ID for tracking
func (rp *ResponseProcessor) GenerateRequestID() string {
	return fmt.Sprintf("req-%s", uuid.New().String())
}

// ExtractImageReferences extracts image references from the prompt
func (rp *ResponseProcessor) ExtractImageReferences(input *ExecuteTurn1Input) map[string]string {
	imageRefs := make(map[string]string)

	// Extract current prompt
	currentPrompt, err := ExtractAndValidateCurrentPrompt(input)
	if err != nil {
		return imageRefs
	}

	// For Turn 1, we expect a reference image
	for _, message := range currentPrompt.Messages {
		for _, content := range message.Content {
			if content.Image != nil {
				imageRefs["reference"] = content.Image.Source.S3Location.URI
			}
		}
	}

	return imageRefs
}

// CalculatePerformanceMetrics calculates performance metrics for monitoring
func (rp *ResponseProcessor) CalculatePerformanceMetrics(
	turn1Response *Turn1Response,
	input *ExecuteTurn1Input,
) map[string]interface{} {
	metrics := map[string]interface{}{
		"latencyMs":      turn1Response.LatencyMs,
		"inputTokens":    turn1Response.TokenUsage.Input,
		"outputTokens":   turn1Response.TokenUsage.Output,
		"totalTokens":    turn1Response.TokenUsage.Total,
		"thinkingTokens": turn1Response.TokenUsage.Thinking,
	}

	// Calculate tokens per second
	if turn1Response.LatencyMs > 0 {
		tokensPerSecond := float64(turn1Response.TokenUsage.Total) / (float64(turn1Response.LatencyMs) / 1000.0)
		metrics["tokensPerSecond"] = tokensPerSecond
	}

	// Add verification type specific metrics
	metrics["verificationType"] = input.VerificationContext.VerificationType
	metrics["turnNumber"] = 1

	return metrics
}

// ValidateResponseContent validates the content of the response
func (rp *ResponseProcessor) ValidateResponseContent(response *BedrockResponse) error {
	if response == nil {
		return NewValidationError("Response cannot be nil", nil)
	}

	if len(response.Content) == 0 {
		return NewBedrockError(
			"Response contains no content",
			"EMPTY_CONTENT",
			false,
		)
	}

	// Check for text content
	hasText := false
	for _, content := range response.Content {
		if content.Type == "text" && content.Text != nil && *content.Text != "" {
			hasText = true
			break
		}
	}

	if !hasText {
		return NewBedrockError(
			"Response contains no text content",
			"NO_TEXT_CONTENT",
			false,
		)
	}

	return nil
}

// FormatResponseForLogging formats response for logging (sanitized)
func (rp *ResponseProcessor) FormatResponseForLogging(response *Turn1Response) map[string]interface{} {
	return map[string]interface{}{
		"turnId":        response.TurnID,
		"timestamp":     response.Timestamp,
		"latencyMs":     response.LatencyMs,
		"tokenUsage":    response.TokenUsage,
		"analysisStage": response.AnalysisStage,
		"modelId":       response.BedrockMetadata.ModelID,
		"responseLength": len(response.Response.Content),
		"hasThinking":   response.Response.Thinking != nil,
	}
}

// ExtractAnalysisInsights extracts key insights from the Turn 1 response
// This is use case specific and helps prepare context for Turn 2
func (rp *ResponseProcessor) ExtractAnalysisInsights(
	turn1Response *Turn1Response,
	verificationType string,
) map[string]interface{} {
	insights := make(map[string]interface{})

	responseContent := turn1Response.Response.Content
	
	switch verificationType {
	case VerificationTypeLayoutVsChecking:
		// For UC1, extract layout validation insights
		insights["layoutValidated"] = true
		insights["structureConfirmed"] = strings.Contains(responseContent, "confirmed") || 
			strings.Contains(responseContent, "validated")
		insights["analysisType"] = "REFERENCE_LAYOUT_ANALYSIS"
		
		// Check for specific confirmations
		if strings.Contains(responseContent, "rows") && strings.Contains(responseContent, "columns") {
			insights["structuralAnalysisComplete"] = true
		}
		
	case VerificationTypePreviousVsCurrent:
		// For UC2, extract baseline state insights
		insights["baselineEstablished"] = true
		insights["analysisType"] = "PREVIOUS_STATE_ANALYSIS"
		
		// Extract state information
		if strings.Contains(responseContent, "empty") {
			insights["hasEmptyPositions"] = true
		}
		if strings.Contains(responseContent, "partial") {
			insights["hasPartialRows"] = true
		}
	}

	// Common extractions
	insights["responseLength"] = len(responseContent)
	insights["hasDetailedAnalysis"] = len(responseContent) > 500
	insights["confidenceIndicators"] = rp.extractConfidenceIndicators(responseContent)

	return insights
}

// extractConfidenceIndicators extracts confidence indicators from response text
func (rp *ResponseProcessor) extractConfidenceIndicators(content string) []string {
	indicators := []string{}
	
	// Look for confidence-related keywords
	confidenceKeywords := []string{
		"confident", "clear", "visible", "confirmed", "validated",
		"uncertain", "unclear", "difficult", "obscured", "partial",
	}
	
	contentLower := strings.ToLower(content)
	for _, keyword := range confidenceKeywords {
		if strings.Contains(contentLower, keyword) {
			indicators = append(indicators, keyword)
		}
	}
	
	return indicators
}

// PrepareMetadataForTurn2 prepares metadata that will be useful for Turn 2
func (rp *ResponseProcessor) PrepareMetadataForTurn2(
	turn1Response *Turn1Response,
	input *ExecuteTurn1Input,
) map[string]interface{} {
	metadata := map[string]interface{}{
		"turn1Completed":    true,
		"turn1Timestamp":    turn1Response.Timestamp,
		"turn1LatencyMs":    turn1Response.LatencyMs,
		"turn1TokenUsage":   turn1Response.TokenUsage,
		"verificationType":  input.VerificationContext.VerificationType,
		"analysisStage":     turn1Response.AnalysisStage,
		"baselineEstablished": true,
	}

	// Add insights from Turn 1
	insights := rp.ExtractAnalysisInsights(turn1Response, input.VerificationContext.VerificationType)
	metadata["turn1Insights"] = insights

	return metadata
}

// FormatSuccessResponse formats the successful response
func (rp *ResponseProcessor) FormatSuccessResponse(
	input *ExecuteTurn1Input,
	turn1Response *Turn1Response,
) *ExecuteTurn1Output {
	// Update verification context
	updatedContext := input.VerificationContext
	updatedContext.Status = "TURN1_COMPLETED"
	
	// Set Turn 1 timestamp if not already set
	if updatedContext.TurnTimestamps.Turn1 == nil {
		now := time.Now()
		updatedContext.TurnTimestamps.Turn1 = &now
	}

	// Create conversation state
	conversationState := rp.CreateConversationHistory(&updatedContext, turn1Response)

	return &ExecuteTurn1Output{
		VerificationContext: updatedContext,
		Turn1Response:      *turn1Response,
		ConversationState:  *conversationState,
	}
}