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
	log    logger.Logger
}

// NewHistoricalContextLoader creates a new instance of HistoricalContextLoader
func NewHistoricalContextLoader(dynamo services.DynamoDBService, log logger.Logger) *HistoricalContextLoader {
	return &HistoricalContextLoader{
		dynamo: dynamo,
		log:    log,
	}
}

// LoadHistoricalContext loads historical context for PREVIOUS_VS_CURRENT verification type
func (h *HistoricalContextLoader) LoadHistoricalContext(ctx context.Context, req *models.Turn1Request, contextLogger logger.Logger) (time.Duration, error) {
	if req.VerificationContext.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return 0, nil
	}
	
	historicalStart := time.Now()
	
	// Extract checking image URL from the reference image S3 key
	checkingImageUrl := extractCheckingImageUrl(req.S3Refs.Images.ReferenceBase64.Key)
	
	if checkingImageUrl == "" {
		return time.Since(historicalStart), nil
	}
	
	contextLogger.Info("Loading historical verification context", map[string]interface{}{
		"checking_image_url": checkingImageUrl,
		"verification_type":  req.VerificationContext.VerificationType,
	})
	
	// Query for previous verification using the checking image URL
	previousVerification, err := h.dynamo.QueryPreviousVerification(ctx, checkingImageUrl)
	if err != nil {
		// Log warning but continue - historical context is optional enhancement
		contextLogger.Warn("Failed to load historical verification context", map[string]interface{}{
			"error":              err.Error(),
			"checking_image_url": checkingImageUrl,
		})
		return time.Since(historicalStart), nil
	}
	
	if previousVerification == nil {
		return time.Since(historicalStart), nil
	}
	
	// Populate historical context with previous verification data
	req.VerificationContext.HistoricalContext = map[string]interface{}{
		"PreviousVerificationAt":         previousVerification.VerificationAt,
		"PreviousVerificationStatus":     previousVerification.CurrentStatus,
		"PreviousVerificationId":         previousVerification.VerificationId,
		"HoursSinceLastVerification":     calculateHoursSince(previousVerification.VerificationAt),
	}
	
	// Add layout information from the previous verification
	if previousVerification.LayoutId > 0 {
		req.VerificationContext.HistoricalContext["LayoutId"] = previousVerification.LayoutId
		req.VerificationContext.HistoricalContext["LayoutPrefix"] = previousVerification.LayoutPrefix
	}
	
	// For now, set default row/column information
	// In a real implementation, this would come from additional DynamoDB attributes
	// or a separate layout metadata query
	req.VerificationContext.HistoricalContext["RowCount"] = 4
	req.VerificationContext.HistoricalContext["ColumnCount"] = 10
	req.VerificationContext.HistoricalContext["RowLabels"] = []string{"A", "B", "C", "D"}
	
	contextLogger.Info("Successfully loaded historical verification context", map[string]interface{}{
		"previous_verification_id":   previousVerification.VerificationId,
		"previous_verification_at":   previousVerification.VerificationAt,
		"hours_since_last":          req.VerificationContext.HistoricalContext["HoursSinceLastVerification"],
	})
	
	return time.Since(historicalStart), nil
}