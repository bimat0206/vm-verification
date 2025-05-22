// internal/services/s3.go - CLEAN AND FOCUSED S3 STATE MANAGEMENT
package services

import (
	"context"
	"fmt"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// S3StateManager defines S3-based state persistence operations for ExecuteTurn1Combined
type S3StateManager interface {
	// Core operations for Turn1Combined
	LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error)
	LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error)
	StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error)
	StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error)
	
	// Enhanced operations for comprehensive tracking
	StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error)
	StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error)
	StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error)
	LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error)
}

// s3Manager implements S3StateManager with enhanced capabilities
type s3Manager struct {
	stateManager s3state.Manager
	bucket       string
}

// NewS3StateManager creates an enhanced S3StateManager
func NewS3StateManager(bucket string) (S3StateManager, error) {
	mgr, err := s3state.New(bucket)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to create S3 state manager", false).
			WithContext("bucket", bucket)
	}
	
	return &s3Manager{
		stateManager: mgr,
		bucket:       bucket,
	}, nil
}

// LoadSystemPrompt retrieves system prompt from S3
func (m *s3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	if err := m.validateReference(ref, "system_prompt"); err != nil {
		return "", err
	}
	
	var prompt string
	stateRef := m.toStateReference(ref)
	
	if err := m.stateManager.RetrieveJSON(stateRef, &prompt); err != nil {
		return "", errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load system prompt", true).
			WithContext("s3_key", ref.Key).
			WithContext("bucket", ref.Bucket)
	}
	
	return prompt, nil
}

// LoadBase64Image retrieves Base64 image data from S3
func (m *s3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	if err := m.validateReference(ref, "base64_image"); err != nil {
		return "", err
	}
	
	var imageData string
	stateRef := m.toStateReference(ref)
	
	if err := m.stateManager.RetrieveJSON(stateRef, &imageData); err != nil {
		return "", errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load Base64 image", true).
			WithContext("s3_key", ref.Key).
			WithContext("bucket", ref.Bucket).
			WithContext("estimated_size", ref.Size)
	}
	
	return imageData, nil
}

// StoreRawResponse stores raw Bedrock response
func (m *s3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing raw response",
			map[string]interface{}{"operation": "store_raw_response"})
	}
	
	key := fmt.Sprintf("%s/turn1-raw-response.json", verificationID)
	stateRef, err := m.stateManager.StoreJSON("responses", key, data)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store raw response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}
	
	return m.fromStateReference(stateRef), nil
}

// StoreProcessedAnalysis stores processed analysis results
func (m *s3Manager) StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing processed analysis",
			map[string]interface{}{"operation": "store_processed_analysis"})
	}
	
	key := fmt.Sprintf("%s/turn1-processed-analysis.json", verificationID)
	stateRef, err := m.stateManager.StoreJSON("processing", key, analysis)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processed analysis", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}
	
	return m.fromStateReference(stateRef), nil
}

// StoreConversationTurn stores conversation turn data
func (m *s3Manager) StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error) {
	if verificationID == "" || turnData == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and turn data required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"turn_data_nil":         turnData == nil,
			})
	}
	
	key := fmt.Sprintf("%s/conversation-turn%d.json", verificationID, turnData.TurnId)
	stateRef, err := m.stateManager.StoreJSON("responses", key, turnData)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store conversation turn", true).
			WithContext("verification_id", verificationID).
			WithContext("turn_id", turnData.TurnId)
	}
	
	return m.fromStateReference(stateRef), nil
}

// StoreTemplateProcessor stores template processing results
func (m *s3Manager) StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error) {
	if verificationID == "" || processor == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and template processor required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"processor_nil":         processor == nil,
			})
	}
	
	key := fmt.Sprintf("%s/template-processor.json", verificationID)
	stateRef, err := m.stateManager.StoreJSON("processing", key, processor)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store template processor", true).
			WithContext("verification_id", verificationID).
			WithContext("template_id", processor.Template.TemplateId)
	}
	
	return m.fromStateReference(stateRef), nil
}

// StoreProcessingMetrics stores processing metrics
func (m *s3Manager) StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error) {
	if verificationID == "" || metrics == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and processing metrics required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"metrics_nil":           metrics == nil,
			})
	}
	
	key := fmt.Sprintf("%s/processing-metrics.json", verificationID)
	stateRef, err := m.stateManager.StoreJSON("processing", key, metrics)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store processing metrics", true).
			WithContext("verification_id", verificationID)
	}
	
	return m.fromStateReference(stateRef), nil
}

// LoadProcessingState loads processing state by type
func (m *s3Manager) LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error) {
	if verificationID == "" || stateType == "" {
		return nil, errors.NewValidationError(
			"verification ID and state type required",
			map[string]interface{}{
				"verification_id": verificationID,
				"state_type":      stateType,
			})
	}
	
	key := fmt.Sprintf("%s/%s.json", verificationID, stateType)
	stateRef := &s3state.Reference{
		Bucket: m.bucket,
		Key:    fmt.Sprintf("processing/%s", key),
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

// Helper methods

// validateReference validates S3 reference
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