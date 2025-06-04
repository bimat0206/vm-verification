package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	configaws "github.com/aws/aws-sdk-go-v2/config"

	"workflow-function/FetchHistoricalVerification/internal"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// handler is the Lambda handler function
func handler(ctx context.Context, event internal.InputEvent) (internal.OutputEvent, error) {
	// Load config first
	cfg := internal.LoadConfig()
	
	// Initialize dependencies
	deps, err := initDependencies(ctx, cfg)
	if err != nil {
		return internal.OutputEvent{}, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

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
		return internal.OutputEvent{}, fmt.Errorf("input validation error: %w", err)
	}

	// Use the dynamoRepo directly
	service := internal.NewHistoricalVerificationService(deps.GetDynamoRepo(), logger)

	// Process the request
	result, err := service.FetchHistoricalVerification(ctx, event.VerificationContext)
	if err != nil {
		logger.Error("Error fetching historical verification", map[string]interface{}{
			"error": err.Error(),
		})
		return internal.OutputEvent{}, fmt.Errorf("error fetching historical verification: %w", err)
	}

	// Return the result
	logger.Info("Successfully retrieved historical verification", map[string]interface{}{
		"previousVerificationId": result.PreviousVerificationID,
		"hoursSince":             result.HoursSinceLastVerification,
	})

	return internal.OutputEvent{
		HistoricalContext: result,
	}, nil
}

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context, config internal.ConfigVars) (*internal.Dependencies, error) {
	// Load AWS config with the region from our config
	awsCfg, err := configaws.LoadDefaultConfig(ctx, configaws.WithRegion(config.Region))
	if err != nil {
		return nil, err
	}
	return internal.NewDependencies(awsCfg, config), nil
}

// validateInput validates the input parameters
func validateInput(ctx schema.VerificationContext) error {
	// Check required fields
	if ctx.VerificationId == "" {
		return errors.NewMissingFieldError("verificationId")
	}

	if ctx.VerificationType == "" {
		return errors.NewMissingFieldError("verificationType")
	}

	// Ensure verificationType is 'PREVIOUS_VS_CURRENT'
	if ctx.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return errors.NewValidationError(
			"invalid verificationType, expected 'PREVIOUS_VS_CURRENT'",
			map[string]interface{}{
				"expected": schema.VerificationTypePreviousVsCurrent,
				"actual":   ctx.VerificationType,
			},
		)
	}

	if ctx.ReferenceImageUrl == "" {
		return errors.NewMissingFieldError("referenceImageUrl")
	}

	if ctx.CheckingImageUrl == "" {
		return errors.NewMissingFieldError("checkingImageUrl")
	}

	if ctx.VendingMachineId == "" {
		return errors.NewMissingFieldError("vendingMachineId")
	}

	// Verify S3 URL format for reference image
	if !isValidS3Url(ctx.ReferenceImageUrl) {
		return errors.NewValidationError("invalid reference image URL format, expected s3:// prefix",
			map[string]interface{}{"url": ctx.ReferenceImageUrl})
	}

	// For previous_vs_current, reference image should be in the checking bucket
	if !isCheckingBucketURL(ctx.ReferenceImageUrl) {
		return errors.NewValidationError(
			"for PREVIOUS_VS_CURRENT verification, referenceImageUrl must point to the checking bucket",
			map[string]interface{}{"url": ctx.ReferenceImageUrl},
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
	checkingBucket := os.Getenv("CHECKING_BUCKET")
	return len(url) > len(checkingBucket) && url[5:5+len(checkingBucket)] == checkingBucket
}

func main() {
	lambda.Start(handler)
}