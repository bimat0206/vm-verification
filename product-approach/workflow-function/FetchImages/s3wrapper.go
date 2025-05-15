package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// S3UtilsWrapper wraps the shared s3utils package with specific functionality needed for FetchImages
type S3UtilsWrapper struct {
	s3utils   *s3utils.S3Utils
	s3Client  *s3.Client
	config    aws.Config // Always have the config available
	logger    logger.Logger
}

// NewS3Utils creates a new S3UtilsWrapper with AWS config
func NewS3Utils(config aws.Config, log logger.Logger) *S3UtilsWrapper {
	s3Client := s3.NewFromConfig(config)
	
	return &S3UtilsWrapper{
		s3utils:  s3utils.NewWithConfig(config, log),
		s3Client: s3Client,
		config:   config, // Store the config
		logger:   log.WithFields(map[string]interface{}{
			"component": "s3wrapper",
		}),
	}
}

// NewS3UtilsWithConfig creates a new S3UtilsWrapper with explicit config (kept for compatibility)
func NewS3UtilsWithConfig(config aws.Config, log logger.Logger) *S3UtilsWrapper {
	return NewS3Utils(config, log)
}

// GetS3ImageMetadata fetches S3 object metadata and converts it to ImageMetadata format
func (u *S3UtilsWrapper) GetS3ImageMetadata(ctx context.Context, s3url string) (ImageMetadata, error) {
	u.logger.Info("Getting S3 image metadata", map[string]interface{}{
		"s3url": s3url,
	})

	// Parse S3 URL
	parsed, err := u.s3utils.ParseS3URL(s3url)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to parse S3 URL: %w", err)
	}

	// Validate image exists and get metadata
	metadata, err := u.getImageMetadataWithBucketOwner(ctx, s3url, parsed)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Parse content length properly
	contentLength, err := strconv.ParseInt(metadata["contentLength"], 10, 64)
	if err != nil {
		u.logger.Warn("Failed to parse content length, defaulting to 0", map[string]interface{}{
			"contentLength": metadata["contentLength"],
			"error": err.Error(),
		})
		contentLength = 0
	}

	// Create the response with Bedrock-compatible format
	result := ImageMetadata{
		ContentType:  metadata["contentType"],
		Size:         contentLength,
		LastModified: metadata["lastModified"],
		ETag:         metadata["etag"],
		BucketOwner:  metadata["bucketOwner"],
		Bucket:       parsed.Bucket,
		Key:          parsed.Key,
		BedrockFormat: u.formatForBedrock(s3url, metadata["bucketOwner"]),
	}

	u.logger.Info("Successfully retrieved S3 image metadata", map[string]interface{}{
		"contentType":   result.ContentType,
		"size":          result.Size,
		"bucketOwner":   result.BucketOwner,
		"lastModified":  result.LastModified,
	})

	return result, nil
}

// getImageMetadataWithBucketOwner extends the shared s3utils function to include bucket owner
func (u *S3UtilsWrapper) getImageMetadataWithBucketOwner(ctx context.Context, s3url string, parsed s3utils.S3URL) (map[string]string, error) {
	// Get object metadata using HeadObject
	headOutput, err := u.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
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
	
	// Handle LastModified
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
	bucketOwner, err := u.getBucketOwner(ctx, parsed.Bucket)
	if err != nil {
		u.logger.Warn("Could not retrieve bucket owner", map[string]interface{}{
			"bucket": parsed.Bucket,
			"error": err.Error(),
		})
		// Don't fail the entire operation for missing bucket owner
		// Set it to empty string and let Bedrock handle it
		bucketOwner = ""
	}
	metadata["bucketOwner"] = bucketOwner

	return metadata, nil
}

// getBucketOwner retrieves the bucket owner using multiple fallback methods
func (u *S3UtilsWrapper) getBucketOwner(ctx context.Context, bucket string) (string, error) {
	u.logger.Debug("Attempting to retrieve bucket owner", map[string]interface{}{
		"bucket": bucket,
		"hasConfig": u.config.Region != "", // Check if config is valid
	})

	// Method 1: Try GetBucketAcl
	aclOutput, err := u.s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
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
	
	// Method 3: Use STS GetCallerIdentity - config should always be available now
	if u.config.Region != "" { // Check if config is valid
		stsClient := sts.NewFromConfig(u.config)
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
		u.logger.Warn("AWS config appears to be invalid", map[string]interface{}{
			"bucket": bucket,
			"configRegion": u.config.Region,
		})
	}
	
	return "", fmt.Errorf("could not determine bucket owner for bucket %s: GetBucketAcl failed (%v), no AWS_ACCOUNT_ID env var, STS failed", bucket, err)
}

// formatForBedrock creates the Bedrock-compatible S3 URI format
func (u *S3UtilsWrapper) formatForBedrock(s3url, bucketOwner string) map[string]interface{} {
	if bucketOwner == "" {
		u.logger.Warn("Bucket owner is empty, Bedrock format may not work correctly", map[string]interface{}{
			"s3url": s3url,
		})
		// Still create the format but without bucketOwner
		return map[string]interface{}{
			"format": "png", // Default to PNG, could be made dynamic
			"source": map[string]interface{}{
				"s3Location": map[string]interface{}{
					"uri": s3url,
					// Omit bucketOwner if not available
				},
			},
		}
	}
	
	return map[string]interface{}{
		"format": "png", // Default to PNG, could be made dynamic
		"source": map[string]interface{}{
			"s3Location": map[string]interface{}{
				"uri":         s3url,
				"bucketOwner": bucketOwner,
			},
		},
	}
}

// ValidateImageForBedrock validates that an image meets Bedrock requirements
func (u *S3UtilsWrapper) ValidateImageForBedrock(ctx context.Context, s3url string) error {
	// Parse URL
	parsed, err := u.s3utils.ParseS3URL(s3url)
	if err != nil {
		return fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Get image metadata
	metadata, err := u.getImageMetadataWithBucketOwner(ctx, s3url, parsed)
	if err != nil {
		return fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Validate content type
	contentType := metadata["contentType"]
	if !u.s3utils.IsImageContentType(contentType) {
		return fmt.Errorf("invalid content type: %s, expected image/png or image/jpeg", contentType)
	}

	// Validate size (Bedrock limit is typically 100MB per image)
	size, err := strconv.ParseInt(metadata["contentLength"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid content length: %s", metadata["contentLength"])
	}
	
	const maxSize = 100 * 1024 * 1024 // 100MB
	if size > maxSize {
		return fmt.Errorf("image too large: %d bytes (max %d bytes for Bedrock)", size, maxSize)
	}

	// Note: We no longer fail validation if bucket owner is missing
	// Let Bedrock handle this gracefully
	if metadata["bucketOwner"] == "" {
		u.logger.Warn("Bucket owner not found, but continuing validation", map[string]interface{}{
			"s3url": s3url,
		})
	}

	u.logger.Info("Image validated for Bedrock", map[string]interface{}{
		"s3url":        s3url,
		"contentType":  contentType,
		"size":         size,
		"bucketOwner":  metadata["bucketOwner"],
	})

	return nil
}