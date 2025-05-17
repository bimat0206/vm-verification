package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	
	"github.com/aws/smithy-go"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Predefined errors
var (
	ErrImageTooLarge       = errors.New("image exceeds maximum allowed size")
	ErrInvalidContentType  = errors.New("invalid content type for image")
	ErrResourceNotFound    = errors.New("resource not found in S3")
)

// S3Validator provides methods to validate S3 resources
type S3Validator struct {
	s3Client *S3Client
	urlParser *S3URLParser
	logger   logger.Logger
}

// NewS3Validator creates a new S3Validator
func NewS3Validator(s3Client *S3Client, urlParser *S3URLParser, log logger.Logger) *S3Validator {
	return &S3Validator{
		s3Client: s3Client,
		urlParser: urlParser,
		logger:   log.WithFields(map[string]interface{}{"component": "S3Validator"}),
	}
}

// ValidateImageExists checks if an image exists and validates its size and type
func (v *S3Validator) ValidateImageExists(ctx context.Context, s3URL string, maxSize int64) (bool, error) {
	// Parse the S3 URL to get bucket and key
	bucket, key, err := v.urlParser.ParseS3URL(s3URL)
	if err != nil {
		return false, err
	}
	
	// Check if the file extension is valid for an image
	if !v.urlParser.IsValidImageExtension(key) {
		return false, fmt.Errorf("%w: URL %s has invalid image extension", ErrInvalidContentType, s3URL)
	}
	
	// Check if the object exists in S3
	headResult, err := v.s3Client.HeadObject(ctx, bucket, key)
	if err != nil {
		var apiErr smithy.APIError
		// Check if the error is a NoSuchKey error
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey") {
			return false, fmt.Errorf("%w: %s", ErrResourceNotFound, s3URL)
		}
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}
	
	// Check if the content type is valid for an image
	contentType := ""
	if headResult.ContentType != nil {
		contentType = *headResult.ContentType
	}
	
	if !isValidImageContentType(contentType) && contentType != "" {
		return false, fmt.Errorf("%w: %s has content type %s", ErrInvalidContentType, s3URL, contentType)
	}
	
	// Check if the image size is within limits
	if headResult.ContentLength != nil && *headResult.ContentLength > maxSize {
		return false, fmt.Errorf("%w: %s size %d exceeds limit %d", 
			ErrImageTooLarge, s3URL, *headResult.ContentLength, maxSize)
	}
	
	return true, nil
}

// ValidateImagesInParallel validates multiple S3 images concurrently
func (v *S3Validator) ValidateImagesInParallel(
	ctx context.Context, 
	imageURLs []string, 
	maxSize int64,
) (*schema.ResourceValidation, error) {
	result := &schema.ResourceValidation{
		ValidationTimestamp: schema.FormatISO8601(),
	}
	
	if len(imageURLs) == 0 {
		return result, nil
	}
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(imageURLs))
	
	// Create a validation function to reuse
	validateImage := func(url string, resultField *bool) {
		defer wg.Done()
		
		exists, err := v.ValidateImageExists(ctx, url, maxSize)
		if err != nil {
			errChan <- fmt.Errorf("failed to validate %s: %w", url, err)
			return
		}
		
		*resultField = exists
		errChan <- nil
	}
	
	// Validate all images in parallel
	for i, url := range imageURLs {
		wg.Add(1)
		
		// First image is reference, second is checking
		if i == 0 {
			go validateImage(url, &result.ReferenceImageExists)
		} else if i == 1 {
			go validateImage(url, &result.CheckingImageExists)
		}
	}
	
	// Close channel after all goroutines finish
	go func() {
		wg.Wait()
		close(errChan)
	}()
	
	// Collect errors
	for err := range errChan {
		if err != nil {
			return result, err
		}
	}
	
	return result, nil
}

// isValidImageContentType checks if the content type is valid for an image
func isValidImageContentType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}
	
	return validTypes[contentType]
}