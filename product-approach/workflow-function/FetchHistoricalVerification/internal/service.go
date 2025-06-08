package internal

import (
	"context"
	"fmt"
	"math"
	"time"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// HistoricalVerificationService handles fetching historical verification data
type HistoricalVerificationService struct {
	repo   *DynamoDBRepository
	logger logger.Logger
}

// NewHistoricalVerificationService creates a new service instance
func NewHistoricalVerificationService(repo *DynamoDBRepository, log logger.Logger) *HistoricalVerificationService {
	return &HistoricalVerificationService{
		repo:   repo,
		logger: log.WithFields(map[string]interface{}{
			"component": "historical-verification-service",
		}),
	}
}

// FetchHistoricalVerification retrieves historical verification data for "previous_vs_current" verification type
func (s *HistoricalVerificationService) FetchHistoricalVerification(ctx context.Context, verificationCtx schema.VerificationContext) (HistoricalContext, error) {
	s.logger.Info("Fetching historical verification", map[string]interface{}{
		"referenceImageUrl": verificationCtx.ReferenceImageUrl,
		"verificationId":    verificationCtx.VerificationId,
	})

	// Query DynamoDB to find most recent verification using referenceImageUrl as checking image
	// Exclude the current verification to find genuinely previous verification
	verification, err := s.repo.FindPreviousVerification(ctx, verificationCtx.ReferenceImageUrl, verificationCtx.VerificationId, verificationCtx.VerificationAt)
	if err != nil {
		s.logger.Warn("No previous verification found, creating fallback context", map[string]interface{}{
			"error": err.Error(),
		})
		// Return fallback context when no historical data is found
		return s.createFallbackContext(verificationCtx), nil
	}

	s.logger.Info("Found previous verification", map[string]interface{}{
		"previousVerificationId": verification.VerificationID,
		"previousVerificationAt": verification.VerificationAt,
	})

	return s.createHistoricalContext(ctx, verificationCtx, verification)
}

// createFallbackContext creates a fallback context when no historical data is found
func (s *HistoricalVerificationService) createFallbackContext(verificationCtx schema.VerificationContext) HistoricalContext {
	return HistoricalContext{
		VerificationID:              verificationCtx.VerificationId,
		VerificationType:            verificationCtx.VerificationType,
		ReferenceImageUrl:           verificationCtx.ReferenceImageUrl,
		CheckingImageUrl:            verificationCtx.CheckingImageUrl,
		HistoricalDataFound:         false,
		SourceType:                  "NO_HISTORICAL_DATA",
		PreviousVerification:        nil,
		TemporalContext:             nil,
		MachineStructure:            nil,
		PreviousVerificationSummary: nil,
	}
}

// createHistoricalContext creates a complete historical context with found data
func (s *HistoricalVerificationService) createHistoricalContext(ctx context.Context, verificationCtx schema.VerificationContext, verification *VerificationRecord) (HistoricalContext, error) {
	// Calculate temporal context
	temporalCtx, err := s.calculateTemporalContext(verification.VerificationAt, verificationCtx.VerificationAt)
	if err != nil {
		s.logger.Warn("Failed to calculate temporal context", map[string]interface{}{
			"error": err.Error(),
		})
		// Use default values if calculation fails
		temporalCtx = &TemporalContext{
			HoursSinceLastVerification: 0,
			DaysSinceLastVerification:  0,
			BusinessDaysSince:          0,
		}
	}

	// Create previous verification details
	prevVerification := &PreviousVerification{
		VerificationID:     verification.VerificationID,
		VerificationAt:     verification.VerificationAt,
		VerificationStatus: verification.VerificationStatus,
		VendingMachineID:   s.formatNullableString(verification.VendingMachineID),
		Location:           s.extractLocation(verification),
		LayoutID:           s.formatNullableString(s.extractLayoutID(verification)),
		LayoutPrefix:       s.formatNullableString(s.extractLayoutPrefix(verification)),
	}

	// Create verification summary
	prevSummary := s.createPreviousVerificationSummary(verification.VerificationSummary)

	// Create enhanced machine structure with totalPositions
	enhancedMachineStructure := s.createEnhancedMachineStructure(verification.MachineStructure)

	// Get turn2Processed URL from DynamoDB ConversationHistory table
	turn2ProcessedUrl, err := s.repo.GetTurn2ProcessedPath(ctx, verification.VerificationID)
	if err != nil {
		s.logger.Warn("Failed to get Turn2ProcessedPath from DynamoDB, using generated URL", map[string]interface{}{
			"error":                  err.Error(),
			"previousVerificationId": verification.VerificationID,
		})
		// Fallback to generated URL if DynamoDB query fails
		turn2ProcessedUrl = s.generateTurn2ProcessedUrl(verification.VerificationID, verification.VerificationAt)
	}

	return HistoricalContext{
		VerificationID:              verificationCtx.VerificationId,
		VerificationType:            verificationCtx.VerificationType,
		ReferenceImageUrl:           verificationCtx.ReferenceImageUrl,
		CheckingImageUrl:            verificationCtx.CheckingImageUrl,
		HistoricalDataFound:         true,
		Turn2Processed:              turn2ProcessedUrl,
		SourceType:                  "DYNAMODB_QUERY_RESULT",
		PreviousVerification:        prevVerification,
		TemporalContext:             temporalCtx,
		MachineStructure:            enhancedMachineStructure,
		PreviousVerificationSummary: prevSummary,
	}, nil
}

// calculateTemporalContext calculates temporal information between two verification times
func (s *HistoricalVerificationService) calculateTemporalContext(prevTimeStr, currentTimeStr string) (*TemporalContext, error) {
	prevTime, err := time.Parse(time.RFC3339, prevTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous time: %w", err)
	}

	currentTime, err := time.Parse(time.RFC3339, currentTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current time: %w", err)
	}

	duration := currentTime.Sub(prevTime)
	hours := duration.Hours()
	days := hours / 24
	businessDays := s.calculateBusinessDays(prevTime, currentTime)

	return &TemporalContext{
		HoursSinceLastVerification: math.Round(hours*100) / 100, // Round to 2 decimal places
		DaysSinceLastVerification:  math.Round(days*100) / 100,  // Round to 2 decimal places
		BusinessDaysSince:          businessDays,
	}, nil
}

// calculateBusinessDays calculates the number of business days between two dates
func (s *HistoricalVerificationService) calculateBusinessDays(start, end time.Time) int {
	if start.After(end) {
		return 0
	}

	businessDays := 0
	current := start

	for current.Before(end) || current.Equal(end.Truncate(24*time.Hour)) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			businessDays++
		}
		current = current.AddDate(0, 0, 1)
	}

	return businessDays
}

// formatNullableString formats a string with null indicator if empty
func (s *HistoricalVerificationService) formatNullableString(value string) string {
	if value == "" {
		return "null"
	}
	return value + "|null"
}

// extractLocation extracts location from verification record
func (s *HistoricalVerificationService) extractLocation(verification *VerificationRecord) string {
	// For now, return a default location since it's not in the current schema
	// This should be updated when location field is added to the verification record
	return "Office Building A, Floor 3"
}

// extractLayoutID extracts layout ID from verification record
func (s *HistoricalVerificationService) extractLayoutID(verification *VerificationRecord) string {
	// Layout ID is not currently stored in verification records for PREVIOUS_VS_CURRENT type
	// Return empty string which will be formatted as "null"
	return ""
}

// extractLayoutPrefix extracts layout prefix from verification record
func (s *HistoricalVerificationService) extractLayoutPrefix(verification *VerificationRecord) string {
	// Layout prefix is not currently stored in verification records for PREVIOUS_VS_CURRENT type
	// Return empty string which will be formatted as "null"
	return ""
}

// createPreviousVerificationSummary creates a previous verification summary from schema.VerificationSummary
func (s *HistoricalVerificationService) createPreviousVerificationSummary(summary schema.VerificationSummary) *PreviousVerificationSummary {
	// Calculate total positions from machine structure (default 60 for 6x10 grid)
	totalPositions := 60

	// Create discrepancy details
	discrepancyDetails := &DiscrepancyDetails{
		MissingProducts:       0,
		IncorrectProductTypes: 0,
		UnexpectedProducts:    summary.DiscrepantPositions,
	}

	// Format accuracy percentage
	accuracy := fmt.Sprintf("%.0f%% (%d/%d)",
		summary.OverallAccuracy*100,
		summary.CorrectPositions,
		totalPositions)

	// Format confidence percentage
	confidence := fmt.Sprintf("%.0f%%", summary.OverallConfidence*100)

	return &PreviousVerificationSummary{
		TotalPositionsChecked:    totalPositions,
		CorrectPositions:         summary.CorrectPositions,
		DiscrepantPositions:      summary.DiscrepantPositions,
		DiscrepancyDetails:       discrepancyDetails,
		EmptyPositionsInChecking: summary.EmptyPositionsCount,
		OverallAccuracy:          accuracy,
		OverallConfidence:        confidence,
		VerificationOutcome:      summary.VerificationOutcome,
	}
}

// createEnhancedMachineStructure creates an enhanced machine structure with totalPositions
func (s *HistoricalVerificationService) createEnhancedMachineStructure(machineStructure schema.MachineStructure) *EnhancedMachineStructure {
	totalPositions := machineStructure.RowCount * machineStructure.ColumnsPerRow

	return &EnhancedMachineStructure{
		RowCount:       machineStructure.RowCount,
		ColumnsPerRow:  machineStructure.ColumnsPerRow,
		RowOrder:       machineStructure.RowOrder,
		ColumnOrder:    machineStructure.ColumnOrder,
		TotalPositions: totalPositions,
	}
}

// generateTurn2ProcessedUrl generates the S3 URL for the turn2 processed response of the previous verification
func (s *HistoricalVerificationService) generateTurn2ProcessedUrl(verificationID, verificationAt string) string {
	// Parse the verification timestamp to extract date components
	verificationTime, err := time.Parse(time.RFC3339, verificationAt)
	if err != nil {
		s.logger.Warn("Failed to parse verification time for turn2 URL generation", map[string]interface{}{
			"error":          err.Error(),
			"verificationAt": verificationAt,
		})
		// Use current time as fallback
		verificationTime = time.Now()
	}

	// Extract date components
	year := verificationTime.Format("2006")
	month := verificationTime.Format("01")
	day := verificationTime.Format("02")

	// Construct the S3 URL following the expected pattern:
	// s3://bucket/year/month/day/verificationId/responses/turn2-processed-response.md
	// Note: We need to get the bucket name from configuration or environment
	bucketName := s.getStateBucketName()

	return fmt.Sprintf("s3://%s/%s/%s/%s/%s/responses/turn2-processed-response.md",
		bucketName, year, month, day, verificationID)
}

// getStateBucketName retrieves the state bucket name from configuration
func (s *HistoricalVerificationService) getStateBucketName() string {
	// This should ideally come from configuration or environment variables
	// For now, using the pattern from the example
	return "kootoro-dev-s3-state-f6d3xl"
}