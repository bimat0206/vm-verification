// examples/basic_usage.go
package main

import (
	"fmt"
	"log"

	"github.com/kootoro/s3state"
)

func main() {
	// Initialize the state manager
	manager, err := s3state.New("my-state-bucket")
	if err != nil {
		log.Fatal("Failed to create state manager:", err)
	}

	// Example 1: Store and retrieve JSON data
	data := map[string]interface{}{
		"message": "Hello, World!",
		"count":   42,
	}

	// Store JSON data
	ref, err := manager.StoreJSON(s3state.CategoryProcessing, "test-data.json", data)
	if err != nil {
		log.Fatal("Failed to store JSON:", err)
	}

	fmt.Printf("Stored data at: %s\n", ref.String())

	// Retrieve JSON data
	var retrieved map[string]interface{}
	err = manager.RetrieveJSON(ref, &retrieved)
	if err != nil {
		log.Fatal("Failed to retrieve JSON:", err)
	}

	fmt.Printf("Retrieved data: %+v\n", retrieved)

	// Example 2: Store and retrieve raw bytes
	rawData := []byte("This is raw binary data")

	ref2, err := manager.Store(s3state.CategoryImages, "binary-data.bin", rawData)
	if err != nil {
		log.Fatal("Failed to store raw data:", err)
	}

	retrievedData, err := manager.Retrieve(ref2)
	if err != nil {
		log.Fatal("Failed to retrieve raw data:", err)
	}

	fmt.Printf("Retrieved %d bytes\n", len(retrievedData))
}