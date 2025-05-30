package handler

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// ContextLoader handles concurrent loading of system prompt and base64 image
type ContextLoader struct {
	s3  services.S3StateManager
	log logger.Logger
}

// NewContextLoader creates a new instance of ContextLoader
func NewContextLoader(s3 services.S3StateManager, log logger.Logger) *ContextLoader {
	return &ContextLoader{
		s3:  s3,
		log: log,
	}
}

// LoadResult contains the results of context loading
type LoadResult struct {
	SystemPrompt     string
	Base64Image      string
	ImageFormat      string
	Duration         time.Duration
	Error            error
}

// LoadContext loads system prompt and base64 image concurrently (legacy Turn1 method)
// This method is deprecated for Turn2 processing. Use LoadContextTurn2 instead.
func (c *ContextLoader) LoadContext(ctx context.Context, req interface{}) *LoadResult {
	startTime := time.Now()
	result := &LoadResult{
		Error:    fmt.Errorf("LoadContext method is deprecated for Turn2 processing. Use LoadContextTurn2 instead"),
		Duration: time.Since(startTime),
	}
	return result
}

// loadWithRetry implements exponential backoff retry logic for S3 operations
func (c *ContextLoader) loadWithRetry(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond
	const maxDelay = 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			if attempt > 0 {
				c.log.Info("retry_successful", map[string]interface{}{
					"attempt": attempt + 1,
					"total_attempts": maxRetries,
				})
			}
			return result, nil
		}

		lastErr = err
		
		// Check if error is retryable
		if wfErr, ok := err.(*errors.WorkflowError); ok && !wfErr.Retryable {
			c.log.Debug("non_retryable_error_encountered", map[string]interface{}{
				"error": err.Error(),
				"attempt": attempt + 1,
			})
			break
		}

		// Don't retry on the last attempt
		if attempt == maxRetries-1 {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		if delay > maxDelay {
			delay = maxDelay
		}
		
		c.log.Debug("retrying_operation", map[string]interface{}{
			"attempt": attempt + 1,
			"max_attempts": maxRetries,
			"delay_ms": delay.Milliseconds(),
			"error": err.Error(),
		})

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	c.log.Error("all_retry_attempts_failed", map[string]interface{}{
		"max_attempts": maxRetries,
		"final_error": lastErr.Error(),
	})
	
	return nil, lastErr
}

// LoadContextTurn2 loads system prompt and checking image concurrently for Turn2
func (c *ContextLoader) LoadContextTurn2(ctx context.Context, req *models.Turn2Request) *LoadResult {
	startTime := time.Now()
	result := &LoadResult{}

	var (
		systemPrompt     string
		base64Img        string
		imageFormat      = req.S3Refs.Images.CheckingImageFormat
		loadErr          error
		errorMutex       sync.Mutex // Protect loadErr from race conditions
	)

	// Helper function to safely set error (only sets the first error encountered)
	setError := func(err error) {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		if loadErr == nil { // Only set the first error
			loadErr = err
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2) // 2 concurrent operations

	// Load system prompt
	go func() {
		defer wg.Done()

		c.log.Debug("loading_system_prompt_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Prompts.System.Bucket,
			"key":    req.S3Refs.Prompts.System.Key,
			"size":   req.S3Refs.Prompts.System.Size,
		})

		sp, err := c.loadWithRetry(ctx, func() (interface{}, error) {
			return c.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		})
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load system prompt", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "system_prompt_load").
				WithContext("expected_content_type", "string")

			setError(errors.SetVerificationID(enrichedErr, req.VerificationID))
			return
		}

		c.log.Debug("system_prompt_loaded_successfully", map[string]interface{}{
			"prompt_length": len(sp.(string)),
			"bucket":        req.S3Refs.Prompts.System.Bucket,
			"key":           req.S3Refs.Prompts.System.Key,
		})

		systemPrompt = sp.(string)
	}()

	// Load checking image
	go func() {
		defer wg.Done()

		c.log.Debug("loading_checking_image_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Images.CheckingBase64.Bucket,
			"key":    req.S3Refs.Images.CheckingBase64.Key,
			"size":   req.S3Refs.Images.CheckingBase64.Size,
		})

		img, err := c.loadWithRetry(ctx, func() (interface{}, error) {
			return c.s3.LoadBase64Image(ctx, req.S3Refs.Images.CheckingBase64)
		})
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load checking image", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Images.CheckingBase64.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "checking_image_load").
				WithContext("expected_content_type", "base64")

			setError(errors.SetVerificationID(enrichedErr, req.VerificationID))
			return
		}

		c.log.Debug("checking_image_loaded_successfully", map[string]interface{}{
			"image_length": len(img.(string)),
			"bucket":       req.S3Refs.Images.CheckingBase64.Bucket,
			"key":          req.S3Refs.Images.CheckingBase64.Key,
		})

		base64Img = img.(string)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	if loadErr != nil {
		result.Error = loadErr
		result.Duration = time.Since(startTime)
		return result
	}

	// Set result fields
	result.SystemPrompt = systemPrompt
	result.Base64Image = base64Img
	result.ImageFormat = imageFormat
	result.Duration = time.Since(startTime)

	c.log.Info("turn2_context_loading_completed_successfully", map[string]interface{}{
		"system_prompt_length":  len(systemPrompt),
		"base64_image_length":   len(base64Img),
		"total_duration_ms":     result.Duration.Milliseconds(),
		"concurrent_operations": 2,
	})

	return result
}
