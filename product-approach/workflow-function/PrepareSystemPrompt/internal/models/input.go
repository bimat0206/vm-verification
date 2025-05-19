package models

import (
	"encoding/json"
	
	"workflow-function/shared/schema"
	"workflow-function/shared/s3state"
)

// InputType represents the type of input received by the function
type InputType string

const (
	// InputTypeS3Reference indicates the input contains S3 references to load state
	InputTypeS3Reference InputType = "S3_REFERENCE"
	
	// InputTypeDirectJSON indicates the input contains direct JSON with verification context
	InputTypeDirectJSON InputType = "DIRECT_JSON"
)

// Input adapts the Lambda input event to internal structures
type Input struct {
	Type                InputType                 `json:"-"`
	S3Envelope          *s3state.Envelope         `json:"-"`
	VerificationContext *schema.VerificationContext `json:"verificationContext"`
	LayoutMetadata      map[string]interface{}    `json:"layoutMetadata,omitempty"`
	HistoricalContext   map[string]interface{}    `json:"historicalContext,omitempty"`
	Images              *schema.ImageData         `json:"images,omitempty"`
	TurnNumber          int                       `json:"turnNumber,omitempty"`
	IncludeImage        string                    `json:"includeImage,omitempty"`
}

// S3ReferenceInput represents an input with S3 references to state
type S3ReferenceInput struct {
	VerificationID string               `json:"verificationId"`
	S3References   map[string]*s3state.Reference `json:"s3References"`
	DatePartition  string               `json:"datePartition,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Input
func (i *Input) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal as S3 reference envelope
	var s3Input S3ReferenceInput
	if err := json.Unmarshal(data, &s3Input); err == nil && s3Input.VerificationID != "" && s3Input.S3References != nil {
		// This is an S3 reference input
		i.Type = InputTypeS3Reference
		i.S3Envelope = &s3state.Envelope{
			VerificationID: s3Input.VerificationID,
			References:     s3Input.S3References,
		}
		return nil
	}
	
	// If not S3 reference, try direct JSON
	var temp struct {
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
		LayoutMetadata      json.RawMessage            `json:"layoutMetadata,omitempty"`
		HistoricalContext   json.RawMessage            `json:"historicalContext,omitempty"`
		Images              *schema.ImageData          `json:"images,omitempty"`
		TurnNumber          int                        `json:"turnNumber,omitempty"`
		IncludeImage        string                     `json:"includeImage,omitempty"`
	}
	
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// This is a direct JSON input
	i.Type = InputTypeDirectJSON
	i.VerificationContext = temp.VerificationContext
	i.Images = temp.Images
	i.TurnNumber = temp.TurnNumber
	i.IncludeImage = temp.IncludeImage
	
	// Parse layout metadata if present
	if len(temp.LayoutMetadata) > 0 {
		var layoutMetadata map[string]interface{}
		if err := json.Unmarshal(temp.LayoutMetadata, &layoutMetadata); err != nil {
			return err
		}
		i.LayoutMetadata = layoutMetadata
	}
	
	// Parse historical context if present
	if len(temp.HistoricalContext) > 0 {
		var historicalContext map[string]interface{}
		if err := json.Unmarshal(temp.HistoricalContext, &historicalContext); err != nil {
			return err
		}
		i.HistoricalContext = historicalContext
	}
	
	return nil
}

// CreateWorkflowState creates a workflow state from the input
func (i *Input) CreateWorkflowState() *schema.WorkflowState {
	if i.Type == InputTypeS3Reference && i.S3Envelope != nil {
		// For S3 reference inputs, the state will be loaded from S3
		// Return nil to indicate this
		return nil
	}
	
	// For direct JSON inputs, create a new workflow state
	state := &schema.WorkflowState{
		SchemaVersion:      "1.0.0",
		VerificationContext: i.VerificationContext,
		Images:             i.Images,
	}
	
	// Add layout metadata if present
	if i.LayoutMetadata != nil {
		state.LayoutMetadata = i.LayoutMetadata
	}
	
	// Add historical context if present
	if i.HistoricalContext != nil {
		state.HistoricalContext = i.HistoricalContext
	}
	
	return state
}

// GetVerificationID returns the verification ID from the input
func (i *Input) GetVerificationID() string {
	if i.Type == InputTypeS3Reference && i.S3Envelope != nil {
		return i.S3Envelope.VerificationID
	}
	
	if i.VerificationContext != nil {
		return i.VerificationContext.VerificationId
	}
	
	return ""
}