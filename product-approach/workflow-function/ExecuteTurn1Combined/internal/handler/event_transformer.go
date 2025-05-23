// internal/handler/event_transformer.go - COMPREHENSIVE SCHEMA INTEGRATION
package handler

import (
	"context"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
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
	SchemaVersion  string                         `json:"schemaVersion"`
	S3References   map[string]models.S3Reference `json:"s3References"`
	VerificationID string                         `json:"verificationId"`
	Status         string                         `json:"status"`
}

// TransformStepFunctionEvent provides comprehensive transformation with schema integration
func (e *EventTransformer) TransformStepFunctionEvent(ctx context.Context, event StepFunctionEvent) (*models.Turn1Request, error) {
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
	initRef, exists := event.S3References["processing_initialization"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing processing_initialization reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs":  getMapKeys(event.S3References),
			})
	}
	
	transformLogger.Info("loading_initialization_data", map[string]interface{}{
		"bucket":         initRef.Bucket,
		"key":            initRef.Key,
		"size":           initRef.Size,
		"schema_version": schema.SchemaVersion,
	})
	
	// STRATEGIC RESOLUTION: Use schema-integrated initialization data loader
	initData, err := e.s3.LoadInitializationData(ctx, initRef)
	if err != nil {
		transformLogger.Error("initialization_data_load_failed", map[string]interface{}{
			"error":  err.Error(),
			"bucket": initRef.Bucket,
			"key":    initRef.Key,
		})
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load initialization data", true).
			WithContext("s3_key", initRef.Key).
			WithContext("verification_id", event.VerificationID).
			WithContext("schema_integration", "failed")
	}
	
	transformLogger.Info("initialization_data_loaded_successfully", map[string]interface{}{
		"verification_id":        initData.VerificationContext.VerificationId,
		"verification_type":      initData.VerificationContext.VerificationType,
		"vending_machine_id":     initData.VerificationContext.VendingMachineId,
		"layout_id":             initData.VerificationContext.LayoutId,
		"layout_prefix":         initData.VerificationContext.LayoutPrefix,
		"system_prompt_id":      initData.SystemPrompt.PromptID,
		"system_prompt_version": initData.SystemPrompt.PromptVersion,
		"has_layout_metadata":   initData.LayoutMetadata != nil,
		"schema_validation":     "passed",
	})
	
	// STRATEGIC STAGE 2: Load image metadata with comprehensive validation
	metadataRef, exists := event.S3References["images_metadata"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing images_metadata reference",
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
		"reference_image_bucket":     metadata.ReferenceImage.StorageMetadata.Bucket,
		"reference_image_key":        metadata.ReferenceImage.StorageMetadata.Key,
		"reference_image_size":       metadata.ReferenceImage.StorageMetadata.StoredSize,
		"reference_bedrock_compat":   metadata.ReferenceImage.Validation.BedrockCompatible,
		"checking_image_bucket":      metadata.CheckingImage.StorageMetadata.Bucket,
		"checking_image_key":         metadata.CheckingImage.StorageMetadata.Key,
		"checking_image_size":        metadata.CheckingImage.StorageMetadata.StoredSize,
		"checking_bedrock_compat":    metadata.CheckingImage.Validation.BedrockCompatible,
		"parallel_processing":        metadata.ProcessingMetadata.ParallelProcessing,
		"total_images_processed":     metadata.ProcessingMetadata.TotalImagesProcessed,
	})
	
	// STRATEGIC STAGE 3: Validate system prompt reference
	systemPromptRef, exists := event.S3References["prompts_system"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing prompts_system reference",
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
    if layoutMetadataRef, exists := event.S3References["processing_layout-metadata"]; exists {
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
            // Strategic schema integration: Populate VerificationContext LayoutMetadata
            if initData.VerificationContext.LayoutId == 0 {
                initData.VerificationContext.LayoutId = layoutData.LayoutId
            }
            if initData.VerificationContext.LayoutPrefix == "" {
                initData.VerificationContext.LayoutPrefix = layoutData.LayoutPrefix
            }
            if initData.VerificationContext.VendingMachineId == "" {
                initData.VerificationContext.VendingMachineId = layoutData.VendingMachineId
            }
            
            // Store the layout metadata for later use in the local verification context
            // This will be converted and included in the Turn1Request
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
        transformLogger.Info("layout_metadata_not_available", map[string]interface{}{
            "verification_type": initData.VerificationContext.VerificationType,
            "impact":           "using_embedded_layout_data_from_initialization",
        })
    }
}
	
	// STRATEGIC STAGE 5: Turn1Request construction with schema conversion
	req := &models.Turn1Request{
		VerificationID:      event.VerificationID,
		VerificationContext: convertSchemaToLocalVerificationContext(initData.VerificationContext, initData.LayoutMetadata),
		S3Refs: models.Turn1RequestS3Refs{
			Prompts: models.PromptRefs{
				System: systemPromptRef,
			},
			Images: models.ImageRefs{
				ReferenceBase64: models.S3Reference{
					Bucket: metadata.ReferenceImage.StorageMetadata.Bucket,
					Key:    metadata.ReferenceImage.StorageMetadata.Key,
					Size:   metadata.ReferenceImage.StorageMetadata.StoredSize,
				},
			},
		},
	}
	
	// STRATEGIC STAGE 6: Comprehensive transformation completion with schema validation
	transformLogger.Info("transformation_completed_successfully", map[string]interface{}{
		"verification_id":         req.VerificationID,
		"verification_type":       req.VerificationContext.VerificationType,
		"vending_machine_id":      req.VerificationContext.VendingMachineId,
		"layout_id":              req.VerificationContext.LayoutId,
		"layout_prefix":          req.VerificationContext.LayoutPrefix,
		"system_prompt_key":       req.S3Refs.Prompts.System.Key,
		"reference_image_key":     req.S3Refs.Images.ReferenceBase64.Key,
		"reference_image_size":    req.S3Refs.Images.ReferenceBase64.Size,
		"has_layout_metadata":     req.VerificationContext.LayoutMetadata != nil,
		"has_historical_context":  req.VerificationContext.HistoricalContext != nil,
		"transformation_stages":   6,
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
func getMapKeys(m map[string]models.S3Reference) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}