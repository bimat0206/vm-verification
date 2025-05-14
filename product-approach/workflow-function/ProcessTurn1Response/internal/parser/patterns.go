package parser

import (
	"encoding/json"
	"fmt"
	"workflow-function/shared/logger"
	"workflow-function/product-approach/workflow-function/ProcessTurn1Response/internal/types"
)

// ParsingPatterns contains configurable patterns for parsing
type ParsingPatterns struct {
	ContentFields       []string                      `json:"contentFields"`
	ThinkingFields      []string                      `json:"thinkingFields"`
	StructuredMarkers   []string                      `json:"structuredMarkers"`
	SectionPatterns     map[string]*types.ResponsePattern `json:"sectionPatterns"`
	MachineStructure    []string                      `json:"machineStructurePatterns"`
	RowStatus           []string                      `json:"rowStatusPatterns"`
	EmptyPositions      []string                      `json:"emptyPositionPatterns"`
	Observations        []string                      `json:"observationKeywords"`
	QuantityPatterns    []string                      `json:"quantityPatterns"`
	PositionPattern     string                        `json:"positionPattern"`
	RowPattern          string                        `json:"rowPattern"`
	ColumnPattern       string                        `json:"columnPattern"`
	MinContentLength    int                           `json:"minContentLength"`
	MaxSectionLength    int                           `json:"maxSectionLength"`
}

// loadParsingPatterns loads parsing patterns from configuration or context
func loadParsingPatterns(context *types.ParsingContext) (*ParsingPatterns, error) {
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
		SectionPatterns:   make(map[string]*types.ResponsePattern),
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
func getPatternsFromConfig(config *types.ProcessingConfig) *ParsingPatterns {
	patterns := getDefaultParsingPatterns()
	
	// Adjust patterns based on configuration
	if config.ExtractMachineStructure {
		patterns.SectionPatterns["machine_structure"] = &types.ResponsePattern{
			Name:     "Machine Structure",
			Pattern:  `(?s)(\d+)\s+.*?(\d+)`,
			Required: false,
		}
	}
	
	if config.ValidateCompleteness {
		patterns.SectionPatterns["completeness_check"] = &types.ResponsePattern{
			Name:     "Completeness Check",
			Pattern:  `(?i)(complete|incomplete|partial)`,
			Required: false,
		}
	}
	
	return patterns
}

// mapToStruct converts a map to a struct
func mapToStruct(data map[string]interface{}, target interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you might use reflection or a JSON library
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}