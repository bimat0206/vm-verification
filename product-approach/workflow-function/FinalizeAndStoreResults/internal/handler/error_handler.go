package handler

import (
	"context"
	"fmt"
	"time"

	"workflow-function/FinalizeWithErrorFunction/internal/models"
	"workflow-function/FinalizeWithErrorFunction/internal/services"
	"workflow-function/FinalizeWithErrorFunction/internal/utils"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// ErrorHandler processes workflow errors
type ErrorHandler struct {
	logger         logger.Logger
	s3Manager      s3state.Manager
	contextService *services.ContextService
}

// NewErrorHandler constructs ErrorHandler
func NewErrorHandler(log logger.Logger, mgr s3state.Manager) *ErrorHandler {
	return &ErrorHandler{
		logger:         log,
		s3Manager:      mgr,
		contextService: services.NewContextService(log, mgr),
	}
}

// ProcessError handles the FinalizeWithErrorInput and returns FinalizeWithErrorOutput
func (h *ErrorHandler) ProcessError(ctx context.Context, input *models.FinalizeWithErrorInput) (*models.FinalizeWithErrorOutput, error) {
	h.logger.Info("Processing error", map[string]interface{}{
		"verificationId": input.VerificationID,
		"errorStage":     input.ErrorStage,
	})

	// Parse Step Functions error cause
	if input.Error.Cause != nil {
		parsed, err := utils.ParseStepFunctionsErrorCause(*input.Error.Cause)
		if err != nil {
			h.logger.Warn("Failed to parse error cause", map[string]interface{}{"error": err.Error()})
		} else {
			input.Error.ParsedCause = parsed
		}
	}

	// Construct schema.ErrorInfo
	now := time.Now().UTC().Format(time.RFC3339)
	errInfo := schema.ErrorInfo{
		Code:    "WORKFLOW_ERROR",
		Message: fmt.Sprintf("Stage %s failed", input.ErrorStage),
		Details: map[string]interface{}{
			"stepError": input.Error,
		},
		Timestamp: now,
	}

	// Load existing context
	var initRef *schema.S3Reference
	if ref, ok := input.PartialS3References["processing_initialization"]; ok {
		initRef = &ref
	}

	vc, _ := h.contextService.LoadOrCreateVerificationContext(ctx, input.VerificationID, initRef, schema.StatusVerificationFailed, &errInfo)

	// Update context with error
	h.contextService.UpdateVerificationContextWithError(vc, input.ErrorStage, errInfo, utils.DetermineSpecificErrorStatus)

	// Save updated context back to S3
	if initRef != nil {
		if _, err := h.s3Manager.StoreJSON("", initRef.Key, vc); err != nil {
			h.logger.Error("Failed to store updated verification context", map[string]interface{}{"error": err.Error()})
		}
	}

	// Optionally store error summary
	outputRefs := make(map[string]schema.S3Reference)
	for k, v := range input.PartialS3References {
		outputRefs[k] = v
	}

	summary := map[string]interface{}{
		"message":    "Verification failed.",
		"errorStage": input.ErrorStage,
		"errorCode":  errInfo.Code,
	}

	output := &models.FinalizeWithErrorOutput{
		SchemaVersion:       schema.SchemaVersion,
		VerificationID:      input.VerificationID,
		S3References:        outputRefs,
		Status:              vc.Status,
		Error:               errInfo,
		VerificationContext: vc,
		Summary:             summary,
	}

	h.logger.Info("Finalized error processing", map[string]interface{}{
		"verificationId": input.VerificationID,
		"status":         vc.Status,
	})

	return output, nil
}
