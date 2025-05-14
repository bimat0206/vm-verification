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

The function supports the following environment variables:

- `LOG_LEVEL`: Sets the logging level (default: "info")
- `AWS_REGION`: AWS region for services (default: from AWS Lambda environment)
- `DYNAMODB_TABLE`: DynamoDB table name (required for persistence)
- `S3_BUCKET`: S3 bucket for artifact storage (optional)

## Troubleshooting

### Common Issues

1. **Module Resolution Errors**: Ensure the `go.work` file includes all required modules and that replace directives are correctly configured in `go.mod`.

2. **Docker Build Failures**: Check that the Docker context includes both the function code and shared modules.

3. **Runtime Errors**: Verify environment variables are correctly set and that IAM permissions allow access to required AWS services.

Check the CHANGELOG.md for version history and recent changes.
