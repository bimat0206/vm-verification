package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"./config"
	"./logger"
	"./renderer"
	"./s3utils"
)

func handler(ctx context.Context, event events.EventBridgeEvent) error {
	// Initialize S3 client and logger
	cfg := config.GetConfig()
	s3Client, err := s3utils.NewS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize S3 client: %v", err)
	}
	log := logger.NewLogger(s3Client, cfg.BucketName)

	// Parse event to get bucket and object key
	bucket, objectKey, err := s3utils.ParseEvent(event)
	if err != nil {
		log.Error(ctx, 0, fmt.Sprintf("failed to parse event: %v", err))
		return err
	}

	// Skip if not a JSON file in the "raw" folder
	if !s3utils.IsValidRawFile(objectKey) {
		log.Info(ctx, 0, "skipping non-JSON or non-raw file: "+objectKey)
		return nil
	}

	// Download and parse the layout JSON
	layout, err := s3utils.DownloadAndParseLayout(ctx, s3Client, cfg.BucketName, objectKey)
	if err != nil {
		log.Error(ctx, 0, fmt.Sprintf("failed to download or parse layout: %v", err))
		return err
	}

	// Log start of processing
	log.Info(ctx, layout.LayoutID, "starting layout processing")

	// Render the layout to an image
	imgBytes, err := renderer.RenderLayoutToBytes(layout)
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to render layout: %v", err))
		return err
	}

	// Generate processed key with date-time and layoutId
	processedKey := s3utils.GenerateProcessedKey(objectKey, layout.LayoutID)

	// Upload the image to S3
	err = s3utils.UploadImage(ctx, s3Client, cfg.BucketName, processedKey, imgBytes)
	if err != nil {
		log.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to upload image: %v", err))
		return err
	}

	// Log success
	log.Info(ctx, layout.LayoutID, fmt.Sprintf("successfully generated and uploaded image to %s", processedKey))
	return nil
}

func main() {
	lambda.Start(handler)
}