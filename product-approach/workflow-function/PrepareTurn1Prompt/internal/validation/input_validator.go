package validation

import (
	"strings"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"prepare-turn1/internal/state"
)

// Validator handles input validation
type Validator struct {
	log logger.Logger
}

// NewValidator creates a new validator with the given logger
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		log: log,
	}
}

// ValidateInput validates the input parameters
func (v *Validator) ValidateInput(input *state.Input) error {
	// Basic validation
	if input == nil {
		return errors.NewValidationError("Input cannot be nil", nil)
	}

	// Validate required fields
	if input.VerificationID == "" {
		return errors.NewMissingFieldError("verificationId")
	}

	// Turn number validation (must be 1 for Turn 1)
	if input.TurnNumber != 1 {
		return errors.NewInvalidFieldError("turnNumber", input.TurnNumber, "1")
	}

	// Include image validation (must be "reference" for Turn 1)
	if input.IncludeImage != "reference" {
		return errors.NewInvalidFieldError("includeImage", input.IncludeImage, "reference")
	}

	// Validate verification type if available
	if input.VerificationType != "" {
		if err := v.validateVerificationType(input.VerificationType); err != nil {
			return err
		}
	}

	// Validate S3 references
	if err := state.ValidateReferences(input); err != nil {
		return err
	}
	
	// Log reference information for debugging
	if input.References != nil {
		referenceCategories := make(map[string]int)
		for key := range input.References {
			parts := strings.Split(key, "_")
			if len(parts) > 0 {
				category := parts[0]
				referenceCategories[category]++
			}
		}
		
		v.log.Info("Input references by category", map[string]interface{}{
			"categories": referenceCategories,
			"totalCount": len(input.References),
			"verificationId": input.VerificationID,
		})
	}

	return nil
}

// validateVerificationType validates that the verification type is supported
func (v *Validator) validateVerificationType(verificationType string) error {
	validTypes := []string{schema.VerificationTypeLayoutVsChecking, schema.VerificationTypePreviousVsCurrent}
	
	isValidType := false
	for _, vt := range validTypes {
		if verificationType == vt {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return errors.NewInvalidFieldError("verificationType", 
			verificationType, 
			"one of LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT")
	}
	
	return nil
}

// ValidateImageReferences validates that image references contain the necessary Base64 storage information
func (v *Validator) ValidateImageReferences(images *schema.ImageData) error {
	if images == nil {
		return errors.NewValidationError("Image data is nil", nil)
	}
	
	// Validate reference image
	refImage := images.GetReference()
	if refImage == nil {
		return errors.NewValidationError("Reference image is required", nil)
	}
	
	// Check Base64 storage references for S3 temporary storage
	if refImage.StorageMethod == schema.StorageMethodS3Temporary {
		if refImage.Base64Generated {
			// If Base64 is generated, ensure storage references are set
			if refImage.Base64S3Bucket == "" {
				return errors.NewValidationError("Reference image missing Base64 S3 bucket", nil)
			}
			
			if refImage.GetBase64S3Key() == "" {
				return errors.NewValidationError("Reference image missing Base64 S3 key", nil)
			}
			
			v.log.Info("Validated Base64 storage references", map[string]interface{}{
				"bucket": refImage.Base64S3Bucket,
				"key": refImage.GetBase64S3Key(),
				"verificationId": refImage.VerificationId,
			})
		} else {
			// If Base64 is not generated, log a warning but don't fail
			v.log.Warn("Reference image Base64 not generated yet", map[string]interface{}{
				"url": refImage.URL,
				"storageMethod": refImage.StorageMethod,
				"verificationId": refImage.VerificationId,
			})
		}
	}
	
	return nil
}

// ValidateReferenceStructure ensures that the references contain necessary categories
func (v *Validator) ValidateReferenceStructure(references map[string]*state.Reference) error {
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
	
	// Log found categories
	categoryList := make([]string, 0, len(categories))
	for category := range categories {
		categoryList = append(categoryList, category)
	}
	
	v.log.Info("Reference categories found", map[string]interface{}{
		"categories": categoryList,
		"count": len(categories),
	})
	
	// Check for required categories
	requiredCategories := []string{
		state.CategoryInitialization,
		state.CategoryImages,
		state.CategoryPrompts,
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
				"found": categoryList,
			})
	}
	
	return nil
}

// ValidateReferenceAccumulation checks if the output contains all references from the input
func (v *Validator) ValidateReferenceAccumulation(input, output map[string]*state.Reference) error {
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
	
	// Log successful accumulation
	v.log.Info("Reference accumulation validated", map[string]interface{}{
		"inputRefCount": len(input),
		"outputRefCount": len(output),
		"newRefCount": len(output) - len(input),
	})
	
	return nil
}