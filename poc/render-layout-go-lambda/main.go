package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"vending-machine-layout-generator/config"
	"vending-machine-layout-generator/dynamodb"
	"vending-machine-layout-generator/logger"
	"vending-machine-layout-generator/prefix"
	"vending-machine-layout-generator/renderer"
	"vending-machine-layout-generator/s3utils"
)

func handler(ctx context.Context, event events.EventBridgeEvent) error {
	// 1. INPUT VALIDATION
	// Parse event to get bucket and object key
	bucket, objectKey, err := s3utils.ParseEvent(event)
	if err != nil {
		return fmt.Errorf("failed to parse event: %v", err)
	}
	
	// Check if JSON file size is within limits (e.g., < 10 MB) using S3 metadata
	fileSize, err := s3utils.GetObjectSize(ctx, bucket, objectKey)
	if err != nil {
		return fmt.Errorf("failed to get object size: %v", err)
	}
	
	// 10MB limit
	if fileSize > 10*1024*1024 {
		return fmt.Errorf("file size exceeds limit (10MB): %d bytes", fileSize)
	}
	
	// Skip if not a JSON file in the "raw" folder
	if !s3utils.IsValidRawFile(objectKey) {
		return nil
	}
	
	// Dynamically set the bucket name from the event
	config.UpdateBucketName(bucket)
	
	// 2. EVENT PROCESSING
	// Initialize S3 client and logger with the dynamic bucket
	s3Client, err := s3utils.NewS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize S3 client: %v", err)
	}
	log := logger.NewLogger(s3Client, bucket)
	
	// Generate execution ID for traceability
	executionID := fmt.Sprintf("exec-%d", os.Getpid())
	log.Info(ctx, 0, fmt.Sprintf("Starting execution with ID: %s for object: %s", executionID, objectKey))
	
	// 3. LAYOUT FILE PROCESSING
	// Download and parse the layout JSON
	layout, err := s3utils.DownloadAndParseLayout(ctx, s3Client, bucket, objectKey)
	if err != nil {
		log.Error(ctx, 0, fmt.Sprintf("failed to download or parse layout: %v", err))
		return err
	}
	
	// Log start of processing with layout ID
	log.Info(ctx, layout.LayoutID, "starting layout processing")
	
	// Validate layout data - ensure it has required structure
	if layout.LayoutID <= 0 {
		log.Error(ctx, layout.LayoutID, "invalid layout ID")
		return fmt.Errorf("invalid layout ID: %d", layout.LayoutID)
	}
	
	if len(layout.SubLayoutList) == 0 || len(layout.SubLayoutList[0].TrayList) == 0 {
		log.Error(ctx, layout.LayoutID, "layout contains no trays")
		return fmt.Errorf("layout contains no trays")
	}
	
	// 4. UNIQUE IDENTIFIER GENERATION
	// Generate unique layoutPrefix
	prefixGenerator := prefix.NewGenerator()
	layoutPrefix, err := prefixGenerator.GenerateLayoutPrefix()
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to generate layout prefix: %v", err))
		return err
	}
	
	log.Info(ctx, layout.LayoutID, fmt.Sprintf("generated layout prefix: %s", layoutPrefix))
	
	// 5. IMAGE RENDERING
	// Render the layout to an image
	imgBytes, err := renderer.RenderLayoutToBytes(*layout)
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to render layout: %v", err))
		return err
	}
	
	// 6. OUTPUT GENERATION
	// Generate processed key with date-time, layoutId and layoutPrefix
	processedKey := s3utils.GenerateProcessedKey(objectKey, layout.LayoutID, layoutPrefix)
	
	// Upload the image to S3
	err = s3utils.UploadImage(ctx, s3Client, bucket, processedKey, imgBytes)
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to upload image: %v", err))
		return err
	}
	
	log.Info(ctx, layout.LayoutID, fmt.Sprintf("successfully generated and uploaded image to %s", processedKey))
	
	// 7. METADATA STORAGE
	// Initialize DynamoDB manager
	dynamoRegion := os.Getenv("AWS_REGION")
	if dynamoRegion == "" {
		dynamoRegion = "us-east-1" // Default region
	}
	
	// Get DynamoDB table name from environment variable
	tableName := os.Getenv("DYNAMODB_LAYOUT_TABLE")
	if tableName == "" {
		log.Error(ctx, layout.LayoutID, "DYNAMODB_LAYOUT_TABLE environment variable not set")
		return fmt.Errorf("DYNAMODB_LAYOUT_TABLE environment variable not set")
	}
	
	// Initialize DynamoDB manager
	dbManager, err := dynamodb.NewManager(ctx, dynamoRegion, tableName)
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to initialize DynamoDB manager: %v", err))
		return err
	}
	
	// Store layout metadata in DynamoDB
	err = dbManager.StoreLayoutMetadata(ctx, layout, layoutPrefix, bucket, processedKey, objectKey)
	if err != nil {
		// If it's a conditional check failure, the item already exists, which is not a fatal error
		if !strings.Contains(err.Error(), "already exists") {
			log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to store layout metadata: %v", err))
			return err
		}
		log.Info(ctx, layout.LayoutID, fmt.Sprintf("layout metadata already exists: %v", err))
	} else {
		log.Info(ctx, layout.LayoutID, "successfully stored layout metadata in DynamoDB")
	}
	
	return nil
}

func main() {
	lambda.Start(handler)
}