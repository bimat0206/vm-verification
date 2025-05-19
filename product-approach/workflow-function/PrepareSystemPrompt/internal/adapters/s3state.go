package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/s3state"
	
	"workflow-function/PrepareSystemPrompt/internal/config"
)

// S3StateAdapter wraps shared/s3state functionality
type S3StateAdapter struct {
	stateManager *s3state.Manager
	config       *config.Config
	logger       logger.Logger
}

// NewS3StateAdapter creates a new S3 state adapter
func NewS3StateAdapter(cfg *config.Config, log logger.Logger) (*S3StateAdapter, error) {
	manager, err := s3state.New(cfg.StateBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 state manager: %w", err)
	}
	
	return &S3StateAdapter{
		stateManager: &manager,
		config:       cfg,
		logger:       log,
	}, nil
}

// GenerateDateBasedKey creates a date-based key path
func (s *S3StateAdapter) GenerateDateBasedKey(verificationID, category, filename string) string {
	datePath := s.config.CurrentDatePartition()
	return fmt.Sprintf("%s/%s/%s/%s", datePath, verificationID, category, filename)
}

// GenerateDateBasedKeyWithPartition creates a date-based key with custom partition
func (s *S3StateAdapter) GenerateDateBasedKeyWithPartition(datePartition, verificationID, category, filename string) string {
	return fmt.Sprintf("%s/%s/%s/%s", datePartition, verificationID, category, filename)
}

// LoadVerificationContext loads verification context from S3 state
func (s *S3StateAdapter) LoadVerificationContext(datePartition, verificationID string) (*schema.VerificationContext, error) {
	key := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "processing", s3state.InitializationFile)
	
	ref := &s3state.Reference{
		Bucket: s.config.StateBucket,
		Key:    key,
	}
	
	var state schema.WorkflowState
	err := (*s.stateManager).RetrieveJSON(ref, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to load verification context: %w", err)
	}
	
	return state.VerificationContext, nil
}

// LoadLayoutMetadata loads layout metadata from S3 state
func (s *S3StateAdapter) LoadLayoutMetadata(datePartition, verificationID string) (map[string]interface{}, error) {
	// First try standard location
	key := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "processing", s3state.LayoutMetadataFile)
	
	ref := &s3state.Reference{
		Bucket: s.config.StateBucket,
		Key:    key,
	}
	
	var metadata map[string]interface{}
	err := (*s.stateManager).RetrieveJSON(ref, &metadata)
	if err == nil {
		s.logger.Info("Layout metadata loaded from state bucket", map[string]interface{}{
			"verificationId": verificationID,
			"bucket": s.config.StateBucket,
			"key": key,
		})
		return metadata, nil
	}

	// If not found, try alternative locations
	keyAlt := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "metadata", "layout-metadata.json")
	refAlt := &s3state.Reference{
		Bucket: s.config.StateBucket,
		Key:    keyAlt,
	}
	
	err = (*s.stateManager).RetrieveJSON(refAlt, &metadata)
	if err == nil {
		s.logger.Info("Layout metadata loaded from alternate location in state bucket", map[string]interface{}{
			"verificationId": verificationID,
			"bucket": s.config.StateBucket,
			"key": keyAlt,
		})
		return metadata, nil
	}
	
	// If still not found, try from reference bucket if layoutId is available
	vCtx, vErr := s.LoadVerificationContext(datePartition, verificationID)
	if vErr == nil && vCtx != nil && vCtx.LayoutId > 0 {
		s.logger.Info("Attempting to load layout metadata from reference bucket", map[string]interface{}{
			"verificationId": verificationID,
			"layoutId": vCtx.LayoutId,
		})
		
		// Try to load from reference bucket
		var refBucketKey string
		if vCtx.LayoutPrefix != "" {
			refBucketKey = fmt.Sprintf("layouts/%d_%s/layout-metadata.json", vCtx.LayoutId, vCtx.LayoutPrefix)
		} else {
			refBucketKey = fmt.Sprintf("layouts/%d/layout-metadata.json", vCtx.LayoutId)
		}
		
		refRefBucket := &s3state.Reference{
			Bucket: s.config.ReferenceBucket,
			Key:    refBucketKey,
		}
		
		err = (*s.stateManager).RetrieveJSON(refRefBucket, &metadata)
		if err == nil {
			s.logger.Info("Layout metadata loaded from reference bucket", map[string]interface{}{
				"verificationId": verificationID,
				"bucket": s.config.ReferenceBucket,
				"key": refBucketKey,
			})
			return metadata, nil
		}
	}
	
	return nil, fmt.Errorf("failed to load layout metadata: %w", err)
}

// LoadHistoricalContext loads historical context from S3 state
func (s *S3StateAdapter) LoadHistoricalContext(datePartition, verificationID string) (map[string]interface{}, error) {
	key := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "processing", s3state.HistoricalContextFile)
	
	ref := &s3state.Reference{
		Bucket: s.config.StateBucket,
		Key:    key,
	}
	
	var context map[string]interface{}
	err := (*s.stateManager).RetrieveJSON(ref, &context)
	if err != nil {
		return nil, fmt.Errorf("failed to load historical context: %w", err)
	}
	
	return context, nil
}

// LoadWorkflowState loads complete workflow state from S3
func (s *S3StateAdapter) LoadWorkflowState(datePartition, verificationID string) (*schema.WorkflowState, error) {
	key := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "processing", s3state.InitializationFile)
	
	ref := &s3state.Reference{
		Bucket: s.config.StateBucket,
		Key:    key,
	}
	
	var state schema.WorkflowState
	err := (*s.stateManager).RetrieveJSON(ref, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow state: %w", err)
	}
	
	return &state, nil
}

// UpdateVerificationContextFromS3 updates the verification context in a workflow state from an S3 reference
func (s *S3StateAdapter) UpdateVerificationContextFromS3(state *schema.WorkflowState, ref *s3state.Reference) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}
	
	if ref == nil {
		return fmt.Errorf("reference is nil")
	}
	
	// First try to load as a full workflow state
	var refState schema.WorkflowState
	err := (*s.stateManager).RetrieveJSON(ref, &refState)
	if err == nil && refState.VerificationContext != nil {
		s.logger.Info("Successfully loaded full state with verification context", map[string]interface{}{
			"bucket": ref.Bucket,
			"key": ref.Key,
			"verification_id": refState.VerificationContext.VerificationId,
		})
		state.VerificationContext = refState.VerificationContext
		return nil
	}
	
	// If that fails, try to load the raw data structure to check if it matches the initialization.json format
	var rawData map[string]interface{}
	err = (*s.stateManager).RetrieveJSON(ref, &rawData)
	if err == nil {
		// Check if this is a direct initialization.json format (not wrapped in WorkflowState)
		if verificationId, hasVerifId := rawData["verificationId"]; hasVerifId && verificationId != "" {
			if verificationType, hasVerifType := rawData["verificationType"]; hasVerifType && verificationType != "" {
				s.logger.Info("Found direct initialization.json format", map[string]interface{}{
					"bucket": ref.Bucket,
					"key": ref.Key,
					"verification_id": verificationId,
					"verification_type": verificationType,
				})
				
				// Convert the raw data directly to a verification context
				rawBytes, err := json.Marshal(rawData)
				if err == nil {
					var vContext schema.VerificationContext
					if err := json.Unmarshal(rawBytes, &vContext); err == nil {
						s.logger.Info("Successfully converted initialization.json to verification context", map[string]interface{}{
							"verification_id": vContext.VerificationId,
							"verification_type": vContext.VerificationType,
						})
						state.VerificationContext = &vContext
						
						// Also extract any additional information that might be useful
						if layoutId, ok := rawData["layoutId"]; ok {
							if layoutIdFloat, ok := layoutId.(float64); ok {
								state.VerificationContext.LayoutId = int(layoutIdFloat)
							}
						}
						
						return nil
					}
				}
				
				// If the unmarshal failed, manually construct the verification context
				vContext := &schema.VerificationContext{
					VerificationId: rawData["verificationId"].(string),
				}
				
				// Add verification type if available
				if verifType, ok := rawData["verificationType"].(string); ok {
					vContext.VerificationType = verifType
				}
				
				// Add vendor machine ID if available
				if vmId, ok := rawData["vendingMachineId"].(string); ok {
					vContext.VendingMachineId = vmId
				}
				
				// Add verification time if available
				if verifAt, ok := rawData["verificationAt"].(string); ok {
					vContext.VerificationAt = verifAt
				}
				
				// Add status if available
				if status, ok := rawData["status"].(string); ok {
					vContext.Status = status
				}
				
				// Add layout ID if available
				if layoutId, ok := rawData["layoutId"].(float64); ok {
					vContext.LayoutId = int(layoutId)
				}
				
				// Add layout prefix if available
				if layoutPrefix, ok := rawData["layoutPrefix"].(string); ok {
					vContext.LayoutPrefix = layoutPrefix
				}
				
				state.VerificationContext = vContext
				s.logger.Info("Manually constructed verification context from initialization.json", map[string]interface{}{
					"verification_id": vContext.VerificationId,
					"verification_type": vContext.VerificationType,
				})
				return nil
			}
		}
		
		// Check for a verification context key
		if vCtxObj, ok := rawData["verificationContext"]; ok {
			// Convert to JSON and back to properly parse the object
			if vCtxBytes, err := json.Marshal(vCtxObj); err == nil {
				var vCtx schema.VerificationContext
				if err := json.Unmarshal(vCtxBytes, &vCtx); err == nil && vCtx.VerificationId != "" {
					s.logger.Info("Extracted verification context from JSON object", map[string]interface{}{
						"bucket": ref.Bucket,
						"key": ref.Key,
						"verification_id": vCtx.VerificationId,
					})
					state.VerificationContext = &vCtx
					return nil
				}
			}
		}
	}
	
	// If all else fails, try to load as a direct verification context
	var vContext schema.VerificationContext
	err = (*s.stateManager).RetrieveJSON(ref, &vContext)
	if err == nil && vContext.VerificationId != "" {
		s.logger.Info("Successfully loaded direct verification context", map[string]interface{}{
			"bucket": ref.Bucket,
			"key": ref.Key,
			"verification_id": vContext.VerificationId,
		})
		state.VerificationContext = &vContext
		return nil
	}
	
	return fmt.Errorf("failed to extract verification context from reference")
}

// StoreSystemPrompt stores a system prompt in S3 state
func (s *S3StateAdapter) StoreSystemPrompt(datePartition, verificationID string, prompt *schema.SystemPrompt) (*s3state.Reference, error) {
	// Create the full key path including the prompts category
	// Format: {datePartition}/{verificationId}/prompts/system-prompt.json
	key := fmt.Sprintf("%s/%s/prompts/system-prompt.json", datePartition, verificationID)
	
	// Use empty category to avoid duplication since we're building the full path ourselves
	ref, err := (*s.stateManager).StoreJSON("", key, prompt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to store system prompt: %w", err)
	}
	
	s.logger.Info("Stored system prompt", map[string]interface{}{
		"bucket": ref.Bucket,
		"key": ref.Key,
		"size": ref.Size,
	})
	
	return ref, nil
}

// CreateS3Envelope creates a new S3 state envelope
func (s *S3StateAdapter) CreateS3Envelope(verificationID string) *s3state.Envelope {
	return s3state.NewEnvelope(verificationID)
}

// UpdateWorkflowState updates a workflow state in S3
func (s *S3StateAdapter) UpdateWorkflowState(datePartition, verificationID string, state *schema.WorkflowState) (*s3state.Reference, error) {
	key := s.GenerateDateBasedKeyWithPartition(datePartition, verificationID, "processing", s3state.InitializationFile)
	
	// Use empty category to avoid duplication - the category is already included in the key
	ref, err := (*s.stateManager).StoreJSON("", key, state)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow state: %w", err)
	}
	
	s.logger.Info("Updated workflow state", map[string]interface{}{
		"bucket": ref.Bucket,
		"key": ref.Key,
		"size": ref.Size,
	})
	
	return ref, nil
}

// CombineReferences combines references from different operations
func (s *S3StateAdapter) CombineReferences(envelope *s3state.Envelope, refs map[string]*s3state.Reference) {
	for name, ref := range refs {
		envelope.AddReference(name, ref)
	}
}

// AccumulateAllReferences preserves all input references and adds new references
// This is critical for ensuring that references from previous steps in the workflow
// are preserved and forwarded to subsequent steps
func (s *S3StateAdapter) AccumulateAllReferences(inputReferences map[string]*s3state.Reference, newReferences map[string]*s3state.Reference) map[string]*s3state.Reference {
	// Create a fresh output map to hold the complete reference collection
	outputReferences := make(map[string]*s3state.Reference)
	
	// First phase: Preserve all incoming references from previous functions
	for referenceKey, referenceValue := range inputReferences {
		outputReferences[referenceKey] = referenceValue
		s.logger.Info("Preserving input reference", map[string]interface{}{
			"key": referenceKey,
			"bucket": referenceValue.Bucket,
			"s3_key": referenceValue.Key,
		})
	}
	
	// Second phase: Add all newly created references (overwrites if same key)
	for referenceKey, referenceValue := range newReferences {
		outputReferences[referenceKey] = referenceValue
		s.logger.Info("Adding new reference", map[string]interface{}{
			"key": referenceKey,
			"bucket": referenceValue.Bucket,
			"s3_key": referenceValue.Key,
		})
	}
	
	return outputReferences
}

// GetDatePartitionFromReference extracts the date partition from a reference
func (s *S3StateAdapter) GetDatePartitionFromReference(reference *s3state.Reference) (string, error) {
	if reference == nil || reference.Key == "" {
		return "", fmt.Errorf("invalid reference")
	}
	
	// Parse the reference key
	// Handle two possible formats:
	// 1. YYYY/MM/DD/verificationId/...
	// 2. category/YYYY/MM/DD/verificationId/...
	parts := strings.Split(reference.Key, "/")
	
	// If key doesn't have enough parts, return error
	if len(parts) < 4 {
		return "", fmt.Errorf("reference key does not contain date partition: %s", reference.Key)
	}
	
	// Check for numeric first segment (year) to determine format
	var yearIndex int
	if len(parts[0]) == 4 && strings.HasPrefix(parts[0], "20") {
		// Format 1: YYYY/MM/DD/verificationId/...
		yearIndex = 0
	} else if len(parts) >= 5 && len(parts[1]) == 4 && strings.HasPrefix(parts[1], "20") {
		// Format 2: category/YYYY/MM/DD/verificationId/...
		yearIndex = 1
	} else {
		return "", fmt.Errorf("could not find date partition in key: %s", reference.Key)
	}
	
	// Make sure we have enough parts to extract the date
	if yearIndex+2 >= len(parts) {
		return "", fmt.Errorf("reference key does not contain complete date partition: %s", reference.Key)
	}
	
	// Return the date partition
	return fmt.Sprintf("%s/%s/%s", parts[yearIndex], parts[yearIndex+1], parts[yearIndex+2]), nil
}

// LoadStateFromEnvelope loads state from an S3 envelope
func (s *S3StateAdapter) LoadStateFromEnvelope(ctx context.Context, envelope *s3state.Envelope) (*schema.WorkflowState, string, error) {
	if envelope == nil {
		return nil, "", fmt.Errorf("envelope is nil")
	}
	
	// Find the initialization reference
	var initRef *s3state.Reference
	for key, ref := range envelope.References {
		if key == "processing_initialization" || strings.HasSuffix(key, "_initialization") {
			initRef = ref
			break
		}
	}
	
	if initRef == nil {
		return nil, "", fmt.Errorf("initialization reference not found in envelope")
	}
	
	// Extract date partition
	datePartition, err := s.GetDatePartitionFromReference(initRef)
	if err != nil {
		return nil, "", err
	}
	
	// Load state
	var state schema.WorkflowState
	err = (*s.stateManager).RetrieveJSON(initRef, &state)
	if err != nil {
		return nil, datePartition, fmt.Errorf("failed to load workflow state: %w", err)
	}
	
	// Log detailed information about the loaded state
	s.logger.Info("Loaded workflow state", map[string]interface{}{
		"bucket": initRef.Bucket,
		"key": initRef.Key,
		"schema_version": state.SchemaVersion,
		"has_verification_context": state.VerificationContext != nil,
	})
	
	// Try to load the raw JSON to understand the structure
	var rawState map[string]interface{}
	rawErr := (*s.stateManager).RetrieveJSON(initRef, &rawState)
	if rawErr == nil {
		// Check if there's a directly embedded verification context at the root level
		directVerifCtx := false
		
		for key := range rawState {
			if key == "verificationId" {
				directVerifCtx = true
				break
			}
		}
		
		if directVerifCtx {
			// This isn't a proper WorkflowState, but rather a direct verification context
			// Convert the raw state to a verification context directly
			verificationBytes, err := json.Marshal(rawState)
			if err == nil {
				var verificationContext schema.VerificationContext
				if err := json.Unmarshal(verificationBytes, &verificationContext); err == nil && verificationContext.VerificationId != "" {
					s.logger.Info("Found direct verification context in state file", map[string]interface{}{
						"verification_id": verificationContext.VerificationId,
						"verification_type": verificationContext.VerificationType,
					})
					state.VerificationContext = &verificationContext
				}
			}
		}
	}
	
	// Verify that verification context exists
	if state.VerificationContext == nil {
		s.logger.Error("State loaded successfully but verification context is nil", map[string]interface{}{
			"bucket": initRef.Bucket,
			"key":    initRef.Key,
			"envelope_verification_id": envelope.VerificationID,
			"state_schema_version": state.SchemaVersion,
		})
		
		// Log the full state for debugging
		stateJSON, _ := json.Marshal(state)
		s.logger.Info("State content", map[string]interface{}{
			"state": string(stateJSON),
		})
		
		// Try to initialize a minimal verification context from the envelope
		if envelope != nil && envelope.VerificationID != "" {
			s.logger.Info("Attempting to create verification context from envelope", map[string]interface{}{
				"verificationId": envelope.VerificationID,
				"references": fmt.Sprintf("%v", envelope.References),
			})
			
			// Create a more complete verification context with default values
			verificationType := schema.VerificationTypeLayoutVsChecking // Default type
			
			// Try to determine verification type from reference paths
			for key := range envelope.References {
				if strings.Contains(strings.ToLower(key), "layout") {
					verificationType = schema.VerificationTypeLayoutVsChecking
					s.logger.Info("Determined verification type from references", map[string]interface{}{
						"type": "LAYOUT_VS_CHECKING",
						"key": key,
					})
					break
				} else if strings.Contains(strings.ToLower(key), "previous") {
					verificationType = schema.VerificationTypePreviousVsCurrent
					s.logger.Info("Determined verification type from references", map[string]interface{}{
						"type": "PREVIOUS_VS_CURRENT",
						"key": key,
					})
					break
				}
			}
			
			// Look for initialization.json file in the references and try to read it directly
			for key, ref := range envelope.References {
				if strings.Contains(key, "initialization") && ref != nil {
					// Try to load the file directly to extract more details
					var initState schema.WorkflowState
					initErr := (*s.stateManager).RetrieveJSON(ref, &initState)
					if initErr == nil && initState.VerificationContext != nil {
						s.logger.Info("Found initialization file with verification context", map[string]interface{}{
							"key": key,
							"verification_id": initState.VerificationContext.VerificationId,
						})
						// Use this verification context instead
						state.VerificationContext = initState.VerificationContext
						return &state, datePartition, nil
					}
				}
			}
			
			// Create new context with available information
			state.VerificationContext = &schema.VerificationContext{
				VerificationId: envelope.VerificationID,
				Status:         "INITIALIZED",
				VerificationType: verificationType,
			}
			
			s.logger.Info("Created new verification context", map[string]interface{}{
				"verification_id": state.VerificationContext.VerificationId,
				"verification_type": state.VerificationContext.VerificationType,
			})
		}
		
		// If we still don't have a verification context, return an error
		if state.VerificationContext == nil {
			return nil, datePartition, fmt.Errorf("verification context is nil in state and could not be recovered: empty state or missing verification ID in envelope")
		}
	}
	
	return &state, datePartition, nil
}