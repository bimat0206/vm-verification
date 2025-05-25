package handler

import (
	"context"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// HistoricalContextLoader handles loading historical verification context
type HistoricalContextLoader struct {
	dynamo services.DynamoDBService
	s3     services.S3StateManager
	log    logger.Logger
}

// NewHistoricalContextLoader creates a new instance of HistoricalContextLoader
func NewHistoricalContextLoader(s3 services.S3StateManager, dynamo services.DynamoDBService, log logger.Logger) *HistoricalContextLoader {
	return &HistoricalContextLoader{
		dynamo: dynamo,
		s3:     s3,
		log:    log,
	}
}

// LoadHistoricalContext loads historical context for PREVIOUS_VS_CURRENT verification type
func (h *HistoricalContextLoader) LoadHistoricalContext(ctx context.Context, req *models.Turn1Request, contextLogger logger.Logger) (time.Duration, error) {
	if req.VerificationContext.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return 0, nil
	}

	historicalStart := time.Now()

	if ref := req.S3Refs.Processing.HistoricalContext; ref.Key != "" {
		contextLogger.Info("loading_historical_context_from_s3", map[string]interface{}{
			"bucket": ref.Bucket,
			"key":    ref.Key,
		})

		var data map[string]interface{}
		err := h.s3.LoadJSON(ctx, ref, &data)
		if err != nil {
			contextLogger.Warn("failed_to_load_historical_context_from_s3", map[string]interface{}{
				"error": err.Error(),
				"key":   ref.Key,
			})
			return time.Since(historicalStart), nil
		}

		if data != nil {
			if req.VerificationContext.HistoricalContext == nil {
				req.VerificationContext.HistoricalContext = make(map[string]interface{})
			}
			for k, v := range data {
				req.VerificationContext.HistoricalContext[k] = v
			}

			if ts, ok := data["PreviousVerificationAt"].(string); ok {
				req.VerificationContext.HistoricalContext["HoursSinceLastVerification"] = calculateHoursSince(ts)
			}

			contextLogger.Info("historical_context_loaded_from_s3", map[string]interface{}{
				"key": ref.Key,
			})
		}
		return time.Since(historicalStart), nil
	}

	contextLogger.Warn("historical_context_s3_reference_missing", map[string]interface{}{
		"verification_id": req.VerificationID,
	})

	return time.Since(historicalStart), nil
}
