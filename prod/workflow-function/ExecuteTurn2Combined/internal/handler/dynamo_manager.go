package handler

import (
	"context"
	"time"

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
	config config.Config
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
func NewDynamoManager(dynamo services.DynamoDBService, cfg config.Config, log logger.Logger) *DynamoManager {
	return &DynamoManager{dynamo: dynamo, log: log, config: cfg}
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
		d.logEnhancedDynamoDBError(err, "UpdateVerificationStatusEnhanced", verificationID, map[string]interface{}{
			"verificationAt": initialVerificationAt,
			"status":         statusEntry.Status,
			"stage":          statusEntry.Stage,
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateConversationTurn(ctx, verificationID, turnEntry); err != nil {
		d.logEnhancedDynamoDBError(err, "UpdateConversationTurn", verificationID, map[string]interface{}{
			"turnId":    turnEntry.TurnId,
			"stage":     turnEntry.Stage,
			"timestamp": turnEntry.Timestamp,
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateTurn1CompletionDetails(ctx, verificationID, initialVerificationAt, statusEntry, turn1Metrics, processedMarkdownRef, nil); err != nil {
		d.logEnhancedDynamoDBError(err, "UpdateTurn1CompletionDetails", verificationID, map[string]interface{}{
			"verificationAt":     initialVerificationAt,
			"hasMetrics":        turn1Metrics != nil,
			"hasProcessedRef":   processedMarkdownRef != nil,
		})
		dynamoOK = false
	}

	return dynamoOK
}

// UpdateTurn2Completion persists Turn2 processing results and status
func (d *DynamoManager) UpdateTurn2Completion(ctx context.Context, res Turn2Result) bool {
	dynamoOK := true

	// Debug logging to track verificationID flow
	d.log.Debug("UpdateTurn2Completion_called", map[string]interface{}{
		"verification_id":        res.VerificationID,
		"verification_id_empty":  res.VerificationID == "",
		"verification_id_length": len(res.VerificationID),
		"verification_at":        res.VerificationAt,
		"turn_id":               res.TurnEntry.TurnId,
		"turn_stage":            res.TurnEntry.Stage,
	})

	if err := d.dynamo.UpdateVerificationStatusEnhanced(ctx, res.VerificationID, res.VerificationAt, res.StatusEntry); err != nil {
		d.logEnhancedDynamoDBError(err, "UpdateVerificationStatusEnhanced", res.VerificationID, map[string]interface{}{
			"verificationAt": res.VerificationAt,
			"status":         res.StatusEntry.Status,
			"stage":          res.StatusEntry.Stage,
		})
		dynamoOK = false
	}

	// Additional debug logging before UpdateConversationTurn
	d.log.Debug("before_UpdateConversationTurn", map[string]interface{}{
		"verification_id":        res.VerificationID,
		"verification_id_empty":  res.VerificationID == "",
		"turn_entry_not_nil":     res.TurnEntry != nil,
	})

	if err := d.dynamo.UpdateConversationTurn(ctx, res.VerificationID, res.TurnEntry); err != nil {
		d.logEnhancedDynamoDBError(err, "UpdateConversationTurn", res.VerificationID, map[string]interface{}{
			"turnId":    res.TurnEntry.TurnId,
			"stage":     res.TurnEntry.Stage,
			"timestamp": res.TurnEntry.Timestamp,
		})
		dynamoOK = false
	}

	if err := d.dynamo.UpdateTurn2CompletionDetails(ctx, res.VerificationID, res.VerificationAt, res.StatusEntry, res.Metrics, res.ProcessedMarkdownRef, res.VerificationStatus, res.Discrepancies, res.ComparisonSummary, res.ConversationRef); err != nil {
		d.logEnhancedDynamoDBError(err, "UpdateTurn2CompletionDetails", res.VerificationID, map[string]interface{}{
			"verificationAt":       res.VerificationAt,
			"hasMetrics":          res.Metrics != nil,
			"hasProcessedRef":     res.ProcessedMarkdownRef != nil,
			"verificationStatus":  res.VerificationStatus,
			"discrepancyCount":    len(res.Discrepancies),
			"hasComparisonSummary": res.ComparisonSummary != "",
			"hasConversationRef":  res.ConversationRef != nil,
		})
		dynamoOK = false
	}

	return dynamoOK
}

// logEnhancedDynamoDBError provides comprehensive DynamoDB error logging with detailed context
func (d *DynamoManager) logEnhancedDynamoDBError(err error, operation string, verificationID string, context map[string]interface{}) {
	startTime := time.Now()

	// Analyze the error using the shared errors package
	var enhancedErr *errors.WorkflowError
	if workflowErr, ok := err.(*errors.WorkflowError); ok {
		enhancedErr = workflowErr
	} else {
		// Use the shared error analysis function
		enhancedErr = errors.AnalyzeDynamoDBError(operation, d.getTableNameForOperation(operation), err)
		enhancedErr.VerificationID = verificationID
	}

	// Create comprehensive logging context
	logContext := map[string]interface{}{
		"operation":        operation,
		"verification_id":  verificationID,
		"error_type":       string(enhancedErr.Type),
		"error_code":       enhancedErr.Code,
		"error_message":    enhancedErr.Message,
		"severity":         string(enhancedErr.Severity),
		"category":         string(enhancedErr.Category),
		"retryable":        enhancedErr.Retryable,
		"retry_strategy":   string(enhancedErr.RetryStrategy),
		"api_source":       string(enhancedErr.APISource),
		"table_name":       enhancedErr.TableName,
		"component":        "DynamoManager",
		"timestamp":        startTime.Format(time.RFC3339),
	}

	// Add operation-specific context
	for key, value := range context {
		logContext[key] = value
	}

	// Add error details if available
	if enhancedErr.Details != nil {
		logContext["error_details"] = enhancedErr.Details
	}

	// Add suggestions and recovery hints
	if len(enhancedErr.Suggestions) > 0 {
		logContext["suggestions"] = enhancedErr.Suggestions
	}
	if len(enhancedErr.RecoveryHints) > 0 {
		logContext["recovery_hints"] = enhancedErr.RecoveryHints
	}

	// Log with appropriate level based on severity
	switch enhancedErr.Severity {
	case errors.ErrorSeverityCritical:
		d.log.Error("dynamodb_operation_critical_failure", logContext)
	case errors.ErrorSeverityHigh:
		d.log.Error("dynamodb_operation_failed", logContext)
	case errors.ErrorSeverityMedium:
		d.log.Warn("dynamodb_operation_failed", logContext)
	case errors.ErrorSeverityLow:
		d.log.Warn("dynamodb_operation_retryable_failure", logContext)
	default:
		d.log.Warn("dynamodb_operation_failed", logContext)
	}
}

// getTableNameForOperation returns the appropriate table name for the given operation
func (d *DynamoManager) getTableNameForOperation(operation string) string {
	switch operation {
	case "UpdateVerificationStatusEnhanced", "UpdateTurn1CompletionDetails", "UpdateTurn2CompletionDetails":
		return d.config.AWS.DynamoDBVerificationTable
	case "UpdateConversationTurn":
		return d.config.AWS.DynamoDBConversationTable
	default:
		return "unknown_table"
	}
}
