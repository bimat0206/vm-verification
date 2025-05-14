package parser

import (
	"testing"
	
	"workflow-function/shared/logger"
	"workflow-function/product-approach/workflow-function/ProcessTurn1Response/internal/types"
)

// mockLogger implements a simple logger for testing
type mockLogger struct{}

func (m *mockLogger) Info(msg string, data map[string]interface{}) {}
func (m *mockLogger) Debug(msg string, data map[string]interface{}) {}
func (m *mockLogger) Warn(msg string, data map[string]interface{}) {}
func (m *mockLogger) Error(msg string, data map[string]interface{}) {}

// TestParserCreation tests that the parser can be created correctly
func TestParserCreation(t *testing.T) {
	log := &mockLogger{}
	parser := NewParser(log)
	
	if parser == nil {
		t.Fatal("Failed to create parser instance")
	}
}

// TestParseEmptyResponse tests the parser behavior with an empty response
func TestParseEmptyResponse(t *testing.T) {
	log := &mockLogger{}
	parser := NewParser(log)
	
	// Test with empty response
	result := parser.ParseValidationResponse(map[string]interface{}{})
	
	// Should return fallback result
	if result["fallback"] != true {
		t.Error("Expected fallback result for empty response")
	}
}

// TestParseSimpleResponse tests the parser behavior with a simple response
func TestParseSimpleResponse(t *testing.T) {
	log := &mockLogger{}
	parser := NewParser(log)
	
	// Simple response with some content
	response := map[string]interface{}{
		"content": "The machine has 6 rows and 10 columns. The structure is confirmed.",
	}
	
	result := parser.ParseMachineStructure(response)
	
	// Should extract simple structure
	if rowCount, ok := result["rowCount"].(int); !ok || rowCount != 6 {
		t.Errorf("Expected rowCount=6, got %v", result["rowCount"])
	}
	
	if colCount, ok := result["columnsPerRow"].(int); !ok || colCount != 10 {
		t.Errorf("Expected columnsPerRow=10, got %v", result["columnsPerRow"])
	}
}

// TestMachineStateExtraction tests the extraction of machine state data
func TestMachineStateExtraction(t *testing.T) {
	log := &mockLogger{}
	config := types.DefaultProcessingConfig()
	context := &types.ParsingContext{
		VerificationType:    "",
		HasHistoricalContext: false,
		ParsingConfig:       config,
	}
	
	responseParser := NewResponseParser(config, context, log)
	
	// Test with a sample response
	response := map[string]interface{}{
		"content": `
			Row A: Fully stocked with 10 items.
			Row B: Partially filled, some products visible.
			Row C: Empty, coil visible.
			
			Empty positions: A3, B2, B5, C1, C2, C3.
		`,
	}
	
	parsed, err := responseParser.ParseResponse(response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Check if sections were parsed
	if len(parsed.ParsedSections) == 0 {
		t.Error("No sections were parsed")
	}
	
	// Check extracted data
	if len(parsed.ExtractedData) == 0 {
		t.Error("No data was extracted")
	}
	
	// Extract positions
	emptyPos := responseParser.extractEmptyPositionsData(parsed)
	if len(emptyPos) == 0 {
		t.Error("No empty positions extracted")
	}
	
	// Check for specific positions
	positionFound := false
	for _, pos := range emptyPos {
		if pos == "A3" {
			positionFound = true
			break
		}
	}
	
	if !positionFound {
		t.Error("Expected position A3 not found in extracted data")
	}
}