package storage

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"workflow-function/shared/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds configuration for S3 operations
type S3Config struct {
	ReferenceBucket string // Bucket for reference images
	CheckingBucket  string // Bucket for checking images
	ResultsBucket   string // Bucket for result images
	MaxImageSize    int64  // Maximum allowed image size in bytes (default: 10MB)
}

// S3Manager provides utilities for S3 operations
type S3Manager struct {
	client *s3.Client
	logger logger.Logger
	config S3Config
}

// S3URL represents a parsed S3 URL
type S3URL struct {
	Bucket string // S3 bucket name
	Key    string // S3 object key
}

// NewS3Manager creates a new S3Manager instance
func NewS3Manager(client *s3.Client, log logger.Logger, config S3Config) *S3Manager {
	// Set default max size if not specified
	if config.MaxImageSize <= 0 {
		config.MaxImageSize = 10 * 1024 * 1024 // 10MB
	}

	return &S3Manager{
		client: client,
		logger: log.WithFields(map[string]interface{}{
			"component": "s3_manager",
		}),
		config: config,
	}
}

// ValidateImageExists checks if an image exists in S3 and has valid properties
func (m *S3Manager) ValidateImageExists(ctx context.Context, s3Url string) (bool, error) {
	// Parse S3 URL
	parsed, err := m.ParseS3URL(s3Url)
	if err != nil {
		return false, NewValidationError(
			"Invalid S3 URL",
			map[string]interface{}{"url": s3Url, "error": err.Error()},
		)
	}

	m.logger.Debug("Validating image existence", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
	})

	// Check if image exists
	headOutput, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return false, NewS3Error(
			fmt.Sprintf("Image not found in S3: %s", s3Url),
			"NOT_FOUND",
			false,
			err,
		)
	}

	// Validate content type is an image
	contentType := aws.ToString(headOutput.ContentType)
	if !m.IsImageContentType(contentType) {
		return false, NewValidationError(
			fmt.Sprintf("Invalid content type: %s, expected image/*", contentType),
			map[string]interface{}{"contentType": contentType},
		)
	}

	// Check if image size is within limits
	if headOutput.ContentLength != nil && *headOutput.ContentLength > m.config.MaxImageSize {
		return false, NewValidationError(
			fmt.Sprintf("Image too large: %d bytes, maximum allowed: %d bytes", *headOutput.ContentLength, m.config.MaxImageSize),
			map[string]interface{}{
				"size":     *headOutput.ContentLength,
				"maxSize":  m.config.MaxImageSize,
			},
		)
	}

	m.logger.Debug("Image validated successfully", map[string]interface{}{
		"contentType":   contentType,
		"contentLength": headOutput.ContentLength,
	})

	return true, nil
}

// GetImageMetadata retrieves metadata for an S3 image
func (m *S3Manager) GetImageMetadata(ctx context.Context, s3Url string) (map[string]string, error) {
	// Parse S3 URL
	parsed, err := m.ParseS3URL(s3Url)
	if err != nil {
		return nil, NewValidationError(
			"Invalid S3 URL",
			map[string]interface{}{"url": s3Url, "error": err.Error()},
		)
	}

	m.logger.Debug("Getting S3 object metadata", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
	})

	// Get object metadata
	headOutput, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return nil, NewS3Error(
			"Failed to get image metadata",
			"HEAD_ERROR",
			true,
			err,
		)
	}

	// Extract metadata
	metadata := make(map[string]string)
	
	// Add standard metadata with proper nil handling
	metadata["contentType"] = aws.ToString(headOutput.ContentType)
	
	// Handle ContentLength properly
	if headOutput.ContentLength != nil {
		metadata["contentLength"] = fmt.Sprintf("%d", *headOutput.ContentLength)
	} else {
		metadata["contentLength"] = "0"
	}
	
	// Handle LastModified properly
	if headOutput.LastModified != nil {
		metadata["lastModified"] = headOutput.LastModified.Format("2006-01-02T15:04:05Z")
	} else {
		metadata["lastModified"] = ""
	}
	
	// Handle ETag properly (remove quotes if present)
	if headOutput.ETag != nil {
		etag := *headOutput.ETag
		// Remove surrounding quotes if present
		if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		metadata["etag"] = etag
	} else {
		metadata["etag"] = ""
	}
	
	metadata["bucket"] = parsed.Bucket
	metadata["key"] = parsed.Key
	
	// Extract any custom metadata
	for k, v := range headOutput.Metadata {
		metadata[k] = v
	}

	m.logger.Debug("Successfully retrieved image metadata", map[string]interface{}{
		"contentType":   metadata["contentType"],
		"contentLength": metadata["contentLength"],
		"etag":          metadata["etag"],
	})

	return metadata, nil
}

// ParseS3URLs parses both reference and checking image URLs
func (m *S3Manager) ParseS3URLs(refUrl, checkUrl string) (S3URL, S3URL, error) {
	// Parse reference image URL
	refParsed, err := m.ParseS3URL(refUrl)
	if err != nil {
		return S3URL{}, S3URL{}, NewValidationError(
			"Invalid reference image URL",
			map[string]interface{}{"url": refUrl, "error": err.Error()},
		)
	}
	
	// Parse checking image URL
	checkParsed, err := m.ParseS3URL(checkUrl)
	if err != nil {
		return S3URL{}, S3URL{}, NewValidationError(
			"Invalid checking image URL",
			map[string]interface{}{"url": checkUrl, "error": err.Error()},
		)
	}
	
	return refParsed, checkParsed, nil
}

// ParseS3URL parses an S3 URL and returns the bucket and key
func (m *S3Manager) ParseS3URL(s3Url string) (S3URL, error) {
	// Support both formats: s3://bucket/key and https://bucket.s3.region.amazonaws.com/key
	if strings.HasPrefix(s3Url, "s3://") {
		parts := strings.SplitN(strings.TrimPrefix(s3Url, "s3://"), "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return S3URL{}, fmt.Errorf("invalid S3 URL format: missing bucket or key")
		}
		
		if !m.IsValidBucketName(parts[0]) {
			return S3URL{}, fmt.Errorf("invalid S3 bucket name")
		}
		
		return S3URL{
			Bucket: parts[0],
			Key:    parts[1],
		}, nil
	} else if contains(s3Url, ".s3.") {
		re := regexp.MustCompile(`https://([^.]+)\.s3\.[^/]+\.amazonaws\.com/(.+)`)
		matches := re.FindStringSubmatch(s3Url)
		if len(matches) != 3 || matches[1] == "" || matches[2] == "" {
			return S3URL{}, fmt.Errorf("invalid S3 URL format: unable to extract bucket and key")
		}
		
		if !m.IsValidBucketName(matches[1]) {
			return S3URL{}, fmt.Errorf("invalid S3 bucket name")
		}
		
		return S3URL{
			Bucket: matches[1],
			Key:    matches[2],
		}, nil
	}
	return S3URL{}, fmt.Errorf("unsupported S3 URL format")
}

// BuildS3URL constructs an S3 URL from bucket and key
func (m *S3Manager) BuildS3URL(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}

// IsValidBucketName checks if a bucket name is valid
func (m *S3Manager) IsValidBucketName(bucket string) bool {
	// AWS bucket naming rules
	if len(bucket) < 3 || len(bucket) > 63 {
		return false
	}
	
	// Must consist of lowercase letters, numbers, dots, and hyphens
	match, _ := regexp.MatchString("^[a-z0-9.-]+$", bucket)
	if !match {
		return false
	}
	
	// Must start and end with a letter or number
	if match, _ := regexp.MatchString("^[a-z0-9]", bucket); !match {
		return false
	}
	
	if match, _ := regexp.MatchString("[a-z0-9]$", bucket); !match {
		return false
	}
	
	// Cannot contain consecutive periods
	if strings.Contains(bucket, "..") {
		return false
	}
	
	// Cannot be formatted as an IP address
	if match, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", bucket); match {
		return false
	}
	
	return true
}

// IsImageContentType checks if a content type is a valid image type
func (m *S3Manager) IsImageContentType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}
	
	// Convert to lowercase for comparison
	contentType = strings.ToLower(contentType)
	
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	
	// Also check if it starts with image/
	return strings.HasPrefix(contentType, "image/")
}

// IsValidImageExtension checks if a file has a valid image extension
func (m *S3Manager) IsValidImageExtension(key string) bool {
	validExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tiff"}
	lower := strings.ToLower(key)
	
	for _, ext := range validExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	
	return false
}

// ValidateImageForBedrock validates that an image meets Bedrock requirements
func (m *S3Manager) ValidateImageForBedrock(ctx context.Context, s3Url string) error {
	// Parse URL
	if _, err := m.ParseS3URL(s3Url); err != nil {
		return NewValidationError(
			"Invalid S3 URL",
			map[string]interface{}{"url": s3Url, "error": err.Error()},
		)
	}

	// Get image metadata
	metadata, err := m.GetImageMetadata(ctx, s3Url)
	if err != nil {
		return NewS3Error(
			"Failed to get image metadata",
			"METADATA_ERROR",
			true,
			err,
		)
	}

	// Validate content type
	contentType := metadata["contentType"]
	if !m.IsImageContentType(contentType) {
		return NewValidationError(
			fmt.Sprintf("Invalid content type: %s, expected image/png or image/jpeg", contentType),
			map[string]interface{}{"contentType": contentType},
		)
	}

	// Bedrock specifically supports PNG and JPEG
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/jpg" {
		m.logger.Warn("Content type may not be fully supported by Bedrock", map[string]interface{}{
			"contentType": contentType,
			"s3Url": s3Url,
		})
	}

	// Parse and validate size
	contentLength := metadata["contentLength"]
	if contentLength == "" {
		return NewValidationError(
			"Content length not available",
			map[string]interface{}{"url": s3Url},
		)
	}

	size := int64(0)
	if _, err := fmt.Sscanf(contentLength, "%d", &size); err != nil {
		return NewValidationError(
			fmt.Sprintf("Invalid content length: %s", contentLength),
			map[string]interface{}{"contentLength": contentLength},
		)
	}
	
	// Bedrock limit is typically 100MB per image
	const maxSize = 100 * 1024 * 1024 // 100MB
	if size > maxSize {
		return NewValidationError(
			fmt.Sprintf("Image too large: %d bytes (max %d bytes for Bedrock)", size, maxSize),
			map[string]interface{}{"size": size, "maxSize": maxSize},
		)
	}

	m.logger.Info("Image validated for Bedrock", map[string]interface{}{
		"s3Url":        s3Url,
		"contentType":  contentType,
		"size":         size,
	})

	return nil
}
