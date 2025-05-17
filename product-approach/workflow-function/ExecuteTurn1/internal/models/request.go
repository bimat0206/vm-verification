package request

import (
	"encoding/json"
	//"fmt"
	"time"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
)

// --- Lambda Input/Output Contracts ---

// ExecuteTurn1Request represents the input to the ExecuteTurn1 Lambda function.
// Uses shared schema.WorkflowState for the payload.
type ExecuteTurn1Request struct {
	WorkflowState schema.WorkflowState `json:"workflowState"`
}

// ExecuteTurn1Response represents the output of the ExecuteTurn1 Lambda function.
// Mirrors the contract for Step Functions and error propagation.
type ExecuteTurn1Response struct {
	WorkflowState *schema.WorkflowState  `json:"workflowState,omitempty"`
	Error         *errors.WorkflowError  `json:"error,omitempty"`
}

// --- Core Helper Functions ---

// Validate checks the request's WorkflowState using schema validators and logs context.
func (req *ExecuteTurn1Request) Validate(log logger.Logger) error {
	// Ensure schema version is set to the latest supported version
	if req.WorkflowState.SchemaVersion == "" || req.WorkflowState.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{
			"from": req.WorkflowState.SchemaVersion, 
			"to": schema.SchemaVersion,
		})
		req.WorkflowState.SchemaVersion = schema.SchemaVersion
	}

	// Validate main workflow state
	valErrs := schema.ValidateWorkflowState(&req.WorkflowState)
	if len(valErrs) > 0 {
		log.Error("WorkflowState validation failed", map[string]interface{}{"validationErrors": valErrs.Error()})
		return errors.NewValidationError("Invalid WorkflowState", map[string]interface{}{"validationErrors": valErrs.Error()})
	}

	// Validate image data
	if errs := schema.ValidateImageData(req.WorkflowState.Images, true); len(errs) > 0 {
		log.Error("ImageData validation failed", map[string]interface{}{"validationErrors": errs.Error()})
		return errors.NewValidationError("Invalid ImageData", map[string]interface{}{"validationErrors": errs.Error()})
	}

	// Validate current prompt
	if errs := schema.ValidateCurrentPrompt(req.WorkflowState.CurrentPrompt, true); len(errs) > 0 {
		log.Error("CurrentPrompt validation failed", map[string]interface{}{"validationErrors": errs.Error()})
		return errors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{"validationErrors": errs.Error()})
	}

	// Validate Bedrock config
	if errs := schema.ValidateBedrockConfig(req.WorkflowState.BedrockConfig); len(errs) > 0 {
		log.Error("BedrockConfig validation failed", map[string]interface{}{"validationErrors": errs.Error()})
		return errors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{"validationErrors": errs.Error()})
	}

	return nil
}

// --- Utility Functions for Tests/Integration ---

// NewRequestFromJSON parses a raw JSON event into an ExecuteTurn1Request.
func NewRequestFromJSON(event []byte, log logger.Logger) (*ExecuteTurn1Request, error) {
	var req ExecuteTurn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		log.Error("Failed to unmarshal ExecuteTurn1Request", map[string]interface{}{"error": err.Error()})
		return nil, errors.NewValidationError("Failed to unmarshal request", map[string]interface{}{"error": err.Error()})
	}
	return &req, nil
}

// NewResponse creates a successful response.
func NewResponse(state *schema.WorkflowState) *ExecuteTurn1Response {
	return &ExecuteTurn1Response{
		WorkflowState: state,
	}
}

// NewErrorResponse creates a standardized error response.
func NewErrorResponse(state *schema.WorkflowState, err error, log logger.Logger, status string) *ExecuteTurn1Response {
	var wfErr *errors.WorkflowError
	if e, ok := err.(*errors.WorkflowError); ok {
		wfErr = e
	} else {
		wfErr = errors.WrapError(err, errors.ErrorTypeInternal, "Unexpected error", false)
	}

	// Attach error to the state for Step Functions compatibility
	if state != nil && state.VerificationContext != nil {
		// Convert WorkflowError to schema.ErrorInfo
		state.VerificationContext.Error = &schema.ErrorInfo{
			Code:      wfErr.Code,
			Message:   wfErr.Message,
			Details:   wfErr.Context,
			Timestamp: schema.FormatISO8601(),
		}
		state.VerificationContext.Status = status
		state.VerificationContext.VerificationAt = schema.FormatISO8601()
	}

	log.Error("Creating error response", map[string]interface{}{
		"error":  wfErr.Error(),
		"status": status,
	})

	return &ExecuteTurn1Response{
		WorkflowState: state,
		Error:         wfErr,
	}
}

// --- (Optional) Utilities for Handler/Orchestration ---

// SetVerificationStatus is a helper to update the status field in the workflow state.
func SetVerificationStatus(state *schema.WorkflowState, status string) {
	if state != nil && state.VerificationContext != nil {
		state.VerificationContext.Status = status
		state.VerificationContext.VerificationAt = schema.FormatISO8601()
	}
}

// FormatISO8601 utility returns current UTC in RFC3339.
func FormatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}
