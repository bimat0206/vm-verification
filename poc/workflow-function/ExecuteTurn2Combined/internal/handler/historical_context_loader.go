package handler

import (
	"context"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/logger"
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

// LoadHistoricalContext loads historical context for PREVIOUS_VS_CURRENT verification type (legacy Turn1 method)
// This method is deprecated for Turn2 processing.
func (h *HistoricalContextLoader) LoadHistoricalContext(ctx context.Context, req interface{}, contextLogger logger.Logger) (time.Duration, error) {
	// This method is deprecated for Turn2 processing
	// Historical context loading is now handled in LoadContextTurn2
	contextLogger.Warn("LoadHistoricalContext method is deprecated for Turn2 processing", nil)
	return 0, fmt.Errorf("LoadHistoricalContext method is deprecated for Turn2 processing")
}
