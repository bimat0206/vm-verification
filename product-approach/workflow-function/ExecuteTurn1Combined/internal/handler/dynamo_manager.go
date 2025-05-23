package handler

import (
	"context"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
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
) <-chan error {
	updateComplete := make(chan error, 2)
	contextLogger := d.log.WithCorrelationId(verificationID)
	
	go func() {
		// Update verification status
		updateErr := d.dynamo.UpdateVerificationStatus(ctx, verificationID, models.StatusTurn1Completed, tokenUsage)
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
		}
		updateComplete <- updateErr
		
		// Record conversation history
		conversationTurn := &models.ConversationTurn{
			VerificationID:   verificationID,
			TurnID:           1,
			RawResponseRef:   rawRef,
			ProcessedRef:     procRef,
			TokenUsage:       tokenUsage,
			BedrockRequestID: requestID,
			Timestamp:        time.Now(),
		}
		
		historyErr := d.dynamo.RecordConversationTurn(ctx, conversationTurn)
		if historyErr != nil {
			contextLogger.Warn("conversation history recording failed", map[string]interface{}{
				"error": historyErr.Error(),
				"table": d.cfg.AWS.DynamoDBConversationTable,
			})
		}
		updateComplete <- historyErr
	}()
	
	return updateComplete
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
					"error": err.Error(),
				})
			}
		case <-updateTimeout:
			contextLogger.Warn("async updates timed out", map[string]interface{}{
				"updates_received": updatesReceived,
				"timeout_seconds": timeout.Seconds(),
			})
			return
		}
	}
}