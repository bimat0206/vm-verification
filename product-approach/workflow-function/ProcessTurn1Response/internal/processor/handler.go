package processor

import (
	"context"
	"fmt"
	"time"

	"workflow-function/ProcessTurn1Response/internal/storage"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Handler handles the ProcessTurn1Response Lambda function
type Handler struct {
	log        logger.Logger
	processor  *Processor
	deps       *storage.Dependencies
}

// NewHandler creates a new handler instance with all dependencies
func NewHandler(log logger.Logger) *Handler {
	// Initialize dependencies
	deps := GetStorageDependencies(log)
	
	// Create processor
	processor := NewProcessor(log, deps)

	return &Handler{
		log:       log,
		processor: processor,
		deps:      deps,
	}
}

// Handle processes the Turn 1 response and extracts reference analysis
func (h *Handler) Handle(ctx context.Context, input schema.WorkflowState) (schema.WorkflowState, error) {
	startTime := time.Now()
	
	// Extract verification ID for correlation
	verificationId := ""
	if input.VerificationContext != nil {
		verificationId = input.VerificationContext.VerificationId
	}
	
	// Create logger with correlation ID
	log := h.log.WithCorrelationId(verificationId)
	
	log.Info("ProcessTurn1Response function invoked", map[string]interface{}{
		"inputStatus":   getStatus(input.VerificationContext),
		"hasTurn1Response": input.Turn1Response != nil,
	})

	// Validate input has required Turn 1 response
	if err := h.validateInputBasics(input); err != nil {
		log.Error("Input validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return h.createErrorResponse(input, "INPUT_VALIDATION_ERROR", err.Error()), nil
	}

	// Process Turn 1 response
	result, err := h.processor.ProcessTurn1Response(ctx, input)
	if err != nil {
		log.Error("Turn 1 response processing failed", map[string]interface{}{
			"error": err.Error(),
		})
		return h.createErrorResponse(input, "PROCESSING_ERROR", err.Error()), nil
	}

	// Update status to TURN1_PROCESSED
	if result.VerificationContext != nil {
		result.VerificationContext.Status = schema.StatusTurn1Processed
	}

	// Log completion
	duration := time.Since(startTime)
	log.Info("ProcessTurn1Response completed successfully", map[string]interface{}{
		"duration":     duration.Milliseconds(),
		"outputStatus": getStatus(result.VerificationContext),
		"analysisType": getAnalysisType(result.ReferenceAnalysis),
	})

	return result, nil
}

// validateInputBasics performs basic validation on the input state
func (h *Handler) validateInputBasics(input schema.WorkflowState) error {
	// Check schema version compatibility
	if input.SchemaVersion != "" && input.SchemaVersion != schema.SchemaVersion {
		return fmt.Errorf("unsupported schema version: %s (supported: %s)", 
			input.SchemaVersion, schema.SchemaVersion)
	}

	// Ensure verification context exists
	if input.VerificationContext == nil {
		return fmt.Errorf("verification context is required")
	}

	// Check for Turn 1 response
	if input.Turn1Response == nil || len(input.Turn1Response) == 0 {
		return fmt.Errorf("turn1Response is required")
	}

	// Verify current status allows processing
	status := input.VerificationContext.Status
	if status != schema.StatusTurn1Completed {
		return fmt.Errorf("invalid status for processing: %s (expected: %s)", 
			status, schema.StatusTurn1Completed)
	}

	return nil
}

// createErrorResponse creates a standardized error response
func (h *Handler) createErrorResponse(input schema.WorkflowState, errorCode, errorMessage string) schema.WorkflowState {
	// Create error info
	errorInfo := &schema.ErrorInfo{
		Code:      errorCode,
		Message:   errorMessage,
		Timestamp: schema.FormatISO8601(),
		Details: map[string]interface{}{
			"function": "ProcessTurn1Response",
			"stage":    "Turn1Processing",
		},
	}

	// Create response with error
	response := input
	response.Error = errorInfo
	
	// Update verification context if it exists
	if response.VerificationContext != nil {
		response.VerificationContext.Status = schema.StatusVerificationFailed
		response.VerificationContext.Error = errorInfo
	}

	return response
}

// Helper functions

// getStatus safely gets status from verification context
func getStatus(ctx *schema.VerificationContext) string {
	if ctx == nil {
		return "UNKNOWN"
	}
	return ctx.Status
}

// getAnalysisType determines the type of analysis performed
func getAnalysisType(analysis map[string]interface{}) string {
	if analysis == nil {
		return "NONE"
	}
	
	if sourceType, ok := analysis["sourceType"].(string); ok {
		return sourceType
	}
	
	if status, ok := analysis["status"].(string); ok {
		return status
	}
	
	return "UNKNOWN"
}
