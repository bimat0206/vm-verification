// Package handler provides the entry point for the ProcessTurn1Response Lambda function
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"workflow-function/ProcessTurn1Response/internal/processor"
	"workflow-function/ProcessTurn1Response/internal/state"
	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/shared/schema"
)

// ErrorCategory represents standardized error categories for reporting
type ErrorCategory string

const (
	// Error categories
	ErrorCategoryInput       ErrorCategory = "INPUT_VALIDATION_ERROR"
	ErrorCategoryState       ErrorCategory = "STATE_MANAGEMENT_ERROR"
	ErrorCategoryProcessing  ErrorCategory = "PROCESSING_ERROR"
	ErrorCategoryStorage     ErrorCategory = "STORAGE_ERROR"
	ErrorCategoryUnexpected  ErrorCategory = "UNEXPECTED_ERROR"
)

// Handler is responsible for coordinating the workflow of the ProcessTurn1Response Lambda
type Handler struct {
	logger       *slog.Logger
	processor    processor.Processor
	stateManager *state.StateManager
}

// New creates a new Handler instance with all dependencies
func New(logger *slog.Logger) (*Handler, error) {
	// Create a new state manager
	correlationID := "" // This will be updated when we receive the input
	stateManager, err := state.New(logger, correlationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Create processor
	processor := processor.New(logger)

	return &Handler{
		logger:       logger,
		processor:    processor,
		stateManager: stateManager,
	}, nil
}

// Handle is the entry point for the Lambda function
func (h *Handler) Handle(ctx context.Context, input interface{}) (interface{}, error) {
	startTime := time.Now()

	// Step 1: Load and validate workflow state
	workflowState, err := h.stateManager.LoadWorkflowState(ctx, input)
	if err != nil {
		return h.handleError(ctx, nil, ErrorCategoryInput, "Failed to load workflow state", err)
	}

	// Update correlation ID with verification ID once we have it
	verificationID := workflowState.VerificationID
	h.logger = h.logger.With("verificationId", verificationID)
	h.stateManager = &state.StateManager{} // Replace with a proper update method when available

	h.logger.Info("ProcessTurn1Response function invoked", 
		"status", workflowState.Status,
		"hasTurn1Response", workflowState.Turn1Response != nil,
	)

	// Step 2: Prepare state for processing
	err = h.stateManager.PrepareStateForProcessing(ctx, workflowState)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryState, "Failed to prepare state for processing", err)
	}

	// Step 3: Determine the processing path
	processingPath, err := h.stateManager.DetermineProcessingPath(ctx, workflowState)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryInput, "Failed to determine processing path", err)
	}

	h.logger.Info("Processing path determined", "path", processingPath)

	// Step 4: Extract Turn1 response content
	responseContent, err := h.stateManager.ExtractResponseContentFromState(workflowState)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryInput, "Failed to extract response content", err)
	}

	// Step 5: Load historical data if relevant
	var historicalData *types.HistoricalEnhancement
	if processingPath == types.PathHistoricalEnhancement {
		historicalData, err = h.stateManager.ExtractHistoricalData(ctx, workflowState)
		if err != nil {
			h.logger.Warn("Failed to extract historical data, continuing without it", "error", err.Error())
			// Continue processing without historical data
		}
	}

	// Step 6: Process the Turn1 response based on the determined path
	processingResult, err := h.processor.ProcessTurn1Response(ctx, responseContent, processingPath, historicalData)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryProcessing, "Failed to process Turn1 response", err)
	}

	// Step 7: Store processing result in S3
	err = h.stateManager.StoreProcessingResult(ctx, workflowState, processingResult)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryStorage, "Failed to store processing result", err)
	}

	// Step 8: Build context for Turn2
	turn2Context, err := h.stateManager.BuildTurn2Context(ctx, workflowState, processingResult)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryState, "Failed to build Turn2 context", err)
	}

	// Step 9: Store context for Turn2 in S3
	err = h.stateManager.StoreContextForTurn2(ctx, workflowState, turn2Context)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryStorage, "Failed to store Turn2 context", err)
	}

	// Step 10: Store reference analysis in S3
	err = h.stateManager.StoreReferenceAnalysis(ctx, workflowState, processingResult.ReferenceAnalysis)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryStorage, "Failed to store reference analysis", err)
	}

	// Step 11: Update workflow state status and save final state
	workflowState.Status = schema.StatusTurn1Processed
	err = h.stateManager.UpdateWorkflowState(ctx, workflowState, schema.StatusTurn1Processed)
	if err != nil {
		return h.handleError(ctx, workflowState, ErrorCategoryState, "Failed to update workflow state", err)
	}

	// Log completion
	duration := time.Since(startTime)
	h.logger.Info("ProcessTurn1Response completed successfully",
		"duration", duration.Milliseconds(),
		"processingPath", processingPath,
		"status", workflowState.Status,
	)

	// Convert to envelope for return
	return h.stateManager.GetEnvelopeFromState(workflowState), nil
}

// This method is now implemented in error_handler.go
// The duplicate has been removed to avoid compilation errors