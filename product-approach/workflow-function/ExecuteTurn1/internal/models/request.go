// internal/models/request.go

package models

import (
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// ExecuteTurn1Request wraps the incoming Step Functions payload for ExecuteTurn1.
type ExecuteTurn1Request struct {
	WorkflowState schema.WorkflowState `json:"workflowState" validate:"required"`
}

// ExecuteTurn1Response is returned back to Step Functions.
type ExecuteTurn1Response struct {
	WorkflowState schema.WorkflowState  `json:"workflowState"`
	Error         *wferrors.WorkflowError `json:"error,omitempty"`
}

// RequestValidator handles validation of incoming requests
type RequestValidator struct {
	logger logger.Logger
}

// NewRequestValidator creates a new request validator instance
func NewRequestValidator(log logger.Logger) *RequestValidator {
	return &RequestValidator{
		logger: log.WithFields(map[string]interface{}{
			"component": "RequestValidator",
		}),
	}
}

// Validate performs comprehensive validation of the ExecuteTurn1Request.
// This validation is designed to catch structural issues early, before image processing.
func (req *ExecuteTurn1Request) Validate(log logger.Logger) error {
	validator := NewRequestValidator(log)
	return validator.ValidateRequest(req)
}

// ValidateRequest performs the actual validation logic
func (v *RequestValidator) ValidateRequest(req *ExecuteTurn1Request) error {
	v.logger.Debug("Starting request validation", nil)

	// Step 1: Validate basic structure and required fields
	if err := v.validateBasicStructure(req); err != nil {
		return err
	}

	// Step 2: Validate workflow state structure
	if err := v.validateWorkflowStateStructure(&req.WorkflowState); err != nil {
		return err
	}

	// Step 3: Validate prompt structure (without requiring images)
	if err := v.validatePromptStructure(req.WorkflowState.CurrentPrompt); err != nil {
		return err
	}

	// Step 4: Validate Bedrock configuration
	if err := v.validateBedrockConfiguration(req.WorkflowState.BedrockConfig); err != nil {
		return err
	}

	// Step 5: Validate verification context
	if err := v.validateVerificationContext(req.WorkflowState.VerificationContext); err != nil {
		return err
	}

	v.logger.Debug("Request validation completed successfully", map[string]interface{}{
		"verificationId": req.WorkflowState.VerificationContext.VerificationId,
	})

	return nil
}

// validateBasicStructure performs basic nil checks and structure validation
func (v *RequestValidator) validateBasicStructure(req *ExecuteTurn1Request) error {
	if req == nil {
		v.logger.Error("Request is nil", nil)
		return wferrors.NewValidationError("Request cannot be nil", nil)
	}

	// Check for required nested structures
	requiredFields := map[string]interface{}{
		"VerificationContext": req.WorkflowState.VerificationContext,
		"CurrentPrompt":       req.WorkflowState.CurrentPrompt,
		"BedrockConfig":       req.WorkflowState.BedrockConfig,
	}

	for fieldName, fieldValue := range requiredFields {
		if fieldValue == nil {
			v.logger.Error("Required field is nil", map[string]interface{}{
				"field": fieldName,
			})
			return wferrors.NewValidationError(
				fmt.Sprintf("%s is required but was nil", fieldName),
				map[string]interface{}{"field": fieldName},
			)
		}
	}

	return nil
}

// validateWorkflowStateStructure validates the overall workflow state using schema validation
func (v *RequestValidator) validateWorkflowStateStructure(state *schema.WorkflowState) error {
	v.logger.Debug("Validating workflow state structure", map[string]interface{}{
		"schemaVersion":  state.SchemaVersion,
		"verificationId": state.VerificationContext.VerificationId,
	})

	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		v.logger.Error("WorkflowState validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"verificationId":   state.VerificationContext.VerificationId,
			"schemaVersion":    state.SchemaVersion,
		})
		return wferrors.NewValidationError("Invalid WorkflowState structure", map[string]interface{}{
			"validationErrors": errs.Error(),
			"verificationId":   state.VerificationContext.VerificationId,
		})
	}

	return nil
}

// validatePromptStructure validates the current prompt structure without requiring images
func (v *RequestValidator) validatePromptStructure(prompt *schema.CurrentPrompt) error {
	v.logger.Debug("Validating prompt structure", map[string]interface{}{
		"promptId":     prompt.PromptId,
		"messageCount": len(prompt.Messages),
	})

	// Validate prompt structure WITHOUT requiring images
	// Images will be validated later after Base64 generation
	if errs := schema.ValidateCurrentPrompt(prompt, false); len(errs) > 0 {
		v.logger.Error("CurrentPrompt validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         prompt.PromptId,
			"messageCount":     len(prompt.Messages),
		})
		return wferrors.NewValidationError("Invalid CurrentPrompt structure", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         prompt.PromptId,
		})
	}

	// Additional message-level validation
	if err := v.validatePromptMessages(prompt); err != nil {
		return err
	}

	return nil
}

// validatePromptMessages performs detailed validation of prompt messages
func (v *RequestValidator) validatePromptMessages(prompt *schema.CurrentPrompt) error {
	if len(prompt.Messages) == 0 {
		v.logger.Error("Prompt has no messages", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
		return wferrors.NewValidationError("CurrentPrompt must have at least one message", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
	}

	// Validate first message (required for ExecuteTurn1)
	firstMsg := prompt.Messages[0]
	if firstMsg.Role == "" {
		v.logger.Error("First message has empty role", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
		return wferrors.NewValidationError("First message must have a role", map[string]interface{}{
			"promptId": prompt.PromptId,
		})
	}

	// Check if message has content
	if len(firstMsg.Content) == 0 {
		v.logger.Error("First message has no content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
		return wferrors.NewValidationError("First message must have content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
	}

	// Validate text content exists
	if firstMsg.Content[0].Text == "" {
		v.logger.Error("First message has empty text content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
		return wferrors.NewValidationError("First message must have text content", map[string]interface{}{
			"promptId": prompt.PromptId,
			"role":     firstMsg.Role,
		})
	}

	return nil
}

// validateBedrockConfiguration validates the Bedrock configuration
func (v *RequestValidator) validateBedrockConfiguration(config *schema.BedrockConfig) error {
	v.logger.Debug("Validating Bedrock configuration", map[string]interface{}{
		// ModelId not in schema.BedrockConfig
		"anthropicVersion": config.AnthropicVersion,
		"maxTokens":        config.MaxTokens,
	})

	if errs := schema.ValidateBedrockConfig(config); len(errs) > 0 {
		v.logger.Error("BedrockConfig validation failed", map[string]interface{}{
			"validationErrors":  errs.Error(),
			"anthropicVersion":  config.AnthropicVersion,
			"maxTokens":         config.MaxTokens,
		})
		return wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Additional validation for critical fields
	// ModelId validation removed as it's not in the schema

	if config.AnthropicVersion == "" {
		v.logger.Error("BedrockConfig missing AnthropicVersion", nil)
		return wferrors.NewValidationError("BedrockConfig.AnthropicVersion is required", nil)
	}

	if config.MaxTokens <= 0 {
		v.logger.Error("BedrockConfig has invalid MaxTokens", map[string]interface{}{
			"maxTokens": config.MaxTokens,
		})
		return wferrors.NewValidationError("BedrockConfig.MaxTokens must be greater than 0", map[string]interface{}{
			"maxTokens": config.MaxTokens,
		})
	}

	return nil
}

// validateVerificationContext validates the verification context
func (v *RequestValidator) validateVerificationContext(context *schema.VerificationContext) error {
	v.logger.Debug("Validating verification context", map[string]interface{}{
		"verificationId": context.VerificationId,
		"status":         context.Status,
	})

	// Check required fields
	if context.VerificationId == "" {
		v.logger.Error("VerificationContext missing VerificationId", nil)
		return wferrors.NewValidationError("VerificationContext.VerificationId is required", nil)
	}

	// Validate status is not in a terminal error state
	// Error states to check against
	errorStates := []string{
		schema.StatusBedrockProcessingFailed,
		"TURN1_FAILED", // Replace with correct constants from schema
		"TURN2_FAILED", // Replace with correct constants from schema
		schema.StatusVerificationFailed,
	}

	for _, errorState := range errorStates {
		if context.Status == errorState {
			v.logger.Error("VerificationContext is in error state", map[string]interface{}{
				"status":         context.Status,
				"verificationId": context.VerificationId,
			})
			return wferrors.NewValidationError("VerificationContext is in error state", map[string]interface{}{
				"status":         context.Status,
				"verificationId": context.VerificationId,
			})
		}
	}

	return nil
}

// NewRequestFromJSON unmarshals the raw Lambda event into an ExecuteTurn1Request
func NewRequestFromJSON(event []byte, log logger.Logger) (*ExecuteTurn1Request, error) {
	log.Debug("Parsing JSON request", map[string]interface{}{
		"eventSize": len(event),
	})

	if len(event) == 0 {
		log.Error("Empty event payload", nil)
		return nil, wferrors.NewValidationError("Event payload is empty", nil)
	}

	var req ExecuteTurn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		log.Error("Failed to parse ExecuteTurn1Request", map[string]interface{}{
			"error":     err.Error(),
			"eventSize": len(event),
		})
		return nil, wferrors.NewValidationError("Invalid JSON format", map[string]interface{}{
			"parseError": err.Error(),
			"eventSize":  len(event),
		})
	}

	log.Debug("JSON request parsed successfully", map[string]interface{}{
		"verificationId": req.WorkflowState.VerificationContext.VerificationId,
		"schemaVersion":  req.WorkflowState.SchemaVersion,
	})

	return &req, nil
}

// NewResponse creates a successful response wrapper for Step Functions
func NewResponse(state *schema.WorkflowState, log logger.Logger) *ExecuteTurn1Response {
	log.Debug("Creating successful response", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"status":         state.VerificationContext.Status,
	})

	return &ExecuteTurn1Response{
		WorkflowState: *state,
		Error:         nil,
	}
}

// NewErrorResponse creates an error response that attaches the WorkflowError to the state
func NewErrorResponse(
	state *schema.WorkflowState,
	wfErr *wferrors.WorkflowError,
	log logger.Logger,
	status string,
) *ExecuteTurn1Response {
	if wfErr == nil {
		log.Error("Attempting to create error response with nil error", nil)
		wfErr = wferrors.NewInternalError("UnknownError", fmt.Errorf("nil error provided"))
	}

	log.Error("Creating error response", map[string]interface{}{
		"error":          wfErr.Error(),
		"errorCode":      wfErr.Code,
		"retryable":      wfErr.Retryable,
		"status":         status,
		"verificationId": state.VerificationContext.VerificationId,
	})

	// Update verification context with error information
	if state.VerificationContext != nil {
		state.VerificationContext.Status = status
		state.VerificationContext.Error = &schema.ErrorInfo{
			Code:      wfErr.Code,
			Message:   wfErr.Message,
			Details:   wfErr.Context,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		state.VerificationContext.VerificationAt = time.Now().UTC().Format(time.RFC3339)
	}

	return &ExecuteTurn1Response{
		WorkflowState: *state,
		Error:         wfErr,
	}
}

// ValidateAndSanitize performs both validation and basic sanitization of the request
func (req *ExecuteTurn1Request) ValidateAndSanitize(log logger.Logger) error {
	// First validate the request
	if err := req.Validate(log); err != nil {
		return err
	}

	// Perform basic sanitization
	return req.sanitize(log)
}

// sanitize performs basic sanitization of the request data
func (req *ExecuteTurn1Request) sanitize(log logger.Logger) error {
	log.Debug("Sanitizing request", nil)

	// Ensure schema version is current
	if req.WorkflowState.SchemaVersion == "" || req.WorkflowState.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{
			"from": req.WorkflowState.SchemaVersion,
			"to":   schema.SchemaVersion,
		})
		req.WorkflowState.SchemaVersion = schema.SchemaVersion
	}

	// Ensure timestamps are in correct format
	if req.WorkflowState.VerificationContext.VerificationAt == "" {
		req.WorkflowState.VerificationContext.VerificationAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Initialize conversation state if not present
	if req.WorkflowState.ConversationState == nil {
		log.Debug("Initializing conversation state", nil)
		req.WorkflowState.ConversationState = &schema.ConversationState{
			CurrentTurn: 0,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}

	log.Debug("Request sanitization completed", nil)
	return nil
}

// GetVerificationID is a helper method to safely extract the verification ID
func (req *ExecuteTurn1Request) GetVerificationID() string {
	if req.WorkflowState.VerificationContext != nil {
		return req.WorkflowState.VerificationContext.VerificationId
	}
	return ""
}

// GetPromptID is a helper method to safely extract the prompt ID
func (req *ExecuteTurn1Request) GetPromptID() string {
	if req.WorkflowState.CurrentPrompt != nil {
		return req.WorkflowState.CurrentPrompt.PromptId
	}
	return ""
}

// HasImages is a helper method to check if the request contains image data
func (req *ExecuteTurn1Request) HasImages() bool {
	return req.WorkflowState.Images != nil && 
		   (req.WorkflowState.Images.Reference != nil || req.WorkflowState.Images.ReferenceImage != nil)
}

// String implements the Stringer interface for better logging
func (req *ExecuteTurn1Request) String() string {
	return fmt.Sprintf("ExecuteTurn1Request{VerificationID: %s, PromptID: %s, HasImages: %t}",
		req.GetVerificationID(),
		req.GetPromptID(),
		req.HasImages(),
	)
}

// String implements the Stringer interface for ExecuteTurn1Response
func (resp *ExecuteTurn1Response) String() string {
	verificationId := ""
	status := ""
	if resp.WorkflowState.VerificationContext != nil {
		verificationId = resp.WorkflowState.VerificationContext.VerificationId
		status = resp.WorkflowState.VerificationContext.Status
	}

	errorCode := ""
	if resp.Error != nil {
		errorCode = resp.Error.Code
	}

	return fmt.Sprintf("ExecuteTurn1Response{VerificationID: %s, Status: %s, Error: %s}",
		verificationId,
		status,
		errorCode,
	)
}