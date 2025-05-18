// internal/request.go

package handler

import (
    "encoding/json"
    "workflow-function/shared/schema"
    "workflow-function/shared/logger"
    "workflow-function/shared/errors"
    "time"
)

// ExecuteTurn1Request wraps the incoming Step Functions payload.
type ExecuteTurn1Request struct {
    WorkflowState schema.WorkflowState `json:"workflowState"`
}

// ExecuteTurn1Response is returned back to Step Functions.
type ExecuteTurn1Response struct {
    WorkflowState schema.WorkflowState `json:"workflowState"`
    Error         *errors.WorkflowError `json:"error,omitempty"`
}

// Validate performs only lightweight sanity checks (no Base64) on the incoming request.
func (req *ExecuteTurn1Request) Validate(log logger.Logger) error {
    // Validate workflow structure
    if errs := schema.ValidateWorkflowState(&req.WorkflowState); len(errs) > 0 {
        log.Error("WorkflowState validation failed", map[string]interface{}{"validationErrors": errs.Error()})
        return errors.NewValidationError("Invalid WorkflowState", map[string]interface{}{"validationErrors": errs.Error()})
    }

    // Validate prompt and Bedrock config (images will be validated after Base64 generation)
    if errs := schema.ValidateCurrentPrompt(req.WorkflowState.CurrentPrompt, true); len(errs) > 0 {
        log.Error("CurrentPrompt validation failed", map[string]interface{}{"validationErrors": errs.Error()})
        return errors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{"validationErrors": errs.Error()})
    }
    if errs := schema.ValidateBedrockConfig(req.WorkflowState.BedrockConfig); len(errs) > 0 {
        log.Error("BedrockConfig validation failed", map[string]interface{}{"validationErrors": errs.Error()})
        return errors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{"validationErrors": errs.Error()})
    }

    return nil
}

// NewRequestFromJSON unmarshals the raw Lambda event.
func NewRequestFromJSON(event []byte, log logger.Logger) (*ExecuteTurn1Request, error) {
    var req ExecuteTurn1Request
    if err := json.Unmarshal(event, &req); err != nil {
        log.Error("Failed to parse ExecuteTurn1Request", map[string]interface{}{"error": err.Error()})
        return nil, errors.NewValidationError("Invalid input format", map[string]interface{}{"error": err.Error()})
    }
    return &req, nil
}

// NewResponse wraps a successful state for Step Functions.
func NewResponse(state *schema.WorkflowState) *ExecuteTurn1Response {
    return &ExecuteTurn1Response{WorkflowState: *state}
}

// NewErrorResponse attaches a WorkflowError into the state for Step Functions.
func NewErrorResponse(state *schema.WorkflowState, wfErr *errors.WorkflowError, log logger.Logger, status string) *ExecuteTurn1Response {
    // Attach error info back into the state
    if state.VerificationContext != nil {
        state.VerificationContext.Status = status
        state.VerificationContext.Error = &schema.ErrorInfo{
            Code:      wfErr.Code,
            Message:   wfErr.Message,
            Details:   wfErr.Context,
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        }
    }
    log.Error("Creating error response", map[string]interface{}{"error": wfErr.Error(), "status": status})
    return &ExecuteTurn1Response{WorkflowState: *state, Error: wfErr}
}
