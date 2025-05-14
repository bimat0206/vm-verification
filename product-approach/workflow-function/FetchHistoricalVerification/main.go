package main

import (
	"context"
	"fmt"
	
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"workflow-function/shared/schema"
)

// Handler is the Lambda handler function
func handler(ctx context.Context, event InputEvent) (OutputEvent, error) {
	// Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		return OutputEvent{}, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Configure dependencies with environment variables
	cfg := LoadConfig()
	deps.ConfigureDbUtils(cfg)

	logger := deps.GetLogger()
	logger.Info("Processing event", map[string]interface{}{
		"verificationId":   event.VerificationContext.VerificationId,
		"verificationType": event.VerificationContext.VerificationType,
	})

	// Validate input
	if err := validateInput(event.VerificationContext); err != nil {
		logger.Error("Input validation error", map[string]interface{}{
			"error": err.Error(),
		})
		return OutputEvent{}, fmt.Errorf("input validation error: %w", err)
	}

	// Create service instances
	dbWrapper := NewDBWrapper(deps.GetDynamoUtil(), logger)
	service := NewHistoricalVerificationService(dbWrapper, logger)

	// Process the request
	result, err := service.FetchHistoricalVerification(ctx, event.VerificationContext)
	if err != nil {
		logger.Error("Error fetching historical verification", map[string]interface{}{
			"error": err.Error(),
		})
		return OutputEvent{}, fmt.Errorf("error fetching historical verification: %w", err)
	}

	// Return the result
	logger.Info("Successfully retrieved historical verification", map[string]interface{}{
		"previousVerificationId": result.PreviousVerificationID,
		"hoursSince":             result.HoursSinceLastVerification,
	})

	return OutputEvent{
		HistoricalContext: result,
	}, nil
}

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context) (*Dependencies, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return NewDependencies(awsCfg), nil
}

// validateInput validates the input parameters
func validateInput(ctx schema.VerificationContext) error {
	// Check required fields
	if ctx.VerificationId == "" {
		return NewValidationError("missing verificationId", nil)
	}

	if ctx.VerificationType == "" {
		return NewValidationError("missing verificationType", nil)
	}

	// Ensure verificationType is 'PREVIOUS_VS_CURRENT'
	if ctx.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return NewValidationError(
			"invalid verificationType, expected 'PREVIOUS_VS_CURRENT'",
			map[string]string{
				"expected": schema.VerificationTypePreviousVsCurrent,
				"actual":   ctx.VerificationType,
			},
		)
	}

	if ctx.ReferenceImageUrl == "" {
		return NewValidationError("missing referenceImageUrl", nil)
	}

	if ctx.CheckingImageUrl == "" {
		return NewValidationError("missing checkingImageUrl", nil)
	}

	if ctx.VendingMachineId == "" {
		return NewValidationError("missing vendingMachineId", nil)
	}

	// Verify S3 URL format for reference image
	if !isValidS3Url(ctx.ReferenceImageUrl) {
		return NewValidationError("invalid reference image URL format, expected s3:// prefix", 
			map[string]string{"url": ctx.ReferenceImageUrl})
	}

	// For previous_vs_current, reference image should be in the checking bucket
	if !isCheckingBucketURL(ctx.ReferenceImageUrl) {
		return NewValidationError(
			"for PREVIOUS_VS_CURRENT verification, referenceImageUrl must point to the checking bucket",
			map[string]string{"url": ctx.ReferenceImageUrl},
		)
	}

	return nil
}

// isValidS3Url checks if the URL has the s3:// prefix
func isValidS3Url(url string) bool {
	return len(url) > 5 && url[:5] == "s3://"
}

// isCheckingBucketURL checks if the URL is from the checking bucket
func isCheckingBucketURL(url string) bool {
	checkingBucket := getEnv(EnvCheckingBucketName)
	return len(url) > len(checkingBucket) && url[5:5+len(checkingBucket)] == checkingBucket
}

func main() {
	lambda.Start(handler)
}