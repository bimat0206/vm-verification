package s3utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"api_images_upload_render/config"
	"api_images_upload_render/renderer"
)

func NewS3Client(ctx context.Context) (*s3.Client, error) {
	// Get config to use region setting
	cfg := config.GetConfig()
	
	// Load configuration with region from config
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, 
		awsconfig.WithRegion(cfg.S3Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}
	return s3.NewFromConfig(awsCfg), nil
}

// GetObjectSize returns the size of an S3 object in bytes
func GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	s3Client, err := NewS3Client(ctx)
	if err != nil {
		return 0, err
	}
	
	headOutput, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get head object for %s: %v", key, err)
	}
	
	if headOutput.ContentLength == nil {
		return 0, fmt.Errorf("content length is nil for object %s", key)
	}
	return *headOutput.ContentLength, nil
}

func IsValidRawFile(objectKey string) bool {
	return strings.HasPrefix(objectKey, "raw/") && filepath.Ext(objectKey) == ".json"
}

func DownloadAndParseLayout(ctx context.Context, s3Client *s3.Client, bucket, objectKey string) (*renderer.Layout, error) {
	getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %v", objectKey, bucket, err)
	}
	defer getObjectOutput.Body.Close()

	data, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %v", err)
	}

	var layout renderer.Layout
	err = json.Unmarshal(data, &layout)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &layout, nil
}

// GenerateProcessedKey creates a processed key with the correct path structure
// Format: processed/{year}/{month}/{date}/{layoutId}_{layoutPrefix}_reference_image.png
func GenerateProcessedKey(objectKey string, layoutID int64, layoutPrefix string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	date := now.Format("02")
	
	return fmt.Sprintf("processed/%s/%s/%s/%d_%s_reference_image.png", 
		year, month, date, layoutID, layoutPrefix)
}

func UploadImage(ctx context.Context, s3Client *s3.Client, bucket, key string, imgBytes []byte) error {
	contentType := "image/png"
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(imgBytes),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to %s: %v", key, err)
	}
	return nil
}

func UploadLog(ctx context.Context, s3Client *s3.Client, bucket, key string, logData []byte) error {
	contentType := "application/json"
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(logData),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload log to %s: %v", key, err)
	}
	return nil
}

// UploadFile uploads a file to S3 with the specified content type
func UploadFile(ctx context.Context, s3Client *s3.Client, bucket, key string, data []byte, contentType string) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to %s: %v", key, err)
	}
	return nil
}

// IsLayoutJSON checks if the uploaded JSON file is a valid layout file
func IsLayoutJSON(ctx context.Context, s3Client *s3.Client, bucket, key string) (bool, error) {
	// Download and try to parse as layout
	getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get object %s from bucket %s: %v", key, bucket, err)
	}
	defer getObjectOutput.Body.Close()

	data, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read object data: %v", err)
	}

	var layout renderer.Layout
	err = json.Unmarshal(data, &layout)
	if err != nil {
		return false, nil // Not a valid layout JSON, but not an error
	}

	// Check if it has the required structure for a layout
	if layout.LayoutID <= 0 {
		return false, nil
	}

	if len(layout.SubLayoutList) == 0 {
		return false, nil
	}

	return true, nil
}
