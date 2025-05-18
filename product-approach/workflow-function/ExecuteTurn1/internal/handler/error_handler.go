package handler

import (
	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// createAndLogError centralizes error creation, logging, and state updates
func (h *Handler) createAndLogError(
	state *schema.WorkflowState,
	err error,
	log logger.Logger,
	status string,
) error {
	// Ensure we have a WorkflowError
	var wfErr *wferrors.WorkflowError
	if e, ok := err.(*wferrors.WorkflowError); ok {
		wfErr = e
	} else {
		wfErr = wferrors.WrapError(err, wferrors.ErrorTypeInternal, "unexpected error", false)
	}

	// Log the error with full context
	log.Error("ExecuteTurn1 error", map[string]interface{}{
		"error":     wfErr.Error(),
		"errorCode": wfErr.Code,
		"retryable": wfErr.Retryable,
		"context":   wfErr.Context,
	})

	// Update state with error information
	if state.VerificationContext != nil {
		state.VerificationContext.Error = &schema.ErrorInfo{
			Code:      wfErr.Code,
			Message:   wfErr.Message,
			Details:   wfErr.Context,
			Timestamp: schema.FormatISO8601(),
		}
		state.VerificationContext.Status = status
	}

	return wfErr
}