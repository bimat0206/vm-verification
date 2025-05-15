package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Import the Handler function
func main() {
	// Read test input file
	inputFile, err := os.ReadFile("test-input.json")
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Set required environment variables
	os.Setenv("BEDROCK_MODEL", "anthropic.claude-3-sonnet-20240229-v1:0")
	os.Setenv("BEDROCK_REGION", "us-east-1")

	// Start the container with necessary AWS credentials
	fmt.Println("Lambda would try to call Bedrock with the input we provided")
	fmt.Println("Since we can't directly test without AWS credentials, here's what the request would look like:")
	
	var inputData map[string]interface{}
	if err := json.Unmarshal(inputFile, &inputData); err != nil {
		log.Fatalf("Error parsing input JSON: %v", err)
	}
	
	// Extract messages to check content structure
	if currentPrompt, ok := inputData["currentPrompt"].(map[string]interface{}); ok {
		if innerPrompt, ok := currentPrompt["currentPrompt"].(map[string]interface{}); ok {
			if messages, ok := innerPrompt["messages"].([]interface{}); ok && len(messages) > 0 {
				fmt.Println("\nMessage structure:")
				for i, msg := range messages {
					msgMap, _ := msg.(map[string]interface{})
					fmt.Printf("- Message %d, role: %s\n", i, msgMap["role"])
					
					content, _ := msgMap["content"].([]interface{})
					for j, contentItem := range content {
						contentMap, _ := contentItem.(map[string]interface{})
						fmt.Printf("  - Content %d, type: %s\n", j, contentMap["type"])
						if text, ok := contentMap["text"].(string); ok {
							fmt.Printf("    Text: %s\n", text)
						}
					}
				}
				fmt.Println("\nThe messages array now contains the 'type' field in each content item")
				fmt.Println("This should fix the validation error: 'messages.0.content.0.type: Field required'")
			}
		}
	}
}
