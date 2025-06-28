# ProcessTurn1Response Function

## Overview
Processes and extracts structured data from Turn 1 Bedrock responses, transforming raw AI output into normalized data structures for Turn 2 analysis. The function follows a reference-based architecture using S3 for state management.

## Responsibilities
- Parse Bedrock response content and thinking sections
- Extract reference state information
- Handle both use cases (LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT)
- Prepare context for Turn 2 processing
- Store and manage state in S3 using reference-based architecture

## Code Architecture
The codebase follows a modern architecture with clear separation of concerns:

- **cmd/main.go**: Lambda entry point and bootstrap
- **internal/**: Core implementation modules
  - **handler/**: Lambda request orchestration
    - **handler.go**: Main workflow coordination
    - **error_handler.go**: Standardized error responses
  - **processor/**: Business logic processing
    - **processor.go**: Processing coordination
    - **paths.go**: Specialized processing paths
    - **error_handler.go**: Processing-specific errors
  - **parser/**: Response parsing
    - **parser.go**: Core parser interface
    - **extractor.go**: Data extraction utilities
    - **patterns.go**: Parsing patterns configuration
  - **state/**: State management
    - **manager.go**: S3StateManager integration
    - **operations.go**: High-level state operations
    - **error_handler.go**: State-related error handling
  - **validator/**: Data validation
    - **validator.go**: Comprehensive validation layer
  - **errors/**: Centralized error handling
    - **errors.go**: Custom error types and utilities
    - **error_converter.go**: Error type conversion
    - **storage_error.go**: Storage-specific errors
  - **config/**: Configuration management
    - **config.go**: Environment and settings

## Reference-Based Architecture

The function utilizes a reference-based architecture where:

1. **State References**: Data is stored in S3 and referenced by keys rather than passing complete data in memory
2. **Envelope Pattern**: A lightweight envelope contains references to all relevant data, reducing payload sizes
3. **Selective Loading**: Data is loaded from S3 only when needed, improving memory efficiency
4. **Categorized Storage**: Different data types are stored in specific S3 categories (processing, conversations, etc.)

This architecture offers several advantages:
- Reduced memory usage and payload sizes
- Better scalability for large response processing
- Simplified integration with Step Functions workflows
- More reliable state persistence across function invocations

## Input/Output
- **Input**: WorkflowState with S3 references to Turn1 data
- **Output**: Updated WorkflowState with new references to processed analysis

## Use Cases Handled
1. **UC1: LAYOUT_VS_CHECKING** - Simple validation flow (PathValidationFlow)
2. **UC2 with Historical**: Enhancement with existing data (PathHistoricalEnhancement)
3. **UC2 without Historical**: Fresh extraction from response (PathFreshExtraction)

## Validator Layer

The validator provides comprehensive validation for different aspects of Turn 1 response processing:

- **ValidatorInterface**: Interface defining all validation methods
- **Validation Types**:
  - Reference Analysis Validation
  - Machine Structure Validation
  - Processing Result Validation
  - Historical Enhancement Validation
  - Extracted State Validation
  - Turn 2 Context Validation

## Error Handling System

The function implements a robust error handling system:

- **Error Categories**: Input, Process, State, System
- **Error Severity Levels**: Debug, Info, Warning, Error, Critical
- **Structured Errors**: All errors include operation, category, code, and message
- **Error Propagation**: Errors are wrapped with additional context as they propagate
- **API Error Responses**: Standardized error formats for API consumers

Example:
```go
// Creating a processing error
err := errors.ProcessingError("ExtractStructure", "Failed to extract machine structure", innerErr)

// Creating an input validation error with details
details := map[string]interface{}{"field": "rows", "expected": "> 0", "actual": 0}
err := errors.ValidationError("ValidateInput", "Invalid machine structure", details)

// Converting to API response
errorResponse := errors.ConvertToErrorInfo(err)
```

## Dependencies
- shared/schema: Data structures and validation
- shared/logger: Structured logging
- shared/s3state: S3-based state management
- shared/errors: Error handling utilities

## Development

### Local Development with Go Workspace

This project uses Go workspaces to manage local module dependencies. The workspace is configured in the root `go.work` file.

```bash
# Run go mod tidy to ensure dependencies are properly resolved
go mod tidy

# Build the Lambda function locally
go build -o main ./cmd

# Run tests
go test ./...
```

### Docker Build

The project includes a multi-stage Dockerfile that optimizes the build process:

```bash
# Build the Docker image
docker build -t process-turn1-response -f Dockerfile ../..

# Run the Docker image locally (for testing)
docker run --rm -p 9000:8080 process-turn1-response

# Test the function with AWS Lambda Runtime Interface Emulator
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{...}'
```

### Environment Variables

The function requires the following environment variables:

- `LOG_LEVEL`: Sets the logging level (default: "info")
- `AWS_REGION`: AWS region for services (default: from AWS Lambda environment)
- `S3_STATE_BUCKET`: S3 bucket for state storage (required)
- `DYNAMODB_CONVERSATION_TABLE`: DynamoDB table for conversation history (optional)

## Troubleshooting

### Common Issues

1. **S3 Reference Errors**: Ensure the S3 bucket exists and the function has proper IAM permissions.

2. **Missing Required References**: Check that all required S3 references are included in the input state.

3. **Invalid Schema Version**: Verify that the input WorkflowState has a compatible schema version.

4. **Parse Failures**: Review the Turn1Response format to ensure it meets the expected structure.

Check the CHANGELOG.md for version history and recent changes.