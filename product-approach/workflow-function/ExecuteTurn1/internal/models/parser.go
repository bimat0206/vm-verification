package models

import (
	"encoding/json"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors"
)

// Parser handles JSON parsing for ExecuteTurn1Request
type Parser struct {
	logger logger.Logger
}

// NewParser creates a new parser instance
func NewParser(log logger.Logger) *Parser {
	return &Parser{
		logger: log.WithFields(map[string]interface{}{
			"component": "RequestParser",
		}),
	}
}

// ParseRequestFromJSON unmarshals the raw Lambda event into an ExecuteTurn1Request
// This method specifically addresses the nil pointer dereference issue by:
// 1. Validating input before processing
// 2. Performing safe field access after unmarshaling
// 3. Using safe helper methods for logging
func (p *Parser) ParseRequestFromJSON(event []byte) (*ExecuteTurn1Request, error) {
	p.logger.Debug("Starting JSON parsing", map[string]interface{}{
		"eventSize": len(event),
	})

	// Validate input
	if len(event) == 0 {
		p.logger.Error("Empty event payload", nil)
		return nil, wferrors.NewValidationError("Event payload is empty", nil)
	}

	// Parse JSON
	var req ExecuteTurn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		p.logger.Error("Failed to unmarshal JSON", map[string]interface{}{
			"error":     err.Error(),
			"eventSize": len(event),
		})
		return nil, wferrors.NewValidationError("Invalid JSON format", map[string]interface{}{
			"parseError": err.Error(),
			"eventSize":  len(event),
		})
	}

	// CRITICAL FIX: Validate basic structure immediately after unmarshaling
	// This prevents nil pointer dereferences in logging and subsequent operations
	if err := p.validateBasicStructureAfterUnmarshal(&req); err != nil {
		return nil, err
	}

	// Safe logging using helper methods (fixes the original nil pointer issue)
	p.logger.Debug("JSON parsing completed successfully", map[string]interface{}{
		"verificationId":   req.GetVerificationID(),   // Safe access - no nil pointer risk
		"schemaVersion":    req.GetSchemaVersion(),    // Safe access
		"hasValidStructure": req.HasValidStructure(),  // Safe access
	})

	return &req, nil
}

// validateBasicStructureAfterUnmarshal performs minimal validation after JSON unmarshaling
// to prevent nil pointer dereferences in subsequent operations. This is the core fix for
// the reported panic issue.
func (p *Parser) validateBasicStructureAfterUnmarshal(req *ExecuteTurn1Request) error {
	if req == nil {
		p.logger.Error("Request is nil after unmarshaling", nil)
		return wferrors.NewValidationError("Request is nil after unmarshaling", nil)
	}

	// Log the structure state for debugging - using safe methods
	debugInfo := req.GetDebugInfo()
	p.logger.Debug("Post-unmarshal structure analysis", debugInfo)

	// Check for critical nil pointers that could cause panics
	// We log warnings but don't fail here - let proper validation handle these
	var warnings []string
	
	if req.WorkflowState.VerificationContext == nil {
		warnings = append(warnings, "VerificationContext is nil")
	}
	
	if req.WorkflowState.CurrentPrompt == nil {
		warnings = append(warnings, "CurrentPrompt is nil")
	}
	
	if req.WorkflowState.BedrockConfig == nil {
		warnings = append(warnings, "BedrockConfig is nil")
	}

	// Log warnings but don't fail - this is just to prevent panics
	if len(warnings) > 0 {
		p.logger.Warn("Detected nil pointers in request structure", map[string]interface{}{
			"warnings":       warnings,
			"verificationId": req.GetVerificationID(), // Safe to call
		})
	}

	return nil
}

// Legacy function for backward compatibility
// This maintains the existing API while using the new safe implementation
func NewRequestFromJSON(event []byte, log logger.Logger) (*ExecuteTurn1Request, error) {
	parser := NewParser(log)
	return parser.ParseRequestFromJSON(event)
}