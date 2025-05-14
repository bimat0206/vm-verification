package processor

import (
	"context"
	"fmt"
	"strings"

	"product-approach/workflow-function/shared/logger"
	"product-approach/workflow-function/shared/schema"
)

// ProcessingResult represents the result of Turn 1 response processing
type ProcessingResult struct {
	// SourceType indicates the source of the analysis
	SourceType string
	
	// AnalysisData contains the extracted analysis data
	AnalysisData map[string]interface{}
	
	// ContextForTurn2 contains prepared context for Turn 2
	ContextForTurn2 map[string]interface{}
	
	// ProcessingPath indicates which processing path was used
	ProcessingPath string
}

// Dependencies is an alias for dependencies.Dependencies
type Dependencies struct {
	// Placeholder for dependencies
	Logger logger.Logger
}

// Parser is an alias for parser.Parser
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

// Validator handles validation of processing results
type Validator struct {
	logger logger.Logger
}

// ValidateReferenceAnalysis validates the reference analysis
func (v *Validator) ValidateReferenceAnalysis(analysis map[string]interface{}) error {
	// This is a placeholder implementation
	if analysis == nil {
		return fmt.Errorf("analysis is nil")
	}
	return nil
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(log logger.Logger) *Dependencies {
	return &Dependencies{
		Logger: log,
	}
}

// NewParser creates a new Parser instance
func NewParser(log logger.Logger) Parser {
	// This is a placeholder implementation
	return nil
}

// NewValidator creates a new Validator instance
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		logger: log,
	}
}

// Processor handles the core processing logic for Turn 1 responses
type Processor struct {
	log    logger.Logger
	deps   *Dependencies
	parser Parser
}

// NewProcessor creates a new processor instance
func NewProcessor(log logger.Logger, deps *Dependencies) *Processor {
	return &Processor{
		log:    log,
		deps:   deps,
		parser: NewParser(log),
	}
}

// ProcessTurn1Response processes the Turn 1 response based on verification type
func (p *Processor) ProcessTurn1Response(ctx context.Context, input schema.WorkflowState) (schema.WorkflowState, error) {
	log := p.log.WithFields(map[string]interface{}{
		"verificationType": input.VerificationContext.VerificationType,
		"verificationId":   input.VerificationContext.VerificationId,
	})

	log.Info("Starting Turn 1 response processing", nil)

	// Determine processing path
	processingPath := p.determineProcessingPath(input)
	log.Info("Processing path determined", map[string]interface{}{
		"path": processingPath,
	})

	// Route to appropriate processing method
	var result *ProcessingResult
	var err error

	switch processingPath {
	case "UC1_VALIDATION":
		result, err = p.processValidationFlow(input)
	case "UC2_HISTORICAL_ENHANCEMENT":
		result, err = p.processHistoricalEnhancement(input)
	case "UC2_FRESH_EXTRACTION":
		result, err = p.processFreshExtraction(input)
	default:
		return input, fmt.Errorf("unknown processing path: %s", processingPath)
	}

	if err != nil {
		log.Error("Processing failed", map[string]interface{}{
			"error": err.Error(),
			"path":  processingPath,
		})
		return input, err
	}

	// Build output workflow state
	output := p.buildOutputState(input, result)
	
	log.Info("Turn 1 response processing completed", map[string]interface{}{
		"sourceType":    result.SourceType,
		"analysisItems": len(result.AnalysisData),
	})

	return output, nil
}

// determineProcessingPath determines which processing path to take
func (p *Processor) determineProcessingPath(input schema.WorkflowState) string {
	verificationType := input.VerificationContext.VerificationType
	
	// Check verification type constants from schema
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		return "UC1_VALIDATION"
	}
	
	if verificationType == schema.VerificationTypePreviousVsCurrent {
		// Check if historical data is available
		if input.HistoricalContext != nil && len(input.HistoricalContext) > 0 {
			return "UC2_HISTORICAL_ENHANCEMENT"
		}
		return "UC2_FRESH_EXTRACTION"
	}
	
	return "UNKNOWN"
}

// processValidationFlow handles UC1 validation flow
func (p *Processor) processValidationFlow(input schema.WorkflowState) (*ProcessingResult, error) {
	response := input.Turn1Response
	
	// Parse the response for validation indicators
	validationResults := p.parser.ParseValidationResponse(response)
	
	// Extract basic observations
	observations := p.parser.ExtractObservations(response)
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":            "VALIDATION_COMPLETE",
		"sourceType":        "REFERENCE_VALIDATION",
		"validationResults": validationResults,
		"basicObservations": observations,
		"contextForTurn2": map[string]interface{}{
			"referenceValidated":    true,
			"useSystemPromptReference": true,
			"validationPassed":     validationResults["structureConfirmed"],
			"readyForTurn2":        true,
		},
	}
	
	return &ProcessingResult{
		SourceType:      "REFERENCE_VALIDATION",
		AnalysisData:    analysisData,
		ContextForTurn2: analysisData["contextForTurn2"].(map[string]interface{}),
		ProcessingPath:  "UC1_VALIDATION",
	}, nil
}

// processHistoricalEnhancement handles UC2 with historical context
func (p *Processor) processHistoricalEnhancement(input schema.WorkflowState) (*ProcessingResult, error) {
	response := input.Turn1Response
	historicalContext := input.HistoricalContext
	
	// Parse visual confirmation from response
	visualAnalysis := p.parser.ParseVisualAnalysis(response)
	
	// Extract and enhance historical baseline
	enhancedBaseline := p.enhanceHistoricalBaseline(historicalContext, visualAnalysis)
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":             "EXTRACTION_COMPLETE",
		"sourceType":         "HISTORICAL_WITH_VISUAL_CONFIRMATION",
		"historicalBaseline": enhancedBaseline,
		"machineStructure":   p.extractMachineStructure(historicalContext),
		"knownIssues":        p.extractKnownIssues(historicalContext),
		"contextForTurn2": map[string]interface{}{
			"baselineEstablished":    true,
			"useHistoricalAsReference": true,
			"focusAreas":            p.identifyFocusAreas(historicalContext),
			"knownIssues":           p.extractKnownIssuesList(historicalContext),
		},
	}
	
	return &ProcessingResult{
		SourceType:      "HISTORICAL_WITH_VISUAL_CONFIRMATION",
		AnalysisData:    analysisData,
		ContextForTurn2: analysisData["contextForTurn2"].(map[string]interface{}),
		ProcessingPath:  "UC2_HISTORICAL_ENHANCEMENT",
	}, nil
}

// processFreshExtraction handles UC2 without historical context
func (p *Processor) processFreshExtraction(input schema.WorkflowState) (*ProcessingResult, error) {
	response := input.Turn1Response
	
	// Parse comprehensive analysis from response
	extractedStructure := p.parser.ParseMachineStructure(response)
	extractedState := p.parser.ParseMachineState(response)
	
	// Create analysis data
	analysisData := map[string]interface{}{
		"status":              "EXTRACTION_COMPLETE",
		"sourceType":          "FRESH_VISUAL_ANALYSIS",
		"extractedStructure":  extractedStructure,
		"extractedState":     extractedState,
		"contextForTurn2": map[string]interface{}{
			"baselineSource":        "EXTRACTED_STATE",
			"useHistoricalData":     false,
			"extractedDataAvailable": true,
			"readyForTurn2":         true,
		},
	}
	
	return &ProcessingResult{
		SourceType:      "FRESH_VISUAL_ANALYSIS",
		AnalysisData:    analysisData,
		ContextForTurn2: analysisData["contextForTurn2"].(map[string]interface{}),
		ProcessingPath:  "UC2_FRESH_EXTRACTION",
	}, nil
}

// enhanceHistoricalBaseline combines historical data with visual analysis
func (p *Processor) enhanceHistoricalBaseline(historical map[string]interface{}, visual map[string]interface{}) map[string]interface{} {
	enhanced := make(map[string]interface{})
	
	// Copy historical data
	for key, value := range historical {
		enhanced[key] = value
	}
	
	// Add visual confirmations
	if visual != nil {
		enhanced["visualConfirmation"] = visual
		enhanced["enhancementTimestamp"] = schema.FormatISO8601()
	}
	
	return enhanced
}

// extractMachineStructure extracts machine structure from context
func (p *Processor) extractMachineStructure(context map[string]interface{}) map[string]interface{} {
	if structure, ok := context["machineStructure"].(map[string]interface{}); ok {
		return structure
	}
	
	// Return default structure if not found
	return map[string]interface{}{
		"extracted": false,
		"fallback":  true,
	}
}

// extractKnownIssues extracts known issues from historical context
func (p *Processor) extractKnownIssues(context map[string]interface{}) map[string]interface{} {
	issues := make(map[string]interface{})
	
	// Extract verification summary if available
	if summary, ok := context["verificationSummary"].(map[string]interface{}); ok {
		// Convert summary fields to known issues format
		for key, value := range summary {
			if strings.Contains(strings.ToLower(key), "empty") ||
			   strings.Contains(strings.ToLower(key), "incorrect") ||
			   strings.Contains(strings.ToLower(key), "missing") ||
			   strings.Contains(strings.ToLower(key), "discrepant") {
				issues[key] = value
			}
		}
	}
	
	return issues
}

// identifyFocusAreas identifies areas that need special attention in Turn 2
func (p *Processor) identifyFocusAreas(context map[string]interface{}) []string {
	focusAreas := []string{}
	
	// Extract areas with known issues
	if checkingStatus, ok := context["checkingStatus"].(map[string]interface{}); ok {
		for key, value := range checkingStatus {
			if status, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(status), "empty") ||
				   strings.Contains(strings.ToLower(status), "incorrect") ||
				   strings.Contains(strings.ToLower(status), "changed") {
					focusAreas = append(focusAreas, key)
				}
			}
		}
	}
	
	return focusAreas
}

// extractKnownIssuesList extracts a list of known issue types
func (p *Processor) extractKnownIssuesList(context map[string]interface{}) []string {
	issues := []string{}
	
	if summary, ok := context["verificationSummary"].(map[string]interface{}); ok {
		if status, ok := summary["verificationStatus"].(string); ok && status != "CORRECT" {
			// Parse outcome for issue types
			if outcome, ok := summary["verificationOutcome"].(string); ok {
				if strings.Contains(strings.ToLower(outcome), "empty") {
					issues = append(issues, "empty_rows")
				}
				if strings.Contains(strings.ToLower(outcome), "incorrect") {
					issues = append(issues, "incorrect_products")
				}
			}
		}
	}
	
	return issues
}

// buildOutputState constructs the output workflow state
func (p *Processor) buildOutputState(input schema.WorkflowState, result *ProcessingResult) schema.WorkflowState {
	output := input
	
	// Set reference analysis
	output.ReferenceAnalysis = result.AnalysisData
	
	// Update schema version
	output.SchemaVersion = schema.SchemaVersion
	
	// Add processing metadata
	if output.ReferenceAnalysis != nil {
		output.ReferenceAnalysis["processingMetadata"] = map[string]interface{}{
			"processingPath":   result.ProcessingPath,
			"processedAt":      schema.FormatISO8601(),
			"sourceType":       result.SourceType,
			"contextPrepared":  true,
		}
	}
	
	return output
}
