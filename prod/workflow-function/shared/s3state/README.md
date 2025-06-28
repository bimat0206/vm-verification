```markdown
# S3 State Management Package

A simple Go package for managing state data in S3 with category-based organization, designed specifically for AWS Lambda functions and Step Functions workflows.

## Overview

This package provides a standardized way to store and retrieve state data across AWS Lambda functions using S3 as the storage backend. It organizes data into categories and provides a consistent interface for state management in serverless workflows.

## Features

- **Simple API**: Easy-to-use interface with just 5 core methods
- **Category Organization**: Automatic organization by data type (images, prompts, responses, processing)
- **Step Functions Integration**: Lightweight reference passing between workflow states
- **JSON Support**: Built-in JSON marshaling/unmarshaling
- **Error Handling**: Comprehensive error types with retry logic
- **AWS Native**: Built on AWS SDK v2 with proper authentication

## Installation

Since this is a local package, import it directly in your Go modules:

```go
import "path/to/s3state"
```

Or add it to your `go.mod`:

```
replace github.com/kootoro/s3state => ./path/to/s3state
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/kootoro/s3state"
)

func main() {
    // Initialize manager
    manager, err := s3state.New("my-state-bucket")
    if err != nil {
        log.Fatal(err)
    }

    // Store JSON data
    data := map[string]interface{}{"message": "Hello, World!"}
    ref, err := manager.StoreJSON(s3state.CategoryProcessing, "test.json", data)
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve JSON data
    var result map[string]interface{}
    err = manager.RetrieveJSON(ref, &result)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Retrieved: %+v", result)
}
```

## Categories

The package organizes data into four standard categories:

- **`images`**: Image data (Base64 encoded) and metadata
- **`prompts`**: AI prompts and conversation data  
- **`responses`**: AI responses and conversation history
- **`processing`**: Processed analysis and intermediate results

## Key Format

All S3 keys follow the format: `{verificationId}/{category}/{filename}`

Examples:
- `verif-123/images/metadata.json`
- `verif-123/prompts/system-prompt.json`
- `verif-123/processing/final-results.json`

## API Reference

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

### Core Types

```go
type Reference struct {
    Bucket string `json:"bucket"`
    Key    string `json:"key"`
    Size   int64  `json:"size,omitempty"`
}

type Envelope struct {
    VerificationID string                        `json:"verificationId"`
    References     map[string]*Reference         `json:"s3References"`
    Status         string                        `json:"status"`
    Summary        map[string]interface{}        `json:"summary,omitempty"`
}
```

### Category Constants

```go
const (
    CategoryImages     = "images"
    CategoryPrompts    = "prompts"
    CategoryResponses  = "responses"
    CategoryProcessing = "processing"
)
```

### Standard Filenames

The package provides constants for common filenames:

```go
// Images category
ImageMetadataFile     = "metadata"
ReferenceBase64File   = "reference-base64.base64"
CheckingBase64File    = "checking-base64.base64"

// Prompts category
SystemPromptFile      = "system-prompt"
Turn1PromptFile       = "turn1-prompt"
Turn2PromptFile       = "turn2-prompt"

// Responses category
Turn1ResponseFile     = "turn1-raw-response"
Turn2ResponseFile     = "turn2-raw-response"

// Processing category
InitializationFile    = "initialization.json"
LayoutMetadataFile    = "layout-metadata.json"
HistoricalContextFile = "historical-context.json"
Turn1AnalysisFile     = "turn1-analysis.json"
Turn2AnalysisFile     = "turn2-analysis.json"
FinalResultsFile      = "final-results.json"
```

## Error Handling

The package provides structured error handling with different error types:

```go
// Error types
const (
    ErrorTypeS3Operation   = "S3_OPERATION_ERROR"
    ErrorTypeValidation    = "VALIDATION_ERROR"
    ErrorTypeJSONOperation = "JSON_OPERATION_ERROR"
    ErrorTypeReference     = "REFERENCE_ERROR"
    ErrorTypeCategory      = "CATEGORY_ERROR"
    ErrorTypeInternal      = "INTERNAL_ERROR"
)

// Check error types
if s3state.IsS3Error(err) {
    // Handle S3-specific error
}

if s3state.IsRetryable(err) {
    // Implement retry logic
}
```

## Lambda Integration

### Basic Lambda Function

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

## Environment Requirements

- AWS credentials configured (IAM role, environment variables, or AWS config)
- S3 bucket with appropriate permissions
- Go 1.19 or later

## Package Structure

```
s3state/
├── manager.go          # Main S3StateManager implementation
├── reference.go        # Reference and Envelope types
├── categories.go       # Category definitions and utilities
├── errors.go          # Error types and handling
├── examples/          # Usage examples
└── README.md          # This file
```

## Prerequisites

1. **AWS Credentials**: Ensure AWS credentials are configured
2. **S3 Bucket**: Create an S3 bucket for state storage
3. **IAM Permissions**: Required permissions:
   - `s3:GetObject`
   - `s3:PutObject`
   - `s3:ListBucket`

## Testing

Run tests with:

```bash
go test ./...
```

For integration tests with actual S3:

```bash
AWS_REGION=us-east-1 S3_BUCKET=test-bucket go test -tags=integration ./...
```

## Best Practices

1. **Error Handling**: Always check and handle errors appropriately
2. **Resource Cleanup**: The package handles S3 lifecycle automatically
3. **Validation**: Use provided validation functions before operations
4. **Category Usage**: Stick to the standard categories for consistency
5. **Reference Passing**: Pass references in Step Functions, not full data

## Limitations

- Maximum object size limited by S3 (5TB)
- No built-in compression (add manually if needed)
- No cross-region replication support
- Lambda memory limits apply for large objects

## Support

For issues or questions, refer to the examples directory or contact the development team.
```

```markdown
# S3 State Management Examples

This directory contains practical examples demonstrating how to use the S3 State Management package in various scenarios.

## Prerequisites

Before running the examples, ensure you have:

1. **AWS Credentials**: Configure AWS credentials via:
   - AWS CLI: `aws configure`
   - Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
   - IAM role (recommended for Lambda)

2. **S3 Bucket**: Create an S3 bucket for testing:
   ```bash
   aws s3 mb s3://your-test-bucket
   ```

3. **Environment Variables**: Set the bucket name:
   ```bash
   export STATE_BUCKET=your-test-bucket
   ```

## Examples Overview

### 1. basic_usage.go
**Purpose**: Demonstrates basic store and retrieve operations

**What it shows**:
- Storing and retrieving JSON data
- Storing and retrieving raw binary data
- Basic error handling

**Run it**:
```bash
cd examples
go run basic_usage.go
```

### 2. lambda_integration.go
**Purpose**: Shows how to integrate with AWS Lambda and Step Functions

**What it shows**:
- Lambda function handler structure
- Loading envelopes from Step Functions input
- Processing data and updating envelopes
- Returning envelopes to Step Functions

**Note**: This is a complete Lambda function. To test:
```bash
cd examples
go build -o lambda_handler lambda_integration.go
# Deploy to AWS Lambda or use SAM/Serverless
```

### 3. envelope_operations.go
**Purpose**: Comprehensive envelope manipulation examples

**What it shows**:
- Creating and configuring envelopes
- Adding references and metadata
- Validating envelopes
- Working with categories
- JSON serialization for Step Functions

**Run it**:
```bash
cd examples
go run envelope_operations.go
```

### 4. category_operations.go
**Purpose**: Working with the category system

**What it shows**:
- Using category constants
- Key generation and parsing
- Category validation
- Standard filename usage
- Reference key building

**Run it**:
```bash
cd examples
go run category_operations.go
```

### 5. error_handling.go
**Purpose**: Comprehensive error handling patterns

**What it shows**:
- Creating different error types
- Error type checking
- Validation helpers
- Error collection with ErrorList
- Retry logic patterns

**Run it**:
```bash
cd examples
go run error_handling.go
```

### 6. real_world_usage.go
**Purpose**: Simulates a complete verification workflow

**What it shows**:
- Multi-step Lambda function workflow
- Data passing between functions
- State progression through workflow stages
- Real-world data structures and processing

**Run it**:
```bash
cd examples
go run real_world_usage.go
```

## Running Examples

### Individual Examples

Run any example individually:

```bash
cd examples
go run <example_name>.go
```

### All Examples

Run all examples in sequence:

```bash
cd examples
for file in *.go; do
    echo "Running $file..."
    go run "$file"
    echo "---"
done
```

### With Custom Bucket

Override the default bucket:

```bash
cd examples
STATE_BUCKET=my-custom-bucket go run basic_usage.go
```

## Example Data Flow

Here's how the examples relate to a real workflow:

```
1. basic_usage.go
   └── Learn basic operations

2. category_operations.go
   └── Understand organization

3. envelope_operations.go
   └── Master state management

4. error_handling.go
   └── Handle edge cases

5. lambda_integration.go
   └── Build Lambda functions

6. real_world_usage.go
   └── Complete workflow
```

## Testing Examples

### Verify S3 Connectivity

```bash
cd examples
go run -c "
package main
import (
    \"fmt\"
    \"github.com/kootoro/s3state\"
)
func main() {
    manager, err := s3state.New(\"your-test-bucket\")
    if err != nil {
        fmt.Printf(\"Error: %v\n\", err)
        return
    }
    fmt.Println(\"Successfully connected to S3\")
}"
```


### S3 Operations Logging

To see actual S3 operations, add logging to examples:

```go
import "log"

// Add before creating manager
log.Println("Creating S3 state manager...")

// Add after operations
log.Printf("Stored object: %s", ref.String())
```




## Best Practices from Examples

1. **Always handle errors**: Every example shows proper error handling
2. **Use envelopes for Step Functions**: Pass references, not data
3. **Validate before operations**: Use validation helpers
4. **Organize by categories**: Follow the standard category structure
5. **Use standard filenames**: Leverage provided constants

