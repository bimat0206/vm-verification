package main

import (
	"context"
	"log"
	"time"
)

// HistoricalVerificationService handles fetching historical verification data
type HistoricalVerificationService struct {
	db     *DynamoDBClient
	logger *log.Logger
}

// NewHistoricalVerificationService creates a new service instance
func NewHistoricalVerificationService(db *DynamoDBClient, logger *log.Logger) *HistoricalVerificationService {
	return &HistoricalVerificationService{
		db:     db,
		logger: logger,
	}
}

// FetchHistoricalVerification retrieves historical verification data for "previous_vs_current" verification type
func (s *HistoricalVerificationService) FetchHistoricalVerification(ctx context.Context, verificationCtx VerificationContext) (HistoricalContext, error) {
	s.logger.Printf("Fetching historical verification for reference image: %s", verificationCtx.ReferenceImageURL)

	// Query DynamoDB to find most recent verification using referenceImageUrl as checking image
	verification, err := s.db.QueryMostRecentVerificationByCheckingImage(ctx, verificationCtx.ReferenceImageURL)
	if err != nil {
		return HistoricalContext{}, err
	}

	s.logger.Printf("Found previous verification: %s at %s", verification.VerificationID, verification.VerificationAt)

	// Calculate hours since last verification
	prevTime, err := time.Parse(time.RFC3339, verification.VerificationAt)
	if err != nil {
		s.logger.Printf("Warning: Could not parse previous verification time: %v", err)
		prevTime = time.Now() // Fallback to current time
	}

	currentTime, err := time.Parse(time.RFC3339, verificationCtx.VerificationAt)
	if err != nil {
		s.logger.Printf("Warning: Could not parse current verification time: %v", err)
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