package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
)

// Custom error types
var (
	ErrMissingRequiredField    = errors.New("missing required field")
	ErrInvalidVerificationType  = errors.New("invalid verification type")
	ErrSameReferenceAndChecking = errors.New("reference and checking images cannot be the same")
)

// InitializeService handles the logic for initializing verifications
type InitializeService struct {
	config           Config
	logger           logger.Logger
	verificationRepo *VerificationRepository
	layoutRepo       *LayoutRepository
	s3Validator      *S3Validator
	s3URLParser      *S3URLParser
	s3StateManager   *S3StateManagerWrapper
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
		config:           cfg,
		logger:           log,
		verificationRepo: verificationRepo,
		layoutRepo:       layoutRepo,
		s3Validator:      s3Validator,
		s3URLParser:      s3URLParser,
		s3StateManager:   s3StateManager,
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

	// Attempt to Lookup LayoutId and LayoutPrefix if missing for LAYOUT_VS_CHECKING
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking &&
		(verificationContext.LayoutId == 0 || verificationContext.LayoutPrefix == "") &&
		verificationContext.ReferenceImageUrl != "" {

		s.logger.Info("LayoutId/LayoutPrefix not fully provided for LAYOUT_VS_CHECKING with referenceImageUrl. Attempting lookup.", map[string]interface{}{
			"verificationId":    verificationContext.VerificationId,
			"referenceImageUrl": verificationContext.ReferenceImageUrl,
			"currentLayoutId":   verificationContext.LayoutId,
			"currentLayoutPrefix": verificationContext.LayoutPrefix,
		})

		retrievedLayout, lookupErr := s.layoutRepo.GetLayoutByReferenceImage(ctx, verificationContext.ReferenceImageUrl)
		if lookupErr != nil {
			s.logger.Error("Failed to retrieve layout metadata from referenceImageUrl during initialization", map[string]interface{}{
				"verificationId":    verificationContext.VerificationId,
				"referenceImageUrl": verificationContext.ReferenceImageUrl,
				"error":             lookupErr.Error(),
			})
			
			errorInfo := &schema.ErrorInfo{
				Code:      "LAYOUT_LOOKUP_FAILED",
				Message:   fmt.Sprintf("cannot retrieve layout metadata from referenceImageUrl: %v", lookupErr),
				Timestamp: schema.FormatISO8601(),
				Details: map[string]interface{}{
					"referenceImageUrl": verificationContext.ReferenceImageUrl,
					"lookupError":       lookupErr.Error(),
				},
			}
			
			verificationContext.Error = errorInfo
			verificationContext.Status = schema.StatusInitializationFailed
			
			basicEnvelope, stateErr := s.createStateStructure(ctx, verificationContext)
			if stateErr != nil {
				s.logger.Error("Failed to create state structure after layout lookup failure", map[string]interface{}{
					"error": stateErr.Error(),
					"verificationId": verificationContext.VerificationId,
				})
				return nil, fmt.Errorf("layout lookup failed and unable to save state: %w", lookupErr)
			}
			
			envelope := NewExtendedEnvelope(basicEnvelope)
			envelope.VerificationContext = verificationContext
			envelope.Envelope.SetStatus(schema.StatusInitializationFailed)
			return envelope, fmt.Errorf("layout lookup failed: %w", lookupErr)
		}
		
		if retrievedLayout == nil {
			s.logger.Error("Layout lookup returned nil result for referenceImageUrl", map[string]interface{}{
				"verificationId":    verificationContext.VerificationId,
				"referenceImageUrl": verificationContext.ReferenceImageUrl,
			})
			
			errorInfo := &schema.ErrorInfo{
				Code:      "LAYOUT_NOT_FOUND",
				Message:   "layout metadata not found for the provided referenceImageUrl",
				Timestamp: schema.FormatISO8601(),
				Details: map[string]interface{}{
					"referenceImageUrl": verificationContext.ReferenceImageUrl,
				},
			}
			
			verificationContext.Error = errorInfo
			verificationContext.Status = schema.StatusInitializationFailed
			
			basicEnvelope, stateErr := s.createStateStructure(ctx, verificationContext)
			if stateErr != nil {
				s.logger.Error("Failed to create state structure after layout not found", map[string]interface{}{
					"error": stateErr.Error(),
					"verificationId": verificationContext.VerificationId,
				})
				return nil, errors.New("layout not found and unable to save state")
			}
			
			envelope := NewExtendedEnvelope(basicEnvelope)
			envelope.VerificationContext = verificationContext
			envelope.Envelope.SetStatus(schema.StatusInitializationFailed)
			return envelope, errors.New("layout metadata not found for referenceImageUrl")
		}
		
		if retrievedLayout != nil {
			s.logger.Info("Successfully retrieved layout by referenceImageUrl", map[string]interface{}{
				"verificationId":        verificationContext.VerificationId,
				"retrievedLayoutId":     retrievedLayout.LayoutId,
				"retrievedLayoutPrefix": retrievedLayout.LayoutPrefix,
			})
			
			// Update verification context with retrieved layout metadata
			if verificationContext.LayoutId == 0 && retrievedLayout.LayoutId != 0 {
				verificationContext.LayoutId = retrievedLayout.LayoutId
			}
			if verificationContext.LayoutPrefix == "" && retrievedLayout.LayoutPrefix != "" {
				verificationContext.LayoutPrefix = retrievedLayout.LayoutPrefix
			}
			
			// Verify that we now have both required fields
			if verificationContext.LayoutId == 0 || verificationContext.LayoutPrefix == "" {
				s.logger.Error("Layout lookup succeeded but layoutId or layoutPrefix is still missing", map[string]interface{}{
					"verificationId":        verificationContext.VerificationId,
					"retrievedLayoutId":     retrievedLayout.LayoutId,
					"retrievedLayoutPrefix": retrievedLayout.LayoutPrefix,
					"finalLayoutId":        verificationContext.LayoutId,
					"finalLayoutPrefix":    verificationContext.LayoutPrefix,
				})
				
				errorInfo := &schema.ErrorInfo{
					Code:      "INVALID_LAYOUT_METADATA",
					Message:   "retrieved layout metadata is incomplete (missing layoutId or layoutPrefix)",
					Timestamp: schema.FormatISO8601(),
					Details: map[string]interface{}{
						"referenceImageUrl":     verificationContext.ReferenceImageUrl,
						"retrievedLayoutId":     retrievedLayout.LayoutId,
						"retrievedLayoutPrefix": retrievedLayout.LayoutPrefix,
					},
				}
				
				verificationContext.Error = errorInfo
				verificationContext.Status = schema.StatusInitializationFailed
				
				basicEnvelope, stateErr := s.createStateStructure(ctx, verificationContext)
				if stateErr != nil {
					s.logger.Error("Failed to create state structure after invalid layout metadata", map[string]interface{}{
						"error": stateErr.Error(),
						"verificationId": verificationContext.VerificationId,
					})
					return nil, errors.New("invalid layout metadata and unable to save state")
				}
				
				envelope := NewExtendedEnvelope(basicEnvelope)
				envelope.VerificationContext = verificationContext
				envelope.Envelope.SetStatus(schema.StatusInitializationFailed)
				return envelope, errors.New("retrieved layout metadata is incomplete")
			}
			
			s.logger.Info("Layout metadata successfully populated from referenceImageUrl", map[string]interface{}{
				"verificationId":   verificationContext.VerificationId,
				"layoutId":         verificationContext.LayoutId,
				"layoutPrefix":     verificationContext.LayoutPrefix,
			})
		}
	}

	// Validate the verification context
	if err := s.validateVerificationContext(verificationContext); err != nil {
		s.logger.Error("Verification context validation failed", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify resources (S3 images, layout metadata)
	resourceValidation, err := s.verifyResources(ctx, verificationContext)
	if err != nil {
		s.logger.Error("Resource verification failed", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})

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

		basicEnvelope, stateErr := s.createStateStructure(ctx, verificationContext)
		if stateErr != nil {
			s.logger.Error("Failed to create state structure after resource validation failure", map[string]interface{}{
				"error": stateErr.Error(),
				"verificationId": verificationContext.VerificationId,
			})
		} else if basicEnvelope != nil {
			envelope := NewExtendedEnvelope(basicEnvelope)
			envelope.VerificationContext = verificationContext
			envelope.Envelope.SetStatus(schema.StatusInitializationFailed)
			return envelope, err 
		}
		return nil, err
	}

	verificationContext.ResourceValidation = resourceValidation
	verificationContext.Status = schema.StatusVerificationInitialized

	basicEnvelope, err := s.createStateStructure(ctx, verificationContext)
	if err != nil {
		s.logger.Error("Failed to create state structure", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to create state structure: %w", err)
	}
	
	var s3InitRef *s3state.Reference
	if basicEnvelope.References != nil {
		s3InitRef = basicEnvelope.References["processing_initialization"]
	}
	if s3InitRef == nil {
		s.logger.Warn("No 'processing_initialization' S3 reference found in envelope to store in DynamoDB.", map[string]interface{}{
			"verificationId": verificationContext.VerificationId,
		})
	}

	err = s.verificationRepo.StoreMinimalRecord(ctx, verificationContext, s3InitRef)
	if err != nil {
		s.logger.Error("Failed to store verification record", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to store verification record: %w", err)
	}

	basicEnvelope.AddSummary("verificationType", verificationContext.VerificationType)
	basicEnvelope.AddSummary("resourcesValidated", []string{"referenceImage", "checkingImage", "layoutMetadata"})
	basicEnvelope.AddSummary("contextEstablished", true)
	basicEnvelope.AddSummary("stateStructureCreated", true)

	envelope := NewExtendedEnvelope(basicEnvelope)
	envelope.VerificationContext = verificationContext

	s.logger.Info("Verification initialized successfully", map[string]interface{}{
		"verificationId":   verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"status":           verificationContext.Status,
		"s3StateBucket":    s.config.StateBucket,
	})

	return envelope, nil
}

// createStateStructure creates the S3 state structure and saves the initialization context
func (s *InitializeService) createStateStructure(ctx context.Context, verificationContext *schema.VerificationContext) (*s3state.Envelope, error) {
	err := s.s3StateManager.InitializeStateStructure(ctx, verificationContext.VerificationId)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize state structure: %w", err)
	}

	envelope := s.s3StateManager.CreateEnvelope(verificationContext.VerificationId)
	envelope.SetStatus(verificationContext.Status)

	err = s.s3StateManager.SaveContext(envelope, verificationContext)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification context: %w", err)
	}

	return envelope, nil
}

// validateVerificationContext validates the verification context
func (s *InitializeService) validateVerificationContext(verificationContext *schema.VerificationContext) error {
	if verificationContext.VerificationType == "" {
		return fmt.Errorf("%w: verificationType", ErrMissingRequiredField)
	}

	if verificationContext.VerificationType != schema.VerificationTypeLayoutVsChecking &&
		verificationContext.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return fmt.Errorf("%w: %s", ErrInvalidVerificationType, verificationContext.VerificationType)
	}

	// referenceImageUrl and checkingImageUrl are always required.
	if verificationContext.ReferenceImageUrl == "" {
		return fmt.Errorf("%w: referenceImageUrl", ErrMissingRequiredField)
	}
	if verificationContext.CheckingImageUrl == "" {
		return fmt.Errorf("%w: checkingImageUrl", ErrMissingRequiredField)
	}

	// --- BEGIN MODIFICATION: Relaxed validation for LayoutId and LayoutPrefix ---
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// The following checks are now removed as per user request.
		// It's assumed that if layoutId/layoutPrefix are not provided,
		// the lookup attempt via referenceImageUrl in the Process method should populate them.
		// If that lookup also fails, and these fields are critical for downstream
		// (e.g., in verifyResources), an error will be raised at that point.
		/*
			if verificationContext.LayoutId == 0 {
				return fmt.Errorf("%w: layoutId required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
			}
			if verificationContext.LayoutPrefix == "" {
				return fmt.Errorf("%w: layoutPrefix required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
			}
		*/
		s.logger.Info("For LAYOUT_VS_CHECKING, validation for LayoutId and LayoutPrefix is relaxed. "+
			"Their presence relies on direct input or lookup via referenceImageUrl.", map[string]interface{}{
			"verificationId": verificationContext.VerificationId,
			"layoutId":       verificationContext.LayoutId,
			"layoutPrefix":   verificationContext.LayoutPrefix,
		})
	}
	// --- END MODIFICATION ---

	_, _, err := s.s3URLParser.ParseS3URLs(verificationContext.ReferenceImageUrl, verificationContext.CheckingImageUrl)
	if err != nil {
		return err // Return parsing errors
	}

	if verificationContext.ReferenceImageUrl == verificationContext.CheckingImageUrl {
		return ErrSameReferenceAndChecking
	}

	return nil
}

// verifyResources checks if all resources (images, layout) exist
func (s *InitializeService) verifyResources(ctx context.Context, verificationContext *schema.VerificationContext) (*schema.ResourceValidation, error) {
	resourceValidation := &schema.ResourceValidation{
		ValidationTimestamp: schema.FormatISO8601(),
	}

	const maxImageSize = 5 * 1024 * 1024 // 5MB

	imageURLs := []string{verificationContext.ReferenceImageUrl, verificationContext.CheckingImageUrl}
	validationResult, err := s.s3Validator.ValidateImagesInParallel(ctx, imageURLs, maxImageSize)
	if err != nil {
		if validationResult != nil {
			resourceValidation.ReferenceImageExists = validationResult.ReferenceImageExists
			resourceValidation.CheckingImageExists = validationResult.CheckingImageExists
		}
		return resourceValidation, err
	}

	resourceValidation.ReferenceImageExists = validationResult.ReferenceImageExists
	resourceValidation.CheckingImageExists = validationResult.CheckingImageExists

	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// This check critically depends on LayoutId and LayoutPrefix being correctly populated,
		// either from the request or via the lookup using referenceImageUrl.
		// If they are 0 or "" here, this check will likely fail as intended.
		if verificationContext.LayoutId == 0 || verificationContext.LayoutPrefix == "" {
			s.logger.Warn("LayoutId or LayoutPrefix is missing for LAYOUT_VS_CHECKING before verifying layout existence. Layout verification will likely fail.", map[string]interface{}{
				"verificationId":    verificationContext.VerificationId,
				"layoutId":          verificationContext.LayoutId,
				"layoutPrefix":      verificationContext.LayoutPrefix,
			})
			// Return an error here because attempting to verify a layout with ID 0 or empty prefix is problematic.
			return resourceValidation, fmt.Errorf("cannot verify layout: LayoutId or LayoutPrefix is missing for LAYOUT_VS_CHECKING. "+
				"Ensure referenceImageUrl maps to a valid layout or provide LayoutId/Prefix directly. "+
				"Current LayoutId: %d, LayoutPrefix: '%s'", verificationContext.LayoutId, verificationContext.LayoutPrefix)
		}

		exists, err := s.layoutRepo.VerifyLayoutExists(ctx, verificationContext.LayoutId, verificationContext.LayoutPrefix)
		if err != nil {
			return resourceValidation, fmt.Errorf("error checking layout: %w", err)
		}
		if !exists {
			return resourceValidation, fmt.Errorf("layout with ID %d and prefix '%s' not found",
				verificationContext.LayoutId, verificationContext.LayoutPrefix)
		}
		resourceValidation.LayoutExists = true
	}

	return resourceValidation, nil
}

// generateVerificationId creates a unique verification ID
func (s *InitializeService) generateVerificationId() string {
	now := time.Now().UTC()
	ts := now.Format("20060102150405")
	randomSuffix := uuid.New().String()[0:4]
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

	if request.SchemaVersion != "" && request.VerificationContext != nil {
		if vc, ok := request.VerificationContext.(*schema.VerificationContext); ok {
			verificationContext = vc
			if verificationContext.VerificationId == "" {
				verificationContext.VerificationId = s.generateVerificationId()
				s.logger.Info("Generated missing verificationId for provided context", map[string]interface{}{
					"verificationId": verificationContext.VerificationId,
				})
			}
			if verificationContext.VerificationAt == "" {
				now := time.Now().UTC()
				verificationContext.VerificationAt = now.Format(time.RFC3339)
				s.logger.Info("Setting missing verificationAt for provided context", map[string]interface{}{
					"verificationId": verificationContext.VerificationId,
					"verificationAt": verificationContext.VerificationAt,
				})
			}
			s.logger.Info("Using provided verification context with schema version", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
				"schemaVersion":  request.SchemaVersion,
			})
			return verificationContext, nil
		}
		s.logger.Error("Invalid verification context type provided with schema version", map[string]interface{}{
			"type": fmt.Sprintf("%T", request.VerificationContext),
		})
		return nil, fmt.Errorf("invalid verification context type with schema version")
	}

	verificationId := s.generateVerificationId()
	now := time.Now().UTC()
	isoTimestamp := now.Format(time.RFC3339)

	verificationContext = &schema.VerificationContext{
		VerificationId:      verificationId,
		VerificationAt:      isoTimestamp,
		Status:              schema.StatusVerificationInitialized,
		VerificationType:    request.VerificationType,
		VendingMachineId:    request.VendingMachineId,
		ReferenceImageUrl:   request.ReferenceImageUrl,
		CheckingImageUrl:    request.CheckingImageUrl,
	}

	if request.VerificationType == schema.VerificationTypeLayoutVsChecking {
		if request.LayoutId != 0 {
			verificationContext.LayoutId = request.LayoutId
		}
		if request.LayoutPrefix != "" {
			verificationContext.LayoutPrefix = request.LayoutPrefix
		}
	}

	if request.VerificationType == schema.VerificationTypePreviousVsCurrent {
		verificationContext.PreviousVerificationId = request.PreviousVerificationId
	}

	verificationContext.TurnTimestamps = &schema.TurnTimestamps{
		Initialized: isoTimestamp,
	}

	verificationContext.ConversationType = "two-turn"
	verificationContext.TurnConfig = &schema.TurnConfig{
		MaxTurns:           2,
		ReferenceImageTurn: 1,
		CheckingImageTurn:  2,
	}

	if request.ConversationConfig.Type != "" {
		verificationContext.ConversationType = request.ConversationConfig.Type
	}
	if request.ConversationConfig.MaxTurns > 0 {
		verificationContext.TurnConfig.MaxTurns = request.ConversationConfig.MaxTurns
	}

	requestTimestamp := request.RequestTimestamp
	if requestTimestamp == "" {
		requestTimestamp = isoTimestamp
	}

	verificationContext.RequestMetadata = &schema.RequestMetadata{
		RequestId:         request.RequestId,
		RequestTimestamp:  requestTimestamp,
		ProcessingStarted: isoTimestamp,
	}

	s.logger.Debug("Created verification context from request parameters", map[string]interface{}{
		"verificationId":   verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
	})

	return verificationContext, nil
}
