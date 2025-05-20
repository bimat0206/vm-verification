package state

import (
	"workflow-function/shared/errors"
	"workflow-function/shared/s3state"
)

// Reference categories for S3 state management
const (
	CategoryInitialization = "initialization"
	CategoryImages         = "images"
	CategoryProcessing     = "processing"
	CategoryPrompts        = "prompts"
	CategoryResponses      = "responses"
)

// Reference keys for standard file names
const (
	KeyInitialization = "initialization.json"
	KeyMetadata       = "metadata.json"
	KeyTurn1Prompt    = "turn1-prompt.json"
	KeyTurn1Metrics   = "turn1-metrics.json"
)

// Input represents the Lambda function input with S3 references
type Input struct {
	References           map[string]*s3state.Reference `json:"references"`
	S3References         map[string]*s3state.Reference `json:"s3References"` // Alternative field name used in Step Functions
	VerificationID       string                        `json:"verificationId"`
	VerificationType     string                        `json:"verificationType"`
	TurnNumber           int                           `json:"turnNumber"`
	IncludeImage         string                        `json:"includeImage"`
	EnableS3StateManager bool                          `json:"enableS3StateManager"`
	Status               string                        `json:"status"`
	SchemaVersion        string                        `json:"schemaVersion"`
	Summary              map[string]interface{}        `json:"summary"` // Summary field from PrepareSystemPrompt
}

// Output represents the Lambda function output with S3 references
type Output struct {
	References       map[string]*s3state.Reference `json:"references"`
	VerificationID   string                        `json:"verificationId"`
	VerificationType string                        `json:"verificationType"`
	Status           string                        `json:"status"`
}

// GetReferenceKey builds a standard reference key for accessing references
func GetReferenceKey(category, dataType string) string {
	return category + "_" + dataType
}

// ValidateReferences checks if required references exist in the input
func ValidateReferences(input *Input) error {
	if input == nil {
		return errors.NewValidationError("Input is nil", nil)
	}

	// If S3References is present but References is not, use S3References
	if input.References == nil && input.S3References != nil {
		input.References = input.S3References
	}

	if input.References == nil {
		return errors.NewValidationError("References map is nil", nil)
	}

	// Required references for Turn 1 processing
	requiredRefs := []string{
		GetReferenceKey(CategoryProcessing, "initialization"),
		GetReferenceKey(CategoryPrompts, "system"),
	}

	// Try alternative prefixes
	alternativeRefs := map[string][]string{
		GetReferenceKey(CategoryProcessing, "initialization"): {
			"initialization_initialization",
			"processing_initialization",
		},
		GetReferenceKey(CategoryPrompts, "system"): {
			"prompts_system",
			"prompts_system_prompt",
			"prompts_system-prompt",
		},
	}

	missingRefs := make([]string, 0)
	for _, refKey := range requiredRefs {
		// Check the primary key
		if _, exists := input.References[refKey]; exists {
			continue
		}

		// Check alternative keys
		found := false
		for _, altKey := range alternativeRefs[refKey] {
			if _, exists := input.References[altKey]; exists {
				found = true
				break
			}
		}

		if !found {
			missingRefs = append(missingRefs, refKey)
		}
	}

	if len(missingRefs) > 0 {
		return errors.NewValidationError("Missing required references", 
			map[string]interface{}{
				"missing": missingRefs,
			})
	}

	return nil
}

// CopyReferences is a helper function to copy references from one map to another
func CopyReferences(dest, src map[string]*s3state.Reference) {
	if dest == nil || src == nil {
		return
	}
	
	for k, v := range src {
		dest[k] = v
	}
}

// NewOutput creates a new output with initialized references map
// Modified to accept and preserve existing references
func NewOutput(verificationID, verificationType, status string, existingRefs map[string]*s3state.Reference) *Output {
	out := &Output{
		References:       make(map[string]*s3state.Reference),
		VerificationID:   verificationID,
		VerificationType: verificationType,
		Status:           status,
	}
	
	// Preserve all existing references
	if existingRefs != nil {
		CopyReferences(out.References, existingRefs)
	}
	
	return out
}

// AddReference adds a reference to the output
func (o *Output) AddReference(category, dataType string, ref *s3state.Reference) {
	if o.References == nil {
		o.References = make(map[string]*s3state.Reference)
	}
	o.References[GetReferenceKey(category, dataType)] = ref
}

// EnvelopeToInput converts an S3 state envelope to input format
// Enhanced to properly handle references
func EnvelopeToInput(envelope *s3state.Envelope) (*Input, error) {
	if envelope == nil {
		return nil, errors.NewValidationError("Envelope is nil", nil)
	}

	input := &Input{
		References:           make(map[string]*s3state.Reference),
		S3References:         make(map[string]*s3state.Reference),
		VerificationID:       envelope.VerificationID,
		EnableS3StateManager: true,
		TurnNumber:           1,                    // Always set to 1 for PrepareTurn1Prompt
		IncludeImage:         "reference",          // Default to "reference" for Turn 1
		Status:               envelope.Status,      // Use status from envelope
	}
	
	// Ensure proper copying of all references to both maps for compatibility
	if envelope.References != nil {
		for k, v := range envelope.References {
			input.References[k] = v
			input.S3References[k] = v
		}
	}

	// Check for available references to determine if we can proceed
	hasInitRef := false
	
	// Check various possible keys for the initialization reference
	possibleInitKeys := []string{
		GetReferenceKey(CategoryProcessing, "initialization"),
		GetReferenceKey(CategoryInitialization, "initialization"),
		"initialization_initialization",
		"processing_initialization",
	}
	
	for _, key := range possibleInitKeys {
		if ref, exists := envelope.References[key]; exists && ref != nil {
			hasInitRef = true
			break
		}
	}
	
	if !hasInitRef {
		return nil, errors.NewValidationError("Initialization reference not found or nil", 
			map[string]interface{}{
				"availableRefs": envelope.References,
			})
	}

	return input, nil
}

// OutputToEnvelope converts output to an S3 state envelope
func OutputToEnvelope(output *Output) *s3state.Envelope {
	return &s3state.Envelope{
		References:     output.References,
		VerificationID: output.VerificationID,
		Status:         output.Status,
	}
}