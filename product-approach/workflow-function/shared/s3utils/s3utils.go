package s3utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"workflow-function/shared/logger"
)

// S3URL represents a parsed S3 URL
type S3URL struct {
	Bucket string // S3 bucket name
	Key    string // S3 object key
}

// S3Utils provides utilities for S3 operations
type S3Utils struct {
	client *s3.Client
	config *aws.Config // Store the config for STS operations
	logger logger.Logger
}

// Config holds configuration for S3Utils
type Config struct {
	ReferenceBucket string // Bucket for reference images
	CheckingBucket  string // Bucket for checking images
	ResultsBucket   string // Bucket for result images
	MaxImageSize    int64  // Maximum allowed image size in bytes (default: 10MB)
}

// New creates a new S3Utils instance
func New(client *s3.Client, log logger.Logger) *S3Utils {
	return &S3Utils{
		client: client,
		config: nil, // No config available
		logger: log.WithFields(map[string]interface{}{
			"component": "s3utils",
		}),
	}
}

// NewWithConfig creates a new S3Utils instance with AWS config
func NewWithConfig(config aws.Config, log logger.Logger) *S3Utils {
	return &S3Utils{
		client: s3.NewFromConfig(config),
		config: &config, // Store the config
		logger: log.WithFields(map[string]interface{}{
			"component": "s3utils",
		}),
	}
}

// ValidateImageExists checks if an image exists in S3 and has valid properties
func (u *S3Utils) ValidateImageExists(ctx context.Context, s3Url string, maxSize int64) (bool, error) {
	// Default max size to 10MB if not specified
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 10MB
	}

	// Parse S3 URL
	parsed, err := u.ParseS3URL(s3Url)
	if err != nil {
		return false, fmt.Errorf("invalid S3 URL: %w", err)
	}

	u.logger.Debug("Validating image existence", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
	})

	// Check if image exists
	headOutput, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return false, fmt.Errorf("image not found in S3: %w", err)
	}

	// Validate content type is an image
	contentType := aws.ToString(headOutput.ContentType)
	if !u.IsImageContentType(contentType) {
		return false, fmt.Errorf("invalid content type: %s, expected image/*", contentType)
	}

	// Check if image size is within limits
	if headOutput.ContentLength != nil && *headOutput.ContentLength > maxSize {
		return false, fmt.Errorf("image too large: %d bytes, maximum allowed: %d bytes", *headOutput.ContentLength, maxSize)
	}

	u.logger.Debug("Image validated successfully", map[string]interface{}{
		"contentType":   contentType,
		"contentLength": headOutput.ContentLength,
	})

	return true, nil
}

// GetImageMetadata retrieves metadata for an S3 image
func (u *S3Utils) GetImageMetadata(ctx context.Context, s3Url string) (map[string]string, error) {
	// Parse S3 URL
	parsed, err := u.ParseS3URL(s3Url)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 URL: %w", err)
	}

	u.logger.Debug("Getting S3 object metadata", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
	})

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

	// Get bucket owner
	bucketOwner, err := u.GetBucketOwner(ctx, parsed.Bucket)
	if err != nil {
		u.logger.Warn("Could not retrieve bucket owner", map[string]interface{}{
			"bucket": parsed.Bucket,
			"error": err.Error(),
		})
		bucketOwner = ""
	}
	metadata["bucketOwner"] = bucketOwner

	u.logger.Debug("Successfully retrieved image metadata", map[string]interface{}{
		"contentType":   metadata["contentType"],
		"contentLength": metadata["contentLength"],
		"etag":          metadata["etag"],
		"bucketOwner":   bucketOwner,
	})

	return metadata, nil
}

// GetBucketOwner retrieves the bucket owner using multiple fallback methods
func (u *S3Utils) GetBucketOwner(ctx context.Context, bucket string) (string, error) {
	u.logger.Debug("Attempting to retrieve bucket owner", map[string]interface{}{
		"bucket": bucket,
	})

	// Method 1: Try GetBucketAcl
	aclOutput, err := u.client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err == nil && aclOutput.Owner != nil && aclOutput.Owner.ID != nil {
		u.logger.Debug("Retrieved bucket owner from GetBucketAcl", map[string]interface{}{
			"bucket": bucket,
			"owner": *aclOutput.Owner.ID,
		})
		return *aclOutput.Owner.ID, nil
	}
	
	u.logger.Debug("GetBucketAcl failed, trying fallback methods", map[string]interface{}{
		"bucket": bucket,
		"error": err.Error(),
	})
	
	// Method 2: Check environment variable
	if accountID := os.Getenv("AWS_ACCOUNT_ID"); accountID != "" {
		u.logger.Debug("Using bucket owner from environment variable", map[string]interface{}{
			"bucket": bucket,
			"accountId": accountID,
		})
		return accountID, nil
	}
	
	// Method 3: Use STS GetCallerIdentity (only if config is available)
	if u.config != nil {
		stsClient := sts.NewFromConfig(*u.config)
		identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err == nil && identity.Account != nil {
			u.logger.Debug("Retrieved bucket owner from STS GetCallerIdentity", map[string]interface{}{
				"bucket": bucket,
				"account": *identity.Account,
			})
			return *identity.Account, nil
		}
		
		u.logger.Debug("STS GetCallerIdentity failed", map[string]interface{}{
			"bucket": bucket,
			"error": err.Error(),
		})
	} else {
		u.logger.Debug("No AWS config available for STS operation", map[string]interface{}{
			"bucket": bucket,
		})
	}
	
	return "", fmt.Errorf("could not determine bucket owner for bucket %s: GetBucketAcl failed (%v), no AWS_ACCOUNT_ID env var, %s", bucket, err, 
		func() string {
			if u.config == nil {
				return "no AWS config for STS"
			}
			return "STS failed"
		}())
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

// GetPresignedURL generates a presigned URL for an S3 object
func (u *S3Utils) GetPresignedURL(ctx context.Context, bucket, key string, expireSeconds int) (string, error) {
	// This function is stubbed for now - actual implementation would use the AWS SDK presigner
	u.logger.Debug("Getting presigned URL", map[string]interface{}{
		"bucket":        bucket,
		"key":           key,
		"expireSeconds": expireSeconds,
	})
	
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?presigned=true", bucket, key), nil
}

// BuildS3URL constructs an S3 URL from bucket and key
func (u *S3Utils) BuildS3URL(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
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
func (u *S3Utils) IsValidImageExtension(key string) bool {
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
func (u *S3Utils) ValidateImageForBedrock(ctx context.Context, s3Url string) error {
	// Parse URL
	if _, err := u.ParseS3URL(s3Url); err != nil {
		return fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Get image metadata
	metadata, err := u.GetImageMetadata(ctx, s3Url)
	if err != nil {
		return fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Validate content type
	contentType := metadata["contentType"]
	if !u.IsImageContentType(contentType) {
		return fmt.Errorf("invalid content type: %s, expected image/png or image/jpeg", contentType)
	}

	// Bedrock specifically supports PNG and JPEG
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/jpg" {
		u.logger.Warn("Content type may not be fully supported by Bedrock", map[string]interface{}{
			"contentType": contentType,
			"s3Url": s3Url,
		})
	}

	// Parse and validate size
	contentLength := metadata["contentLength"]
	if contentLength == "" {
		return fmt.Errorf("content length not available")
	}

	size := int64(0)
	if _, err := fmt.Sscanf(contentLength, "%d", &size); err != nil {
		return fmt.Errorf("invalid content length: %s", contentLength)
	}
	
	// Bedrock limit is typically 100MB per image
	const maxSize = 100 * 1024 * 1024 // 100MB
	if size > maxSize {
		return fmt.Errorf("image too large: %d bytes (max %d bytes for Bedrock)", size, maxSize)
	}

	// Note: We no longer fail validation if bucket owner is missing
	// Let Bedrock handle this gracefully
	if metadata["bucketOwner"] == "" {
		u.logger.Warn("Bucket owner not found, but continuing validation", map[string]interface{}{
			"s3Url": s3Url,
		})
	}

	u.logger.Info("Image validated for Bedrock", map[string]interface{}{
		"s3Url":        s3Url,
		"contentType":  contentType,
		"size":         size,
		"bucketOwner":  metadata["bucketOwner"],
	})

	return nil
}
