# ProcessTurn1Response Function

## Overview
Processes and extracts structured data from Turn 1 Bedrock responses, transforming raw AI output into normalized data structures for Turn 2 analysis.

## Responsibilities
- Parse Bedrock response content and thinking sections
- Extract reference state information
- Handle both use cases (LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT)
- Prepare context for Turn 2 processing

## Code Structure
The codebase is organized using a standard Go project layout:

- **cmd/main.go**: Lambda entry point and bootstrap
- **internal/**: Core implementation modules
  - **parser/**: Response parsing (refactored into multiple specialized files)
    - **parser_interface.go**: Public parser interface
    - **patterns.go**: Parsing patterns management
    - **response_parser.go**: Main parsing logic
    - **extractors.go**: Data extraction utilities
    - **machine_structure_parser.go**: Structure parsing
    - **state_parser.go**: State parsing and validation
  - **processor/**: Business logic processing
  - **types/**: Domain models and data structures
  - **validator/**: Data validation utilities
  - **dependencies/**: External service integration

## Input/Output
- **Input**: WorkflowState with turn1Response
- **Output**: WorkflowState with referenceAnalysis and updated status

## Use Cases Handled
1. **UC1: LAYOUT_VS_CHECKING** - Simple validation flow
2. **UC2 with Historical**: Enhancement with existing data
3. **UC2 without Historical**: Fresh extraction from response

## Dependencies
- shared/schema: Data structures and validation
- shared/logger: Structured logging
- shared/s3utils: S3 operations (if needed)
- shared/dbutils: DynamoDB operations (if needed)

## Development
To build and test the function locally:

```bash
# Build the Lambda function locally
go build -o main ./cmd/process-turn1-response

# Run tests
go test ./...
```

Check the CHANGELOG.md for version history and recent changes.