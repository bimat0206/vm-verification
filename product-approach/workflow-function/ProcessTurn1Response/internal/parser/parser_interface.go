package parser

import (
	"workflow-function/shared/logger"
	"workflow-function/product-approach/workflow-function/ProcessTurn1Response/internal/types"
)

// Parser interface defines the methods needed by the Processor
type Parser interface {
	// ParseValidationResponse parses validation indicators from response
	ParseValidationResponse(response map[string]interface{}) map[string]interface{}
	
	// ParseVisualAnalysis parses visual analysis from response
	ParseVisualAnalysis(response map[string]interface{}) map[string]interface{}
	
	// ParseMachineStructure parses machine structure from response
	ParseMachineStructure(response map[string]interface{}) map[string]interface{}
	
	// ParseMachineState parses machine state from response
	ParseMachineState(response map[string]interface{}) map[string]interface{}
	
	// ExtractObservations extracts observations from response
	ExtractObservations(response map[string]interface{}) map[string]interface{}
}

// ParserImpl implements the Parser interface using ResponseParser
type ParserImpl struct {
	responseParser *ResponseParser
	logger         logger.Logger
}

// NewParser creates a new Parser implementation
func NewParser(log logger.Logger) Parser {
	// Create default processing config and context
	config := types.DefaultProcessingConfig()
	context := &types.ParsingContext{
		VerificationType:    "",
		HasHistoricalContext: false,
		ParsingConfig:       config,
	}
	
	responseParser := NewResponseParser(config, context, log)
	
	impl := &ParserImpl{
		responseParser: responseParser,
		logger:         log,
	}
	
	// Return the interface, not the implementation
	var parser Parser = impl
	return parser
}