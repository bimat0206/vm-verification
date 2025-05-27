package handler

import (
	"context"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// DynamoManager wraps DynamoDB operations used by the handler.
type DynamoManager struct {
	dynamo services.DynamoDBService
	log    logger.Logger
}

// NewDynamoManager creates a DynamoManager instance.
func NewDynamoManager(dynamo services.DynamoDBService, _ config.Config, log logger.Logger) *DynamoManager {
	return &DynamoManager{dynamo: dynamo, log: log}
}

// Update writes the final status and conversation turn. It returns true
// only if both writes succeed.
func (d *DynamoManager) UpdateTurn1Completion(
	ctx context.Context,
	verificationID string,
	initialVerificationAt string,
	statusEntry schema.StatusHistoryEntry,
	turnEntry *schema.TurnResponse,
	turn1Metrics *schema.TurnMetrics,
	processedMarkdownRef *models.S3Reference,
) bool {
	dynamoOK := true

	if err := d.dynamo.UpdateVerificationStatusEnhanced(ctx, verificationID, initialVerificationAt, statusEntry); err != nil {
		d.log.Warn("dynamodb status update failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateConversationTurn(ctx, verificationID, turnEntry); err != nil {
		d.log.Warn("conversation history recording failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateTurn1CompletionDetails(ctx, verificationID, initialVerificationAt, statusEntry, turn1Metrics, processedMarkdownRef); err != nil {
		d.log.Warn("dynamodb update turn1 completion details failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	return dynamoOK
}
