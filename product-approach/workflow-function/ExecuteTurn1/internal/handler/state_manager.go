package handler

import (
	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
)

// updateWorkflowStateWithResults updates the workflow state with Turn 1 results
func (h *Handler) updateWorkflowStateWithResults(
	state *schema.WorkflowState,
	turnResponse *schema.TurnResponse,
	log logger.Logger,
) {
	log.Debug("Updating workflow state with results", map[string]interface{}{
		"turnId": turnResponse.TurnId,
	})

	// Update Turn1Response
	state.Turn1Response = map[string]interface{}{"turnResponse": turnResponse}

	// Update verification context
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.VerificationContext.VerificationAt = schema.FormatISO8601()
	state.VerificationContext.Error = nil

	// Update conversation state
	state.ConversationState = h.updateConversationState(state, turnResponse)

	log.Debug("Workflow state updated successfully", map[string]interface{}{
		"status":      state.VerificationContext.Status,
		"currentTurn": state.ConversationState.CurrentTurn,
	})
}

// updateConversationState appends the turn to the conversation history
func (h *Handler) updateConversationState(
	state *schema.WorkflowState,
	turn *schema.TurnResponse,
) *schema.ConversationState {
	cs := state.ConversationState
	if cs == nil {
		cs = &schema.ConversationState{
			CurrentTurn: turn1ID,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}
	
	cs.History = append(cs.History, *turn)
	cs.CurrentTurn = turn1ID
	return cs
}