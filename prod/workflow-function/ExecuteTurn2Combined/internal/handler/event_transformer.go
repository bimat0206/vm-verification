// internal/handler/event_transformer.go - COMPREHENSIVE SCHEMA INTEGRATION
package handler

import (
	"context"
	"fmt"
	"strings"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// EventTransformer orchestrates Step Functions event transformation with strategic schema integration
type EventTransformer struct {
	s3  services.S3StateManager
	log logger.Logger
}

// NewEventTransformer creates a strategically enhanced event transformer
func NewEventTransformer(s3 services.S3StateManager, log logger.Logger) *EventTransformer {
	return &EventTransformer{
		s3:  s3,
		log: log,
	}
}

// StepFunctionEvent represents the structured input from Step Functions orchestration
type StepFunctionEvent struct {
	SchemaVersion  string                 `json:"schemaVersion"`
	S3References   map[string]interface{} `json:"s3References"`
	VerificationID string                 `json:"verificationId"`
	Status         string                 `json:"status"`
}

// TransformStepFunctionEvent provides comprehensive transformation with schema integration
func (e *EventTransformer) TransformStepFunctionEvent(ctx context.Context, event StepFunctionEvent) (*models.Turn2Request, error) {
	transformLogger := e.log.WithFields(map[string]interface{}{
		"operation":         "transform_step_function_event",
		"verification_id":   event.VerificationID,
		"schema_version":    event.SchemaVersion,
		"status":            event.Status,
		"shared_schema_ver": schema.SchemaVersion,
	})

	transformLogger.Info("step_function_transformation_started", map[string]interface{}{
		"s3_references_count": len(event.S3References),
		"available_refs":      getMapKeys(event.S3References),
		"schema_integration":  "comprehensive",
	})

	// STRATEGIC STAGE 1: Load initialization data using schema-integrated loader
	initVal, exists := event.S3References["processing_initialization"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing processing_initialization",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	initRef, ok := convertToS3Reference(initVal)
	if !ok {
		return nil, errors.NewValidationError(
			"invalid processing_initialization",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	// Ensure the key points to the /processing/initialization.json path
	if !strings.Contains(initRef.Key, "/processing/initialization.json") && strings.HasSuffix(initRef.Key, "/initialization.json") {
		base := strings.TrimSuffix(initRef.Key, "/initialization.json")
		newKey := fmt.Sprintf("%s/processing/initialization.json", strings.TrimSuffix(base, "/"))
		transformLogger.Info("adjusting_initialization_path", map[string]interface{}{
			"from": initRef.Key,
			"to":   newKey,
		})
		initRef.Key = newKey
	}

	transformLogger.Info("turn2_processing_initialization_load_start", map[string]interface{}{
		"bucket": initRef.Bucket,
		"key":    initRef.Key,
		"size":   initRef.Size,
	})

	// STRATEGIC RESOLUTION: Use schema-integrated initialization data loader with fallback handling
	transformLogger.Debug("attempting_to_load_initialization_data", map[string]interface{}{
		"bucket": initRef.Bucket,
		"key":    initRef.Key,
		"size":   initRef.Size,
	})

	initData, err := e.s3.LoadInitializationData(ctx, initRef)
	if err != nil {
		transformLogger.Error("initialization_data_load_failed", map[string]interface{}{
			"error":  err.Error(),
			"bucket": initRef.Bucket,
			"key":    initRef.Key,
		})

		// initialization.json is REQUIRED - fail fast if missing
		transformLogger.Error("initialization_file_missing_critical_error", map[string]interface{}{
			"verification_id": event.VerificationID,
			"missing_key":     initRef.Key,
			"missing_bucket":  initRef.Bucket,
			"error_message":   err.Error(),
			"impact":          "ExecuteTurn2Combined cannot proceed without initialization.json",
			"workflow_issue":  "UPSTREAM_FAILURE",
			"investigation_steps": []string{
				"1. Check Initialize Lambda logs for verification ID: " + event.VerificationID,
				"2. Verify S3 bucket contains: " + initRef.Key,
				"3. Check Step Functions execution state transitions",
				"4. Verify ExecuteTurn1Combined completed successfully",
			},
			"expected_workflow": "Initialize → ExecuteTurn1Combined → ExecuteTurn2Combined",
			"file_creator":      "Initialize Lambda Function",
			"file_updater":      "ExecuteTurn1Combined",
			"file_consumer":     "ExecuteTurn2Combined",
		})

		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"CRITICAL: initialization.json file is required and missing", true).
			WithContext("s3_key", initRef.Key).
			WithContext("verification_id", event.VerificationID).
			WithContext("upstream_issue", "ExecuteTurn1Combined should create initialization.json").
			WithContext("schema_integration", "failed")
	}

	transformLogger.Info("initialization_data_loaded_successfully", map[string]interface{}{
		"verification_id":       initData.VerificationContext.VerificationId,
		"verification_type":     initData.VerificationContext.VerificationType,
		"vending_machine_id":    initData.VerificationContext.VendingMachineId,
		"layout_id":             initData.VerificationContext.LayoutId,
		"layout_prefix":         initData.VerificationContext.LayoutPrefix,
		"system_prompt_id":      initData.SystemPrompt.PromptID,
		"system_prompt_version": initData.SystemPrompt.PromptVersion,
		"has_layout_metadata":   initData.LayoutMetadata != nil,
		"schema_validation":     "passed",
	})

	// STRATEGIC STAGE 2: Load image metadata with comprehensive validation
	metadataVal, exists := event.S3References["images_metadata"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing images_metadata reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	metadataRef, ok := convertToS3Reference(metadataVal)
	if !ok {
		return nil, errors.NewValidationError(
			"invalid images_metadata reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	transformLogger.Info("loading_image_metadata", map[string]interface{}{
		"bucket": metadataRef.Bucket,
		"key":    metadataRef.Key,
		"size":   metadataRef.Size,
	})

	// STRATEGIC ARCHITECTURE: Use schema-enhanced image metadata loader
	metadata, err := e.s3.LoadImageMetadata(ctx, metadataRef)
	if err != nil {
		transformLogger.Error("image_metadata_load_failed", map[string]interface{}{
			"error":  err.Error(),
			"bucket": metadataRef.Bucket,
			"key":    metadataRef.Key,
		})
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load image metadata", true).
			WithContext("s3_key", metadataRef.Key).
			WithContext("verification_id", event.VerificationID)
	}

	transformLogger.Info("image_metadata_loaded_successfully", map[string]interface{}{
		"reference_image_bucket":   metadata.ReferenceImage.StorageMetadata.Bucket,
		"reference_image_key":      metadata.ReferenceImage.StorageMetadata.Key,
		"reference_image_size":     metadata.ReferenceImage.StorageMetadata.StoredSize,
		"reference_bedrock_compat": metadata.ReferenceImage.Validation.BedrockCompatible,
		"checking_image_bucket":    metadata.CheckingImage.StorageMetadata.Bucket,
		"checking_image_key":       metadata.CheckingImage.StorageMetadata.Key,
		"checking_image_size":      metadata.CheckingImage.StorageMetadata.StoredSize,
		"checking_bedrock_compat":  metadata.CheckingImage.Validation.BedrockCompatible,
		"parallel_processing":      metadata.ProcessingMetadata.ParallelProcessing,
		"total_images_processed":   metadata.ProcessingMetadata.TotalImagesProcessed,
	})

	// STRATEGIC STAGE 3: Validate system prompt reference
	systemPromptVal, exists := event.S3References["prompts_system"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing prompts_system reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	systemPromptRef, ok := convertToS3Reference(systemPromptVal)
	if !ok {
		return nil, errors.NewValidationError(
			"invalid prompts_system reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}

	transformLogger.Info("system_prompt_reference_validated", map[string]interface{}{
		"bucket": systemPromptRef.Bucket,
		"key":    systemPromptRef.Key,
		"size":   systemPromptRef.Size,
	})

	// STRATEGIC STAGE 4: Enhanced layout metadata integration with schema validation
	if initData.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		if layoutMetaVal, exists := event.S3References["processing_layout-metadata"]; exists {
			if layoutMetadataRef, ok := convertToS3Reference(layoutMetaVal); ok {
				transformLogger.Info("loading_layout_metadata", map[string]interface{}{
					"bucket":         layoutMetadataRef.Bucket,
					"key":            layoutMetadataRef.Key,
					"schema_version": schema.SchemaVersion,
				})

				// STRATEGIC ENHANCEMENT: Use shared schema layout metadata loader
				layoutData, err := e.s3.LoadLayoutMetadata(ctx, layoutMetadataRef)
				if err != nil {
					transformLogger.Warn("layout_metadata_load_failed", map[string]interface{}{
						"error":  err.Error(),
						"key":    layoutMetadataRef.Key,
						"impact": "continuing_with_embedded_layout_data",
					})
				} else {
					if initData.VerificationContext.LayoutId == 0 {
						initData.VerificationContext.LayoutId = layoutData.LayoutId
					}
					if initData.VerificationContext.LayoutPrefix == "" {
						initData.VerificationContext.LayoutPrefix = layoutData.LayoutPrefix
					}
					if initData.VerificationContext.VendingMachineId == "" {
						initData.VerificationContext.VendingMachineId = layoutData.VendingMachineId
					}

					if initData.LayoutMetadata == nil {
						initData.LayoutMetadata = layoutData
					}

					transformLogger.Info("layout_metadata_integrated_successfully", map[string]interface{}{
						"layout_id":          layoutData.LayoutId,
						"layout_prefix":      layoutData.LayoutPrefix,
						"vending_machine_id": layoutData.VendingMachineId,
						"location":           layoutData.Location,
						"product_positions":  len(layoutData.ProductPositionMap),
						"schema_validation":  "passed",
					})
				}
			} else {
				transformLogger.Warn("layout_metadata_reference_invalid", map[string]interface{}{
					"verification_id": event.VerificationID,
				})
			}
		} else {
			transformLogger.Info("layout_metadata_not_available", map[string]interface{}{
				"verification_type": initData.VerificationContext.VerificationType,
				"impact":            "using_embedded_layout_data_from_initialization",
			})
		}
	}

	// STRATEGIC STAGE 5: Extract Turn1 response references - simplified approach
	var turn1ProcessedRef, turn1RawRef, turn1ConvRef models.S3Reference

	if v, exists := event.S3References["responses_turn1Processed"]; exists {
		if ref, ok := convertToS3Reference(v); ok {
			turn1ProcessedRef = ref
		}
	}
	if v, exists := event.S3References["responses_turn1Raw"]; exists {
		if ref, ok := convertToS3Reference(v); ok {
			turn1RawRef = ref
		}
	}
	if v, exists := event.S3References["conversation_turn1"]; exists {
		if ref, ok := convertToS3Reference(v); ok {
			turn1ConvRef = ref
		}
	}

	if turn1ProcessedRef.Key == "" {
		if v, exists := event.S3References["processing_turn1_processed_response"]; exists {
			if ref, ok := convertToS3Reference(v); ok {
				turn1ProcessedRef = ref
			}
		}
	}
	if turn1RawRef.Key == "" {
		if v, exists := event.S3References["responses_turn1_raw_response"]; exists {
			if ref, ok := convertToS3Reference(v); ok {
				turn1RawRef = ref
			}
		}
	}

	if responsesVal, exists := event.S3References["responses"]; exists {
		if respMap, ok := responsesVal.(map[string]interface{}); ok {
			if turn1ProcessedRef.Key == "" {
				if v, ok2 := respMap["turn1Processed"]; ok2 {
					if ref, ok3 := convertToS3Reference(v); ok3 {
						turn1ProcessedRef = ref
					}
				}
			}
			if turn1RawRef.Key == "" {
				if v, ok2 := respMap["turn1Raw"]; ok2 {
					if ref, ok3 := convertToS3Reference(v); ok3 {
						turn1RawRef = ref
					}
				}
			}
		}
	}

	transformLogger.Info("turn1_references_extracted", map[string]interface{}{
		"processed_key":  turn1ProcessedRef.Key,
		"raw_key":        turn1RawRef.Key,
		"processed_size": turn1ProcessedRef.Size,
		"raw_size":       turn1RawRef.Size,
	})

	// STRATEGIC VERIFICATION ID RESOLUTION
	// 1. Primary Source: Use VerificationId from the loaded initialization.json
	verificationID := initData.VerificationContext.VerificationId
	idSource := "initialization_data"

	// 2. Secondary Source: Fallback to the Step Function event's verificationId
	if verificationID == "" {
		if event.VerificationID != "" {
			verificationID = event.VerificationID
			idSource = "step_function_event"
			transformLogger.Warn("using_event_verification_id_as_fallback", map[string]interface{}{
				"verification_id": verificationID,
			})
		}
	}

	// 3. Tertiary Source: As a last resort, extract from the S3 key path
	if verificationID == "" {
		if extractedID := extractVerificationIDFromKey(initRef.Key); extractedID != "" {
			verificationID = extractedID
			idSource = "extracted_from_s3_key"
			transformLogger.Warn("extracted_verification_id_from_s3_key_as_last_resort", map[string]interface{}{
				"s3_key":          initRef.Key,
				"verification_id": verificationID,
			})
		}
	}

	transformLogger.Info("verification_id_resolved", map[string]interface{}{
		"final_verification_id": verificationID,
		"source":                idSource,
	})

	// STRATEGIC STAGE 6: Turn2Request construction with schema conversion
	req := &models.Turn2Request{
		VerificationID:      verificationID,
		VerificationContext: convertSchemaToLocalVerificationContext(initData.VerificationContext, initData.LayoutMetadata),
		S3Refs: models.Turn2RequestS3Refs{
			Prompts: models.PromptRefs{
				System: systemPromptRef,
			},
			Images: models.Turn2ImageRefs{
				CheckingBase64: models.S3Reference{
					Bucket: metadata.CheckingImage.StorageMetadata.Bucket,
					Key:    metadata.CheckingImage.StorageMetadata.Key,
					Size:   metadata.CheckingImage.StorageMetadata.StoredSize,
				},
				CheckingImageFormat: schema.Base64Helpers.DetectImageFormat(
					metadata.CheckingImage.OriginalMetadata.ContentType,
					metadata.CheckingImage.OriginalMetadata.SourceKey,
				),
			},
			Processing: models.ProcessingReferences{},
			Turn1: models.Turn1References{
				ProcessedResponse: turn1ProcessedRef,
				RawResponse:       turn1RawRef,
				Conversation:      turn1ConvRef,
			},
		},
		InputInitializationFileRef: initRef,
		InputS3References:          convertS3ReferencesToInterface(event.S3References),
	}

	// Populate historical context S3 reference for PREVIOUS_VS_CURRENT
	if initData.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		if v, ok := event.S3References["processing_historical_context"]; ok {
			if histCtxRef, ok := convertToS3Reference(v); ok {
				req.S3Refs.Processing.HistoricalContext = histCtxRef
				transformLogger.Info("historical_context_s3_reference_found_in_event", map[string]interface{}{
					"bucket": histCtxRef.Bucket,
					"key":    histCtxRef.Key,
				})
			}
		} else {
			transformLogger.Warn("s3_reference_for_processing_historical_context_not_found_in_event_for_uc2", map[string]interface{}{
				"verification_id": event.VerificationID,
			})
		}
	}

	// Additional debug check for verificationID after construction
	if req.VerificationID == "" {
		transformLogger.Error("CRITICAL_verification_id_empty_after_construction", map[string]interface{}{
			"event_verification_id":       event.VerificationID,
			"init_verification_id":        initData.VerificationContext.VerificationId,
			"final_verification_id":       verificationID,
			"req_verification_id":         req.VerificationID,
			"req_verification_id_pointer": &req.VerificationID,
		})
	} else {
		transformLogger.Info("verification_id_successfully_set", map[string]interface{}{
			"source": func() string {
				if verificationID == initData.VerificationContext.VerificationId {
					return "initialization_data"
				} else if verificationID == event.VerificationID {
					return "step_function_event"
				} else {
					return "extracted_from_s3_key"
				}
			}(),
			"verification_id": req.VerificationID,
		})
	}

	// STRATEGIC STAGE 6: Comprehensive transformation completion with schema validation
	transformLogger.Info("transformation_completed_successfully", map[string]interface{}{
		"verification_id":        req.VerificationID,
		"verification_type":      req.VerificationContext.VerificationType,
		"vending_machine_id":     req.VerificationContext.VendingMachineId,
		"layout_id":              req.VerificationContext.LayoutId,
		"layout_prefix":          req.VerificationContext.LayoutPrefix,
		"system_prompt_key":      req.S3Refs.Prompts.System.Key,
		"checking_image_key":     req.S3Refs.Images.CheckingBase64.Key,
		"checking_image_size":    req.S3Refs.Images.CheckingBase64.Size,
		"checking_image_format":  req.S3Refs.Images.CheckingImageFormat,
		"has_layout_metadata":    req.VerificationContext.LayoutMetadata != nil,
		"has_historical_context": req.VerificationContext.HistoricalContext != nil,
		"transformation_stages":  6,
		"schema_version":         schema.SchemaVersion,
		"integration_status":     "comprehensive_success",
	})

	return req, nil
}

// ===================================================================
// STRATEGIC SCHEMA CONVERSION FUNCTIONS
// ===================================================================

// convertSchemaToLocalVerificationContext converts shared schema to local model with comprehensive mapping
func convertSchemaToLocalVerificationContext(schemaCtx schema.VerificationContext, layoutMetadata *schema.LayoutMetadata) models.VerificationContext {
	localCtx := models.VerificationContext{
		VerificationAt:    schemaCtx.VerificationAt,
		VerificationType:  schemaCtx.VerificationType,
		VendingMachineId:  schemaCtx.VendingMachineId,
		LayoutId:          schemaCtx.LayoutId,
		LayoutPrefix:      schemaCtx.LayoutPrefix,
		LayoutMetadata:    extractLayoutMetadataMap(layoutMetadata),
		HistoricalContext: extractHistoricalContextMap(schemaCtx),
	}

	return localCtx
}

// extractLayoutMetadataMap extracts layout metadata into map format for backward compatibility
func extractLayoutMetadataMap(layoutMetadata *schema.LayoutMetadata) map[string]interface{} {
	if layoutMetadata == nil {
		return nil
	}

	return map[string]interface{}{
		"layoutId":           layoutMetadata.LayoutId,
		"layoutPrefix":       layoutMetadata.LayoutPrefix,
		"vendingMachineId":   layoutMetadata.VendingMachineId,
		"location":           layoutMetadata.Location,
		"machineStructure":   layoutMetadata.MachineStructure,
		"productPositionMap": layoutMetadata.ProductPositionMap,
		"referenceImageUrl":  layoutMetadata.ReferenceImageUrl,
		"sourceJsonUrl":      layoutMetadata.SourceJsonUrl,
		"createdAt":          layoutMetadata.CreatedAt,
		"updatedAt":          layoutMetadata.UpdatedAt,
	}
}

// extractHistoricalContextMap extracts historical context for PREVIOUS_VS_CURRENT verification
func extractHistoricalContextMap(schemaCtx schema.VerificationContext) map[string]interface{} {
	if schemaCtx.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return nil
	}

	// Extract historical context from schema VerificationContext
	historicalContext := make(map[string]interface{})

	if schemaCtx.PreviousVerificationId != "" {
		historicalContext["PreviousVerificationId"] = schemaCtx.PreviousVerificationId
	}

	// Additional historical context extraction can be added here
	// based on shared schema VerificationContext fields

	if len(historicalContext) == 0 {
		return nil
	}

	return historicalContext
}

// getMapKeys provides clean key extraction for debugging and validation
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// convertS3ReferencesToInterface converts map[string]models.S3Reference to map[string]interface{}
func convertS3ReferencesToInterface(refs map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(refs))
	for k, v := range refs {
		result[k] = v
	}
	return result
}

func convertToS3Reference(v interface{}) (models.S3Reference, bool) {
	if v == nil {
		return models.S3Reference{}, false
	}
	switch val := v.(type) {
	case models.S3Reference:
		return val, true
	case map[string]interface{}:
		var ref models.S3Reference
		if b, ok := val["bucket"].(string); ok {
			ref.Bucket = b
		}
		if k, ok := val["key"].(string); ok {
			ref.Key = k
		}
		if s, ok := val["size"]; ok {
			switch sz := s.(type) {
			case float64:
				ref.Size = int64(sz)
			case int:
				ref.Size = int64(sz)
			case int64:
				ref.Size = sz
			}
		}
		if ref.Bucket == "" && ref.Key == "" && ref.Size == 0 {
			return models.S3Reference{}, false
		}
		return ref, true
	default:
		return models.S3Reference{}, false
	}
}

// extractVerificationIDFromKey extracts verification ID from S3 key
func extractVerificationIDFromKey(key string) string {
	parts := strings.Split(key, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "verif-") {
			return part
		}
	}
	return ""
}
