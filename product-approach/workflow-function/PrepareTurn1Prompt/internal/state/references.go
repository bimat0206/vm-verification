package state

import (
	"strings"
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
	KeyInitialization = "initialization"
	KeyMetadata       = "metadata"
	KeyTurn1Prompt    = "turn1-prompt"
	KeyTurn1Metrics   = "turn1-metrics"
)

// Input represents the Lambda function input with S3 references
type Input struct {
	References           map[string]*s3state.Reference `json:"references"`
	S3References         map[string]*s3state.Reference `json:"s3References"` // Main field name used in Step Functions
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
	S3References     map[string]*s3state.Reference `json:"s3References"` // Changed to s3References for consistency
	VerificationID   string                        `json:"verificationId"`
	VerificationType string                        `json:"verificationType"`
	Status           string                        `json:"status"`
}

// GetReferenceKey builds a standard reference key for accessing references
func GetReferenceKey(category, dataType string) string {
	// Simple standardized format: category_datatype
	// Avoid including any paths or verification IDs in the reference key
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
		GetReferenceKey(CategoryInitialization, "initialization"),
		GetReferenceKey(CategoryPrompts, "system"),
	}

	// Try alternative prefixes
	alternativeRefs := map[string][]string{
		GetReferenceKey(CategoryInitialization, "initialization"): {
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
// Preserved existing references using s3References field instead of references
func NewOutput(verificationID, verificationType, status string, existingRefs map[string]*s3state.Reference) *Output {
	out := &Output{
		S3References:     make(map[string]*s3state.Reference), // Changed to S3References
		VerificationID:   verificationID,
		VerificationType: verificationType,
		Status:           status,
	}
	
	// Preserve all existing references
	if existingRefs != nil {
		CopyReferences(out.S3References, existingRefs) // Changed to S3References
	}
	
	return out
}

// AddReference adds a reference to the output
func (o *Output) AddReference(category, dataType string, ref *s3state.Reference) {
	if o.S3References == nil { // Changed to S3References
		o.S3References = make(map[string]*s3state.Reference) // Changed to S3References
	}
	// Use the standardized reference key format
	o.S3References[GetReferenceKey(category, dataType)] = ref // Changed to S3References
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
			// Standardize keys during copy to ensure consistency
			standardizedKey := standardizeReferenceKey(k)
			input.References[standardizedKey] = v
			input.S3References[standardizedKey] = v
		}
	}

	// Check for available references to determine if we can proceed
	hasInitRef := false
	
	// Check various possible keys for the initialization reference
	possibleInitKeys := []string{
		GetReferenceKey(CategoryInitialization, "initialization"),
		GetReferenceKey(CategoryProcessing, "initialization"),
		"initialization_initialization",
		"processing_initialization",
	}
	
	for _, key := range possibleInitKeys {
		if ref, exists := input.References[key]; exists && ref != nil {
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
		References:     output.S3References, // Changed to S3References
		VerificationID: output.VerificationID,
		Status:         output.Status,
	}
}

// standardizeReferenceKey helps standardize existing reference keys
// This is used to fix potentially malformed keys from existing state
func standardizeReferenceKey(key string) string {
	// Check if the key contains a path with verification ID
	if strings.Contains(key, "/") {
		// Handle potentially complex keys with embedded paths
		parts := strings.Split(key, "_")
		if len(parts) >= 2 {
			// Extract the category from the first part
			category := parts[0]
			
			// Extract the data type from the last path segment
			pathParts := strings.Split(parts[1], "/")
			dataType := strings.TrimSuffix(pathParts[len(pathParts)-1], ".json")
			
			// Return standardized key
			return GetReferenceKey(category, dataType)
		}
	}
	
	// If the key can't be standardized, return as is
	return key
}

// ValidateReferenceStructure ensures that the references contain necessary categories
func ValidateReferenceStructure(references map[string]*s3state.Reference) error {
	if references == nil {
		return errors.NewValidationError("References map is nil", nil)
	}
	
	// Check for critical reference categories
	categories := make(map[string]bool)
	for key := range references {
		parts := strings.Split(key, "_")
		if len(parts) > 0 {
			categories[parts[0]] = true
		}
	}
	
	// Check for required categories
	requiredCategories := []string{
		CategoryInitialization,
		CategoryImages,
		CategoryPrompts,
	}
	
	missingCategories := make([]string, 0)
	for _, category := range requiredCategories {
		if !categories[category] {
			missingCategories = append(missingCategories, category)
		}
	}
	
	if len(missingCategories) > 0 {
		return errors.NewValidationError("Missing required reference categories", 
			map[string]interface{}{
				"missing": missingCategories,
				"found": categories,
			})
	}
	
	return nil
}

// ValidateReferenceAccumulation checks if the output contains all references from the input
func ValidateReferenceAccumulation(input, output map[string]*s3state.Reference) error {
	if input == nil || output == nil {
		return errors.NewValidationError("Input or output references map is nil", nil)
	}
	
	// Check that all input references exist in output
	missingRefs := make([]string, 0)
	for key := range input {
		if _, exists := output[key]; !exists {
			missingRefs = append(missingRefs, key)
		}
	}
	
	if len(missingRefs) > 0 {
		return errors.NewValidationError("Output missing references from input", 
			map[string]interface{}{
				"missingRefs": missingRefs,
				"inputRefCount": len(input),
				"outputRefCount": len(output),
			})
	}
	
	return nil
}

// CategorizeReferences groups references by category for analysis
func CategorizeReferences(references map[string]*s3state.Reference) map[string][]string {
	categorized := make(map[string][]string)
	
	for key := range references {
		parts := strings.Split(key, "_")
		if len(parts) > 0 {
			category := parts[0]
			if _, exists := categorized[category]; !exists {
				categorized[category] = make([]string, 0)
			}
			categorized[category] = append(categorized[category], key)
		}
	}
	
	return categorized
}