// internal/services/s3_enhanced.go - ENHANCED S3 OPERATIONS
package services

import (
	"context"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// S3EnhancedOperations provides additional S3 operations for comprehensive state management
type S3EnhancedOperations struct {
	stateManager s3state.Manager
	bucket       string
}

// NewS3EnhancedOperations creates enhanced S3 operations
func NewS3EnhancedOperations(bucket string) (*S3EnhancedOperations, error) {
	mgr, err := s3state.New(bucket)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to create enhanced S3 operations", false).
			WithContext("bucket", bucket)
	}

	return &S3EnhancedOperations{
		stateManager: mgr,
		bucket:       bucket,
	}, nil
}

// StoreWorkflowState stores complete workflow state
func (e *S3EnhancedOperations) StoreWorkflowState(ctx context.Context, verificationID string, state *schema.WorkflowState) (models.S3Reference, error) {
	if verificationID == "" || state == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and workflow state required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"state_nil":             state == nil,
			})
	}

	// Validate workflow state
	if validationErrors := schema.ValidateWorkflowState(state); len(validationErrors) > 0 {
		return models.S3Reference{}, errors.NewValidationError(
			"workflow state validation failed",
			map[string]interface{}{
				"validation_errors": validationErrors.Error(),
			})
	}

	key := fmt.Sprintf("%s/workflow-state.json", verificationID)
	stateRef, err := e.stateManager.StoreJSON("processing", key, state)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store workflow state", true).
			WithContext("verification_id", verificationID)
	}

	return e.fromStateReference(stateRef), nil
}

// LoadWorkflowState loads complete workflow state
func (e *S3EnhancedOperations) LoadWorkflowState(ctx context.Context, verificationID string) (*schema.WorkflowState, error) {
	if verificationID == "" {
		return nil, errors.NewValidationError(
			"verification ID required",
			map[string]interface{}{"operation": "load_workflow_state"})
	}

	key := fmt.Sprintf("processing/%s/workflow-state.json", verificationID)
	stateRef := &s3state.Reference{
		Bucket: e.bucket,
		Key:    key,
	}

	var state schema.WorkflowState
	if err := e.stateManager.RetrieveJSON(stateRef, &state); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load workflow state", true).
			WithContext("verification_id", verificationID)
	}

	return &state, nil
}

// StoreStatusHistory stores status history for tracking
func (e *S3EnhancedOperations) StoreStatusHistory(ctx context.Context, verificationID string, history []schema.StatusHistoryEntry) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing status history",
			map[string]interface{}{"history_count": len(history)})
	}

	key := fmt.Sprintf("%s/status-history.json", verificationID)
	stateRef, err := e.stateManager.StoreJSON("processing", key, history)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store status history", true).
			WithContext("verification_id", verificationID).
			WithContext("history_count", len(history))
	}

	return e.fromStateReference(stateRef), nil
}

// AppendStatusHistory appends a new status entry to existing history
func (e *S3EnhancedOperations) AppendStatusHistory(ctx context.Context, verificationID string, entry schema.StatusHistoryEntry) (models.S3Reference, error) {
	// Load existing history
	history, err := e.LoadStatusHistory(ctx, verificationID)
	if err != nil {
		// If not found, start with empty history
		history = []schema.StatusHistoryEntry{}
	}

	// Append new entry
	history = append(history, entry)

	// Store updated history
	return e.StoreStatusHistory(ctx, verificationID, history)
}

// LoadStatusHistory loads status history
func (e *S3EnhancedOperations) LoadStatusHistory(ctx context.Context, verificationID string) ([]schema.StatusHistoryEntry, error) {
	if verificationID == "" {
		return nil, errors.NewValidationError(
			"verification ID required",
			map[string]interface{}{"operation": "load_status_history"})
	}

	key := fmt.Sprintf("processing/%s/status-history.json", verificationID)
	stateRef := &s3state.Reference{
		Bucket: e.bucket,
		Key:    key,
	}

	var history []schema.StatusHistoryEntry
	if err := e.stateManager.RetrieveJSON(stateRef, &history); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load status history", true).
			WithContext("verification_id", verificationID)
	}

	return history, nil
}

// StoreConversationHistory stores complete conversation history
func (e *S3EnhancedOperations) StoreConversationHistory(ctx context.Context, verificationID string, conversation *schema.ConversationTracker) (models.S3Reference, error) {
	if verificationID == "" || conversation == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and conversation required",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"conversation_nil":      conversation == nil,
			})
	}

	// Validate conversation tracker
	if validationErrors := schema.ValidateConversationTracker(conversation); len(validationErrors) > 0 {
		return models.S3Reference{}, errors.NewValidationError(
			"conversation tracker validation failed",
			map[string]interface{}{
				"validation_errors": validationErrors.Error(),
			})
	}

	key := fmt.Sprintf("%s/conversation-history.json", verificationID)
	stateRef, err := e.stateManager.StoreJSON("responses", key, conversation)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store conversation history", true).
			WithContext("verification_id", verificationID).
			WithContext("current_turn", conversation.CurrentTurn)
	}

	return e.fromStateReference(stateRef), nil
}

// CreateStateSnapshot creates a complete state snapshot
func (e *S3EnhancedOperations) CreateStateSnapshot(ctx context.Context, verificationID string, state interface{}) (models.S3Reference, error) {
	if verificationID == "" || state == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID and state required for snapshot",
			map[string]interface{}{
				"verification_id_empty": verificationID == "",
				"state_nil":             state == nil,
			})
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	key := fmt.Sprintf("%s/snapshots/state-%s.json", verificationID, timestamp)

	stateRef, err := e.stateManager.StoreJSON("processing", key, state)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to create state snapshot", true).
			WithContext("verification_id", verificationID).
			WithContext("timestamp", timestamp)
	}

	return e.fromStateReference(stateRef), nil
}

// CleanupExpiredStates removes expired temporary states
func (e *S3EnhancedOperations) CleanupExpiredStates(ctx context.Context, verificationID string, olderThan time.Duration) error {
	if verificationID == "" {
		return errors.NewValidationError(
			"verification ID required for cleanup",
			map[string]interface{}{"older_than_hours": olderThan.Hours()})
	}

	// In a full implementation, this would:
	// 1. List objects with the verification ID prefix
	// 2. Check timestamps
	// 3. Delete expired objects
	// For now, this is a placeholder

	return nil
}

// Helper methods

// fromStateReference converts state reference to model reference
func (e *S3EnhancedOperations) fromStateReference(ref *s3state.Reference) models.S3Reference {
	return models.S3Reference{
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   ref.Size,
	}
}

// LoadInitializationData loads initialization data from S3
// This method provides interface consistency with S3StateManager
func (e *S3EnhancedOperations) LoadInitializationData(ctx context.Context, ref models.S3Reference) (*InitializationData, error) {
	if ref.Bucket == "" || ref.Key == "" {
		return nil, errors.NewValidationError(
			"initialization data reference validation failed",
			map[string]interface{}{
				"bucket_empty": ref.Bucket == "",
				"key_empty":    ref.Key == "",
			})
	}

	var initData InitializationData
	stateRef := &s3state.Reference{
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   ref.Size,
	}

	if err := e.stateManager.RetrieveJSON(stateRef, &initData); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load initialization data", true).
			WithContext("s3_key", ref.Key).
			WithContext("bucket", ref.Bucket)
	}

	return &initData, nil
}

// LoadImageMetadata loads image metadata from S3
// This method provides interface consistency with S3StateManager
func (e *S3EnhancedOperations) LoadImageMetadata(ctx context.Context, ref models.S3Reference) (*ImageMetadata, error) {
	if ref.Bucket == "" || ref.Key == "" {
		return nil, errors.NewValidationError(
			"image metadata reference validation failed",
			map[string]interface{}{
				"bucket_empty": ref.Bucket == "",
				"key_empty":    ref.Key == "",
			})
	}

	var metadata ImageMetadata
	stateRef := &s3state.Reference{
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   ref.Size,
	}

	if err := e.stateManager.RetrieveJSON(stateRef, &metadata); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to load image metadata", true).
			WithContext("s3_key", ref.Key).
			WithContext("bucket", ref.Bucket)
	}

	return &metadata, nil
}
