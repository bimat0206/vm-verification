package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"workflow-function/shared/logger"
)

// ResponseParser handles parsing of Bedrock responses
type ResponseParser struct {
	config  *ProcessingConfig
	context *ParsingContext
	logger  logger.Logger
	patterns *ParsingPatterns
}

// ParsingPatterns contains configurable patterns for parsing
type ParsingPatterns struct {
	ContentFields       []string            `json:"contentFields"`
	ThinkingFields      []string            `json:"thinkingFields"`
	StructuredMarkers   []string            `json:"structuredMarkers"`
	SectionPatterns     map[string]*ResponsePattern `json:"sectionPatterns"`
	MachineStructure    []string            `json:"machineStructurePatterns"`
	RowStatus           []string            `json:"rowStatusPatterns"`
	EmptyPositions      []string            `json:"emptyPositionPatterns"`
	Observations        []string            `json:"observationKeywords"`
	QuantityPatterns    []string            `json:"quantityPatterns"`
	PositionPattern     string              `json:"positionPattern"`
	RowPattern          string              `json:"rowPattern"`
	ColumnPattern       string              `json:"columnPattern"`
	MinContentLength    int                 `json:"minContentLength"`
	MaxSectionLength    int                 `json:"maxSectionLength"`
}

// NewResponseParser creates a new response parser
func NewResponseParser(config *ProcessingConfig, context *ParsingContext, log logger.Logger) *ResponseParser {
	patterns, err := loadParsingPatterns(context)
	if err != nil {
		log.Warn("Failed to load parsing patterns, using defaults", map[string]interface{}{
			"error": err.Error(),
		})
		patterns = getDefaultParsingPatterns()
	}
	
	return &ResponseParser{
		config:   config,
		context:  context,
		logger:   log,
		patterns: patterns,
	}
}

// loadParsingPatterns loads parsing patterns from configuration or context
func loadParsingPatterns(context *ParsingContext) (*ParsingPatterns, error) {
	// Try to load from layout metadata if available
	if context.LayoutMetadata != nil {
		if patternsData, exists := context.LayoutMetadata["parsingPatterns"]; exists {
			patterns := &ParsingPatterns{}
			if data, ok := patternsData.(map[string]interface{}); ok {
				if err := mapToStruct(data, patterns); err != nil {
					return nil, fmt.Errorf("failed to parse patterns from layout metadata: %w", err)
				}
				return patterns, nil
			}
		}
	}
	
	// Try to load from historical context if available
	if context.HistoricalContext != nil {
		if patternsData, exists := context.HistoricalContext["parsingPatterns"]; exists {
			patterns := &ParsingPatterns{}
			if data, ok := patternsData.(map[string]interface{}); ok {
				if err := mapToStruct(data, patterns); err != nil {
					return nil, fmt.Errorf("failed to parse patterns from historical context: %w", err)
				}
				return patterns, nil
			}
		}
	}
	
	// Load from configuration if available
	if context.ParsingConfig != nil {
		return getPatternsFromConfig(context.ParsingConfig), nil
	}
	
	return getDefaultParsingPatterns(), nil
}

// getDefaultParsingPatterns returns minimal default patterns
func getDefaultParsingPatterns() *ParsingPatterns {
	return &ParsingPatterns{
		ContentFields:     []string{"content", "response", "text", "output", "result"},
		ThinkingFields:    []string{"thinking", "reasoning", "analysis"},
		StructuredMarkers: []string{},
		SectionPatterns:   make(map[string]*ResponsePattern),
		MachineStructure:  []string{`(?i)(\d+)\s+.*?(\d+)`},
		RowStatus:         []string{`(?i)([A-Z]+)[:\s]+([^.]+\.)`},
		EmptyPositions:    []string{`([A-Z]\d+)`},
		Observations:      []string{"visible", "appears", "note"},
		QuantityPatterns:  []string{`(\d+)`},
		PositionPattern:   `[A-Z]\d+`,
		RowPattern:        `[A-Z]`,
		ColumnPattern:     `\d+`,
		MinContentLength:  50,
		MaxSectionLength:  10000,
	}
}

// getPatternsFromConfig derives patterns from processing configuration
func getPatternsFromConfig(config *ProcessingConfig) *ParsingPatterns {
	patterns := getDefaultParsingPatterns()
	
	// Adjust patterns based on configuration
	if config.ExtractMachineStructure {
		patterns.SectionPatterns["machine_structure"] = &ResponsePattern{
			Name:     "Machine Structure",
			Pattern:  `(?s)(\d+)\s+.*?(\d+)`,
			Required: false,
		}
	}
	
	if config.ValidateCompleteness {
		patterns.SectionPatterns["completeness_check"] = &ResponsePattern{
			Name:     "Completeness Check",
			Pattern:  `(?i)(complete|incomplete|partial)`,
			Required: false,
		}
	}
	
	return patterns
}

// mapToStruct converts a map to a struct (simple implementation)
func mapToStruct(data map[string]interface{}, target interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you might use reflection or a JSON library
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// ParseResponse parses a raw Bedrock response into structured data
func (p *ResponseParser) ParseResponse(turn1Response map[string]interface{}) (*ParsedResponse, error) {
	startTime := time.Now()
	
	p.logger.Debug("Starting response parsing", map[string]interface{}{
		"responseKeys": p.getMapKeys(turn1Response),
		"hasConfig":    p.config != nil,
		"hasContext":   p.context != nil,
	})
	
	parsed := &ParsedResponse{
		ParsedSections: make(map[string]string),
		ExtractedData:  make(map[string]interface{}),
		ParsingErrors:  []string{},
	}
	
	// Extract main content
	mainContent, err := p.extractMainContent(turn1Response)
	if err != nil {
		parsed.ParsingErrors = append(parsed.ParsingErrors, fmt.Sprintf("Failed to extract main content: %v", err))
		if !p.config.FallbackToTextParsing {
			return parsed, err
		}
	}
	parsed.MainContent = mainContent
	
	// Extract thinking content if available
	thinkingContent, _ := p.extractThinkingContent(turn1Response)
	parsed.ThinkingContent = thinkingContent
	
	// Determine if response is structured
	parsed.IsStructured = p.isStructuredResponse(mainContent)
	
	// Parse sections based on structure
	if parsed.IsStructured {
		err = p.parseStructuredResponse(parsed)
	} else {
		err = p.parseUnstructuredResponse(parsed)
	}
	
	if err != nil && !p.config.FallbackToTextParsing {
		return parsed, err
	}
	
	// Extract structured data
	p.extractStructuredData(parsed)
	
	p.logger.Debug("Response parsing completed", map[string]interface{}{
		"isStructured":    parsed.IsStructured,
		"sectionsFound":   len(parsed.ParsedSections),
		"hasErrors":       len(parsed.ParsingErrors) > 0,
		"processingTime":  time.Since(startTime),
	})
	
	return parsed, nil
}

// extractMainContent extracts the main response content
func (p *ResponseParser) extractMainContent(response map[string]interface{}) (string, error) {
	// Try configured content fields
	for _, field := range p.patterns.ContentFields {
		if value, exists := response[field]; exists {
			switch v := value.(type) {
			case string:
				return v, nil
			case map[string]interface{}:
				// Handle nested content structures
				if text, ok := v["text"].(string); ok {
					return text, nil
				}
			}
		}
	}
	
	// Fallback: try to find any significant string content
	for key, value := range response {
		if str, ok := value.(string); ok && len(str) > p.patterns.MinContentLength {
			p.logger.Warn("Using fallback content extraction", map[string]interface{}{
				"fieldName":     key,
				"contentLength": len(str),
			})
			return str, nil
		}
	}
	
	return "", fmt.Errorf("no main content found in response")
}

// extractThinkingContent extracts the thinking/reasoning content
func (p *ResponseParser) extractThinkingContent(response map[string]interface{}) (string, error) {
	for _, field := range p.patterns.ThinkingFields {
		if value, exists := response[field]; exists {
			if str, ok := value.(string); ok {
				return str, nil
			}
		}
	}
	
	return "", nil
}

// isStructuredResponse determines if the response is in a structured format
func (p *ResponseParser) isStructuredResponse(content string) bool {
	// Check for configured structured markers
	content = strings.ToUpper(content)
	for _, marker := range p.patterns.StructuredMarkers {
		if strings.Contains(content, strings.ToUpper(marker)) {
			return true
		}
	}
	
	// Check for JSON-like structure
	if strings.Contains(content, "{") && strings.Contains(content, "}") {
		return true
	}
	
	// Check for markdown-like structure
	if strings.Contains(content, "**") || strings.Contains(content, "##") {
		return true
	}
	
	return false
}

// parseStructuredResponse parses a structured response
func (p *ResponseParser) parseStructuredResponse(parsed *ParsedResponse) error {
	content := parsed.MainContent
	
	// Extract sections using configured patterns
	for sectionName, pattern := range p.patterns.SectionPatterns {
		if matches := p.extractWithPattern(content, pattern.Pattern); len(matches) > 0 {
			sectionContent := strings.TrimSpace(matches[0])
			if len(sectionContent) <= p.patterns.MaxSectionLength {
				parsed.ParsedSections[sectionName] = sectionContent
			} else {
				parsed.ParsingErrors = append(parsed.ParsingErrors, 
					fmt.Sprintf("Section %s exceeds maximum length", sectionName))
			}
		}
	}
	
	return nil
}

// parseUnstructuredResponse parses an unstructured text response
func (p *ResponseParser) parseUnstructuredResponse(parsed *ParsedResponse) error {
	content := parsed.MainContent
	sections := make(map[string]string)
	
	// Extract information using configured patterns
	if machineInfo := p.extractUsingPatterns(content, p.patterns.MachineStructure, "machine_structure"); machineInfo != "" {
		sections["machine_structure"] = machineInfo
	}
	
	if rowInfo := p.extractUsingPatterns(content, p.patterns.RowStatus, "row_status"); rowInfo != "" {
		sections["row_status"] = rowInfo
	}
	
	if emptyInfo := p.extractUsingPatterns(content, p.patterns.EmptyPositions, "empty_positions"); emptyInfo != "" {
		sections["empty_positions"] = emptyInfo
	}
	
	if observations := p.extractObservations(content); observations != "" {
		sections["observations"] = observations
	}
	
	parsed.ParsedSections = sections
	return nil
}

// extractUsingPatterns extracts information using a list of patterns
func (p *ResponseParser) extractUsingPatterns(content string, patterns []string, sectionType string) string {
	for _, pattern := range patterns {
		if matches := p.extractWithPattern(content, pattern); len(matches) > 0 {
			result := strings.TrimSpace(matches[0])
			if result != "" {
				p.logger.Debug("Pattern matched", map[string]interface{}{
					"sectionType": sectionType,
					"pattern":     pattern,
					"resultLen":   len(result),
				})
				return result
			}
		}
	}
	return ""
}

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

// extractStructuredData extracts structured data from parsed sections
func (p *ResponseParser) extractStructuredData(parsed *ParsedResponse) {
	// Extract machine structure if available
	if structureData := p.extractMachineStructureData(parsed); structureData != nil {
		parsed.ExtractedData["machineStructure"] = structureData
	}
	
	// Extract row states if available
	if rowStates := p.extractRowStatesData(parsed); len(rowStates) > 0 {
		parsed.ExtractedData["rowStates"] = rowStates
	}
	
	// Extract empty positions if available
	if emptyPositions := p.extractEmptyPositionsData(parsed); len(emptyPositions) > 0 {
		parsed.ExtractedData["emptyPositions"] = emptyPositions
	}
	
	// Extract other data based on available sections
	for sectionName, sectionContent := range parsed.ParsedSections {
		if sectionName != "machine_structure" && sectionName != "row_status" && sectionName != "empty_positions" {
			parsed.ExtractedData[sectionName] = sectionContent
		}
	}
}

// extractMachineStructureData extracts machine structure data
func (p *ResponseParser) extractMachineStructureData(parsed *ParsedResponse) *MachineStructure {
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
				structure := &MachineStructure{
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

// parseMachineStructureFromSection parses machine structure from a specific section
func (p *ResponseParser) parseMachineStructureFromSection(section string) *MachineStructure {
	// Try different parsing approaches
	parsers := []func(string) *MachineStructure{
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
func (p *ResponseParser) parseStructuredFormat(section string) *MachineStructure {
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
				structure := &MachineStructure{
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
func (p *ResponseParser) parseNaturalLanguageFormat(section string) *MachineStructure {
	// Extract numbers and context for natural language
	words := strings.Fields(strings.ToLower(section))
	var rowCount, colCount int
	var err error
	
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
			
			if strings.Contains(context, "row") && rowCount == 0 {
				rowCount = num
			} else if (strings.Contains(context, "column") || strings.Contains(context, "slot")) && colCount == 0 {
				colCount = num
			}
		}
	}
	
	if rowCount > 0 && colCount > 0 {
		structure := &MachineStructure{
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
func (p *ResponseParser) parseNumberSequenceFormat(section string) *MachineStructure {
	// Extract all numbers from the section
	numberPattern := `\d+`
	numbers := regexp.MustCompile(numberPattern).FindAllString(section, -1)
	
	if len(numbers) >= 2 {
		// Try to identify which might be rows vs columns based on typical values
		rowCount, _ := strconv.Atoi(numbers[0])
		colCount, _ := strconv.Atoi(numbers[1])
		
		// Basic sanity check
		if rowCount > 0 && colCount > 0 && rowCount <= 26 && colCount <= 100 {
			structure := &MachineStructure{
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
	if strings.Contains(rangeStr, "-") {
		return p.parseRangeSequence(rangeStr, sequenceType)
	}
	
	// Handle comma-separated format (e.g., "A, B, C" or "1, 2, 3")
	if strings.Contains(rangeStr, ",") {
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

// extractRowStatesData extracts row state information
func (p *ResponseParser) extractRowStatesData(parsed *ParsedResponse) map[string]*RowState {
	states := make(map[string]*RowState)
	
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
func (p *ResponseParser) parseRowStatesFromSection(section string) map[string]*RowState {
	states := make(map[string]*RowState)
	
	// Use configured patterns for row status
	for _, pattern := range p.patterns.RowStatus {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(section, -1)
		
		for _, match := range matches {
			if len(match) >= 3 {
				row := match[1]
				statusText := strings.TrimSpace(match[2])
				
				state := &RowState{
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
func (p *ResponseParser) extractEmptyPositionsData(parsed *ParsedResponse) []string {
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

// extractWithPattern extracts text using a regex pattern
func (p *ResponseParser) extractWithPattern(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		return matches[1:]
	}
	
	return []string{}
}

// getMapKeys gets keys from a map for logging
func (p *ResponseParser) getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
// Parser interface defines the methods needed by the Processor
type Parser interface {
	ParseValidationResponse(response map[string]interface{}) map[string]interface{}
	ParseVisualAnalysis(response map[string]interface{}) map[string]interface{}
	ParseMachineStructure(response map[string]interface{}) map[string]interface{}
	ParseMachineState(response map[string]interface{}) map[string]interface{}
	ExtractObservations(response map[string]interface{}) map[string]interface{}
}

// ParserImpl implements the Parser interface using ResponseParser
type ParserImpl struct {
	responseParser *ResponseParser
	logger         logger.Logger
}

// NewParser creates a new Parser implementation
func NewParser(log logger.Logger) *Parser {
	// Create default processing config and context
	config := DefaultProcessingConfig()
	context := &ParsingContext{
		VerificationType: "",
		HasHistoricalContext: false,
		ParsingConfig: config,
	}
	
	responseParser := NewResponseParser(config, context, log)
	
	impl := &ParserImpl{
		responseParser: responseParser,
		logger:         log,
	}
	
	// Return the interface, not the implementation
	var parser Parser = impl
	return &parser
}

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
		if ms, ok := structure.(*MachineStructure); ok {
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
		if ms, ok := structure.(*MachineStructure); ok {
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
	if rowStates, ok := result["rowStates"].(map[string]*RowState); ok {
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
func (p *ParserImpl) calculateStateSummary(rowStates map[string]*RowState) map[string]interface{} {
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