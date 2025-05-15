package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"workflow-function/shared/logger"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// ResponseParser handles parsing of Bedrock responses
type ResponseParser struct {
	config   *types.ProcessingConfig
	context  *types.ParsingContext
	logger   logger.Logger
	patterns *ParsingPatterns
}

// NewResponseParser creates a new response parser
func NewResponseParser(config *types.ProcessingConfig, context *types.ParsingContext, log logger.Logger) *ResponseParser {
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

// ParseResponse parses a raw Bedrock response into structured data
func (p *ResponseParser) ParseResponse(turn1Response map[string]interface{}) (*types.ParsedResponse, error) {
	startTime := time.Now()
	
	p.logger.Debug("Starting response parsing", map[string]interface{}{
		"responseKeys": p.getMapKeys(turn1Response),
		"hasConfig":    p.config != nil,
		"hasContext":   p.context != nil,
	})
	
	parsed := &types.ParsedResponse{
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
func (p *ResponseParser) parseStructuredResponse(parsed *types.ParsedResponse) error {
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
func (p *ResponseParser) parseUnstructuredResponse(parsed *types.ParsedResponse) error {
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

// extractWithPattern extracts text using a regex pattern
func (p *ResponseParser) extractWithPattern(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		return matches[1:]
	}
	
	return []string{}
}

// extractStructuredData extracts structured data from parsed sections
func (p *ResponseParser) extractStructuredData(parsed *types.ParsedResponse) {
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

// getMapKeys gets keys from a map for logging
func (p *ResponseParser) getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
