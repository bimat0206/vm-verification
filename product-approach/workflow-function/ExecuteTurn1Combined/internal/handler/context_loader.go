package handler

import (
	"context"
	"sync"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
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
	SystemPrompt string
	Base64Image  string
	Duration     time.Duration
	Error        error
}

// LoadContext loads system prompt and base64 image concurrently
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

	// Load system prompt
	go func() {
		defer wg.Done()
		sp, err := c.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3, 
				"failed to load system prompt", true)
			
			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Prompts.System.Key).
				WithContext("stage", "context_loading")
			
			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}
		systemPrompt = sp
	}()

	// Load base64 image
	go func() {
		defer wg.Done()
		img, err := c.s3.LoadBase64Image(ctx, req.S3Refs.Images.ReferenceBase64)
		if err != nil {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeS3, 
				"failed to load reference image", true)
			
			enrichedErr := wrappedErr.WithContext("s3_key", req.S3Refs.Images.ReferenceBase64.Key).
				WithContext("stage", "context_loading").
				WithContext("image_size", req.S3Refs.Images.ReferenceBase64.Size)
			
			loadErr = errors.SetVerificationID(enrichedErr, req.VerificationID)
			return
		}
		base64Img = img
	}()

	wg.Wait()
	
	result.SystemPrompt = systemPrompt
	result.Base64Image = base64Img
	result.Duration = time.Since(startTime)
	result.Error = loadErr
	
	return result
}