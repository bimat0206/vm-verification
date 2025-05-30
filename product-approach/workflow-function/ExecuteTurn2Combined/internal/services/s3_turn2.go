package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// LoadTurn1ProcessedResponse loads the processed Turn1 response from S3
func (m *s3Manager) LoadTurn1ProcessedResponse(ctx context.Context, ref models.S3Reference) (*schema.Turn1ProcessedResponse, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_turn1_processed_response_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "turn1_processed_response_load",
	})

	if err := m.validateReference(ref, "turn1_processed_response"); err != nil {
		return nil, err
	}

	// Get raw bytes
	stateRef := m.toStateReference(ref)
	raw, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_turn1_processed_response_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "get_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read Turn1 processed response: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return nil, wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Parse the Turn1 processed response
	var parsedData bedrockparser.ParsedTurn1Data
	if err := json.Unmarshal(raw, &parsedData); err != nil {
		m.logger.Error("s3_turn1_processed_response_format_error", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
			"bytes":  len(raw),
		})
		return nil, &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "BadTurn1ProcessedResponse",
			Message:   fmt.Sprintf("expected valid Turn1 processed response, got err %v", err),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	// Convert to schema.Turn1ProcessedResponse
	response := &schema.Turn1ProcessedResponse{
		InitialConfirmation: parsedData.InitialConfirmation,
		MachineStructure:    fmt.Sprintf("%v", parsedData.MachineStructure),
		ReferenceRowStatus:  fmt.Sprintf("%v", parsedData.ReferenceRowStatus),
	}

	duration := time.Since(startTime)
	m.logger.Info("turn1_processed_response_loaded_successfully", map[string]interface{}{
		"bucket":      ref.Bucket,
		"key":         ref.Key,
		"duration_ms": duration.Milliseconds(),
	})

	return response, nil
}

// LoadTurn1RawResponse loads the raw Turn1 response from S3
func (m *s3Manager) LoadTurn1RawResponse(ctx context.Context, ref models.S3Reference) (json.RawMessage, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_turn1_raw_response_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "turn1_raw_response_load",
	})

	if err := m.validateReference(ref, "turn1_raw_response"); err != nil {
		return nil, err
	}

	// Get raw bytes
	stateRef := m.toStateReference(ref)
	raw, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_turn1_raw_response_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "get_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read Turn1 raw response: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return nil, wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Validate it's valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal(raw, &jsonObj); err != nil {
		m.logger.Error("s3_turn1_raw_response_format_error", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
			"bytes":  len(raw),
		})
		return nil, &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "BadTurn1RawResponse",
			Message:   fmt.Sprintf("expected valid JSON, got err %v", err),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	duration := time.Since(startTime)
	m.logger.Info("turn1_raw_response_loaded_successfully", map[string]interface{}{
		"bucket":      ref.Bucket,
		"key":         ref.Key,
		"bytes":       len(raw),
		"duration_ms": duration.Milliseconds(),
	})

	return raw, nil
}

// StoreTurn2Response stores the Turn2 response
func (m *s3Manager) StoreTurn2Response(ctx context.Context, verificationID string, response *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 response",
			map[string]interface{}{"operation": "store_turn2_response"})
	}
	if response == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"response cannot be nil for storing Turn2 response",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "processing/turn2-processed-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, response)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2RawResponse stores the raw Turn2 Bedrock response
func (m *s3Manager) StoreTurn2RawResponse(ctx context.Context, verificationID string, raw interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 raw response",
			map[string]interface{}{"operation": "store_turn2_raw"})
	}

	key := "responses/turn2-raw-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, raw)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 raw response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2ProcessedResponse stores the processed Turn2 analysis
func (m *s3Manager) StoreTurn2ProcessedResponse(ctx context.Context, verificationID string, processed *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 processed response",
			map[string]interface{}{"operation": "store_turn2_processed"})
	}
	if processed == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"processed data cannot be nil",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "processing/turn2-processed-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, processed)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 processed response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2Markdown stores the Markdown version of the Turn2 analysis
func (m *s3Manager) StoreTurn2Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 markdown",
			map[string]interface{}{"operation": "store_turn2_markdown"})
	}

	key := "responses/turn2-processed-response.md"
	dataBytes := []byte(markdownContent)
	stateRef, err := m.stateManager.StoreWithContentType(m.datePath(verificationID), key, dataBytes, "text/markdown; charset=utf-8")
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 markdown", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2Conversation stores full conversation messages for turn2
func (m *s3Manager) StoreTurn2Conversation(ctx context.Context, verificationID string, messages []map[string]interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing turn2 conversation",
			map[string]interface{}{"operation": "store_turn2_conversation"})
	}

	key := "responses/turn2-conversation.json"
	data := map[string]interface{}{
		"verificationId": verificationID,
		"turnId":         2,
		"messages":       messages,
		"timestamp":      schema.FormatISO8601(),
	}
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, data)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store turn2 conversation", true).
			WithContext("verification_id", verificationID)
	}
	return m.fromStateReference(stateRef), nil
}
