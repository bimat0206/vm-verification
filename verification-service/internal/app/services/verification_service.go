package services

import (
	"context"
	"fmt"
	"time"

	"verification-service/internal/domain/models"
	domainServices "verification-service/internal/domain/services"
	"verification-service/internal/infrastructure/logger"
	"verification-service/pkg/errors"
)

// VerificationRepository defines the interface for verification data storage
type VerificationRepository interface {
	CreateVerification(ctx context.Context, verification *models.VerificationContext) error
	GetVerification(ctx context.Context, id string) (*models.VerificationContext, error)
	UpdateVerificationStatus(ctx context.Context, id string, status models.VerificationStatus) error
	StoreReferenceAnalysis(ctx context.Context, verificationID string, analysis *models.ReferenceAnalysis) error
	StoreCheckingAnalysis(ctx context.Context, verificationID string, analysis *models.CheckingAnalysis) error
	StoreFinalResults(ctx context.Context, verificationID string, results *models.VerificationResult) error
	StoreResultImageURL(ctx context.Context, verificationID string, url string) error
	ListVerifications(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.VerificationResult, int, error)
}

// ImageServiceInterface defines the interface for image operations
type ImageServiceInterface interface {
	FetchReferenceImage(ctx context.Context, url string) ([]byte, map[string]interface{}, error)
	FetchCheckingImage(ctx context.Context, url string) ([]byte, error)
	GetImageMetadata(ctx context.Context, imageBytes []byte) (map[string]interface{}, error)
	ImageToBase64(imageBytes []byte) string
	StoreResultImage(ctx context.Context, verificationID string, imageBytes []byte) (string, error)
}

// VisualizationServiceInterface defines the interface for generating result visualizations
type VisualizationServiceInterface interface {
	GenerateVisualization(
		ctx context.Context,
		verificationContext *models.VerificationContext,
		results *models.VerificationResult,
		referenceImage []byte,
		checkingImage []byte,
	) (string, error) // Returns S3 URL
}

// NotificationServiceInterface defines the interface for sending notifications
type NotificationServiceInterface interface {
	SendNotifications(
		ctx context.Context,
		verificationContext *models.VerificationContext,
		results *models.VerificationResult,
		resultImageURL string,
	) error
}

// AIProvider defines the interface for AI operations
type AIProvider interface {
	InvokeModel(
		ctx context.Context,
		systemPrompt string,
		userPrompt string,
		images []string, // Base64 encoded
		conversationContext map[string]interface{},
	) (string, map[string]interface{}, error) // Returns response, metadata, error
}

// VerificationService handles the verification workflow
type VerificationService struct {
	repository         VerificationRepository
	imageService       ImageServiceInterface
	verificationEngine domainServices.VerificationEngine
	promptGenerator    domainServices.PromptGenerator
	responseAnalyzer   domainServices.ResponseAnalyzer
	aiProvider         AIProvider
	visualizationService VisualizationServiceInterface
	notificationService NotificationServiceInterface
	logger             *logger.StandardLogger
}

// NewVerificationService creates a new verification service
func NewVerificationService(
	repository VerificationRepository,
	imageService ImageServiceInterface,
	verificationEngine domainServices.VerificationEngine,
	promptGenerator domainServices.PromptGenerator,
	responseAnalyzer domainServices.ResponseAnalyzer,
	aiProvider AIProvider,
	visualizationService VisualizationServiceInterface,
	notificationService NotificationServiceInterface,
) *VerificationService {
	return &VerificationService{
		repository:         repository,
		imageService:       imageService,
		verificationEngine: verificationEngine,
		promptGenerator:    promptGenerator,
		responseAnalyzer:   responseAnalyzer,
		aiProvider:         aiProvider,
		visualizationService: visualizationService,
		notificationService: notificationService,
		logger:             logger.NewLogger(),
	}
}

// InitiateVerification starts a new verification process
func (s *VerificationService) InitiateVerification(ctx context.Context, params map[string]interface{}) (*models.VerificationContext, error) {
	s.logger.Debug("Starting verification process with params: %v", params)
	
	// Validate input parameters
	referenceImageURL, ok := params["referenceImageUrl"].(string)
	if !ok || referenceImageURL == "" {
		s.logger.Error("Missing or invalid referenceImageUrl")
		return nil, errors.NewValidationError("referenceImageUrl is required")
	}
	s.logger.Debug("Reference image URL: %s", referenceImageURL)

	checkingImageURL, ok := params["checkingImageUrl"].(string)
	if !ok || checkingImageURL == "" {
		s.logger.Error("Missing or invalid checkingImageUrl")
		return nil, errors.NewValidationError("checkingImageUrl is required")
	}
	s.logger.Debug("Checking image URL: %s", checkingImageURL)

	vendingMachineID, ok := params["vendingMachineId"].(string)
	if !ok || vendingMachineID == "" {
		s.logger.Error("Missing or invalid vendingMachineId")
		return nil, errors.NewValidationError("vendingMachineId is required")
	}
	s.logger.Debug("Vending machine ID: %s", vendingMachineID)

	layoutID, ok := params["layoutId"].(float64)
	if !ok {
		// Try to convert from different numeric types
		if intVal, ok := params["layoutId"].(int); ok {
			layoutID = float64(intVal)
			s.logger.Debug("Converted layoutId from int to float64: %f", layoutID)
		} else {
			s.logger.Error("Missing or invalid layoutId")
			return nil, errors.NewValidationError("layoutId is required")
		}
	}
	s.logger.Debug("Layout ID: %f", layoutID)

	layoutPrefix, ok := params["layoutPrefix"].(string)
	if !ok || layoutPrefix == "" {
		s.logger.Error("Missing or invalid layoutPrefix")
		return nil, errors.NewValidationError("layoutPrefix is required")
	}
	s.logger.Debug("Layout prefix: %s", layoutPrefix)

	// Generate verification ID
	now := time.Now()
	verificationID := fmt.Sprintf("verif-%s", now.Format("20060102150405"))
	s.logger.Debug("Generated verification ID: %s", verificationID)

	// Create verification context
	verificationContext := &models.VerificationContext{
		VerificationID:      verificationID,
		VerificationAt:      now,
		Status:              models.StatusInitialized,
		VendingMachineID:    vendingMachineID,
		LayoutID:            int(layoutID),
		LayoutPrefix:        layoutPrefix,
		ReferenceImageURL:   referenceImageURL,
		CheckingImageURL:    checkingImageURL,
		CurrentTurn:         0,
		TurnConfig: models.TurnConfig{
			MaxTurns:           2,
			ReferenceImageTurn: 1,
			CheckingImageTurn:  2,
			TurnTimeoutMs:      60000, // 60 seconds
		},
		TurnTimestamps: map[string]time.Time{
			"initialized": now,
		},
		NotificationEnabled: params["notificationEnabled"] == true,
	}
	s.logger.Debug("Created verification context: %+v", verificationContext)

	// Set processing metadata
	verificationContext.ProcessingMetadata.RequestID = fmt.Sprintf("req-%s", now.Format("20060102150405"))
	verificationContext.ProcessingMetadata.StartTime = now
	verificationContext.ProcessingMetadata.RetryCount = 0
	verificationContext.ProcessingMetadata.Timeout = 300000 // 5 minutes
	s.logger.Debug("Set processing metadata: %+v", verificationContext.ProcessingMetadata)

	// Store initial verification record
	s.logger.Debug("Storing verification record in DynamoDB")
	if err := s.repository.CreateVerification(ctx, verificationContext); err != nil {
		s.logger.Error("Failed to create verification record: %v", err)
		return nil, fmt.Errorf("failed to create verification record: %w", err)
	}
	s.logger.Info("Successfully stored verification record with ID: %s", verificationID)

	// Start async verification process
	s.logger.Debug("Starting async verification process")
	go s.runVerificationProcess(context.Background(), verificationContext)

	return verificationContext, nil
}

// runVerificationProcess handles the entire verification workflow
func (s *VerificationService) runVerificationProcess(ctx context.Context, verificationContext *models.VerificationContext) {
	s.logger.Info("Starting verification workflow for ID: %s", verificationContext.VerificationID)
	
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic recovered in verification process: %v", r)
			s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusError)
		}
	}()

	// 1. Fetch images
	s.logger.Debug("Fetching reference image from: %s", verificationContext.ReferenceImageURL)
	referenceImage, layoutMetadata, err := s.imageService.FetchReferenceImage(ctx, verificationContext.ReferenceImageURL)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to fetch reference image")
		return
	}
	s.logger.Debug("Successfully fetched reference image, size: %d bytes", len(referenceImage))
	s.logger.Debug("Layout metadata: %+v", layoutMetadata)

	s.logger.Debug("Fetching checking image from: %s", verificationContext.CheckingImageURL)
	checkingImage, err := s.imageService.FetchCheckingImage(ctx, verificationContext.CheckingImageURL)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to fetch checking image")
		return
	}
	s.logger.Debug("Successfully fetched checking image, size: %d bytes", len(checkingImage))

	s.logger.Debug("Updating verification status to IMAGES_FETCHED")
	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusImagesFetched); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 2. Generate system prompt
	systemPrompt, err := s.promptGenerator.GenerateSystemPrompt(ctx, *verificationContext, layoutMetadata)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to generate system prompt")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusSystemPromptReady); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 3. Turn 1: Reference Layout Analysis
	// Generate Turn 1 prompt
	referenceImageB64 := s.imageService.ImageToBase64(referenceImage)
	turn1Prompt, err := s.promptGenerator.GenerateTurn1Prompt(ctx, *verificationContext, layoutMetadata, referenceImageB64)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to generate Turn 1 prompt")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn1PromptReady); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// Invoke model for Turn 1
	turn1Response, _, err := s.aiProvider.InvokeModel(ctx, systemPrompt, turn1Prompt, []string{referenceImageB64}, nil)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to invoke model for Turn 1")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn1Processing); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// Process Turn 1 response
	referenceAnalysis, err := s.responseAnalyzer.ProcessTurn1Response(ctx, turn1Response, layoutMetadata)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to process Turn 1 response")
		return
	}

	// Store reference analysis
	if err := s.repository.StoreReferenceAnalysis(ctx, verificationContext.VerificationID, referenceAnalysis); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to store reference analysis")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn1Completed); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 4. Turn 2: Checking Image Verification
	// Generate Turn 2 prompt
	checkingImageB64 := s.imageService.ImageToBase64(checkingImage)
	turn2Prompt, err := s.promptGenerator.GenerateTurn2Prompt(ctx, *verificationContext, layoutMetadata, checkingImageB64, referenceAnalysis)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to generate Turn 2 prompt")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn2PromptReady); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// Prepare conversation context
	conversationContext := map[string]interface{}{
		"referenceAnalysis": referenceAnalysis,
	}

	// Invoke model for Turn 2
	turn2Response, _, err := s.aiProvider.InvokeModel(ctx, systemPrompt, turn2Prompt, []string{checkingImageB64}, conversationContext)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to invoke model for Turn 2")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn2Processing); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// Process Turn 2 response
	checkingAnalysis, err := s.responseAnalyzer.ProcessTurn2Response(ctx, turn2Response, referenceAnalysis)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to process Turn 2 response")
		return
	}

	// Store checking analysis
	if err := s.repository.StoreCheckingAnalysis(ctx, verificationContext.VerificationID, checkingAnalysis); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to store checking analysis")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusTurn2Completed); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 5. Finalize Results
	result := s.finalizeResults(*verificationContext, referenceAnalysis, checkingAnalysis, layoutMetadata)

	// Store final results
	if err := s.repository.StoreFinalResults(ctx, verificationContext.VerificationID, result); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to store final results")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusResultsFinalized); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 6. Generate Visualization
	resultImageURL, err := s.visualizationService.GenerateVisualization(ctx, verificationContext, result, referenceImage, checkingImage)
	if err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to generate visualization")
		return
	}

	// Store result image URL
	if err := s.repository.StoreResultImageURL(ctx, verificationContext.VerificationID, resultImageURL); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to store result image URL")
		return
	}

	if err := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusResultsStored); err != nil {
		s.handleError(ctx, verificationContext, err, "Failed to update verification status")
		return
	}

	// 7. Send Notifications (if enabled)
	if verificationContext.NotificationEnabled {
		if err := s.notificationService.SendNotifications(ctx, verificationContext, result, resultImageURL); err != nil {
			// Log error but continue
			fmt.Printf("Failed to send notifications: %v\n", err)
		} else {
			s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusNotificationSent)
		}
	}
}

// finalizeResults combines reference and checking analysis to create the final result
func (s *VerificationService) finalizeResults(
	verificationContext models.VerificationContext,
	referenceAnalysis *models.ReferenceAnalysis,
	checkingAnalysis *models.CheckingAnalysis,
	layoutMetadata map[string]interface{},
) *models.VerificationResult {
	// Create a map of rows with discrepancies
	discrepantRows := make(map[string]bool)
	for _, discrepancy := range checkingAnalysis.Discrepancies {
		if len(discrepancy.Position) > 0 {
			row := string(discrepancy.Position[0])
			discrepantRows[row] = true
		}
	}

	// Find corrected rows (rows without discrepancies)
	correctedRows := []string{}
	for _, row := range referenceAnalysis.MachineStructure.RowOrder {
		if !discrepantRows[row] {
			correctedRows = append(correctedRows, row)
		}
	}

	// Create result
	result := &models.VerificationResult{
		VerificationID:    verificationContext.VerificationID,
		VerificationAt:    verificationContext.VerificationAt,
		Status:            checkingAnalysis.VerificationStatus,
		VendingMachineID:  verificationContext.VendingMachineID,
		LayoutID:          verificationContext.LayoutID,
		LayoutPrefix:      verificationContext.LayoutPrefix,
		ReferenceImageURL: verificationContext.ReferenceImageURL,
		CheckingImageURL:  verificationContext.CheckingImageURL,
		MachineStructure:  referenceAnalysis.MachineStructure,
		InitialConfirmation: referenceAnalysis.InitialConfirmation,
		CorrectedRows:     correctedRows,
		EmptySlotReport:   checkingAnalysis.EmptySlotReport,
		Discrepancies:     checkingAnalysis.Discrepancies,
	}

	// Extract reference and checking status
	result.ReferenceStatus = make(map[string]string)
	result.CheckingStatus = make(map[string]string)

	for row, analysis := range referenceAnalysis.RowAnalysis {
		if description, ok := analysis["description"].(string); ok {
			result.ReferenceStatus[row] = description
		}
	}

	for row, analysis := range checkingAnalysis.RowAnalysis {
		if description, ok := analysis["description"].(string); ok {
			result.CheckingStatus[row] = description
		}
	}

	// Calculate verification summary
	totalPositions := len(referenceAnalysis.MachineStructure.RowOrder) * referenceAnalysis.MachineStructure.ColumnsPerRow
	correctPositions := totalPositions - len(checkingAnalysis.Discrepancies)
	
	if correctPositions < 0 {
		correctPositions = 0 // Safeguard
	}

	// Count discrepancy types
	missingProducts := 0
	incorrectProductTypes := 0
	unexpectedProducts := 0
	
	for _, d := range checkingAnalysis.Discrepancies {
		switch d.Issue {
		case models.DiscrepancyMissingProduct:
			missingProducts++
		case models.DiscrepancyIncorrectProductType:
			incorrectProductTypes++
		case models.DiscrepancyUnexpectedProduct:
			unexpectedProducts++
		}
	}

	// Set summary
	result.VerificationSummary.TotalPositionsChecked = totalPositions
	result.VerificationSummary.CorrectPositions = correctPositions
	result.VerificationSummary.DiscrepantPositions = len(checkingAnalysis.Discrepancies)
	result.VerificationSummary.MissingProducts = missingProducts
	result.VerificationSummary.IncorrectProductTypes = incorrectProductTypes
	result.VerificationSummary.UnexpectedProducts = unexpectedProducts
	result.VerificationSummary.EmptyPositionsCount = checkingAnalysis.EmptySlotReport.TotalEmpty
	
	// Calculate accuracy
	if totalPositions > 0 {
		result.VerificationSummary.OverallAccuracy = float64(correctPositions) / float64(totalPositions) * 100
	} else {
		result.VerificationSummary.OverallAccuracy = 0
	}
	
	result.VerificationSummary.OverallConfidence = checkingAnalysis.Confidence
	result.VerificationSummary.VerificationStatus = checkingAnalysis.VerificationStatus

	// Generate outcome message
	if len(checkingAnalysis.Discrepancies) == 0 {
		result.VerificationSummary.VerificationOutcome = "No discrepancies detected."
	} else {
		summaries := []string{}
		rowDiscrepancies := make(map[string][]models.DiscrepancyType)
		
		for _, d := range checkingAnalysis.Discrepancies {
			if len(d.Position) > 0 {
				row := string(d.Position[0])
				rowDiscrepancies[row] = append(rowDiscrepancies[row], d.Issue)
			}
		}
		
		for row, issues := range rowDiscrepancies {
			// De-duplicate issues
			uniqueIssues := make(map[models.DiscrepancyType]bool)
			for _, issue := range issues {
				uniqueIssues[issue] = true
			}
			
			issueTypes := []models.DiscrepancyType{}
			for issue := range uniqueIssues {
				issueTypes = append(issueTypes, issue)
			}
			
			// Create description
			description := fmt.Sprintf("Row %s contains ", row)
			if len(issueTypes) == 1 {
				description += string(issueTypes[0])
			} else if len(issueTypes) == 2 {
				description += string(issueTypes[0]) + " and " + string(issueTypes[1])
			} else {
				for i, issue := range issueTypes[:len(issueTypes)-1] {
					if i > 0 {
						description += ", "
					}
					description += string(issue)
				}
				description += ", and " + string(issueTypes[len(issueTypes)-1])
			}
			
			summaries = append(summaries, description)
		}
		
		if len(summaries) > 0 {
			result.VerificationSummary.VerificationOutcome = "Discrepancies Detected - "
			for i, summary := range summaries {
				if i > 0 {
					result.VerificationSummary.VerificationOutcome += " and "
				}
				result.VerificationSummary.VerificationOutcome += summary
			}
		} else {
			result.VerificationSummary.VerificationOutcome = "Discrepancies Detected"
		}
	}

	// Set metadata
	result.Metadata.BedrockModel = "Claude 3.7 Sonnet"
	result.Metadata.CompletedAt = time.Now()
	result.Metadata.ProcessingTime = int(time.Since(verificationContext.ProcessingMetadata.StartTime).Milliseconds())

	return result
}

// handleError processes errors during verification
func (s *VerificationService) handleError(ctx context.Context, verificationContext *models.VerificationContext, err error, message string) {
	// Log error with details
	s.logger.Error("%s - %v", message, err)
	s.logger.Error("Verification ID: %s, VendingMachineID: %s", 
		verificationContext.VerificationID, 
		verificationContext.VendingMachineID)
	
	// Log additional context
	s.logger.Debug("Error context - Reference URL: %s", verificationContext.ReferenceImageURL)
	s.logger.Debug("Error context - Checking URL: %s", verificationContext.CheckingImageURL)
	s.logger.Debug("Error context - Current status: %s", verificationContext.Status)
	
	// Update verification status
	updateErr := s.repository.UpdateVerificationStatus(ctx, verificationContext.VerificationID, models.StatusError)
	if updateErr != nil {
		s.logger.Error("Failed to update verification status to ERROR: %v", updateErr)
	}
	
	// TODO: Implement more sophisticated error handling with recovery strategies
}

// GetVerification retrieves a verification by ID
func (s *VerificationService) GetVerification(ctx context.Context, id string) (*models.VerificationResult, error) {
	s.logger.Debug("Getting verification with ID: %s", id)
	
	// Get verification from repository
	// This is a simplified implementation - in a real system,
	// you would combine data from multiple sources based on the verification status
	
	// For this example, we'll assume the repository returns a complete result
	context, err := s.repository.GetVerification(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get verification with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}
	
	s.logger.Debug("Retrieved verification context: %+v", context)
	
	// In a real implementation, this would return the appropriate data based on the verification status
	// For now, we'll return a simplified result
	result := &models.VerificationResult{
		VerificationID:    context.VerificationID,
		VerificationAt:    context.VerificationAt,
		Status:            context.Status,
		VendingMachineID:  context.VendingMachineID,
		LayoutID:          context.LayoutID,
		LayoutPrefix:      context.LayoutPrefix,
		ReferenceImageURL: context.ReferenceImageURL,
		CheckingImageURL:  context.CheckingImageURL,
	}
	
	s.logger.Debug("Returning verification result: %+v", result)
	return result, nil
}

// ListVerifications retrieves a list of verifications with filtering
func (s *VerificationService) ListVerifications(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.VerificationResult, int, error) {
	s.logger.Debug("Listing verifications with filters: %v, limit: %d, offset: %d", filters, limit, offset)
	
	results, total, err := s.repository.ListVerifications(ctx, filters, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list verifications: %v", err)
		return nil, 0, err
	}
	
	s.logger.Debug("Retrieved %d verifications out of %d total", len(results), total)
	return results, total, nil
}
