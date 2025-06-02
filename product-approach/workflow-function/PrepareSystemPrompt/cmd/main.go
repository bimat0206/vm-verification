package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/PrepareSystemPrompt/internal/config"
	"workflow-function/PrepareSystemPrompt/internal/handlers"
	"workflow-function/PrepareSystemPrompt/internal/models"
)

var handler *handlers.Handler

func init() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize handler
	handler, err = handlers.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	log.Printf("Successfully initialized PrepareSystemPrompt Lambda")
}

// HandleRequest is the Lambda handler function with enhanced input parsing
func HandleRequest(ctx context.Context, event interface{}) (json.RawMessage, error) {
	start := time.Now()
	
	// Log the incoming event type and basic info
	log.Printf("Received event type: %T", event)
	
	// Parse the incoming event with enhanced error handling
	var raw map[string]interface{}
	var parseErr error
	
	switch e := event.(type) {
	case json.RawMessage:
		// Direct JSON message
		if len(e) == 0 {
			log.Printf("Error: Empty JSON input received")
			return nil, fmt.Errorf("failed to parse event: empty JSON input")
		}
		
		// Check for basic JSON structure
		jsonStr := string(e)
		if !strings.HasPrefix(jsonStr, "{") || !strings.HasSuffix(jsonStr, "}") {
			log.Printf("Error: Invalid JSON structure - size: %d, preview: %s", len(e), jsonStr[:min(100, len(jsonStr))])
			return nil, fmt.Errorf("failed to parse event: invalid JSON structure")
		}
		
		parseErr = json.Unmarshal(e, &raw)
		if parseErr != nil {
			log.Printf("Error parsing JSON RawMessage: %v, size: %d, preview: %s", parseErr, len(e), jsonStr[:min(200, len(jsonStr))])
		} else {
			log.Printf("Successfully parsed JSON RawMessage, size: %d", len(e))
		}
		
	case map[string]interface{}:
		// Step Functions / direct map - use directly
		raw = e
		log.Printf("Processing Step Functions request with keys: %v", getMapKeys(e))
		
	case string:
		// String input - try to parse as JSON
		if e == "" {
			log.Printf("Error: Empty string input received")
			return nil, fmt.Errorf("failed to parse event: empty string input")
		}
		
		parseErr = json.Unmarshal([]byte(e), &raw)
		if parseErr != nil {
			log.Printf("Error parsing string input: %v, content: %s", parseErr, e[:min(200, len(e))])
		} else {
			log.Printf("Successfully parsed string input, length: %d", len(e))
		}
		
	default:
		// Fallback: try to marshal and unmarshal
		jsonBytes, err := json.Marshal(e)
		if err != nil {
			log.Printf("Error marshaling unknown event type %T: %v", e, err)
			return nil, fmt.Errorf("unknown event format: %w", err)
		}
		
		parseErr = json.Unmarshal(jsonBytes, &raw)
		if parseErr != nil {
			log.Printf("Error parsing marshaled event: %v, size: %d", parseErr, len(jsonBytes))
		} else {
			log.Printf("Successfully parsed unknown event type %T, size: %d", e, len(jsonBytes))
		}
	}

	// Handle parsing errors
	if parseErr != nil {
		log.Printf("Failed to parse event: %v", parseErr)
		
		// Provide more specific error messages
		if strings.Contains(parseErr.Error(), "unexpected end of JSON input") {
			return nil, fmt.Errorf("failed to parse event detail: JSON input appears to be truncated or incomplete")
		}
		
		return nil, fmt.Errorf("failed to parse event detail: %w", parseErr)
	}

	// Validate we have some data
	if raw == nil {
		log.Printf("Error: Parsed event is nil")
		return nil, fmt.Errorf("failed to parse event: no data received")
	}

	log.Printf("Event parsed successfully with keys: %v", getMapKeys(raw))

	// Convert the parsed map back to JSON for the handler
	eventBytes, err := json.Marshal(raw)
	if err != nil {
		log.Printf("Error re-marshaling parsed event: %v", err)
		return nil, fmt.Errorf("failed to process parsed event: %w", err)
	}

	// Create a models.Input from the parsed data
	var input models.Input
	if err := json.Unmarshal(eventBytes, &input); err != nil {
		log.Printf("Error creating Input model: %v", err)
		log.Printf("Raw data keys: %v", getMapKeys(raw))
		
		// Try to create a basic input structure if direct parsing fails
		input = models.Input{
			Type: models.InputTypeS3Reference, // Default type
		}
		
		// Try to extract common fields manually
		if verificationId, ok := raw["verificationId"].(string); ok {
			log.Printf("Found verificationId: %s", verificationId)
		}
		
		if status, ok := raw["status"].(string); ok {
			log.Printf("Found status: %s", status)
		}
		
		if s3References, ok := raw["s3References"].(map[string]interface{}); ok {
			log.Printf("Found s3References with keys: %v", getMapKeys(s3References))
		}
		
		// If we still can't parse, return the original error with more context
		return nil, fmt.Errorf("invalid input format - failed to parse into models.Input: %w (available keys: %v)", err, getMapKeys(raw))
	}

	log.Printf("Successfully created Input model, type: %s", input.Type)

	// Handle request using the handler
	response, err := handler.HandleRequest(ctx, eventBytes)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		return nil, err
	}

	log.Printf("Successfully processed request in %v", time.Since(start))
	return response, nil
}

// Helper functions

// getMapKeys extracts keys from a map for logging
func getMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	lambda.Start(HandleRequest)
}
