package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParallelFetch executes the S3 and DynamoDB fetches concurrently with Base64 encoding
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

	log.Info("Starting parallel fetch operations with Base64 encoding", map[string]interface{}{
		"referenceUrl":         referenceUrl,
		"checkingUrl":          checkingUrl,
		"layoutId":             layoutId,
		"layoutPrefix":         layoutPrefix,
		"prevVerificationId":   prevVerificationId,
		"maxImageSize":         config.MaxImageSize,
		"startTime":            startTime.Format(time.RFC3339),
	})

	// S3: Reference image with Base64 encoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		
		// Create S3 wrapper with max image size from config
		s3Wrapper := deps.NewS3WrapperWithSize(config.MaxImageSize)
		
		// Download and encode image
		meta, err := s3Wrapper.GetS3ImageWithBase64(ctx, referenceUrl)
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
			log.Info("Successfully fetched and encoded reference image", map[string]interface{}{
				"contentType":    meta.ContentType,
				"size":           meta.Size,
				"base64Length":   len(meta.Base64Data),
				"imageFormat":    meta.ImageFormat,
				"duration":       time.Since(fetchStart).Milliseconds(),
			})
		}
	}()

	// S3: Checking image with Base64 encoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchStart := time.Now()
		
		// Create S3 wrapper with max image size from config
		s3Wrapper := deps.NewS3WrapperWithSize(config.MaxImageSize)
		
		// Download and encode image
		meta, err := s3Wrapper.GetS3ImageWithBase64(ctx, checkingUrl)
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
			log.Info("Successfully fetched and encoded checking image", map[string]interface{}{
				"contentType":    meta.ContentType,
				"size":           meta.Size,
				"base64Length":   len(meta.Base64Data),
				"imageFormat":    meta.ImageFormat,
				"duration":       time.Since(fetchStart).Milliseconds(),
			})
		}
	}()

	// DynamoDB: Layout metadata (only for LAYOUT_VS_CHECKING)
	if layoutId != 0 && layoutPrefix != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchStart := time.Now()
			
			// Initialize the DB wrapper
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
			
			// Fetch the layout metadata
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
			
			// Initialize the DB wrapper
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
	
	// Log final results summary with Base64 metrics
	totalDuration := time.Since(startTime)
	log.Info("Parallel fetch operations completed", map[string]interface{}{
		"totalDuration":            totalDuration.Milliseconds(),
		"errorCount":               len(results.Errors),
		"referenceImageFetched":    results.ReferenceMeta.HasBase64Data(),
		"checkingImageFetched":     results.CheckingMeta.HasBase64Data(),
		"layoutMetadataFetched":    len(results.LayoutMeta) > 0,
		"historicalContextFetched": len(results.HistoricalContext) > 0,
		"referenceSize":            results.ReferenceMeta.Size,
		"checkingSize":             results.CheckingMeta.Size,
		"referenceBase64Length":    len(results.ReferenceMeta.Base64Data),
		"checkingBase64Length":     len(results.CheckingMeta.Base64Data),
		"referenceFormat":          results.ReferenceMeta.ImageFormat,
		"checkingFormat":           results.CheckingMeta.ImageFormat,
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
	
	// Validate that we have Base64 data for both images if no errors

	// Validate that we have Base64 data for both images if no errors
	if len(results.Errors) == 0 {
		if !results.ReferenceMeta.HasBase64Data() {
			log.Warn("Reference image missing Base64 data despite no errors", map[string]interface{}{
				"verificationId": "unknown", // You can pass this from context if available
			})
		}
		if !results.CheckingMeta.HasBase64Data() {
			log.Warn("Checking image missing Base64 data despite no errors", map[string]interface{}{
				"verificationId": "unknown", // You can pass this from context if available
			})
		}
	}
	
	return results
}