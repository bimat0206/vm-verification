package handler

import (
	"regexp"
	"strings"
	"time"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// ResponseProcessor handles Bedrock turn response parsing and workflow state updates.
type ResponseProcessor struct {
	logger         logger.Logger
	enableThinking bool
	thinkingBudget int
}

// NewResponseProcessor constructs a new ResponseProcessor with thinking settings.
func NewResponseProcessor(log logger.Logger, enableThinking bool, thinkingBudget int) *ResponseProcessor {
	return &ResponseProcessor{
		logger:         log.WithFields(map[string]interface{}{"component": "ResponseProcessor"}),
		enableThinking: enableThinking,
		thinkingBudget: thinkingBudget,
	}
}

// ProcessResponse transforms a BedrockApiResponse into a TurnResponse, applying thinking extraction and token normalization.
func (p *ResponseProcessor) ProcessResponse(
	bedrockResp schema.BedrockApiResponse,
	latencyMs int64,
	stage string,
	prompt string,
	imageUrls map[string]string,
	tokenUsage *schema.TokenUsage,
) (*schema.TurnResponse, error) {
	p.logger.Info("Processing Bedrock response", map[string]interface{}{"stage": stage, "latencyMs": latencyMs})

	// Extract thinking if enabled
	var thinking string
	if p.enableThinking {
		if bedrockResp.Thinking != "" {
			thinking = bedrockResp.Thinking
		} else {
			thinking = extractThinkingContent(bedrockResp.Content)
		}
		if p.thinkingBudget > 0 && len(thinking) > p.thinkingBudget {
			thinking = thinking[:p.thinkingBudget]
		}
	}

	// Build TurnResponse
	turnResp := &schema.TurnResponse{
		TurnId:    determineTurnID(stage),
		Timestamp: schema.FormatISO8601(),
		Prompt:    prompt,
		ImageUrls: imageUrls,
		Response:  bedrockResp,
		LatencyMs: latencyMs,
		TokenUsage: mergeTokenUsage(tokenUsage, bedrockResp),
		Stage:     stage,
	}

	// Attach thinking to metadata
	if thinking != "" {
		if turnResp.Metadata == nil {
			turnResp.Metadata = make(map[string]interface{})
		}
		turnResp.Metadata["thinking"] = thinking
	}

	return turnResp, nil
}

// UpdateWorkflowState appends the TurnResponse to history and updates context status.
func (p *ResponseProcessor) UpdateWorkflowState(
	state *schema.WorkflowState,
	turnResp *schema.TurnResponse,
) *schema.WorkflowState {
	p.logger.Info("Updating WorkflowState with response", map[string]interface{}{"turnId": turnResp.TurnId})

	// Initialize conversation state if needed
	if state.ConversationState == nil {
		state.ConversationState = &schema.ConversationState{
			CurrentTurn: turnResp.TurnId,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}
	state.ConversationState.History = append(state.ConversationState.History, *turnResp)
	state.ConversationState.CurrentTurn = turnResp.TurnId

	// Update verification context
	state.Turn1Response = map[string]interface{}{"turnResponse": turnResp}
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.VerificationContext.Error = nil
	state.VerificationContext.VerificationAt = schema.FormatISO8601()

	return state
}

// ExtractBedrockError maps generic errors to WorkflowError for Step Functions.
func (p *ResponseProcessor) ExtractBedrockError(err error) *wferrors.WorkflowError {
	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
		return wfErr
	}
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "ServiceException"), strings.Contains(errStr, "ThrottlingException"):
		return wferrors.NewBedrockError("Bedrock API error", "BEDROCK_API_ERROR", true).
			WithContext("rawError", errStr)
	case strings.Contains(errStr, "validation"):
		return wferrors.NewBedrockError("Bedrock validation error", "BEDROCK_VALIDATION_ERROR", false).
			WithContext("rawError", errStr)
	case strings.Contains(errStr, "timeout"), strings.Contains(errStr, "deadline exceeded"):
		return wferrors.NewTimeoutError("Bedrock invocation", time.Duration(0)).
			WithContext("rawError", errStr)
	default:
		return wferrors.NewBedrockError("Bedrock API error", "BEDROCK_ERROR", true).
			WithContext("rawError", errStr)
	}
}

// determineTurnID infers the numeric turn based on stage.
func determineTurnID(stage string) int {
	switch stage {
	case schema.StatusTurn1Completed:
		return 1
	case schema.StatusTurn2Completed:
		return 2
	default:
		return 0
	}
}

// mergeTokenUsage combines provided tokenUsage with values from BedrockApiResponse.
func mergeTokenUsage(
	usage *schema.TokenUsage,
	resp schema.BedrockApiResponse,
) *schema.TokenUsage {
	// Since token usage is not directly in BedrockApiResponse, 
	// we just return the provided usage or create a new one
	if usage == nil {
		// Construct fresh usage with estimated values
		return &schema.TokenUsage{
			InputTokens:    0, // Set default or estimated values
			OutputTokens:   len(resp.Content) / 4, // Rough estimate
			ThinkingTokens: func() int { if resp.Thinking != "" { return len(resp.Thinking) / 4 } else { return 0 } }(),
			TotalTokens:    len(resp.Content) / 4, // Will be updated when actual counts are available
		}
	}
	// Fill thinking tokens if present in response
	if usage.ThinkingTokens == 0 && resp.Thinking != "" {
		usage.ThinkingTokens = len(resp.Thinking) / 4 // Rough estimate
		usage.TotalTokens += usage.ThinkingTokens
	}
	return usage
}

// extractThinkingContent pulls out content inside <thinking> tags.
func extractThinkingContent(text string) string {
	regex := regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	m := regex.FindStringSubmatch(text)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}