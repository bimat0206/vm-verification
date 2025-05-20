package state

import (
	"encoding/json"
	"fmt"
	"time"
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

	// Load both images from the combined metadata file
	if err := l.loadImagesFromMetadata(input, state); err != nil {
		return nil, fmt.Errorf("failed to load images from metadata: %w", err)
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

	// Load initialization data - try all possible formats
	// First try nested format (with verificationContext field)
	var initDataNested struct {
		SchemaVersion      string                   `json:"schemaVersion"`
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
		SystemPrompt       *schema.SystemPrompt     `json:"systemPrompt,omitempty"`
		LayoutMetadata     map[string]interface{}   `json:"layoutMetadata,omitempty"`
	}

	// Get the raw JSON first for additional attempts if needed
	var rawJSON map[string]interface{}
	if err := l.s3Manager.RetrieveJSON(initRef, &rawJSON); err != nil {
		return errors.NewInternalError("initialization-load-raw", err)
	}
	
	// Try to parse as nested structure
	if err := l.s3Manager.RetrieveJSON(initRef, &initDataNested); err != nil {
		l.log.Warn("Failed to load as nested structure, will try alternative formats", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Check if we got a nested structure with verificationContext
	if initDataNested.VerificationContext != nil {
		// Set verification context in state using the nested structure
		state.VerificationContext = initDataNested.VerificationContext
		
		// Also set system prompt and layout metadata if available
		if initDataNested.SystemPrompt != nil {
			state.SystemPrompt = initDataNested.SystemPrompt
			l.log.Info("Loaded system prompt from nested initialization structure", nil)
		}
		
		if initDataNested.LayoutMetadata != nil {
			state.LayoutMetadata = initDataNested.LayoutMetadata
			l.log.Info("Loaded layout metadata from nested initialization structure", nil)
		}
		
		l.log.Info("Loaded verification context from nested structure", map[string]interface{}{
			"verificationId": state.VerificationContext.VerificationId,
			"verificationType": state.VerificationContext.VerificationType,
		})
	} else {
		// Try direct VerificationContext fields in the JSON
		if verContextRaw, exists := rawJSON["verificationContext"].(map[string]interface{}); exists {
			// Convert the raw map to a proper VerificationContext
			verContextJSON, err := json.Marshal(verContextRaw)
			if err != nil {
				l.log.Warn("Failed to marshal verification context from raw JSON", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				var verContext schema.VerificationContext
				if err := json.Unmarshal(verContextJSON, &verContext); err != nil {
					l.log.Warn("Failed to unmarshal verification context from JSON", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					state.VerificationContext = &verContext
					l.log.Info("Loaded verification context from raw JSON map", map[string]interface{}{
						"verificationId": state.VerificationContext.VerificationId,
						"verificationType": state.VerificationContext.VerificationType,
					})
				}
			}
		}
		
		// If still nil, try loading as a direct VerificationContext object
		if state.VerificationContext == nil {
			var initData schema.VerificationContext
			if err := l.s3Manager.RetrieveJSON(initRef, &initData); err != nil {
				l.log.Warn("Failed to load as direct structure, will create a new VerificationContext", map[string]interface{}{
					"error": err.Error(),
				})
				// Create a new empty verification context as a fallback
				state.VerificationContext = &schema.VerificationContext{}
			} else {
				// Set verification context in state from direct structure
				state.VerificationContext = &initData
				l.log.Info("Loaded verification context directly", map[string]interface{}{
					"verificationId": state.VerificationContext.VerificationId,
					"verificationType": state.VerificationContext.VerificationType,
				})
			}
		}
	}
	
	// Validate that we have a valid verification context
	if state.VerificationContext == nil {
		return errors.NewInternalError("initialization-load", 
			fmt.Errorf("failed to load verification context in any format"))
	}
	
	// Ensure all required fields are set in the verification context
	// This is required for validation to pass
	if state.VerificationContext.VerificationId == "" && input.VerificationID != "" {
		state.VerificationContext.VerificationId = input.VerificationID
		l.log.Info("Set verificationId from input", map[string]interface{}{
			"verificationId": input.VerificationID,
		})
	}
	
	// Set verificationType if missing - try multiple sources
	if state.VerificationContext.VerificationType == "" {
		// First try input.VerificationType
		if input.VerificationType != "" {
			state.VerificationContext.VerificationType = input.VerificationType
			l.log.Info("Set verificationType from input.VerificationType", map[string]interface{}{
				"verificationType": input.VerificationType,
			})
		} else if input.Summary != nil {
			// Try from summary if available
			if verType, ok := input.Summary["verificationType"].(string); ok && verType != "" {
				state.VerificationContext.VerificationType = verType
				l.log.Info("Set verificationType from input.Summary", map[string]interface{}{
					"verificationType": verType,
				})
			}
		}
	}
	
	// Set status if missing
	if state.VerificationContext.Status == "" && input.Status != "" {
		state.VerificationContext.Status = input.Status
		l.log.Info("Set status from input", map[string]interface{}{
			"status": input.Status,
		})
	}
	
	// Set verificationAt if missing
	if state.VerificationContext.VerificationAt == "" {
		// Use current time if not available
		state.VerificationContext.VerificationAt = time.Now().UTC().Format(time.RFC3339)
		l.log.Info("Set verificationAt to current time", map[string]interface{}{
			"verificationAt": state.VerificationContext.VerificationAt,
		})
	}

	// Log the verification context to help with debugging
	l.log.Info("Verification context after processing", map[string]interface{}{
		"verificationId":   state.VerificationContext.VerificationId,
		"verificationType": state.VerificationContext.VerificationType,
		"verificationAt":   state.VerificationContext.VerificationAt,
		"status":           state.VerificationContext.Status,
	})

	return nil
}

// loadImagesFromMetadata loads both reference and checking images from combined metadata file
func (l *Loader) loadImagesFromMetadata(input *Input, state *schema.WorkflowState) error {
	// Get images metadata reference
	metadataRef, exists := input.References[GetReferenceKey(CategoryImages, "metadata")]
	if !exists {
		return errors.NewValidationError("Images metadata reference not found", map[string]interface{}{
			"expectedKey": GetReferenceKey(CategoryImages, "metadata"),
			"availableRefs": func() []string {
				keys := make([]string, 0, len(input.References))
				for k := range input.References {
					keys = append(keys, k)
				}
				return keys
			}(),
		})
	}

	// First try to load the metadata as the new format (complex structure)
	var complexMetadata struct {
		VerificationId     string                 `json:"verificationId"`
		VerificationType   string                 `json:"verificationType"`
		ReferenceImage     map[string]interface{} `json:"referenceImage"`
		CheckingImage      map[string]interface{} `json:"checkingImage"`
		ProcessingMetadata map[string]interface{} `json:"processingMetadata"`
		Version            string                 `json:"version"`
	}

	if err := l.s3Manager.RetrieveJSON(metadataRef, &complexMetadata); err != nil {
		// If we can't parse as complex structure, try the old format
		var oldMetadata map[string]*schema.ImageInfo
		if oldErr := l.s3Manager.RetrieveJSON(metadataRef, &oldMetadata); oldErr != nil {
			return errors.NewInternalError("images-metadata-load", err)
		}
		
		// Process old format
		return l.processOldFormatMetadata(oldMetadata, state)
	}
	
	// Process the new complex format
	return l.processNewFormatMetadata(complexMetadata, state)
}

// processOldFormatMetadata handles the old metadata format (map[string]*schema.ImageInfo)
func (l *Loader) processOldFormatMetadata(metadata map[string]*schema.ImageInfo, state *schema.WorkflowState) error {
	// Extract reference image from metadata (required for Turn 1)
	if refImage, ok := metadata["referenceImage"]; ok && refImage != nil {
		// Set the image in the state
		state.Images.Reference = refImage
		state.Images.ReferenceImage = refImage // Also set in legacy format for compatibility
		
		l.log.Info("Loaded reference image from metadata (old format)", map[string]interface{}{
			"url":           refImage.URL,
			"storageMethod": refImage.StorageMethod,
			"format":        refImage.Format,
			"size":          refImage.Size,
		})
	} else {
		return errors.NewValidationError("Reference image not found in metadata", map[string]interface{}{
			"availableImages": func() []string {
				keys := make([]string, 0, len(metadata))
				for k := range metadata {
					keys = append(keys, k)
				}
				return keys
			}(),
		})
	}

	// Extract checking image from metadata (optional for Turn 1)
	if checkImage, ok := metadata["checkingImage"]; ok && checkImage != nil {
		// Set the image in the state
		state.Images.Checking = checkImage
		state.Images.CheckingImage = checkImage // Also set in legacy format for compatibility
		
		l.log.Info("Loaded checking image from metadata (old format)", map[string]interface{}{
			"url":           checkImage.URL,
			"storageMethod": checkImage.StorageMethod,
			"format":        checkImage.Format,
			"size":          checkImage.Size,
		})
	} else {
		// Checking image is not critical for Turn 1, so just log a warning
		l.log.Warn("Checking image not found in metadata (old format)", map[string]interface{}{
			"availableImages": func() []string {
				keys := make([]string, 0, len(metadata))
				for k := range metadata {
					keys = append(keys, k)
				}
				return keys
			}(),
		})
	}

	return nil
}

// processNewFormatMetadata handles the new metadata format (complex structure)
func (l *Loader) processNewFormatMetadata(metadata struct {
	VerificationId     string                 `json:"verificationId"`
	VerificationType   string                 `json:"verificationType"`
	ReferenceImage     map[string]interface{} `json:"referenceImage"`
	CheckingImage      map[string]interface{} `json:"checkingImage"`
	ProcessingMetadata map[string]interface{} `json:"processingMetadata"`
	Version            string                 `json:"version"`
}, state *schema.WorkflowState) error {
	// Process reference image (required for Turn 1)
	if metadata.ReferenceImage != nil {
		// Extract storage metadata
		storageMetadata, ok := metadata.ReferenceImage["storageMetadata"].(map[string]interface{})
		if !ok {
			return errors.NewValidationError("Reference image storage metadata not found or invalid", nil)
		}
		
		// Extract original metadata
		originalMetadata, ok := metadata.ReferenceImage["originalMetadata"].(map[string]interface{})
		if !ok {
			return errors.NewValidationError("Reference image original metadata not found or invalid", nil)
		}
		
		// Create ImageInfo from the complex structure
		refImage := &schema.ImageInfo{
			// From originalMetadata
			URL:         getStringValue(originalMetadata, "sourceUrl"),
			S3Key:       getStringValue(originalMetadata, "sourceKey"),
			S3Bucket:    getStringValue(originalMetadata, "sourceBucket"),
			ContentType: getStringValue(originalMetadata, "contentType"),
			Size:        getInt64Value(originalMetadata, "originalSize"),
			
			// From storageMetadata
			Base64S3Bucket: getStringValue(storageMetadata, "bucket"),
			Base64S3Key:    getStringValue(storageMetadata, "key"),
			
			// Set storage method
			StorageMethod:   "s3-temporary",
			Base64Generated: true,
		}
		
		// Set the image in the state
		state.Images.Reference = refImage
		state.Images.ReferenceImage = refImage // Also set in legacy format for compatibility
		
		l.log.Info("Loaded reference image from metadata (new format)", map[string]interface{}{
			"url":           refImage.URL,
			"storageMethod": refImage.StorageMethod,
			"base64S3Bucket": refImage.Base64S3Bucket,
			"base64S3Key":    refImage.Base64S3Key,
		})
	} else {
		return errors.NewValidationError("Reference image not found in metadata", nil)
	}
	
	// Process checking image (optional for Turn 1)
	if metadata.CheckingImage != nil {
		// Extract storage metadata
		storageMetadata, ok := metadata.CheckingImage["storageMetadata"].(map[string]interface{})
		if !ok {
			l.log.Warn("Checking image storage metadata not found or invalid", nil)
			return nil
		}
		
		// Extract original metadata
		originalMetadata, ok := metadata.CheckingImage["originalMetadata"].(map[string]interface{})
		if !ok {
			l.log.Warn("Checking image original metadata not found or invalid", nil)
			return nil
		}
		
		// Create ImageInfo from the complex structure
		checkImage := &schema.ImageInfo{
			// From originalMetadata
			URL:         getStringValue(originalMetadata, "sourceUrl"),
			S3Key:       getStringValue(originalMetadata, "sourceKey"),
			S3Bucket:    getStringValue(originalMetadata, "sourceBucket"),
			ContentType: getStringValue(originalMetadata, "contentType"),
			Size:        getInt64Value(originalMetadata, "originalSize"),
			
			// From storageMetadata
			Base64S3Bucket: getStringValue(storageMetadata, "bucket"),
			Base64S3Key:    getStringValue(storageMetadata, "key"),
			
			// Set storage method
			StorageMethod:   "s3-temporary",
			Base64Generated: true,
		}
		
		// Set the image in the state
		state.Images.Checking = checkImage
		state.Images.CheckingImage = checkImage // Also set in legacy format for compatibility
		
		l.log.Info("Loaded checking image from metadata (new format)", map[string]interface{}{
			"url":           checkImage.URL,
			"storageMethod": checkImage.StorageMethod,
			"base64S3Bucket": checkImage.Base64S3Bucket,
			"base64S3Key":    checkImage.Base64S3Key,
		})
	} else {
		// Checking image is not critical for Turn 1, so just log a warning
		l.log.Warn("Checking image not found in metadata (new format)", nil)
	}
	
	return nil
}

// Helper functions for extracting values from maps
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt64Value(m map[string]interface{}, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	}
	return 0
}

// loadLayoutMetadata loads layout metadata from processing reference
func (l *Loader) loadLayoutMetadata(input *Input, state *schema.WorkflowState) error {
	// Try different possible keys for layout metadata
	possibleKeys := []string{
		GetReferenceKey(CategoryProcessing, "layout_metadata"),
		GetReferenceKey(CategoryProcessing, "layout-metadata"),
		"processing_layout_metadata",
		"processing_layout-metadata",
	}
	
	var layoutRef *s3state.Reference
	var foundKey string
	
	for _, key := range possibleKeys {
		if ref, exists := input.References[key]; exists && ref != nil {
			layoutRef = ref
			foundKey = key
			break
		}
	}
	
	if layoutRef == nil {
		return errors.NewValidationError("Layout metadata reference not found", map[string]interface{}{
			"triedKeys": possibleKeys,
			"availableRefs": func() []string {
				keys := make([]string, 0, len(input.References))
				for k := range input.References {
					keys = append(keys, k)
				}
				return keys
			}(),
		})
	}

	l.log.Info("Found layout metadata reference", map[string]interface{}{
		"key":    foundKey,
		"bucket": layoutRef.Bucket,
		"path":   layoutRef.Key,
	})

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
	// Try different possible keys for historical context
	possibleKeys := []string{
		GetReferenceKey(CategoryProcessing, "historical_context"),
		GetReferenceKey(CategoryProcessing, "historical-context"),
		"processing_historical_context",
		"processing_historical-context",
	}
	
	var historicalRef *s3state.Reference
	var foundKey string
	
	for _, key := range possibleKeys {
		if ref, exists := input.References[key]; exists && ref != nil {
			historicalRef = ref
			foundKey = key
			break
		}
	}
	
	if historicalRef == nil {
		return errors.NewValidationError("Historical context reference not found", map[string]interface{}{
			"triedKeys": possibleKeys,
			"availableRefs": func() []string {
				keys := make([]string, 0, len(input.References))
				for k := range input.References {
					keys = append(keys, k)
				}
				return keys
			}(),
		})
	}

	l.log.Info("Found historical context reference", map[string]interface{}{
		"key":    foundKey,
		"bucket": historicalRef.Bucket,
		"path":   historicalRef.Key,
	})

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