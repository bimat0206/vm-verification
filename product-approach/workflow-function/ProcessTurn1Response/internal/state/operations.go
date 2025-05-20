// Package state provides S3 state management for the ProcessTurn1Response Lambda
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// Constants for additional operation names
const (
	OpGetVerificationType   = "GetVerificationType"
)

// ExtractHistoricalData extracts historical data from the workflow state
func (sm *StateManager) ExtractHistoricalData(ctx context.Context, state *WorkflowState) (*types.HistoricalEnhancement, error) {
	if state == nil {
		return nil, s3state.NewValidationError(OpExtractHistoricalData, "workflow state is nil")
	}

	// Check if historical context is available
	historicalData, err := sm.LoadHistoricalContext(ctx, state)
	if err != nil {
		// Log the error but continue without historical data
		sm.logger.Warn("Failed to load historical context", "error", err)
		return nil, nil
	}

	// If no historical data, return nil without error (it's optional)
	if historicalData == nil || len(historicalData) == 0 {
		sm.logger.Info("No historical data available for enhancement")
		return nil, nil
	}

	// Build the historical enhancement structure
	enhancement := &types.HistoricalEnhancement{
		BaselineData: historicalData,
		VisualConfirmation: make(map[string]interface{}),
		NewObservations: []string{},
		Discrepancies: []string{},
		EnrichedBaseline: make(map[string]interface{}),
	}

	sm.logger.Info("Successfully extracted historical data for enhancement", 
		"dataSize", len(historicalData),
	)

	return enhancement, nil
}

// BuildTurn2Context constructs the context needed for Turn 2 processing
func (sm *StateManager) BuildTurn2Context(ctx context.Context, state *WorkflowState, result *types.Turn1ProcessingResult) (map[string]interface{}, error) {
	if state == nil {
		return nil, s3state.NewValidationError(OpBuildTurn2Context, "workflow state is nil")
	}

	if result == nil {
		return nil, s3state.NewValidationError(OpBuildTurn2Context, "processing result is nil")
	}

	// Initialize Turn 2 context
	turn2Context := make(map[string]interface{})

	// Add verification metadata
	turn2Context["verificationId"] = state.VerificationID
	turn2Context["processingPath"] = result.SourceType
	turn2Context["processingTimestamp"] = time.Now().Format(time.RFC3339)

	// Add reference analysis
	if result.ReferenceAnalysis != nil && len(result.ReferenceAnalysis) > 0 {
		turn2Context["referenceAnalysis"] = result.ReferenceAnalysis
	}

	// Add extracted structure if available
	if result.ExtractedStructure != nil {
		turn2Context["machineStructure"] = result.ExtractedStructure
	}

	// Add contextual data specific to the processing path
	switch result.SourceType {
	case types.PathValidationFlow:
		// For validation flow, include expected structure for validation
		if state.Metadata != nil {
			if layoutInfo, ok := state.Metadata["layoutInfo"].(map[string]interface{}); ok {
				turn2Context["expectedLayout"] = layoutInfo
			}
		}

	case types.PathHistoricalEnhancement:
		// For historical enhancement, include the enriched baseline
		if result.ContextForTurn2 != nil {
			if enrichedData, ok := result.ContextForTurn2["enrichedBaseline"].(map[string]interface{}); ok {
				turn2Context["historicalContext"] = enrichedData
			}
		}

	case types.PathFreshExtraction:
		// For fresh extraction, include the complete extracted state
		if result.ContextForTurn2 != nil {
			turn2Context["extractedState"] = result.ContextForTurn2
		}
	}

	// Add preprocessing metadata
	metadata := map[string]interface{}{
		"processingTime": result.ProcessingMetadata.ProcessingDuration.String(),
		"sourceType": result.SourceType,
		"extractedElements": result.ProcessingMetadata.ExtractedElements,
		"validationsPassed": result.ProcessingMetadata.ValidationsPassed,
		"validationsFailed": result.ProcessingMetadata.ValidationsFailed,
	}

	// Add warnings if present
	if result.HasWarnings() {
		metadata["warnings"] = result.Warnings
	}

	turn2Context["processingMetadata"] = metadata

	sm.logger.Info("Successfully built context for Turn 2",
		"contextSize", len(turn2Context),
		"processingPath", result.SourceType,
	)

	return turn2Context, nil
}

// GetVerificationType extracts the verification type from the workflow state
func (sm *StateManager) GetVerificationType(state *WorkflowState) (string, error) {
	if state == nil {
		return "", s3state.NewValidationError(OpGetVerificationType, "workflow state is nil")
	}

	// Try to get from metadata
	if state.Metadata != nil {
		if verificationType, ok := state.Metadata["verificationType"].(string); ok && verificationType != "" {
			return verificationType, nil
		}
	}

	// Try to get from turn1Response
	if state.Turn1Response != nil {
		if metadata, ok := state.Turn1Response["metadata"].(map[string]interface{}); ok {
			if verificationType, ok := metadata["verificationType"].(string); ok && verificationType != "" {
				return verificationType, nil
			}
		}
	}

	// Default to PREVIOUS_VS_CURRENT if not found (most common case)
	sm.logger.Warn("Verification type not found in workflow state, defaulting to PREVIOUS_VS_CURRENT")
	return schema.VerificationTypePreviousVsCurrent, nil
}

// DetermineProcessingPath determines the appropriate processing path based on verification type and historical data
func (sm *StateManager) DetermineProcessingPath(ctx context.Context, state *WorkflowState) (types.ProcessingPath, error) {
	if state == nil {
		return "", s3state.NewValidationError(OpDetermineProcessingPath, "workflow state is nil")
	}

	// Get verification type
	verificationType, err := sm.GetVerificationType(state)
	if err != nil {
		return "", s3state.WrapError(OpDetermineProcessingPath, err)
	}

	// Check for historical context
	historicalRef := state.References[fmt.Sprintf("%s_%s", s3state.CategoryProcessing, "historical-context")]
	hasHistorical := historicalRef != nil && historicalRef.IsValid()

	// Determine processing path based on verification type and historical context
	var path types.ProcessingPath
	switch verificationType {
	case schema.VerificationTypeLayoutVsChecking:
		// For layout vs checking, always use validation flow
		path = types.PathValidationFlow
		sm.logger.Info("Using validation flow for LAYOUT_VS_CHECKING verification type")

	case schema.VerificationTypePreviousVsCurrent:
		if hasHistorical {
			// Use historical enhancement when historical data is available
			path = types.PathHistoricalEnhancement
			sm.logger.Info("Using historical enhancement with available historical data")
		} else {
			// Use fresh extraction when no historical data is available
			path = types.PathFreshExtraction
			sm.logger.Info("Using fresh extraction as no historical data is available")
		}

	default:
		// Default to fresh extraction for unknown verification types
		path = types.PathFreshExtraction
		sm.logger.Warn("Unknown verification type, defaulting to fresh extraction", "verificationType", verificationType)
	}

	// Validate the selected path
	if !path.IsValid() {
		return "", s3state.NewValidationError(OpDetermineProcessingPath, fmt.Sprintf("invalid processing path: %s", path))
	}

	return path, nil
}

// ValidateWorkflowState performs comprehensive validation of the workflow state
func (sm *StateManager) ValidateWorkflowState(ctx context.Context, state *WorkflowState) error {
	if state == nil {
		return s3state.NewValidationError("ValidateWorkflowState", "workflow state is nil")
	}

	errorList := s3state.NewErrorList()

	// Check required fields
	if state.VerificationID == "" {
		errorList.Add(s3state.NewValidationError("ValidateWorkflowState", "verification ID is required"))
	}

	if state.Status == "" {
		errorList.Add(s3state.NewValidationError("ValidateWorkflowState", "status is required"))
	}

	// Check Turn1Response availability
	if state.Turn1Response == nil || len(state.Turn1Response) == 0 {
		// Try to load it from S3 if not present
		err := sm.LoadTurn1Response(ctx, state)
		if err != nil {
			errorList.Add(s3state.WrapError("ValidateWorkflowState", fmt.Errorf("turn1Response is required: %w", err)))
		}
	}

	// Validate references
	if state.References != nil {
		// Ensure at least the turn1-response reference exists
		responseRef := state.References[fmt.Sprintf("%s_%s", s3state.CategoryResponses, "turn1-response")]
		if responseRef == nil {
			errorList.Add(s3state.NewReferenceError("ValidateWorkflowState", "turn1-response reference is required"))
		} else if !responseRef.IsValid() {
			errorList.Add(s3state.NewReferenceError("ValidateWorkflowState", "turn1-response reference is invalid"))
		}
	} else {
		errorList.Add(s3state.NewValidationError("ValidateWorkflowState", "references map is required"))
	}

	return errorList.ToError()
}

// PrepareStateForProcessing prepares the workflow state for processing
func (sm *StateManager) PrepareStateForProcessing(ctx context.Context, state *WorkflowState) error {
	// Validate the workflow state
	if err := sm.ValidateWorkflowState(ctx, state); err != nil {
		return err
	}

	// Ensure Turn1Response is loaded
	if state.Turn1Response == nil || len(state.Turn1Response) == 0 {
		if err := sm.LoadTurn1Response(ctx, state); err != nil {
			return err
		}
	}

	// Determine and set the processing path
	path, err := sm.DetermineProcessingPath(ctx, state)
	if err != nil {
		return err
	}

	// Update state status to indicate processing has begun
	state.Status = "PROCESSING_TURN1"

	// Add processing metadata to state
	if state.Metadata == nil {
		state.Metadata = make(map[string]interface{})
	}
	state.Metadata["processingStartTime"] = time.Now().Format(time.RFC3339)
	state.Metadata["processingPath"] = path.String()

	// Update state in S3
	return sm.UpdateWorkflowState(ctx, state, "")
}

// ExtractResponseContentFromState extracts the actual response content from the Turn1Response in state
func (sm *StateManager) ExtractResponseContentFromState(state *WorkflowState) (string, error) {
	if state == nil || state.Turn1Response == nil {
		return "", s3state.NewValidationError("ExtractResponseContent", "workflow state or turn1Response is nil")
	}

	// Try to get the content from different possible locations in the JSON structure
	if content, ok := state.Turn1Response["content"].(string); ok && content != "" {
		return content, nil
	}

	if response, ok := state.Turn1Response["response"].(map[string]interface{}); ok {
		if content, ok := response["content"].(string); ok && content != "" {
			return content, nil
		}
	}

	// Try to get from completion
	if completion, ok := state.Turn1Response["completion"].(map[string]interface{}); ok {
		if content, ok := completion["content"].(string); ok && content != "" {
			return content, nil
		}
	}

	// If we can't find it directly, try to marshal the entire structure to string
	jsonBytes, err := json.Marshal(state.Turn1Response)
	if err != nil {
		return "", s3state.NewJSONError("ExtractResponseContent", "failed to marshal turn1Response", err)
	}

	return string(jsonBytes), nil
}