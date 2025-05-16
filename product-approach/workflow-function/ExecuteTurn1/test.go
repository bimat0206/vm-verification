package main

import (
	"encoding/json"
	"log"

	"workflow-function/shared/bedrock"
)

// TestBytesImageSource tests the bytes-based image source
func TestBytesImageSource() {
	// Create a test image source with bytes (base64)
	imageSource := bedrock.ImageSource{
		Type:  "bytes",
		Bytes: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", // 1x1 transparent PNG
	}

	// Test image source validation
	err := bedrock.ValidateImageSource(imageSource)
	if err != nil {
		log.Fatalf("Image source validation failed: %v", err)
	}
	log.Printf("✅ Image source validation passed")

	// Create a test image block with bytes source
	imageBlock := &bedrock.ImageBlock{
		Format: "png",
		Source: imageSource,
	}

	// Create a content block with the image
	contentBlock := bedrock.ContentBlock{
		Type:  "image",
		Image: imageBlock,
	}

	// Create a message wrapper with the image content
	messageWrapper := bedrock.MessageWrapper{
		Role:    "user",
		Content: []bedrock.ContentBlock{contentBlock},
	}

	// Test message wrapper validation
	err = bedrock.ValidateMessageWrapper(messageWrapper)
	if err != nil {
		log.Fatalf("Message wrapper validation failed: %v", err)
	}
	log.Printf("✅ Message wrapper validation passed")

	// Create a converse request with the message
	converseRequest := bedrock.CreateConverseRequest(
		"anthropic.claude-3-sonnet-20240229-v1:0",
		[]bedrock.MessageWrapper{messageWrapper},
		"You are an assistant that helps analyze images.",
		24000,
	)

	// Test converse request validation
	err = bedrock.ValidateConverseRequest(converseRequest)
	if err != nil {
		log.Fatalf("Converse request validation failed: %v", err)
	}
	log.Printf("✅ Converse request validation passed")

	// Test serialization to JSON
	requestJSON, err := json.MarshalIndent(converseRequest, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal converse request: %v", err)
	}
	log.Printf("✅ JSON serialization passed")
	log.Printf("JSON output (truncated):\n%s", string(requestJSON[:200])+"...")

	log.Printf("✅ All tests passed")
}

// TestS3LocationImageSource tests the S3-based image source
func TestS3LocationImageSource() {
	// Create a test image source with S3 location (bucketOwner is optional now)
	imageSource := bedrock.ImageSource{
		Type: "s3Location",
		S3Location: bedrock.S3Location{
			URI: "s3://test-bucket/test-image.png",
			// No bucketOwner provided, should still be valid
		},
	}

	// Test image source validation
	err := bedrock.ValidateImageSource(imageSource)
	if err != nil {
		log.Fatalf("Image source validation failed: %v", err)
	}
	log.Printf("✅ S3 image source validation passed (with empty bucketOwner)")

	// Add a bucketOwner and test again
	imageSource.S3Location.BucketOwner = "123456789012"
	err = bedrock.ValidateImageSource(imageSource)
	if err != nil {
		log.Fatalf("Image source validation failed: %v", err)
	}
	log.Printf("✅ S3 image source validation passed (with bucketOwner)")

	log.Printf("✅ All S3 location tests passed")
}

// TestImageContentCreation tests creating image content with both methods
func TestImageContentCreation() {
	// Test creating with bytes
	bytesContent := bedrock.CreateImageContentFromBytes(
		"png",
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=",
	)
	if bytesContent.Type != "image" || bytesContent.Image == nil || 
	   bytesContent.Image.Source.Type != "bytes" || bytesContent.Image.Source.Bytes == "" {
		log.Fatalf("Failed to create bytes image content correctly")
	}
	log.Printf("✅ CreateImageContentFromBytes passed")

	// Test creating with S3 (bucketOwner is optional now)
	s3Content := bedrock.CreateImageContentFromS3(
		"png",
		"s3://test-bucket/test-image.png",
		"", // No bucketOwner
	)
	if s3Content.Type != "image" || s3Content.Image == nil || 
	   s3Content.Image.Source.Type != "s3Location" || s3Content.Image.Source.S3Location.URI == "" {
		log.Fatalf("Failed to create S3 image content correctly")
	}
	log.Printf("✅ CreateImageContentFromS3 passed (without bucketOwner)")

	// Test creating with S3 and bucketOwner
	s3Content = bedrock.CreateImageContentFromS3(
		"png",
		"s3://test-bucket/test-image.png",
		"123456789012",
	)
	if s3Content.Image.Source.S3Location.BucketOwner != "123456789012" {
		log.Fatalf("Failed to set bucketOwner correctly")
	}
	log.Printf("✅ CreateImageContentFromS3 passed (with bucketOwner)")

	log.Printf("✅ All image content creation tests passed")
}

// RunAllTests runs all the validation tests
func RunAllTests() {
	log.Printf("Starting validation tests...")
	TestBytesImageSource()
	TestS3LocationImageSource()
	TestImageContentCreation()
	log.Printf("✅ All tests completed successfully")
}

func main() {
	RunAllTests()
}