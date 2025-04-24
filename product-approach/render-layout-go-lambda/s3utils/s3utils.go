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

	"github.com/aws/aws-lambda-go/events"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"vending-machine-layout-generator/config"
	"vending-machine-layout-generator/renderer"
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

func ParseEvent(event events.EventBridgeEvent) (string, string, error) {
	detail := make(map[string]interface{})
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		return "", "", fmt.Errorf("failed to parse event detail: %v", err)
	}

	bucketInfo, ok := detail["bucket"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("bucket info not found in event")
	}

	objectInfo, ok := detail["object"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("object info not found in event")
	}

	bucket, ok := bucketInfo["name"].(string)
	if !ok {
		return "", "", fmt.Errorf("bucket name not found in event")
	}

	objectKey, ok := objectInfo["key"].(string)
	if !ok {
		return "", "", fmt.Errorf("object key not found in event")
	}

	return bucket, objectKey, nil
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