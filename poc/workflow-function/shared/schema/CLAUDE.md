# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a shared schema package for a vending machine verification system built on AWS serverless architecture. The package defines standardized data models, validation logic, and helper functions used across various Lambda functions and Step Functions in the workflow.

The main purpose of the library is to facilitate a consistent schema between different workflow components that handle verification of vending machine layouts against physical checking images using Amazon Bedrock's multimodal AI capabilities.

## Key Components and Architecture

### Core Data Model

1. **VerificationContext**: Central context object flowing through the entire workflow
2. **ImageData**: Structure for reference and checking images with S3-based Base64 support
3. **ImageInfo**: Detailed image information including S3 metadata and references for Base64 data
4. **WorkflowState**: Complete representation of workflow state at any point
5. **ConversationState**: State of conversation with Amazon Bedrock
6. **CurrentPrompt**: Enhanced prompt structure supporting both text and Bedrock messages
7. **BedrockMessage**: Native format for Amazon Bedrock API messages

### File Structure

- **constants.go**: Status constants and other constants
- **core.go**: Core types and helper functions
- **image_info.go**: ImageInfo struct and methods
- **image_data.go**: ImageData struct and methods
- **bedrock.go**: Bedrock-related types and functions
- **s3_helpers.go**: S3 storage helpers
- **validation.go**: Validation functions
- **types.go**: Additional type definitions

### Workflow Structure

The schema supports this verification workflow:

1. Initialize verification request
2. Fetch reference and checking images
3. Prepare system prompt for Bedrock
4. Process first turn with reference image
5. Process second turn with checking image
6. Compare images and finalize results
7. Store results and send notifications

## Common Development Tasks

### Building and Testing

```bash
# Run all tests
go test ./...

# Run specific tests
go test -v ./... -run TestValidateImageInfo

# Build the package
go build ./...
```

### Working with S3-Based Base64 Storage

The schema uses S3 to store Base64-encoded image data for Bedrock API calls. When developing functions that use this package:

```bash
# Creating an ImageInfo with S3 storage
s3Config := &schema.S3StorageConfig{
    TempBase64Bucket: "temp-base64-bucket",
}
builder := schema.NewS3ImageInfoBuilder(s3Config, s3Client)
imageInfo, _ := builder.
    WithS3Info(s3URL, s3Key, s3Bucket).
    WithImageDataAndS3Storage(imageBytes, contentType, filename, "reference").
    Build()
err := builder.Upload()

# Retrieving Base64 data from S3
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)
base64Data, err := retriever.RetrieveBase64Data(imageInfo)
```

### Creating Bedrock Messages

```bash
# Build a Bedrock message with S3 retrieval
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)
message := schema.NewBedrockMessageBuilder("user", retriever).
    AddText("Analyze this vending machine image").
    AddImageWithS3Retrieval(imageInfo).
    Build()
```

### Validation

```bash
# Validate a verification context
errors := schema.ValidateVerificationContext(context)

# Validate image data with S3-based Base64 requirement
errors := schema.ValidateImageData(imageData, true)
```

## Best Practices

1. **Use S3 for Base64 Storage**
   - Always use the S3 storage helpers for Base64 data
   - The schema version 1.3.0 requires S3-based storage for all Base64 data

2. **Validate Before Bedrock API Calls**
   - Always validate image formats are supported (png, jpeg, jpg)
   - Check Base64 size limits before API calls (20MB limit)

3. **Use Builder Patterns**
   - Use the provided builders rather than manually constructing complex objects
   - `S3ImageInfoBuilder`, `BedrockMessageBuilder`, and `CurrentPromptBuilder` ensure proper construction

4. **Maintain Backward Compatibility**
   - The schema supports both legacy and new field names in some structs
   - Always use the getter methods like `images.GetReference()` which handle both formats

5. **Handle Errors Properly**
   - All validation functions return detailed error information
   - Return validation errors to the caller for proper handling

## Version Information

- Current schema version: 1.3.0
- Major change in 1.3.0: Migration to S3-only storage for Base64 data