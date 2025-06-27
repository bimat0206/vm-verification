# Parser Package

This package handles the parsing of Turn 1 responses in the vending machine verification workflow. It extracts structured data from AI model responses for further processing.

## Files Organization

The parser package is organized into multiple files, each with specific responsibilities:

- **parser_interface.go**: Defines the primary public interface for using the parser
- **patterns.go**: Manages parsing patterns and configuration
- **response_parser.go**: Core response parsing functionality
- **extractors.go**: Data extraction utilities
- **machine_structure_parser.go**: Specialized parsing for machine structure
- **state_parser.go**: Machine state and validation parsing

## Usage

To use the parser, create an instance using `NewParser()` with a logger:

```go
parser := parser.NewParser(logger)

// Parse validation response
validationResult := parser.ParseValidationResponse(responseMap)

// Parse machine structure
structureResult := parser.ParseMachineStructure(responseMap)

// Parse machine state
stateResult := parser.ParseMachineState(responseMap)

// Extract observations
observations := parser.ExtractObservations(responseMap)
```

## Parsing Strategy

The parser uses a multi-stage approach to extract data:

1. **Content Extraction**: Extract main content and thinking sections
2. **Structure Detection**: Determine if the response has a structured format
3. **Section Parsing**: Parse specific sections like machine structure, row status, etc.
4. **Data Extraction**: Extract structured data from sections
5. **Specialized Parsing**: Apply specialized parsing for machine structure, row state, etc.

## Testing

Use the included tests to verify parser functionality:

```
go test -v ./internal/parser
```