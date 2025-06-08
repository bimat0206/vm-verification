package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	configaws "github.com/aws/aws-sdk-go-v2/config"

	"workflow-function/FetchHistoricalVerification/internal"
	"workflow-function/shared/errors"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (internal.OutputEvent, error) {
	cfg := internal.LoadConfig()

	deps, err := initDependencies(ctx, cfg)
	if err != nil {
		return internal.OutputEvent{}, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	logger := deps.GetLogger()

	envelope, err := s3state.LoadEnvelope(event)
	if err != nil {
		logger.Error("Failed to load envelope", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("failed to load envelope: %w", err)
	}

	initRef := envelope.GetReference("processing_initialization")
	if initRef == nil {
		return internal.OutputEvent{}, fmt.Errorf("initialization reference missing")
	}

	var initData struct {
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
	}
	if err := deps.GetStateManager().RetrieveJSON(initRef, &initData); err != nil {
		logger.Error("Failed to load initialization", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("failed to load initialization: %w", err)
	}
	if initData.VerificationContext == nil {
		return internal.OutputEvent{}, fmt.Errorf("verificationContext missing in initialization")
	}
	vCtx := *initData.VerificationContext

	if err := validateInput(vCtx); err != nil {
		logger.Error("Input validation error", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("input validation error: %w", err)
	}

	service := internal.NewHistoricalVerificationService(deps.GetDynamoRepo(), logger)
	result, err := service.FetchHistoricalVerification(ctx, vCtx)
	if err != nil {
		logger.Error("Error fetching historical verification", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("error fetching historical verification: %w", err)
	}

	if err := deps.GetStateManager().SaveToEnvelope(envelope, s3state.CategoryProcessing, s3state.HistoricalContextFile, result); err != nil {
		logger.Error("Failed to store historical context", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("failed to store historical context: %w", err)
	}

	envelope.SetStatus(schema.StatusHistoricalContextLoaded)

	// Create verification context for output
	verificationContext := createVerificationContext(vCtx, result)

	return internal.OutputEvent{
		VerificationID:      envelope.VerificationID,
		S3References:        envelope.References,
		Status:              envelope.Status,
		VerificationContext: verificationContext,
	}, nil
}

func initDependencies(ctx context.Context, config internal.ConfigVars) (*internal.Dependencies, error) {
	awsCfg, err := configaws.LoadDefaultConfig(ctx, configaws.WithRegion(config.Region))
	if err != nil {
		return nil, err
	}
	deps, err := internal.NewDependencies(awsCfg, config)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func validateInput(ctx schema.VerificationContext) error {
	if ctx.VerificationId == "" {
		return errors.NewMissingFieldError("verificationId")
	}
	if ctx.VerificationType == "" {
		return errors.NewMissingFieldError("verificationType")
	}
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
	return nil
}

// createVerificationContext creates an enhanced verification context for the output
func createVerificationContext(inputCtx schema.VerificationContext, historicalResult internal.HistoricalContext) *internal.EnhancedVerificationContext {
	now := time.Now().UTC().Format(time.RFC3339)

	// Create the base verification context
	baseVerificationContext := &schema.VerificationContext{
		VerificationId:    inputCtx.VerificationId,
		VerificationAt:    inputCtx.VerificationAt,
		Status:            schema.StatusHistoricalContextLoaded,
		VerificationType:  inputCtx.VerificationType,
		ReferenceImageUrl: inputCtx.ReferenceImageUrl,
		CheckingImageUrl:  inputCtx.CheckingImageUrl,
		VendingMachineId:  inputCtx.VendingMachineId,
		LayoutId:          inputCtx.LayoutId,
		LayoutPrefix:      inputCtx.LayoutPrefix,
		PreviousVerificationId: inputCtx.PreviousVerificationId,
		ResourceValidation: &schema.ResourceValidation{
			ReferenceImageExists: true, // Assume true since we're processing
			CheckingImageExists:  true, // Assume true since we're processing
			ValidationTimestamp:  now,
		},
		RequestMetadata: inputCtx.RequestMetadata,
		TurnConfig:      inputCtx.TurnConfig,
		TurnTimestamps:  inputCtx.TurnTimestamps,
		Error:           inputCtx.Error,
	}

	// Create enhanced verification context with historical data
	enhancedContext := &internal.EnhancedVerificationContext{
		VerificationContext: baseVerificationContext,
		HistoricalDataFound: historicalResult.HistoricalDataFound,
		SourceType:          historicalResult.SourceType,
		Turn2Processed:      historicalResult.Turn2Processed,
	}

	// Add previous verification data if available
	if historicalResult.PreviousVerification != nil {
		enhancedContext.PreviousVerificationId = historicalResult.PreviousVerification.VerificationID
		enhancedContext.PreviousVerificationAt = historicalResult.PreviousVerification.VerificationAt
		enhancedContext.PreviousStatus = historicalResult.PreviousVerification.VerificationStatus
	}

	return enhancedContext
}

func main() {
	lambda.Start(handler)
}
