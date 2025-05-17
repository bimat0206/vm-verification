package parser

import (
	"regexp"
	"strconv"
	"strings"
	
	"workflow-function/ProcessTurn1Response/internal/types"
)

// extractObservations extracts general observations using keywords from configuration
func (p *ResponseParser) extractObservations(content string) string {
	// Define default observation keywords
	observationKeywords := []string{"visible", "appears", "note"}
	
	sentences := strings.Split(content, ".")
	var observations []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 10 {
			continue
		}
		
		lowerSentence := strings.ToLower(sentence)
		for _, keyword := range observationKeywords {
			if contains(lowerSentence, keyword) {
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
		// Use the basic implementation to avoid duplicate method issues
		// Extract row and column counts from the section
		rowColPattern := `(?i)(\d+)\s*rows?\s+.*?(\d+)\s*columns?`
		matches := regexp.MustCompile(rowColPattern).FindStringSubmatch(structureSection)
		
		if len(matches) >= 3 {
			rowCount, rowErr := strconv.Atoi(matches[1])
			colCount, colErr := strconv.Atoi(matches[2])
			
			if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
				// Create structure
				structure := &types.MachineStructure{
					RowCount:           rowCount,
					ColumnsPerRow:      colCount,
					StructureConfirmed: true,
				}
				
				// Extract row and column orders
				structure.RowOrder = p.extractRowOrder(structureSection, rowCount)
				structure.ColumnOrder = p.extractColumnOrder(structureSection, colCount)
				structure.CalculateTotalPositions()
				
				return structure
			}
		}
		
		// Try fallback pattern
		fallbackPattern := `(?i)(\d+).*?(\d+)`
		matches = regexp.MustCompile(fallbackPattern).FindStringSubmatch(structureSection)
		
		if len(matches) >= 3 {
			rowCount, rowErr := strconv.Atoi(matches[1])
			colCount, colErr := strconv.Atoi(matches[2])
			
			if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
				structure := &types.MachineStructure{
					RowCount:           rowCount,
					ColumnsPerRow:      colCount,
					StructureConfirmed: true,
				}
				
				structure.RowOrder = p.extractRowOrder(structureSection, rowCount)
				structure.ColumnOrder = p.extractColumnOrder(structureSection, colCount)
				structure.CalculateTotalPositions()
				return structure
			}
		}
	}
	
	// Try to extract from content using patterns
	content := parsed.MainContent
	
	// Define default machine structure pattern
	machineStructurePatterns := []string{`(?i)(\d+)\s+.*?(\d+)`}
	
	for _, pattern := range machineStructurePatterns {
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
	
	// Use default patterns for row status
	rowStatusPatterns := []string{`(?i)([A-Z]+)[:\s]+([^.]+\.)`}
	for _, pattern := range rowStatusPatterns {
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
		if contains(text, keyword) {
			return status
		}
	}
	
	// Heuristic analysis
	if contains(text, "visible") && !contains(text, "coil") {
		return "Partial"
	} else if contains(text, "coil") {
		return "Empty"
	}
	
	return "Unknown"
}

// extractQuantityFromText extracts quantity using default patterns
func (p *ResponseParser) extractQuantityFromText(text string) int {
	// Define default quantity patterns
	quantityPatterns := []string{`(\d+)`}
	for _, pattern := range quantityPatterns {
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
	
	// Try to extract from other sections using default patterns
	emptyPositionPatterns := []string{`([A-Z]\d+)`}
	for _, pattern := range emptyPositionPatterns {
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

// parsePositionsList parses a list of positions using default pattern
func (p *ResponseParser) parsePositionsList(text string) []string {
	// Define default position pattern
	positionPattern := `[A-Z]\d+`
	matches := regexp.MustCompile(positionPattern).FindAllString(text, -1)
	
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

// extractRowOrder extracts row designations (A, B, C, etc.)
func (p *ResponseParser) extractRowOrder(content string, rowCount int) []string {
	rows := make([]string, 0, rowCount)
	
	// Try to extract from content using common patterns
	rowPattern := `[A-Z]`
	matches := regexp.MustCompile(rowPattern).FindAllString(content, -1)
	
	// Filter for unique row identifiers
	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match] && len(rows) < rowCount {
			seen[match] = true
			rows = append(rows, match)
		}
	}
	
	// If not enough rows found, generate default (A, B, C, etc.)
	if len(rows) < rowCount {
		rows = make([]string, rowCount)
		for i := 0; i < rowCount; i++ {
			rows[i] = string('A' + i)
		}
	}
	
	return rows
}

// extractColumnOrder extracts column designations (1, 2, 3, etc.)
func (p *ResponseParser) extractColumnOrder(content string, colCount int) []string {
	cols := make([]string, 0, colCount)
	
	// Try to extract from content using common patterns
	colPattern := `\d+`
	matches := regexp.MustCompile(colPattern).FindAllString(content, -1)
	
	// Filter for unique column identifiers
	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match] && len(cols) < colCount {
			seen[match] = true
			cols = append(cols, match)
		}
	}
	
	// If not enough columns found, generate default (1, 2, 3, etc.)
	if len(cols) < colCount {
		cols = make([]string, colCount)
		for i := 0; i < colCount; i++ {
			cols[i] = strconv.Itoa(i + 1)
		}
	}
	
	return cols
}

// parseMachineStructureFromSection parses machine structure from a section
func (p *ResponseParser) parseMachineStructureFromSection(section string) *types.MachineStructure {
	// Try to find row and column counts
	rowColPattern := `(?i)(\d+)\s*rows?\s+.*?(\d+)\s*columns?`
	matches := regexp.MustCompile(rowColPattern).FindStringSubmatch(section)
	
	if len(matches) >= 3 {
		rowCount, rowErr := strconv.Atoi(matches[1])
		colCount, colErr := strconv.Atoi(matches[2])
		
		if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
			// Create structure
			structure := &types.MachineStructure{
				RowCount:           rowCount,
				ColumnsPerRow:      colCount,
				StructureConfirmed: true,
			}
			
			// Extract row and column orders
			structure.RowOrder = p.extractRowOrder(section, rowCount)
			structure.ColumnOrder = p.extractColumnOrder(section, colCount)
			structure.CalculateTotalPositions()
			
			return structure
		}
	}
	
	// Fallback: Try more generic patterns
	fallbackPattern := `(?i)(\d+).*?(\d+)`
	matches = regexp.MustCompile(fallbackPattern).FindStringSubmatch(section)
	
	if len(matches) >= 3 {
		// Assume first number is rows, second is columns
		rowCount, rowErr := strconv.Atoi(matches[1])
		colCount, colErr := strconv.Atoi(matches[2])
		
		if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
			structure := &types.MachineStructure{
				RowCount:           rowCount,
				ColumnsPerRow:      colCount,
				StructureConfirmed: true,
			}
			
			// Generate default row/column orders
			structure.RowOrder = make([]string, rowCount)
			for i := 0; i < rowCount; i++ {
				structure.RowOrder[i] = string('A' + i)
			}
			
			structure.ColumnOrder = make([]string, colCount)
			for i := 0; i < colCount; i++ {
				structure.ColumnOrder[i] = strconv.Itoa(i + 1)
			}
			
			structure.CalculateTotalPositions()
			return structure
		}
	}
	
	return nil
}
