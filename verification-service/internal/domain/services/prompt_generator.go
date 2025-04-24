package services

import (
	"context"

	"verification-service/internal/domain/models"
)

// PromptGenerator defines the interface for generating AI prompts
type PromptGenerator interface {
	// GenerateSystemPrompt creates the base system prompt for the verification process
	GenerateSystemPrompt(
		ctx context.Context,
		verificationContext models.VerificationContext,
		layoutMetadata map[string]interface{},
	) (string, error)

	// GenerateTurn1Prompt creates the prompt for reference layout analysis
	GenerateTurn1Prompt(
		ctx context.Context,
		verificationContext models.VerificationContext,
		layoutMetadata map[string]interface{}, 
		referenceImageB64 string,
	) (string, error)

	// GenerateTurn2Prompt creates the prompt for checking image comparison
	GenerateTurn2Prompt(
		ctx context.Context,
		verificationContext models.VerificationContext,
		layoutMetadata map[string]interface{},
		checkingImageB64 string,
		referenceAnalysis *models.ReferenceAnalysis,
	) (string, error)
}