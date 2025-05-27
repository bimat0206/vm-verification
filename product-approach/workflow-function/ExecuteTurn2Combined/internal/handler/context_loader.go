package handler

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
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
	Turn1Response    *schema.Turn1ProcessedResponse
	Turn1RawResponse json.RawMessage
	Duration         time.Duration
	Error            error
}

// LoadContext loads system prompt and base64 image concurrently for Turn1
func (c *ContextLoader) LoadContext(ctx context.Context, req *models.Turn1Request) *LoadResult {
	startTime := time.Now()
	result := &LoadResult{}

	var (
		systemPrompt string
		base64Img    string
		loadErr      error
	)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Load system prompt - ENHANCED ERROR HANDLING
	go func() {
		defer wg.Done()

		// STRATEGIC CHANGE: Add context validation
		c.log.Debug("loading_system_prompt_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Prompts.System.Bucket,
			"key":    req.S3Refs.Prompts.System.Key,
			"size":   req.S3Refs.Prompts.System.Size,
		})

		sp, err := c.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load system prompt", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "system_prompt_load").
				WithContext("expected_content_type", "string")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("system_prompt_loaded_successfully", map[string]interface{}{
			"prompt_length": len(sp),
			"bucket":        req.S3Refs.Prompts.System.Bucket,
			"key":           req.S3Refs.Prompts.System.Key,
		})

		systemPrompt = sp
	}()

	// Load base64 image - ENHANCED ERROR HANDLING
	go func() {
		defer wg.Done()

		c.log.Debug("loading_base64_image_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Images.ReferenceBase64.Bucket,
			"key":    req.S3Refs.Images.ReferenceBase64.Key,
			"size":   req.S3Refs.Images.ReferenceBase64.Size,
		})

		img, err := c.s3.LoadBase64Image(ctx, req.S3Refs.Images.ReferenceBase64)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load reference image", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Images.ReferenceBase64.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "base64_image_load").
				WithContext("image_size", req.S3Refs.Images.ReferenceBase64.Size).
				WithContext("expected_content_type", "string")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("base64_image_loaded_successfully", map[string]interface{}{
			"image_data_length": len(img),
			"bucket":            req.S3Refs.Images.ReferenceBase64.Bucket,
			"key":               req.S3Refs.Images.ReferenceBase64.Key,
		})

		base64Img = img
	}()

	wg.Wait()

	result.SystemPrompt = systemPrompt
	result.Base64Image = base64Img
	result.Duration = time.Since(startTime)
	result.Error = loadErr

	if loadErr == nil {
		c.log.Info("context_loading_completed_successfully", map[string]interface{}{
			"system_prompt_length":  len(systemPrompt),
			"base64_image_length":   len(base64Img),
			"total_duration_ms":     result.Duration.Milliseconds(),
			"concurrent_operations": 2,
		})
	}

	return result
}

// LoadContextTurn2 loads system prompt, checking image, and Turn1 results concurrently for Turn2
func (c *ContextLoader) LoadContextTurn2(ctx context.Context, req *models.Turn2Request) *LoadResult {
	startTime := time.Now()
	result := &LoadResult{}

	var (
		systemPrompt     string
		base64Img        string
		turn1Response    *schema.Turn1ProcessedResponse
		turn1RawResponse json.RawMessage
		loadErr          error
	)

	wg := sync.WaitGroup{}
	wg.Add(4) // 4 concurrent operations

	// Load system prompt
	go func() {
		defer wg.Done()

		c.log.Debug("loading_system_prompt_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Prompts.System.Bucket,
			"key":    req.S3Refs.Prompts.System.Key,
			"size":   req.S3Refs.Prompts.System.Size,
		})

		sp, err := c.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load system prompt", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "system_prompt_load").
				WithContext("expected_content_type", "string")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("system_prompt_loaded_successfully", map[string]interface{}{
			"prompt_length": len(sp),
			"bucket":        req.S3Refs.Prompts.System.Bucket,
			"key":           req.S3Refs.Prompts.System.Key,
		})

		systemPrompt = sp
	}()

	// Load checking image
	go func() {
		defer wg.Done()

		c.log.Debug("loading_checking_image_concurrently", map[string]interface{}{
			"bucket": req.S3Refs.Images.CheckingBase64.Bucket,
			"key":    req.S3Refs.Images.CheckingBase64.Key,
			"size":   req.S3Refs.Images.CheckingBase64.Size,
		})

		img, err := c.s3.LoadBase64Image(ctx, req.S3Refs.Images.CheckingBase64)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load checking image", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Images.CheckingBase64.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "checking_image_load").
				WithContext("expected_content_type", "base64")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("checking_image_loaded_successfully", map[string]interface{}{
			"image_length": len(img),
			"bucket":       req.S3Refs.Images.CheckingBase64.Bucket,
			"key":          req.S3Refs.Images.CheckingBase64.Key,
		})

		base64Img = img
	}()

	// Load Turn1 processed response
	go func() {
		defer wg.Done()

		c.log.Debug("loading_turn1_processed_response", map[string]interface{}{
			"bucket": req.S3Refs.Turn1.ProcessedResponse.Bucket,
			"key":    req.S3Refs.Turn1.ProcessedResponse.Key,
			"size":   req.S3Refs.Turn1.ProcessedResponse.Size,
		})

		processedResponse, err := c.s3.LoadTurn1ProcessedResponse(ctx, req.S3Refs.Turn1.ProcessedResponse)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load Turn1 processed response", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Turn1.ProcessedResponse.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "turn1_processed_response_load").
				WithContext("expected_content_type", "json")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("turn1_processed_response_loaded_successfully", map[string]interface{}{
			"bucket": req.S3Refs.Turn1.ProcessedResponse.Bucket,
			"key":    req.S3Refs.Turn1.ProcessedResponse.Key,
		})

		turn1Response = processedResponse
	}()

	// Load Turn1 raw response
	go func() {
		defer wg.Done()

		c.log.Debug("loading_turn1_raw_response", map[string]interface{}{
			"bucket": req.S3Refs.Turn1.RawResponse.Bucket,
			"key":    req.S3Refs.Turn1.RawResponse.Key,
			"size":   req.S3Refs.Turn1.RawResponse.Size,
		})

		rawResponse, err := c.s3.LoadTurn1RawResponse(ctx, req.S3Refs.Turn1.RawResponse)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3,
				"failed to load Turn1 raw response", true)

			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Turn1.RawResponse.Key).
				WithContext("stage", "context_loading").
				WithContext("operation", "turn1_raw_response_load").
				WithContext("expected_content_type", "json")

			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}

		c.log.Debug("turn1_raw_response_loaded_successfully", map[string]interface{}{
			"bucket": req.S3Refs.Turn1.RawResponse.Bucket,
			"key":    req.S3Refs.Turn1.RawResponse.Key,
		})

		turn1RawResponse = rawResponse
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	if loadErr != nil {
		result.Error = loadErr
		return result
	}

	// Set result fields
	result.SystemPrompt = systemPrompt
	result.Base64Image = base64Img
	result.Turn1Response = turn1Response
	result.Turn1RawResponse = turn1RawResponse
	result.Duration = time.Since(startTime)

	if loadErr == nil {
		c.log.Info("turn2_context_loading_completed_successfully", map[string]interface{}{
			"system_prompt_length":  len(systemPrompt),
			"base64_image_length":   len(base64Img),
			"turn1_response_loaded": turn1Response != nil,
			"turn1_raw_loaded":     len(turn1RawResponse) > 0,
			"total_duration_ms":    result.Duration.Milliseconds(),
			"concurrent_operations": 4,
		})
	}

	return result
}
