package main

import (
	"context"
	//"fmt"
	"strconv"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// S3UtilsWrapper wraps the shared s3utils package with specific functionality needed for FetchImages
type S3UtilsWrapper struct {
	s3utils *s3utils.S3Utils
}

// NewS3Utils creates a new S3UtilsWrapper
func NewS3Utils(client *s3.Client, log logger.Logger) *S3UtilsWrapper {
	return &S3UtilsWrapper{
		s3utils: s3utils.New(client, log),
	}
}

// GetS3ImageMetadata fetches S3 object metadata and converts it to ImageMetadata format
func (u *S3UtilsWrapper) GetS3ImageMetadata(ctx context.Context, s3url string) (ImageMetadata, error) {
	// Parse S3 URL
	parsed, err := u.s3utils.ParseS3URL(s3url)
	if err != nil {
		return ImageMetadata{}, err
	}

	// Get object metadata
	metadata, err := u.s3utils.GetImageMetadata(ctx, s3url)
	if err != nil {
		return ImageMetadata{}, err
	}

	// Convert to ImageMetadata format
	contentLength, _ := strconv.ParseInt(metadata["contentLength"], 10, 64)
	return ImageMetadata{
		ContentType:  metadata["contentType"],
		Size:         contentLength,
		LastModified: metadata["lastModified"],
		ETag:         metadata["etag"],
		BucketOwner:  metadata["bucketOwner"],
		Bucket:       parsed.Bucket,
		Key:          parsed.Key,
	}, nil
}