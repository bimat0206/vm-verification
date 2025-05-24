package handler

import (
	"context"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// DynamoManager handles DynamoDB operations
type DynamoManager struct {
	dynamo services.DynamoDBService
	cfg    config.Config
	log    logger.Logger
}

// NewDynamoManager creates a new instance of DynamoManager
func NewDynamoManager(dynamo services.DynamoDBService, cfg config.Config, log logger.Logger) *DynamoManager {
	return &DynamoManager{
		dynamo: dynamo,
		cfg:    cfg,
		log:    log,
	}
}

// UpdateAsync performs asynchronous DynamoDB updates
func (d *DynamoManager) UpdateAsync(
	ctx context.Context,
	verificationID string,
	tokenUsage models.TokenUsage,
	requestID string,
	rawRef, procRef models.S3Reference,
) (<-chan error, *bool) {
	updateComplete := make(chan error, 2)
	dynamoOK := new(bool)
	*dynamoOK = true // Initialize to true
	contextLogger := d.log.WithCorrelationId(verificationID)

	go func() {
		// Build status entry using schema types
		statusEntry := schema.StatusHistoryEntry{
			Status:           schema.StatusTurn1Completed,
			Timestamp:        schema.FormatISO8601(),
			FunctionName:     "ExecuteTurn1Combined",
			ProcessingTimeMs: 0, // Will be set by caller
			Stage:            "turn1_completion",
		}

		// Update verification status with enhanced method
		updateErr := d.dynamo.UpdateVerificationStatusEnhanced(ctx, verificationID, statusEntry)
		if updateErr != nil {
			asyncLogger := contextLogger.WithFields(map[string]interface{}{
				"async_operation": "verification_status_update",
				"table":           d.cfg.AWS.DynamoDBVerificationTable,
			})

			if workflowErr, ok := updateErr.(*errors.WorkflowError); ok {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error_type": string(workflowErr.Type),
					"error_code": workflowErr.Code,
					"retryable":  workflowErr.Retryable,
				})
			} else {
				asyncLogger.Warn("dynamodb status update failed", map[string]interface{}{
					"error": updateErr.Error(),
				})
			}
			
			// Set dynamoOK flag to false on error
			*dynamoOK = false
		}
		updateComplete <- updateErr

		// Build turn response for conversation history
		turnResponse := &schema.TurnResponse{
			TurnId:     1,
			Timestamp:  schema.FormatISO8601(),
			Prompt:     "", // Would be filled with actual prompt
			ImageUrls:  map[string]string{},
			Response:   schema.BedrockApiResponse{RequestId: requestID},
			LatencyMs:  0, // Will be set by caller
			TokenUsage: &tokenUsage,
			Stage:      "REFERENCE_ANALYSIS",
			Metadata: map[string]interface{}{
				"model_id":        d.cfg.AWS.BedrockModel,
				"verification_id": verificationID,
				"function_name":   "ExecuteTurn1Combined",
			},
		}

		// Update conversation turn with the method available in the interface
		historyErr := d.dynamo.UpdateConversationTurn(ctx, verificationID, turnResponse)
		if historyErr != nil {
			contextLogger.Warn("conversation history recording failed", map[string]interface{}{
				"error": historyErr.Error(),
				"table": d.cfg.AWS.DynamoDBConversationTable,
			})
			
			// Set dynamoOK flag to false on error
			*dynamoOK = false
		}
		updateComplete <- historyErr
	}()

	return updateComplete, dynamoOK
}

// WaitForUpdates waits for async updates with timeout
func (d *DynamoManager) WaitForUpdates(updateComplete <-chan error, timeout time.Duration, contextLogger logger.Logger) {
	updateTimeout := time.After(timeout)
	updatesReceived := 0

	for updatesReceived < 2 {
		select {
		case err := <-updateComplete:
			updatesReceived++
			if err != nil {
				contextLogger.Warn("async update error received", map[string]interface{}{
					"update_number": updatesReceived,
					"error":         err.Error(),
				})
			}
		case <-updateTimeout:
			contextLogger.Warn("async updates timed out", map[string]interface{}{
				"updates_received": updatesReceived,
				"timeout_seconds":  timeout.Seconds(),
			})
			return
		}
	}
}
