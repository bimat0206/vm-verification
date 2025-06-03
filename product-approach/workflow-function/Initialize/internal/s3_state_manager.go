package internal

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
)

// S3StateManagerWrapper wraps the s3state.Manager to provide additional functionality
type S3StateManagerWrapper struct {
	manager    s3state.Manager
	logger     logger.Logger
	config     Config
	s3Client   *S3Client
}

// NewS3StateManager creates a new S3 state manager wrapper
func NewS3StateManager(ctx context.Context, awsConfig aws.Config, cfg Config, log logger.Logger) (*S3StateManagerWrapper, error) {
	if cfg.StateBucket == "" {
		return nil, fmt.Errorf("STATE_BUCKET environment variable is not set")
	}

	// Create s3state.Manager
	manager, err := s3state.New(cfg.StateBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 state manager: %w", err)
	}
	
	// Create S3 client for additional operations
	s3Client := NewS3Client(awsConfig, cfg, log)

	return &S3StateManagerWrapper{
		manager:    manager,
		logger:     log.WithFields(map[string]interface{}{"component": "S3StateManager"}),
		config:     cfg,
		s3Client:   s3Client,
	}, nil
}

// getDateBasedPath returns the date-based path for a verification ID
func (s *S3StateManagerWrapper) getDateBasedPath(verificationId string) string {
	// Extract date components from the current time
	now := time.Now().UTC()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	
	// Return the path structure: {year}/{month}/{day}/{verificationId}
	return fmt.Sprintf("%s/%s/%s/%s", year, month, day, verificationId)
}

// InitializeStateStructure creates the initial state folder structure for a verification
func (s *S3StateManagerWrapper) InitializeStateStructure(ctx context.Context, verificationId string) error {
	s.logger.Info("Initializing S3 state structure", map[string]interface{}{
		"verificationId": verificationId,
		"bucket":         s.config.StateBucket,
	})

	// Get the date-based path
	basePath := s.getDateBasedPath(verificationId)
	
	// Create folder structure using empty objects with trailing slashes
	folders := []string{
		fmt.Sprintf("%s/", basePath),
		fmt.Sprintf("%s/%s/", basePath, s3state.CategoryImages),
		fmt.Sprintf("%s/%s/", basePath, s3state.CategoryPrompts),
		fmt.Sprintf("%s/%s/", basePath, s3state.CategoryResponses),
		fmt.Sprintf("%s/%s/", basePath, s3state.CategoryProcessing),
	}

	// Create folders in parallel
	errorChan := make(chan error, len(folders))
	for _, folder := range folders {
		go func(folderKey string) {
			_, err := s.s3Client.Client().PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(s.config.StateBucket),
				Key:         aws.String(folderKey),
				ContentType: aws.String("application/x-directory"),
				Body:        nil, // Empty body for directory marker
			})
			errorChan <- err
		}(folder)
	}

	// Collect errors
	var errors []error
	for i := 0; i < len(folders); i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to create %d folder(s): %v", len(errors), errors[0])
	}

	s.logger.Info("S3 state structure initialized successfully", map[string]interface{}{
		"verificationId": verificationId,
		"folderCount":    len(folders),
	})

	return nil
}

// CreateEnvelope creates a new S3 state envelope for a verification ID
func (s *S3StateManagerWrapper) CreateEnvelope(verificationId string) *s3state.Envelope {
	envelope := s3state.NewEnvelope(verificationId)
	
	// Log path structure being used
	s.logger.Info("Created envelope with date-based path", map[string]interface{}{
		"verificationId": verificationId,
		"basePath": s.getDateBasedPath(verificationId),
	})
	
	return envelope
}

// SaveContext stores the verification context in the S3 state
func (s *S3StateManagerWrapper) SaveContext(envelope *s3state.Envelope, data interface{}) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	
	// Get the date-based path for the verification ID
	verificationId := envelope.VerificationID
	datePath := s.getDateBasedPath(verificationId)
	
	// Create the key with the date-based structure
	category := s3state.CategoryProcessing
	filename := s3state.InitializationFile
	key := fmt.Sprintf("%s/%s", category, filename)
	
	// Store the data with the full path
	ref, err := s.manager.StoreJSON(datePath, key, data)
	if err != nil {
		return err
	}
	
	// Add reference to envelope
	refKey := fmt.Sprintf("%s_%s", category, strings.TrimSuffix(filename, ".json"))
	envelope.AddReference(refKey, ref)
	
	return nil
}

// Manager returns the underlying s3state.Manager
func (s *S3StateManagerWrapper) Manager() s3state.Manager {
	return s.manager
}
