package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
	"workflow-function/shared/schema"
)

// Custom error types for better error handling
var (
	ErrMissingRequiredField   = errors.New("missing required field")
	ErrInvalidImageFormat     = errors.New("invalid image format")
	ErrInvalidVerificationType = errors.New("invalid verification type")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrDatabaseFailure        = errors.New("database operation failed")
)

// InitService handles the business logic for the verification initialization
type InitService struct {
	deps    *Dependencies
	config  ConfigVars
	s3Utils *s3utils.S3Utils
	dbUtils *dbutils.DynamoDBUtils
	logger  logger.Logger
}

// NewInitService creates a new instance of InitService
func NewInitService(deps *Dependencies, config ConfigVars) *InitService {
	return &InitService{
		deps:    deps,
		config:  config,
		s3Utils: deps.GetS3Util(),
		dbUtils: deps.GetDynamoUtil(),
		logger:  deps.GetLogger().WithFields(map[string]interface{}{"component": "InitService"}),
	}
}

// Process handles the entire verification initialization workflow
func (s *InitService) Process(ctx context.Context, request InitRequest) (*schema.VerificationContext, error) {
	// If using the new standardized schema format
	if request.SchemaVersion != "" && request.VerificationContext != nil {
		// Validate the verification context
		errors := schema.ValidateVerificationContext(request.VerificationContext)
		if len(errors) > 0 {
			s.logger.Error("Verification context validation failed", map[string]interface{}{
				"errors": errors.Error(),
			})
			return nil, fmt.Errorf("validation failed: %s", errors.Error())
		}
		
		// No need to set status since state machine handles it
		
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
			
			request.VerificationContext.Error = errorInfo
			return request.VerificationContext, err
		}
		
		// Update the verification context with resource validation
		request.VerificationContext.ResourceValidation = resourceValidation
		
		// Update the verification ID if not already set
		if request.VerificationContext.VerificationId == "" {
			request.VerificationContext.VerificationId = s.generateVerificationId()
		}
		
		// Store in DynamoDB
		if err := s.dbUtils.StoreVerificationRecord(ctx, request.VerificationContext); err != nil {
			s.logger.Error("Failed to store verification record", map[string]interface{}{
				"error": err.Error(),
				"verificationId": request.VerificationContext.VerificationId,
			})
			return nil, err
		}
		
		s.logger.Info("Verification initialized successfully with standardized schema", map[string]interface{}{
			"verificationId": request.VerificationContext.VerificationId,
			"verificationType": request.VerificationContext.VerificationType,
			"status": request.VerificationContext.Status,
			"schemaVersion": request.SchemaVersion,
		})
		
		return request.VerificationContext, nil
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
		previousVerification, err := s.dbUtils.GetVerificationRecord(ctx, request.PreviousVerificationId)
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
		previousVerification, err := s.dbUtils.FindPreviousVerification(ctx, request.ReferenceImageUrl)
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
	if err := s.dbUtils.StoreVerificationRecord(ctx, verificationContext); err != nil {
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
func (s *InitService) validateRequest(request InitRequest) error {
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
	} else if request.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// For PREVIOUS_VS_CURRENT, both previousVerificationId and vendingMachineId are optional
		// No additional validation needed here
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

	// If reference and checking images are the same, reject
	if request.ReferenceImageUrl == request.CheckingImageUrl {
		return fmt.Errorf("reference and checking images cannot be the same")
	}

	return nil
}

// verifyResources checks if all resources (images, layout) exist
func (s *InitService) verifyResources(ctx context.Context, request InitRequest) (*schema.ResourceValidation, error) {
	// Create validation result
	resourceValidation := &schema.ResourceValidation{
		ValidationTimestamp: schema.FormatISO8601(),
	}

	// Create error channel and wait group for concurrent verification
	errChan := make(chan error, 3)
	var wg sync.WaitGroup
	
	// Determine image URLs based on request format
	var referenceImageUrl, checkingImageUrl string
	var verificationType string
	var layoutId int
	var layoutPrefix string
	
	if request.VerificationContext != nil {
		// Get from verification context
		referenceImageUrl = request.VerificationContext.ReferenceImageUrl
		checkingImageUrl = request.VerificationContext.CheckingImageUrl
		verificationType = request.VerificationContext.VerificationType
		layoutId = request.VerificationContext.LayoutId
		layoutPrefix = request.VerificationContext.LayoutPrefix
	} else {
		// Get from direct fields
		referenceImageUrl = request.ReferenceImageUrl
		checkingImageUrl = request.CheckingImageUrl
		verificationType = request.VerificationType
		layoutId = request.LayoutId
		layoutPrefix = request.LayoutPrefix
	}
	
	// Always verify both images
	wg.Add(2)

	// If LAYOUT_VS_CHECKING, also verify layout
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		wg.Add(1)
	}

	// Use a constant max size for all images (10MB)
	const maxImageSize = 10 * 1024 * 1024

	// Verify reference image existence in parallel
	go func() {
		defer wg.Done()
		exists, err := s.s3Utils.ValidateImageExists(ctx, referenceImageUrl, maxImageSize)
		if err != nil {
			errChan <- fmt.Errorf("reference image verification failed: %w", err)
			return
		}
		resourceValidation.ReferenceImageExists = exists
		errChan <- nil
	}()

	// Verify checking image existence in parallel
	go func() {
		defer wg.Done()
		exists, err := s.s3Utils.ValidateImageExists(ctx, checkingImageUrl, maxImageSize)
		if err != nil {
			errChan <- fmt.Errorf("checking image verification failed: %w", err)
			return
		}
		resourceValidation.CheckingImageExists = exists
		errChan <- nil
	}()

	// Verify layout exists in DynamoDB only for LAYOUT_VS_CHECKING
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		go func() {
			defer wg.Done()
			exists, err := s.dbUtils.VerifyLayoutExists(ctx, layoutId, layoutPrefix)
			if err != nil {
				errChan <- fmt.Errorf("error checking layout: %w", err)
				return
			}
			if !exists {
				errChan <- fmt.Errorf("layout with ID %d and prefix %s not found", 
					layoutId, layoutPrefix)
				return
			}
			resourceValidation.LayoutExists = true
			errChan <- nil
		}()
	}

	// Use a separate goroutine to close the channel after all workers are done
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors from all goroutines
	for err := range errChan {
		if err != nil {
			return resourceValidation, err
		}
	}

	return resourceValidation, nil
}

// -------------------------
// Historical Context Models (canonical definitions)
// -------------------------

// HistoricalContext represents historical verification data
type HistoricalContext struct {
	PreviousVerificationId     string               `json:"previousVerificationId"`
	PreviousVerificationAt     string               `json:"previousVerificationAt"`
	PreviousVerificationStatus string               `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64              `json:"hoursSinceLastVerification"`
	MachineStructure           *MachineStructure    `json:"machineStructure,omitempty"`
	VerificationSummary        *VerificationSummary `json:"verificationSummary,omitempty"`
	CheckingStatus             map[string]string    `json:"checkingStatus,omitempty"`
}

// MachineStructure contains information about the vending machine layout
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// VerificationSummary contains summary information from a previous verification
type VerificationSummary struct {
	TotalPositionsChecked  int     `json:"totalPositionsChecked"`
	CorrectPositions       int     `json:"correctPositions"`
	DiscrepantPositions    int     `json:"discrepantPositions"`
	MissingProducts        int     `json:"missingProducts"`
	IncorrectProductTypes  int     `json:"incorrectProductTypes"`
	UnexpectedProducts     int     `json:"unexpectedProducts"`
	EmptyPositionsCount    int     `json:"emptyPositionsCount"`
	OverallAccuracy        float64 `json:"overallAccuracy"`
	OverallConfidence      float64 `json:"overallConfidence"`
	VerificationStatus     string  `json:"verificationStatus"`
	VerificationOutcome    string  `json:"verificationOutcome"`
}

// createHistoricalContext creates a historical context from a previous verification
func (s *InitService) createHistoricalContext(previous *schema.VerificationContext) *HistoricalContext {
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
func (s *InitService) generateVerificationId() string {
	// Generate a unique verification ID: verif-<timestamp>-<4-char random>
	now := time.Now().UTC()
	ts := now.Format("20060102150405")                // e.g. "20250421153025"
	randomSuffix := uuid.New().String()[0:4]          // e.g. "a1b4"
	base := fmt.Sprintf("%s%s", s.config.VerificationPrefix, ts)
	return fmt.Sprintf("%s-%s", base, randomSuffix)
}

// createVerificationContext creates the verification context with a unique ID
func (s *InitService) createVerificationContext(
	request InitRequest, 
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

	// Handle conversation configuration
	if request.ConversationConfig != nil {
		verificationContext.ConversationType = request.ConversationConfig.Type
		verificationContext.TurnConfig = &schema.TurnConfig{
			MaxTurns:           request.ConversationConfig.MaxTurns,
			ReferenceImageTurn: 1,
			CheckingImageTurn:  2,
		}
	} else {
		// Default to two-turn configuration
		verificationContext.ConversationType = "two-turn"
		verificationContext.TurnConfig = &schema.TurnConfig{
			MaxTurns:           2,
			ReferenceImageTurn: 1,
			CheckingImageTurn:  2,
		}
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