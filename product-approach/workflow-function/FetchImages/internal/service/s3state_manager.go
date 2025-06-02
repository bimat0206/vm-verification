// Package service provides business logic implementations for the FetchImages function
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"workflow-function/FetchImages/internal/config"
	"workflow-function/FetchImages/internal/models"
)

// S3StateManager wraps the s3state.Manager to provide additional functionality
type S3StateManager struct {
	manager    s3state.Manager
	logger     logger.Logger
	config     config.Config
	s3Client   *s3.Client
}

// NewS3StateManager creates a new S3 state manager
func NewS3StateManager(ctx context.Context, awsConfig aws.Config, cfg config.Config, log logger.Logger) (*S3StateManager, error) {
	if cfg.StateBucket == "" {
		return nil, fmt.Errorf("STATE_BUCKET environment variable is not set")
	}

	// Create s3state.Manager
	manager, err := s3state.New(cfg.StateBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 state manager: %w", err)
	}
	
	// Create S3 client for additional operations
	s3Client := s3.NewFromConfig(awsConfig)

	return &S3StateManager{
		manager:    manager,
		logger:     log.WithFields(map[string]interface{}{"component": "S3StateManager"}),
		config:     cfg,
		s3Client:   s3Client,
	}, nil
}

// LoadVerificationContext loads the verification context from the S3 state
func (s *S3StateManager) LoadVerificationContext(envelope *s3state.Envelope) (interface{}, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	
	// Find the initialization reference in the envelope
	initRef := envelope.GetReference("processing_initialization")
	if initRef == nil {
		return nil, fmt.Errorf("initialization reference not found in envelope")
	}
	
	// Create a map to hold the loaded context
	var context interface{}
	
	// Load the context from S3
	err := s.manager.RetrieveJSON(initRef, &context)
	if err != nil {
		return nil, fmt.Errorf("failed to load verification context: %w", err)
	}
	
	return context, nil
}

// StoreImageMetadata stores image metadata in the images category
func (s *S3StateManager) StoreImageMetadata(envelope *s3state.Envelope, metadata interface{}) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	
	// Store metadata in images category
	err := s.manager.SaveToEnvelope(
		envelope,
		s3state.CategoryImages,
		s3state.ImageMetadataFile,
		metadata,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store image metadata: %w", err)
	}
	
	return nil
}

// StoreLayoutMetadata stores layout metadata in the processing category
func (s *S3StateManager) StoreLayoutMetadata(envelope *s3state.Envelope, layoutMetadata interface{}) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	
	// Store layout metadata in processing category
	err := s.manager.SaveToEnvelope(
		envelope,
		s3state.CategoryProcessing,
		s3state.LayoutMetadataFile,
		layoutMetadata,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store layout metadata: %w", err)
	}
	
	return nil
}

// StoreHistoricalContext stores historical context in the processing category
func (s *S3StateManager) StoreHistoricalContext(envelope *s3state.Envelope, historicalContext interface{}) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	
	// Store historical context in processing category
	err := s.manager.SaveToEnvelope(
		envelope,
		s3state.CategoryProcessing,
		s3state.HistoricalContextFile,
		historicalContext,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store historical context: %w", err)
	}
	
	return nil
}

// UpdateEnvelopeStatus updates the status in the envelope
func (s *S3StateManager) UpdateEnvelopeStatus(envelope *s3state.Envelope, status string) {
	if envelope == nil {
		s.logger.Error("Cannot update status for nil envelope", map[string]interface{}{
			"status": status,
		})
		return
	}
	
	envelope.SetStatus(status)
	
	s.logger.Info("Updated envelope status", map[string]interface{}{
		"verificationId": envelope.VerificationID,
		"status":         status,
	})
}

// AddSummary adds a summary field to the envelope
func (s *S3StateManager) AddSummary(envelope *s3state.Envelope, key string, value interface{}) {
	if envelope == nil {
		s.logger.Error("Cannot add summary to nil envelope", map[string]interface{}{
			"key":   key,
			"value": value,
		})
		return
	}
	
	envelope.AddSummary(key, value)
}

// LoadEnvelope loads an envelope from the input with enhanced error handling
func (s *S3StateManager) LoadEnvelope(input interface{}) (*s3state.Envelope, error) {
	// Check if input is a FetchImagesRequest, and convert it to an Envelope
	if req, ok := input.(*models.FetchImagesRequest); ok {
		envelope := &s3state.Envelope{
			VerificationID: req.VerificationId,
			References:     req.S3References,
			Status:         req.Status,
			Summary:        make(map[string]interface{}),
		}

		// Check for inherited errors from previous workflow steps
		s.checkForInheritedErrors(envelope, req)

		// If the conversion succeeded, return the envelope
		if envelope.VerificationID != "" {
			s.logger.Info("Created envelope from FetchImagesRequest", map[string]interface{}{
				"verificationId": envelope.VerificationID,
				"status":        envelope.Status,
				"referenceCount": len(envelope.References),
				"hasInheritedErrors": envelope.Summary["hasInheritedErrors"],
			})
			return envelope, nil
		}
	}

	// For other input types, use the standard method
	envelope, err := s3state.LoadEnvelope(input)
	if err != nil {
		return nil, fmt.Errorf("failed to load envelope: %w", err)
	}

	if envelope.VerificationID == "" {
		return nil, fmt.Errorf("envelope is missing verification ID")
	}

	s.logger.Info("Loaded envelope", map[string]interface{}{
		"verificationId": envelope.VerificationID,
		"status":         envelope.Status,
		"referenceCount": len(envelope.References),
	})

	return envelope, nil
}

// checkForInheritedErrors examines the request for errors from previous workflow steps
func (s *S3StateManager) checkForInheritedErrors(envelope *s3state.Envelope, req *models.FetchImagesRequest) {
	hasInheritedErrors := false
	inheritedErrorSources := []string{}

	// Check if the status indicates a previous error
	if strings.Contains(strings.ToUpper(req.Status), "ERROR") ||
	   strings.Contains(strings.ToUpper(req.Status), "FAILED") {
		hasInheritedErrors = true
		inheritedErrorSources = append(inheritedErrorSources, "status_field")
		s.logger.Warn("Detected error status from previous step", map[string]interface{}{
			"status": req.Status,
			"verificationId": req.VerificationId,
		})
	}

	// Check for error references in S3References
	if req.S3References != nil {
		for key, ref := range req.S3References {
			if strings.Contains(strings.ToLower(key), "error") {
				hasInheritedErrors = true
				inheritedErrorSources = append(inheritedErrorSources, fmt.Sprintf("s3_reference_%s", key))
				s.logger.Warn("Detected error reference from previous step", map[string]interface{}{
					"referenceKey": key,
					"bucket": ref.Bucket,
					"key": ref.Key,
					"verificationId": req.VerificationId,
				})
			}
		}
	}

	// Store inherited error information in summary
	envelope.Summary["hasInheritedErrors"] = hasInheritedErrors
	if hasInheritedErrors {
		envelope.Summary["inheritedErrorSources"] = inheritedErrorSources
		envelope.Summary["inheritedErrorDetectedAt"] = time.Now().UTC().Format(time.RFC3339)

		s.logger.Info("Inherited errors detected and catalogued", map[string]interface{}{
			"verificationId": req.VerificationId,
			"errorSources": inheritedErrorSources,
		})
	}
}

// GetTimeBasedKey generates a time-based key to ensure uniqueness
func (s *S3StateManager) GetTimeBasedKey(baseName string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", baseName, timestamp)
}

// Manager returns the underlying s3state.Manager
func (s *S3StateManager) Manager() s3state.Manager {
	return s.manager
}