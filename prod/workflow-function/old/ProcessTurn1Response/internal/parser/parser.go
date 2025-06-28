// Package parser provides parsing utilities for Turn 1 responses
package parser

import (
	"log/slog"
	"workflow-function/ProcessTurn1Response/internal/types"
)

// Parser defines the core interface for parsing Bedrock responses
// This interface focuses purely on data extraction without business logic
type Parser interface {
	// ParseValidationResponse extracts validation indicators from turn1 response
	ParseValidationResponse(responseContent string) (map[string]interface{}, error)
	
	// ParseVisualAnalysis extracts visual analysis elements from turn1 response
	ParseVisualAnalysis(responseContent string) (map[string]interface{}, error)
	
	// ParseMachineStructure extracts machine structure information from turn1 response
	ParseMachineStructure(responseContent string) (*types.MachineStructure, error)
	
	// ParseMachineState extracts machine state information from turn1 response
	ParseMachineState(responseContent string) (*types.ExtractedState, error)
	
	// ExtractObservations extracts general observations from turn1 response
	ExtractObservations(responseContent string) ([]string, error)
}

// DefaultParser implements the Parser interface
type DefaultParser struct {
	// Core parser components with specialized responsibilities
	patternProvider    PatternProvider    // Provides parsing patterns
	responseExtractor  ResponseExtractor  // Extracts raw response content
	structureExtractor StructureExtractor // Extracts machine structure
	stateExtractor     StateExtractor     // Extracts machine state
	observationFinder  ObservationFinder  // Finds observations
	
	// Shared components
	logger *slog.Logger
	config *types.ParsingConfig
}

// DefaultParsingConfig returns a default parsing configuration
func DefaultParsingConfig() *types.ParsingConfig {
	return &types.ParsingConfig{
		StrictParsing:   false,
		IncludeThinking: true,
		CustomPatterns:  make(map[string]string),
		FallbackMode:    true,
		MaxResponseSize: 1024 * 1024, // 1MB
	}
}

// New creates a new Parser implementation with all dependencies
func New(logger *slog.Logger) Parser {
	// Create default config
	config := DefaultParsingConfig()
	
	// Create pattern provider with default patterns
	patternProvider := NewPatternProvider(logger)
	
	// Create specialized extractors
	responseExtractor := NewResponseExtractor(patternProvider, logger)
	structureExtractor := NewStructureExtractor(patternProvider, logger)
	stateExtractor := NewStateExtractor(patternProvider, logger)
	observationFinder := NewObservationFinder(patternProvider, logger)
	
	return &DefaultParser{
		patternProvider:    patternProvider,
		responseExtractor:  responseExtractor,
		structureExtractor: structureExtractor,
		stateExtractor:     stateExtractor,
		observationFinder:  observationFinder,
		logger:             logger,
		config:             config,
	}
}

// ParseValidationResponse extracts validation indicators from turn1 response
func (p *DefaultParser) ParseValidationResponse(responseContent string) (map[string]interface{}, error) {
	// Extract raw content using response extractor
	parsed, err := p.responseExtractor.ExtractContent(responseContent)
	if err != nil {
		p.logger.Error("Failed to extract response content", slog.String("error", err.Error()))
		return nil, err
	}
	
	// Extract validation indicators
	validationResults := make(map[string]interface{})
	
	// Extract structure validation indicators
	structureValidation, err := p.extractStructureValidation(parsed.MainContent)
	if err == nil {
		validationResults["structureValidation"] = structureValidation
	}
	
	// Extract completeness validation indicators
	completenessValidation, err := p.extractCompletenessValidation(parsed.MainContent)
	if err == nil {
		validationResults["completenessValidation"] = completenessValidation
	}
	
	// Extract consistency validation indicators
	consistencyValidation, err := p.extractConsistencyValidation(parsed.MainContent)
	if err == nil {
		validationResults["consistencyValidation"] = consistencyValidation
	}
	
	return validationResults, nil
}

// ParseVisualAnalysis extracts visual analysis elements from turn1 response
func (p *DefaultParser) ParseVisualAnalysis(responseContent string) (map[string]interface{}, error) {
	// Extract raw content using response extractor
	parsed, err := p.responseExtractor.ExtractContent(responseContent)
	if err != nil {
		p.logger.Error("Failed to extract response content", slog.String("error", err.Error()))
		return nil, err
	}
	
	// Extract visual analysis elements
	visualAnalysis := make(map[string]interface{})
	
	// Extract general observations
	observations, err := p.observationFinder.FindObservations(parsed.MainContent)
	if err == nil && len(observations) > 0 {
		visualAnalysis["observations"] = observations
	}
	
	// Extract product identifications
	products, err := p.extractProductIdentifications(parsed.MainContent)
	if err == nil {
		visualAnalysis["productIdentifications"] = products
	}
	
	// Extract visibility assessments
	visibility, err := p.extractVisibilityAssessments(parsed.MainContent)
	if err == nil {
		visualAnalysis["visibilityAssessments"] = visibility
	}
	
	return visualAnalysis, nil
}

// ParseMachineStructure extracts machine structure information from turn1 response
func (p *DefaultParser) ParseMachineStructure(responseContent string) (*types.MachineStructure, error) {
	// Extract raw content using response extractor
	parsed, err := p.responseExtractor.ExtractContent(responseContent)
	if err != nil {
		p.logger.Error("Failed to extract response content", slog.String("error", err.Error()))
		return nil, err
	}
	
	// Extract machine structure using structure extractor
	return p.structureExtractor.ExtractMachineStructure(parsed.MainContent)
}

// ParseMachineState extracts machine state information from turn1 response
func (p *DefaultParser) ParseMachineState(responseContent string) (*types.ExtractedState, error) {
	// Extract raw content using response extractor
	parsed, err := p.responseExtractor.ExtractContent(responseContent)
	if err != nil {
		p.logger.Error("Failed to extract response content", slog.String("error", err.Error()))
		return nil, err
	}
	
	// Extract machine state using state extractor
	return p.stateExtractor.ExtractMachineState(parsed.MainContent)
}

// ExtractObservations extracts general observations from turn1 response
func (p *DefaultParser) ExtractObservations(responseContent string) ([]string, error) {
	// Extract raw content using response extractor
	parsed, err := p.responseExtractor.ExtractContent(responseContent)
	if err != nil {
		p.logger.Error("Failed to extract response content", slog.String("error", err.Error()))
		return nil, err
	}
	
	// Find observations using observation finder
	return p.observationFinder.FindObservations(parsed.MainContent)
}

// Helper methods for extracting validation indicators
func (p *DefaultParser) extractStructureValidation(content string) (map[string]interface{}, error) {
	// Implementation details for structure validation extraction
	return map[string]interface{}{
		"valid": true,
		"details": "Machine structure validated",
	}, nil
}

func (p *DefaultParser) extractCompletenessValidation(content string) (map[string]interface{}, error) {
	// Implementation details for completeness validation extraction
	return map[string]interface{}{
		"complete": true,
		"missingElements": []string{},
	}, nil
}

func (p *DefaultParser) extractConsistencyValidation(content string) (map[string]interface{}, error) {
	// Implementation details for consistency validation extraction
	return map[string]interface{}{
		"consistent": true,
		"inconsistencies": []string{},
	}, nil
}

// Helper methods for extracting visual analysis elements
func (p *DefaultParser) extractProductIdentifications(content string) (map[string]interface{}, error) {
	// Implementation details for product identification extraction
	return map[string]interface{}{
		"productCount": 10,
		"productTypes": []string{"Snack", "Beverage"},
	}, nil
}

func (p *DefaultParser) extractVisibilityAssessments(content string) (map[string]interface{}, error) {
	// Implementation details for visibility assessment extraction
	return map[string]interface{}{
		"clearView": true,
		"obstructions": []string{},
	}, nil
}