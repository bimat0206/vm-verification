// Package repository provides data access implementations for the FetchImages function
package repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
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
	// Check for empty URL
	if s3url == "" {
		return S3URL{}, fmt.Errorf("S3 URL is empty")
	}

	// Handle s3:// format
	if strings.HasPrefix(s3url, "s3://") {
		parts := strings.SplitN(s3url[5:], "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return S3URL{}, fmt.Errorf("invalid S3 URL format: %s", s3url)
		}
		return S3URL{
			Bucket: parts[0],
			Key:    parts[1],
		}, nil
	}

	// Handle https://bucket-name.s3.region.amazonaws.com/path/to/object format
	if strings.HasPrefix(s3url, "https://") {
		// Remove https:// prefix
		url := s3url[8:]
		
		// Check for bucket-name.s3.region.amazonaws.com format
		if strings.Contains(url, ".s3.") && strings.Contains(url, ".amazonaws.com/") {
			parts := strings.SplitN(url, "/", 2)
			if len(parts) != 2 {
				return S3URL{}, fmt.Errorf("invalid HTTPS S3 URL format: %s", s3url)
			}
			
			// Extract bucket name from hostname
			hostParts := strings.Split(parts[0], ".")
			if len(hostParts) < 4 {
				return S3URL{}, fmt.Errorf("invalid S3 hostname format: %s", parts[0])
			}
			
			bucket := hostParts[0]
			key := parts[1]
			
			if bucket == "" || key == "" {
				return S3URL{}, fmt.Errorf("empty bucket or key in S3 URL: %s", s3url)
			}
			
			return S3URL{
				Bucket: bucket,
				Key:    key,
			}, nil
		}
		
		// Check for s3.region.amazonaws.com/bucket-name/path format
		if strings.Contains(url, "s3.") && strings.Contains(url, ".amazonaws.com/") {
			parts := strings.SplitN(url, "/", 3)
			if len(parts) < 3 {
				return S3URL{}, fmt.Errorf("invalid path-style S3 URL format: %s", s3url)
			}
			
			bucket := parts[1]
			key := parts[2]
			
			if bucket == "" || key == "" {
				return S3URL{}, fmt.Errorf("empty bucket or key in path-style S3 URL: %s", s3url)
			}
			
			return S3URL{
				Bucket: bucket,
				Key:    key,
			}, nil
		}
	}

	return S3URL{}, fmt.Errorf("unsupported S3 URL format: %s (supported formats: s3://, https://bucket.s3.region.amazonaws.com/, https://s3.region.amazonaws.com/bucket/)", s3url)
}

// DownloadAndConvertToBase64 downloads an image from S3 and returns its Base64 representation and metadata
func (r *S3Repository) DownloadAndConvertToBase64(ctx context.Context, s3url string) (string, *schema.ImageInfo, error) {
	parsed, err := ParseS3URL(s3url)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse S3 URL: %w", err)
	}

	r.logger.Info("Downloading image from S3", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
	})

	obj, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to download object: %w", err)
	}
	defer obj.Body.Close()

	data, err := io.ReadAll(obj.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read object body: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(data)

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		// image.DecodeConfig may fail for unsupported formats; log but continue
		r.logger.Error("Failed to decode image config", map[string]interface{}{"error": err.Error()})
	}

	info := &schema.ImageInfo{
		URL:               s3url,
		S3Bucket:          parsed.Bucket,
		S3Key:             parsed.Key,
		Size:              int64(len(data)),
		Base64Size:        int64(len(base64Str)),
		Width:             cfg.Width,
		Height:            cfg.Height,
		Format:            format,
		StorageMethod:     schema.StorageMethodS3Temporary,
		Base64Generated:   true,
		StorageDecisionAt: schema.FormatISO8601(),
	}

	if obj.ContentType != nil {
		info.ContentType = *obj.ContentType
	}
	if obj.LastModified != nil {
		info.LastModified = obj.LastModified.Format(time.RFC3339)
	}
	if obj.ETag != nil {
		etag := *obj.ETag
		if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		info.ETag = etag
	}

	// If format from DecodeConfig failed, try to derive from content type
	if info.Format == "" || info.Format == "unknown" {
		switch info.ContentType {
		case "image/png":
			info.Format = "png"
		case "image/jpeg", "image/jpg":
			info.Format = "jpeg"
		case "image/gif":
			info.Format = "gif"
		case "image/webp":
			info.Format = "webp"
		default:
			info.Format = format
		}
	}

	r.logger.Info("Successfully downloaded and encoded image", map[string]interface{}{
		"key":    parsed.Key,
		"size":   info.Size,
		"format": info.Format,
	})

	return base64Str, info, nil
}
