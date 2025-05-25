// Package processor provides Turn1 processing capabilities
package processor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"workflow-function/ProcessTurn1Response/internal/parser"
	"workflow-function/ProcessTurn1Response/internal/types"
	"workflow-function/ProcessTurn1Response/internal/validator"
)

// Processor interface defines the contract for processing Turn1 responses
type Processor interface {
	// ProcessTurn1Response processes the Turn1 response based on the processing path
	ProcessTurn1Response(ctx context.Context, responseContent string, processingPath types.ProcessingPath, historicalData *types.HistoricalEnhancement) (*types.Turn1ProcessingResult, error)
}

// DefaultProcessor implements the Processor interface
type DefaultProcessor struct {
	logger     *slog.Logger
	parser     parser.Parser
	validator  validator.ValidatorInterface
}

// New creates a new processor instance
func New(logger *slog.Logger) Processor {
	// Adapt slog logger to shared logger interface for validator
	logAdapter := NewSlogLoggerAdapter(logger)
	
	return &DefaultProcessor{
		logger:    logger,
		parser:    parser.New(logger),
		validator: validator.NewValidator(logAdapter),
	}
}

// ProcessTurn1Response processes the Turn1 response based on the processing path
func (p *DefaultProcessor) ProcessTurn1Response(
	ctx context.Context,
	responseContent string,
	processingPath types.ProcessingPath,
	historicalData *types.HistoricalEnhancement,
) (*types.Turn1ProcessingResult, error) {
	startTime := TimeNow()
	
	p.logger.Info("Starting Turn1 response processing", 
		"processingPath", processingPath,
		"contentLength", len(responseContent),
		"hasHistoricalData", historicalData != nil,
	)

	// Create processing metadata
	metadata := &types.ProcessingMetadata{
		ProcessingStartTime: startTime,
		ProcessingPath:      processingPath,
		ResponseSize:        int64(len(responseContent)),
	}

	// Initialize result with common fields
	result := &types.Turn1ProcessingResult{
		Status:             "PROCESSING",
		SourceType:         processingPath,
		ProcessingMetadata: metadata,
	}

	// Select the appropriate processing path handler
	var processor PathProcessor
	var err error

	switch processingPath {
	case types.PathValidationFlow:
		processor = NewValidationFlowProcessor(p.logger, p.parser, p.validator)
	case types.PathHistoricalEnhancement:
		processor = NewHistoricalEnhancementProcessor(p.logger, p.parser, p.validator)
	case types.PathFreshExtraction:
		processor = NewFreshExtractionProcessor(p.logger, p.parser, p.validator)
	default:
		return nil, NewProcessingError("Unknown processing path", processingPath.String())
	}

	// Process using the selected handler
	err = processor.Process(ctx, responseContent, historicalData, result)
	if err != nil {
		return nil, err
	}

	// Set completion time and duration
	metadata.ProcessingEndTime = TimeNow()
	metadata.ProcessingDuration = metadata.ProcessingEndTime.Sub(metadata.ProcessingStartTime)
	
	// Update status based on success
	result.Status = "EXTRACTION_COMPLETE"
	
	p.logger.Info("Turn1 response processing completed", 
		"duration", metadata.ProcessingDuration.String(),
		"extractedElements", metadata.ExtractedElements,
	)

	return result, nil
}

// TimeFormat is the standard time format for timestamps
const TimeFormat = time.RFC3339

// TimeNow returns the current time, can be mocked for testing
var TimeNow = func() time.Time {
	return time.Now()
}

// FormatTime formats a time according to the standard format
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// ProcessorOption is a function for configuring a processor
type ProcessorOption func(*DefaultProcessor)

// WithParser sets a custom parser for the processor
func WithParser(p parser.Parser) ProcessorOption {
	return func(dp *DefaultProcessor) {
		dp.parser = p
	}
}

// WithValidator sets a custom validator for the processor
func WithValidator(v validator.ValidatorInterface) ProcessorOption {
	return func(dp *DefaultProcessor) {
		dp.validator = v
	}
}

// NewWithOptions creates a new processor with options
func NewWithOptions(logger *slog.Logger, options ...ProcessorOption) Processor {
	// Adapt slog logger to shared logger interface for validator
	logAdapter := NewSlogLoggerAdapter(logger)
	
	proc := &DefaultProcessor{
		logger:    logger,
		parser:    parser.New(logger),
		validator: validator.NewValidator(logAdapter),
	}
	
	for _, option := range options {
		option(proc)
	}
	
	return proc
}

// Error types

// ProcessingError represents an error during Turn1 response processing
type ProcessingError struct {
	Message   string
	Details   string
	Category  string
}

// NewProcessingError creates a new ProcessingError
func NewProcessingError(message, details string) *ProcessingError {
	return &ProcessingError{
		Message:  message,
		Details:  details,
		Category: "PROCESSING_ERROR",
	}
}

// Error implements the error interface
func (e *ProcessingError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// ParsingError represents an error during response parsing
type ParsingError struct {
	Message   string
	InnerErr  error
	Category  string
}

// NewParsingError creates a new ParsingError
func NewParsingError(message string, err error) *ParsingError {
	return &ParsingError{
		Message:  message,
		InnerErr: err,
		Category: "PARSING_ERROR",
	}
}

// Error implements the error interface
func (e *ParsingError) Error() string {
	if e.InnerErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.InnerErr)
	}
	return e.Message
}

// ValidationError represents an error during validation
type ValidationError struct {
	Message   string
	Field     string
	Expected  interface{}
	Actual    interface{}
	Category  string
}

// NewValidationError creates a new ValidationError
func NewValidationError(message, field string, expected, actual interface{}) *ValidationError {
	return &ValidationError{
		Message:  message,
		Field:    field,
		Expected: expected,
		Actual:   actual,
		Category: "VALIDATION_ERROR",
	}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: field '%s', expected: %v, actual: %v", 
			e.Message, e.Field, e.Expected, e.Actual)
	}
	return e.Message
}