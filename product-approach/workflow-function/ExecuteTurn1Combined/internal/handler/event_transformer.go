package handler

import (
	"context"
	"encoding/json"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// EventTransformer handles transformation of Step Functions events to internal request format
type EventTransformer struct {
	s3  services.S3StateManager
	log logger.Logger
}

// NewEventTransformer creates a new instance of EventTransformer
func NewEventTransformer(s3 services.S3StateManager, log logger.Logger) *EventTransformer {
	return &EventTransformer{
		s3:  s3,
		log: log,
	}
}

// StepFunctionEvent represents the event format from Step Functions
type StepFunctionEvent struct {
	SchemaVersion  string                         `json:"schemaVersion"`
	S3References   map[string]models.S3Reference `json:"s3References"`
	VerificationID string                         `json:"verificationId"`
	Status         string                         `json:"status"`
}

// TransformStepFunctionEvent transforms the Step Functions event format to Turn1Request format
func (e *EventTransformer) TransformStepFunctionEvent(ctx context.Context, event StepFunctionEvent) (*models.Turn1Request, error) {
	// Create logger for this transformation
	transformLogger := e.log.WithFields(map[string]interface{}{
		"operation": "transform_step_function_event",
		"verification_id": event.VerificationID,
	})
	
	// 1. Load initialization.json to get verificationContext
	initRef, exists := event.S3References["processing_initialization"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing processing_initialization reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs": getMapKeys(event.S3References),
			})
	}
	
	transformLogger.Info("loading_initialization_data", map[string]interface{}{
		"bucket": initRef.Bucket,
		"key": initRef.Key,
		"size": initRef.Size,
	})
	
	// Load initialization data
	var initData struct {
		SchemaVersion       string                      `json:"schemaVersion"`
		VerificationContext models.VerificationContext `json:"verificationContext"`
		SystemPrompt        struct {
			Content         string `json:"content"`
			PromptID        string `json:"promptId"`
			PromptVersion   string `json:"promptVersion"`
		} `json:"systemPrompt"`
		LayoutMetadata interface{} `json:"layoutMetadata"`
	}
	
	initJSON, err := e.s3.LoadSystemPrompt(ctx, initRef) // Using LoadSystemPrompt as it loads JSON
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load initialization data", true).
			WithContext("s3_key", initRef.Key).
			WithContext("verification_id", event.VerificationID)
	}
	
	if err := json.Unmarshal([]byte(initJSON), &initData); err != nil {
		return nil, errors.NewValidationError(
			"failed to parse initialization data",
			map[string]interface{}{
				"parse_error": err.Error(),
				"verification_id": event.VerificationID,
			})
	}
	
	transformLogger.Info("initialization_data_loaded", map[string]interface{}{
		"verification_type": initData.VerificationContext.VerificationType,
		"layout_id": initData.VerificationContext.LayoutId,
		"layout_prefix": initData.VerificationContext.LayoutPrefix,
		"vending_machine_id": initData.VerificationContext.VendingMachineId,
	})
	
	// 2. Load metadata.json to get image references
	metadataRef, exists := event.S3References["images_metadata"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing images_metadata reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs": getMapKeys(event.S3References),
			})
	}
	
	transformLogger.Info("loading_image_metadata", map[string]interface{}{
		"bucket": metadataRef.Bucket,
		"key": metadataRef.Key,
		"size": metadataRef.Size,
	})
	
	// Load metadata
	var metadata struct {
		VerificationID   string `json:"verificationId"`
		VerificationType string `json:"verificationType"`
		ReferenceImage   struct {
			StorageMetadata struct {
				Bucket string `json:"bucket"`
				Key    string `json:"key"`
				Size   int64  `json:"storedSize"`
			} `json:"storageMetadata"`
			Base64Metadata struct {
				EncodedSize int64 `json:"encodedSize"`
			} `json:"base64Metadata"`
		} `json:"referenceImage"`
		CheckingImage struct {
			StorageMetadata struct {
				Bucket string `json:"bucket"`
				Key    string `json:"key"`
				Size   int64  `json:"storedSize"`
			} `json:"storageMetadata"`
		} `json:"checkingImage"`
	}
	
	metadataJSON, err := e.s3.LoadSystemPrompt(ctx, metadataRef) // Using LoadSystemPrompt as it loads JSON
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load image metadata", true).
			WithContext("s3_key", metadataRef.Key).
			WithContext("verification_id", event.VerificationID)
	}
	
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return nil, errors.NewValidationError(
			"failed to parse image metadata",
			map[string]interface{}{
				"parse_error": err.Error(),
				"verification_id": event.VerificationID,
			})
	}
	
	transformLogger.Info("image_metadata_loaded", map[string]interface{}{
		"reference_image_bucket": metadata.ReferenceImage.StorageMetadata.Bucket,
		"reference_image_key": metadata.ReferenceImage.StorageMetadata.Key,
		"reference_image_size": metadata.ReferenceImage.StorageMetadata.Size,
	})
	
	// 3. Get system prompt reference
	systemPromptRef, exists := event.S3References["prompts_system"]
	if !exists {
		return nil, errors.NewValidationError(
			"missing prompts_system reference",
			map[string]interface{}{
				"verification_id": event.VerificationID,
				"available_refs": getMapKeys(event.S3References),
			})
	}
	
	// 4. Load layout metadata if present and verificationType is LAYOUT_VS_CHECKING
	if initData.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		layoutMetadataRef, exists := event.S3References["processing_layout-metadata"]
		if exists {
			transformLogger.Info("loading_layout_metadata", map[string]interface{}{
				"bucket": layoutMetadataRef.Bucket,
				"key": layoutMetadataRef.Key,
			})
			
			layoutMetadataJSON, err := e.s3.LoadSystemPrompt(ctx, layoutMetadataRef)
			if err != nil {
				transformLogger.Warn("failed_to_load_layout_metadata", map[string]interface{}{
					"error": err.Error(),
					"key": layoutMetadataRef.Key,
				})
			} else {
				var layoutData map[string]interface{}
				if err := json.Unmarshal([]byte(layoutMetadataJSON), &layoutData); err == nil {
					initData.VerificationContext.LayoutMetadata = layoutData
					transformLogger.Info("layout_metadata_loaded", map[string]interface{}{
						"layout_id": layoutData["layoutId"],
						"layout_prefix": layoutData["layoutPrefix"],
					})
				}
			}
		}
	}
	
	// 5. Build the Turn1Request structure
	req := &models.Turn1Request{
		VerificationID:      event.VerificationID,
		VerificationContext: initData.VerificationContext,
		S3Refs: models.Turn1RequestS3Refs{
			Prompts: models.PromptRefs{
				System: systemPromptRef,
			},
			Images: models.ImageRefs{
				ReferenceBase64: models.S3Reference{
					Bucket: metadata.ReferenceImage.StorageMetadata.Bucket,
					Key:    metadata.ReferenceImage.StorageMetadata.Key,
					Size:   metadata.ReferenceImage.StorageMetadata.Size,
				},
			},
		},
	}
	
	transformLogger.Info("transformation_completed", map[string]interface{}{
		"verification_id": req.VerificationID,
		"verification_type": req.VerificationContext.VerificationType,
		"system_prompt_key": req.S3Refs.Prompts.System.Key,
		"reference_image_key": req.S3Refs.Images.ReferenceBase64.Key,
	})
	
	return req, nil
}

// Helper function to get map keys for logging
func getMapKeys(m map[string]models.S3Reference) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}