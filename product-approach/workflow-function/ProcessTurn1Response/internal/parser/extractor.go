// Package parser provides parsing utilities for Turn 1 responses
package parser

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// ResponseExtractor extracts and processes raw response content
type ResponseExtractor interface {
	// ExtractContent extracts structured content from raw response
	ExtractContent(responseContent string) (*types.ParsedResponse, error)
	
	// ExtractMainContent extracts the main content from response JSON
	ExtractMainContent(response map[string]interface{}) (string, error)
	
	// ExtractThinkingContent extracts the thinking/reasoning content
	ExtractThinkingContent(response map[string]interface{}) (string, error)
	
	// ParseSections extracts structured sections from content
	ParseSections(content string) (map[string]string, error)
}

// StructureExtractor extracts machine structure information
type StructureExtractor interface {
	// ExtractMachineStructure extracts the machine structure
	ExtractMachineStructure(content string) (*types.MachineStructure, error)
	
	// ExtractRowOrder extracts row designations (A, B, C, etc.)
	ExtractRowOrder(content string, rowCount int) ([]string, error)
	
	// ExtractColumnOrder extracts column designations (1, 2, 3, etc.)
	ExtractColumnOrder(content string, colCount int) ([]string, error)
}

// StateExtractor extracts machine state information
type StateExtractor interface {
	// ExtractMachineState extracts the overall state of the machine
	ExtractMachineState(content string) (*types.ExtractedState, error)
	
	// ExtractRowStates extracts individual row states
	ExtractRowStates(content string) (map[string]*types.RowState, error)
	
	// ExtractPositions extracts position information
	ExtractPositions(content string) ([]string, []string, error)
}

// ObservationFinder extracts observations and notes
type ObservationFinder interface {
	// FindObservations extracts general observations from content
	FindObservations(content string) ([]string, error)
}

// DefaultResponseExtractor implements ResponseExtractor
type DefaultResponseExtractor struct {
	patterns PatternProvider
	logger   *slog.Logger
}

// DefaultStructureExtractor implements StructureExtractor
type DefaultStructureExtractor struct {
	patterns PatternProvider
	logger   *slog.Logger
}

// DefaultStateExtractor implements StateExtractor
type DefaultStateExtractor struct {
	patterns PatternProvider
	logger   *slog.Logger
}

// DefaultObservationFinder implements ObservationFinder
type DefaultObservationFinder struct {
	patterns PatternProvider
	logger   *slog.Logger
}

// NewResponseExtractor creates a new ResponseExtractor
func NewResponseExtractor(patterns PatternProvider, logger *slog.Logger) ResponseExtractor {
	return &DefaultResponseExtractor{
		patterns: patterns,
		logger:   logger,
	}
}

// NewStructureExtractor creates a new StructureExtractor
func NewStructureExtractor(patterns PatternProvider, logger *slog.Logger) StructureExtractor {
	return &DefaultStructureExtractor{
		patterns: patterns,
		logger:   logger,
	}
}

// NewStateExtractor creates a new StateExtractor
func NewStateExtractor(patterns PatternProvider, logger *slog.Logger) StateExtractor {
	return &DefaultStateExtractor{
		patterns: patterns,
		logger:   logger,
	}
}

// NewObservationFinder creates a new ObservationFinder
func NewObservationFinder(patterns PatternProvider, logger *slog.Logger) ObservationFinder {
	return &DefaultObservationFinder{
		patterns: patterns,
		logger:   logger,
	}
}

// ExtractContent extracts structured content from raw response
func (re *DefaultResponseExtractor) ExtractContent(responseContent string) (*types.ParsedResponse, error) {
	if responseContent == "" {
		return nil, errors.New("response content is empty")
	}
	
	parsed := &types.ParsedResponse{
		MainContent:    responseContent,
		ParsedSections: make(map[string]string),
		ExtractedData:  make(map[string]interface{}),
		ParsingErrors:  []string{},
	}
	
	// Attempt to parse structured sections
	sections, err := re.ParseSections(responseContent)
	if err == nil {
		parsed.ParsedSections = sections
		parsed.IsStructured = true
	} else {
		// Fallback to simpler extraction
		re.logger.Warn("Failed to parse structured sections, using fallback", 
			slog.String("error", err.Error()))
		
		// Extract sections using simple patterns
		machineStructure := re.extractSectionWithPattern(responseContent, 
			re.patterns.GetPattern(PatternTypeMachineStructure))
		if machineStructure != "" {
			parsed.ParsedSections["machine_structure"] = machineStructure
		}
		
		rowStatus := re.extractSectionWithPattern(responseContent, 
			re.patterns.GetPattern(PatternTypeRowStatus))
		if rowStatus != "" {
			parsed.ParsedSections["row_status"] = rowStatus
		}
		
		emptyPositions := re.extractSectionWithPattern(responseContent, 
			re.patterns.GetPattern(PatternTypeEmptyPositions))
		if emptyPositions != "" {
			parsed.ParsedSections["empty_positions"] = emptyPositions
		}
	}
	
	return parsed, nil
}

// ExtractMainContent extracts the main content from response JSON
func (re *DefaultResponseExtractor) ExtractMainContent(response map[string]interface{}) (string, error) {
	// Try known content fields based on pattern configuration
	contentFields := re.patterns.GetContentFields()
	for _, field := range contentFields {
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
	minContentLength := re.patterns.GetMinContentLength()
	for key, value := range response {
		if str, ok := value.(string); ok && len(str) > minContentLength {
			re.logger.Warn("Using fallback content extraction", 
				slog.String("fieldName", key),
				slog.Int("contentLength", len(str)))
			return str, nil
		}
	}
	
	return "", fmt.Errorf("no main content found in response")
}

// ExtractThinkingContent extracts the thinking/reasoning content
func (re *DefaultResponseExtractor) ExtractThinkingContent(response map[string]interface{}) (string, error) {
	thinkingFields := re.patterns.GetThinkingFields()
	for _, field := range thinkingFields {
		if value, exists := response[field]; exists {
			if str, ok := value.(string); ok {
				return str, nil
			}
		}
	}
	
	return "", nil // Not finding thinking content is okay
}

// ParseSections extracts structured sections from content
func (re *DefaultResponseExtractor) ParseSections(content string) (map[string]string, error) {
	sections := make(map[string]string)
	
	// Check for markdown-style sections
	if strings.Contains(content, "#") {
		// Split by markdown headers
		headerPattern := `(?m)^#{1,3}\s+([^#\n]+)$`
		headerMatches := regexp.MustCompile(headerPattern).FindAllStringSubmatchIndex(content, -1)
		
		if len(headerMatches) > 0 {
			for i, match := range headerMatches {
				sectionStart := match[1] // End of header
				var sectionEnd int
				
				if i < len(headerMatches)-1 {
					sectionEnd = headerMatches[i+1][0] // Start of next header
				} else {
					sectionEnd = len(content)
				}
				
				headerName := content[match[2]:match[3]]
				sectionContent := strings.TrimSpace(content[sectionStart:sectionEnd])
				
				// Convert header to section key
				sectionKey := normalizeHeaderToKey(headerName)
				if sectionKey != "" && sectionContent != "" {
					sections[sectionKey] = sectionContent
				}
			}
		}
	}
	
	// If no markdown sections found, try JSON block extraction
	if len(sections) == 0 && (strings.Contains(content, "{") && strings.Contains(content, "}")) {
		jsonPattern := `\{([^{}]*(\{[^{}]*\}[^{}]*)*)\}`
		jsonMatches := regexp.MustCompile(jsonPattern).FindAllString(content, -1)
		
		if len(jsonMatches) > 0 {
			// For simplicity, just store the JSON blocks
			for i, match := range jsonMatches {
				sections[fmt.Sprintf("json_block_%d", i)] = match
			}
		}
	}
	
	// If still no sections, try to extract based on defined patterns
	if len(sections) == 0 {
		sectionPatterns := re.patterns.GetSectionPatterns()
		for sectionName, pattern := range sectionPatterns {
			if match := re.extractSectionWithPattern(content, pattern); match != "" {
				sections[sectionName] = match
			}
		}
	}
	
	if len(sections) == 0 {
		return nil, errors.New("no structured sections found")
	}
	
	return sections, nil
}

// extractSectionWithPattern extracts a section using a regex pattern
func (re *DefaultResponseExtractor) extractSectionWithPattern(content, pattern string) string {
	re1 := regexp.MustCompile(pattern)
	matches := re1.FindStringSubmatch(content)
	
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// ExtractMachineStructure extracts machine structure from content
func (se *DefaultStructureExtractor) ExtractMachineStructure(content string) (*types.MachineStructure, error) {
	// Use the row-column pattern
	rowColPattern := se.patterns.GetPattern(PatternTypeRowColumn)
	matches := regexp.MustCompile(rowColPattern).FindStringSubmatch(content)
	
	if len(matches) >= 3 {
		rowCount, rowErr := strconv.Atoi(matches[1])
		colCount, colErr := strconv.Atoi(matches[2])
		
		if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
			structure := &types.MachineStructure{
				RowCount:           rowCount,
				ColumnsPerRow:      colCount,
				StructureConfirmed: true,
			}
			
			rowOrder, err := se.ExtractRowOrder(content, rowCount)
			if err == nil {
				structure.RowOrder = rowOrder
			} else {
				// Generate default row order (A, B, C, etc.)
				structure.RowOrder = generateDefaultRowOrder(rowCount)
			}
			
			colOrder, err := se.ExtractColumnOrder(content, colCount)
			if err == nil {
				structure.ColumnOrder = colOrder
			} else {
				// Generate default column order (1, 2, 3, etc.)
				structure.ColumnOrder = generateDefaultColumnOrder(colCount)
			}
			
			structure.CalculateTotalPositions()
			return structure, nil
		}
	}
	
	// Try fallback pattern
	fallbackPattern := se.patterns.GetPattern(PatternTypeFallbackStructure)
	matches = regexp.MustCompile(fallbackPattern).FindStringSubmatch(content)
	
	if len(matches) >= 3 {
		rowCount, rowErr := strconv.Atoi(matches[1])
		colCount, colErr := strconv.Atoi(matches[2])
		
		if rowErr == nil && colErr == nil && rowCount > 0 && colCount > 0 {
			structure := &types.MachineStructure{
				RowCount:           rowCount,
				ColumnsPerRow:      colCount,
				StructureConfirmed: true,
				RowOrder:           generateDefaultRowOrder(rowCount),
				ColumnOrder:        generateDefaultColumnOrder(colCount),
			}
			
			structure.CalculateTotalPositions()
			return structure, nil
		}
	}
	
	return nil, errors.New("failed to extract machine structure")
}

// ExtractRowOrder extracts row designations (A, B, C, etc.) from content
func (se *DefaultStructureExtractor) ExtractRowOrder(content string, rowCount int) ([]string, error) {
	if rowCount <= 0 {
		return nil, errors.New("invalid row count")
	}
	
	rowPattern := se.patterns.GetPattern(PatternTypeRow)
	matches := regexp.MustCompile(rowPattern).FindAllString(content, -1)
	
	// Filter for unique row identifiers
	seen := make(map[string]bool)
	rows := make([]string, 0, rowCount)
	
	for _, match := range matches {
		if !seen[match] && len(rows) < rowCount {
			seen[match] = true
			rows = append(rows, match)
		}
	}
	
	if len(rows) != rowCount {
		return nil, fmt.Errorf("found %d rows, expected %d", len(rows), rowCount)
	}
	
	return rows, nil
}

// ExtractColumnOrder extracts column designations (1, 2, 3, etc.) from content
func (se *DefaultStructureExtractor) ExtractColumnOrder(content string, colCount int) ([]string, error) {
	if colCount <= 0 {
		return nil, errors.New("invalid column count")
	}
	
	colPattern := se.patterns.GetPattern(PatternTypeColumn)
	matches := regexp.MustCompile(colPattern).FindAllString(content, -1)
	
	// Filter for unique column identifiers
	seen := make(map[string]bool)
	cols := make([]string, 0, colCount)
	
	for _, match := range matches {
		if !seen[match] && len(cols) < colCount {
			seen[match] = true
			cols = append(cols, match)
		}
	}
	
	if len(cols) != colCount {
		return nil, fmt.Errorf("found %d columns, expected %d", len(cols), colCount)
	}
	
	return cols, nil
}

// ExtractMachineState extracts the overall state of the machine
func (ste *DefaultStateExtractor) ExtractMachineState(content string) (*types.ExtractedState, error) {
	// Create the extracted state
	state := &types.ExtractedState{
		RowStates:       make(map[string]*types.RowState),
		EmptyPositions:  []string{},
		FilledPositions: []string{},
		Observations:    []string{},
	}
	
	// Extract row states
	rowStates, err := ste.ExtractRowStates(content)
	if err == nil {
		state.RowStates = rowStates
	} else {
		ste.logger.Warn("Failed to extract row states", slog.String("error", err.Error()))
	}
	
	// Extract positions
	emptyPositions, filledPositions, err := ste.ExtractPositions(content)
	if err == nil {
		state.EmptyPositions = emptyPositions
		state.FilledPositions = filledPositions
		state.TotalEmptyCount = len(emptyPositions)
		state.TotalFilledCount = len(filledPositions)
	} else {
		ste.logger.Warn("Failed to extract positions", slog.String("error", err.Error()))
	}
	
	// Determine overall status
	state.OverallStatus = determineOverallStatus(state.TotalFilledCount, state.TotalEmptyCount)
	
	// Find observations
	observationFinder := &DefaultObservationFinder{
		patterns: ste.patterns,
		logger:   ste.logger,
	}
	observations, err := observationFinder.FindObservations(content)
	if err == nil {
		state.Observations = observations
	}
	
	return state, nil
}

// ExtractRowStates extracts individual row states
func (ste *DefaultStateExtractor) ExtractRowStates(content string) (map[string]*types.RowState, error) {
	rowStates := make(map[string]*types.RowState)
	
	// Use row status pattern to extract status for each row
	rowStatusPattern := ste.patterns.GetPattern(PatternTypeRowStatus)
	matches := regexp.MustCompile(rowStatusPattern).FindAllStringSubmatch(content, -1)
	
	if len(matches) == 0 {
		return nil, errors.New("no row status information found")
	}
	
	for _, match := range matches {
		if len(match) >= 3 {
			row := match[1]
			statusText := strings.TrimSpace(match[2])
			
			state := &types.RowState{
				Status:   ste.parseRowStatus(statusText),
				Notes:    statusText,
				Quantity: ste.extractQuantity(statusText),
			}
			
			// Extract additional details if available
			if len(match) >= 4 && match[3] != "" {
				state.Notes += " - " + strings.TrimSpace(match[3])
			}
			
			rowStates[row] = state
		}
	}
	
	return rowStates, nil
}

// ExtractPositions extracts empty and filled positions
func (ste *DefaultStateExtractor) ExtractPositions(content string) ([]string, []string, error) {
	var emptyPositions, filledPositions []string
	
	// Extract empty positions
	emptyPattern := ste.patterns.GetPattern(PatternTypeEmptyPositions)
	emptyMatches := regexp.MustCompile(emptyPattern).FindAllStringSubmatch(content, -1)
	
	for _, match := range emptyMatches {
		if len(match) >= 2 {
			positions := ste.parsePositions(match[1])
			emptyPositions = append(emptyPositions, positions...)
		}
	}
	
	// Extract filled positions
	filledPattern := ste.patterns.GetPattern(PatternTypeFilledPositions)
	filledMatches := regexp.MustCompile(filledPattern).FindAllStringSubmatch(content, -1)
	
	for _, match := range filledMatches {
		if len(match) >= 2 {
			positions := ste.parsePositions(match[1])
			filledPositions = append(filledPositions, positions...)
		}
	}
	
	// Remove duplicates
	emptyPositions = removeDuplicates(emptyPositions)
	filledPositions = removeDuplicates(filledPositions)
	
	return emptyPositions, filledPositions, nil
}

// parseRowStatus determines the status of a row from text
func (ste *DefaultStateExtractor) parseRowStatus(text string) string {
	text = strings.ToLower(text)
	
	// Check status keywords
	statusKeywords := map[string]string{
		"full":     "Full",
		"complete": "Full",
		"filled":   "Full",
		"partial":  "Partial",
		"some":     "Partial",
		"few":      "Partial",
		"empty":    "Empty",
		"vacant":   "Empty",
		"no ":      "Empty",
	}
	
	for keyword, status := range statusKeywords {
		if strings.Contains(text, keyword) {
			return status
		}
	}
	
	// Heuristic rules
	if strings.Contains(text, "visible") && !strings.Contains(text, "coil") {
		return "Partial"
	} else if strings.Contains(text, "coil") {
		return "Empty"
	}
	
	return "Unknown"
}

// extractQuantity extracts quantity information from text
func (ste *DefaultStateExtractor) extractQuantity(text string) int {
	quantityPattern := ste.patterns.GetPattern(PatternTypeQuantity)
	matches := regexp.MustCompile(quantityPattern).FindStringSubmatch(text)
	
	if len(matches) >= 2 {
		quantity, err := strconv.Atoi(matches[1])
		if err == nil && quantity >= 0 {
			return quantity
		}
	}
	
	return 0
}

// parsePositions parses position identifiers from text
func (ste *DefaultStateExtractor) parsePositions(text string) []string {
	positionPattern := ste.patterns.GetPattern(PatternTypePosition)
	matches := regexp.MustCompile(positionPattern).FindAllString(text, -1)
	return matches
}

// FindObservations extracts general observations from content
func (of *DefaultObservationFinder) FindObservations(content string) ([]string, error) {
	var observations []string
	
	// Split content into sentences
	sentences := splitIntoSentences(content)
	
	// Filter sentences containing observation keywords
	observationKeywords := of.patterns.GetObservationKeywords()
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 10 { // Skip very short sentences
			continue
		}
		
		lowerSentence := strings.ToLower(sentence)
		for _, keyword := range observationKeywords {
			if strings.Contains(lowerSentence, keyword) {
				observations = append(observations, sentence)
				break
			}
		}
	}
	
	if len(observations) == 0 {
		return nil, errors.New("no observations found")
	}
	
	return observations, nil
}

// Helper functions

// splitIntoSentences splits text into sentences
func splitIntoSentences(text string) []string {
	// Simple sentence splitting
	sentences := []string{}
	
	// Split by common sentence terminators
	parts := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}
	
	return sentences
}

// normalizeHeaderToKey converts a section header to a normalized key
func normalizeHeaderToKey(header string) string {
	header = strings.TrimSpace(header)
	header = strings.ToLower(header)
	
	// Replace spaces with underscores
	header = regexp.MustCompile(`\s+`).ReplaceAllString(header, "_")
	
	// Remove special characters
	header = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(header, "")
	
	return header
}

// generateDefaultRowOrder generates default row identifiers (A, B, C, etc.)
func generateDefaultRowOrder(rowCount int) []string {
	rows := make([]string, rowCount)
	for i := 0; i < rowCount; i++ {
		rows[i] = string(rune('A' + i))
	}
	return rows
}

// generateDefaultColumnOrder generates default column identifiers (1, 2, 3, etc.)
func generateDefaultColumnOrder(colCount int) []string {
	cols := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		cols[i] = strconv.Itoa(i + 1)
	}
	return cols
}

// determineOverallStatus determines the overall status of the machine
func determineOverallStatus(filledCount, emptyCount int) string {
	if filledCount == 0 && emptyCount > 0 {
		return "Empty"
	} else if emptyCount == 0 && filledCount > 0 {
		return "Full"
	} else if filledCount > 0 && emptyCount > 0 {
		return "Partial"
	}
	
	return "Unknown"
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	unique := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}
	
	return unique
}