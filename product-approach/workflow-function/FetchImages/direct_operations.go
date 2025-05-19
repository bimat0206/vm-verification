package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
)

// Constants for data processing
const (
	// Base64ExpansionFactor is the approximate expansion factor when converting binary to Base64
	Base64ExpansionFactor = 1.37
	
	// MaxUsableResponseSize is the maximum usable response size for Lambda (6MB)
	// This accounts for the Lambda response limit of 6MB minus overhead for JSON structure
	MaxUsableResponseSize = 6 * 1024 * 1024
	
	// ResponseOverheadBuffer is the buffer for JSON structure overhead (256KB)
	ResponseOverheadBuffer = 256 * 1024
)

// GetS3ImageWithBase64 downloads the image from S3 and returns metadata with Base64 encoded data
// using the dynamic response size calculation approach
func GetS3ImageWithBase64(
	ctx context.Context,
	s3Client *s3.Client,
	s3url string,
	config ConfigVars,
	responseSizeTracker *ResponseSizeTracker,
	log logger.Logger,
) (ImageMetadata, error) {
	log.Info("Downloading and encoding S3 image with dynamic response size calculation", map[string]interface{}{
		"s3url":                s3url,
		"maxInlineBase64Size":  config.MaxInlineBase64Size,
		"maxUsableResponseSize": MaxUsableResponseSize,
		"tempBase64Bucket":     config.TempBase64Bucket,
	})

	// Parse S3 URL
	parsed, err := ParseS3URL(s3url)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to parse S3 URL: %w", err)
	}

	// Validate image size before downloading
	if err := validateImageSize(ctx, s3Client, parsed, config.MaxImageSize, log); err != nil {
		return ImageMetadata{}, fmt.Errorf("image validation failed: %w", err)
	}

	// Get metadata
	metadata, err := getImageMetadata(ctx, s3Client, parsed)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Estimate Base64 size before downloading
	estimatedBase64Size := int64(float64(metadata.Size) * Base64ExpansionFactor)
	
	log.Debug("Estimated Base64 size", map[string]interface{}{
		"originalSize":         metadata.Size,
		"estimatedBase64Size":  estimatedBase64Size,
		"expansionFactor":      Base64ExpansionFactor,
	})

	// Download the image
	log.Debug("Starting image download", map[string]interface{}{
		"bucket": parsed.Bucket,
		"key":    parsed.Key,
		"size":   metadata.Size,
	})

	getOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to download image: %w", err)
	}
	defer getOutput.Body.Close()

	// Read image bytes
	imageBytes, err := io.ReadAll(getOutput.Body)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to read image bytes: %w", err)
	}

	// Convert to Base64
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	actualBase64Size := int64(len(base64Data))

	// Determine storage method with dynamic response size consideration
	storageMethod := determineStorageMethodWithResponseSize(actualBase64Size, s3url, config.MaxInlineBase64Size, responseSizeTracker, log)
	
	log.Info("Determined storage method with response size calculation", map[string]interface{}{
		"actualBase64Size":      actualBase64Size,
		"estimatedBase64Size":   estimatedBase64Size,
		"storageMethod":         storageMethod,
		"currentTotalSize":      responseSizeTracker.GetTotalSize(),
		"maxUsableResponseSize": MaxUsableResponseSize,
		"wouldExceedLimit":      responseSizeTracker.WouldExceedLimit(actualBase64Size),
	})

	// Store Base64 data according to determined method
	if err := storeBase64Data(ctx, s3Client, &metadata, base64Data, storageMethod, config.TempBase64Bucket, log); err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to store Base64 data: %w", err)
	}

	// Update response size tracker
	updateResponseSizeTracker(actualBase64Size, storageMethod, s3url, responseSizeTracker)

	// Update metadata with format information
	metadata.UpdateImageFormat()
	metadata.UpdateBedrockFormat()

	log.Info("Successfully processed image with dynamic storage", map[string]interface{}{
		"contentType":           metadata.ContentType,
		"size":                  metadata.Size,
		"actualBase64Size":      actualBase64Size,
		"storageMethod":         metadata.StorageMethod,
		"imageFormat":           metadata.ImageFormat,
		"currentTotalBase64":    responseSizeTracker.GetTotalSize(),
		"responseUtilization":   float64(responseSizeTracker.GetTotalSize()) / float64(MaxUsableResponseSize) * 100,
	})

	return metadata, nil
}

// determineStorageMethodWithResponseSize determines storage method considering total response size
func determineStorageMethodWithResponseSize(
	base64Length int64,
	imageURL string,
	maxInlineBase64Size int64,
	responseSizeTracker *ResponseSizeTracker,
	log logger.Logger,
) string {
	// First check if this single image would exceed our usable response size
	if base64Length > MaxUsableResponseSize {
		log.Warn("Single image Base64 exceeds usable response size", map[string]interface{}{
			"base64Length":         base64Length,
			"maxUsableResponseSize": MaxUsableResponseSize,
			"imageURL":             imageURL,
		})
		return StorageMethodS3Temporary
	}

	// Check if adding this Base64 would exceed the total response size limit
	if responseSizeTracker.WouldExceedLimit(base64Length) {
		log.Info("Using S3 storage due to total response size limit", map[string]interface{}{
			"base64Length":         base64Length,
			"currentTotalSize":     responseSizeTracker.GetTotalSize(),
			"wouldBeTotalSize":     responseSizeTracker.GetTotalSize() + base64Length,
			"maxUsableResponseSize": MaxUsableResponseSize,
			"imageURL":             imageURL,
		})
		return StorageMethodS3Temporary
	}

	// Apply individual image threshold
	if base64Length <= maxInlineBase64Size {
		return StorageMethodInline
	}

	log.Info("Using S3 storage due to individual size threshold", map[string]interface{}{
		"base64Length":         base64Length,
		"maxInlineBase64Size":  maxInlineBase64Size,
		"imageURL":             imageURL,
	})
	return StorageMethodS3Temporary
}

// updateResponseSizeTracker updates the response size tracker based on storage method
func updateResponseSizeTracker(
	base64Size int64,
	storageMethod string,
	imageURL string,
	responseSizeTracker *ResponseSizeTracker,
) {
	// Only count toward response size if stored inline
	if storageMethod == StorageMethodInline {
		// Determine if this is reference or checking image based on URL
		if strings.Contains(imageURL, "reference") || strings.Contains(imageURL, "processed") {
			responseSizeTracker.UpdateReferenceSize(base64Size)
		} else {
			responseSizeTracker.UpdateCheckingSize(base64Size)
		}
	}
}

// storeBase64Data stores Base64 data according to the specified method
func storeBase64Data(
	ctx context.Context,
	s3Client *s3.Client,
	metadata *ImageMetadata,
	base64Data string,
	storageMethod string,
	tempBase64Bucket string,
	log logger.Logger,
) error {
	switch storageMethod {
	case StorageMethodInline:
		metadata.SetInlineStorage(base64Data)
		log.Debug("Stored Base64 data inline", map[string]interface{}{
			"length": len(base64Data),
		})
		return nil
		
	case StorageMethodS3Temporary:
		if tempBase64Bucket == "" {
			return fmt.Errorf("temporary S3 bucket not configured for large Base64 storage")
		}
		
		// Generate unique key for temporary storage
		tempKey, err := generateTempKey(metadata)
		if err != nil {
			return fmt.Errorf("failed to generate temporary key: %w", err)
		}
		
		// Store in S3
		if err := storeBase64InS3(ctx, s3Client, base64Data, tempBase64Bucket, tempKey); err != nil {
			return fmt.Errorf("failed to store Base64 in S3: %w", err)
		}
		
		metadata.SetS3Storage(tempBase64Bucket, tempKey)
		log.Debug("Stored Base64 data in S3", map[string]interface{}{
			"bucket": tempBase64Bucket,
			"key":    tempKey,
			"length": len(base64Data),
		})
		return nil
		
	default:
		return fmt.Errorf("unknown storage method: %s", storageMethod)
	}
}

// storeBase64InS3 stores Base64 data in the temporary S3 bucket
func storeBase64InS3(
	ctx context.Context,
	s3Client *s3.Client,
	base64Data string,
	bucket string,
	key string,
) error {
	// Add lifecycle metadata for automatic cleanup
	metadata := map[string]string{
		"Content-Type":       "text/plain",
		"x-amz-meta-type":    "base64-data",
		"x-amz-meta-created": time.Now().UTC().Format(time.RFC3339),
		"x-amz-meta-size":    fmt.Sprintf("%d", len(base64Data)),
	}
	
	putInput := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     strings.NewReader(base64Data),
		Metadata: metadata,
		// Set server-side encryption
		ServerSideEncryption: "AES256",
	}
	
	// Add lifecycle configuration for automatic cleanup (24 hours)
	// This would typically be configured at the bucket level
	
	_, err := s3Client.PutObject(ctx, putInput)
	return err
}

// generateTempKey generates a unique key for temporary Base64 storage
func generateTempKey(metadata *ImageMetadata) (string, error) {
	// Generate random suffix for uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomSuffix := base64.URLEncoding.EncodeToString(randomBytes)[:8]
	
	// Create structured key path
	timestamp := time.Now().UTC().Format("2006-01-02/15-04-05")
	imageFormat := metadata.GetImageFormat()
	
	key := fmt.Sprintf("temp-base64/%s/%s-%s.base64", 
		timestamp, imageFormat, randomSuffix)
	
	return key, nil
}

// validateImageSize checks the image size before downloading
func validateImageSize(
	ctx context.Context,
	s3Client *s3.Client,
	parsed S3URL,
	maxImageSize int64,
	log logger.Logger,
) error {
	// Get object metadata to check size before downloading
	headOutput, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object metadata: %w", err)
	}
	
	// Check size
	if headOutput.ContentLength != nil && *headOutput.ContentLength > maxImageSize {
		return fmt.Errorf("image too large: %d bytes (max %d bytes)", 
			*headOutput.ContentLength, maxImageSize)
	}
	
	// Validate content type
	if headOutput.ContentType != nil {
		contentType := *headOutput.ContentType
		if !IsImageContentType(contentType) {
			return fmt.Errorf("invalid content type: %s (expected image type)", contentType)
		}
	}
	
	return nil
}

// getImageMetadata gets object metadata and creates ImageMetadata struct
func getImageMetadata(
	ctx context.Context,
	s3Client *s3.Client,
	parsed S3URL,
) (ImageMetadata, error) {
	// Get object metadata using HeadObject
	headOutput, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(parsed.Bucket),
		Key:    aws.String(parsed.Key),
	})
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Create metadata object
	metadata := ImageMetadata{
		Bucket: parsed.Bucket,
		Key:    parsed.Key,
	}

	// Set content type
	if headOutput.ContentType != nil {
		metadata.ContentType = *headOutput.ContentType
	} else {
		metadata.ContentType = "application/octet-stream"
	}
	
	// Set size
	if headOutput.ContentLength != nil {
		metadata.Size = *headOutput.ContentLength
	}
	
	// Set last modified
	if headOutput.LastModified != nil {
		metadata.LastModified = headOutput.LastModified.Format("2006-01-02T15:04:05Z")
	}
	
	// Set ETag (remove quotes if present)
	if headOutput.ETag != nil {
		etag := *headOutput.ETag
		if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		metadata.ETag = etag
	}

	return metadata, nil
}

// ConvertImageToS3Storage converts an inline-stored image to S3 storage
func ConvertImageToS3Storage(
	ctx context.Context,
	s3Client *s3.Client,
	config ConfigVars,
	meta *ImageMetadata,
	tracker *ResponseSizeTracker,
	log logger.Logger,
) error {
	if meta.StorageMethod != StorageMethodInline {
		return fmt.Errorf("image is not stored inline, cannot convert")
	}
	
	if meta.Base64Data == "" {
		return fmt.Errorf("no inline Base64 data to convert")
	}
	
	log.Info("Converting inline storage to S3 storage", map[string]interface{}{
		"currentSize": len(meta.Base64Data),
		"bucket":      meta.Bucket,
		"key":         meta.Key,
	})
	
	// Store current inline data
	base64Data := meta.Base64Data
	
	// Generate new temporary key
	tempKey, err := generateTempKey(meta)
	if err != nil {
		return fmt.Errorf("failed to generate temporary key: %w", err)
	}
	
	// Store in S3
	if err := storeBase64InS3(ctx, s3Client, base64Data, config.TempBase64Bucket, tempKey); err != nil {
		return fmt.Errorf("failed to convert to S3 storage: %w", err)
	}
	
	// Update metadata to reflect S3 storage
	meta.SetS3Storage(config.TempBase64Bucket, tempKey)
	
	// Update response size tracker to remove the inline data size
	imageSize := int64(len(base64Data))
	tracker.mu.Lock()
	if meta.Bucket != "" && strings.Contains(meta.Bucket, "reference") {
		tracker.referenceBase64Size = 0
	} else {
		tracker.checkingBase64Size = 0
	}
	tracker.updateTotalSize()
	tracker.mu.Unlock()
	
	log.Info("Successfully converted to S3 storage", map[string]interface{}{
		"tempBucket":        config.TempBase64Bucket,
		"tempKey":           tempKey,
		"convertedSize":     imageSize,
		"newTotalBase64":    tracker.GetTotalSize(),
	})
	
	return nil
}

// ValidateLayoutExists checks if a layout exists before attempting to fetch it
func ValidateLayoutExists(
	ctx context.Context,
	dbClient *dynamodb.Client,
	tableName string,
	layoutId int,
	layoutPrefix string,
) (bool, error) {
	// Create the key for the DynamoDB query
	key := map[string]types.AttributeValue{
		"layoutId": &types.AttributeValueMemberN{
			Value: strconv.Itoa(layoutId),
		},
		"layoutPrefix": &types.AttributeValueMemberS{
			Value: layoutPrefix,
		},
	}
	
	// Get the item from DynamoDB
	getItemOutput, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
		ProjectionExpression: aws.String("layoutId"), // Only retrieve the key to minimize data transfer
	})
	
	if err != nil {
		return false, fmt.Errorf("failed to check if layout exists: %w", err)
	}
	
	// Check if the item exists
	return getItemOutput.Item != nil, nil
}

// FetchLayoutMetadataWithFallback attempts to fetch layout metadata with graceful fallback
func FetchLayoutMetadataWithFallback(
	ctx context.Context,
	dbClient *dynamodb.Client,
	tableName string,
	layoutId int,
	layoutPrefix string,
) (map[string]interface{}, error) {
	// Try to fetch the complete layout metadata
	result, err := FetchLayoutMetadata(ctx, dbClient, tableName, layoutId, layoutPrefix)
	if err != nil {
		// If the layout doesn't exist, return a minimal structure
		// This can happen if the layout hasn't been fully processed yet
		return map[string]interface{}{
			"layoutId":        layoutId,
			"layoutPrefix":    layoutPrefix,
			"status":          "partial",
			"error":           err.Error(),
			"machineStructure": map[string]interface{}{
				"rowCount":      6,  // Default values
				"columnsPerRow": 10, // Default values
				"rowOrder":      []string{"A", "B", "C", "D", "E", "F"},
				"columnOrder":   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			"productPositionMap": make(map[string]interface{}),
		}, nil
	}
	
	// Add status to indicate successful retrieval
	result["status"] = "complete"
	return result, nil
}

// FetchLayoutMetadata retrieves complete layout metadata from DynamoDB.
func FetchLayoutMetadata(
	ctx context.Context,
	dbClient *dynamodb.Client,
	tableName string,
	layoutId int,
	layoutPrefix string,
) (map[string]interface{}, error) {
	// Create the key for the DynamoDB query
	key := map[string]types.AttributeValue{
		"layoutId": &types.AttributeValueMemberN{
			Value: strconv.Itoa(layoutId),
		},
		"layoutPrefix": &types.AttributeValueMemberS{
			Value: layoutPrefix,
		},
	}
	
	// Get the item from DynamoDB
	getItemOutput, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve layout metadata from DynamoDB: %w", err)
	}
	
	// Check if the item exists
	if getItemOutput.Item == nil {
		return nil, fmt.Errorf("layout not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix)
	}
	
	// Unmarshal the DynamoDB item into a LayoutMetadata struct
	var layout LayoutMetadata
	err = attributevalue.UnmarshalMap(getItemOutput.Item, &layout)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal layout metadata: %w", err)
	}
	
	// Convert the structured layout metadata to a map for the response
	result := map[string]interface{}{
		"layoutId":           layout.LayoutId,
		"layoutPrefix":       layout.LayoutPrefix,
		"vendingMachineId":   layout.VendingMachineId,
		"location":           layout.Location,
		"createdAt":          layout.CreatedAt,
		"updatedAt":          layout.UpdatedAt,
		"referenceImageUrl":  layout.ReferenceImageUrl,
		"sourceJsonUrl":      layout.SourceJsonUrl,
		"machineStructure":   layout.MachineStructure,
		"productPositionMap": layout.ProductPositionMap,
	}

	// Add derived fields for convenience
	if machineStruct := layout.MachineStructure; machineStruct != nil {
		if rowCount, exists := machineStruct["rowCount"]; exists {
			result["derivedRowCount"] = rowCount
		}
		if columnsPerRow, exists := machineStruct["columnsPerRow"]; exists {
			result["derivedColumnsPerRow"] = columnsPerRow
		}
	}

	// Add product count if productPositionMap exists
	if prodMap := layout.ProductPositionMap; prodMap != nil {
		result["totalProductPositions"] = len(prodMap)
	}

	return result, nil
}

// FetchHistoricalContext retrieves previous verification from DynamoDB.
func FetchHistoricalContext(
	ctx context.Context,
	dbClient *dynamodb.Client,
	tableName string,
	verificationId string,
) (map[string]interface{}, error) {
	// Create the key for the DynamoDB query
	key := map[string]types.AttributeValue{
		"verificationId": &types.AttributeValueMemberS{
			Value: verificationId,
		},
	}
	
	// Get the item from DynamoDB
	getItemOutput, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical verification: %w", err)
	}
	
	// Check if the item exists
	if getItemOutput.Item == nil {
		return nil, fmt.Errorf("verification not found: verificationId=%s", verificationId)
	}
	
	// Unmarshal the DynamoDB item into a VerificationRecord struct
	var verificationCtx VerificationRecord
	err = attributevalue.UnmarshalMap(getItemOutput.Item, &verificationCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification record: %w", err)
	}
	
	// Convert the verification record to a historical context map
	return verificationCtx.ToHistoricalContext(), nil
}

// UpdateVerificationStatus updates verification status in DynamoDB.
func UpdateVerificationStatus(
	ctx context.Context,
	dbClient *dynamodb.Client,
	tableName string,
	verificationId string,
	status string,
) error {
	// Create the key for the DynamoDB update
	key := map[string]types.AttributeValue{
		"verificationId": &types.AttributeValueMemberS{
			Value: verificationId,
		},
	}
	
	// Create the update expression
	updateExpression := "SET #status = :status"
	expressionAttributeNames := map[string]string{
		"#status": "status",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{
			Value: status,
		},
	}
	
	// Update the item in DynamoDB
	_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	
	if err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}
	
	return nil
}
