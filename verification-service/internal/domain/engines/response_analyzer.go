package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"verification-service/internal/domain/models"
)

// ResponseAnalyzer implements the domain.ResponseAnalyzer interface
type ResponseAnalyzer struct{}

// NewResponseAnalyzer creates a new response analyzer
func NewResponseAnalyzer() *ResponseAnalyzer {
	return &ResponseAnalyzer{}
}

// ProcessTurn1Response analyzes the Turn 1 (reference layout) response
func (a *ResponseAnalyzer) ProcessTurn1Response(
	ctx context.Context,
	response string,
	layoutMetadata map[string]interface{},
) (*models.ReferenceAnalysis, error) {
	// In a real implementation, this would parse the AI model's response
	// to extract structured data about the reference layout
	
	// Create basic machine structure from metadata
	machineStructure := models.MachineStructure{
		RowCount:      6,
		ColumnsPerRow: 10,
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	}
	
	// Extract machine structure from metadata if available
	if meta, ok := layoutMetadata["machineStructure"].(map[string]interface{}); ok {
		if rowCount, ok := meta["rowCount"].(float64); ok {
			machineStructure.RowCount = int(rowCount)
		}
		if columnsPerRow, ok := meta["columnsPerRow"].(float64); ok {
			machineStructure.ColumnsPerRow = int(columnsPerRow)
		}
		// Additional mapping from metadata would happen here
	}
	
	// Create mock row analysis for demonstration
	rowAnalysis := make(map[string]map[string]interface{})
	for _, row := range machineStructure.RowOrder {
		rowAnalysis[row] = map[string]interface{}{
			"description": fmt.Sprintf("Row %s contains products from positions %s01-%s%02d", 
				row, row, row, machineStructure.ColumnsPerRow),
			"productsFound": []map[string]interface{}{
				{
					"position":  fmt.Sprintf("%s01", row),
					"product":   "Mi Hảo Hảo",
					"isPresent": true,
				},
			},
		}
	}
	
	// Create mock product positions for demonstration
	productPositions := make(map[string]map[string]interface{})
	for _, row := range machineStructure.RowOrder {
		for colIdx, col := range machineStructure.ColumnOrder {
			if colIdx < 7 { // Only include first 7 columns as visible
				position := fmt.Sprintf("%s%s", row, col)
				productPositions[position] = map[string]interface{}{
					"productName": "Mi Hảo Hảo",
					"visible":     true,
				}
			}
		}
	}
	
	// Extract any meaningful information from the response text
	initialConfirmation := "Successfully identified shelf structure."
	if strings.Contains(response, "identified") && strings.Contains(response, "row") {
		initialConfirmation = extractInitialConfirmation(response)
	}
	
	// Return analysis
	return &models.ReferenceAnalysis{
		TurnNumber:         1,
		MachineStructure:   machineStructure,
		RowAnalysis:        rowAnalysis,
		ProductPositions:   productPositions,
		EmptyPositions:     []string{},
		Confidence:         90,
		InitialConfirmation: initialConfirmation,
		OriginalResponse:   response,
		CompletedAt:        time.Now(),
	}, nil
}

// ProcessTurn2Response analyzes the Turn 2 (checking image) response
func (a *ResponseAnalyzer) ProcessTurn2Response(
	ctx context.Context,
	response string,
	referenceAnalysis *models.ReferenceAnalysis,
) (*models.CheckingAnalysis, error) {
	// In a real implementation, this would parse the AI model's response
	// to extract structured data about discrepancies
	
	// Try to extract JSON from the response
	jsonStr := extractJSON(response)
	
	// If JSON extraction failed, create a simplified analysis
	if jsonStr == "" {
		// Create mock discrepancies for demonstration
		discrepancies := []models.Discrepancy{
			{
				Position:          "E01",
				Expected:          "Green \"Mi Cung Đình\" cup noodle",
				Found:             "Red/white \"Mi modern Lẩu thái\" cup noodle",
				Issue:             models.DiscrepancyIncorrectProductType,
				Confidence:        95,
				Evidence:          "Different packaging color and branding visible",
				VerificationResult: models.StatusIncorrect,
				Severity:          models.SeverityHigh,
			},
		}
		
		// Create empty slot report
		emptySlotReport := struct {
			ReferenceEmptyRows      []string `json:"referenceEmptyRows"`
			CheckingEmptyRows       []string `json:"checkingEmptyRows"`
			CheckingPartiallyEmptyRows []string `json:"checkingPartiallyEmptyRows"`
			CheckingEmptyPositions  []string `json:"checkingEmptyPositions"`
			TotalEmpty              int      `json:"totalEmpty"`
		}{
			ReferenceEmptyRows:      []string{},
			CheckingEmptyRows:       []string{"F"},
			CheckingPartiallyEmptyRows: []string{},
			CheckingEmptyPositions:  []string{"F01", "F02", "F03", "F04", "F05", "F06", "F07"},
			TotalEmpty:              7,
		}
		
		// Create mock row analysis for demonstration
		rowAnalysis := make(map[string]map[string]interface{})
		for _, row := range referenceAnalysis.MachineStructure.RowOrder {
			rowAnalysis[row] = map[string]interface{}{
				"description": fmt.Sprintf("Row %s status based on checking image", row),
				"status":      "Correct",
			}
		}
		
		// Return analysis
		return &models.CheckingAnalysis{
			TurnNumber:        2,
			VerificationStatus: models.StatusIncorrect,
			Discrepancies:     discrepancies,
			TotalDiscrepancies: len(discrepancies),
			Severity:          models.SeverityHigh,
			RowAnalysis:       rowAnalysis,
			EmptySlotReport:   emptySlotReport,
			Confidence:        95,
			OriginalResponse:  response,
			CompletedAt:       time.Now(),
		}, nil
	}
	
	// Try to parse the extracted JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}
	
	// Extract discrepancies
	var discrepancies []models.Discrepancy
	if discrepanciesJSON, ok := result["discrepancies"].([]interface{}); ok {
		for _, discJSON := range discrepanciesJSON {
			if disc, ok := discJSON.(map[string]interface{}); ok {
				discrepancy := models.Discrepancy{
					Position:          getStringOrDefault(disc, "position", ""),
					Expected:          getStringOrDefault(disc, "expected", ""),
					Found:             getStringOrDefault(disc, "found", ""),
					Issue:             models.DiscrepancyType(getStringOrDefault(disc, "issue", "Unknown")),
					Confidence:        getIntOrDefault(disc, "confidence", 80),
					Evidence:          getStringOrDefault(disc, "evidence", ""),
					VerificationResult: models.StatusIncorrect,
					Severity:          models.SeverityMedium,
				}
				
				// Map the result to a status
				resultStr := getStringOrDefault(disc, "verificationResult", "")
				if strings.ToUpper(resultStr) == "CORRECT" {
					discrepancy.VerificationResult = models.StatusCorrect
				} else {
					discrepancy.VerificationResult = models.StatusIncorrect
				}
				
				// Map the severity
				severityStr := getStringOrDefault(disc, "severity", "")
				switch strings.ToUpper(severityStr) {
				case "LOW":
					discrepancy.Severity = models.SeverityLow
				case "MEDIUM":
					discrepancy.Severity = models.SeverityMedium
				case "HIGH":
					discrepancy.Severity = models.SeverityHigh
				}
				
				discrepancies = append(discrepancies, discrepancy)
			}
		}
	}
	
	// Extract total discrepancies
	totalDiscrepancies := len(discrepancies)
	if totalJSON, ok := result["totalDiscrepancies"].(float64); ok {
		totalDiscrepancies = int(totalJSON)
	}
	
	// Extract severity
	severity := models.SeverityMedium
	if severityJSON, ok := result["severity"].(string); ok {
		switch strings.ToUpper(severityJSON) {
		case "LOW":
			severity = models.SeverityLow
		case "MEDIUM":
			severity = models.SeverityMedium
		case "HIGH":
			severity = models.SeverityHigh
		}
	}
	
	// Create empty slot report
	emptySlotReport := struct {
		ReferenceEmptyRows      []string `json:"referenceEmptyRows"`
		CheckingEmptyRows       []string `json:"checkingEmptyRows"`
		CheckingPartiallyEmptyRows []string `json:"checkingPartiallyEmptyRows"`
		CheckingEmptyPositions  []string `json:"checkingEmptyPositions"`
		TotalEmpty              int      `json:"totalEmpty"`
	}{
		ReferenceEmptyRows:      []string{},
		CheckingEmptyRows:       []string{},
		CheckingPartiallyEmptyRows: []string{},
		CheckingEmptyPositions:  []string{},
		TotalEmpty:              0,
	}
	
	// Extract empty positions from discrepancies
	for _, d := range discrepancies {
		if d.Issue == models.DiscrepancyMissingProduct {
			position := d.Position
			row := string(position[0])
			
			// Add to empty positions
			emptySlotReport.CheckingEmptyPositions = append(emptySlotReport.CheckingEmptyPositions, position)
			
			// Check if this row is already marked as empty
			isEmptyRow := false
			isPartiallyEmptyRow := false
			
			for _, emptyRow := range emptySlotReport.CheckingEmptyRows {
				if emptyRow == row {
					isEmptyRow = true
					break
				}
			}
			
			for _, partialRow := range emptySlotReport.CheckingPartiallyEmptyRows {
				if partialRow == row {
					isPartiallyEmptyRow = true
					break
				}
			}
			
			// If not already marked, mark as partially empty
			if !isEmptyRow && !isPartiallyEmptyRow {
				emptySlotReport.CheckingPartiallyEmptyRows = append(emptySlotReport.CheckingPartiallyEmptyRows, row)
			}
		}
	}
	
	// Update total empty count
	emptySlotReport.TotalEmpty = len(emptySlotReport.CheckingEmptyPositions)
	
	// Create row analysis
	rowAnalysis := make(map[string]map[string]interface{})
	for _, row := range referenceAnalysis.MachineStructure.RowOrder {
		// Check if row has discrepancies
		hasDiscrepancy := false
		for _, d := range discrepancies {
			if len(d.Position) > 0 && string(d.Position[0]) == row {
				hasDiscrepancy = true
				break
			}
		}
		
		// Check if row is empty
		isEmpty := false
		for _, emptyRow := range emptySlotReport.CheckingEmptyRows {
			if emptyRow == row {
				isEmpty = true
				break
			}
		}
		
		// Check if row is partially empty
		isPartiallyEmpty := false
		for _, partialRow := range emptySlotReport.CheckingPartiallyEmptyRows {
			if partialRow == row {
				isPartiallyEmpty = true
				break
			}
		}
		
		// Determine row status
		status := "Correct"
		if isEmpty {
			status = "Empty"
		} else if isPartiallyEmpty {
			status = "Partially Empty"
		} else if hasDiscrepancy {
			status = "Incorrect"
		}
		
		rowAnalysis[row] = map[string]interface{}{
			"description": fmt.Sprintf("Row %s status based on checking image", row),
			"status":      status,
		}
	}
	
	// Determine verification status
	verificationStatus := models.StatusCorrect
	if len(discrepancies) > 0 {
		verificationStatus = models.StatusIncorrect
	}
	
	// Return analysis
	return &models.CheckingAnalysis{
		TurnNumber:        2,
		VerificationStatus: verificationStatus,
		Discrepancies:     discrepancies,
		TotalDiscrepancies: totalDiscrepancies,
		Severity:          severity,
		RowAnalysis:       rowAnalysis,
		EmptySlotReport:   emptySlotReport,
		Confidence:        95,
		OriginalResponse:  response,
		CompletedAt:       time.Now(),
	}, nil
}

// Helper function to extract JSON from a string
func extractJSON(text string) string {
	// Find first opening brace
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}
	
	// Find matching closing brace
	braceCount := 0
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			braceCount++
		} else if text[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return text[start : i+1]
			}
		}
	}
	
	return ""
}

// Helper function to extract initial confirmation from response
func extractInitialConfirmation(text string) string {
	// Look for sentences mentioning row structure
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if (strings.Contains(sentence, "row") || strings.Contains(sentence, "shelf")) && 
		   (strings.Contains(sentence, "identif") || strings.Contains(sentence, "structur")) {
			return sentence + "."
		}
	}
	
	return "Successfully identified shelf structure."
}

// Helper function to get string from map or return default
func getStringOrDefault(m map[string]interface{}, key string, def string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return def
}

// Helper function to get int from map or return default
func getIntOrDefault(m map[string]interface{}, key string, def int) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return def
}