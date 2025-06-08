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

	// Create enhanced verification context with historical data
	enhancedVerificationContext := createEnhancedVerificationContext(vCtx, result)

	// Update the initialization.json file with the enhanced verification context
	if err := updateInitializationFile(deps.GetStateManager(), initRef, enhancedVerificationContext, logger); err != nil {
		logger.Error("Failed to update initialization file", map[string]interface{}{"error": err.Error()})
		return internal.OutputEvent{}, fmt.Errorf("failed to update initialization file: %w", err)
	}

	return internal.OutputEvent{
		VerificationID:      envelope.VerificationID,
		S3References:        envelope.References,
		Status:              envelope.Status,
		VerificationContext: enhancedVerificationContext,
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

// createEnhancedVerificationContext creates an enhanced verification context for updating initialization.json
func createEnhancedVerificationContext(inputCtx schema.VerificationContext, historicalResult internal.HistoricalContext) *internal.EnhancedVerificationContext {
	now := time.Now().UTC().Format(time.RFC3339)

	// Create the base verification context with all original fields preserved
	baseVerificationContext := &schema.VerificationContext{
		VerificationId:         inputCtx.VerificationId,
		VerificationAt:         inputCtx.VerificationAt,
		Status:                 schema.StatusHistoricalContextLoaded,
		VerificationType:       inputCtx.VerificationType,
		ConversationType:       inputCtx.ConversationType,
		VendingMachineId:       inputCtx.VendingMachineId,
		LayoutId:               inputCtx.LayoutId,
		LayoutPrefix:           inputCtx.LayoutPrefix,
		ReferenceImageUrl:      inputCtx.ReferenceImageUrl,
		CheckingImageUrl:       inputCtx.CheckingImageUrl,
		PreviousVerificationId: inputCtx.PreviousVerificationId,
		ResourceValidation:     inputCtx.ResourceValidation,
		RequestMetadata:        inputCtx.RequestMetadata,
		TurnConfig:             inputCtx.TurnConfig,
		TurnTimestamps:         inputCtx.TurnTimestamps,
		LastUpdatedAt:          now,
		Error:                  inputCtx.Error,
		// Copy enhanced fields if they exist
		CurrentStatus:     inputCtx.CurrentStatus,
		StatusHistory:     inputCtx.StatusHistory,
		ProcessingMetrics: inputCtx.ProcessingMetrics,
		ErrorTracking:     inputCtx.ErrorTracking,
	}

	// Ensure ResourceValidation exists
	if baseVerificationContext.ResourceValidation == nil {
		baseVerificationContext.ResourceValidation = &schema.ResourceValidation{
			ReferenceImageExists: true, // Assume true since we're processing
			CheckingImageExists:  true, // Assume true since we're processing
			ValidationTimestamp:  now,
		}
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

// updateInitializationFile updates the initialization.json file with enhanced verification context
func updateInitializationFile(stateManager s3state.Manager, initRef *s3state.Reference, enhancedCtx *internal.EnhancedVerificationContext, logger interface{}) error {
	// Load the current initialization data
	var initData struct {
		SchemaVersion       string                      `json:"schemaVersion"`
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
		SystemPrompt        map[string]interface{}      `json:"systemPrompt,omitempty"`
		LayoutMetadata      interface{}                 `json:"layoutMetadata,omitempty"`
	}

	if err := stateManager.RetrieveJSON(initRef, &initData); err != nil {
		return fmt.Errorf("failed to load initialization data: %w", err)
	}

	// Update the verification context with enhanced data
	// First, preserve the original verification context fields
	originalCtx := initData.VerificationContext
	if originalCtx != nil {
		// Update the status and timestamp
		originalCtx.Status = schema.StatusHistoricalContextLoaded
		originalCtx.LastUpdatedAt = enhancedCtx.VerificationContext.LastUpdatedAt

		// Add the enhanced fields as custom fields in the verification context
		// Since we can't modify the schema.VerificationContext directly, we'll create a custom structure
	}

	// Create a new structure that includes both the original verification context and the enhanced fields
	enhancedInitData := map[string]interface{}{
		"schemaVersion": initData.SchemaVersion,
		"verificationContext": map[string]interface{}{
			// Copy all original fields
			"verificationId":         enhancedCtx.VerificationContext.VerificationId,
			"verificationAt":         enhancedCtx.VerificationContext.VerificationAt,
			"status":                 enhancedCtx.VerificationContext.Status,
			"verificationType":       enhancedCtx.VerificationContext.VerificationType,
			"conversationType":       enhancedCtx.VerificationContext.ConversationType,
			"vendingMachineId":       enhancedCtx.VerificationContext.VendingMachineId,
			"layoutId":               enhancedCtx.VerificationContext.LayoutId,
			"layoutPrefix":           enhancedCtx.VerificationContext.LayoutPrefix,
			"referenceImageUrl":      enhancedCtx.VerificationContext.ReferenceImageUrl,
			"checkingImageUrl":       enhancedCtx.VerificationContext.CheckingImageUrl,
			"previousVerificationId": enhancedCtx.VerificationContext.PreviousVerificationId,
			"resourceValidation":     enhancedCtx.VerificationContext.ResourceValidation,
			"requestMetadata":        enhancedCtx.VerificationContext.RequestMetadata,
			"turnConfig":             enhancedCtx.VerificationContext.TurnConfig,
			"turnTimestamps":         enhancedCtx.VerificationContext.TurnTimestamps,
			"lastUpdatedAt":          enhancedCtx.VerificationContext.LastUpdatedAt,
			"error":                  enhancedCtx.VerificationContext.Error,
			"currentStatus":          enhancedCtx.VerificationContext.CurrentStatus,
			"statusHistory":          enhancedCtx.VerificationContext.StatusHistory,
			"processingMetrics":      enhancedCtx.VerificationContext.ProcessingMetrics,
			"errorTracking":          enhancedCtx.VerificationContext.ErrorTracking,
			// Add the enhanced historical fields
			"previousVerificationAt": enhancedCtx.PreviousVerificationAt,
			"turn2Processed":         enhancedCtx.Turn2Processed,
			"historicalDataFound":    enhancedCtx.HistoricalDataFound,
			"sourceType":             enhancedCtx.SourceType,
			"previousStatus":         enhancedCtx.PreviousStatus,
		},
		"systemPrompt":   initData.SystemPrompt,
		"layoutMetadata": initData.LayoutMetadata,
	}

	// Save the updated initialization data back to S3
	_, err := stateManager.StoreJSON("", initRef.Key, enhancedInitData)
	if err != nil {
		return fmt.Errorf("failed to save updated initialization data: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
