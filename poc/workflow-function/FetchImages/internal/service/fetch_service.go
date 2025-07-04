// Package service provides business logic implementations for the FetchImages function
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"workflow-function/shared/errors"
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

// ProcessRequest orchestrates the fetching of image metadata and related data with comprehensive error logging
func (s *FetchService) ProcessRequest(
	ctx context.Context,
	request *models.FetchImagesRequest,
) (*models.FetchImagesResponse, error) {
	processingStartTime := time.Now()
	correlationID := fmt.Sprintf("fetch-service-%d", processingStartTime.UnixNano())

	s.logger.Info("Processing FetchImages request", map[string]interface{}{
		"verificationId": request.VerificationId,
		"correlationId":  correlationID,
		"timestamp":      processingStartTime.Format(time.RFC3339),
	})

	// Load or create envelope with enhanced error handling
	envelope, err := s.stateManager.LoadEnvelope(request)
	if err != nil {
		enhancedErr := errors.NewInternalError("FetchService", err).
			WithCorrelationID(correlationID).
			WithContext("operation", "LoadEnvelope").
			WithContext("verificationId", request.VerificationId)

		s.logger.Error("Failed to load state envelope", map[string]interface{}{
			"error":          err.Error(),
			"correlationId":  correlationID,
			"verificationId": request.VerificationId,
		})
		return nil, enhancedErr
	}

	// Load verification context with enhanced error handling
	context, err := s.loadVerificationContext(ctx, envelope, request)
	if err != nil {
		enhancedErr := errors.NewInternalError("FetchService", err).
			WithCorrelationID(correlationID).
			WithContext("operation", "LoadVerificationContext").
			WithContext("verificationId", request.VerificationId)

		s.logger.Error("Failed to load verification context", map[string]interface{}{
			"error":          err.Error(),
			"correlationId":  correlationID,
			"verificationId": request.VerificationId,
		})
		return nil, enhancedErr
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
		s.logger.Info("Extracting verification context from map", map[string]interface{}{
			"mapKeys":                   getMapKeys(v),
			"previousVerificationId":    getStringValue(v, "previousVerificationId"),
			"rawPreviousVerificationId": v["previousVerificationId"],
		})

		// Try JSON marshaling/unmarshaling approach first for better type conversion
		jsonBytes, err := json.Marshal(v)
		if err == nil {
			s.logger.Info("JSON marshaling successful", map[string]interface{}{
				"jsonString": string(jsonBytes),
			})
			
			var tempVC schema.VerificationContext
			if err := json.Unmarshal(jsonBytes, &tempVC); err == nil {
				verificationContext = &tempVC
				s.logger.Info("Successfully converted via JSON marshaling", map[string]interface{}{
					"verificationId":         verificationContext.VerificationId,
					"verificationType":       verificationContext.VerificationType,
					"previousVerificationId": verificationContext.PreviousVerificationId,
				})
			} else {
				s.logger.Error("JSON unmarshaling failed, falling back to manual extraction", map[string]interface{}{
					"error":      err.Error(),
					"jsonString": string(jsonBytes),
				})
				// Fall back to manual extraction
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
			}
		} else {
			s.logger.Error("JSON marshaling failed, using manual extraction", map[string]interface{}{
				"error": err.Error(),
			})
			// Fall back to manual extraction
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
		}

		s.logger.Info("Final extracted verification context", map[string]interface{}{
			"verificationId":         verificationContext.VerificationId,
			"verificationType":       verificationContext.VerificationType,
			"previousVerificationId": verificationContext.PreviousVerificationId,
			"isEmpty":                verificationContext.PreviousVerificationId == "",
		})
	default:
		enhancedErr := errors.NewValidationError("Unsupported verification context type", map[string]interface{}{
			"correlationId":    correlationID,
			"verificationId":   request.VerificationId,
			"receivedType":     fmt.Sprintf("%T", context),
			"expectedType":     "schema.VerificationContext",
		}).WithCorrelationID(correlationID).WithComponent("FetchService")

		s.logger.Error("Unsupported verification context type", map[string]interface{}{
			"correlationId":  correlationID,
			"receivedType":   fmt.Sprintf("%T", context),
			"expectedType":   "schema.VerificationContext",
		})
		return nil, enhancedErr
	}

	// Set status to IMAGES_FETCHED
	verificationContext.Status = schema.StatusImagesFetched

	// Extract sourceType and historicalDataFound from raw verification context data
	var sourceType string
	var historicalDataFound bool

	// Try to extract these fields from the raw verification context data
	if rawVCData, ok := context.(map[string]interface{}); ok {
		if st, exists := rawVCData["sourceType"]; exists {
			if stStr, ok := st.(string); ok {
				sourceType = stStr
			}
		}
		if hdf, exists := rawVCData["historicalDataFound"]; exists {
			if hdfBool, ok := hdf.(bool); ok {
				historicalDataFound = hdfBool
			}
		}
	}

	// Determine if we need layout or historical data based on verification type
	var prevVerificationId string
	if verificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// Add comprehensive debugging before validation
		s.logger.Info("DEBUGGING: About to validate previousVerificationId", map[string]interface{}{
			"verificationType":                    verificationContext.VerificationType,
			"expectedType":                       schema.VerificationTypePreviousVsCurrent,
			"typesMatch":                         verificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent,
			"previousVerificationId":             verificationContext.PreviousVerificationId,
			"previousVerificationIdLength":       len(verificationContext.PreviousVerificationId),
			"previousVerificationIdIsEmpty":      verificationContext.PreviousVerificationId == "",
			"sourceType":                         sourceType,
			"historicalDataFound":                historicalDataFound,
			"fullVerificationContext":            fmt.Sprintf("%+v", verificationContext),
		})

		// Check if we should bypass previousVerificationId requirement
		// Allow bypass when sourceType is "NO_HISTORICAL_DATA" or historicalDataFound is false
		shouldBypassPreviousVerificationId := sourceType == "NO_HISTORICAL_DATA" || !historicalDataFound

		if shouldBypassPreviousVerificationId {
			s.logger.Info("BYPASSING previousVerificationId validation", map[string]interface{}{
				"reason":               "No historical data available",
				"sourceType":           sourceType,
				"historicalDataFound":  historicalDataFound,
				"verificationId":       verificationContext.VerificationId,
			})
			// Set empty string to indicate no historical verification needed
			prevVerificationId = ""
		} else {
			// Make sure previousVerificationId exists for PREVIOUS_VS_CURRENT when historical data is expected
			if verificationContext.PreviousVerificationId == "" {
				enhancedErr := errors.NewMissingFieldError("previousVerificationId").
					WithCorrelationID(correlationID).
					WithComponent("FetchService").
					WithContext("verificationType", verificationContext.VerificationType).
					WithContext("verificationId", verificationContext.VerificationId).
					WithContext("sourceType", sourceType).
					WithContext("historicalDataFound", historicalDataFound)

				s.logger.Error("VALIDATION FAILED: previousVerificationId is empty", map[string]interface{}{
					"correlationId":              correlationID,
					"verificationType":           verificationContext.VerificationType,
					"verificationId":             verificationContext.VerificationId,
					"previousVerificationId":     verificationContext.PreviousVerificationId,
					"previousVerificationIdLen":  len(verificationContext.PreviousVerificationId),
					"sourceType":                 sourceType,
					"historicalDataFound":        historicalDataFound,
					"allFields":                  fmt.Sprintf("%+v", verificationContext),
				})
				return nil, enhancedErr
			}
			prevVerificationId = verificationContext.PreviousVerificationId
			s.logger.Info("VALIDATION PASSED: previousVerificationId found", map[string]interface{}{
				"previousVerificationId": prevVerificationId,
			})
		}
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
		enhancedErr := errors.NewInternalError("FetchService", err).
			WithCorrelationID(correlationID).
			WithContext("operation", "FetchAllDataInParallel").
			WithContext("verificationId", verificationContext.VerificationId).
			WithContext("verificationType", verificationContext.VerificationType)

		s.logger.Error("Failed to fetch required data", map[string]interface{}{
			"error":            err.Error(),
			"correlationId":    correlationID,
			"verificationId":   verificationContext.VerificationId,
			"verificationType": verificationContext.VerificationType,
		})
		return nil, enhancedErr
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
		enhancedErr := errors.NewInternalError("FetchService", err).
			WithCorrelationID(correlationID).
			WithContext("operation", "StoreImageMetadata").
			WithContext("verificationId", verificationContext.VerificationId)

		s.logger.Error("Failed to store image metadata", map[string]interface{}{
			"error":          err.Error(),
			"correlationId":  correlationID,
			"verificationId": verificationContext.VerificationId,
		})
		return nil, enhancedErr
	}

	// Store layout metadata if available
	if len(results.LayoutMeta) > 0 {
		if err := s.stateManager.StoreLayoutMetadata(envelope, results.LayoutMeta); err != nil {
			enhancedErr := errors.NewInternalError("FetchService", err).
				WithCorrelationID(correlationID).
				WithContext("operation", "StoreLayoutMetadata").
				WithContext("verificationId", verificationContext.VerificationId)

			s.logger.Error("Failed to store layout metadata", map[string]interface{}{
				"error":          err.Error(),
				"correlationId":  correlationID,
				"verificationId": verificationContext.VerificationId,
			})
			return nil, enhancedErr
		}
	}

	// Store historical context if available
	if len(results.HistoricalContext) > 0 {
		if err := s.stateManager.StoreHistoricalContext(envelope, results.HistoricalContext); err != nil {
			enhancedErr := errors.NewInternalError("FetchService", err).
				WithCorrelationID(correlationID).
				WithContext("operation", "StoreHistoricalContext").
				WithContext("verificationId", verificationContext.VerificationId)

			s.logger.Error("Failed to store historical context", map[string]interface{}{
				"error":          err.Error(),
				"correlationId":  correlationID,
				"verificationId": verificationContext.VerificationId,
			})
			return nil, enhancedErr
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

	// Create and return response with success metrics
	processingTime := time.Since(processingStartTime)
	s.logger.Info("Successfully completed FetchImages processing", map[string]interface{}{
		"verificationId":   verificationContext.VerificationId,
		"correlationId":    correlationID,
		"processingTimeMs": processingTime.Milliseconds(),
		"status":           schema.StatusImagesFetched,
		"referenceCount":   len(envelope.References),
		"timestamp":        time.Now().Format(time.RFC3339),
	})

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

		s.logger.Error("DEBUGGING: Raw data loaded from S3", map[string]interface{}{
			"rawDataType": fmt.Sprintf("%T", rawData),
			"rawData":     fmt.Sprintf("%+v", rawData),
		})

		// Try to parse as InitializationData structure first
		if dataMap, ok := rawData.(map[string]interface{}); ok {
			s.logger.Error("DEBUGGING: Data is a map", map[string]interface{}{
				"mapKeys": getMapKeys(dataMap),
				"mapSize": len(dataMap),
			})

			// Check if this is the new InitializationData format
			if schemaVersion, hasSchema := dataMap["schemaVersion"]; hasSchema {
				s.logger.Info("Found InitializationData with schema version", map[string]interface{}{
					"schemaVersion": schemaVersion,
				})

				// Extract verificationContext from InitializationData
				if vcData, hasVC := dataMap["verificationContext"]; hasVC {
					s.logger.Error("DEBUGGING: Found verificationContext in InitializationData", map[string]interface{}{
						"vcDataType": fmt.Sprintf("%T", vcData),
						"vcData":     fmt.Sprintf("%+v", vcData),
					})
					
					// If vcData is a map, check for previousVerificationId
					if vcMap, isMap := vcData.(map[string]interface{}); isMap {
						s.logger.Error("DEBUGGING: verificationContext is a map", map[string]interface{}{
							"vcMapKeys":                  getMapKeys(vcMap),
							"previousVerificationId":     vcMap["previousVerificationId"],
							"previousVerificationIdType": fmt.Sprintf("%T", vcMap["previousVerificationId"]),
						})
					}
					
					return vcData, nil
				} else {
					return nil, fmt.Errorf("verificationContext not found in InitializationData")
				}
			} else {
				// This might be legacy format - return as is
				s.logger.Info("Found legacy format initialization data", map[string]interface{}{
					"dataKeys": getMapKeys(dataMap),
				})
				
				// Check if this legacy format has previousVerificationId
				if prevId, hasPrevId := dataMap["previousVerificationId"]; hasPrevId {
					s.logger.Error("DEBUGGING: Found previousVerificationId in legacy format", map[string]interface{}{
						"previousVerificationId":     prevId,
						"previousVerificationIdType": fmt.Sprintf("%T", prevId),
					})
				}
				
				return rawData, nil
			}
		}

		// If not a map, return as is (might be direct VerificationContext)
		s.logger.Error("DEBUGGING: Raw data is not a map", map[string]interface{}{
			"rawDataType": fmt.Sprintf("%T", rawData),
		})
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
			
			// Add extensive debugging before the fetch
			s.logger.Info("Starting historical verification fetch", map[string]interface{}{
				"previousVerificationId": prevVerificationId,
				"previousVerificationIdType": fmt.Sprintf("%T", prevVerificationId),
				"previousVerificationIdLength": len(prevVerificationId),
				"tableName": s.dynamoDBRepo.GetTableName(),
			})
			
			// Perform the fetch with enhanced error diagnostics
			historicalContext, err := s.dynamoDBRepo.FetchHistoricalVerification(ctx, prevVerificationId)
			mu.Lock()
			defer mu.Unlock()
			
			if err != nil {
				// Enhanced error logging with more context
				errorDetails := map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"error":                  err.Error(),
					"errorType":              fmt.Sprintf("%T", err),
				}
				
				results.Errors = append(results.Errors, fmt.Errorf("failed to fetch historical verification data: %w", err))
				s.logger.Error("Failed to fetch historical verification", errorDetails)
			} else {
				// Add more details about the fetched data
				s.logger.Info("Successfully fetched historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"dataKeys": getMapKeys(historicalContext),
					"verificationAt": historicalContext["verificationAt"],
					"dataSize": len(fmt.Sprintf("%v", historicalContext)),
				})
				results.HistoricalContext = historicalContext
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
	val, exists := m[key]
	if !exists {
		return ""
	}

	// Handle different types that might be stored as strings
	switch v := val.(type) {
	case string:
		return v
	case nil:
		return ""
	case *string:
		if v != nil {
			return *v
		}
		return ""
	default:
		// Try to convert to string - this should preserve the actual value
		str := fmt.Sprintf("%v", v)
		// Don't return empty representations like "<nil>" or "0"
		if str == "<nil>" {
			return ""
		}
		return str
	}
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
