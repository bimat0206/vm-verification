package parser

import (
	"strings"
	
	"workflow-function/product-approach/workflow-function/ProcessTurn1Response/internal/types"
)

// ParseValidationResponse parses validation indicators from response
func (p *ParserImpl) ParseValidationResponse(response map[string]interface{}) map[string]interface{} {
	parsed, err := p.responseParser.ParseResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse validation response", map[string]interface{}{
			"error": err.Error(),
		})
		return p.createFallbackValidationResult()
	}
	
	result := make(map[string]interface{})
	
	// Extract structure confirmation
	if structure, exists := parsed.ExtractedData["machineStructure"]; exists {
		if ms, ok := structure.(*types.MachineStructure); ok {
			result["structureConfirmed"] = ms.StructureConfirmed
			result["rowCount"] = ms.RowCount
			result["columnsPerRow"] = ms.ColumnsPerRow
		}
	}
	
	// Extract validation indicators from content
	content := strings.ToLower(parsed.MainContent)
	result["layoutMatches"] = strings.Contains(content, "confirmed") || strings.Contains(content, "verified")
	result["productTypesConfirmed"] = p.countProductTypes(parsed.MainContent)
	result["fullyStocked"] = strings.Contains(content, "full") && !strings.Contains(content, "empty")
	
	return result
}

// ParseVisualAnalysis parses visual analysis from response
func (p *ParserImpl) ParseVisualAnalysis(response map[string]interface{}) map[string]interface{} {
	parsed, err := p.responseParser.ParseResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse visual analysis", map[string]interface{}{
			"error": err.Error(),
		})
		return make(map[string]interface{})
	}
	
	result := make(map[string]interface{})
	
	// Extract visual observations
	if observations, exists := parsed.ParsedSections["observations"]; exists {
		result["visualObservations"] = observations
	}
	
	// Extract confirmations
	content := strings.ToLower(parsed.MainContent)
	result["matchesHistorical"] = strings.Contains(content, "matches") || strings.Contains(content, "confirms")
	result["newObservations"] = p.extractNewObservations(parsed.MainContent)
	
	return result
}

// ParseMachineStructure parses machine structure from response
func (p *ParserImpl) ParseMachineStructure(response map[string]interface{}) map[string]interface{} {
	parsed, err := p.responseParser.ParseResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse machine structure", map[string]interface{}{
			"error": err.Error(),
		})
		return p.createFallbackStructure()
	}
	
	result := make(map[string]interface{})
	
	// Extract machine structure
	if structure, exists := parsed.ExtractedData["machineStructure"]; exists {
		if ms, ok := structure.(*types.MachineStructure); ok {
			result["rowCount"] = ms.RowCount
			result["columnsPerRow"] = ms.ColumnsPerRow
			result["rowOrder"] = ms.RowOrder
			result["columnOrder"] = ms.ColumnOrder
			result["totalPositions"] = ms.TotalPositions
			result["structureConfirmed"] = ms.StructureConfirmed
		}
	}
	
	return result
}

// ParseMachineState parses machine state from response
func (p *ParserImpl) ParseMachineState(response map[string]interface{}) map[string]interface{} {
	parsed, err := p.responseParser.ParseResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse machine state", map[string]interface{}{
			"error": err.Error(),
		})
		return make(map[string]interface{})
	}
	
	result := make(map[string]interface{})
	
	// Extract row states
	if rowStates, exists := parsed.ExtractedData["rowStates"]; exists {
		result["rowStates"] = rowStates
	}
	
	// Extract empty positions
	if emptyPos, exists := parsed.ExtractedData["emptyPositions"]; exists {
		result["emptyPositions"] = emptyPos
	}
	
	// Calculate summary
	if rowStates, ok := result["rowStates"].(map[string]*types.RowState); ok {
		summary := p.calculateStateSummary(rowStates)
		result["summary"] = summary
	}
	
	return result
}

// ExtractObservations extracts observations from response
func (p *ParserImpl) ExtractObservations(response map[string]interface{}) map[string]interface{} {
	parsed, err := p.responseParser.ParseResponse(response)
	if err != nil {
		p.logger.Warn("Failed to extract observations", map[string]interface{}{
			"error": err.Error(),
		})
		return make(map[string]interface{})
	}
	
	result := make(map[string]interface{})
	
	// Extract observations from sections
	if observations, exists := parsed.ParsedSections["observations"]; exists {
		result["observations"] = observations
	}
	
	// Extract from main content
	result["mainObservations"] = p.extractGeneralObservations(parsed.MainContent)
	
	return result
}

// Helper methods

// createFallbackValidationResult creates a fallback validation result
func (p *ParserImpl) createFallbackValidationResult() map[string]interface{} {
	return map[string]interface{}{
		"structureConfirmed":   false,
		"layoutMatches":        false,
		"productTypesConfirmed": 0,
		"fullyStocked":         false,
		"fallback":             true,
	}
}

// createFallbackStructure creates a fallback structure
func (p *ParserImpl) createFallbackStructure() map[string]interface{} {
	return map[string]interface{}{
		"rowCount":           6,    // Default assumption
		"columnsPerRow":      10,   // Default assumption
		"structureConfirmed": false,
		"fallback":           true,
	}
}

// countProductTypes counts product types mentioned in content
func (p *ParserImpl) countProductTypes(content string) int {
	// Simple heuristic - count unique product patterns
	productIndicators := []string{"cup noodles", "drink", "snack", "water", "soda"}
	count := 0
	content = strings.ToLower(content)
	
	for _, indicator := range productIndicators {
		if strings.Contains(content, indicator) {
			count++
		}
	}
	
	return count
}

// extractNewObservations extracts new observations from content
func (p *ParserImpl) extractNewObservations(content string) []string {
	observations := []string{}
	
	// Look for observation keywords
	keywords := []string{"note", "observe", "visible", "appears", "seems"}
	sentences := strings.Split(content, ".")
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		lowerSentence := strings.ToLower(sentence)
		
		for _, keyword := range keywords {
			if strings.Contains(lowerSentence, keyword) && len(sentence) > 10 {
				observations = append(observations, sentence)
				break
			}
		}
	}
	
	return observations
}

// calculateStateSummary calculates summary from row states
func (p *ParserImpl) calculateStateSummary(rowStates map[string]*types.RowState) map[string]interface{} {
	summary := map[string]interface{}{
		"totalRows":    len(rowStates),
		"fullRows":     0,
		"partialRows":  0,
		"emptyRows":    0,
		"totalProducts": 0,
	}
	
	for _, state := range rowStates {
		switch state.Status {
		case "Full":
			summary["fullRows"] = summary["fullRows"].(int) + 1
		case "Partial":
			summary["partialRows"] = summary["partialRows"].(int) + 1
		case "Empty":
			summary["emptyRows"] = summary["emptyRows"].(int) + 1
		}
		summary["totalProducts"] = summary["totalProducts"].(int) + state.Quantity
	}
	
	return summary
}

// extractGeneralObservations extracts general observations
func (p *ParserImpl) extractGeneralObservations(content string) string {
	// Extract sentences that contain observation keywords
	keywords := []string{"visible", "appears", "shows", "contains", "note"}
	sentences := strings.Split(content, ".")
	var observations []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 10 {
			continue
		}
		
		lowerSentence := strings.ToLower(sentence)
		for _, keyword := range keywords {
			if strings.Contains(lowerSentence, keyword) {
				observations = append(observations, sentence)
				break
			}
		}
	}
	
	return strings.Join(observations, ". ")
}