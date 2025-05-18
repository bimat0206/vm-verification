package models

import (
	"time"
	"strings"
	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// Sanitizer handles request sanitization and normalization
type Sanitizer struct {
	logger logger.Logger
}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer(log logger.Logger) *Sanitizer {
	return &Sanitizer{
		logger: log.WithFields(map[string]interface{}{
			"component": "RequestSanitizer",
		}),
	}
}

// SanitizeRequest performs sanitization and normalization of the request
func (s *Sanitizer) SanitizeRequest(req *ExecuteTurn1Request) error {
	if req == nil {
		s.logger.Error("Cannot sanitize nil request", nil)
		return wferrors.NewValidationError("Cannot sanitize nil request", nil)
	}

	s.logger.Debug("Starting request sanitization", map[string]interface{}{
		"verificationId": req.GetVerificationID(),
	})

	// Update schema version to current
	s.ensureCurrentSchemaVersion(req)

	// Normalize timestamps
	s.normalizeTimestamps(req)

	// Initialize optional structures
	s.initializeOptionalStructures(req)

	// Clean up string fields
	s.cleanStringFields(req)

	s.logger.Debug("Request sanitization completed", map[string]interface{}{
		"verificationId": req.GetVerificationID(),
		"schemaVersion":  req.GetSchemaVersion(),
	})

	return nil
}

// ensureCurrentSchemaVersion updates the schema version to current if needed
func (s *Sanitizer) ensureCurrentSchemaVersion(req *ExecuteTurn1Request) {
	currentVersion := req.GetSchemaVersion()
	if currentVersion != schema.SchemaVersion {
		s.logger.Debug("Updating schema version", map[string]interface{}{
			"from":           currentVersion,
			"to":             schema.SchemaVersion,
			"verificationId": req.GetVerificationID(),
		})
		req.WorkflowState.SchemaVersion = schema.SchemaVersion
	}
}

// normalizeTimestamps ensures timestamps are in the correct format
func (s *Sanitizer) normalizeTimestamps(req *ExecuteTurn1Request) {
	if req.WorkflowState.VerificationContext != nil {
		// Set VerificationAt if missing or empty
		if req.WorkflowState.VerificationContext.VerificationAt == "" {
			timestamp := time.Now().UTC().Format(time.RFC3339)
			s.logger.Debug("Setting default VerificationAt timestamp", map[string]interface{}{
				"timestamp":      timestamp,
				"verificationId": req.GetVerificationID(),
			})
			req.WorkflowState.VerificationContext.VerificationAt = timestamp
		}
	}
}

// initializeOptionalStructures creates missing optional structures with sensible defaults
func (s *Sanitizer) initializeOptionalStructures(req *ExecuteTurn1Request) {
	// Initialize conversation state if missing
	if req.WorkflowState.ConversationState == nil {
		s.logger.Debug("Initializing conversation state", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		req.WorkflowState.ConversationState = &schema.ConversationState{
			CurrentTurn: 0,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}

	// Ensure history is initialized (not nil)
	if req.WorkflowState.ConversationState.History == nil {
		s.logger.Debug("Initializing conversation history", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
		})
		req.WorkflowState.ConversationState.History = []interface{}{}
	}

	// Set default MaxTurns if not set or invalid
	if req.WorkflowState.ConversationState.MaxTurns <= 0 {
		s.logger.Debug("Setting default MaxTurns", map[string]interface{}{
			"verificationId": req.GetVerificationID(),
			"maxTurns":       2,
		})
		req.WorkflowState.ConversationState.MaxTurns = 2
	}
}

// cleanStringFields trims whitespace from string fields and handles empty strings
func (s *Sanitizer) cleanStringFields(req *ExecuteTurn1Request) {
	// Clean verification context strings
	if req.WorkflowState.VerificationContext != nil {
		// Note: We don't modify VerificationId as it should not be changed
		// Just ensure status is properly formatted
		if req.WorkflowState.VerificationContext.Status != "" {
			// Status is typically set by the system, but ensure it's trimmed
			req.WorkflowState.VerificationContext.Status = 
				normalizeString(req.WorkflowState.VerificationContext.Status)
		}
	}

	// Clean prompt strings
	if req.WorkflowState.CurrentPrompt != nil {
		req.WorkflowState.CurrentPrompt.PromptId = 
			normalizeString(req.WorkflowState.CurrentPrompt.PromptId)

		// Clean message content
		for i := range req.WorkflowState.CurrentPrompt.Messages {
			msg := &req.WorkflowState.CurrentPrompt.Messages[i]
			msg.Role = normalizeString(msg.Role)
			
			// Clean text content in message
			for j := range msg.Content {
				content := &msg.Content[j]
				content.Text = normalizeString(content.Text)
			}
		}
	}

	// Clean Bedrock config strings
	if req.WorkflowState.BedrockConfig != nil {
		req.WorkflowState.BedrockConfig.AnthropicVersion = 
			normalizeString(req.WorkflowState.BedrockConfig.AnthropicVersion)
	}
}

// normalizeString trims whitespace and handles empty strings consistently
func normalizeString(s string) string {
	// Trim leading and trailing whitespace
	trimmed := strings.TrimSpace(s)
	return trimmed
}

// SanitizeAndValidate performs both sanitization and validation in the correct order
func (s *Sanitizer) SanitizeAndValidate(req *ExecuteTurn1Request) error {
	// First sanitize
	if err := s.SanitizeRequest(req); err != nil {
		return err
	}

	// Then validate
	validator := NewValidator(s.logger)
	return validator.ValidateRequest(req)
}

// Backward compatibility method - maintains existing API
func (req *ExecuteTurn1Request) ValidateAndSanitize(log logger.Logger) error {
	// First validate (existing behavior)
	if err := req.Validate(log); err != nil {
		return err
	}

	// Then sanitize
	sanitizer := NewSanitizer(log)
	return sanitizer.SanitizeRequest(req)
}

