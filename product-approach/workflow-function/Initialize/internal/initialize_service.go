package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid" 
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// Custom error types
var (
	ErrMissingRequiredField   = errors.New("missing required field")
	ErrInvalidVerificationType = errors.New("invalid verification type")
	ErrSameReferenceAndChecking = errors.New("reference and checking images cannot be the same")
)

// InitializeService handles the logic for initializing verifications
type InitializeService struct {
	config             Config
	logger             logger.Logger
	verificationRepo   *VerificationRepository
	layoutRepo         *LayoutRepository
	s3Validator        *S3Validator
	s3URLParser        *S3URLParser
	s3StateManager     *S3StateManagerWrapper
}

// NewInitializeService creates a new instance of InitializeService
func NewInitializeService(ctx context.Context, awsConfig aws.Config, cfg Config) (*InitializeService, error) {
	// Create logger
	log := logger.New("verification", "InitializeFunction")
	
	// Create S3 client
	s3Client := NewS3Client(awsConfig, cfg, log)
	s3URLParser := NewS3URLParser(cfg, log)
	s3Validator := NewS3Validator(s3Client, s3URLParser, log)
	
	// Create S3 state manager
	s3StateManager, err := NewS3StateManager(ctx, awsConfig, cfg, log)
	if err != nil {
		log.Error("Failed to create S3 state manager", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create S3 state manager: %w", err)
	}
	
	// Create DynamoDB client
	dbClient := NewDynamoDBClient(awsConfig, cfg, log)
	verificationRepo := NewVerificationRepository(dbClient, cfg, log)
	layoutRepo := NewLayoutRepository(dbClient, cfg, log)
	
	return &InitializeService{
		config:             cfg,
		logger:             log,
		verificationRepo:   verificationRepo,
		layoutRepo:         layoutRepo,
		s3Validator:        s3Validator,
		s3URLParser:        s3URLParser,
		s3StateManager:     s3StateManager,
	}, nil
}

// Process handles the verification initialization process
func (s *InitializeService) Process(ctx context.Context, request ProcessRequest) (*ExtendedEnvelope, error) {
	// Create a verification context from the request
	verificationContext, err := s.createVerificationContext(request)
	if err != nil {
		s.logger.Error("Failed to create verification context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create verification context: %w", err)
	}
	
	// Validate the verification context
	if err := s.validateVerificationContext(verificationContext); err != nil {
		s.logger.Error("Verification context validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Verify resources (S3 images, layout metadata)
	resourceValidation, err := s.verifyResources(ctx, verificationContext)
	if err != nil {
		s.logger.Error("Resource verification failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		// Create standard error structure
		errorInfo := &schema.ErrorInfo{
			Code:      "RESOURCE_VALIDATION_FAILED",
			Message:   err.Error(),
			Timestamp: schema.FormatISO8601(),
			Details: map[string]interface{}{
				"resourceValidation": resourceValidation,
			},
		}
		
		verificationContext.Error = errorInfo
		verificationContext.Status = schema.StatusInitializationFailed
		
		// Still try to create the S3 state structure with error info
		basicEnvelope, stateErr := s.createStateStructure(ctx, verificationContext)
		if stateErr != nil {
			// Log but return the original error
			s.logger.Error("Failed to create state structure after resource validation failure", map[string]interface{}{
				"error": stateErr.Error(),
			})
		} else if basicEnvelope != nil {
			// Create extended envelope with verification context
			envelope := NewExtendedEnvelope(basicEnvelope)
			envelope.VerificationContext = verificationContext
			
			// Return the envelope with error status
			envelope.Envelope.SetStatus(schema.StatusInitializationFailed)
			return envelope, err
		}
		
		return nil, err
	}
	
	// Update the verification context with resource validation
	verificationContext.ResourceValidation = resourceValidation
	verificationContext.Status = schema.StatusVerificationInitialized
	
	// Create S3 state structure and save initialization context
	basicEnvelope, err := s.createStateStructure(ctx, verificationContext)
	if err != nil {
		s.logger.Error("Failed to create state structure", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to create state structure: %w", err)
	}
	
	// Store minimal record in DynamoDB with S3 reference
	err = s.verificationRepo.StoreMinimalRecord(ctx, verificationContext, basicEnvelope.References["processing_initialization"])
	if err != nil {
		s.logger.Error("Failed to store verification record", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to store verification record: %w", err)
	}
	
	// Set response summary with information Step Functions will look for
	basicEnvelope.AddSummary("verificationType", verificationContext.VerificationType)
	basicEnvelope.AddSummary("resourcesValidated", []string{"referenceImage", "checkingImage", "layoutMetadata"})
	basicEnvelope.AddSummary("contextEstablished", true)
	basicEnvelope.AddSummary("stateStructureCreated", true)
	
	// Create extended envelope with verification context for Step Functions compatibility
	envelope := NewExtendedEnvelope(basicEnvelope)
	envelope.VerificationContext = verificationContext
	
	s.logger.Info("Verification initialized successfully", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"status": verificationContext.Status,
		"s3StateBucket": s.config.StateBucket,
	})
	
	return envelope, nil
}

// createStateStructure creates the S3 state structure and saves the initialization context
func (s *InitializeService) createStateStructure(ctx context.Context, verificationContext *schema.VerificationContext) (*s3state.Envelope, error) {
	// Create the folder structure
	err := s.s3StateManager.InitializeStateStructure(ctx, verificationContext.VerificationId)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize state structure: %w", err)
	}
	
	// Create envelope
	envelope := s.s3StateManager.CreateEnvelope(verificationContext.VerificationId)
	envelope.SetStatus(verificationContext.Status)
	
	// Save verification context to S3
	err = s.s3StateManager.SaveContext(envelope, verificationContext)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification context: %w", err)
	}
	
	return envelope, nil
}

// validateVerificationContext validates the verification context
func (s *InitializeService) validateVerificationContext(verificationContext *schema.VerificationContext) error {
	// Check verification type
	if verificationContext.VerificationType == "" {
		return fmt.Errorf("%w: verificationType", ErrMissingRequiredField)
	}
	
	// Verify verification type is supported
	if verificationContext.VerificationType != schema.VerificationTypeLayoutVsChecking && 
	   verificationContext.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return fmt.Errorf("%w: %s", ErrInvalidVerificationType, verificationContext.VerificationType)
	}

	// Check required fields for all verification types
	if verificationContext.ReferenceImageUrl == "" {
		return fmt.Errorf("%w: referenceImageUrl", ErrMissingRequiredField)
	}
	if verificationContext.CheckingImageUrl == "" {
		return fmt.Errorf("%w: checkingImageUrl", ErrMissingRequiredField)
	}
	
	// Check required fields specific to verification type
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// Layout vs. Checking requires layoutId and layoutPrefix
		if verificationContext.LayoutId == 0 {
			return fmt.Errorf("%w: layoutId required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
		}
		if verificationContext.LayoutPrefix == "" {
			return fmt.Errorf("%w: layoutPrefix required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
		}
	}

	// Validate S3 URLs format
	_, _, err := s.s3URLParser.ParseS3URLs(verificationContext.ReferenceImageUrl, verificationContext.CheckingImageUrl)
	if err != nil {
		return err
	}

	// If reference and checking images are the same, reject
	if verificationContext.ReferenceImageUrl == verificationContext.CheckingImageUrl {
		return ErrSameReferenceAndChecking
	}

	return nil
}

// verifyResources checks if all resources (images, layout) exist
func (s *InitializeService) verifyResources(ctx context.Context, verificationContext *schema.VerificationContext) (*schema.ResourceValidation, error) {
	// Create validation result
	resourceValidation := &schema.ResourceValidation{
		ValidationTimestamp: schema.FormatISO8601(),
	}
	
	// Calculate max image size based on lambda payload limitations
	// Typically Lambda has a 6MB payload limit, use a safe value below that
	const maxImageSize = 5 * 1024 * 1024 // 5MB should be safe for Lambda payloads
	
	// Validate both images
	imageURLs := []string{verificationContext.ReferenceImageUrl, verificationContext.CheckingImageUrl}
	validationResult, err := s.s3Validator.ValidateImagesInParallel(ctx, imageURLs, maxImageSize)
	if err != nil {
		return resourceValidation, err
	}
	
	// Copy validation results
	resourceValidation.ReferenceImageExists = validationResult.ReferenceImageExists
	resourceValidation.CheckingImageExists = validationResult.CheckingImageExists
	
	// Verify layout exists in DynamoDB only for LAYOUT_VS_CHECKING
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		exists, err := s.layoutRepo.VerifyLayoutExists(ctx, verificationContext.LayoutId, verificationContext.LayoutPrefix)
		if err != nil {
			return resourceValidation, fmt.Errorf("error checking layout: %w", err)
		}
		if !exists {
			return resourceValidation, fmt.Errorf("layout with ID %d and prefix %s not found", 
				verificationContext.LayoutId, verificationContext.LayoutPrefix)
		}
		resourceValidation.LayoutExists = true
	}
	
	return resourceValidation, nil
}

// generateVerificationId creates a unique verification ID
func (s *InitializeService) generateVerificationId() string {
	// Generate a unique verification ID: verif-<timestamp>-<4-char random>
	now := time.Now().UTC()
	ts := now.Format("20060102150405")                // e.g. "20250421153025"
	randomSuffix := uuid.New().String()[0:4]          // e.g. "a1b4"
	base := fmt.Sprintf("%s%s", s.config.VerificationPrefix, ts)
	return fmt.Sprintf("%s-%s", base, randomSuffix)
}

// Logger returns the service logger
func (s *InitializeService) Logger() logger.Logger {
	return s.logger
}

// createVerificationContext creates a new verification context from the request
func (s *InitializeService) createVerificationContext(request ProcessRequest) (*schema.VerificationContext, error) {
	var verificationContext *schema.VerificationContext
	
	// If using the standardized schema format with VerificationContext
	if request.SchemaVersion != "" && request.VerificationContext != nil {
		// Try to use the provided verification context
		if vc, ok := request.VerificationContext.(*schema.VerificationContext); ok {
			verificationContext = vc
			
			// Ensure verificationId is set
			if verificationContext.VerificationId == "" {
				verificationContext.VerificationId = s.generateVerificationId()
			}
			
			// Ensure verificationAt is set (critical for DynamoDB)
			if verificationContext.VerificationAt == "" {
				now := time.Now().UTC()
				verificationContext.VerificationAt = now.Format(time.RFC3339)
				s.logger.Info("Setting missing verificationAt", map[string]interface{}{
					"verificationAt": verificationContext.VerificationAt,
				})
			}
			
			// Store schema version in log for tracking
			s.logger.Info("Using schema version", map[string]interface{}{
				"schemaVersion": request.SchemaVersion,
			})
			
			s.logger.Info("Using provided verification context", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
				"verificationAt": verificationContext.VerificationAt,
				"schemaVersion": request.SchemaVersion,
			})
			
			return verificationContext, nil
		}
		
		// If we can't use the provided verification context, log an error
		s.logger.Error("Invalid verification context type", map[string]interface{}{
			"type": fmt.Sprintf("%T", request.VerificationContext),
		})
		return nil, fmt.Errorf("invalid verification context type")
	}
	
	// Create a new verification context from the request parameters
	// Generate a unique verification ID
	verificationId := s.generateVerificationId()

	// Format ISO timestamp
	now := time.Now().UTC()
	isoTimestamp := now.Format(time.RFC3339)

	// Create verification context
	verificationContext = &schema.VerificationContext{
		VerificationId:      verificationId,
		VerificationAt:      isoTimestamp,
		Status:              schema.StatusVerificationInitialized,
		VerificationType:    request.VerificationType,
		VendingMachineId:    request.VendingMachineId,
		ReferenceImageUrl:   request.ReferenceImageUrl,
		CheckingImageUrl:    request.CheckingImageUrl,
		NotificationEnabled: request.NotificationEnabled,
	}

	// Set layout details for LAYOUT_VS_CHECKING
	if request.VerificationType == schema.VerificationTypeLayoutVsChecking {
		verificationContext.LayoutId = request.LayoutId
		verificationContext.LayoutPrefix = request.LayoutPrefix
	}

	// Set previousVerificationId for PREVIOUS_VS_CURRENT
	if request.VerificationType == schema.VerificationTypePreviousVsCurrent {
		verificationContext.PreviousVerificationId = request.PreviousVerificationId
		
		// If previous verification ID is not provided, try to find it
		if verificationContext.PreviousVerificationId == "" {
			// Use the context from the Process function
			previousVerification, err := s.verificationRepo.FindPreviousVerification(context.Background(), request.ReferenceImageUrl)
			if err == nil && previousVerification != nil {
				verificationContext.PreviousVerificationId = previousVerification.VerificationId
				s.logger.Info("Found previous verification", map[string]interface{}{
					"previousVerificationId": verificationContext.PreviousVerificationId,
				})
			}
		}
	}

	// Initialize turn timestamps
	verificationContext.TurnTimestamps = &schema.TurnTimestamps{
		Initialized: isoTimestamp,
	}

	// Set conversation configuration with defaults
	verificationContext.ConversationType = "two-turn" // Default
	verificationContext.TurnConfig = &schema.TurnConfig{
		MaxTurns:           2, // Default
		ReferenceImageTurn: 1,
		CheckingImageTurn:  2,
	}
	
	// Override defaults if provided
	if request.ConversationConfig.Type != "" {
		verificationContext.ConversationType = request.ConversationConfig.Type
	}
	if request.ConversationConfig.MaxTurns > 0 {
		verificationContext.TurnConfig.MaxTurns = request.ConversationConfig.MaxTurns
	}

	// Set request metadata
	requestTimestamp := request.RequestTimestamp
	if requestTimestamp == "" {
		requestTimestamp = isoTimestamp
	}
	
	verificationContext.RequestMetadata = &schema.RequestMetadata{
		RequestId:         request.RequestId,
		RequestTimestamp:  requestTimestamp,
		ProcessingStarted: isoTimestamp,
	}

	s.logger.Debug("Created verification context", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"status": verificationContext.Status,
	})

	return verificationContext, nil
}