# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `s3state` package provides a standardized way to manage state data in S3 for AWS Lambda functions and Step Functions workflows. It organizes data into categories (images, prompts, responses, processing) and provides a consistent interface for state management in serverless applications.

## Key Concepts

1. **Manager Interface**: The core interface for all S3 state operations.
2. **References**: Lightweight pointers to S3 objects (bucket, key, size).
3. **Envelopes**: Containers that hold references and state for passing between Lambda functions.
4. **Categories**: Standardized organization for different types of data.

## Development Commands

### Building the Package

```bash
# Run tests
go test ./...

# Run integration tests (requires AWS credentials)
AWS_REGION=us-east-1 S3_BUCKET=test-bucket go test -tags=integration ./...
```

## Architecture

The package follows a modular architecture with clear separation of concerns:

1. **manager.go**: Core implementation of the Manager interface with S3 operations
2. **reference.go**: Reference and Envelope types for state tracking
3. **categories.go**: Category organization and standard filenames
4. **errors.go**: Comprehensive error types and handling

### Manager Interface

```go
type Manager interface {
    Store(category, key string, data []byte) (*Reference, error)
    Retrieve(ref *Reference) ([]byte, error)
    StoreJSON(category, key string, data interface{}) (*Reference, error)
    RetrieveJSON(ref *Reference, target interface{}) error
    SaveToEnvelope(envelope *Envelope, category, filename string, data interface{}) error
}
```

### Reference & Envelope Types

```go
type Reference struct {
    Bucket string `json:"bucket"`
    Key    string `json:"key"`
    Size   int64  `json:"size,omitempty"`
}

type Envelope struct {
    VerificationID string                `json:"verificationId"`
    References     map[string]*Reference `json:"s3References"`
    Status         string                `json:"status"`
    Summary        map[string]interface{} `json:"summary,omitempty"`
}
```

### Standard Categories

- `images`: Image data (Base64 encoded) and metadata
- `prompts`: AI prompts and conversation data
- `responses`: AI responses and conversation history
- `processing`: Processed analysis and intermediate results

### Error Handling

The package provides structured error types:
- `ErrorTypeS3Operation`: S3-specific errors
- `ErrorTypeValidation`: Input validation errors
- `ErrorTypeJSONOperation`: JSON marshaling/unmarshaling errors
- `ErrorTypeReference`: Reference-related errors
- `ErrorTypeCategory`: Category-related errors
- `ErrorTypeInternal`: Other internal errors

## Lambda Integration

When integrating with AWS Lambda, follow this pattern:

```go
func Handler(ctx context.Context, input interface{}) (*s3state.Envelope, error) {
    // Load envelope from Step Functions input
    envelope, err := s3state.LoadEnvelope(input)
    if err != nil {
        return nil, err
    }

    // Initialize state manager
    manager, err := s3state.New(os.Getenv("STATE_BUCKET"))
    if err != nil {
        return nil, err
    }

    // Process data
    result := processData()

    // Store result in envelope
    err = manager.SaveToEnvelope(envelope, s3state.CategoryProcessing, "result.json", result)
    if err != nil {
        return nil, err
    }

    // Update status
    envelope.SetStatus("COMPLETED")
    return envelope, nil
}
```

## Best Practices

1. **Category Organization**: Use standard categories for consistency.
2. **Error Handling**: Always check and handle errors with proper type checking.
3. **Reference Passing**: Pass references in envelopes between functions, not full data.
4. **Validation**: Use provided validation functions before operations.
5. **Envelopes**: Maintain the envelope structure throughout the workflow.