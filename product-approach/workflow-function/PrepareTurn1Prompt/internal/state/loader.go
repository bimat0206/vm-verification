package state

import (
	"fmt"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// Loader handles loading state data from S3
type Loader struct {
	s3Manager s3state.Manager
	log       logger.Logger
}

// NewLoader creates a new state loader with the given S3 manager and logger
func NewLoader(s3Manager s3state.Manager, log logger.Logger) *Loader {
	return &Loader{
		s3Manager: s3Manager,
		log:       log,
	}
}

// LoadWorkflowState loads the complete workflow state from S3 references
func (l *Loader) LoadWorkflowState(input *Input) (*schema.WorkflowState, error) {
	// Create a new workflow state
	state := &schema.WorkflowState{
		SchemaVersion: schema.SchemaVersion,
		Images:        &schema.ImageData{},
	}

	// Load verification context from initialization reference
	if err := l.loadVerificationContext(input, state); err != nil {
		return nil, fmt.Errorf("failed to load verification context: %w", err)
	}

	// Set turn number and include image from input
	state.CurrentPrompt = &schema.CurrentPrompt{
		TurnNumber:   input.TurnNumber,
		IncludeImage: input.IncludeImage,
	}

	// Load reference image
	if err := l.loadReferenceImage(input, state); err != nil {
		return nil, fmt.Errorf("failed to load reference image: %w", err)
	}

	// Load checking image if available (not required for Turn 1)
	if err := l.loadCheckingImage(input, state); err != nil {
		l.log.Warn("Checking image not loaded (may not be required for Turn 1)", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Load layout metadata for LAYOUT_VS_CHECKING verification type
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		if err := l.loadLayoutMetadata(input, state); err != nil {
			return nil, fmt.Errorf("failed to load layout metadata: %w", err)
		}
	}

	// Load historical context for PREVIOUS_VS_CURRENT verification type
	if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		if err := l.loadHistoricalContext(input, state); err != nil {
			l.log.Warn("Historical context not loaded (may not be required for Turn 1)", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Load system prompt if available
	if err := l.loadSystemPrompt(input, state); err != nil {
		l.log.Warn("System prompt not loaded (using default)", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return state, nil
}

// loadVerificationContext loads the verification context from initialization reference
func (l *Loader) loadVerificationContext(input *Input, state *schema.WorkflowState) error {
	// If S3References is present but References is not, use S3References
	if input.References == nil && input.S3References != nil {
		input.References = input.S3References
	}

	// Try different possible keys for initialization
	possibleKeys := []string{
		GetReferenceKey(CategoryProcessing, "initialization"),
		GetReferenceKey(CategoryInitialization, "initialization"),
		"initialization_initialization",
		"processing_initialization",
	}
	
	var initRef *s3state.Reference
	var foundKey string
	
	for _, key := range possibleKeys {
		if ref, exists := input.References[key]; exists && ref != nil {
			initRef = ref
			foundKey = key
			break
		}
	}
	
	if initRef == nil {
		// Log available references to help troubleshooting
		refKeys := make([]string, 0, len(input.References))
		for k := range input.References {
			refKeys = append(refKeys, k)
		}
		
		return errors.NewValidationError("Initialization reference not found", 
			map[string]interface{}{
				"availableRefs": refKeys,
				"triedKeys":     possibleKeys,
			})
	}
	
	l.log.Info("Found initialization reference", map[string]interface{}{
		"key":    foundKey,
		"bucket": initRef.Bucket,
		"path":   initRef.Key,
	})

	// Load initialization data
	var initData schema.VerificationContext
	if err := l.s3Manager.RetrieveJSON(initRef, &initData); err != nil {
		return errors.NewInternalError("initialization-load", err)
	}

	// Set verification context in state
	state.VerificationContext = &initData

	return nil
}

// loadReferenceImage loads the reference image from images reference or creates from URL in verification context
func (l *Loader) loadReferenceImage(input *Input, state *schema.WorkflowState) error {
	// Get reference image reference
	refImageRef, exists := input.References[GetReferenceKey(CategoryImages, "reference")]
	if !exists {
		// Reference not found in references map, try to create from URL in verification context
		if state.VerificationContext != nil && state.VerificationContext.ReferenceImageUrl != "" {
			l.log.Info("Creating reference image info from URL in verification context", map[string]interface{}{
				"url": state.VerificationContext.ReferenceImageUrl,
			})
			
			// Create image info from URL
			referenceImageInfo := &schema.ImageInfo{
				URL:           state.VerificationContext.ReferenceImageUrl,
				StorageMethod: "s3-temporary",
				ContentType:   "image/png", // Assume PNG, will be updated during processing
				LastModified:  state.VerificationContext.VerificationAt,
			}
			
			// Set reference image in state
			state.Images.Reference = referenceImageInfo
			state.Images.ReferenceImage = referenceImageInfo // Also set in legacy format
			
			return nil
		}
		
		// No reference URL available, return error
		return errors.NewValidationError("Reference image reference not found and no URL in verification context", nil)
	}

	// Load reference image data from reference
	var refImageData schema.ImageInfo
	if err := l.s3Manager.RetrieveJSON(refImageRef, &refImageData); err != nil {
		return errors.NewInternalError("reference-image-load", err)
	}

	// Set reference image in state
	state.Images.Reference = &refImageData
	state.Images.ReferenceImage = &refImageData // Also set in legacy format

	return nil
}

// loadCheckingImage loads the checking image from images reference (if available)
func (l *Loader) loadCheckingImage(input *Input, state *schema.WorkflowState) error {
	// Get checking image reference
	checkImageRef, exists := input.References[GetReferenceKey(CategoryImages, "checking")]
	if !exists {
		// Reference not found in references map, try to create from URL in verification context
		if state.VerificationContext != nil && state.VerificationContext.CheckingImageUrl != "" {
			l.log.Info("Creating checking image info from URL in verification context", map[string]interface{}{
				"url": state.VerificationContext.CheckingImageUrl,
			})
			
			// Create image info from URL
			checkingImageInfo := &schema.ImageInfo{
				URL:           state.VerificationContext.CheckingImageUrl,
				StorageMethod: "s3-temporary",
				ContentType:   "image/png", // Assume PNG, will be updated during processing
				LastModified:  state.VerificationContext.VerificationAt,
			}
			
			// Set checking image in state
			state.Images.Checking = checkingImageInfo
			state.Images.CheckingImage = checkingImageInfo // Also set in legacy format
			
			return nil
		}
		
		// No checking URL available, return error (this is not critical for Turn 1)
		return errors.NewValidationError("Checking image reference not found and no URL in verification context", nil)
	}

	// Load checking image data from reference
	var checkImageData schema.ImageInfo
	if err := l.s3Manager.RetrieveJSON(checkImageRef, &checkImageData); err != nil {
		return errors.NewInternalError("checking-image-load", err)
	}

	// Set checking image in state
	state.Images.Checking = &checkImageData
	state.Images.CheckingImage = &checkImageData // Also set in legacy format

	return nil
}

// loadLayoutMetadata loads layout metadata from processing reference
func (l *Loader) loadLayoutMetadata(input *Input, state *schema.WorkflowState) error {
	// Get layout metadata reference
	layoutRef, exists := input.References[GetReferenceKey(CategoryProcessing, "layout_metadata")]
	if !exists {
		return errors.NewValidationError("Layout metadata reference not found", nil)
	}

	// Load layout metadata
	var layoutData map[string]interface{}
	if err := l.s3Manager.RetrieveJSON(layoutRef, &layoutData); err != nil {
		return errors.NewInternalError("layout-metadata-load", err)
	}

	// Set layout metadata in state
	state.LayoutMetadata = layoutData

	return nil
}

// loadHistoricalContext loads historical context from processing reference
func (l *Loader) loadHistoricalContext(input *Input, state *schema.WorkflowState) error {
	// Get historical context reference
	historicalRef, exists := input.References[GetReferenceKey(CategoryProcessing, "historical_context")]
	if !exists {
		return errors.NewValidationError("Historical context reference not found", nil)
	}

	// Load historical context
	var historicalData map[string]interface{}
	if err := l.s3Manager.RetrieveJSON(historicalRef, &historicalData); err != nil {
		return errors.NewInternalError("historical-context-load", err)
	}

	// Set historical context in state
	state.HistoricalContext = historicalData

	return nil
}

// loadSystemPrompt loads system prompt from prompts reference
func (l *Loader) loadSystemPrompt(input *Input, state *schema.WorkflowState) error {
	// If S3References is present but References is not, use S3References
	if input.References == nil && input.S3References != nil {
		input.References = input.S3References
	}

	// Try different possible keys for system prompt
	possibleKeys := []string{
		GetReferenceKey(CategoryPrompts, "system_prompt"),
		GetReferenceKey(CategoryPrompts, "system-prompt"),
		GetReferenceKey(CategoryPrompts, "system"),
		"prompts_system",
		"prompts_system_prompt",
		"prompts_system-prompt",
	}
	
	var systemPromptRef *s3state.Reference
	var foundKey string
	
	for _, key := range possibleKeys {
		if ref, exists := input.References[key]; exists && ref != nil {
			systemPromptRef = ref
			foundKey = key
			break
		}
	}
	
	if systemPromptRef == nil {
		// Log available references to help troubleshooting
		refKeys := make([]string, 0, len(input.References))
		for k := range input.References {
			refKeys = append(refKeys, k)
		}
		
		return errors.NewValidationError("System prompt reference not found", 
			map[string]interface{}{
				"availableRefs": refKeys,
				"triedKeys":     possibleKeys,
			})
	}
	
	l.log.Info("Found system prompt reference", map[string]interface{}{
		"key":    foundKey,
		"bucket": systemPromptRef.Bucket,
		"path":   systemPromptRef.Key,
	})

	// Load system prompt
	var systemPromptData schema.SystemPrompt
	if err := l.s3Manager.RetrieveJSON(systemPromptRef, &systemPromptData); err != nil {
		return errors.NewInternalError("system-prompt-load", err)
	}

	// Set system prompt in state
	state.SystemPrompt = &systemPromptData

	return nil
}