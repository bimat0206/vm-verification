package s3state

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Reference represents an S3 object reference
type Reference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Size   int64  `json:"size,omitempty"`
}

// Envelope represents the state envelope passed between Step Functions
type Envelope struct {
	VerificationID string               `json:"verificationId"`
	References     map[string]*Reference `json:"s3References"`
	Status         string               `json:"status"`
	Summary        map[string]interface{} `json:"summary,omitempty"`
}

// NewReference creates a new S3 reference
func NewReference(bucket, key string, size int64) *Reference {
	return &Reference{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}
}

// NewEnvelope creates a new envelope for a verification
func NewEnvelope(verificationID string) *Envelope {
	return &Envelope{
		VerificationID: verificationID,
		References:     make(map[string]*Reference),
		Status:         "INITIALIZED",
		Summary:        make(map[string]interface{}),
	}
}

// IsValid checks if the reference has required fields
func (r *Reference) IsValid() bool {
	return r != nil && r.Bucket != "" && r.Key != ""
}

// GetCategory extracts the category from the S3 key
func (r *Reference) GetCategory() string {
	if r == nil || r.Key == "" {
		return ""
	}
	
	parts := strings.Split(r.Key, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	
	return ""
}

// GetFilename extracts the filename from the S3 key
func (r *Reference) GetFilename() string {
	if r == nil || r.Key == "" {
		return ""
	}
	
	parts := strings.Split(r.Key, "/")
	if len(parts) >= 1 {
		return parts[len(parts)-1]
	}
	
	return ""
}

// String returns a string representation of the reference
func (r *Reference) String() string {
	if r == nil {
		return "<nil reference>"
	}
	return fmt.Sprintf("s3://%s/%s", r.Bucket, r.Key)
}

// AddReference adds a reference to the envelope
func (e *Envelope) AddReference(name string, ref *Reference) {
	if e.References == nil {
		e.References = make(map[string]*Reference)
	}
	e.References[name] = ref
}

// GetReference gets a reference from the envelope by name
func (e *Envelope) GetReference(name string) *Reference {
	if e.References == nil {
		return nil
	}
	return e.References[name]
}

// HasReference checks if a reference exists in the envelope
func (e *Envelope) HasReference(name string) bool {
	return e.GetReference(name) != nil
}

// SetStatus updates the envelope status
func (e *Envelope) SetStatus(status string) {
	e.Status = status
}

// AddSummary adds a key-value pair to the summary
func (e *Envelope) AddSummary(key string, value interface{}) {
	if e.Summary == nil {
		e.Summary = make(map[string]interface{})
	}
	e.Summary[key] = value
}

// GetSummary gets a value from the summary
func (e *Envelope) GetSummary(key string) interface{} {
	if e.Summary == nil {
		return nil
	}
	return e.Summary[key]
}

// ToJSON converts the envelope to JSON bytes
func (e *Envelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an envelope from JSON bytes
func FromJSON(data []byte) (*Envelope, error) {
	var envelope Envelope
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}
	return &envelope, nil
}

// LoadEnvelope loads an envelope from Step Functions input
func LoadEnvelope(input interface{}) (*Envelope, error) {
	// Handle different input types
	switch v := input.(type) {
	case *Envelope:
		return v, nil
	case Envelope:
		return &v, nil
	case map[string]interface{}:
		// Convert map to JSON then to Envelope
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		return FromJSON(data)
	case []byte:
		return FromJSON(v)
	case string:
		return FromJSON([]byte(v))
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}
}

// GetReferencesByCategory returns all references that belong to a specific category
func (e *Envelope) GetReferencesByCategory(category string) map[string]*Reference {
	result := make(map[string]*Reference)
	
	for name, ref := range e.References {
		if ref != nil && ref.GetCategory() == category {
			result[name] = ref
		}
	}
	
	return result
}

// ListCategories returns all unique categories present in the envelope
func (e *Envelope) ListCategories() []string {
	categories := make(map[string]bool)
	
	for _, ref := range e.References {
		if ref != nil {
			category := ref.GetCategory()
			if category != "" {
				categories[category] = true
			}
		}
	}
	
	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	
	return result
}

// Validate checks if the envelope has valid structure
func (e *Envelope) Validate() error {
	if e == nil {
		return fmt.Errorf("envelope is nil")
	}
	
	if e.VerificationID == "" {
		return fmt.Errorf("verification ID is required")
	}
	
	// Validate all references
	for name, ref := range e.References {
		if !ref.IsValid() {
			return fmt.Errorf("invalid reference '%s': missing bucket or key", name)
		}
	}
	
	return nil
}

// Clone creates a deep copy of the envelope
func (e *Envelope) Clone() *Envelope {
	if e == nil {
		return nil
	}
	
	clone := &Envelope{
		VerificationID: e.VerificationID,
		Status:         e.Status,
		References:     make(map[string]*Reference),
		Summary:        make(map[string]interface{}),
	}
	
	// Copy references
	for name, ref := range e.References {
		if ref != nil {
			clone.References[name] = &Reference{
				Bucket: ref.Bucket,
				Key:    ref.Key,
				Size:   ref.Size,
			}
		}
	}
	
	// Copy summary
	for key, value := range e.Summary {
		clone.Summary[key] = value
	}
	
	return clone
}