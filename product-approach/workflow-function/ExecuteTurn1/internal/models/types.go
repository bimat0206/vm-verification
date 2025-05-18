package models

import (
	"fmt"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
)

// ExecuteTurn1Request wraps the incoming Step Functions payload for ExecuteTurn1.
type ExecuteTurn1Request struct {
	WorkflowState schema.WorkflowState `json:"workflowState" validate:"required"`
}

// ExecuteTurn1Response is returned back to Step Functions.
type ExecuteTurn1Response struct {
	WorkflowState schema.WorkflowState    `json:"workflowState"`
	Error         *wferrors.WorkflowError `json:"error,omitempty"`
}

// Safe field access helpers - prevent nil pointer dereferences

// GetVerificationID safely extracts the verification ID
func (req *ExecuteTurn1Request) GetVerificationID() string {
	if req == nil || req.WorkflowState.VerificationContext == nil {
		return ""
	}
	return req.WorkflowState.VerificationContext.VerificationId
}

// GetPromptID safely extracts the prompt ID
func (req *ExecuteTurn1Request) GetPromptID() string {
	if req == nil || req.WorkflowState.CurrentPrompt == nil {
		return ""
	}
	return req.WorkflowState.CurrentPrompt.PromptId
}

// GetSchemaVersion safely extracts the schema version
func (req *ExecuteTurn1Request) GetSchemaVersion() string {
	if req == nil {
		return ""
	}
	return req.WorkflowState.SchemaVersion
}

// GetCurrentStatus safely extracts the current status
func (req *ExecuteTurn1Request) GetCurrentStatus() string {
	if req == nil || req.WorkflowState.VerificationContext == nil {
		return ""
	}
	return req.WorkflowState.VerificationContext.Status
}

// HasImages checks if the request contains image data
func (req *ExecuteTurn1Request) HasImages() bool {
	if req == nil || req.WorkflowState.Images == nil {
		return false
	}
	return req.WorkflowState.Images.Reference != nil || req.WorkflowState.Images.ReferenceImage != nil
}

// HasValidStructure checks if the request has the minimum required structure
func (req *ExecuteTurn1Request) HasValidStructure() bool {
	return req != nil &&
		req.WorkflowState.VerificationContext != nil &&
		req.WorkflowState.CurrentPrompt != nil &&
		req.WorkflowState.BedrockConfig != nil
}

// GetDebugInfo safely collects debug information about the request structure
func (req *ExecuteTurn1Request) GetDebugInfo() map[string]interface{} {
	if req == nil {
		return map[string]interface{}{"error": "request is nil"}
	}

	info := map[string]interface{}{
		"hasVerificationContext": req.WorkflowState.VerificationContext != nil,
		"hasCurrentPrompt":       req.WorkflowState.CurrentPrompt != nil,
		"hasBedrockConfig":       req.WorkflowState.BedrockConfig != nil,
		"hasImages":              req.HasImages(),
		"hasConversationState":   req.WorkflowState.ConversationState != nil,
		"schemaVersion":          req.WorkflowState.SchemaVersion,
	}

	// Add verification context details if available
	if req.WorkflowState.VerificationContext != nil {
		info["verificationId"] = req.WorkflowState.VerificationContext.VerificationId
		info["status"] = req.WorkflowState.VerificationContext.Status
		info["verificationAt"] = req.WorkflowState.VerificationContext.VerificationAt
	}

	// Add prompt details if available
	if req.WorkflowState.CurrentPrompt != nil {
		info["promptId"] = req.WorkflowState.CurrentPrompt.PromptId
		info["messageCount"] = len(req.WorkflowState.CurrentPrompt.Messages)
	}

	// Add bedrock config details if available
	if req.WorkflowState.BedrockConfig != nil {
		info["maxTokens"] = req.WorkflowState.BedrockConfig.MaxTokens
		info["anthropicVersion"] = req.WorkflowState.BedrockConfig.AnthropicVersion
	}

	return info
}

// String implementations for better logging

// String implements the Stringer interface for ExecuteTurn1Request
func (req *ExecuteTurn1Request) String() string {
	if req == nil {
		return "ExecuteTurn1Request{nil}"
	}
	return fmt.Sprintf("ExecuteTurn1Request{VerificationID: %s, PromptID: %s, HasImages: %t}",
		req.GetVerificationID(),
		req.GetPromptID(),
		req.HasImages(),
	)
}

// String implements the Stringer interface for ExecuteTurn1Response
func (resp *ExecuteTurn1Response) String() string {
	if resp == nil {
		return "ExecuteTurn1Response{nil}"
	}

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