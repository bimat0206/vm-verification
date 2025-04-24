package s3

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"strings"
"time"
"bytes"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ImageService implements image operations using S3
type ImageService struct {
	client           *s3.Client
	referenceBucket  string
	checkingBucket   string
	resultsBucket    string
}

// NewImageService creates a new S3-based image service
func NewImageService(
	region string,
	referenceBucket string,
	checkingBucket string,
	resultsBucket string,
) (*ImageService, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &ImageService{
		client:           client,
		referenceBucket:  referenceBucket,
		checkingBucket:   checkingBucket,
		resultsBucket:    resultsBucket,
	}, nil
}

// parseS3URL parses an S3 URL into bucket and key
func parseS3URL(url string) (bucket, key string, err error) {
	if !strings.HasPrefix(url, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URL: %s", url)
	}

	parts := strings.TrimPrefix(url, "s3://")
	firstSlash := strings.IndexByte(parts, '/')
	if firstSlash == -1 {
		return "", "", fmt.Errorf("invalid S3 URL format: %s", url)
	}

	bucket = parts[:firstSlash]
	key = parts[firstSlash+1:]

	return bucket, key, nil
}

// FetchReferenceImage retrieves a reference image from S3
func (s *ImageService) FetchReferenceImage(ctx context.Context, url string) ([]byte, map[string]interface{}, error) {
	// Parse S3 URL
	bucket, key, err := parseS3URL(url)
	if err != nil {
		return nil, nil, err
	}

	// Get object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get reference image from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the object body
	imageBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read reference image: %w", err)
	}

	// Parse metadata path
	metadataKey := filepath.Join(filepath.Dir(key), "metadata.json")

	// Get metadata from S3
	metadataResult, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(metadataKey),
	})

	var layoutMetadata map[string]interface{}
	if err == nil {
		defer metadataResult.Body.Close()

		// Read the metadata JSON
		metadataBytes, err := io.ReadAll(metadataResult.Body)
		if err == nil {
			// Parse the metadata JSON
			if err := json.Unmarshal(metadataBytes, &layoutMetadata); err != nil {
				// Log the error but continue without metadata
				fmt.Printf("Failed to parse layout metadata: %v\n", err)
			}
		}
	}

	// If no metadata found, use an empty map
	if layoutMetadata == nil {
		layoutMetadata = make(map[string]interface{})
	}

	// Add image metadata
	imageMetadata, err := s.GetImageMetadata(ctx, imageBytes)
	if err == nil {
		// Merge image metadata with layout metadata
		for k, v := range imageMetadata {
			layoutMetadata[k] = v
		}
	}

	return imageBytes, layoutMetadata, nil
}

// FetchCheckingImage retrieves a checking image from S3
func (s *ImageService) FetchCheckingImage(ctx context.Context, url string) ([]byte, error) {
	// Parse S3 URL
	bucket, key, err := parseS3URL(url)
	if err != nil {
		return nil, err
	}

	// Get object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get checking image from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the object body
	imageBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read checking image: %w", err)
	}

	return imageBytes, nil
}

// GetImageMetadata extracts metadata from an image
func (s *ImageService) GetImageMetadata(ctx context.Context, imageBytes []byte) (map[string]interface{}, error) {
	// In a real implementation, this would use an image processing library
	// to extract metadata such as dimensions, format, etc.

	// For this example, we'll just return basic metadata
	metadata := map[string]interface{}{
		"format": detectImageFormat(imageBytes),
		"size":   len(imageBytes),
		// In a real implementation, you would include width and height
	}

	return metadata, nil
}

// detectImageFormat detects the format of an image from its bytes
func detectImageFormat(imageBytes []byte) string {
	if len(imageBytes) < 4 {
		return "unknown"
	}

	// Check for PNG
	if imageBytes[0] == 0x89 && imageBytes[1] == 0x50 && imageBytes[2] == 0x4E && imageBytes[3] == 0x47 {
		return "png"
	}

	// Check for JPEG
	if imageBytes[0] == 0xFF && imageBytes[1] == 0xD8 {
		return "jpeg"
	}

	// Check for GIF
	if imageBytes[0] == 0x47 && imageBytes[1] == 0x49 && imageBytes[2] == 0x46 {
		return "gif"
	}

	return "unknown"
}

// ImageToBase64 converts an image to base64
func (s *ImageService) ImageToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}

// StoreResultImage stores a result image in S3
func (s *ImageService) StoreResultImage(ctx context.Context, verificationID string, imageBytes []byte) (string, error) {
	// Generate key for result image
	timestamp := time.Now().Format("2006/01/02")
	key := fmt.Sprintf("%s/%s/result.jpg", timestamp, verificationID)

	// Put object in S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.resultsBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageBytes),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to store result image in S3: %w", err)
	}

	// Return S3 URL
	return fmt.Sprintf("s3://%s/%s", s.resultsBucket, key), nil
}