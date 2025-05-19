// Package repository provides data access implementations for the FetchImages function
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// S3Repository handles S3 operations for image metadata
type S3Repository struct {
	client *s3.Client
	logger logger.Logger
}

// NewS3Repository creates a new S3Repository instance
func NewS3Repository(client *s3.Client, log logger.Logger) *S3Repository {
	return &S3Repository{
		client: client,
		logger: log.WithFields(map[string]interface{}{"component": "S3Repository"}),
	}
}

// FetchImageMetadata retrieves metadata for an image from S3
func (r *S3Repository) FetchImageMetadata(ctx context.Context, s3url string) (*schema.ImageInfo, error) {
	// Parse the S3 URL to extract bucket and key
	parsed, err := ParseS3URL(s3url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 URL: %w", err)
	}
	
	r.logger.Info("Fetching S3 image metadata", map[string]interface{}{
		"url":    s3url,
		"bucket": parsed.Bucket, 
		"key":    parsed.Key,
	})

	// Get the object metadata using HeadObject
	headOutput, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Create and populate image info
	imgInfo := &schema.ImageInfo{
		URL:      s3url,
		S3Bucket: parsed.Bucket,
		S3Key:    parsed.Key,
	}
	
	// Set size (ContentLength is a pointer, so we need to check if it's nil)
	if headOutput.ContentLength != nil {
		imgInfo.Size = *headOutput.ContentLength
	}
	
	// Set content type
	if headOutput.ContentType != nil {
		imgInfo.ContentType = *headOutput.ContentType
		
		// Set image format based on content type
		switch *headOutput.ContentType {
		case "image/png":
			imgInfo.Format = "png"
		case "image/jpeg", "image/jpg":
			imgInfo.Format = "jpeg"
		case "image/webp":
			imgInfo.Format = "webp"
		default:
			imgInfo.Format = "unknown"
		}
	}
	
	// Set last modified
	if headOutput.LastModified != nil {
		imgInfo.LastModified = headOutput.LastModified.Format(time.RFC3339)
	}
	
	// Set ETag (remove quotes if present)
	if headOutput.ETag != nil {
		etag := *headOutput.ETag
		if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		imgInfo.ETag = etag
	}

	// Configure S3-only storage
	imgInfo.StorageMethod = schema.StorageMethodS3Temporary
	imgInfo.Base64Generated = true
	imgInfo.StorageDecisionAt = schema.FormatISO8601()

	r.logger.Info("Successfully fetched image metadata", map[string]interface{}{
		"url":         s3url,
		"contentType": imgInfo.ContentType,
		"size":        imgInfo.Size,
		"format":      imgInfo.Format,
	})

	return imgInfo, nil
}

// S3URL represents a parsed S3 URL
type S3URL struct {
	Bucket string
	Key    string
	Region string
}

// ParseS3URL parses an S3 URL into bucket and key components
func ParseS3URL(s3url string) (S3URL, error) {
	// This is a simplified placeholder - in production, you would implement
	// a more robust URL parser that handles various S3 URL formats

	// For example:
	// s3://bucket-name/path/to/object
	// https://bucket-name.s3.region.amazonaws.com/path/to/object
	// https://s3.region.amazonaws.com/bucket-name/path/to/object
	
	// For demonstration, we're just handling the s3:// format
	if !strings.HasPrefix(s3url, "s3://") {
		return S3URL{}, fmt.Errorf("unsupported S3 URL format: %s", s3url)
	}
	
	parts := strings.SplitN(s3url[5:], "/", 2)
	if len(parts) != 2 {
		return S3URL{}, fmt.Errorf("invalid S3 URL format: %s", s3url)
	}
	
	return S3URL{
		Bucket: parts[0],
		Key:    parts[1],
	}, nil
}