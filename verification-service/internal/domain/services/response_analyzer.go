package services

import (
	"context"

	"verification-service/internal/domain/models"
)

// ResponseAnalyzer defines the interface for analyzing AI responses
type ResponseAnalyzer interface {
	// ProcessTurn1Response analyzes the Turn 1 (reference layout) response
	ProcessTurn1Response(
		ctx context.Context,
		response string,
		layoutMetadata map[string]interface{},
	) (*models.ReferenceAnalysis, error)

	// ProcessTurn2Response analyzes the Turn 2 (checking image) response
	ProcessTurn2Response(
		ctx context.Context,
		response string,
		referenceAnalysis *models.ReferenceAnalysis,
	) (*models.CheckingAnalysis, error)
}