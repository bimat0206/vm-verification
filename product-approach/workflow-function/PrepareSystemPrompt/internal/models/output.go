package models

import (
	"encoding/json"
	"time"
	
	"workflow-function/shared/schema"
	"workflow-function/shared/s3state"
)

// Response represents the Lambda response
type Response struct {
	VerificationID   string                       `json:"verificationId"`
	VerificationDate string                       `json:"verificationDate"`
	S3References     map[string]*s3state.Reference `json:"s3References"`
	Status           string                       `json:"status"`
	Summary          map[string]interface{}       `json:"summary,omitempty"`
}

// BuildResponse creates a response from an S3 state envelope
func BuildResponse(envelope *s3state.Envelope, datePartition string) *Response {
	resp := &Response{
		VerificationID:   envelope.VerificationID,
		VerificationDate: datePartition,
		S3References:     envelope.References,
		Status:           envelope.Status,
		Summary:          envelope.Summary,
	}
	
	// Default summary if none provided
	if resp.Summary == nil {
		resp.Summary = make(map[string]interface{})
	}
	
	return resp
}

// ToJSON converts the response to JSON
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// SystemPromptReference holds metadata about a stored system prompt
type SystemPromptReference struct {
	Bucket        string              `json:"bucket"`
	Key           string              `json:"key"`
	ETag          string              `json:"etag,omitempty"`
	Size          int64               `json:"size,omitempty"`
	PromptID      string              `json:"promptId"`
	PromptVersion string              `json:"promptVersion"`
	BedrockConfig *schema.BedrockConfig `json:"bedrockConfig,omitempty"`
}

// CreateSummary creates a summary object for the response
func CreateSummary(prompt *schema.SystemPrompt, verificationType string, processingTimeMs int64, modelId string, anthropicVersion string) map[string]interface{} {
	return map[string]interface{}{
		"promptType":        verificationType,
		"estimatedTokens":   len(prompt.Content) / 4, // Rough estimate
		"processingTimeMs":  processingTimeMs,
		"promptTimestamp":   time.Now().UTC().Format(time.RFC3339),
		"promptVersion":     prompt.PromptVersion,
		"modelId":           modelId,
		"anthropicVersion":  anthropicVersion,
	}
}

// CreateCompleteSummary creates a complete summary object from CompleteSystemPrompt
func CreateCompleteSummary(prompt *schema.CompleteSystemPrompt, verificationType string, processingTimeMs int64) map[string]interface{} {
	// Update processing metadata with actual processing time
	prompt.ProcessingMetadata.GenerationTimeMs = processingTimeMs
	
	// Convert the complete system prompt to a map for the summary
	promptBytes, _ := json.Marshal(prompt)
	var promptMap map[string]interface{}
	json.Unmarshal(promptBytes, &promptMap)
	
	return promptMap
}

// AddReferencesToEnvelope adds system prompt references to an envelope
func AddReferencesToEnvelope(
	envelope *s3state.Envelope, 
	systemPromptRef *s3state.Reference,
) {
	// Add system prompt reference
	envelope.AddReference("prompts_system", systemPromptRef)
}

// BuildResponseWithContext creates a response with verification context
func BuildResponseWithContext(
	envelope *s3state.Envelope, 
	verificationContext *schema.VerificationContext,
	datePartition string,
) *Response {
	resp := BuildResponse(envelope, datePartition)
	
	// Only add verification context data if summary doesn't already contain complete structure
	if _, hasVerificationId := resp.Summary["verificationId"]; !hasVerificationId {
		// This is a simple summary, add verification context data
		resp.Summary["verificationType"] = verificationContext.VerificationType
		resp.Summary["vendingMachineId"] = verificationContext.VendingMachineId
	}
	// If summary already has verificationId, it's a complete system prompt structure, don't modify it
	
	return resp
}
