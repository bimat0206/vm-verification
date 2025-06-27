// examples/real_world_usage.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kootoro/s3state"
)

// Example data structures
type ImageMetadata struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Format      string `json:"format"`
	Size        int64  `json:"size"`
	ProcessedAt string `json:"processedAt"`
}

type AnalysisResult struct {
	Success      bool     `json:"success"`
	Discrepancies int      `json:"discrepancies"`
	Confidence   float64  `json:"confidence"`
	Issues       []string `json:"issues"`
}

// FetchImagesFunction simulates the FetchImages Lambda function
func FetchImagesFunction(ctx context.Context, envelope *s3state.Envelope) (*s3state.Envelope, error) {
	manager, err := s3state.New("kootoro-state-bucket")
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Simulate processing images and creating metadata
	metadata := ImageMetadata{
		Width:       1920,
		Height:      1080,
		Format:      "JPEG",
		Size:        2048576,
		ProcessedAt: "2025-04-21T15:30:25Z",
	}

	// Store image metadata
	err = manager.SaveToEnvelope(envelope, s3state.CategoryImages, s3state.ImageMetadataFile, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to save image metadata: %w", err)
	}

	// Simulate storing Base64 image data
	base64Data := map[string]string{
		"reference": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB...", // Truncated for example
		"checking":  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB...", // Truncated for example
	}

	err = manager.SaveToEnvelope(envelope, s3state.CategoryImages, s3state.ReferenceBase64File, base64Data["reference"])
	if err != nil {
		return nil, fmt.Errorf("failed to save reference image: %w", err)
	}

	err = manager.SaveToEnvelope(envelope, s3state.CategoryImages, s3state.CheckingBase64File, base64Data["checking"])
	if err != nil {
		return nil, fmt.Errorf("failed to save checking image: %w", err)
	}

	envelope.SetStatus("IMAGES_FETCHED")
	envelope.AddSummary("imagesProcessed", 2)
	envelope.AddSummary("processingTimeMs", 3500)

	return envelope, nil
}

// ProcessTurn1ResponseFunction simulates the ProcessTurn1Response Lambda function
func ProcessTurn1ResponseFunction(ctx context.Context, envelope *s3state.Envelope) (*s3state.Envelope, error) {
	manager, err := s3state.New("kootoro-state-bucket")
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Load turn 1 response (simulated)
	if ref := envelope.GetReference("responses_turn1"); ref != nil {
		var turn1Response map[string]interface{}
		err = manager.RetrieveJSON(ref, &turn1Response)
		if err != nil {
			return nil, fmt.Errorf("failed to load turn1 response: %w", err)
		}
		fmt.Printf("Processing turn1 response with %d fields\n", len(turn1Response))
	}

	// Create analysis result
	analysis := AnalysisResult{
		Success:       true,
		Discrepancies: 0,
		Confidence:    0.95,
		Issues:        []string{},
	}

	// Store analysis result
	err = manager.SaveToEnvelope(envelope, s3state.CategoryProcessing, s3state.Turn1AnalysisFile, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to save turn1 analysis: %w", err)
	}

	envelope.SetStatus("TURN1_PROCESSED")
	envelope.AddSummary("analysisCompleted", true)

	return envelope, nil
}

// StoreResultsFunction simulates the StoreResults Lambda function
func StoreResultsFunction(ctx context.Context, envelope *s3state.Envelope) (*s3state.Envelope, error) {
	manager, err := s3state.New("kootoro-state-bucket")
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Load final results from processing category
	if ref := envelope.GetReference("processing_final_results"); ref != nil {
		var finalResults map[string]interface{}
		err = manager.RetrieveJSON(ref, &finalResults)
		if err != nil {
			return nil, fmt.Errorf("failed to load final results: %w", err)
		}

		// Simulate storing to DynamoDB and generating visualization
		fmt.Printf("Storing verification results to DynamoDB\n")
		fmt.Printf("Generated result visualization\n")
	}

	envelope.SetStatus("COMPLETED")
	envelope.AddSummary("resultStored", true)
	envelope.AddSummary("resultImageUrl", "s3://results-bucket/verif-123/result.jpg")

	return envelope, nil
}

func main() {
	// Simulate a complete workflow
	fmt.Println("=== Simulating Complete Workflow ===")

	// Start with initial envelope
	envelope := s3state.NewEnvelope("verif-2025042115302500")
	envelope.SetStatus("INITIALIZED")

	ctx := context.Background()

	// Step 1: Fetch Images
	fmt.Println("Step 1: Fetching Images...")
	var err error
	envelope, err = FetchImagesFunction(ctx, envelope)
	if err != nil {
		log.Fatal("FetchImages failed:", err)
	}
	fmt.Printf("Status: %s, References: %d\n", envelope.Status, len(envelope.References))

	// Step 2: Process Turn 1 Response
	fmt.Println("Step 2: Processing Turn 1 Response...")
	envelope, err = ProcessTurn1ResponseFunction(ctx, envelope)
	if err != nil {
		log.Fatal("ProcessTurn1Response failed:", err)
	}
	fmt.Printf("Status: %s, References: %d\n", envelope.Status, len(envelope.References))

	// Step 3: Store Results
	fmt.Println("Step 3: Storing Results...")
	envelope, err = StoreResultsFunction(ctx, envelope)
	if err != nil {
		log.Fatal("StoreResults failed:", err)
	}
	fmt.Printf("Status: %s, References: %d\n", envelope.Status, len(envelope.References))

	// Final envelope summary
	fmt.Println("\n=== Final Envelope Summary ===")
	fmt.Printf("Verification ID: %s\n", envelope.VerificationID)
	fmt.Printf("Status: %s\n", envelope.Status)
	fmt.Printf("Total References: %d\n", len(envelope.References))
	fmt.Printf("Categories: %v\n", envelope.ListCategories())

	// Print all references
	fmt.Println("\n=== All References ===")
	for name, ref := range envelope.References {
		fmt.Printf("  %s: %s\n", name, ref.String())
	}

	// Print summary
	fmt.Println("\n=== Summary ===")
	for key, value := range envelope.Summary {
		fmt.Printf("  %s: %v\n", key, value)
	}
}