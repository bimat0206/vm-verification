// Package parser provides parsing utilities for Turn 1 responses
package parser

import (
	"encoding/json"
	"errors"
	"log/slog"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// PatternType defines types of parsing patterns
type PatternType string

const (
	// Pattern types for structured data extraction
	PatternTypeMachineStructure  PatternType = "machine_structure"
	PatternTypeRowStatus         PatternType = "row_status"
	PatternTypeEmptyPositions    PatternType = "empty_positions"
	PatternTypeFilledPositions   PatternType = "filled_positions"
	PatternTypeObservations      PatternType = "observations"
	PatternTypeQuantity          PatternType = "quantity"
	PatternTypePosition          PatternType = "position"
	PatternTypeRow               PatternType = "row"
	PatternTypeColumn            PatternType = "column"
	PatternTypeRowColumn         PatternType = "row_column"
	PatternTypeFallbackStructure PatternType = "fallback_structure"
	
	// Patterns for validation extraction
	PatternTypeStructureValidation    PatternType = "structure_validation"
	PatternTypeCompletenessValidation PatternType = "completeness_validation"
	PatternTypeConsistencyValidation  PatternType = "consistency_validation"
	
	// Patterns for visual analysis
	PatternTypeProductIdentification PatternType = "product_identification"
	PatternTypeVisibilityAssessment  PatternType = "visibility_assessment"
)

// PatternProvider defines the interface for accessing parsing patterns
type PatternProvider interface {
	// GetPattern retrieves a specific pattern by type
	GetPattern(patternType PatternType) string
	
	// GetSectionPatterns retrieves patterns for content sections
	GetSectionPatterns() map[string]string
	
	// GetContentFields retrieves field names for main content
	GetContentFields() []string
	
	// GetThinkingFields retrieves field names for thinking content
	GetThinkingFields() []string
	
	// GetObservationKeywords retrieves keywords for observations
	GetObservationKeywords() []string
	
	// GetMinContentLength retrieves minimum length for content
	GetMinContentLength() int
	
	// GetMaxSectionLength retrieves maximum length for a section
	GetMaxSectionLength() int
	
	// RegisterCustomPattern registers a custom pattern
	RegisterCustomPattern(patternType PatternType, pattern string) error
}

// DefaultPatternProvider implements PatternProvider
type DefaultPatternProvider struct {
	// Core pattern sets
	patterns          map[PatternType]string
	sectionPatterns   map[string]string
	contentFields     []string
	thinkingFields    []string
	observationKeywords []string
	
	// Configuration values
	minContentLength int
	maxSectionLength int
	
	// Logger for debugging and warning
	logger *slog.Logger
}

// NewPatternProvider creates a new DefaultPatternProvider with default patterns
func NewPatternProvider(logger *slog.Logger) PatternProvider {
	provider := &DefaultPatternProvider{
		patterns:          make(map[PatternType]string),
		sectionPatterns:   make(map[string]string),
		contentFields:     []string{},
		thinkingFields:    []string{},
		observationKeywords: []string{},
		minContentLength:  50,
		maxSectionLength:  10000,
		logger:            logger,
	}
	
	// Initialize with default patterns
	provider.initializeDefaultPatterns()
	
	return provider
}

// GetPattern retrieves a specific pattern by type
func (p *DefaultPatternProvider) GetPattern(patternType PatternType) string {
	if pattern, exists := p.patterns[patternType]; exists {
		return pattern
	}
	
	p.logger.Warn("Pattern not found, using fallback", slog.String("patternType", string(patternType)))
	return getDefaultPattern(patternType)
}

// GetSectionPatterns retrieves patterns for content sections
func (p *DefaultPatternProvider) GetSectionPatterns() map[string]string {
	return p.sectionPatterns
}

// GetContentFields retrieves field names for main content
func (p *DefaultPatternProvider) GetContentFields() []string {
	return p.contentFields
}

// GetThinkingFields retrieves field names for thinking content
func (p *DefaultPatternProvider) GetThinkingFields() []string {
	return p.thinkingFields
}

// GetObservationKeywords retrieves keywords for observations
func (p *DefaultPatternProvider) GetObservationKeywords() []string {
	return p.observationKeywords
}

// GetMinContentLength retrieves minimum length for content
func (p *DefaultPatternProvider) GetMinContentLength() int {
	return p.minContentLength
}

// GetMaxSectionLength retrieves maximum length for a section
func (p *DefaultPatternProvider) GetMaxSectionLength() int {
	return p.maxSectionLength
}

// RegisterCustomPattern registers a custom pattern
func (p *DefaultPatternProvider) RegisterCustomPattern(patternType PatternType, pattern string) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}
	
	p.patterns[patternType] = pattern
	p.logger.Debug("Registered custom pattern", 
		slog.String("patternType", string(patternType)),
		slog.String("pattern", pattern))
	
	return nil
}

// initializeDefaultPatterns sets up the default pattern set
func (p *DefaultPatternProvider) initializeDefaultPatterns() {
	// Initialize pattern maps
	p.patterns = map[PatternType]string{
		// Core structure patterns
		PatternTypeMachineStructure:  `(?is)(machine|vending)\s+structure.*?(\d+)`,
		PatternTypeRowStatus:         `(?m)^## Row ([A-Z])(?:[^*]+)\*\*Status: ([A-Za-z]+)\*\*`,
		PatternTypeEmptyPositions:    `(?i)empty\s+positions?[:\s]+([^.]+)`,
		PatternTypeFilledPositions:   `(?i)filled\s+positions?[:\s]+([^.]+)`,
		PatternTypeQuantity:          `(\d+)`,
		PatternTypePosition:          `(?m)- ([A-Z]\d+): ([^\n]+)`,
		PatternTypeRow:               `[A-Z]`,
		PatternTypeColumn:            `\d+`,
		PatternTypeRowColumn:         `(?i)examining each row from top to bottom \(([A-Z])[^)]+\) and documenting the contents of all (\d+) slots`,
		PatternTypeFallbackStructure: `(?i)(\d+)[^.]*?(\d+)`,
		
		// Validation patterns
		PatternTypeStructureValidation:    `(?i)(structure|layout)[^.]*?(valid|match|confirm)`,
		PatternTypeCompletenessValidation: `(?i)(complete|full|missing|partial)`,
		PatternTypeConsistencyValidation:  `(?i)(consistent|inconsistent|discrepanc)`,
		
		// Visual analysis patterns
		PatternTypeProductIdentification: `(?i)(product|item)[^.]*?(visible|identif|recogn)`,
		PatternTypeVisibilityAssessment:  `(?i)(visib|clear|obstruct)`,
	}
	
	// Section patterns for structured parsing
	p.sectionPatterns = map[string]string{
		"machine_structure": `(?s)(?:#{1,3}\s+)?(?:Machine|Vending)[^#]+`,
		"row_status":        `(?s)(?:#{1,3}\s+)?(?:Row|Status)[^#]+`,
		"empty_positions":   `(?s)(?:#{1,3}\s+)?(?:Empty|Vacant)[^#]+`,
		"observations":      `(?s)(?:#{1,3}\s+)?(?:Observation|Note|Comment)[^#]+`,
	}
	
	// Common field names for content extraction
	p.contentFields = []string{
		"content", "response", "text", "output", "result", "body", "message",
	}
	
	// Common field names for thinking content
	p.thinkingFields = []string{
		"thinking", "reasoning", "analysis", "thought_process", "rationale",
	}
	
	// Keywords that indicate observations
	p.observationKeywords = []string{
		"visible", "appears", "note", "observation", "noticed", "see", "can see",
		"apparent", "evident", "clear", "present", "absent", "missing",
	}
	
	// Configuration values
	p.minContentLength = 50
	p.maxSectionLength = 10000
}

// getDefaultPattern provides a fallback pattern for a given type
func getDefaultPattern(patternType PatternType) string {
	switch patternType {
	case PatternTypeMachineStructure:
		return `(?i)(\d+)\s+.*?(\d+)`
	case PatternTypeRowStatus:
		return `(?i)([A-Z]+)[:\s]+([^.]+\.)`
	case PatternTypeEmptyPositions:
		return `([A-Z]\d+)`
	case PatternTypeFilledPositions:
		return `([A-Z]\d+)`
	case PatternTypeQuantity:
		return `(\d+)`
	case PatternTypePosition:
		return `[A-Z]\d+`
	case PatternTypeRow:
		return `[A-Z]`
	case PatternTypeColumn:
		return `\d+`
	case PatternTypeRowColumn:
		return `(?i)(\d+)\s*rows?[^.]*?(\d+)\s*columns?`
	case PatternTypeFallbackStructure:
		return `(?i)(\d+)[^.]*?(\d+)`
	default:
		return `(.+)`
	}
}

// LoadPatternsFromConfiguration loads patterns from configuration
func LoadPatternsFromConfiguration(config *types.ParsingContext, logger *slog.Logger) PatternProvider {
	provider := NewPatternProvider(logger)
	
	// Try to load from layout metadata if available
	if config.LayoutMetadata != nil {
		if patternsData, exists := config.LayoutMetadata["parsingPatterns"]; exists {
			if err := applyPatternsFromMap(provider, patternsData); err != nil {
				logger.Warn("Failed to apply patterns from layout metadata", 
					slog.String("error", err.Error()))
			}
		}
	}
	
	// Try to load from historical context if available
	if config.HistoricalContext != nil {
		if patternsData, exists := config.HistoricalContext["parsingPatterns"]; exists {
			if err := applyPatternsFromMap(provider, patternsData); err != nil {
				logger.Warn("Failed to apply patterns from historical context", 
					slog.String("error", err.Error()))
			}
		}
	}
	
	return provider
}

// applyPatternsFromMap applies patterns from a map to the provider
func applyPatternsFromMap(provider PatternProvider, patternsData interface{}) error {
	// Convert to map[string]interface{}
	patternsMap, ok := patternsData.(map[string]interface{})
	if !ok {
		return errors.New("patterns data is not a map")
	}
	
	// Marshal and unmarshal to convert cleanly
	data, err := json.Marshal(patternsMap)
	if err != nil {
		return err
	}
	
	// Define pattern mapping struct
	type PatternMapping struct {
		MachineStructure       string   `json:"machineStructurePattern"`
		RowStatus              string   `json:"rowStatusPattern"`
		EmptyPositions         string   `json:"emptyPositionsPattern"`
		FilledPositions        string   `json:"filledPositionsPattern"`
		RowColumn              string   `json:"rowColumnPattern"`
		Position               string   `json:"positionPattern"`
		Row                    string   `json:"rowPattern"`
		Column                 string   `json:"columnPattern"`
		StructureValidation    string   `json:"structureValidationPattern"`
		CompletenessValidation string   `json:"completenessValidationPattern"`
		ConsistencyValidation  string   `json:"consistencyValidationPattern"`
		ObservationKeywords    []string `json:"observationKeywords"`
	}
	
	var mapping PatternMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return err
	}
	
	// Apply patterns from mapping
	if mapping.MachineStructure != "" {
		provider.RegisterCustomPattern(PatternTypeMachineStructure, mapping.MachineStructure)
	}
	
	if mapping.RowStatus != "" {
		provider.RegisterCustomPattern(PatternTypeRowStatus, mapping.RowStatus)
	}
	
	if mapping.EmptyPositions != "" {
		provider.RegisterCustomPattern(PatternTypeEmptyPositions, mapping.EmptyPositions)
	}
	
	if mapping.FilledPositions != "" {
		provider.RegisterCustomPattern(PatternTypeFilledPositions, mapping.FilledPositions)
	}
	
	if mapping.RowColumn != "" {
		provider.RegisterCustomPattern(PatternTypeRowColumn, mapping.RowColumn)
	}
	
	if mapping.Position != "" {
		provider.RegisterCustomPattern(PatternTypePosition, mapping.Position)
	}
	
	if mapping.Row != "" {
		provider.RegisterCustomPattern(PatternTypeRow, mapping.Row)
	}
	
	if mapping.Column != "" {
		provider.RegisterCustomPattern(PatternTypeColumn, mapping.Column)
	}
	
	return nil
}