package handler

import (
	"context"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// StorageManager handles S3 storage operations for responses
type StorageManager struct {
	s3  services.S3StateManager
	cfg config.Config
	log logger.Logger
}

// NewStorageManager creates a new instance of StorageManager
func NewStorageManager(s3 services.S3StateManager, cfg config.Config, log logger.Logger) *StorageManager {
	return &StorageManager{
		s3:  s3,
		cfg: cfg,
		log: log,
	}
}

// StorageResult contains the results of storage operations
type StorageResult struct {
	RawRef       models.S3Reference
	ProcessedRef models.S3Reference
	Duration     time.Duration
	Error        error
}

// StoreResponses stores raw and processed responses to S3
func (s *StorageManager) StoreResponses(ctx context.Context, verificationID string, resp *models.BedrockResponse) *StorageResult {
	startTime := time.Now()
	result := &StorageResult{}
	contextLogger := s.log.WithCorrelationId(verificationID)
	
	// Store raw response
	rawRef, err := s.s3.StoreRawResponse(ctx, verificationID, resp.Raw)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3, 
			"store raw response failed", true).
			WithContext("verification_id", verificationID).
			WithContext("response_size", len(resp.Raw))
		
		enrichedErr := errors.SetVerificationID(s3Err, verificationID)
		
		contextLogger.Warn("s3 raw-store warning", map[string]interface{}{
			"response_size_bytes": len(resp.Raw),
			"bucket":              s.cfg.AWS.S3Bucket,
		})
		
		result.Error = enrichedErr
		result.Duration = time.Since(startTime)
		return result
	}
	
	// Store processed analysis
	procRef, err := s.s3.StoreProcessedAnalysis(ctx, verificationID, resp.Processed)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3, 
			"store processed analysis failed", true).
			WithContext("verification_id", verificationID)
		
		enrichedErr := errors.SetVerificationID(s3Err, verificationID)
		
		contextLogger.Warn("s3 processed-store warning", map[string]interface{}{
			"bucket": s.cfg.AWS.S3Bucket,
		})
		
		result.Error = enrichedErr
		result.Duration = time.Since(startTime)
		return result
	}
	
	result.RawRef = rawRef
	result.ProcessedRef = procRef
	result.Duration = time.Since(startTime)
	
	return result
}

// GetStorageMetadata returns metadata for tracking storage operations
func (s *StorageManager) GetStorageMetadata(result *StorageResult, respSize int) map[string]interface{} {
	return map[string]interface{}{
		"s3_objects_created":   2,
		"raw_response_size":    respSize,
		"processed_ref_key":    result.ProcessedRef.Key,
		"raw_ref_key":          result.RawRef.Key,
	}
}