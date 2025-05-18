package models

import (
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
	"fmt"
)

// Package-level convenience functions for backward compatibility and simplified API

// ParseAndValidateRequest is the main entry point for processing incoming requests
// This combines parsing, validation, and sanitization in a single call
func ParseAndValidateRequest(event []byte, log logger.Logger) (*ExecuteTurn1Request, error) {
	// Parse the JSON
	parser := NewParser(log)
	req, err := parser.ParseRequestFromJSON(event)
	if err != nil {
		return nil, err
	}

	// Validate the request
	validator := NewValidator(log)
	if err := validator.ValidateRequest(req); err != nil {
		return nil, err
	}

	// Basic sanitization (ensure schema version is current, timestamps are set)
	if err := sanitizeRequest(req, log); err != nil {
		return nil, err
	}

	return req, nil
}

// BuildResponse creates an appropriate response based on whether an error occurred
func BuildResponse(state *schema.WorkflowState, err error, log logger.Logger) *ExecuteTurn1Response {
	builder := NewResponseBuilder(log)
	
	if err != nil {
		// Determine appropriate status based on error type
		status := determineErrorStatus(err, log)
		
		// Ensure we have a WorkflowError
		wfErr := builder.EnsureWorkflowError(err)
		
		if state != nil {
			return builder.BuildErrorResponse(state, wfErr, status)
		} else {
			return builder.BuildErrorResponseWithoutState(wfErr, status)
		}
	}
	
	return builder.BuildSuccessResponse(state)
}

// RequestProcessor provides a simple interface for processing requests end-to-end
type RequestProcessor struct {
	parser    *Parser
	validator *Validator
	builder   *ResponseBuilder
	logger    logger.Logger
}

// NewRequestProcessor creates a new request processor with all components
func NewRequestProcessor(log logger.Logger) *RequestProcessor {
	return &RequestProcessor{
		parser:    NewParser(log),
		validator: NewValidator(log),
		builder:   NewResponseBuilder(log),
		logger:    log,
	}
}

// ProcessRequest handles the complete request processing pipeline
func (rp *RequestProcessor) ProcessRequest(event []byte) (*ExecuteTurn1Request, error) {
	// Parse
	req, err := rp.parser.ParseRequestFromJSON(event)
	if err != nil {
		return nil, err
	}

	// Validate
	if err := rp.validator.ValidateRequest(req); err != nil {
		return nil, err
	}

	// Sanitize
	if err := sanitizeRequest(req, rp.logger); err != nil {
		return nil, err
	}

	return req, nil
}

// ProcessResponse handles response creation with appropriate error handling
func (rp *RequestProcessor) ProcessResponse(state *schema.WorkflowState, err error) *ExecuteTurn1Response {
	if err != nil {
		status := determineErrorStatus(err, rp.logger)
		wfErr := rp.builder.EnsureWorkflowError(err)
		
		if state != nil {
			return rp.builder.BuildErrorResponse(state, wfErr, status)
		} else {
			return rp.builder.BuildErrorResponseWithoutState(wfErr, status)
		}
	}
	
	return rp.builder.BuildSuccessResponse(state)
}

// Helper functions

// sanitizeRequest performs basic sanitization of the request
func sanitizeRequest(req *ExecuteTurn1Request, log logger.Logger) error {
	if req == nil {
		return wferrors.NewValidationError("Cannot sanitize nil request", nil)
	}

	// Ensure schema version is current
	if req.WorkflowState.SchemaVersion != schema.SchemaVersion {
		log.Debug("Updating schema version", map[string]interface{}{
			"from": req.WorkflowState.SchemaVersion,
			"to":   schema.SchemaVersion,
			"verificationId": req.GetVerificationID(),
		})
		req.WorkflowState.SchemaVersion = schema.SchemaVersion
	}

	// Initialize conversation state if missing
	if req.WorkflowState.ConversationState == nil {
		log.Debug("Initializing conversation state", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		req.WorkflowState.ConversationState = &schema.ConversationState{
			CurrentTurn: 0,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}

	return nil
}

// determineErrorStatus determines the appropriate status based on error type
func determineErrorStatus(err error, log logger.Logger) string {
	if wfErr, ok := err.(*wferrors.WorkflowError); ok {
		switch wfErr.Type {
		case wferrors.ErrorTypeValidation:
			return schema.StatusBedrockProcessingFailed // Validation errors prevent processing
		case wferrors.ErrorTypeBedrock:
			return schema.StatusBedrockProcessingFailed
		case wferrors.ErrorTypeTimeout:
			return schema.StatusBedrockProcessingFailed
		case wferrors.ErrorTypeInternal:
			return schema.StatusBedrockProcessingFailed
		default:
			return schema.StatusBedrockProcessingFailed
		}
	}

	log.Debug("Unknown error type, using default status", map[string]interface{}{
		"error": err.Error(),
		"errorType": fmt.Sprintf("%T", err),
	})
	return schema.StatusBedrockProcessingFailed
}