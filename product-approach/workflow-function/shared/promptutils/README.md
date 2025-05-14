# PromptUtils Package

A shared Go package for handling template-based prompt generation for Amazon Bedrock (Claude) in the Vending Machine Verification system.

## Overview

The `promptutils` package provides a unified way to:

1. Validate verification request inputs
2. Load and manage templates with versioning
3. Process input data for template rendering
4. Generate system prompts using Go templates
5. Configure Bedrock API requests
6. Handle utilities like logging, image processing, and data formatting

## Installation

Since this is a local module, add it to your Go module as a local dependency:

```go
// In your go.mod file
replace shared/promptutils => ../shared/promptutils
```

## Basic Usage

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"shared/promptutils"
)

var (
	promptProcessor *promptutils.PromptProcessor
)

func init() {
	// Initialize prompt processor with template path
	templateBasePath := "/opt/templates" // Default in container
	promptProcessor = promptutils.NewPromptProcessor(templateBasePath)
}

// HandleRequest is the Lambda handler function
func HandleRequest(ctx context.Context, event json.RawMessage) (promptutils.Response, error) {
	start := time.Now()
	log.Printf("Received event: %s", string(event))
	
	// Process the input using the shared promptutils package
	response, err := promptProcessor.ProcessInput(event)
	if err != nil {
		log.Printf("Error processing input: %v", err)
		return promptutils.Response{}, err
	}
	
	log.Printf("Completed in %v", time.Since(start))
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
```

## Directory Structure

The package is organized into the following modules:

- `promptutils`: Main package with public API and type definitions
  - `bedrock`: Bedrock API client and request/response handling
  - `templates`: Template loading, versioning, and rendering
  - `validator`: Input validation logic
  - `processor`: Data preparation for template rendering
  - `utils`: Helper utilities for formatting, logging, etc.

## Templates

Templates are expected to be organized in the following directory structure:

```
/opt/templates/
├── layout-vs-checking/
│   ├── v1.0.0.tmpl
│   ├── v1.1.0.tmpl
│   └── v1.2.3.tmpl
└── previous-vs-current/
    ├── v1.0.0.tmpl
    └── v1.1.0.tmpl
```

Templates use Go's text/template syntax and can access template data including:

- Verification context information
- Machine structure details
- Product mappings
- Historical verification data
- Any other fields defined in the `TemplateData` struct

## Environment Variables

The package supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates |
| TEMPLATE_VERSION_* | Template version for a specific type (e.g., TEMPLATE_VERSION_LAYOUT_VS_CHECKING) | Latest available |
| ANTHROPIC_VERSION | Anthropic API version for Bedrock | bedrock-2023-05-31 |
| MAX_TOKENS | Maximum tokens for response | 24000 |
| BUDGET_TOKENS | Tokens for Claude's thinking process | 16000 |
| THINKING_TYPE | Claude's thinking mode | enabled |
| PROMPT_VERSION | Default prompt version | 1.0.0 |
| DEBUG | Enable debug logging | false |
| COMPONENT_NAME | Component name for logging | - |
| REFERENCE_BUCKET | S3 bucket for reference images | - |
| CHECKING_BUCKET | S3 bucket for checking images | - |

## Custom Template Functions

The template engine provides several helper functions:

- String manipulation: `split`, `join`
- Math operations: `add`, `sub`, `mul`, `div`
- Comparisons: `gt`, `lt`, `eq`
- Array access: `index`
- Text formatting: `ordinal`

## Example Template

```
You are an AI assistant tasked with verifying vending machine product layouts.
Please analyze the following verification:

Verification ID: {{ .VerificationID }}
Verification Type: {{ .VerificationType }}
Machine ID: {{ .VendingMachineID }}
Location: {{ .Location }}

Machine Structure:
- {{ .RowCount }} rows ({{ .RowLabels }})
- {{ .ColumnCount }} columns ({{ .ColumnLabels }})
- Total positions: {{ .TotalPositions }}

{{ if eq .VerificationType "LAYOUT_VS_CHECKING" }}
Products:
{{ range .ProductMappings }}
- Position {{ .Position }}: {{ .ProductName }} (ID: {{ .ProductID }})
{{ end }}
{{ end }}

{{ if eq .VerificationType "PREVIOUS_VS_CURRENT" }}
Previous Verification:
- Previous ID: {{ .PreviousVerificationID }}
- Performed at: {{ .PreviousVerificationAt }}
- Hours since: {{ .HoursSinceLastVerification }}
{{ end }}
```