package handler

import (
	"context"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// DynamoManager wraps DynamoDB operations used by the handler.
type DynamoManager struct {
	dynamo services.DynamoDBService
	log    logger.Logger
}

// Turn2Result holds data required to finalize Turn2 processing
type Turn2Result struct {
	VerificationID       string
	VerificationAt       string
	StatusEntry          schema.StatusHistoryEntry
	TurnEntry            *schema.TurnResponse
	Metrics              *schema.TurnMetrics
	ProcessedMarkdownRef *models.S3Reference
	VerificationStatus   string
	Discrepancies        []schema.Discrepancy
	ComparisonSummary    string
	ConversationRef      *models.S3Reference
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

	if err := d.dynamo.UpdateTurn1CompletionDetails(ctx, verificationID, initialVerificationAt, statusEntry, turn1Metrics, processedMarkdownRef, nil); err != nil {
		d.log.Warn("dynamodb update turn1 completion details failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	return dynamoOK
}

// UpdateTurn2Completion persists Turn2 processing results and status
func (d *DynamoManager) UpdateTurn2Completion(ctx context.Context, res Turn2Result) bool {
	dynamoOK := true

	if err := d.dynamo.UpdateVerificationStatusEnhanced(ctx, res.VerificationID, res.VerificationAt, res.StatusEntry); err != nil {
		d.log.Warn("dynamodb status update failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateConversationTurn(ctx, res.VerificationID, res.TurnEntry); err != nil {
		d.log.Warn("conversation history recording failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateTurn2CompletionDetails(ctx, res.VerificationID, res.VerificationAt, res.StatusEntry, res.Metrics, res.ProcessedMarkdownRef, res.VerificationStatus, res.Discrepancies, res.ComparisonSummary, res.ConversationRef); err != nil {
		d.log.Warn("dynamodb update turn2 completion details failed", map[string]interface{}{
			"error":     err.Error(),
			"retryable": errors.IsRetryable(err),
		})
		dynamoOK = false
	}

	return dynamoOK
}
