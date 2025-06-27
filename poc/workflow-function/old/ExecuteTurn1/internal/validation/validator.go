package validation

import (
	"fmt"
	"strings"
	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"

	"workflow-function/ExecuteTurn1/internal"
)

// Validator provides validation logic for ExecuteTurn1
type Validator struct {
	logger logger.Logger
}

// NewValidator creates a new validator
func NewValidator(logger logger.Logger) *Validator {
	return &Validator{
		logger: logger.WithFields(map[string]interface{}{"component": "Validator"}),
	}
}

// ValidateStateReferences validates incoming state references
func (v *Validator) ValidateStateReferences(refs *internal.StateReferences) error {
	if refs == nil {
		return wferrors.NewValidationError("StateReferences is nil", nil)
	}

	if refs.VerificationId == "" {
		return wferrors.NewValidationError("VerificationId is required", nil)
	}

	var missingRefs []string
	var optionalMissing []string

	// Required references - critical for functionality
	if refs.Initialization == nil {
		missingRefs = append(missingRefs, "Initialization (processing_initialization)")
	}

	if refs.SystemPrompt == nil {
		missingRefs = append(missingRefs, "SystemPrompt (prompts_system)")
	}

	// Optional references - log warnings but don't fail validation
	if refs.ImageMetadata == nil {
		optionalMissing = append(optionalMissing, "ImageMetadata (images_metadata)")
	}

	// Validate use case specific refs based on other references
	// Since we don't know the verification type yet, we'll check both
	// but only log warnings if they're missing
	if refs.LayoutMetadata == nil && refs.HistoricalContext == nil {
		optionalMissing = append(optionalMissing, "LayoutMetadata or HistoricalContext")
	}

	// Log warnings for optional missing references
	if len(optionalMissing) > 0 {
		v.logger.Warn("Optional state references are missing, but continuing", map[string]interface{}{
			"missingOptionalRefs": optionalMissing,
			"verificationId": refs.VerificationId,
		})
	}

	// Return error only if critical references are missing
	if len(missingRefs) > 0 {
		v.logger.Error("Missing required state references", map[string]interface{}{
			"missingRefs": missingRefs,
			"verificationId": refs.VerificationId,
		})
		return wferrors.NewValidationError("Missing required state references", map[string]interface{}{
			"missingRefs": missingRefs,
		})
	}

	// Validate date partition format
	if refs.DatePartition == "" {
		v.logger.Warn("Date partition is empty, using current date", nil)
	} else {
		// Check if date partition format is correct (YYYY/MM/DD/verificationId)
		parts := len(refs.DatePartition)
		if parts < 10 { // minimum length of "YYYY/MM/DD"
			v.logger.Warn("Date partition format may be incorrect", map[string]interface{}{
				"datePartition": refs.DatePartition,
			})
		}
	}

	v.logger.Info("State references validation passed", map[string]interface{}{
		"verificationId": refs.VerificationId,
		"hasInitialization": refs.Initialization != nil,
		"hasSystemPrompt": refs.SystemPrompt != nil, 
		"hasImageMetadata": refs.ImageMetadata != nil,
	})
	return nil
}

// ValidateWorkflowState validates the complete workflow state
func (v *Validator) ValidateWorkflowState(state *schema.WorkflowState) error {
	if state == nil {
		return wferrors.NewValidationError("WorkflowState is nil", nil)
	}

	// Validate core workflow state using schema validator
	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		v.logger.Error("WorkflowState validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"verificationId":   state.VerificationContext.VerificationId,
		})
		return wferrors.NewValidationError("Invalid WorkflowState", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Additional validation beyond the schema validator
	if err := v.validateRequiredFields(state); err != nil {
		return err
	}

	// Validate current prompt
	if err := v.validateCurrentPrompt(state); err != nil {
		return err
	}

	// Validate use case specific data
	if err := v.validateUseCaseData(state); err != nil {
		return err
	}

	// Validate Bedrock config
	if err := v.validateBedrockConfig(state.BedrockConfig); err != nil {
		return err
	}

	v.logger.Info("WorkflowState validation passed", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"status":         state.VerificationContext.Status,
		"verificationType": state.VerificationContext.VerificationType,
	})

	return nil
}

// validateRequiredFields checks for nil pointers in critical fields with more flexibility
func (v *Validator) validateRequiredFields(state *schema.WorkflowState) error {
    var missingFields []string
    
    if state.CurrentPrompt == nil {
        missingFields = append(missingFields, "CurrentPrompt")
    }

    if state.BedrockConfig == nil {
        missingFields = append(missingFields, "BedrockConfig")
    }

    if state.VerificationContext == nil {
        missingFields = append(missingFields, "VerificationContext")
    } else {
        // Provide more detailed validation for VerificationContext
        var contextErrors []string
        
        if state.VerificationContext.VerificationId == "" {
            contextErrors = append(contextErrors, "VerificationId")
        }
        
        // Only warn about other missing fields so the process can continue
        if state.VerificationContext.VerificationAt == "" || 
           state.VerificationContext.Status == "" || 
           state.VerificationContext.VerificationType == "" {
            v.logger.Warn("VerificationContext missing some recommended fields", map[string]interface{}{
                "hasVerificationAt": state.VerificationContext.VerificationAt != "",
                "hasStatus": state.VerificationContext.Status != "",
                "hasVerificationType": state.VerificationContext.VerificationType != "",
            })
        }
        
        if len(contextErrors) > 0 {
            missingFields = append(missingFields, fmt.Sprintf("VerificationContext fields: %s", strings.Join(contextErrors, ", ")))
        }
    }
    
    if len(missingFields) > 0 {
        return wferrors.NewValidationError("Missing required fields in WorkflowState", map[string]interface{}{
            "missingFields": missingFields,
        })
    }

    return nil
}

// validateUseCaseData validates the use case specific data based on verification type
func (v *Validator) validateUseCaseData(state *schema.WorkflowState) error {
	if state.VerificationContext == nil {
		return nil // Already validated in validateRequiredFields
	}

	// Validate based on verification type
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// For UC1: Check layout metadata
		if state.LayoutMetadata == nil {
			v.logger.Warn("LayoutMetadata is missing for LAYOUT_VS_CHECKING verification", nil)
		}
		
		// Validate required UC1 fields
		if state.VerificationContext.LayoutId == 0 {
			v.logger.Warn("LayoutId is missing for LAYOUT_VS_CHECKING verification", nil)
		}
		
		if state.VerificationContext.LayoutPrefix == "" {
			v.logger.Warn("LayoutPrefix is missing for LAYOUT_VS_CHECKING verification", nil)
		}
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// For UC2: Check historical context
		if state.HistoricalContext == nil {
			v.logger.Warn("HistoricalContext is missing for PREVIOUS_VS_CURRENT verification", nil)
		}
		
		// Validate required UC2 fields
		if state.VerificationContext.PreviousVerificationId == "" {
			v.logger.Warn("PreviousVerificationId is missing for PREVIOUS_VS_CURRENT verification", nil)
		}
	}

	return nil
}

// validateCurrentPrompt validates the current prompt
func (v *Validator) validateCurrentPrompt(state *schema.WorkflowState) error {
	// Skip if current prompt is nil - already caught in validateRequiredFields
	if state.CurrentPrompt == nil {
		return nil
	}

	// Standard schema validation with more permissive image requirement
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, false); len(errs) > 0 {
		v.logger.Error("CurrentPrompt validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
			"promptId":         state.CurrentPrompt.PromptId,
		})
		return wferrors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Check that we have at least one message or text
	if len(state.CurrentPrompt.Messages) == 0 && state.CurrentPrompt.Text == "" {
		return wferrors.NewValidationError("CurrentPrompt has no messages or text", map[string]interface{}{
			"promptId": state.CurrentPrompt.PromptId,
		})
	}

	// If using messages, check content
	if len(state.CurrentPrompt.Messages) > 0 {
		msg := state.CurrentPrompt.Messages[0]
		if len(msg.Content) == 0 {
			return wferrors.NewValidationError("Message has no content", map[string]interface{}{
				"promptId": state.CurrentPrompt.PromptId,
				"role":     msg.Role,
			})
		}
	}

	return nil
}

// validateBedrockConfig validates the Bedrock configuration
func (v *Validator) validateBedrockConfig(state *schema.BedrockConfig) error {
	// Skip if Bedrock config is nil - already caught in validateRequiredFields
	if state == nil {
		return nil
	}

	// Standard schema validation
	if errs := schema.ValidateBedrockConfig(state); len(errs) > 0 {
		v.logger.Error("BedrockConfig validation failed", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
		return wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{
			"validationErrors": errs.Error(),
		})
	}

	// Just check for important fields without enforcing specific values
	if state.MaxTokens <= 0 {
		v.logger.Warn("BedrockConfig.MaxTokens is invalid (zero or negative)", map[string]interface{}{
			"currentValue": state.MaxTokens,
		})
	}

	if state.AnthropicVersion == "" {
		v.logger.Warn("BedrockConfig.AnthropicVersion is empty", nil)
	}

	return nil
}

// ValidateImageInfo validates image information without format checks
func (v *Validator) ValidateImageInfo(img *schema.ImageInfo, requireBase64 bool) error {
    if img == nil {
        return fmt.Errorf("image info is nil")
    }

    // Basic validations (less strict)
    if img.URL == "" {
        v.logger.Warn("Image URL is empty", nil)
        // Continue anyway - not a critical failure
    }
    
    // Base64 validation when required
    if requireBase64 {
        // Check for S3 storage of Base64 data
        if !img.HasBase64Data() {
            return fmt.Errorf("image missing Base64 data reference")
        }
        
        // Skip format validation - already handled in loader
    }
    
    return nil
}

// ValidateImageData validates the complete ImageData structure with simplified checks
func (v *Validator) ValidateImageData(images *schema.ImageData, requireBase64 bool) error {
    if images == nil {
        return fmt.Errorf("images cannot be nil")
    }
    
    // Check Base64 generation flag
    if requireBase64 && !images.Base64Generated {
        v.logger.Warn("Base64Generated flag is false but Base64 is required - this may be fixed by the loader", nil)
        // Continue anyway since the loader should have fixed this
    }
    
    // Validate reference image exists
    ref := images.GetReference()
    if ref == nil {
        return fmt.Errorf("reference image is missing")
    }
    
    if requireBase64 {
        if !ref.HasBase64Data() {
            return fmt.Errorf("reference image missing Base64 data reference")
        }
    }
    
    // Only validate checking image if it's present (might be optional)
	checking := images.GetChecking()
	if checking != nil && requireBase64 && !checking.HasBase64Data() {
		v.logger.Warn("Checking image missing Base64 data reference", nil)
		// Not a critical error for the reference image verification
	}
    
    v.logger.Info("ImageData validation passed", map[string]interface{}{
        "hasReference": ref != nil,
        "hasChecking": checking != nil,
        "base64Generated": images.Base64Generated,
    })
    
    return nil
}

// ValidateTurnResponse validates the Turn1 response
// ValidateTurnResponse validates the Turn1 response
func (v *Validator) ValidateTurnResponse(turnResponse *schema.TurnResponse) error {
	if turnResponse == nil {
		return wferrors.NewValidationError("TurnResponse is nil", nil)
	}

	// Validate required fields
	if turnResponse.TurnId != 1 {
		return wferrors.NewValidationError("Invalid TurnId for Turn1", map[string]interface{}{
			"turnId": turnResponse.TurnId,
		})
	}

	if turnResponse.Timestamp == "" {
		return wferrors.NewValidationError("Missing timestamp", nil)
	}

	if turnResponse.Response.Content == "" {
		return wferrors.NewValidationError("Empty response content", nil)
	}

	v.logger.Info("TurnResponse validation passed", map[string]interface{}{
		"turnId": turnResponse.TurnId,
	})

	return nil
}