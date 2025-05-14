package main

import (
	"context"
	"fmt"
	"time"
	"workflow-function/shared/dbutils"
)

// DBUtilsWrapper wraps the shared dbutils package with specific functionality needed for FetchImages
type DBUtilsWrapper struct {
	dbUtils *dbutils.DynamoDBUtils
}

// NewDBUtils creates a new DBUtilsWrapper
func NewDBUtils(db *dbutils.DynamoDBUtils) *DBUtilsWrapper {
	return &DBUtilsWrapper{
		dbUtils: db,
	}
}

// FetchLayoutMetadata retrieves complete layout metadata from DynamoDB.
func (d *DBUtilsWrapper) FetchLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (map[string]interface{}, error) {
	// Get the complete layout metadata using the shared dbutils package
	layout, err := d.dbUtils.GetLayoutMetadata(ctx, layoutId, layoutPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve layout metadata from DynamoDB: %w", err)
	}

	// Convert the structured layout metadata to a map for the response
	result := map[string]interface{}{
		"layoutId":           layout.LayoutId,
		"layoutPrefix":       layout.LayoutPrefix,
		"vendingMachineId":   layout.VendingMachineId,
		"location":           layout.Location,
		"createdAt":          layout.CreatedAt,
		"updatedAt":          layout.UpdatedAt,
		"referenceImageUrl":  layout.ReferenceImageUrl,
		"sourceJsonUrl":      layout.SourceJsonUrl,
		"machineStructure":   layout.MachineStructure,
		"productPositionMap": layout.ProductPositionMap,
	}

	// Add derived fields for convenience
	// layout.MachineStructure is already map[string]interface{}, no type assertion needed
	if machineStruct := layout.MachineStructure; machineStruct != nil {
		if rowCount, exists := machineStruct["rowCount"]; exists {
			result["derivedRowCount"] = rowCount
		}
		if columnsPerRow, exists := machineStruct["columnsPerRow"]; exists {
			result["derivedColumnsPerRow"] = columnsPerRow
		}
	}

	// Add product count if productPositionMap exists
	// layout.ProductPositionMap is already map[string]interface{}, no type assertion needed
	if prodMap := layout.ProductPositionMap; prodMap != nil {
		result["totalProductPositions"] = len(prodMap)
	}

	return result, nil
}

// FetchHistoricalContext retrieves previous verification from DynamoDB.
func (d *DBUtilsWrapper) FetchHistoricalContext(ctx context.Context, verificationId string) (map[string]interface{}, error) {
	// Get verification record using the shared package
	verificationCtx, err := d.dbUtils.GetVerificationRecord(ctx, verificationId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical verification: %w", err)
	}

	// Calculate hours since verification
	var hoursSince float64 = 0
	if verificationCtx.VerificationAt != "" {
		verTime, err := time.Parse(time.RFC3339, verificationCtx.VerificationAt)
		if err == nil {
			hoursSince = time.Since(verTime).Hours()
		}
	}

	// Extract machine structure from TurnConfig if available
	var machineStructure map[string]interface{}
	if verificationCtx.TurnConfig != nil {
		machineStructure = map[string]interface{}{
			"maxTurns":           verificationCtx.TurnConfig.MaxTurns,
			"referenceImageTurn": verificationCtx.TurnConfig.ReferenceImageTurn,
			"checkingImageTurn":  verificationCtx.TurnConfig.CheckingImageTurn,
		}
	} else {
		machineStructure = make(map[string]interface{})
	}

	// Try to get additional context from the verification record
	// Note: This is a placeholder for additional fields that might be stored in the verification
	verificationSummary := make(map[string]interface{})
	
	// If there are additional fields in the verification context that we need,
	// they can be extracted here. For now, we'll create a basic summary.
	verificationSummary["verificationId"] = verificationCtx.VerificationId
	verificationSummary["verificationType"] = verificationCtx.VerificationType
	verificationSummary["vendingMachineId"] = verificationCtx.VendingMachineId
	
	// Build the historical context response
	result := map[string]interface{}{
		"previousVerificationId":      verificationId,
		"previousVerificationAt":      verificationCtx.VerificationAt,
		"previousVerificationStatus":  verificationCtx.Status,
		"hoursSinceLastVerification":  hoursSince,
		"machineStructure":            machineStructure,
		"verificationSummary":         verificationSummary,
		"vendingMachineId":            verificationCtx.VendingMachineId,
		"verificationType":            verificationCtx.VerificationType,
	}

	// Add request metadata if available
	if verificationCtx.RequestMetadata != nil {
		result["requestMetadata"] = map[string]interface{}{
			"requestId":         verificationCtx.RequestMetadata.RequestId,
			"requestTimestamp":  verificationCtx.RequestMetadata.RequestTimestamp,
			"processingStarted": verificationCtx.RequestMetadata.ProcessingStarted,
		}
	}

	// Add resource validation info if available
	if verificationCtx.ResourceValidation != nil {
		result["resourceValidation"] = map[string]interface{}{
			"layoutExists":         verificationCtx.ResourceValidation.LayoutExists,
			"referenceImageExists": verificationCtx.ResourceValidation.ReferenceImageExists,
			"checkingImageExists":  verificationCtx.ResourceValidation.CheckingImageExists,
			"validationTimestamp": verificationCtx.ResourceValidation.ValidationTimestamp,
		}
	}

	return result, nil
}

// UpdateVerificationStatus updates verification status in DynamoDB.
func (d *DBUtilsWrapper) UpdateVerificationStatus(ctx context.Context, verificationId, status string) error {
	return d.dbUtils.UpdateVerificationStatus(ctx, verificationId, status)
}

// FetchLayoutMetadataWithFallback attempts to fetch layout metadata with graceful fallback
func (d *DBUtilsWrapper) FetchLayoutMetadataWithFallback(ctx context.Context, layoutId int, layoutPrefix string) (map[string]interface{}, error) {
	// Try to fetch the complete layout metadata
	result, err := d.FetchLayoutMetadata(ctx, layoutId, layoutPrefix)
	if err != nil {
		// If the layout doesn't exist, return a minimal structure
		// This can happen if the layout hasn't been fully processed yet
		return map[string]interface{}{
			"layoutId":        layoutId,
			"layoutPrefix":    layoutPrefix,
			"status":          "partial",
			"error":           err.Error(),
			"machineStructure": map[string]interface{}{
				"rowCount":      6,  // Default values
				"columnsPerRow": 10, // Default values
				"rowOrder":      []string{"A", "B", "C", "D", "E", "F"},
				"columnOrder":   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			"productPositionMap": make(map[string]interface{}),
		}, nil
	}
	
	// Add status to indicate successful retrieval
	result["status"] = "complete"
	return result, nil
}

// ValidateLayoutExists checks if a layout exists before attempting to fetch it
func (d *DBUtilsWrapper) ValidateLayoutExists(ctx context.Context, layoutId int, layoutPrefix string) (bool, error) {
	return d.dbUtils.VerifyLayoutExists(ctx, layoutId, layoutPrefix)
}