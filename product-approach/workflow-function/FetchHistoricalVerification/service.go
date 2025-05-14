package main

import (
	"context"
	"time"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// HistoricalVerificationService handles fetching historical verification data
type HistoricalVerificationService struct {
	db     *DBWrapper
	logger logger.Logger
}

// NewHistoricalVerificationService creates a new service instance
func NewHistoricalVerificationService(db *DBWrapper, log logger.Logger) *HistoricalVerificationService {
	return &HistoricalVerificationService{
		db:     db,
		logger: log.WithFields(map[string]interface{}{
			"component": "historical-verification-service",
		}),
	}
}

// FetchHistoricalVerification retrieves historical verification data for "previous_vs_current" verification type
func (s *HistoricalVerificationService) FetchHistoricalVerification(ctx context.Context, verificationCtx schema.VerificationContext) (HistoricalContext, error) {
	s.logger.Info("Fetching historical verification", map[string]interface{}{
		"referenceImageUrl": verificationCtx.ReferenceImageUrl,
		"verificationId":    verificationCtx.VerificationId,
	})

	// Query DynamoDB to find most recent verification using referenceImageUrl as checking image
	verification, err := s.db.QueryMostRecentVerificationByCheckingImage(ctx, verificationCtx.ReferenceImageUrl)
	if err != nil {
		s.logger.Error("Failed to fetch historical verification", map[string]interface{}{
			"error": err.Error(),
		})
		return HistoricalContext{}, err
	}

	s.logger.Info("Found previous verification", map[string]interface{}{
		"previousVerificationId": verification.VerificationID,
		"previousVerificationAt": verification.VerificationAt,
	})

	// Calculate hours since last verification
	prevTime, err := time.Parse(time.RFC3339, verification.VerificationAt)
	if err != nil {
		s.logger.Warn("Could not parse previous verification time", map[string]interface{}{
			"error":              err.Error(),
			"verificationTime":   verification.VerificationAt,
		})
		prevTime = time.Now() // Fallback to current time
	}

	currentTime, err := time.Parse(time.RFC3339, verificationCtx.VerificationAt)
	if err != nil {
		s.logger.Warn("Could not parse current verification time", map[string]interface{}{
			"error":            err.Error(),
			"verificationTime": verificationCtx.VerificationAt,
		})
		currentTime = time.Now() // Fallback to current time
	}

	hoursSinceLastVerification := currentTime.Sub(prevTime).Hours()

	// Create historical context
	historicalContext := HistoricalContext{
		PreviousVerificationID:     verification.VerificationID,
		PreviousVerificationAt:     verification.VerificationAt,
		PreviousVerificationStatus: verification.VerificationStatus,
		HoursSinceLastVerification: hoursSinceLastVerification,
		MachineStructure:           verification.MachineStructure,
		CheckingStatus:             verification.CheckingStatus,
		VerificationSummary:        verification.VerificationSummary,
	}

	return historicalContext, nil
}