package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// ImageBlock represents an image in a content block
type ImageBlock struct {
	Format string      `json:"format"`
	Source ImageSource `json:"source"`
}

// ImageSource represents the source of an image
type ImageSource struct {
	Type       string     `json:"type"`
	S3Location S3Location `json:"s3Location"`
}

// S3Location represents an S3 location
type S3Location struct {
	URI         string `json:"uri"`
	BucketOwner string `json:"bucketOwner,omitempty"`
}

func main() {
	// Test with a properly structured image content item
	testStructure := map[string]interface{}{
		"type": "image",
		"image": map[string]interface{}{
			"format": "png",
			"source": map[string]interface{}{
				"type": "s3",
				"s3Location": map[string]interface{}{
					"uri":         "s3://amzn-s3-demo-bucket/myImage",
					"bucketOwner": "111122223333",
				},
			},
		},
	}

	// Marshal the test structure to JSON
	jsonBytes, err := json.MarshalIndent(testStructure, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal test structure: %v", err)
	}

	fmt.Println("Properly structured image content:")
	fmt.Println(string(jsonBytes))

	// Create the content block as it would appear in the final request
	imageContent := map[string]interface{}{
		"role": "user",
		"content": []map[string]interface{}{
			{
				"type": "image",
				"image": map[string]interface{}{
					"format": "png",
					"source": map[string]interface{}{
						"type": "s3",
						"s3Location": map[string]interface{}{
							"uri":         "s3://amzn-s3-demo-bucket/myImage",
							"bucketOwner": "111122223333",
						},
					},
				},
			},
		},
	}

	// Marshal the full message structure to JSON
	imageJsonBytes, err := json.MarshalIndent(imageContent, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal image content: %v", err)
	}

	fmt.Println("\nFull message structure with image:")
	fmt.Println(string(imageJsonBytes))

	fmt.Println("\nThis structure should satisfy the Bedrock API requirements and fix the validation error.")
}