package main

import (
	"context"
	"fmt"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
)

// DBWrapper wraps the shared dbutils package for this service's specific needs
type DBWrapper struct {
	dbUtils *dbutils.DynamoDBUtils
	logger  logger.Logger
}

// NewDBWrapper creates a new DBWrapper
func NewDBWrapper(dbUtils *dbutils.DynamoDBUtils, log logger.Logger) *DBWrapper {
	return &DBWrapper{
		dbUtils: dbUtils,
		logger:  log.WithFields(map[string]interface{}{
			"component": "db-wrapper",
		}),
	}
}

// QueryMostRecentVerificationByCheckingImage queries the CheckImageIndex to find verifications
// using the provided referenceImageUrl as the checking image
func (d *DBWrapper) QueryMostRecentVerificationByCheckingImage(ctx context.Context, imageURL string) (*VerificationRecord, error) {
	d.logger.Info("Finding previous verification", map[string]interface{}{
		"imageURL": imageURL,
	})

	// Use the shared library's function to find the previous verification
	// This function returns nil if no previous verification is found
	previousVerification, err := d.dbUtils.FindPreviousVerification(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	if previousVerification == nil {
		return nil, fmt.Errorf("no previous verification found for image: %s", imageURL)
	}

	// Convert schema.VerificationContext to our VerificationRecord type
	record := &VerificationRecord{
		VerificationID:     previousVerification.VerificationId,
		VerificationAt:     previousVerification.VerificationAt,
		VerificationType:   previousVerification.VerificationType,
		VendingMachineID:   previousVerification.VendingMachineId,
		CheckingImageURL:   previousVerification.CheckingImageUrl,
		ReferenceImageURL:  previousVerification.ReferenceImageUrl,
		VerificationStatus: previousVerification.Status,
		// Note: Other fields like MachineStructure, CheckingStatus, and VerificationSummary 
		// would need to be retrieved from a different location since they're not part of 
		// the basic schema.VerificationContext

		// For this initial implementation, we'll leave these as empty/default
		// In a full refactoring, we might need to adapt the findPreviousVerification function
		// or create a new one to retrieve the complete verification record
		MachineStructure:    MachineStructure{},
		CheckingStatus:      map[string]string{},
		VerificationSummary: VerificationSummary{},
	}

	return record, nil
}