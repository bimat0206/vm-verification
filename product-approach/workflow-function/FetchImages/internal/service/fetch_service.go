// Package service provides business logic implementations for the FetchImages function
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"

	"workflow-function/FetchImages/internal/models"
	"workflow-function/FetchImages/internal/repository"
)

// FetchService handles the business logic for fetching images and related data
type FetchService struct {
	s3Repo       *repository.S3Repository
	dynamoDBRepo *repository.DynamoDBRepository
	stateManager *S3StateManager
	logger       logger.Logger
}

// NewFetchService creates a new FetchService instance
func NewFetchService(
	s3Repo *repository.S3Repository,
	dynamoDBRepo *repository.DynamoDBRepository,
	stateManager *S3StateManager,
	log logger.Logger,
) *FetchService {
	return &FetchService{
		s3Repo:       s3Repo,
		dynamoDBRepo: dynamoDBRepo,
		stateManager: stateManager,
		logger:       log.WithFields(map[string]interface{}{"component": "FetchService"}),
	}
}

// ProcessRequest orchestrates the fetching of image metadata and related data
func (s *FetchService) ProcessRequest(
	ctx context.Context,
	request *models.FetchImagesRequest,
) (*models.FetchImagesResponse, error) {
	processingStartTime := time.Now()
	s.logger.Info("Processing FetchImages request", map[string]interface{}{
		"verificationId": request.VerificationId,
	})

	// Load or create envelope
	envelope, err := s.stateManager.LoadEnvelope(request)
	if err != nil {
		return nil, models.NewProcessingError("failed to load state envelope", err)
	}

	// Load verification context
	context, err := s.loadVerificationContext(ctx, envelope, request)
	if err != nil {
		return nil, models.NewProcessingError("failed to load verification context", err)
	}

	// Convert to the expected type
	var verificationContext *schema.VerificationContext
	switch v := context.(type) {
	case *schema.VerificationContext:
		verificationContext = v
	case schema.VerificationContext:
		verificationContext = &v
	case map[string]interface{}:
		// Try to extract key fields from map
		verificationContext = &schema.VerificationContext{
			VerificationId:         getStringValue(v, "verificationId"),
			VerificationType:       getStringValue(v, "verificationType"),
			ReferenceImageUrl:      getStringValue(v, "referenceImageUrl"),
			CheckingImageUrl:       getStringValue(v, "checkingImageUrl"),
			LayoutId:               getIntValue(v, "layoutId"),
			LayoutPrefix:           getStringValue(v, "layoutPrefix"),
			PreviousVerificationId: getStringValue(v, "previousVerificationId"),
			VendingMachineId:       getStringValue(v, "vendingMachineId"),
		}
	default:
		return nil, models.NewProcessingError(
			"unsupported verification context type",
			fmt.Errorf("got type %T, expected schema.VerificationContext", context),
		)
	}

	// Set status to IMAGES_FETCHED
	verificationContext.Status = schema.StatusImagesFetched

	// Determine if we need layout or historical data based on verification type
	var prevVerificationId string
	if verificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// Make sure previousVerificationId exists for PREVIOUS_VS_CURRENT
		if verificationContext.PreviousVerificationId == "" {
			return nil, models.NewValidationError(
				"PreviousVerificationId is required for PREVIOUS_VS_CURRENT verification type",
				fmt.Errorf("missing previousVerificationId"),
			)
		}
		prevVerificationId = verificationContext.PreviousVerificationId
	}

	// Execute parallel operations to fetch everything we need
	results, err := s.fetchAllDataInParallel(
		ctx,
		envelope,
		verificationContext.ReferenceImageUrl,
		verificationContext.CheckingImageUrl,
		verificationContext.LayoutId,
		verificationContext.LayoutPrefix,
		prevVerificationId,
	)
	if err != nil {
		return nil, models.NewProcessingError("failed to fetch required data", err)
	}

	// Update envelope status
	s.stateManager.UpdateEnvelopeStatus(envelope, schema.StatusImagesFetched)

	// Create flat metadata structure for backward compatibility
	imgMetadata := &models.ImageMetadata{
		Reference: results.ReferenceMeta,
		Checking:  results.CheckingMeta,
	}

	// Convert to enhanced metadata structure
	enhancedMetadata := models.ConvertToEnhancedMetadata(
		verificationContext.VerificationId,
		verificationContext.VerificationType,
		imgMetadata,
		processingStartTime,
	)

	// Store the enhanced metadata instead of the flat structure
	if err := s.stateManager.StoreImageMetadata(envelope, enhancedMetadata); err != nil {
		return nil, models.NewProcessingError("failed to store image metadata", err)
	}

	// Store layout metadata if available
	if len(results.LayoutMeta) > 0 {
		if err := s.stateManager.StoreLayoutMetadata(envelope, results.LayoutMeta); err != nil {
			return nil, models.NewProcessingError("failed to store layout metadata", err)
		}
	}

	// Store historical context if available
	if len(results.HistoricalContext) > 0 {
		if err := s.stateManager.StoreHistoricalContext(envelope, results.HistoricalContext); err != nil {
			return nil, models.NewProcessingError("failed to store historical context", err)
		}
	}

	// Add summary information
	s.stateManager.AddSummary(envelope, "imagesFetched", true)
	s.stateManager.AddSummary(envelope, "verificationType", verificationContext.VerificationType)
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		s.stateManager.AddSummary(envelope, "layoutId", verificationContext.LayoutId)
	} else {
		s.stateManager.AddSummary(envelope, "previousVerificationId", verificationContext.PreviousVerificationId)
	}

	if envelope != nil && envelope.References != nil {
		delete(envelope.References, "images_reference_base64")
		delete(envelope.References, "images_checking_base64")
	}

	// Create and return response
	return &models.FetchImagesResponse{
		VerificationId: verificationContext.VerificationId,
		S3References:   envelope.References,
		Status:         schema.StatusImagesFetched,
		Summary:        envelope.Summary,
	}, nil
}

// InitializationData represents the structure stored by Initialize function
type InitializationData struct {
	SchemaVersion       string                      `json:"schemaVersion"`
	VerificationContext *schema.VerificationContext `json:"verificationContext"`
	SystemPrompt        map[string]interface{}      `json:"systemPrompt"`
	LayoutMetadata      interface{}                 `json:"layoutMetadata"`
}

// loadVerificationContext loads the verification context from either the envelope or the request
func (s *FetchService) loadVerificationContext(
	ctx context.Context,
	envelope *s3state.Envelope,
	request *models.FetchImagesRequest,
) (interface{}, error) {
	// If using S3 state manager, load from the state
	if ref := envelope.GetReference("processing_initialization"); ref != nil {
		var rawData interface{}
		err := s.stateManager.Manager().RetrieveJSON(ref, &rawData)
		if err != nil {
			s.logger.Error("Failed to load initialization data from S3", map[string]interface{}{
				"reference": ref,
				"error":     err.Error(),
			})
			// Fall back to request context if available
			if request.VerificationContext != nil {
				return request.VerificationContext, nil
			}
			return nil, err
		}

		// Try to parse as InitializationData structure first
		if dataMap, ok := rawData.(map[string]interface{}); ok {
			// Check if this is the new InitializationData format
			if schemaVersion, hasSchema := dataMap["schemaVersion"]; hasSchema {
				s.logger.Info("Found InitializationData with schema version", map[string]interface{}{
					"schemaVersion": schemaVersion,
				})
				
				// Extract verificationContext from InitializationData
				if vcData, hasVC := dataMap["verificationContext"]; hasVC {
					return vcData, nil
				} else {
					return nil, fmt.Errorf("verificationContext not found in InitializationData")
				}
			} else {
				// This might be legacy format - return as is
				s.logger.Info("Found legacy format initialization data", map[string]interface{}{
					"dataKeys": getMapKeys(dataMap),
				})
				return rawData, nil
			}
		}

		// If not a map, return as is (might be direct VerificationContext)
		return rawData, nil
	}

	// Fall back to legacy format if available
	if request.VerificationContext != nil {
		return request.VerificationContext, nil
	}

	return nil, fmt.Errorf("verification context not found in state or request")
}

// ParallelFetchResults holds all the fetched data
type ParallelFetchResults struct {
	ReferenceMeta     *schema.ImageInfo
	CheckingMeta      *schema.ImageInfo
	LayoutMeta        map[string]interface{}
	HistoricalContext map[string]interface{}
	Errors            []error
}

// fetchAllDataInParallel fetches all required data concurrently
func (s *FetchService) fetchAllDataInParallel(
	ctx context.Context,
	envelope *s3state.Envelope,
	referenceUrl string,
	checkingUrl string,
	layoutId int,
	layoutPrefix string,
	prevVerificationId string,
) (*ParallelFetchResults, error) {
	var wg sync.WaitGroup
	results := &ParallelFetchResults{
		LayoutMeta:        make(map[string]interface{}),
		HistoricalContext: make(map[string]interface{}),
		Errors:            []error{},
	}
	var mu sync.Mutex

	// Fetch reference image and Base64
	wg.Add(1)
	go func() {
		defer wg.Done()
		b64, meta, err := s.s3Repo.DownloadAndConvertToBase64(ctx, referenceUrl)
		var ref *s3state.Reference
		if err == nil {
			ref, err = s.stateManager.StoreBase64Image(envelope, "reference", b64)
			if err == nil {
				meta.Base64S3Bucket = ref.Bucket
				meta.SetBase64S3Key(ref.Key, "reference")
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch reference image metadata: %w", err))
			s.logger.Error("Failed to fetch reference image metadata", map[string]interface{}{
				"url":   referenceUrl,
				"error": err.Error(),
			})
		} else {
			results.ReferenceMeta = meta
			s.logger.Info("Successfully processed reference image", map[string]interface{}{
				"url":  referenceUrl,
				"size": meta.Size,
			})
		}
	}()

	// Fetch checking image and Base64
	wg.Add(1)
	go func() {
		defer wg.Done()
		b64, meta, err := s.s3Repo.DownloadAndConvertToBase64(ctx, checkingUrl)
		var ref *s3state.Reference
		if err == nil {
			ref, err = s.stateManager.StoreBase64Image(envelope, "checking", b64)
			if err == nil {
				meta.Base64S3Bucket = ref.Bucket
				meta.SetBase64S3Key(ref.Key, "checking")
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch checking image metadata: %w", err))
			s.logger.Error("Failed to fetch checking image metadata", map[string]interface{}{
				"url":   checkingUrl,
				"error": err.Error(),
			})
		} else {
			results.CheckingMeta = meta
			s.logger.Info("Successfully processed checking image", map[string]interface{}{
				"url":  checkingUrl,
				"size": meta.Size,
			})
		}
	}()

	// Fetch layout metadata if needed
	if layoutId != 0 && layoutPrefix != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// First verify the layout exists
			exists, err := s.dynamoDBRepo.ValidateLayoutExists(ctx, layoutId, layoutPrefix)
			if err != nil {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("failed to validate layout existence: %w", err))
				s.logger.Error("Failed to validate layout existence", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"error":        err.Error(),
				})
				mu.Unlock()
				return
			}

			if !exists {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("layout not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix))
				s.logger.Error("Layout not found", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
				})
				mu.Unlock()
				return
			}

			// Fetch the layout metadata
			layout, err := s.dynamoDBRepo.FetchLayoutMetadata(ctx, layoutId, layoutPrefix)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results.Errors = append(results.Errors, fmt.Errorf("failed to fetch layout metadata: %w", err))
				s.logger.Error("Failed to fetch layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"error":        err.Error(),
				})
			} else {
				// Convert layout to map
				layoutMap := make(map[string]interface{})
				layoutMap["layoutId"] = layout.LayoutId
				layoutMap["layoutPrefix"] = layout.LayoutPrefix
				layoutMap["vendingMachineId"] = layout.VendingMachineId
				layoutMap["location"] = layout.Location
				layoutMap["createdAt"] = layout.CreatedAt
				layoutMap["updatedAt"] = layout.UpdatedAt
				layoutMap["referenceImageUrl"] = layout.ReferenceImageUrl
				layoutMap["sourceJsonUrl"] = layout.SourceJsonUrl
				layoutMap["machineStructure"] = layout.MachineStructure
				layoutMap["productPositionMap"] = layout.ProductPositionMap

				results.LayoutMeta = layoutMap
				s.logger.Info("Successfully fetched layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
				})
			}
		}()
	} else {
		s.logger.Info("Skipping layout metadata fetch", map[string]interface{}{
			"reason": "layoutId or layoutPrefix not provided",
		})
	}

	// Fetch historical verification if needed
	if prevVerificationId != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			historicalContext, err := s.dynamoDBRepo.FetchHistoricalVerification(ctx, prevVerificationId)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results.Errors = append(results.Errors, fmt.Errorf("failed to fetch historical verification data: %w", err))
				s.logger.Error("Failed to fetch historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"error":                  err.Error(),
				})
			} else {
				results.HistoricalContext = historicalContext
				s.logger.Info("Successfully fetched historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
				})
			}
		}()
	} else {
		s.logger.Info("Skipping historical verification fetch", map[string]interface{}{
			"reason": "No previousVerificationId provided",
		})
	}

	// Wait for all fetches to complete
	wg.Wait()

	// Check for errors
	if len(results.Errors) > 0 {
		// Log all errors
		for i, err := range results.Errors {
			s.logger.Error("Error during parallel fetch", map[string]interface{}{
				"errorIndex": i,
				"error":      err.Error(),
			})
		}
		// Return the first error
		return nil, results.Errors[0]
	}

	// Ensure we got all the required data
	if results.ReferenceMeta == nil {
		return nil, fmt.Errorf("failed to fetch reference image metadata")
	}
	if results.CheckingMeta == nil {
		return nil, fmt.Errorf("failed to fetch checking image metadata")
	}

	// For LAYOUT_VS_CHECKING, ensure we have layout metadata
	if layoutId != 0 && layoutPrefix != "" && len(results.LayoutMeta) == 0 {
		return nil, fmt.Errorf("failed to fetch required layout metadata")
	}

	// For PREVIOUS_VS_CURRENT, ensure we have historical context
	if prevVerificationId != "" && len(results.HistoricalContext) == 0 {
		return nil, fmt.Errorf("failed to fetch required historical verification data")
	}

	return results, nil
}

// Helper functions for working with map values

// getStringValue extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getIntValue extracts an int value from a map
func getIntValue(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return 0
}

// getMapKeys returns the keys of a map as a slice of strings
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
