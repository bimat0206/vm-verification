package parser

import (
	"regexp"
	"strconv"
	"strings"
	
	"workflow-function/ProcessTurn1Response/internal/types"
)

// extractObservations extracts general observations using keywords from configuration
func (p *ResponseParser) extractObservations(content string) string {
	if len(p.patterns.Observations) == 0 {
		return ""
	}
	
	sentences := strings.Split(content, ".")
	var observations []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 10 {
			continue
		}
		
		lowerSentence := strings.ToLower(sentence)
		for _, keyword := range p.patterns.Observations {
			if strings.Contains(lowerSentence, strings.ToLower(keyword)) {
				observations = append(observations, sentence)
				break
			}
		}
	}
	
	if len(observations) > 0 {
		return strings.Join(observations, ". ")
	}
	
	return ""
}

// extractMachineStructureData extracts machine structure data
func (p *ResponseParser) extractMachineStructureData(parsed *types.ParsedResponse) *types.MachineStructure {
	// Check for explicit machine structure section
	if structureSection, exists := parsed.ParsedSections["machine_structure"]; exists {
		return p.parseMachineStructureFromSection(structureSection)
	}
	
	// Try to extract from content using patterns
	content := parsed.MainContent
	
	for _, pattern := range p.patterns.MachineStructure {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(content)
		if len(matches) >= 3 {
			rowCount, rowErr := strconv.Atoi(matches[1])
			colCount, colErr := strconv.Atoi(matches[2])
			
			if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
				structure := &types.MachineStructure{
					RowCount:           rowCount,
					ColumnsPerRow:      colCount,
					StructureConfirmed: true,
				}
				
				// Extract row and column orders
				structure.RowOrder = p.extractRowOrder(content, rowCount)
				structure.ColumnOrder = p.extractColumnOrder(content, colCount)
				structure.CalculateTotalPositions()
				
				return structure
			}
		}
	}
	
	return nil
}

// extractRowStatesData extracts row state information
func (p *ResponseParser) extractRowStatesData(parsed *types.ParsedResponse) map[string]*types.RowState {
	states := make(map[string]*types.RowState)
	
	// Check for explicit row status sections
	sectionNames := []string{"row_status_reference", "row_status_checking", "row_status"}
	for _, sectionName := range sectionNames {
		if section, exists := parsed.ParsedSections[sectionName]; exists {
			rowStates := p.parseRowStatesFromSection(section)
			for k, v := range rowStates {
				states[k] = v
			}
		}
	}
	
	return states
}

// parseRowStatesFromSection parses row states from a section
func (p *ResponseParser) parseRowStatesFromSection(section string) map[string]*types.RowState {
	states := make(map[string]*types.RowState)
	
	// Use configured patterns for row status
	for _, pattern := range p.patterns.RowStatus {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(section, -1)
		
		for _, match := range matches {
			if len(match) >= 3 {
				row := match[1]
				statusText := strings.TrimSpace(match[2])
				
				state := &types.RowState{
					Status: p.parseRowStatus(statusText),
					Notes:  statusText,
				}
				
				// Extract additional details if available
				if len(match) >= 4 && match[3] != "" {
					state.Notes += " - " + strings.TrimSpace(match[3])
				}
				
				// Extract quantities using configured patterns
				state.Quantity = p.extractQuantityFromText(statusText)
				
				states[row] = state
			}
		}
	}
	
	return states
}

// parseRowStatus determines row status from text
func (p *ResponseParser) parseRowStatus(text string) string {
	text = strings.ToLower(text)
	
	// Check for explicit status keywords
	statusKeywords := map[string]string{
		"full":    "Full",
		"partial": "Partial",
		"empty":   "Empty",
		"vacant":  "Empty",
	}
	
	for keyword, status := range statusKeywords {
		if strings.Contains(text, keyword) {
			return status
		}
	}
	
	// Heuristic analysis
	if strings.Contains(text, "visible") && !strings.Contains(text, "coil") {
		return "Partial"
	} else if strings.Contains(text, "coil") {
		return "Empty"
	}
	
	return "Unknown"
}

// extractQuantityFromText extracts quantity using configured patterns
func (p *ResponseParser) extractQuantityFromText(text string) int {
	for _, pattern := range p.patterns.QuantityPatterns {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(strings.ToLower(text))
		if len(matches) >= 2 {
			if quantity, err := strconv.Atoi(matches[1]); err == nil {
				return quantity
			}
		}
	}
	
	return 0
}

// extractEmptyPositionsData extracts empty position information
func (p *ResponseParser) extractEmptyPositionsData(parsed *types.ParsedResponse) []string {
	var emptyPositions []string
	
	// Check for explicit empty positions section
	if section, exists := parsed.ParsedSections["empty_positions"]; exists {
		positions := p.parsePositionsList(section)
		emptyPositions = append(emptyPositions, positions...)
	}
	
	// Try to extract from other sections using patterns
	for _, pattern := range p.patterns.EmptyPositions {
		for _, section := range parsed.ParsedSections {
			matches := regexp.MustCompile(pattern).FindAllString(section, -1)
			for _, match := range matches {
				positions := p.parsePositionsList(match)
				emptyPositions = append(emptyPositions, positions...)
			}
		}
	}
	
	// Remove duplicates
	return p.removeDuplicatePositions(emptyPositions)
}

// parsePositionsList parses a list of positions using configured pattern
func (p *ResponseParser) parsePositionsList(text string) []string {
	matches := regexp.MustCompile(p.patterns.PositionPattern).FindAllString(text, -1)
	
	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var unique []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			unique = append(unique, match)
		}
	}
	
	return unique
}

// removeDuplicatePositions removes duplicate positions from slice
func (p *ResponseParser) removeDuplicatePositions(positions []string) []string {
	seen := make(map[string]bool)
	var unique []string
	
	for _, pos := range positions {
		if !seen[pos] {
			seen[pos] = true
			unique = append(unique, pos)
		}
	}
	
	return unique
}
