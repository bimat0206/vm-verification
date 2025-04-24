package services

import (
	"context"

	"verification-service/internal/domain/models"
)

// VerificationEngine defines the interface for the core verification logic
type VerificationEngine interface {
	// AnalyzeReferenceLayout processes the reference layout image (Turn 1)
	AnalyzeReferenceLayout(
		ctx context.Context,
		verificationContext models.VerificationContext,
		referenceImage []byte,
		layoutMetadata map[string]interface{},
	) (*models.ReferenceAnalysis, error)

	// VerifyCheckingImage compares the checking image to the reference layout (Turn 2)
	VerifyCheckingImage(
		ctx context.Context,
		verificationContext models.VerificationContext, 
		checkingImage []byte,
		referenceAnalysis *models.ReferenceAnalysis,
	) (*models.CheckingAnalysis, error)
}