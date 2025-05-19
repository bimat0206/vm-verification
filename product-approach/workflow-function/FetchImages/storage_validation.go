package main

import (
	"fmt"
	"workflow-function/shared/logger"
)

// validateHybridStorageIntegrity validates that both images have proper Base64 data with hybrid storage
func validateHybridStorageIntegrity(results ParallelFetchResults, logger logger.Logger) []error {
	var errors []error
	
	// Validate reference image
	if !results.ReferenceMeta.HasBase64Data() {
		err := fmt.Errorf("reference image missing Base64 data despite successful fetch (storage method: %s)", 
			results.ReferenceMeta.StorageMethod)
		errors = append(errors, err)
		logger.Error("Reference image validation failed", map[string]interface{}{
			"storageMethod": results.ReferenceMeta.StorageMethod,
			"hasInline":     results.ReferenceMeta.HasInlineData(),
			"hasS3":         results.ReferenceMeta.HasS3Storage(),
			"error":         err.Error(),
		})
	}
	
	// Validate checking image
	if !results.CheckingMeta.HasBase64Data() {
		err := fmt.Errorf("checking image missing Base64 data despite successful fetch (storage method: %s)", 
			results.CheckingMeta.StorageMethod)
		errors = append(errors, err)
		logger.Error("Checking image validation failed", map[string]interface{}{
			"storageMethod": results.CheckingMeta.StorageMethod,
			"hasInline":     results.CheckingMeta.HasInlineData(),
			"hasS3":         results.CheckingMeta.HasS3Storage(),
			"error":         err.Error(),
		})
	}
	
	// Validate storage method consistency
	if err := results.ReferenceMeta.ValidateStorageMethod(); err != nil {
		errors = append(errors, fmt.Errorf("reference image storage validation failed: %w", err))
		logger.Error("Reference image storage method validation failed", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	if err := results.CheckingMeta.ValidateStorageMethod(); err != nil {
		errors = append(errors, fmt.Errorf("checking image storage validation failed: %w", err))
		logger.Error("Checking image storage method validation failed", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	if len(errors) == 0 {
		logger.Info("Hybrid storage integrity validation passed", map[string]interface{}{
			"referenceStorage": results.ReferenceMeta.StorageMethod,
			"checkingStorage":  results.CheckingMeta.StorageMethod,
		})
	}
	
	return errors
}

// abs64 returns the absolute value of an int64
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
