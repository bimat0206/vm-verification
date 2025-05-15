package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParallelFetch executes the S3 and DynamoDB fetches concurrently with enhanced error handling and validation.
func ParallelFetch(
	ctx context.Context,
	deps *Dependencies,
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

	log.Info("Starting parallel fetch operations", map[string]interface{}{
		"referenceUrl":         referenceUrl,
		"checkingUrl":          checkingUrl,
		"layoutId":             layoutId,
		"layoutPrefix":         layoutPrefix,
		"prevVerificationId":   prevVerificationId,
		"startTime":            startTime.Format(time.RFC3339),
	})

	// S3: Reference image metadata with validation
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		// Use the config-aware constructor with correct parameters
		s3Wrapper := NewS3Utils(deps.GetAWSConfig(), deps.GetLogger())
		
		// Validate image for Bedrock first
		if err := s3Wrapper.ValidateImageForBedrock(ctx, referenceUrl); err != nil {
			mu.Lock()
			results.Errors = append(results.Errors, fmt.Errorf("reference image validation failed: %w", err))
			log.Error("Reference image validation failed", map[string]interface{}{
				"url":   referenceUrl,
				"error": err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
			mu.Unlock()
			return
		}
		
		// Get metadata
		meta, err := s3Wrapper.GetS3ImageMetadata(ctx, referenceUrl)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch reference image metadata: %w", err))
			log.Error("Failed to fetch reference image metadata", map[string]interface{}{
				"url":      referenceUrl,
				"error":    err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
		} else {
			results.ReferenceMeta = meta
			log.Info("Successfully fetched reference image metadata", map[string]interface{}{
				"contentType":  meta.ContentType,
				"size":         meta.Size,
				"bucketOwner":  meta.BucketOwner,
				"etag":         meta.ETag,
				"duration":     time.Since(fetchStart).Milliseconds(),
			})
		}
	}()

	// S3: Checking image metadata with validation
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		// Use the config-aware constructor with correct parameters
		s3Wrapper := NewS3Utils(deps.GetAWSConfig(), deps.GetLogger())
		
		// Validate image for Bedrock first
		if err := s3Wrapper.ValidateImageForBedrock(ctx, checkingUrl); err != nil {
			mu.Lock()
			results.Errors = append(results.Errors, fmt.Errorf("checking image validation failed: %w", err))
			log.Error("Checking image validation failed", map[string]interface{}{
				"url":   checkingUrl,
				"error": err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
			mu.Unlock()
			return
		}
		
		// Get metadata
		meta, err := s3Wrapper.GetS3ImageMetadata(ctx, checkingUrl)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, fmt.Errorf("failed to fetch checking image metadata: %w", err))
			log.Error("Failed to fetch checking image metadata", map[string]interface{}{
				"url":      checkingUrl,
				"error":    err.Error(),
				"duration": time.Since(fetchStart).Milliseconds(),
			})
		} else {
			results.CheckingMeta = meta
			log.Info("Successfully fetched checking image metadata", map[string]interface{}{
				"contentType":  meta.ContentType,
				"size":         meta.Size,
				"bucketOwner":  meta.BucketOwner,
				"etag":         meta.ETag,
				"duration":     time.Since(fetchStart).Milliseconds(),
			})
		}
	}()

	// DynamoDB: Layout metadata (only for LAYOUT_VS_CHECKING)
	if layoutId != 0 && layoutPrefix != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchStart := time.Now()
			
			// Initialize the DB wrapper if needed
			var dbWrapper *DBUtilsWrapper
			if deps.GetDbUtils() != nil {
				dbWrapper = NewDBUtils(deps.GetDbUtils())
			} else {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("dbUtils not initialized for layout metadata fetch"))
				log.Error("dbUtils not initialized for layout metadata fetch", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"duration":     time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			// Validate layout exists first
			exists, err := dbWrapper.ValidateLayoutExists(ctx, layoutId, layoutPrefix)
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
			
			// Fetch the layout metadata with fallback
			meta, err := dbWrapper.FetchLayoutMetadataWithFallback(ctx, layoutId, layoutPrefix)
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
				// Log additional details about the layout
				logData := map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"duration":     time.Since(fetchStart).Milliseconds(),
				}
				
				// Add machine structure details if available
				if machineStruct, ok := meta["machineStructure"].(map[string]interface{}); ok {
					if rowCount, exists := machineStruct["rowCount"]; exists {
						logData["rowCount"] = rowCount
					}
					if columnsPerRow, exists := machineStruct["columnsPerRow"]; exists {
						logData["columnsPerRow"] = columnsPerRow
					}
				}
				
				// Add product count if available
				if totalPositions, exists := meta["totalProductPositions"]; exists {
					logData["totalProductPositions"] = totalPositions
				}
				
				log.Info("Successfully fetched layout metadata", logData)
			}
		}()
	} else {
		log.Info("Skipping layout metadata fetch", map[string]interface{}{
			"reason": "layoutId or layoutPrefix not provided (expected for PREVIOUS_VS_CURRENT verification)",
		})
	}

	// DynamoDB: Historical context (required for PREVIOUS_VS_CURRENT, skipped for LAYOUT_VS_CHECKING)
	if prevVerificationId != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchStart := time.Now()
			
			// Initialize the DB wrapper if needed
			var dbWrapper *DBUtilsWrapper
			if deps.GetDbUtils() != nil {
				dbWrapper = NewDBUtils(deps.GetDbUtils())
			} else {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("dbUtils not initialized for historical context fetch"))
				log.Error("dbUtils not initialized for historical context fetch", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"duration":               time.Since(fetchStart).Milliseconds(),
				})
				mu.Unlock()
				return
			}
			
			// Fetch historical context
			ctxObj, err := dbWrapper.FetchHistoricalContext(ctx, prevVerificationId)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errMsg := fmt.Sprintf("Failed to fetch historical verification data for ID %s: %v",
					prevVerificationId, err)
				results.Errors = append(results.Errors, fmt.Errorf(errMsg))
				log.Error("Failed to fetch historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"error":                  err.Error(),
					"errorType":              "HistoricalDataFetchError",
					"duration":               time.Since(fetchStart).Milliseconds(),
				})
			} else {
				results.HistoricalContext = ctxObj
				// Log details about the historical context
				logData := map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"duration":               time.Since(fetchStart).Milliseconds(),
				}
				
				// Add additional context details if available
				if verificationStatus, exists := ctxObj["previousVerificationStatus"]; exists {
					logData["verificationStatus"] = verificationStatus
				}
				if verificationAt, exists := ctxObj["previousVerificationAt"]; exists {
					logData["verificationAt"] = verificationAt
				}
				if hoursSince, exists := ctxObj["hoursSinceLastVerification"]; exists {
					logData["hoursSince"] = hoursSince
				}
				if vendingMachineId, exists := ctxObj["vendingMachineId"]; exists {
					logData["vendingMachineId"] = vendingMachineId
				}
				
				log.Info("Successfully fetched historical verification", logData)
			}
		}()
	} else {
		// Log that we're skipping historical context for non-PREVIOUS_VS_CURRENT verifications
		log.Info("Skipping historical context fetch", map[string]interface{}{
			"reason": "No previousVerificationId provided, this is expected for LAYOUT_VS_CHECKING verification type",
		})
	}

	// Wait for all operations to complete
	wg.Wait()
	
	// Log final results summary
	totalDuration := time.Since(startTime)
	log.Info("Parallel fetch operations completed", map[string]interface{}{
		"totalDuration":         totalDuration.Milliseconds(),
		"errorCount":            len(results.Errors),
		"referenceImageFetched": results.ReferenceMeta.Size > 0,
		"checkingImageFetched":  results.CheckingMeta.Size > 0,
		"layoutMetadataFetched": len(results.LayoutMeta) > 0,
		"historicalContextFetched": len(results.HistoricalContext) > 0,
		"totalReferenceSize":    results.ReferenceMeta.Size,
		"totalCheckingSize":     results.CheckingMeta.Size,
	})
	
	// Log individual errors for better debugging
	if len(results.Errors) > 0 {
		for i, err := range results.Errors {
			log.Error("Parallel fetch error", map[string]interface{}{
				"errorIndex": i,
				"error":      err.Error(),
			})
		}
	}
	
	return results
}
