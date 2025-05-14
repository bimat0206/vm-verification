package main

import (
	"context"
	"fmt"
	"sync"
)

// ParallelFetch executes the S3 and DynamoDB fetches concurrently.
func ParallelFetch(
	ctx context.Context,
	deps *Dependencies,
	referenceUrl string,
	checkingUrl string,
	layoutId int,
	layoutPrefix string,
	prevVerificationId string,
) ParallelFetchResults {
	var wg sync.WaitGroup
	results := ParallelFetchResults{
		LayoutMeta:        make(map[string]interface{}),
		HistoricalContext: make(map[string]interface{}),
		Errors:            []error{},
	}
	var mu sync.Mutex
	log := deps.GetLogger()

	wg.Add(2)

	// S3: Reference image metadata
	go func() {
		defer wg.Done()
		s3Wrapper := NewS3Utils(deps.GetS3Client(), deps.GetLogger())
		meta, err := s3Wrapper.GetS3ImageMetadata(ctx, referenceUrl)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, err)
			log.Error("Failed to fetch reference image metadata", map[string]interface{}{
				"url":   referenceUrl,
				"error": err.Error(),
			})
		} else {
			results.ReferenceMeta = meta
			log.Info("Successfully fetched reference image metadata", map[string]interface{}{
				"contentType": meta.ContentType,
				"size":       meta.Size,
			})
		}
	}()

	// S3: Checking image metadata
	go func() {
		defer wg.Done()
		s3Wrapper := NewS3Utils(deps.GetS3Client(), deps.GetLogger())
		meta, err := s3Wrapper.GetS3ImageMetadata(ctx, checkingUrl)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			results.Errors = append(results.Errors, err)
			log.Error("Failed to fetch checking image metadata", map[string]interface{}{
				"url":   checkingUrl,
				"error": err.Error(),
			})
		} else {
			results.CheckingMeta = meta
			log.Info("Successfully fetched checking image metadata", map[string]interface{}{
				"contentType": meta.ContentType,
				"size":       meta.Size,
			})
		}
	}()

	// DynamoDB: Layout metadata (optional)
	if layoutId != 0 && layoutPrefix != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Initialize the DB wrapper if needed
			var dbWrapper *DBUtilsWrapper
			if deps.GetDbUtils() != nil {
				dbWrapper = NewDBUtils(deps.GetDbUtils())
			} else {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("dbUtils not initialized"))
				log.Error("dbUtils not initialized for layout metadata fetch", nil)
				mu.Unlock()
				return
			}
			
			meta, err := dbWrapper.FetchLayoutMetadata(ctx, layoutId, layoutPrefix)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results.Errors = append(results.Errors, err)
				log.Error("Failed to fetch layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
					"error":        err.Error(),
				})
			} else {
				results.LayoutMeta = meta
				log.Info("Successfully fetched layout metadata", map[string]interface{}{
					"layoutId":     layoutId,
					"layoutPrefix": layoutPrefix,
				})
			}
		}()
	}

	// DynamoDB: Historical context (required for PREVIOUS_VS_CURRENT, skipped for LAYOUT_VS_CHECKING)
	if prevVerificationId != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Initialize the DB wrapper if needed
			var dbWrapper *DBUtilsWrapper
			if deps.GetDbUtils() != nil {
				dbWrapper = NewDBUtils(deps.GetDbUtils())
			} else {
				mu.Lock()
				results.Errors = append(results.Errors, fmt.Errorf("dbUtils not initialized"))
				log.Error("dbUtils not initialized for historical context fetch", nil)
				mu.Unlock()
				return
			}
			
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
				})
			} else {
				results.HistoricalContext = ctxObj
				log.Info("Successfully fetched historical verification", map[string]interface{}{
					"previousVerificationId": prevVerificationId,
					"verificationStatus":     ctxObj["previousVerificationStatus"],
					"verificationAt":         ctxObj["previousVerificationAt"],
					"hoursSince":             ctxObj["hoursSinceLastVerification"],
				})
			}
		}()
	} else {
		// Log that we're skipping historical context for non-PREVIOUS_VS_CURRENT verifications
		log.Info("Skipping historical context fetch", map[string]interface{}{
			"reason": "No previousVerificationId provided, this is expected for LAYOUT_VS_CHECKING verification type",
		})
	}

	wg.Wait()
	return results
}