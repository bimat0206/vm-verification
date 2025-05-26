// internal/services/s3.go - STRATEGIC SCHEMA INTEGRATION ARCHITECTURE
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// ===================================================================
// STRATEGIC TYPE DEFINITIONS WITH SCHEMA INTEGRATION
// ===================================================================

// InitializationData leverages shared schema types for architectural coherence
type InitializationData struct {
	SchemaVersion       string                     `json:"schemaVersion"`
	VerificationContext schema.VerificationContext `json:"verificationContext"`      // SHARED SCHEMA INTEGRATION
	SystemPrompt        SystemPromptData           `json:"systemPrompt"`             // ENHANCED STRUCTURE
	LayoutMetadata      *schema.LayoutMetadata     `json:"layoutMetadata,omitempty"` // DIRECT SCHEMA USAGE
}

// SystemPromptData provides strategic mapping for system prompt structure
type SystemPromptData struct {
	Content       string `json:"content"`
	PromptID      string `json:"promptId"`
	PromptVersion string `json:"promptVersion"`
}

// ImageMetadata leverages shared schema concepts with S3-specific enhancements
type ImageMetadata struct {
	VerificationID     string                 `json:"verificationId"`
	VerificationType   string                 `json:"verificationType"`
	ReferenceImage     ImageInfoEnhanced      `json:"referenceImage"`
	CheckingImage      ImageInfoEnhanced      `json:"checkingImage"`
	ProcessingMetadata ProcessingMetadataInfo `json:"processingMetadata"`
	Version            string                 `json:"version"`
}

// ImageInfoEnhanced provides comprehensive image information with S3 storage metadata
type ImageInfoEnhanced struct {
	OriginalMetadata OriginalImageMetadata `json:"originalMetadata"`
	Base64Metadata   Base64ImageMetadata   `json:"base64Metadata"`
	StorageMetadata  StorageImageMetadata  `json:"storageMetadata"`
	ImageType        string                `json:"imageType"`
	Validation       ImageValidation       `json:"validation"`
}

// Strategic metadata structures for comprehensive image processing
type OriginalImageMetadata struct {
	SourceURL       string          `json:"sourceUrl"`
	SourceBucket    string          `json:"sourceBucket"`
	SourceKey       string          `json:"sourceKey"`
	ContentType     string          `json:"contentType"`
	OriginalSize    int64           `json:"originalSize"`
	ImageDimensions ImageDimensions `json:"imageDimensions"`
}

type Base64ImageMetadata struct {
	OriginalSize      int64           `json:"originalSize"`
	EncodedSize       int64           `json:"encodedSize"`
	EncodingTimestamp string          `json:"encodingTimestamp"`
	EncodingMethod    string          `json:"encodingMethod"`
	CompressionRatio  float64         `json:"compressionRatio"`
	QualitySettings   QualitySettings `json:"qualitySettings"`
}

type StorageImageMetadata struct {
	Bucket       string     `json:"bucket"`
	Key          string     `json:"key"`
	StoredSize   int64      `json:"storedSize"`
	StorageClass string     `json:"storageClass"`
	Encryption   Encryption `json:"encryption"`
}

type ImageDimensions struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio float64 `json:"aspectRatio"`
}

type QualitySettings struct {
	Optimized bool `json:"optimized"`
}

type Encryption struct {
	Method string `json:"method"`
}

type ImageValidation struct {
	IsValid           bool `json:"isValid"`
	BedrockCompatible bool `json:"bedrockCompatible"`
	SizeWithinLimits  bool `json:"sizeWithinLimits"`
}

type ProcessingMetadataInfo struct {
	ProcessedAt          string   `json:"processedAt"`
	ProcessingTimeMs     int64    `json:"processingTimeMs"`
	TotalImagesProcessed int      `json:"totalImagesProcessed"`
	ProcessingSteps      []string `json:"processingSteps"`
	ParallelProcessing   bool     `json:"parallelProcessing"`
}

// ===================================================================
// STRATEGIC S3 STATE MANAGER INTERFACE WITH SCHEMA INTEGRATION
// ===================================================================

// S3StateManager defines comprehensive S3 operations with strategic schema integration
type S3StateManager interface {
	// Core content loading operations (preserved backward compatibility)
	LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error)
	LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error)

	// Generic JSON loading capability (architectural enhancement)
	LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error
	// StoreJSONAtReference stores arbitrary JSON data at the provided S3 reference
	StoreJSONAtReference(ctx context.Context, ref models.S3Reference, data interface{}) (models.S3Reference, error)

	// STRATEGIC: Schema-integrated specialized loaders
	LoadInitializationData(ctx context.Context, ref models.S3Reference) (*InitializationData, error)
	LoadImageMetadata(ctx context.Context, ref models.S3Reference) (*ImageMetadata, error)
	LoadLayoutMetadata(ctx context.Context, ref models.S3Reference) (*schema.LayoutMetadata, error)

	// Schema validation integration
	ValidateInitializationData(ctx context.Context, data *InitializationData) error
	ValidateImageMetadata(ctx context.Context, metadata *ImageMetadata) error

	// Storage operations with schema integration
	StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error)
	StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error)
	StorePrompt(ctx context.Context, verificationID string, turn int, prompt interface{}) (models.S3Reference, error)
	StoreProcessedTurn1Response(ctx context.Context, verificationID string, analysisData *bedrockparser.ParsedTurn1Data) (models.S3Reference, error)
	StoreProcessedTurn1Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error)
	StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error)
	StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error)
	StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error)
	LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error)

	// STRATEGIC: Schema-based workflow state operations
	StoreWorkflowState(ctx context.Context, verificationID string, state *schema.WorkflowState) (models.S3Reference, error)
	LoadWorkflowState(ctx context.Context, verificationID string) (*schema.WorkflowState, error)
}

// ===================================================================
// STRATEGIC S3 MANAGER IMPLEMENTATION WITH COMPREHENSIVE INTEGRATION
// ===================================================================

// s3Manager implements strategic S3 state management with schema integration
type s3Manager struct {
	stateManager s3state.Manager
	bucket       string
	logger       logger.Logger
	cfg          config.Config
}

func (m *s3Manager) datePath(verificationID string) string {
	partition := m.cfg.CurrentDatePartition()
	return fmt.Sprintf("%s/%s", partition, verificationID)
}

// NewS3StateManager creates a strategically enhanced S3StateManager
func NewS3StateManager(cfg config.Config, log logger.Logger) (S3StateManager, error) {
	bucket := cfg.AWS.S3Bucket
	log.Info("s3_state_manager_initialization", map[string]interface{}{
		"bucket":         bucket,
		"schema_version": schema.SchemaVersion,
		"integration":    "comprehensive_schema_integration",
	})

	mgr, err := s3state.New(bucket)
	if err != nil {
		log.Error("s3_state_manager_creation_failed", map[string]interface{}{
			"bucket": bucket,
			"error":  err.Error(),
		})
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to create S3 state manager", false).
			WithContext("bucket", bucket)
	}

	return &s3Manager{
		stateManager: mgr,
		bucket:       bucket,
		logger:       log,
		cfg:          cfg,
	}, nil
}

// ===================================================================
// CORE LOADING OPERATIONS (PRESERVED BACKWARD COMPATIBILITY)
// ===================================================================

// LoadSystemPrompt retrieves system prompt from rich JSON format
func (m *s3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_system_prompt_started", map[string]interface{}{
		"bucket":         ref.Bucket,
		"key":            ref.Key,
		"size":           ref.Size,
		"operation":      "system_prompt_load",
		"schema_version": schema.SchemaVersion,
	})

	if err := m.validateReference(ref, "system_prompt"); err != nil {
		return "", err
	}

	// Get raw bytes first - don't unmarshal yet
	stateRef := m.toStateReference(ref)
	raw, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_system_prompt_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "get_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read system prompt: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return "", wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Parse the rich JSON object
	var wrapper struct {
		PromptContent struct {
			SystemMessage string `json:"systemMessage"`
		} `json:"promptContent"`
	}

	if err := json.Unmarshal(raw, &wrapper); err != nil {
		m.logger.Error("s3_system_prompt_format_error", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
			"bytes":  len(raw),
		})
		return "", &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "BadSystemPrompt",
			Message:   fmt.Sprintf("expected rich JSON object, got err %v", err),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	// Validate we got the system message
	if wrapper.PromptContent.SystemMessage == "" {
		return "", &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "MissingSystemMessage",
			Message:   "promptContent.systemMessage is empty or missing",
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	duration := time.Since(startTime)
	m.logger.Debug("system_prompt_loaded", map[string]interface{}{
		"format":         "rich",
		"bytes":          len(raw),
		"message_length": len(wrapper.PromptContent.SystemMessage),
		"duration_ms":    duration.Milliseconds(),
	})

	m.logger.Info("s3_system_prompt_loaded_successfully", map[string]interface{}{
		"bucket":         ref.Bucket,
		"key":            ref.Key,
		"prompt_length":  len(wrapper.PromptContent.SystemMessage),
		"duration_ms":    duration.Milliseconds(),
		"prompt_preview": truncateForLog(wrapper.PromptContent.SystemMessage, 100),
	})

	return wrapper.PromptContent.SystemMessage, nil
}

// LoadBase64Image retrieves Base64 image from .base64 file
func (m *s3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_base64_image_started", map[string]interface{}{
		"bucket":        ref.Bucket,
		"key":           ref.Key,
		"expected_size": ref.Size,
		"operation":     "base64_image_load",
	})

	if err := m.validateReference(ref, "base64_image"); err != nil {
		return "", err
	}

	// Validate file extension
	if !strings.HasSuffix(ref.Key, ".base64") {
		m.logger.Error("s3_base64_image_invalid_extension", map[string]interface{}{
			"bucket": ref.Bucket,
			"key":    ref.Key,
		})
		return "", &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "ExpectBase64Ext",
			Message:   "image object must end with .base64",
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	// Get raw bytes from the .base64 file
	stateRef := m.toStateReference(ref)
	rawBytes, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_base64_image_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "retrieve_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read base64 image: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return "", wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Trim any whitespace from the base64 content
	b := bytes.TrimSpace(rawBytes)
	if len(b) == 0 {
		return "", &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "EmptyBase64",
			Message:   "base64 file is empty",
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	duration := time.Since(startTime)
	m.logger.Debug("image_loaded", map[string]interface{}{
		"format":      ".base64",
		"bytes":       len(b),
		"duration_ms": duration.Milliseconds(),
	})

	m.logger.Info("s3_base64_image_loaded_successfully", map[string]interface{}{
		"bucket":            ref.Bucket,
		"key":               ref.Key,
		"image_data_length": len(b),
		"duration_ms":       duration.Milliseconds(),
		"size_ratio":        calculateSizeRatio(len(b), ref.Size),
	})

	return string(b), nil
}

// LoadJSON provides strategic generic JSON loading with comprehensive error handling
func (m *s3Manager) LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error {
	startTime := time.Now()
	m.logger.Debug("s3_loading_generic_json_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "generic_json_load",
	})

	if err := m.validateReference(ref, "generic_json"); err != nil {
		return err
	}

	stateRef := m.toStateReference(ref)

	if err := m.stateManager.RetrieveJSON(stateRef, target); err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_generic_json_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
		})
		return errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load JSON data", true).
			WithContext("s3_key", ref.Key).
			WithContext("bucket", ref.Bucket).
			WithContext("duration_ms", duration.Milliseconds())
	}

	duration := time.Since(startTime)
	m.logger.Debug("s3_generic_json_loaded_successfully", map[string]interface{}{
		"bucket":      ref.Bucket,
		"key":         ref.Key,
		"duration_ms": duration.Milliseconds(),
	})

	return nil
}

// ===================================================================
// STRATEGIC SCHEMA-INTEGRATED SPECIALIZED LOADERS
// ===================================================================

// LoadInitializationData loads Step Functions initialization data with schema integration
func (m *s3Manager) LoadInitializationData(ctx context.Context, ref models.S3Reference) (*InitializationData, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_initialization_data_started", map[string]interface{}{
		"bucket":         ref.Bucket,
		"key":            ref.Key,
		"size":           ref.Size,
		"operation":      "initialization_data_load",
		"schema_version": schema.SchemaVersion,
	})

	var initData InitializationData

	if err := m.LoadJSON(ctx, ref, &initData); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load initialization data", true).
			WithContext("data_type", "initialization").
			WithContext("schema_version", schema.SchemaVersion)
	}

	// Strategic schema validation integration
	if err := m.ValidateInitializationData(ctx, &initData); err != nil {
		m.logger.Warn("initialization_data_validation_warnings", map[string]interface{}{
			"validation_errors": err.Error(),
			"verification_id":   initData.VerificationContext.VerificationId,
		})
	}

	duration := time.Since(startTime)
	m.logger.Info("initialization_data_loaded_successfully", map[string]interface{}{
		"verification_id":       initData.VerificationContext.VerificationId,
		"verification_type":     initData.VerificationContext.VerificationType,
		"vending_machine_id":    initData.VerificationContext.VendingMachineId,
		"layout_id":             initData.VerificationContext.LayoutId,
		"layout_prefix":         initData.VerificationContext.LayoutPrefix,
		"system_prompt_id":      initData.SystemPrompt.PromptID,
		"system_prompt_version": initData.SystemPrompt.PromptVersion,
		"has_layout_metadata":   initData.LayoutMetadata != nil,
		"schema_version":        initData.SchemaVersion,
		"duration_ms":           duration.Milliseconds(),
	})

	return &initData, nil
}

// LoadImageMetadata loads image processing metadata with comprehensive validation
func (m *s3Manager) LoadImageMetadata(ctx context.Context, ref models.S3Reference) (*ImageMetadata, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_image_metadata_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "image_metadata_load",
	})

	var metadata ImageMetadata

	if err := m.LoadJSON(ctx, ref, &metadata); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load image metadata", true).
			WithContext("data_type", "image_metadata")
	}

	// Strategic metadata validation
	if err := m.ValidateImageMetadata(ctx, &metadata); err != nil {
		m.logger.Warn("image_metadata_validation_warnings", map[string]interface{}{
			"validation_errors": err.Error(),
			"verification_id":   metadata.VerificationID,
		})
	}

	duration := time.Since(startTime)
	m.logger.Info("image_metadata_loaded_successfully", map[string]interface{}{
		"verification_id":        metadata.VerificationID,
		"verification_type":      metadata.VerificationType,
		"reference_image_bucket": metadata.ReferenceImage.StorageMetadata.Bucket,
		"reference_image_key":    metadata.ReferenceImage.StorageMetadata.Key,
		"reference_image_size":   metadata.ReferenceImage.StorageMetadata.StoredSize,
		"checking_image_bucket":  metadata.CheckingImage.StorageMetadata.Bucket,
		"checking_image_key":     metadata.CheckingImage.StorageMetadata.Key,
		"checking_image_size":    metadata.CheckingImage.StorageMetadata.StoredSize,
		"total_images_processed": metadata.ProcessingMetadata.TotalImagesProcessed,
		"parallel_processing":    metadata.ProcessingMetadata.ParallelProcessing,
		"processing_time_ms":     metadata.ProcessingMetadata.ProcessingTimeMs,
		"duration_ms":            duration.Milliseconds(),
	})

	return &metadata, nil
}

// LoadLayoutMetadata loads vending machine layout configuration using shared schema
func (m *s3Manager) LoadLayoutMetadata(ctx context.Context, ref models.S3Reference) (*schema.LayoutMetadata, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_layout_metadata_started", map[string]interface{}{
		"bucket":         ref.Bucket,
		"key":            ref.Key,
		"size":           ref.Size,
		"operation":      "layout_metadata_load",
		"schema_version": schema.SchemaVersion,
	})

	var layoutData schema.LayoutMetadata

	if err := m.LoadJSON(ctx, ref, &layoutData); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load layout metadata", true).
			WithContext("data_type", "layout_metadata").
			WithContext("schema_version", schema.SchemaVersion)
	}

	duration := time.Since(startTime)
	m.logger.Info("layout_metadata_loaded_successfully", map[string]interface{}{
		"layout_id":           layoutData.LayoutId,
		"layout_prefix":       layoutData.LayoutPrefix,
		"vending_machine_id":  layoutData.VendingMachineId,
		"location":            layoutData.Location,
		"machine_structure":   layoutData.MachineStructure,
		"product_positions":   len(layoutData.ProductPositionMap),
		"reference_image_url": layoutData.ReferenceImageUrl,
		"created_at":          layoutData.CreatedAt,
		"updated_at":          layoutData.UpdatedAt,
		"duration_ms":         duration.Milliseconds(),
	})

	return &layoutData, nil
}

// ===================================================================
// STRATEGIC SCHEMA VALIDATION INTEGRATION
// ===================================================================

// ValidateInitializationData performs comprehensive validation using shared schema
func (m *s3Manager) ValidateInitializationData(ctx context.Context, data *InitializationData) error {
	if data == nil {
		return errors.NewValidationError("initialization data cannot be nil", map[string]interface{}{})
	}

	// Strategic schema validation integration
	if validationErrors := schema.ValidateVerificationContextEnhanced(&data.VerificationContext); len(validationErrors) > 0 {
		m.logger.Debug("verification_context_validation_issues", map[string]interface{}{
			"validation_errors": validationErrors.Error(),
			"verification_id":   data.VerificationContext.VerificationId,
		})

		// Return validation warning but don't fail the operation
		return errors.NewValidationError("verification context validation issues", map[string]interface{}{
			"details": validationErrors.Error(),
		})
	}

	// Additional initialization-specific validations
	if data.SystemPrompt.Content == "" {
		return errors.NewValidationError("system prompt content required", map[string]interface{}{
			"verification_id": data.VerificationContext.VerificationId,
		})
	}

	return nil
}

// ValidateImageMetadata performs comprehensive image metadata validation
func (m *s3Manager) ValidateImageMetadata(ctx context.Context, metadata *ImageMetadata) error {
	if metadata == nil {
		return errors.NewValidationError("image metadata cannot be nil", map[string]interface{}{})
	}

	// Validate image information structures
	if metadata.ReferenceImage.StorageMetadata.Bucket == "" || metadata.ReferenceImage.StorageMetadata.Key == "" {
		return errors.NewValidationError("reference image storage metadata incomplete", map[string]interface{}{
			"verification_id": metadata.VerificationID,
		})
	}

	if metadata.CheckingImage.StorageMetadata.Bucket == "" || metadata.CheckingImage.StorageMetadata.Key == "" {
		return errors.NewValidationError("checking image storage metadata incomplete", map[string]interface{}{
			"verification_id": metadata.VerificationID,
		})
	}

	// Validate image sizes and formats
	if !metadata.ReferenceImage.Validation.IsValid || !metadata.CheckingImage.Validation.IsValid {
		return errors.NewValidationError("image validation failed", map[string]interface{}{
			"verification_id":          metadata.VerificationID,
			"reference_image_valid":    metadata.ReferenceImage.Validation.IsValid,
			"checking_image_valid":     metadata.CheckingImage.Validation.IsValid,
			"reference_bedrock_compat": metadata.ReferenceImage.Validation.BedrockCompatible,
			"checking_bedrock_compat":  metadata.CheckingImage.Validation.BedrockCompatible,
		})
	}

	return nil
}

// ===================================================================
// ENHANCED STORAGE OPERATIONS WITH SCHEMA INTEGRATION
// ===================================================================

// StoreWorkflowState stores complete workflow state using shared schema
func (m *s3Manager) StoreWorkflowState(ctx context.Context, verificationID string, state *schema.WorkflowState) (models.S3Reference, error) {
	if verificationID == "" || state == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and workflow state required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"state_nil":             state == nil,
			})
	}

	// Strategic schema validation integration
	if validationErrors := schema.ValidateWorkflowState(state); len(validationErrors) > 0 {
		m.logger.Warn("workflow_state_validation_issues", map[string]interface{}{
			"validation_errors": validationErrors.Error(),
			"verification_id":   verificationID,
		})
	}

	key := fmt.Sprintf("%s/workflow-state.json", verificationID)
	stateRef, err := m.stateManager.StoreJSON("processing", key, state)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store workflow state", true).
			WithContext("verification_id", verificationID).
			WithContext("schema_version", state.SchemaVersion)
	}

	m.logger.Info("workflow_state_stored_successfully", map[string]interface{}{
		"verification_id": verificationID,
		"schema_version":  state.SchemaVersion,
		"bucket":          stateRef.Bucket,
		"key":             stateRef.Key,
	})

	return m.fromStateReference(stateRef), nil
}

// LoadWorkflowState loads complete workflow state with schema validation
func (m *s3Manager) LoadWorkflowState(ctx context.Context, verificationID string) (*schema.WorkflowState, error) {
	if verificationID == "" {
		return nil, errors.NewValidationError(
			"verification ID required",
			map[string]interface{}{"operation": "load_workflow_state"})
	}

	key := fmt.Sprintf("processing/%s/workflow-state.json", verificationID)
	stateRef := &s3state.Reference{
		Bucket: m.bucket,
		Key:    key,
	}

	var state schema.WorkflowState
	if err := m.stateManager.RetrieveJSON(stateRef, &state); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load workflow state", true).
			WithContext("verification_id", verificationID)
	}

	// Strategic schema validation integration
	if validationErrors := schema.ValidateWorkflowState(&state); len(validationErrors) > 0 {
		m.logger.Warn("loaded_workflow_state_validation_issues", map[string]interface{}{
			"validation_errors": validationErrors.Error(),
			"verification_id":   verificationID,
		})
	}

	return &state, nil
}

// ===================================================================
// PRESERVED STORAGE OPERATIONS (BACKWARD COMPATIBILITY)
// ===================================================================

// StoreRawResponse stores raw Bedrock response with enhanced logging
func (m *s3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing raw response",
			map[string]interface{}{"operation": "store_raw_response"})
	}

	key := "responses/turn1-raw-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, data)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store raw response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	m.logger.Info("raw_response_stored_successfully", map[string]interface{}{
		"verification_id": verificationID,
		"bucket":          stateRef.Bucket,
		"key":             stateRef.Key,
		"size":            stateRef.Size,
	})

	return m.fromStateReference(stateRef), nil
}

// StoreProcessedAnalysis stores processed analysis results with validation
func (m *s3Manager) StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing processed analysis",
			map[string]interface{}{"operation": "store_processed_analysis"})
	}

	key := "processing/turn1-processed-analysis.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, analysis)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processed analysis", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreProcessedTurn1Response stores the parsed Turn1 response structure
func (m *s3Manager) StoreProcessedTurn1Response(ctx context.Context, verificationID string, analysisData *bedrockparser.ParsedTurn1Data) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing processed turn1 response",
			map[string]interface{}{"operation": "store_processed_turn1_response"})
	}
	if analysisData == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"analysisData cannot be nil for storing processed turn1 response",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "processing/turn1-processed-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, analysisData)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processed turn1 response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreProcessedTurn1Markdown stores the Markdown version of the Turn 1 analysis
func (m *s3Manager) StoreProcessedTurn1Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing processed turn1 markdown",
			map[string]interface{}{"operation": "store_processed_turn1_markdown"})
	}
	if strings.TrimSpace(markdownContent) == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"markdownContent cannot be empty",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "responses/turn1-processed-response.md"
	dataBytes := []byte(markdownContent)
	stateRef, err := m.stateManager.StoreWithContentType(m.datePath(verificationID), key, dataBytes, "text/markdown; charset=utf-8")
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processed turn1 markdown", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	return m.fromStateReference(stateRef), nil
}

// StorePrompt stores a rendered prompt JSON
func (m *s3Manager) StorePrompt(ctx context.Context, verificationID string, turn int, prompt interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing prompt",
			map[string]interface{}{"operation": "store_prompt"})
	}

	key := fmt.Sprintf("prompts/turn%d-prompt.json", turn)
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, prompt)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store prompt", true).
			WithContext("verification_id", verificationID)
	}

	return m.fromStateReference(stateRef), nil
}

// StoreConversationTurn stores conversation turn data with schema validation
func (m *s3Manager) StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error) {
	if verificationID == "" || turnData == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and turn data required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"turn_data_nil":         turnData == nil,
			})
	}

	key := fmt.Sprintf("responses/conversation-turn%d.json", turnData.TurnId)
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, turnData)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store conversation turn", true).
			WithContext("verification_id", verificationID).
			WithContext("turn_id", turnData.TurnId)
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTemplateProcessor stores template processing results with validation
func (m *s3Manager) StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error) {
	if verificationID == "" || processor == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and template processor required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"processor_nil":         processor == nil,
			})
	}

	// Strategic template processor validation
	if validationErrors := schema.ValidateTemplateProcessor(processor); len(validationErrors) > 0 {
		m.logger.Warn("template_processor_validation_issues", map[string]interface{}{
			"validation_errors": validationErrors.Error(),
			"verification_id":   verificationID,
		})
	}

	key := "processing/template-processor.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, processor)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store template processor", true).
			WithContext("verification_id", verificationID).
			WithContext("template_id", processor.Template.TemplateId)
	}

	return m.fromStateReference(stateRef), nil
}

// StoreProcessingMetrics stores processing metrics with schema validation
func (m *s3Manager) StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error) {
	if verificationID == "" || metrics == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and processing metrics required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"metrics_nil":           metrics == nil,
			})
	}

	// Strategic processing metrics validation
	if validationErrors := schema.ValidateProcessingMetrics(metrics); len(validationErrors) > 0 {
		m.logger.Warn("processing_metrics_validation_issues", map[string]interface{}{
			"validation_errors": validationErrors.Error(),
			"verification_id":   verificationID,
		})
	}

	key := "processing/processing-metrics.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, metrics)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processing metrics", true).
			WithContext("verification_id", verificationID)
	}

	return m.fromStateReference(stateRef), nil
}

// LoadProcessingState loads processing state by type (preserved backward compatibility)
func (m *s3Manager) LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error) {
	if verificationID == "" || stateType == "" {
		return nil, errors.NewValidationError(
			"verification ID and state type required",
			map[string]interface{}{
				"verification_id": verificationID,
				"state_type":      stateType,
			})
	}

	key := fmt.Sprintf("processing/%s.json", stateType)
	stateRef := &s3state.Reference{
		Bucket: m.bucket,
		Key:    fmt.Sprintf("%s/%s", m.datePath(verificationID), key),
	}

	var result interface{}
	if err := m.stateManager.RetrieveJSON(stateRef, &result); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load processing state", true).
			WithContext("verification_id", verificationID).
			WithContext("state_type", stateType)
	}

	return result, nil
}
// StoreJSONAtReference stores arbitrary JSON data at the given S3 reference
func (m *s3Manager) StoreJSONAtReference(ctx context.Context, ref models.S3Reference, data interface{}) (models.S3Reference, error) {
	if err := m.validateReference(ref, "store_json_at_reference"); err != nil {
		return models.S3Reference{}, err
	}

	stateRef, err := m.stateManager.StoreJSON("", ref.Key, data)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store JSON data", true).
			WithContext("bucket", ref.Bucket).
			WithContext("s3_key", ref.Key)
	}

	return m.fromStateReference(stateRef), nil
}
// ===================================================================
// STRATEGIC UTILITY FUNCTIONS
// ===================================================================

// validateReference performs comprehensive input validation
func (m *s3Manager) validateReference(ref models.S3Reference, operation string) error {
	if ref.Bucket == "" {
		return errors.NewValidationError(
			"S3 bucket required",
			map[string]interface{}{"operation": operation})
	}

	if ref.Key == "" {
		return errors.NewValidationError(
			"S3 key required",
			map[string]interface{}{"operation": operation})
	}

	return nil
}

// toStateReference converts model reference to state reference
func (m *s3Manager) toStateReference(ref models.S3Reference) *s3state.Reference {
	return &s3state.Reference{
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   ref.Size,
	}
}

// fromStateReference converts state reference to model reference
func (m *s3Manager) fromStateReference(ref *s3state.Reference) models.S3Reference {
	return models.S3Reference{
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   ref.Size,
	}
}

// truncateForLog provides safe string truncation for logging
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// calculateSizeRatio computes size ratio for validation
func calculateSizeRatio(actual int, expected int64) float64 {
	if expected == 0 {
		return 0
	}
	return float64(actual) / float64(expected)
}
