package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	
	"workflow-function/ProcessTurn1Response/internal/types"
)

// parseMachineStructureFromSection parses machine structure from a specific section
func (p *ResponseParser) parseMachineStructureFromSection(section string) *types.MachineStructure {
	// Try different parsing approaches
	parsers := []func(string) *types.MachineStructure{
		p.parseStructuredFormat,
		p.parseNaturalLanguageFormat,
		p.parseNumberSequenceFormat,
	}
	
	for _, parser := range parsers {
		if structure := parser(section); structure != nil {
			return structure
		}
	}
	
	return nil
}

// parseStructuredFormat parses "Rows: X, Columns: Y" format
func (p *ResponseParser) parseStructuredFormat(section string) *types.MachineStructure {
	patterns := []string{
		`(?i)rows?:\s*(\d+).*?columns?:\s*(\d+)`,
		`(?i)(\d+)\s*rows?.*?(\d+)\s*columns?`,
		`(?i)(\d+)x(\d+)`,
	}
	
	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(section)
		if len(matches) >= 3 {
			rowCount, rowErr := strconv.Atoi(matches[1])
			colCount, colErr := strconv.Atoi(matches[2])
			
			if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
				structure := &types.MachineStructure{
					RowCount:           rowCount,
					ColumnsPerRow:      colCount,
					StructureConfirmed: true,
				}
				structure.CalculateTotalPositions()
				return structure
			}
		}
	}
	
	return nil
}

// parseNaturalLanguageFormat parses natural language descriptions
func (p *ResponseParser) parseNaturalLanguageFormat(section string) *types.MachineStructure {
	// Extract numbers and context for natural language
	words := strings.Fields(strings.ToLower(section))
	var rowCount, colCount int
	
	for i, word := range words {
		if num, numErr := strconv.Atoi(word); numErr == nil {
			// Check context around the number
			context := ""
			if i > 0 {
				context += words[i-1] + " "
			}
			context += word
			if i < len(words)-1 {
				context += " " + words[i+1]
			}
			
			if contains(context, "row") && rowCount == 0 {
				rowCount = num
			} else if (contains(context, "column") || contains(context, "slot")) && colCount == 0 {
				colCount = num
			}
		}
	}
	
	if rowCount > 0 && colCount > 0 {
		structure := &types.MachineStructure{
			RowCount:           rowCount,
			ColumnsPerRow:      colCount,
			StructureConfirmed: true,
		}
		structure.CalculateTotalPositions()
		return structure
	}
	
	return nil
}

// parseNumberSequenceFormat parses sequences of numbers
func (p *ResponseParser) parseNumberSequenceFormat(section string) *types.MachineStructure {
	// Extract all numbers from the section
	numberPattern := `\d+`
	numbers := regexp.MustCompile(numberPattern).FindAllString(section, -1)
	
	if len(numbers) >= 2 {
		// Try to identify which might be rows vs columns based on typical values
		rowCount, _ := strconv.Atoi(numbers[0])
		colCount, _ := strconv.Atoi(numbers[1])
		
		// Basic sanity check
		if rowCount > 0 && colCount > 0 && rowCount <= 26 && colCount <= 100 {
			structure := &types.MachineStructure{
				RowCount:           rowCount,
				ColumnsPerRow:      colCount,
				StructureConfirmed: false, // Less confident since we're guessing
			}
			structure.CalculateTotalPositions()
			return structure
		}
	}
	
	return nil
}

// extractRowOrder extracts the row order dynamically
func (p *ResponseParser) extractRowOrder(content string, expectedCount int) []string {
	// Try to find explicit row order information
	patterns := []string{
		`(?i)rows?\s+([A-Z](?:[-,\s]*[A-Z])*)`,
		`(?i)\(([A-Z](?:[-,\s]*[A-Z])*)\)`,
		`([A-Z]-[A-Z])`,
	}
	
	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(content)
		if len(matches) >= 2 {
			order := p.parseSequence(matches[1], expectedCount, "row")
			if len(order) == expectedCount {
				return order
			}
		}
	}
	
	// Generate default alphabetical sequence
	if expectedCount > 0 && expectedCount <= 26 {
		order := make([]string, expectedCount)
		for i := 0; i < expectedCount; i++ {
			order[i] = string(rune('A' + i))
		}
		return order
	}
	
	return []string{}
}

// extractColumnOrder extracts the column order dynamically
func (p *ResponseParser) extractColumnOrder(content string, expectedCount int) []string {
	// Try to find explicit column order information
	patterns := []string{
		`(?i)columns?\s+([\d-,\s]+)`,
		`(?i)slots?\s+([\d-,\s]+)`,
		`(?i)\(([\d-,\s]+)\)`,
		`(\d+-\d+)`,
	}
	
	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(content)
		if len(matches) >= 2 {
			order := p.parseSequence(matches[1], expectedCount, "column")
			if len(order) == expectedCount {
				return order
			}
		}
	}
	
	// Generate default numerical sequence
	if expectedCount > 0 {
		order := make([]string, expectedCount)
		for i := 0; i < expectedCount; i++ {
			order[i] = fmt.Sprintf("%d", i+1)
		}
		return order
	}
	
	return []string{}
}

// parseSequence parses sequence strings like "A-F" or "1-10"
func (p *ResponseParser) parseSequence(rangeStr string, expectedCount int, sequenceType string) []string {
	rangeStr = strings.TrimSpace(rangeStr)
	
	// Handle range format (e.g., "A-F" or "1-10")
	if contains(rangeStr, "-") {
		return p.parseRangeSequence(rangeStr, sequenceType)
	}
	
	// Handle comma-separated format (e.g., "A, B, C" or "1, 2, 3")
	if contains(rangeStr, ",") {
		return p.parseCommaSequence(rangeStr)
	}
	
	return []string{}
}

// parseRangeSequence parses range sequences like "A-F" or "1-10"
func (p *ResponseParser) parseRangeSequence(rangeStr, sequenceType string) []string {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return []string{}
	}
	
	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])
	
	if sequenceType == "row" && len(start) == 1 && len(end) == 1 {
		// Handle alphabetical range
		startRune := rune(start[0])
		endRune := rune(end[0])
		
		if startRune >= 'A' && startRune <= 'Z' && endRune >= 'A' && endRune <= 'Z' {
			var result []string
			for r := startRune; r <= endRune; r++ {
				result = append(result, string(r))
			}
			return result
		}
	} else if sequenceType == "column" {
		// Handle numerical range
		startNum, startErr := strconv.Atoi(start)
		endNum, endErr := strconv.Atoi(end)
		
		if startErr == nil && endErr == nil && startNum > 0 && endNum >= startNum {
			var result []string
			for i := startNum; i <= endNum; i++ {
				result = append(result, fmt.Sprintf("%d", i))
			}
			return result
		}
	}
	
	return []string{}
}

// parseCommaSequence parses comma-separated sequences
func (p *ResponseParser) parseCommaSequence(sequence string) []string {
	parts := strings.Split(sequence, ",")
	var result []string
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}
