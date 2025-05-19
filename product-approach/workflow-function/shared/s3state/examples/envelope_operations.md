// examples/envelope_operations.go
package main

import (
	"fmt"
	"log"

	"github.com/kootoro/s3state"
)

func main() {
	// Create a new envelope
	envelope := s3state.NewEnvelope("verif-2025042115302500")
	
	// Add some references manually
	ref1 := s3state.NewReference("my-bucket", "verif-2025042115302500/images/metadata.json", 1024)
	envelope.AddReference("images_metadata", ref1)
	
	ref2 := s3state.NewReference("my-bucket", "verif-2025042115302500/prompts/system-prompt.json", 2048)
	envelope.AddReference("prompts_system", ref2)

	// Set status and add summary information
	envelope.SetStatus("PROCESSING")
	envelope.AddSummary("stage", "image_processing")
	envelope.AddSummary("startTime", "2025-04-21T15:30:25Z")

	// Check if references exist
	if envelope.HasReference("images_metadata") {
		fmt.Println("Metadata reference exists")
	}

	// Get references by category
	imageRefs := envelope.GetReferencesByCategory("images")
	fmt.Printf("Found %d image references\n", len(imageRefs))

	// List all categories
	categories := envelope.ListCategories()
	fmt.Printf("Categories: %v\n", categories)

	// Validate envelope
	if err := envelope.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
	} else {
		fmt.Println("Envelope is valid")
	}

	// Convert to JSON (for Step Functions)
	jsonData, err := envelope.ToJSON()
	if err != nil {
		log.Fatal("Failed to convert to JSON:", err)
	}

	fmt.Printf("JSON length: %d bytes\n", len(jsonData))

	// Load from JSON
	envelope2, err := s3state.FromJSON(jsonData)
	if err != nil {
		log.Fatal("Failed to load from JSON:", err)
	}

	fmt.Printf("Loaded envelope with %d references\n", len(envelope2.References))
}