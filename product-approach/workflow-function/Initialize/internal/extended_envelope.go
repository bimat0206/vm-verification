package internal

import (
	"encoding/json"
	
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// ExtendedEnvelope extends s3state.Envelope with additional fields needed for Step Functions
type ExtendedEnvelope struct {
	*s3state.Envelope
	VerificationContext *schema.VerificationContext `json:"verificationContext,omitempty"`
}

// NewExtendedEnvelope creates a new extended envelope wrapping an s3state.Envelope
func NewExtendedEnvelope(envelope *s3state.Envelope) *ExtendedEnvelope {
	return &ExtendedEnvelope{
		Envelope: envelope,
	}
}

// MarshalJSON provides custom JSON marshaling to combine the fields
func (e *ExtendedEnvelope) MarshalJSON() ([]byte, error) {
	// Create a map to hold all fields
	combined := make(map[string]interface{})
	
	// Marshal the base envelope
	baseBytes, err := json.Marshal(e.Envelope)
	if err != nil {
		return nil, err
	}
	
	// Unmarshal into the map
	err = json.Unmarshal(baseBytes, &combined)
	if err != nil {
		return nil, err
	}
	
	// Add the verification context if present
	if e.VerificationContext != nil {
		combined["verificationContext"] = e.VerificationContext
	}
	
	// Marshal the combined map
	return json.Marshal(combined)
}