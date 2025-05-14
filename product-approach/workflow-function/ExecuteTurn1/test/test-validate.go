package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"execute-turn1/internal"
)

func main() {
	// Check if a filename was provided as an argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-validate.go <input-file.json>")
		os.Exit(1)
	}

	// Read the input file
	filename := os.Args[1]
	fmt.Printf("Testing input file: %s\n", filename)
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse the input
	var input internal.ExecuteTurn1Input
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Printf("Failed to unmarshal input: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input parsed successfully\n")
	fmt.Printf("Verification ID: %s\n", input.VerificationContext.VerificationID)
	fmt.Printf("Verification Type: %s\n", input.VerificationContext.VerificationType)

	// Print details of the parsed input structure
	fmt.Println("\nStructure details:")

	// Check CurrentPrompt
	fmt.Printf("CurrentPrompt Messages Length: %d\n", len(input.CurrentPrompt.Messages))
	if len(input.CurrentPrompt.Messages) > 0 {
		fmt.Printf("CurrentPrompt First Message Role: %s\n", input.CurrentPrompt.Messages[0].Role)
	}
	fmt.Printf("CurrentPrompt Turn Number: %d\n", input.CurrentPrompt.TurnNumber)
	fmt.Printf("CurrentPrompt Image Included: %s\n", input.CurrentPrompt.ImageIncluded)

	// Check BedrockConfig
	fmt.Printf("BedrockConfig Anthropic Version: %s\n", input.BedrockConfig.AnthropicVersion)
	fmt.Printf("BedrockConfig Max Tokens: %d\n", input.BedrockConfig.MaxTokens)
	fmt.Printf("BedrockConfig Thinking Type: %s\n", input.BedrockConfig.Thinking.Type)
	fmt.Printf("BedrockConfig Budget Tokens: %d\n", input.BedrockConfig.Thinking.BudgetTokens)

	// Validate the input
	fmt.Println("\nValidating input...")
	if err := internal.ValidateExecuteTurn1Input(&input); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Input validation successful!")
}