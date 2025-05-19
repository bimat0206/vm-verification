// examples/category_operations.go
package main

import (
	"fmt"
	"log"

	"github.com/kootoro/s3state"
)

func main() {
	// Work with categories
	fmt.Println("All categories:", s3state.AllCategories())

	// Validate category
	if s3state.IsValidCategory("images") {
		fmt.Println("'images' is a valid category")
	}

	// Generate keys
	key := s3state.GenerateKey("verif-123", s3state.CategoryImages, "metadata.json")
	fmt.Printf("Generated key: %s\n", key)

	// Parse keys
	verificationID, category, filename, err := s3state.ParseKey(key)
	if err != nil {
		log.Fatal("Failed to parse key:", err)
	}
	fmt.Printf("Parsed - ID: %s, Category: %s, Filename: %s\n", verificationID, category, filename)

	// Get category info
	info, err := s3state.GetCategoryInfo(s3state.CategoryImages)
	if err != nil {
		log.Fatal("Failed to get category info:", err)
	}
	fmt.Printf("Category info: %+v\n", info)

	// Use standard filenames
	filename, err = s3state.GetStandardFilename(s3state.CategoryImages, "metadata")
	if err != nil {
		log.Fatal("Failed to get standard filename:", err)
	}
	fmt.Printf("Standard filename: %s\n", filename)

	// Build reference key for envelope
	refKey := s3state.BuildReferenceKey(s3state.CategoryImages, "metadata.json")
	fmt.Printf("Reference key: %s\n", refKey)

	// Validate key structure
	testKey := "verif-123/images/metadata.json"
	if err := s3state.ValidateKeyStructure(testKey); err != nil {
		log.Printf("Invalid key structure: %v", err)
	} else {
		fmt.Println("Key structure is valid")
	}
}