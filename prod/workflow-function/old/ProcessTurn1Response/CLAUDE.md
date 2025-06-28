# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ProcessTurn1Response is a Go Lambda function that processes and extracts structured data from Turn 1 Bedrock responses in a vending machine verification workflow. The function transforms raw AI output into normalized data structures for Turn 2 analysis.

This Lambda is part of a larger workflow that handles vending machine verification through image analysis:
- It processes the AI response from Turn 1 (initial analysis)
- Extracts reference state information
- Prepares context for Turn 2 processing
- Handles different use cases (layout vs checking and previous vs current)

## Build Commands

```bash
# Build the Lambda function locally
go build -o main ./cmd/process-turn1-response

# Run tests
go test ./...

# Build Docker image
docker build -t process-turn1-response .

# Run the Docker container locally
docker run -p 9000:8080 process-turn1-response

# Invoke the function locally (using AWS Lambda Runtime Interface Emulator)
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{...}'
```

## Code Architecture

The codebase follows a standard Go project layout structure:

1. **Main Entry Point** (`cmd/process-turn1-response/main.go`): 
   - Initializes the Lambda handler
   - Sets up logging and dependencies
   - Starting point for Lambda execution

2. **Internal Packages**:
   - `internal/processor`: Contains the handler and core processing logic
     - `handler.go`: Primary handler logic, validates inputs, processes responses
     - `processor.go`: Core business logic for processing Turn 1 responses
   
   - `internal/parser`: Response parsing utilities
     - `parsers.go`: Parses Bedrock responses using configurable patterns
   
   - `internal/validator`: Data validation
     - `validators.go`: Validates the processed data, ensures completeness
   
   - `internal/types`: Domain models and data structures
     - `types.go`: Defines data structures used throughout the function
   
   - `internal/dependencies`: External service integration
     - `dependencies.go`: Sets up AWS service clients and configurations

## Important Concepts

### Processing Paths

Three main processing paths:

1. **UC1 Validation Flow** (`VALIDATION_FLOW`): 
   - Simple validation for LAYOUT_VS_CHECKING use case
   - Confirms structure matches the expected layout

2. **UC2 Historical Enhancement** (`HISTORICAL_ENHANCEMENT`):
   - Uses historical context to enhance analysis
   - Combines previous verification data with visual confirmation

3. **UC2 Fresh Extraction** (`FRESH_EXTRACTION`):
   - Full extraction when no historical context exists
   - Builds reference state from scratch

### Data Structures

Key data structures:

- `MachineStructure`: Represents vending machine layout (rows, columns)
- `RowState`: Contains state information for a single row
- `ParsedResponse`: Holds parsed Bedrock response data
- `Turn1ProcessingResult`: Contains the final processing result

### Integration Points

- Input: WorkflowState with turn1Response (from Bedrock)
- Output: WorkflowState with referenceAnalysis and updated status
- Uses shared libraries for S3, DynamoDB operations, logging, and schema validation

## Best Practices

1. Maintain the established directory structure
   - Logic goes in `internal/` packages
   - Entry point in `cmd/`
   - Only expose what's necessary to other packages

2. Follow Go idiomatic patterns
   - Use interfaces for dependency injection
   - Keep packages focused on a single responsibility
   - Use descriptive package and type names

3. Code quality
   - Always validate input data using the validator methods
   - Follow the established parsing patterns for consistent extraction
   - Provide appropriate error handling with detailed error messages
   - Use structured logging with correlation IDs for traceability
   - Add fallback mechanisms for graceful degradation when parsing fails