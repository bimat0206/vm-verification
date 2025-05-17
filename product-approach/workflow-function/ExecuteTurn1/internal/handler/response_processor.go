package handler

import (
	//"fmt"
	"regexp"
	"strings"
	"time"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
	//"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// ResponseProcessor handles all Bedrock Turn response processing and workflow state updates.
type ResponseProcessor struct {
	logger         logger.Logger
	enableThinking bool
	thinkingBudget int
}

// NewResponseProcessor constructs a new ResponseProcessor.
func NewResponseProcessor(log logger.Logger, enableThinking bool, thinkingBudget int) *ResponseProcessor {
	return &ResponseProcessor{
		logger:         log.WithFields(map[string]interface{}{"component": "ResponseProcessor"}),
		enableThinking: enableThinking,
		thinkingBudget: thinkingBudget,
	}
}

// ProcessResponse parses Bedrock API response and produces a TurnResponse.
func (p *ResponseProcessor) ProcessResponse(
	bedrockResp schema.BedrockApiResponse,
	latencyMs int64,
	stage string,
	prompt string,
	imageUrls map[string]string,
	tokenUsage *schema.TokenUsage,
) (*schema.TurnResponse, error) {
	p.logger.Info("Processing Bedrock response", map[string]interface{}{
		"stage":     stage,
		"latencyMs": latencyMs,
	})

	thinking := ""
	if p.enableThinking && bedrockResp.Thinking != "" {
		thinking = bedrockResp.Thinking
	} else if p.enableThinking {
		// Optionally extract <thinking>...</thinking> from content if not already parsed
		thinking = extractThinkingContent(bedrockResp.Content)
	}

	// Create the basic response structure
	turnResponse := &schema.TurnResponse{
		TurnId:     1, // Always 1 for Turn1
		Timestamp:  schema.FormatISO8601(),
		Prompt:     prompt,
		ImageUrls:  imageUrls,
		Response:   bedrockResp,
		LatencyMs:  latencyMs,
		TokenUsage: normalizeTokenUsage(tokenUsage),
		Stage:      stage,
	}
	
	// Add the thinking content to the response metadata if it's not a direct field
	if thinking != "" {
		if turnResponse.Metadata == nil {
			turnResponse.Metadata = make(map[string]interface{})
		}
		turnResponse.Metadata["thinking"] = thinking
	}
	
	return turnResponse, nil
}

// UpdateWorkflowState updates the workflow state after processing Turn1.
func (p *ResponseProcessor) UpdateWorkflowState(
	state *schema.WorkflowState,
	turnResponse *schema.TurnResponse,
) *schema.WorkflowState {
	p.logger.Info("Updating WorkflowState with Turn1 response", map[string]interface{}{
		"turnId":    turnResponse.TurnId,
		"latencyMs": turnResponse.LatencyMs,
	})

	// Add Turn1 to conversation history
	if state.ConversationState == nil {
		state.ConversationState = &schema.ConversationState{
			CurrentTurn: 1,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}
	state.ConversationState.History = append(state.ConversationState.History, *turnResponse)
	state.ConversationState.CurrentTurn = 1

	// Set Turn1Response field
	state.Turn1Response = map[string]interface{}{"turnResponse": turnResponse}
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.VerificationContext.Error = nil
	state.VerificationContext.VerificationAt = schema.FormatISO8601()

	return state
}

// ExtractBedrockError converts a generic Bedrock error to a WorkflowError (Step Functions-friendly).
func (p *ResponseProcessor) ExtractBedrockError(err error) *errors.WorkflowError {
	// Already a WorkflowError
	if wfErr, ok := err.(*errors.WorkflowError); ok {
		return wfErr
	}
	// Simple error pattern mapping
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "quota exceeded") || strings.Contains(errStr, "Rate exceeded"):
		return errors.NewBedrockError("Bedrock API quota exceeded", "BEDROCK_QUOTA_EXCEEDED", true).
			WithContext("rawError", errStr)
	case strings.Contains(errStr, "validation"):
		return errors.NewBedrockError("Bedrock API validation error", "BEDROCK_VALIDATION_ERROR", false).
			WithContext("rawError", errStr)
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return errors.NewTimeoutError("Bedrock Converse", 30*time.Second).
			WithContext("rawError", errStr)
	default:
		return errors.NewBedrockError("Bedrock API error", "BEDROCK_ERROR", true).
			WithContext("rawError", errStr)
	}
}

// --- Helpers ---

// extractThinkingContent pulls out <thinking>...</thinking> tags from text, if present.
func extractThinkingContent(text string) string {
	thinkingRegex := regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	matches := thinkingRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// normalizeTokenUsage normalizes token usage accounting from Bedrock or schema.
func normalizeTokenUsage(usage *schema.TokenUsage) *schema.TokenUsage {
	if usage == nil {
		return &schema.TokenUsage{}
	}
	return &schema.TokenUsage{
		InputTokens:   usage.InputTokens,
		OutputTokens:  usage.OutputTokens,
		ThinkingTokens: usage.ThinkingTokens,
		TotalTokens:   usage.TotalTokens,
	}
}

// FormatISO8601 helper returns UTC now in RFC3339.
func FormatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}

