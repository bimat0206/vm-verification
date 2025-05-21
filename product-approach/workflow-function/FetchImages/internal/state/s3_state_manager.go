package state

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3StateManager handles state management for the FetchImages function
type S3StateManager struct {
	client *s3.Client
	logger logger.Logger
	bucket string
}

// NewS3StateManager creates a new S3StateManager instance
func NewS3StateManager(client *s3.Client, log logger.Logger, bucket string) *S3StateManager {
	return &S3StateManager{
		client: client,
		logger: log.WithFields(map[string]interface{}{"component": "S3StateManager"}),
		bucket: bucket,
	}
}

// StoreImageMetadata stores image metadata in S3
func (m *S3StateManager) StoreImageMetadata(ctx context.Context, verificationId string, metadata *schema.ImageMetadata) error {
	key := fmt.Sprintf("%s/images/metadata.json", verificationId)
	
	m.logger.Info("Storing image metadata", map[string]interface{}{
		"verificationId": verificationId,
		"key":           key,
	})

	// Convert metadata to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Upload to S3
	_, err = m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(jsonData)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	m.logger.Info("Successfully stored image metadata", map[string]interface{}{
		"verificationId": verificationId,
		"key":           key,
	})

	return nil
}

// StoreBase64Images stores both reference and checking Base64 images
func (m *S3StateManager) StoreBase64Images(ctx context.Context, verificationId string, referenceBase64, checkingBase64 string) error {
	// Store reference image
	if err := m.storeBase64Image(ctx, verificationId, "reference", referenceBase64); err != nil {
		return fmt.Errorf("failed to store reference Base64 image: %w", err)
	}

	// Store checking image
	if err := m.storeBase64Image(ctx, verificationId, "checking", checkingBase64); err != nil {
		return fmt.Errorf("failed to store checking Base64 image: %w", err)
	}

	return nil
}

// storeBase64Image stores a single Base64 image
func (m *S3StateManager) storeBase64Image(ctx context.Context, verificationId, imageType, base64Str string) error {
	key := fmt.Sprintf("%s/images/%s-base64.json", verificationId, imageType)
	
	m.logger.Info("Storing Base64 image", map[string]interface{}{
		"verificationId": verificationId,
		"imageType":      imageType,
		"key":           key,
	})

	// Create a JSON structure for the Base64 data
	jsonData := fmt.Sprintf(`{"base64": "%s"}`, base64Str)

	// Upload to S3
	_, err := m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to store Base64 image: %w", err)
	}

	m.logger.Info("Successfully stored Base64 image", map[string]interface{}{
		"verificationId": verificationId,
		"imageType":      imageType,
		"key":           key,
	})

	return nil
}

// GetImageMetadata retrieves image metadata from S3
func (m *S3StateManager) GetImageMetadata(ctx context.Context, verificationId string) (*schema.ImageMetadata, error) {
	key := fmt.Sprintf("%s/images/metadata.json", verificationId)
	
	m.logger.Info("Retrieving image metadata", map[string]interface{}{
		"verificationId": verificationId,
		"key":           key,
	})

	// Get the object from S3
	result, err := m.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer result.Body.Close()

	// Parse the JSON data
	var metadata schema.ImageMetadata
	if err := json.NewDecoder(result.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	m.logger.Info("Successfully retrieved image metadata", map[string]interface{}{
		"verificationId": verificationId,
		"key":           key,
	})

	return &metadata, nil
} 