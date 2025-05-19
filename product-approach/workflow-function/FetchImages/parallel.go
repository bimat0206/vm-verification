package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"workflow-function/shared/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ParallelFetch executes the S3 and DynamoDB fetches concurrently with dynamic Base64 response size management
func ParallelFetch(
	ctx context.Context,
	deps *Dependencies,
	config ConfigVars,
	referenceUrl string,
	checkingUrl string,
	layoutId int,
	layoutPrefix string,
	prevVerificationId string,
) ParallelFetchResults {
	startTime := time.Now()
	var wg sync.WaitGroup
	results := ParallelFetchResults{
		LayoutMeta:        make(map[string]interface{}),
		HistoricalContext: make(map[string]interface{}),
		Errors:            []error{},
	}
	var mu sync.Mutex
	log := deps.GetLogger()

	// Create shared response size tracker for coordinated storage decisions
	responseSizeTracker := NewResponseSizeTracker()

	log.Info("Starting parallel fetch operations with dynamic response size management", map[string]interface{}{
		"referenceUrl":          referenceUrl,
		"checkingUrl":           checkingUrl,
		"layoutId":              layoutId,
		"layoutPrefix":          layoutPrefix,
		"prevVerificationId":    prevVerificationId,
		"maxImageSize":          config.MaxImageSize,
		"maxInlineBase64Size":   config.MaxInlineBase64Size,
		"tempBase64Bucket":      config.TempBase64Bucket,
		"maxUsableResponseSize": MaxUsableResponseSize,
		"responseOverheadBuffer": ResponseOverheadBuffer,
		"startTime":             startTime.Format(time.RFC3339),
		"storageConfig":         config.GetStorageConfig(),
	})

	// Pre-estimate total Base64 size by getting image metadata first
	estimatedSizes := make(map[string]int64)
	preEstimateStart := time.Now()
	
	// Estimate reference image size
	wg.Add(1)
	go func() {
		defer wg.Done()
		if size, err := estimateImageBase64Size(ctx, deps, referenceUrl, "reference"); err == nil {
			mu.Lock()
			estimatedSizes["reference"] = size
			mu.Unlock()
		}
	}()

	// Estimate checking image size
	wg.Add(1)
	go func() {
		defer wg.Done()
		if size, err := estimateImageBase64Size(ctx, deps, checkingUrl, "checking"); err == nil {
			mu.Lock()
			estimatedSizes["checking"] = size
			mu.Unlock()
		}
	}()

	wg.Wait() // Wait for size estimations
	
	// Calculate total estimated size and inform the tracker
	mu.Lock()
	totalEstimated := estimatedSizes["reference"] + estimatedSizes["checking"]
	mu.Unlock()
	
	responseSizeTracker.SetEstimatedTotal(totalEstimated)
	
	log.Info("Pre-estimation completed", map[string]interface{}{
		"referenceEstimate":     estimatedSizes["reference"],
		"checkingEstimate":      estimatedSizes["checking"],
		"totalEstimated":        totalEstimated,
		"maxUsableSize":         MaxUsableResponseSize,
		"estimationDuration":    time.Since(preEstimateStart).Milliseconds(),
		"requiresS3Storage":     totalEstimated > MaxUsableResponseSize,
	})

	// S3: Reference image with dynamic response size-aware Base64 encoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		
		// Get S3 client with configuration
		s3Client := deps.GetS3Client()
		
		// Log initial storage configuration with response size context
		log.Debug("Processing reference image with dynamic response size awareness", map[string]interface{}{
			"url":                    referenceUrl,
			"estimatedBase64Size":    estimatedSizes["reference"],
			"totalEstimatedSize":     totalEstimated,
			"maxUsableResponseSize":  MaxUsableResponseSize,
			"storageConfig":          config.GetStorageConfig(),
		})
		
		// Download and encode image with direct S3 client
		meta, err := GetS3ImageWithBase64(ctx, s3Client, referenceUrl, config, responseSizeTracker, log)
		
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch reference image: %w", err))
			log.Error("Failed to fetch reference image", map[string]interface{}{
				"url":      referenceUrl,
				"error":    err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
		} else {
			results.ReferenceMeta = meta
			
			// Log successful fetch with dynamic storage details
			log.Info("Successfully fetched and encoded reference image", map[string]interface{}{
				"contentType":           meta.ContentType,
				"size":                  meta.Size,
				"imageFormat":           meta.ImageFormat,
				"storageMethod":         meta.StorageMethod,
				"hasInlineData":         meta.HasInlineData(),
				"hasS3Storage":          meta.HasS3Storage(),
				"storageInfo":           meta.GetStorageInfo(),
				"responseUtilization":   float64(responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
				"currentTotalBase64":    responseSizeTracker.GetTotalSize(),
				"estimatedSize":         estimatedSizes["reference"],
				"actualBase64Size":      getActualBase64Size(meta),
				"duration":              time.Since(fetchStart).Milliseconds(),
			})
			
			// Additional logging for S3 storage with size reasoning
			if meta.StorageMethod == StorageMethodS3Temporary {
				log.Info("Reference image stored in S3 due to response size optimization", map[string]interface{}{
					"bucket":                 meta.Base64S3Bucket,
					"key":                    meta.Base64S3Key,
					"reason":                 getStorageReason(meta, responseSizeTracker, estimatedSizes["reference"]),
					"totalResponseSizeAfter": responseSizeTracker.GetTotalSize(),
				})
			}
		}
	}()

	// S3: Checking image with dynamic response size-aware Base64 encoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		
		// Get S3 client with configuration
		s3Client := deps.GetS3Client()
		
		// Log initial storage configuration with response size context
		log.Debug("Processing checking image with dynamic response size awareness", map[string]interface{}{
			"url":                    checkingUrl,
			"estimatedBase64Size":    estimatedSizes["checking"],
			"totalEstimatedSize":     totalEstimated,
			"currentResponseSize":    responseSizeTracker.GetTotalSize(),
			"maxUsableResponseSize":  MaxUsableResponseSize,
			"storageConfig":          config.GetStorageConfig(),
		})
		
		// Download and encode image with direct S3 client
		meta, err := GetS3ImageWithBase64(ctx, s3Client, checkingUrl, config, responseSizeTracker, log)
		
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch checking image: %w", err))
			log.Error("Failed to fetch checking image", map[string]interface{}{
				"url":      checkingUrl,
				"error":    err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
		} else {
			results.CheckingMeta = meta
			
			// Log successful fetch with dynamic storage details
			log.Info("Successfully fetched and encoded checking image", map[string]interface{}{
				"contentType":           meta.ContentType,
				"size":                  meta.Size,
				"imageFormat":           meta.ImageFormat,
				"storageMethod":         meta.StorageMethod,
				"hasInlineData":         meta.HasInlineData(),
				"hasS3Storage":          meta.HasS3Storage(),
				"storageInfo":           meta.GetStorageInfo(),
				"responseUtilization":   float64(responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
				"currentTotalBase64":    responseSizeTracker.GetTotalSize(),
				"estimatedSize":         estimatedSizes["checking"],
				"actualBase64Size":      getActualBase64Size(meta),
				"duration":              time.Since(fetchStart).Milliseconds(),
			})
			
			// Additional logging for S3 storage with size reasoning
			if meta.StorageMethod == StorageMethodS3Temporary {
				log.Info("Checking image stored in S3 due to response size optimization", map[string]interface{}{
					"bucket":                 meta.Base64S3Bucket,
					"key":                    meta.Base64S3Key,
					"reason":                 getStorageReason(meta, responseSizeTracker, estimatedSizes["checking"]),
					"totalResponseSizeAfter": responseSizeTracker.GetTotalSize(),
				})
			}
		}
	}()

	// DynamoDB: Layout metadata (only for LAYOUT_VS_CHECKING)
	if layoutId != 0 && layoutPrefix != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchStart := time.Now()
			
			// Get DynamoDB client and table name
			dbClient := deps.GetDBClient()
			layoutTable := deps.GetLayoutTable()
			
			if dbClient == nil {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("DynamoDB client not initialized for layout metadata fetch"))
				log.Error("DynamoDB client not initialized for layout metadata fetch", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			// Validate layout exists first
			exists, err := ValidateLayoutExists(ctx, dbClient, layoutTable, layoutId, layoutPrefix)
			if err != nil {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("failed to validate layout existence: %w", err))
				log.Error("Failed to validate layout existence", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"error":        err.Error(),
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			if !exists {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("layout not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix))
				log.Error("Layout not found in DynamoDB", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			// Fetch the layout metadata
			meta, err := FetchLayoutMetadataWithFallback(ctx, dbClient, layoutTable, layoutId, layoutPrefix)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results.Errors = append(results.Errors, fmt.Errorf("failed to fetch layout metadata: %w", err))
				log.Error("Failed to fetch layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"error":        err.Error(),
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
			} else {
				results.LayoutMeta = meta
				log.Info("Successfully fetched layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
			}
		}()
	} else {
		log.Info("Skipping layout metadata fetch", map[string]interface{}{
			"reason": "layoutId or layoutPrefix not provided (expected for PREVIOUS_VS_CURRENT verification)",
		})
	}

	// DynamoDB: Historical context (required for PREVIOUS_VS_CURRENT)
	if prevVerificationId != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchStart := time.Now()
			
			// Get DynamoDB client and table name
			dbClient := deps.GetDBClient()
			verificationTable := deps.GetVerificationTable()
			
			if dbClient == nil {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("DynamoDB client not initialized for historical context fetch"))
				log.Error("DynamoDB client not initialized for historical context fetch", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"duration":               time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			// Fetch historical context
			ctxObj, err := FetchHistoricalContext(ctx, dbClient, verificationTable, prevVerificationId)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results.Errors = append(results.Errors, fmt.Errorf("failed to fetch historical verification data for ID %s: %w",
					prevVerificationId, err))
				log.Error("Failed to fetch historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"error":                  err.Error(),
					"duration":               time.Since(fetchStart).Milliseconds(),
				})
			} else {
				results.HistoricalContext = ctxObj
				log.Info("Successfully fetched historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"duration":               time.Since(fetchStart).Milliseconds(),
				})
			}
		}()
	} else {
		log.Info("Skipping historical context fetch", map[string]interface{}{
			"reason": "No previousVerificationId provided, expected for LAYOUT_VS_CHECKING verification type",
		})
	}

	// Wait for all operations to complete
	wg.Wait()
	
	// Perform post-processing response size optimization if needed
	if len(results.Errors) == 0 {
		optimized, err := performPostProcessingOptimization(ctx, deps, &results, responseSizeTracker, log)
		if err != nil {
			log.Warn("Post-processing optimization failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else if optimized {
			log.Info("Post-processing optimization applied successfully", map[string]interface{}{
				"finalResponseSize": responseSizeTracker.GetTotalSize(),
				"finalUtilization":  float64(responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
			})
		}
	}
	
	// Log final results summary with dynamic response size metrics
	totalDuration := time.Since(startTime)
	logFinalResultsWithResponseSizeMetrics(results, responseSizeTracker, totalDuration, estimatedSizes, log)
	
	// Log individual errors for better debugging
	if len(results.Errors) > 0 {
		for i, err := range results.Errors {
			log.Error("Parallel fetch error", map[string]interface{}{
				"errorIndex": i,
				"error":      err.Error(),
			})
		}
	}
	
	// Validate hybrid storage integrity if no errors
	if len(results.Errors) == 0 {
		validationErrors := validateDynamicStorageIntegrity(results, responseSizeTracker, log)
		if len(validationErrors) > 0 {
			results.Errors = append(results.Errors, validationErrors...)
		}
	}
	
	return results
}

// estimateImageBase64Size estimates the Base64 size of an image without downloading it
func estimateImageBase64Size(ctx context.Context, deps *Dependencies, s3url, imageType string) (int64, error) {
	log := deps.GetLogger()
	
	// Parse S3 URL
	parsed, err := ParseS3URL(s3url)
	if err != nil {
		return 0, fmt.Errorf("failed to parse S3 URL for %s image: %w", imageType, err)
	}
	
	// Get object metadata
	s3Client := deps.GetS3Client()
	headOutput, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get %s image metadata: %w", imageType, err)
	}
	
	if headOutput.ContentLength == nil {
		return 0, fmt.Errorf("unable to determine %s image size", imageType)
	}
	
	originalSize := *headOutput.ContentLength
	estimatedBase64Size := int64(float64(originalSize) * Base64ExpansionFactor)
	
	log.Debug("Estimated Base64 size", map[string]interface{}{
		"imageType":           imageType,
		"originalSize":        originalSize,
		"estimatedBase64Size": estimatedBase64Size,
		"expansionFactor":     Base64ExpansionFactor,
	})
	
	return estimatedBase64Size, nil
}

// getActualBase64Size returns the actual Base64 size for an image metadata
func getActualBase64Size(meta ImageMetadata) int64 {
	if meta.HasInlineData() {
		return int64(len(meta.Base64Data))
	}
	return 0 // S3-stored data doesn't count toward response size
}

// getStorageReason provides a human-readable reason for the storage method choice
func getStorageReason(meta ImageMetadata, tracker *ResponseSizeTracker, estimatedSize int64) string {
	if meta.StorageMethod == StorageMethodInline {
		return "Image size within inline threshold and total response size acceptable"
	}
	
	actualSize := getActualBase64Size(meta)
	if actualSize == 0 {
		actualSize = estimatedSize // Use estimated if actual not available
	}
	
	currentTotal := tracker.GetTotalSize()
	wouldBeTotal := currentTotal + actualSize
	
	if wouldBeTotal > MaxUsableResponseSize {
		return fmt.Sprintf("Total response size would exceed limit (would be %d, max %d)", 
			wouldBeTotal, MaxUsableResponseSize)
	}
	
	return "Image exceeds individual inline threshold"
}

// performPostProcessingOptimization performs response size optimization after all images are processed
func performPostProcessingOptimization(
	ctx context.Context,
	deps *Dependencies,
	results *ParallelFetchResults,
	tracker *ResponseSizeTracker,
	log logger.Logger,
) (bool, error) {
	currentSize := tracker.GetTotalSize()
	
	// Check if we're within acceptable range
	if currentSize <= MaxUsableResponseSize {
		return false, nil // No optimization needed
	}
	
	log.Warn("Response size exceeds limit, attempting post-processing optimization", map[string]interface{}{
		"currentSize":         currentSize,
		"maxUsableSize":       MaxUsableResponseSize,
		"excessSize":          currentSize - MaxUsableResponseSize,
		"utilizationPercent":  float64(currentSize) / float64(MaxUsableResponseSize) * 100,
	})
	
	optimized := false
	
	// Try to convert the largest inline image to S3 storage
	if results.ReferenceMeta.HasInlineData() && results.CheckingMeta.HasInlineData() {
		// Convert the larger of the two to S3
		refSize := int64(len(results.ReferenceMeta.Base64Data))
		checkSize := int64(len(results.CheckingMeta.Base64Data))
		
		if refSize >= checkSize {
			log.Info("Converting reference image to S3 storage for optimization", map[string]interface{}{
				"referenceBase64Size": refSize,
				"checkingBase64Size":  checkSize,
			})
			if err := convertImageToS3Storage(ctx, deps, &results.ReferenceMeta, tracker, log); err != nil {
				return false, fmt.Errorf("failed to convert reference image to S3: %w", err)
			}
			optimized = true
		} else {
			log.Info("Converting checking image to S3 storage for optimization", map[string]interface{}{
				"referenceBase64Size": refSize,
				"checkingBase64Size":  checkSize,
			})
			if err := convertImageToS3Storage(ctx, deps, &results.CheckingMeta, tracker, log); err != nil {
				return false, fmt.Errorf("failed to convert checking image to S3: %w", err)
			}
			optimized = true
		}
	} else if results.ReferenceMeta.HasInlineData() {
		// Only reference is inline, convert it
		log.Info("Converting reference image to S3 storage for optimization", map[string]interface{}{
			"reason": "Post-processing optimization",
			"method": "reference",
		})
		if err := convertImageToS3Storage(ctx, deps, &results.ReferenceMeta, tracker, log); err != nil {
			return false, fmt.Errorf("failed to convert reference image to S3: %w", err)
		}
		optimized = true
	} else if results.CheckingMeta.HasInlineData() {
		// Only checking is inline, convert it
		log.Info("Converting checking image to S3 storage for optimization", map[string]interface{}{
			"reason": "Post-processing optimization",
			"method": "checking",
		})
		if err := convertImageToS3Storage(ctx, deps, &results.CheckingMeta, tracker, log); err != nil {
			return false, fmt.Errorf("failed to convert checking image to S3: %w", err)
		}
		optimized = true
	}
	
	return optimized, nil
}

// convertImageToS3Storage converts an inline-stored image to S3 storage
func convertImageToS3Storage(
	ctx context.Context,
	deps *Dependencies,
	meta *ImageMetadata,
	tracker *ResponseSizeTracker,
	log logger.Logger,
) error {
	config := deps.GetConfig()
	s3Client := deps.GetS3Client()
	
	return ConvertImageToS3Storage(ctx, s3Client, config, meta, tracker, log)
}

// validateDynamicStorageIntegrity validates storage integrity with response size awareness
func validateDynamicStorageIntegrity(
	results ParallelFetchResults,
	tracker *ResponseSizeTracker,
	log logger.Logger,
) []error {
	var errors []error
	
	// Standard validation
	errors = append(errors, validateHybridStorageIntegrity(results, log)...)
	
	// Additional response size validation
	currentSize := tracker.GetTotalSize()
	if currentSize > MaxUsableResponseSize {
		err := fmt.Errorf("final response size exceeds Lambda limit: %d bytes (max %d bytes)", 
			currentSize, MaxUsableResponseSize)
		errors = append(errors, err)
		log.Error("Response size validation failed", map[string]interface{}{
			"currentSize":       currentSize,
			"maxUsableSize":     MaxUsableResponseSize,
			"excessSize":        currentSize - MaxUsableResponseSize,
			"utilizationPercent": float64(currentSize) / float64(MaxUsableResponseSize) * 100,
		})
	}
	
	// Validate storage method distribution makes sense
	inlineCount := 0
	s3Count := 0
	
	if results.ReferenceMeta.StorageMethod == StorageMethodInline {
		inlineCount++
	} else if results.ReferenceMeta.StorageMethod == StorageMethodS3Temporary {
		s3Count++
	}
	
	if results.CheckingMeta.StorageMethod == StorageMethodInline {
		inlineCount++
	} else if results.CheckingMeta.StorageMethod == StorageMethodS3Temporary {
		s3Count++
	}
	
	log.Info("Dynamic storage integrity validation completed", map[string]interface{}{
		"inlineCount":           inlineCount,
		"s3Count":               s3Count,
		"currentResponseSize":   currentSize,
		"maxUsableResponseSize": MaxUsableResponseSize,
		"withinSizeLimit":       currentSize <= MaxUsableResponseSize,
		"utilizationPercent":    float64(currentSize) / float64(MaxUsableResponseSize) * 100,
	})
	
	return errors
}

// logFinalResultsWithResponseSizeMetrics logs comprehensive final results with response size analysis
func logFinalResultsWithResponseSizeMetrics(
	results ParallelFetchResults,
	tracker *ResponseSizeTracker,
	totalDuration time.Duration,
	estimatedSizes map[string]int64,
	log logger.Logger,
) {
	responseInfo := map[string]interface{}{
		"currentSize":           tracker.GetTotalSize(),
		"maxUsableSize":         MaxUsableResponseSize,
		"utilizationPercent":    float64(tracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
		"withinSizeLimit":       tracker.GetTotalSize() <= MaxUsableResponseSize,
		"referenceBase64Size":   tracker.referenceBase64Size,
		"checkingBase64Size":    tracker.checkingBase64Size,
		"overhead":             ResponseOverheadBuffer,
	}
	
	storageMethodStats := getDetailedStorageMethodStats(results)
	
	estimationAccuracy := map[string]interface{}{
		"referenceEstimated": estimatedSizes["reference"],
		"referenceActual":    getActualBase64Size(results.ReferenceMeta),
		"checkingEstimated":  estimatedSizes["checking"],
		"checkingActual":     getActualBase64Size(results.CheckingMeta),
		"totalEstimated":     estimatedSizes["reference"] + estimatedSizes["checking"],
		"totalActual":        tracker.GetTotalSize(),
	}
	
	// Calculate estimation accuracy
	if estimatedSizes["reference"] > 0 {
		refAccuracy := (1.0 - float64(abs64(estimatedSizes["reference"]-getActualBase64Size(results.ReferenceMeta)))/float64(estimatedSizes["reference"])) * 100
		estimationAccuracy["referenceAccuracy"] = refAccuracy
	}
	
	if estimatedSizes["checking"] > 0 {
		checkAccuracy := (1.0 - float64(abs64(estimatedSizes["checking"]-getActualBase64Size(results.CheckingMeta)))/float64(estimatedSizes["checking"])) * 100
		estimationAccuracy["checkingAccuracy"] = checkAccuracy
	}
	
	log.Info("Parallel fetch completed with dynamic response size management", map[string]interface{}{
		"totalDuration":            totalDuration.Milliseconds(),
		"errorCount":               len(results.Errors),
		"referenceImageFetched":    results.ReferenceMeta.HasBase64Data(),
		"checkingImageFetched":     results.CheckingMeta.HasBase64Data(),
		"layoutMetadataFetched":    len(results.LayoutMeta) > 0,
		"historicalContextFetched": len(results.HistoricalContext) > 0,
		"referenceSize":            results.ReferenceMeta.Size,
		"checkingSize":             results.CheckingMeta.Size,
		"referenceFormat":          results.ReferenceMeta.ImageFormat,
		"checkingFormat":           results.CheckingMeta.ImageFormat,
		"referenceStorageMethod":   results.ReferenceMeta.StorageMethod,
		"checkingStorageMethod":    results.CheckingMeta.StorageMethod,
		"responseSize":             responseInfo,
		"storageMethodStats":       storageMethodStats,
		"estimationAccuracy":       estimationAccuracy,
		"dynamicOptimizationUsed":  tracker.GetTotalSize() < estimatedSizes["reference"]+estimatedSizes["checking"],
	})
}
