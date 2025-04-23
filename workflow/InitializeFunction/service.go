package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Custom error types for better error handling
var (
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidImageFormat   = errors.New("invalid image format")
	ErrResourceNotFound     = errors.New("resource not found")
	ErrDatabaseFailure      = errors.New("database operation failed")
)

// InitService handles the business logic for the verification initialization
type InitService struct {
	deps *Dependencies
	config ConfigVars
	s3Utils *S3Utils
	dbUtils *DynamoDBUtils
	logger Logger
}

// NewInitService creates a new instance of InitService
func NewInitService(deps *Dependencies, config ConfigVars) *InitService {
	s3Util := deps.GetS3Util()
	dbUtil := deps.GetDynamoUtil()
	
	// Set configuration in utilities
	dbUtil.SetConfig(config)
	s3Util.SetConfig(config)
	
	return &InitService{
		deps:    deps,
		config:  config,
		s3Utils: s3Util,
		dbUtils: dbUtil,
		logger:  deps.GetLogger(),
	}
}

// Process handles the entire verification initialization workflow
func (s *InitService) Process(ctx context.Context, request InitRequest) (*VerificationContext, error) {
	s.logger.Info("Starting verification initialization process", map[string]interface{}{
		"layoutId":        request.LayoutId,
		"layoutPrefix":    request.LayoutPrefix,
		"vendingMachineId": request.VendingMachineId,
	})

	// Step 1: Validate the request
	if err := s.validateRequest(request); err != nil {
		s.logger.Error("Request validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 2: Verify all resources exist
	if err := s.verifyResources(ctx, request); err != nil {
		s.logger.Error("Resource verification failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 3: Generate verification ID and context
	verificationContext, err := s.createVerificationContext(request)
	if err != nil {
		s.logger.Error("Failed to create verification context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 4: Store in DynamoDB
	if err := s.dbUtils.StoreVerificationRecord(ctx, verificationContext); err != nil {
		s.logger.Error("Failed to store verification record", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, err
	}

	s.logger.Info("Verification initialized successfully", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"status": verificationContext.Status,
	})

	return verificationContext, nil
}

// validateRequest validates the request parameters
func (s *InitService) validateRequest(request InitRequest) error {
	// Check required fields
	if request.ReferenceImageUrl == "" {
		return fmt.Errorf("%w: referenceImageUrl", ErrMissingRequiredField)
	}
	if request.CheckingImageUrl == "" {
		return fmt.Errorf("%w: checkingImageUrl", ErrMissingRequiredField)
	}
	if request.LayoutId == 0 {
		return fmt.Errorf("%w: layoutId", ErrMissingRequiredField)
	}
	if request.LayoutPrefix == "" {
		return fmt.Errorf("%w: layoutPrefix", ErrMissingRequiredField)
	}

	// Validate S3 URLs format
	refURL, checkURL, err := s.s3Utils.ParseS3URLs(request.ReferenceImageUrl, request.CheckingImageUrl)
	if err != nil {
		return err
	}

	// Validate image extensions
	if !s.s3Utils.IsValidImageExtension(refURL.Key) {
		return fmt.Errorf("%w: reference image must be PNG or JPEG", ErrInvalidImageFormat)
	}
	if !s.s3Utils.IsValidImageExtension(checkURL.Key) {
		return fmt.Errorf("%w: checking image must be PNG or JPEG", ErrInvalidImageFormat)
	}

	return nil
}

// verifyResources checks if all resources (images, layout) exist
func (s *InitService) verifyResources(ctx context.Context, request InitRequest) error {
	// Create error channels for concurrent verification
	errChan := make(chan error, 3)
	defer close(errChan)

	// Verify reference image existence in parallel
	go func() {
		err := s.s3Utils.ValidateImageExists(ctx, request.ReferenceImageUrl)
		if err != nil {
			errChan <- fmt.Errorf("reference image verification failed: %w", err)
			return
		}
		errChan <- nil
	}()

	// Verify checking image existence in parallel
	go func() {
		err := s.s3Utils.ValidateImageExists(ctx, request.CheckingImageUrl)
		if err != nil {
			errChan <- fmt.Errorf("checking image verification failed: %w", err)
			return
		}
		errChan <- nil
	}()

	// Verify layout exists in DynamoDB
	go func() {
		exists, err := s.dbUtils.VerifyLayoutExists(ctx, request.LayoutId, request.LayoutPrefix)
		if err != nil {
			errChan <- fmt.Errorf("error checking layout: %w", err)
			return
		}
		if !exists {
			errChan <- fmt.Errorf("layout with ID %d and prefix %s not found", 
				request.LayoutId, request.LayoutPrefix)
			return
		}
		errChan <- nil
	}()

	// Collect errors from all goroutines
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

// createVerificationContext creates the verification context with a unique ID
func (s *InitService) createVerificationContext(request InitRequest) (*VerificationContext, error) {
	// Generate a unique verification ID
	timestamp := time.Now().UTC()
	formattedTime := timestamp.Format("20060102150405")
	verificationId := fmt.Sprintf("%s%s", s.config.VerificationPrefix, formattedTime)
	
	// Add a random suffix for extra uniqueness
	randomId := uuid.New().String()[0:8]
	verificationId = fmt.Sprintf("%s%s", verificationId, randomId)

	// Format ISO timestamp
	isoTimestamp := timestamp.Format(time.RFC3339)

	// Create verification context
	verificationContext := &VerificationContext{
		VerificationId:    verificationId,
		VerificationAt:    isoTimestamp,
		Status:            StatusInitialized,
		VendingMachineId:  request.VendingMachineId,
		LayoutId:          request.LayoutId,
		LayoutPrefix:      request.LayoutPrefix,
		ReferenceImageUrl: request.ReferenceImageUrl,
		CheckingImageUrl:  request.CheckingImageUrl,
	}

	// Handle conversation configuration
	if request.ConversationConfig != nil {
		verificationContext.ConversationType = request.ConversationConfig.Type
		verificationContext.TurnConfig = &TurnConfig{
			MaxTurns:           request.ConversationConfig.MaxTurns,
			ReferenceImageTurn: 1,
			CheckingImageTurn:  2,
		}
	} else {
		// Default to two-turn configuration
		verificationContext.ConversationType = "two-turn"
		verificationContext.TurnConfig = &TurnConfig{
			MaxTurns:           2,
			ReferenceImageTurn: 1,
			CheckingImageTurn:  2,
		}
	}

	// Initialize turn timestamps
	verificationContext.TurnTimestamps = &TurnTimestamps{
		Initialized: isoTimestamp,
		Turn1:       nil,
		Turn2:       nil,
		Completed:   nil,
	}

	// Set request metadata
	requestTimestamp := request.RequestTimestamp
	if requestTimestamp == "" {
		requestTimestamp = isoTimestamp
	}
	
	verificationContext.RequestMetadata = &RequestMetadata{
		RequestId:         request.RequestId,
		RequestTimestamp:  requestTimestamp,
		ProcessingStarted: isoTimestamp,
	}

	return verificationContext, nil
}