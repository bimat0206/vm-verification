package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"workflow-function/shared/logger"
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
}

// NewInitializeService creates a new instance of InitializeService
func NewInitializeService(ctx context.Context, awsConfig aws.Config, cfg Config) (*InitializeService, error) {
	// Create logger
	log := logger.New("kootoro-verification", "InitializeFunction")
	
	// Create S3 client
	s3Client := NewS3Client(awsConfig, cfg, log)
	s3URLParser := NewS3URLParser(cfg, log)
	s3Validator := NewS3Validator(s3Client, s3URLParser, log)
	
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
	}, nil
}

// Process handles the verification initialization process
func (s *InitializeService) Process(ctx context.Context, request ProcessRequest) (*schema.VerificationContext, error) {
	// If using the new standardized schema format
	if request.SchemaVersion != "" && request.VerificationContext != nil {
		vc, ok := request.VerificationContext.(*schema.VerificationContext)
		if !ok {
			s.logger.Error("Invalid verification context type", map[string]interface{}{
				"type": fmt.Sprintf("%T", request.VerificationContext),
			})
			return nil, fmt.Errorf("invalid verification context type")
		}
		
		// Validate the verification context
		errors := schema.ValidateVerificationContext(vc)
		if len(errors) > 0 {
			s.logger.Error("Verification context validation failed", map[string]interface{}{
				"errors": errors.Error(),
			})
			return nil, fmt.Errorf("validation failed: %s", errors.Error())
		}
		
		// Verify resources
		resourceValidation, err := s.verifyResources(ctx, request)
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
			
			vc.Error = errorInfo
			return vc, err
		}
		
		// Update the verification context with resource validation
		vc.ResourceValidation = resourceValidation
		
		// Always replace the verification ID with our standard format
		// This ensures any UUID from Step Functions is replaced
		vc.VerificationId = s.generateVerificationId()
		
		// Store in DynamoDB
		if err := s.verificationRepo.StoreVerificationRecord(ctx, vc); err != nil {
			s.logger.Error("Failed to store verification record", map[string]interface{}{
				"error": err.Error(),
				"verificationId": vc.VerificationId,
			})
			return nil, err
		}
		
		s.logger.Info("Verification initialized successfully with standardized schema", map[string]interface{}{
			"verificationId": vc.VerificationId,
			"verificationType": vc.VerificationType,
			"status": vc.Status,
			"schemaVersion": request.SchemaVersion,
		})
		
		return vc, nil
	}

	// Legacy flow for backward compatibility
	s.logger.Info("Using legacy format processing", nil)
	
	// Ensure previousVerificationId field is set for PREVIOUS_VS_CURRENT type
	if request.VerificationType == schema.VerificationTypePreviousVsCurrent && request.PreviousVerificationId == "" {
		s.logger.Info("Setting empty previousVerificationId for PREVIOUS_VS_CURRENT verification", nil)
		request.PreviousVerificationId = ""  // Explicitly set to empty string to ensure it exists
	}

	s.logger.Info("Starting verification initialization process", map[string]interface{}{
		"verificationType": request.VerificationType,
		"vendingMachineId": request.VendingMachineId,
		"previousVerificationId": request.PreviousVerificationId,
	})

	// Step 1: Validate the request
	if err := s.validateRequest(request); err != nil {
		s.logger.Error("Request validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 2: Verify all resources exist
	resourceValidation, err := s.verifyResources(ctx, request)
	if err != nil {
		s.logger.Error("Resource verification failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 3: For PREVIOUS_VS_CURRENT, fetch historical context if previousVerificationId is provided
	var historicalContext *HistoricalContext
	if request.VerificationType == schema.VerificationTypePreviousVsCurrent && request.PreviousVerificationId != "" {
		previousVerification, err := s.verificationRepo.GetVerificationRecord(ctx, request.PreviousVerificationId)
		if err != nil {
			s.logger.Warn("Failed to fetch previous verification", map[string]interface{}{
				"previousVerificationId": request.PreviousVerificationId,
				"error": err.Error(),
			})
			// This is a non-critical error, we can continue without historical context
		} else if previousVerification != nil {
			historicalContext = s.createHistoricalContext(previousVerification)
		}
	} else if request.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// Try to find the most recent verification using the reference image as checking image
		previousVerification, err := s.verificationRepo.FindPreviousVerification(ctx, request.ReferenceImageUrl)
		if err != nil {
			s.logger.Warn("Failed to find previous verification", map[string]interface{}{
				"referenceImageUrl": request.ReferenceImageUrl,
				"error": err.Error(),
			})
			// This is a non-critical error, we can continue without historical context
		} else if previousVerification != nil {
			historicalContext = s.createHistoricalContext(previousVerification)
			// Update request with found previousVerificationId
			request.PreviousVerificationId = previousVerification.VerificationId
		}
	}

	// Step 4: Generate verification ID and context
	verificationContext, err := s.createVerificationContext(request, resourceValidation, historicalContext)
	if err != nil {
		s.logger.Error("Failed to create verification context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Step 5: Store in DynamoDB
	if err := s.verificationRepo.StoreVerificationRecord(ctx, verificationContext); err != nil {
		s.logger.Error("Failed to store verification record", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, err
	}

	s.logger.Info("Verification initialized successfully", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"status": verificationContext.Status,
	})

	return verificationContext, nil
}

// validateRequest validates the request parameters
func (s *InitializeService) validateRequest(request ProcessRequest) error {
	// Check verification type
	if request.VerificationType == "" {
		return fmt.Errorf("%w: verificationType", ErrMissingRequiredField)
	}
	
	// Verify verification type is supported
	if request.VerificationType != schema.VerificationTypeLayoutVsChecking && 
	   request.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return fmt.Errorf("%w: %s", ErrInvalidVerificationType, request.VerificationType)
	}

	// Check required fields for all verification types
	if request.ReferenceImageUrl == "" {
		return fmt.Errorf("%w: referenceImageUrl", ErrMissingRequiredField)
	}
	if request.CheckingImageUrl == "" {
		return fmt.Errorf("%w: checkingImageUrl", ErrMissingRequiredField)
	}
	
	// Check required fields specific to verification type
	if request.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// Layout vs. Checking requires layoutId and layoutPrefix
		if request.LayoutId == 0 {
			return fmt.Errorf("%w: layoutId required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
		}
		if request.LayoutPrefix == "" {
			return fmt.Errorf("%w: layoutPrefix required for %s", ErrMissingRequiredField, schema.VerificationTypeLayoutVsChecking)
		}
	}

	// Validate S3 URLs format
	_, _, err := s.s3URLParser.ParseS3URLs(request.ReferenceImageUrl, request.CheckingImageUrl)
	if err != nil {
		return err
	}

	// If reference and checking images are the same, reject
	if request.ReferenceImageUrl == request.CheckingImageUrl {
		return ErrSameReferenceAndChecking
	}

	return nil
}

// verifyResources checks if all resources (images, layout) exist
func (s *InitializeService) verifyResources(ctx context.Context, request ProcessRequest) (*schema.ResourceValidation, error) {
	// Create validation result
	resourceValidation := &schema.ResourceValidation{
		ValidationTimestamp: schema.FormatISO8601(),
	}
	
	// Determine image URLs based on request format
	var referenceImageUrl, checkingImageUrl string
	var verificationType string
	var layoutId int
	var layoutPrefix string
	
	if vc, ok := request.VerificationContext.(*schema.VerificationContext); ok {
		// Get from verification context
		referenceImageUrl = vc.ReferenceImageUrl
		checkingImageUrl = vc.CheckingImageUrl
		verificationType = vc.VerificationType
		layoutId = vc.LayoutId
		layoutPrefix = vc.LayoutPrefix
	} else {
		// Get from direct fields
		referenceImageUrl = request.ReferenceImageUrl
		checkingImageUrl = request.CheckingImageUrl
		verificationType = request.VerificationType
		layoutId = request.LayoutId
		layoutPrefix = request.LayoutPrefix
	}
	
	// Use a constant max size for all images (10MB)
	const maxImageSize = 10 * 1024 * 1024
	
	// Validate both images
	imageURLs := []string{referenceImageUrl, checkingImageUrl}
	validationResult, err := s.s3Validator.ValidateImagesInParallel(ctx, imageURLs, maxImageSize)
	if err != nil {
		return resourceValidation, err
	}
	
	// Copy validation results
	resourceValidation.ReferenceImageExists = validationResult.ReferenceImageExists
	resourceValidation.CheckingImageExists = validationResult.CheckingImageExists
	
	// Verify layout exists in DynamoDB only for LAYOUT_VS_CHECKING
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		exists, err := s.layoutRepo.VerifyLayoutExists(ctx, layoutId, layoutPrefix)
		if err != nil {
			return resourceValidation, fmt.Errorf("error checking layout: %w", err)
		}
		if !exists {
			return resourceValidation, fmt.Errorf("layout with ID %d and prefix %s not found", 
				layoutId, layoutPrefix)
		}
		resourceValidation.LayoutExists = true
	}
	
	return resourceValidation, nil
}

// createHistoricalContext creates a historical context from a previous verification
func (s *InitializeService) createHistoricalContext(previous *schema.VerificationContext) *HistoricalContext {
	// Parse previous verification timestamp
	previousTime, err := time.Parse(time.RFC3339, previous.VerificationAt)
	if err != nil {
		s.logger.Warn("Failed to parse previous verification timestamp", map[string]interface{}{
			"previousVerificationId": previous.VerificationId,
			"timestamp": previous.VerificationAt,
		})
		previousTime = time.Now().UTC() // Fallback to current time
	}

	// Calculate hours since last verification
	hoursSince := time.Since(previousTime).Hours()

	// Basic historical context
	historicalContext := &HistoricalContext{
		PreviousVerificationId:     previous.VerificationId,
		PreviousVerificationAt:     previous.VerificationAt,
		PreviousVerificationStatus: previous.Status,
		HoursSinceLastVerification: hoursSince,
	}

	// Machine structure can be retrieved from previous verification context
	// This is a simplified example, in a real implementation you would extract
	// more detailed information from the previous verification
	machineStructure := &MachineStructure{
		RowCount:      6, // Default values, ideally extracted from previous verification
		ColumnsPerRow: 10,
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	}
	historicalContext.MachineStructure = machineStructure

	// In practice, other fields would be extracted from the actual verification results
	// stored in previous verification, such as checking status and verification summary.
	// This would require the VerificationResults table to store this data.

	return historicalContext
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

// createVerificationContext creates the verification context with a unique ID

// Logger returns the service logger
func (s *InitializeService) Logger() logger.Logger {
	return s.logger
}
// In internal/initialize_service.go
// Update the createVerificationContext method to add additional safety

// createVerificationContext creates the verification context with a unique ID
func (s *InitializeService) createVerificationContext(
	request ProcessRequest, 
	resourceValidation *schema.ResourceValidation,
	historicalContext *HistoricalContext,
) (*schema.VerificationContext, error) {
	// Generate a unique verification ID
	verificationId := s.generateVerificationId()

	// Format ISO timestamp
	now := time.Now().UTC()
	isoTimestamp := now.Format(time.RFC3339)

	// Create verification context
	verificationContext := &schema.VerificationContext{
		VerificationId:    verificationId,
		VerificationAt:    isoTimestamp,
		// Do not set Status field, as it's managed by the Step Functions state machine
		VerificationType:  request.VerificationType,
		VendingMachineId:  request.VendingMachineId,
		ReferenceImageUrl: request.ReferenceImageUrl,
		CheckingImageUrl:  request.CheckingImageUrl,
		ResourceValidation: resourceValidation,
		NotificationEnabled: request.NotificationEnabled,
	}

	// Set layout details for LAYOUT_VS_CHECKING
	if request.VerificationType == schema.VerificationTypeLayoutVsChecking {
		verificationContext.LayoutId = request.LayoutId
		verificationContext.LayoutPrefix = request.LayoutPrefix
	}

	// Set previousVerificationId for PREVIOUS_VS_CURRENT
	// Always include the field (even if empty) for PREVIOUS_VS_CURRENT verification type
	if request.VerificationType == schema.VerificationTypePreviousVsCurrent {
		verificationContext.PreviousVerificationId = request.PreviousVerificationId
		s.logger.Info("Setting previousVerificationId in context", map[string]interface{}{
			"previousVerificationId": verificationContext.PreviousVerificationId,
			"isEmpty": verificationContext.PreviousVerificationId == "",
		})
	}
	
	// Final verification - ensure required fields are not empty
	if verificationContext.VerificationId == "" {
		s.logger.Error("VerificationId is missing in the context", nil)
		return nil, fmt.Errorf("required field verificationId is missing in the context")
	}
	
	s.logger.Debug("Created verification context", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"status": verificationContext.Status,
	})

	// Handle conversation configuration with safety checks
	// Ensure we have valid values for the conversation configuration
	conversationType := request.ConversationConfig.Type
	maxTurns := request.ConversationConfig.MaxTurns
	
	// Apply defaults if values are not provided
	if conversationType == "" {
		conversationType = "two-turn"
	}
	if maxTurns <= 0 {
		maxTurns = 2
	}
	
	// Set the conversation configuration
	verificationContext.ConversationType = conversationType
	verificationContext.TurnConfig = &schema.TurnConfig{
		MaxTurns:           maxTurns,
		ReferenceImageTurn: 1,
		CheckingImageTurn:  2,
	}

	// Initialize turn timestamps
	verificationContext.TurnTimestamps = &schema.TurnTimestamps{
		Initialized: isoTimestamp,
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

	return verificationContext, nil
}