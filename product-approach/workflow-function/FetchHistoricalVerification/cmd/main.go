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
	start := time.Now()
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
	if envelope.Summary == nil {
		envelope.Summary = make(map[string]interface{})
	}
	envelope.Summary["historicalDataFound"] = true
	envelope.Summary["previousVerificationId"] = result.PreviousVerificationID
	envelope.Summary["previousVerificationAt"] = result.PreviousVerificationAt
	envelope.Summary["previousStatus"] = result.PreviousVerificationStatus
	envelope.Summary["processingTimeMs"] = time.Since(start).Milliseconds()

	return internal.OutputEvent{
		VerificationID: envelope.VerificationID,
		S3References:   envelope.References,
		Status:         envelope.Status,
		Summary:        envelope.Summary,
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

func main() {
	lambda.Start(handler)
}
