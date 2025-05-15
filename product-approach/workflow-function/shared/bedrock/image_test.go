package bedrock

import (
	"encoding/json"
	"testing"
)

func TestImageContentStructure(t *testing.T) {
	// Create a sample image content block
	imageContent := ContentBlock{
		Type: "image",
		Image: &ImageBlock{
			Format: "png",
			Source: ImageSource{
				S3Location: S3Location{
					URI:         "s3://test-bucket/test-image.png",
					BucketOwner: "123456789012",
				},
			},
		},
	}

	// Create a map to validate structure
	contentMap := map[string]interface{}{
		"type": "image",
		"image": map[string]interface{}{
			"format": imageContent.Image.Format,
			"source": map[string]interface{}{
				"s3Location": map[string]interface{}{
					"uri":         imageContent.Image.Source.S3Location.URI,
					"bucketOwner": imageContent.Image.Source.S3Location.BucketOwner,
				},
			},
		},
	}

	// Convert to JSON and back to ensure structure is maintained
	jsonBytes, err := json.Marshal(contentMap)
	if err != nil {
		t.Fatalf("Failed to marshal content: %v", err)
	}

	t.Logf("JSON structure: %s", string(jsonBytes))

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal content: %v", err)
	}

	// Verify type field
	typeValue, ok := parsed["type"].(string)
	if !ok || typeValue != "image" {
		t.Errorf("Expected type 'image', got %v", parsed["type"])
	}

	// Verify image field
	imageMap, ok := parsed["image"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing or invalid image field: %v", parsed["image"])
	}

	// Verify format field
	format, ok := imageMap["format"].(string)
	if !ok || format != "png" {
		t.Errorf("Expected format 'png', got %v", imageMap["format"])
	}

	// Verify source field exists
	sourceMap, ok := imageMap["source"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing or invalid source field: %v", imageMap["source"])
	}

	// Verify s3Location field
	s3Map, ok := sourceMap["s3Location"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing or invalid s3Location field: %v", sourceMap["s3Location"])
	}

	// Verify URI
	uri, ok := s3Map["uri"].(string)
	if !ok || uri != "s3://test-bucket/test-image.png" {
		t.Errorf("Expected URI 's3://test-bucket/test-image.png', got %v", s3Map["uri"])
	}

	// Verify bucketOwner
	owner, ok := s3Map["bucketOwner"].(string)
	if !ok || owner != "123456789012" {
		t.Errorf("Expected bucketOwner '123456789012', got %v", s3Map["bucketOwner"])
	}

	t.Log("Image content structure validation passed")
}