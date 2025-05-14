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

// FetchLayoutMetadata retrieves layout metadata from DynamoDB.
func (d *DBUtilsWrapper) FetchLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (map[string]interface{}, error) {
	// Check if layout exists
	exists, err := d.dbUtils.VerifyLayoutExists(ctx, layoutId, layoutPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to verify layout existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("layout metadata not found")
	}

	// For now, we'll just return a basic map, but in a real implementation
	// we would query the actual data from DynamoDB using the shared package
	result := map[string]interface{}{
		"layoutId":      layoutId,
		"layoutPrefix":  layoutPrefix,
		"retrievedWith": "shared dbutils package",
	}

	// Note: In a real implementation, we would be using something like:
	// dynamodbResult, err := d.dbUtils.GetLayoutMetadata(ctx, layoutId, layoutPrefix)
	// and then converting that result to the appropriate format

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
	
	// Convert to a map for the response
	result := map[string]interface{}{
		"previousVerificationId":      verificationId,
		"previousVerificationAt":      verificationCtx.VerificationAt,
		"previousVerificationStatus":  verificationCtx.Status,
		"hoursSinceLastVerification":  hoursSince,
		"machineStructure":            verificationCtx.TurnConfig,
		"verificationSummary":         make(map[string]interface{}),
		"retrievedWith":               "shared dbutils package",
	}

	return result, nil
}

// UpdateVerificationStatus updates verification status in DynamoDB.
func (d *DBUtilsWrapper) UpdateVerificationStatus(ctx context.Context, verificationId, status string) error {
	return d.dbUtils.UpdateVerificationStatus(ctx, verificationId, status)
}