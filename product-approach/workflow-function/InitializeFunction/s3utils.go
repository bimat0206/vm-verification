package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Utils provides utilities for S3 operations
type S3Utils struct {
	client *s3.Client
	logger Logger
	config *ConfigVars
}

// NewS3Utils creates a new S3Utils instance
func NewS3Utils(client *s3.Client, logger Logger) *S3Utils {
	return &S3Utils{
		client: client,
		logger: logger,
		config: &ConfigVars{},
	}
}

// SetConfig sets the configuration for S3Utils
func (u *S3Utils) SetConfig(config ConfigVars) {
	u.config = &config
}

// ValidateImageExists checks if an image exists in S3 and has valid properties
func (u *S3Utils) ValidateImageExists(ctx context.Context, s3Url string) error {
	// Parse S3 URL
	parsed, err := u.ParseS3URL(s3Url)
	if err != nil {
		return fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Check if image exists
	headOutput, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return fmt.Errorf("image not found in S3: %w", err)
	}

	// Validate content type is an image
	contentType := aws.ToString(headOutput.ContentType)
	if !u.IsImageContentType(contentType) {
		return fmt.Errorf("invalid content type: %s, expected image/*", contentType)
	}

	// Check if image size is within limits (e.g., < 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if headOutput.ContentLength != nil && *headOutput.ContentLength > maxSize {
		return fmt.Errorf("image too large: %d bytes, maximum allowed: %d bytes", *headOutput.ContentLength, maxSize)
	}

	return nil
}

// GetImageMetadata retrieves metadata for an S3 image
func (u *S3Utils) GetImageMetadata(ctx context.Context, s3Url string) (map[string]string, error) {
	// Parse S3 URL
	parsed, err := u.ParseS3URL(s3Url)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Get object metadata
	headOutput, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Extract metadata
	metadata := make(map[string]string)
	
	// Add standard metadata
	metadata["contentType"] = aws.ToString(headOutput.ContentType)
	metadata["contentLength"] = fmt.Sprintf("%d", headOutput.ContentLength)
	metadata["lastModified"] = headOutput.LastModified.Format("2006-01-02T15:04:05Z")
	
	// Extract any custom metadata
	for k, v := range headOutput.Metadata {
		metadata[k] = v
	}

	return metadata, nil
}

// ParseS3URLs parses both reference and checking image URLs
func (u *S3Utils) ParseS3URLs(refUrl, checkUrl string) (S3URL, S3URL, error) {
	// Parse reference image URL
	refParsed, err := u.ParseS3URL(refUrl)
	if err != nil {
		return S3URL{}, S3URL{}, fmt.Errorf("invalid reference image URL: %w", err)
	}
	
	// Parse checking image URL
	checkParsed, err := u.ParseS3URL(checkUrl)
	if err != nil {
		return S3URL{}, S3URL{}, fmt.Errorf("invalid checking image URL: %w", err)
	}
	
	return refParsed, checkParsed, nil
}

// ParseS3URL parses an S3 URL and returns the bucket and key
func (u *S3Utils) ParseS3URL(s3Url string) (S3URL, error) {
	// Support both formats: s3://bucket/key and https://bucket.s3.region.amazonaws.com/key
	if strings.HasPrefix(s3Url, "s3://") {
		parts := strings.SplitN(strings.TrimPrefix(s3Url, "s3://"), "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return S3URL{}, errors.New("invalid S3 URL format: missing bucket or key")
		}
		
		if !u.IsValidBucketName(parts[0]) {
			return S3URL{}, errors.New("invalid S3 bucket name")
		}
		
		return S3URL{
			Bucket: parts[0],
			Key:    parts[1],
		}, nil
	} else if strings.Contains(s3Url, ".s3.") {
		re := regexp.MustCompile(`https://([^.]+)\.s3\.[^/]+\.amazonaws\.com/(.+)`)
		matches := re.FindStringSubmatch(s3Url)
		if len(matches) != 3 || matches[1] == "" || matches[2] == "" {
			return S3URL{}, errors.New("invalid S3 URL format: unable to extract bucket and key")
		}
		
		if !u.IsValidBucketName(matches[1]) {
			return S3URL{}, errors.New("invalid S3 bucket name")
		}
		
		return S3URL{
			Bucket: matches[1],
			Key:    matches[2],
		}, nil
	}
	return S3URL{}, errors.New("unsupported S3 URL format")
}

// IsValidBucketName checks if a bucket name is valid
func (u *S3Utils) IsValidBucketName(bucket string) bool {
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
func (u *S3Utils) IsImageContentType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/png",
	}
	
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	
	return false
}

// IsValidImageExtension checks if a file has a valid image extension
func (u *S3Utils) IsValidImageExtension(key string) bool {
	lower := strings.ToLower(key)
	return strings.HasSuffix(lower, ".png") || 
	       strings.HasSuffix(lower, ".jpeg") 
}