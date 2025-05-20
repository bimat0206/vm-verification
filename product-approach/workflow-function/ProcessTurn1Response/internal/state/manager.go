// Package state provides S3 state management for the ProcessTurn1Response Lambda
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/shared/s3state"
)

// Constants for S3 key categories
const (
	// Bucket environment variable
	EnvS3StateBucket = "S3_STATE_BUCKET"
)

// StateManager manages S3 state operations for the ProcessTurn1Response Lambda
type StateManager struct {
	s3Manager    s3state.Manager
	bucket       string
	logger       *slog.Logger
	correlationID string
}

// WorkflowState represents the state that flows through the Step Functions workflow
type WorkflowState struct {
	VerificationID string                 `json:"verificationId"`
	Status         string                 `json:"status"`
	Turn1Response  map[string]interface{} `json:"turn1Response,omitempty"`
	ReferenceAnalysis map[string]interface{} `json:"referenceAnalysis,omitempty"`
	ProcessingResult *types.Turn1ProcessingResult `json:"processingResult,omitempty"`
	References     map[string]*s3state.Reference `json:"references,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// New creates a new StateManager
func New(logger *slog.Logger, correlationID string) (*StateManager, error) {
	bucket := os.Getenv(EnvS3StateBucket)
	if bucket == "" {
		return nil, fmt.Errorf("environment variable %s is not set", EnvS3StateBucket)
	}

	s3Manager, err := s3state.New(bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 state manager: %w", err)
	}

	return &StateManager{
		s3Manager:    s3Manager,
		bucket:       bucket,
		logger:       logger,
		correlationID: correlationID,
	}, nil
}

// LoadWorkflowState loads the workflow state from the Lambda input
func (sm *StateManager) LoadWorkflowState(ctx context.Context, input interface{}) (*WorkflowState, error) {
	// First try to convert the input directly
	if state, ok := input.(*WorkflowState); ok {
		return state, nil
	}

	// Then try to convert from a map
	if data, ok := input.(map[string]interface{}); ok {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, s3state.NewJSONError(OpLoadWorkflowState, "failed to marshal input", err)
		}

		var state WorkflowState
		if err := json.Unmarshal(jsonBytes, &state); err != nil {
			return nil, s3state.NewJSONError(OpLoadWorkflowState, "failed to unmarshal input", err)
		}

		return &state, nil
	}

	// Then try to convert from an s3state.Envelope
	envelope, err := s3state.LoadEnvelope(input)
	if err != nil {
		return nil, s3state.WrapError(OpLoadWorkflowState, err)
	}

	state := &WorkflowState{
		VerificationID: envelope.VerificationID,
		Status:         envelope.Status,
		References:     envelope.References,
		Metadata:       envelope.Summary,
	}

	return state, nil
}

// LoadTurn1Response loads the Turn1 response from S3
func (sm *StateManager) LoadTurn1Response(ctx context.Context, state *WorkflowState) error {
	if state == nil {
		return s3state.NewValidationError(OpLoadTurn1Response, "workflow state is nil")
	}

	// Check if Turn1Response is already present in the state
	if state.Turn1Response != nil && len(state.Turn1Response) > 0 {
		sm.logger.Info("Turn1 response already present in workflow state, skipping load from S3")
		return nil
	}

	// Look for Turn1Response reference in the state
	responseRef := state.References[fmt.Sprintf("%s_%s", s3state.CategoryResponses, "turn1-response")]
	if responseRef == nil {
		return s3state.NewReferenceError(OpLoadTurn1Response, "turn1-response reference not found in workflow state")
	}

	// Validate the reference
	if err := s3state.ValidateReference(responseRef, OpLoadTurn1Response); err != nil {
		return err
	}

	// Retrieve the Turn1 response from S3
	var turn1Response map[string]interface{}
	err := sm.s3Manager.RetrieveJSON(responseRef, &turn1Response)
	if err != nil {
		return s3state.NewS3Error(
			OpLoadTurn1Response, 
			fmt.Sprintf("failed to retrieve Turn1 response from S3: %s", responseRef),
			err,
		)
	}

	// Store the Turn1 response in the workflow state
	state.Turn1Response = turn1Response
	sm.logger.Info("Successfully loaded Turn1 response from S3", 
		"reference", responseRef.String(),
		"size", responseRef.Size,
	)

	return nil
}

// StoreReferenceAnalysis stores the reference analysis result in S3
func (sm *StateManager) StoreReferenceAnalysis(ctx context.Context, state *WorkflowState, analysis map[string]interface{}) error {
	if state == nil {
		return s3state.NewValidationError(OpStoreReferenceAnalysis, "workflow state is nil")
	}

	if analysis == nil || len(analysis) == 0 {
		return s3state.NewValidationError(OpStoreReferenceAnalysis, "reference analysis is empty")
	}

	// Create or get envelope
	envelope := s3state.NewEnvelope(state.VerificationID)
	
	// Copy existing references if available
	if state.References != nil {
		for k, v := range state.References {
			envelope.AddReference(k, v)
		}
	}

	// Store reference analysis in S3
	filename, err := s3state.GetStandardFilename(s3state.CategoryProcessing, "turn1-analysis")
	if err != nil {
		return s3state.NewCategoryError(OpStoreReferenceAnalysis, fmt.Sprintf("failed to get standard filename: %v", err))
	}

	// Save the analysis to the envelope
	err = sm.s3Manager.SaveToEnvelope(envelope, s3state.CategoryProcessing, filename, analysis)
	if err != nil {
		return s3state.NewS3Error(OpStoreReferenceAnalysis, "failed to save reference analysis to S3", err)
	}

	// Update the workflow state with the reference analysis and reference
	state.ReferenceAnalysis = analysis
	state.References = envelope.References

	sm.logger.Info("Successfully stored reference analysis in S3",
		"verificationId", state.VerificationID,
		"category", s3state.CategoryProcessing,
		"filename", filename,
	)

	return nil
}

// UpdateWorkflowState updates the workflow state in S3
func (sm *StateManager) UpdateWorkflowState(ctx context.Context, state *WorkflowState, newStatus string) error {
	if state == nil {
		return s3state.NewValidationError(OpUpdateWorkflowState, "workflow state is nil")
	}

	// Update status if provided
	if newStatus != "" {
		state.Status = newStatus
	}

	// Create envelope from workflow state
	envelope := &s3state.Envelope{
		VerificationID: state.VerificationID,
		Status:         state.Status,
		References:     state.References,
		Summary:        state.Metadata,
	}

	// Validate the envelope
	if err := s3state.ValidateEnvelope(envelope, OpUpdateWorkflowState); err != nil {
		return err
	}

	// Store updated workflow state
	filename := fmt.Sprintf("workflow-state-%s.json", time.Now().Format("20060102-150405"))
	
	ref, err := sm.s3Manager.StoreJSON(s3state.CategoryProcessing, 
		fmt.Sprintf("%s/%s", state.VerificationID, filename), 
		state)
	if err != nil {
		return s3state.NewS3Error(OpUpdateWorkflowState, "failed to store updated workflow state", err)
	}

	sm.logger.Info("Successfully updated workflow state in S3",
		"verificationId", state.VerificationID,
		"status", state.Status,
		"reference", ref.String(),
	)

	return nil
}

// GetEnvelopeFromState converts a WorkflowState to an s3state.Envelope
func (sm *StateManager) GetEnvelopeFromState(state *WorkflowState) *s3state.Envelope {
	if state == nil {
		return nil
	}

	envelope := &s3state.Envelope{
		VerificationID: state.VerificationID,
		Status:         state.Status,
		References:     state.References,
		Summary:        state.Metadata,
	}

	return envelope
}

// LoadReferenceImage loads a reference image from S3
func (sm *StateManager) LoadReferenceImage(ctx context.Context, state *WorkflowState) (map[string]interface{}, error) {
	if state == nil {
		return nil, s3state.NewValidationError("LoadReferenceImage", "workflow state is nil")
	}

	// Look for reference image in the state
	imageRef := state.References[fmt.Sprintf("%s_%s", s3state.CategoryImages, "reference-base64")]
	if imageRef == nil {
		return nil, s3state.NewReferenceError("LoadReferenceImage", "reference image reference not found in workflow state")
	}

	// Validate the reference
	if err := s3state.ValidateReference(imageRef, "LoadReferenceImage"); err != nil {
		return nil, err
	}

	// Retrieve the reference image from S3
	var imageData map[string]interface{}
	err := sm.s3Manager.RetrieveJSON(imageRef, &imageData)
	if err != nil {
		return nil, s3state.NewS3Error(
			"LoadReferenceImage",
			fmt.Sprintf("failed to retrieve reference image from S3: %s", imageRef),
			err,
		)
	}

	sm.logger.Info("Successfully loaded reference image from S3",
		"reference", imageRef.String(),
		"size", imageRef.Size,
	)

	return imageData, nil
}

// StoreProcessingResult stores the processing result in S3
func (sm *StateManager) StoreProcessingResult(ctx context.Context, state *WorkflowState, result *types.Turn1ProcessingResult) error {
	if state == nil {
		return s3state.NewValidationError("StoreProcessingResult", "workflow state is nil")
	}

	if result == nil {
		return s3state.NewValidationError("StoreProcessingResult", "processing result is nil")
	}

	// Create or get envelope
	envelope := sm.GetEnvelopeFromState(state)
	if envelope == nil {
		envelope = s3state.NewEnvelope(state.VerificationID)
	}

	// Store processing result in S3
	filename := fmt.Sprintf("turn1-processing-result-%s.json", time.Now().Format("20060102-150405"))

	// Save the result to the envelope
	err := sm.s3Manager.SaveToEnvelope(envelope, s3state.CategoryProcessing, filename, result)
	if err != nil {
		return s3state.NewS3Error("StoreProcessingResult", "failed to save processing result to S3", err)
	}

	// Update the workflow state with the processing result and references
	state.ProcessingResult = result
	state.References = envelope.References

	sm.logger.Info("Successfully stored processing result in S3",
		"verificationId", state.VerificationID,
		"status", result.Status,
		"sourceType", result.SourceType,
	)

	return nil
}

// LoadHistoricalContext loads historical context data from S3 if available
func (sm *StateManager) LoadHistoricalContext(ctx context.Context, state *WorkflowState) (map[string]interface{}, error) {
	if state == nil {
		return nil, s3state.NewValidationError("LoadHistoricalContext", "workflow state is nil")
	}

	// Look for historical context reference in the state
	histRef := state.References[fmt.Sprintf("%s_%s", s3state.CategoryProcessing, "historical-context")]
	if histRef == nil {
		// Historical context is optional, so we return nil without error
		sm.logger.Info("No historical context reference found, skipping load")
		return nil, nil
	}

	// Validate the reference
	if err := s3state.ValidateReference(histRef, "LoadHistoricalContext"); err != nil {
		return nil, err
	}

	// Retrieve the historical context from S3
	var histData map[string]interface{}
	err := sm.s3Manager.RetrieveJSON(histRef, &histData)
	if err != nil {
		return nil, s3state.NewS3Error(
			"LoadHistoricalContext",
			fmt.Sprintf("failed to retrieve historical context from S3: %s", histRef),
			err,
		)
	}

	sm.logger.Info("Successfully loaded historical context from S3",
		"reference", histRef.String(),
		"size", histRef.Size,
	)

	return histData, nil
}

// StoreContextForTurn2 stores the context for Turn 2 in S3
func (sm *StateManager) StoreContextForTurn2(ctx context.Context, state *WorkflowState, turn2Context map[string]interface{}) error {
	if state == nil {
		return s3state.NewValidationError("StoreContextForTurn2", "workflow state is nil")
	}

	if turn2Context == nil || len(turn2Context) == 0 {
		return s3state.NewValidationError("StoreContextForTurn2", "Turn 2 context is empty")
	}

	// Create or get envelope
	envelope := sm.GetEnvelopeFromState(state)
	if envelope == nil {
		envelope = s3state.NewEnvelope(state.VerificationID)
	}

	// Store Turn 2 context in S3
	filename := "turn2-context.json"

	// Save the context to the envelope
	err := sm.s3Manager.SaveToEnvelope(envelope, s3state.CategoryProcessing, filename, turn2Context)
	if err != nil {
		return s3state.NewS3Error("StoreContextForTurn2", "failed to save Turn 2 context to S3", err)
	}

	// Update the workflow state references
	state.References = envelope.References

	sm.logger.Info("Successfully stored context for Turn 2 in S3",
		"verificationId", state.VerificationID,
		"category", s3state.CategoryProcessing,
		"filename", filename,
	)

	return nil
}