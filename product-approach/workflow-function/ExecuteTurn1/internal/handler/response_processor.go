package handler

import (
	//"encoding/json"
	//"fmt"
	"regexp"
	"strings"

	"workflow-function/ExecuteTurn1/internal/models"
)

// ResponseProcessor handles processing of Bedrock responses
type ResponseProcessor struct {
	enableThinking bool
	thinkingBudget int
}

// NewResponseProcessor creates a new ResponseProcessor
func NewResponseProcessor(enableThinking bool, thinkingBudget int) *ResponseProcessor {
	return &ResponseProcessor{
		enableThinking: enableThinking,
		thinkingBudget: thinkingBudget,
	}
}

// ProcessResponse processes a raw Bedrock response into a structured TurnResponse
func (p *ResponseProcessor) ProcessResponse(
	response models.BedrockResponse,
	latencyMs int64,
	stage string,
) (*models.TurnResponse, error) {
	// Extract text content from the response
	_, thinking := p.extractContent(response)
	
	// Create the turn response
	turnResponse := &models.TurnResponse{
		TurnID:     1, // This is always 1 for ExecuteTurn1
		Timestamp:  models.FormatISO8601(),
		Response:   response,
		LatencyMs:  latencyMs,
		TokenUsage: extractTokenUsage(response.Usage),
		Stage:      stage,
	}
	
	// Add thinking content if enabled and available
	if p.enableThinking && thinking != "" {
		turnResponse.Thinking = thinking
	}
	
	return turnResponse, nil
}

// extractContent extracts the text content and thinking content from the response
func (p *ResponseProcessor) extractContent(response models.BedrockResponse) (string, string) {
	var responseText, thinking string
	
	// Extract text content from all message contents
	for _, content := range response.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}
	
	// If thinking is enabled, try to extract thinking content
	if p.enableThinking {
		thinking = p.extractThinkingContent(responseText)
	}
	
	return responseText, thinking
}

// extractThinkingContent attempts to extract thinking content from a response
// Looks for <thinking>...</thinking> tags in the response
func (p *ResponseProcessor) extractThinkingContent(text string) string {
	thinkingRegex := regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	matches := thinkingRegex.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		// Return the content inside the thinking tags
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// extractTokenUsage extracts token usage information from the Bedrock response
func extractTokenUsage(usage models.TokenUsage) models.TokenUsage {
	return models.TokenUsage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.InputTokens + usage.OutputTokens,
	}
}

// UpdateWorkflowState updates the workflow state with the results of the turn
func (p *ResponseProcessor) UpdateWorkflowState(
	state *models.WorkflowState,
	turnResponse *models.TurnResponse,
) *models.WorkflowState {
	// Add the turn response to the history
	state.TurnHistory = append(state.TurnHistory, *turnResponse)
	
	// Update workflow state with timestamp
	state.Timestamp = models.FormatISO8601()
	
	// Update stage to indicate that Turn1 is complete
	state.Stage = "TURN1_COMPLETE"
	
	// Update status to indicate progress
	state.Status = "IN_PROGRESS"
	
	return state
}

// ExtractBedrockError attempts to extract a structured error from a Bedrock error
func ExtractBedrockError(err error) *models.Error {
	// Check if it's already a structured error
	if structErr, ok := err.(*models.Error); ok {
		return structErr
	}
	
	// Simple parsing based on common Bedrock error patterns
	errStr := err.Error()
	
	// Check for quota exceeded
	if strings.Contains(errStr, "quota exceeded") || strings.Contains(errStr, "Rate exceeded") {
		return models.NewError(
			"BEDROCK_QUOTA_EXCEEDED",
			"Bedrock API quota exceeded",
			true, // This is retryable
		).WithContext("error", errStr)
	}
	
	// Check for validation errors
	if strings.Contains(errStr, "validation") {
		return models.NewError(
			"BEDROCK_VALIDATION_ERROR",
			"Bedrock API validation error",
			false, // Not retryable
		).WithContext("error", errStr)
	}
	
	// Check for timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return models.NewError(
			"BEDROCK_TIMEOUT",
			"Bedrock API call timed out",
			true, // This is retryable
		).WithContext("error", errStr)
	}
	
	// Default to generic error
	return models.NewError(
		"BEDROCK_ERROR",
		"Error calling Bedrock API",
		true, // Default to retryable
	).WithContext("error", errStr)
}
