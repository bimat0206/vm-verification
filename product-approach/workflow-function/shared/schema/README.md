# Shared Schema Package

This package provides standardized data models and validation for the vending machine verification workflow. It ensures consistency across all Lambda functions and Step Functions states, with enhanced support for Base64-encoded images for Bedrock API integration.

## Overview

The shared schema package defines:

1. Core data structures used throughout the workflow
2. Status constants for explicit state transitions
3. Validation functions to ensure data integrity
4. Helper functions for common operations
5. **Base64 image support for Bedrock API calls**
6. **Bedrock message formatting utilities**

## Key Features (v1.1.0)

### Enhanced Image Support
- **Dual Format**: Images are stored with both S3 references (for traceability) and Base64 data (for Bedrock)
- **Automatic Conversion**: Built-in helpers for converting between formats
- **Size Validation**: Automatic validation of Base64 data within Bedrock limits
- **Format Detection**: Intelligent detection of image formats from content type or filename

### Bedrock Integration
- **Native Message Format**: Direct support for Bedrock API message structure
- **Base64 Image Embedding**: Images are embedded as Base64 data in Bedrock messages
- **Builder Patterns**: Easy-to-use builders for constructing Bedrock messages and prompts

## Key Components

### Core Data Structures

- `VerificationContext`: The central context object that flows through the entire workflow
- `ImageData`: Enhanced structure with Base64 support for Bedrock
- `ImageInfo`: Complete image information including S3 metadata and Base64 data
- `ConversationState`: Tracks conversation state with Bedrock
- `BedrockMessage`: Native Bedrock API message format
- `CurrentPrompt`: Enhanced prompt structure supporting both text and Bedrock messages
- `BedrockConfig`: Configuration for Bedrock API calls
- `FinalResults`: Structure for verification results
- `WorkflowState`: Comprehensive state representation

### Status Constants

The package defines a comprehensive set of status constants that model the state machine's explicit status transitions:

```go
// Status constants aligned with state machine
const (
    StatusVerificationRequested  = "VERIFICATION_REQUESTED"
    StatusVerificationInitialized = "VERIFICATION_INITIALIZED"
    StatusFetchingImages         = "FETCHING_IMAGES"
    StatusImagesFetched          = "IMAGES_FETCHED"
    // ... and many more
)
```

### Enhanced Image Structures

```go
// ImageInfo with Base64 support
type ImageInfo struct {
    // S3 References (for traceability)
    URL      string `json:"url"`
    S3Key    string `json:"s3Key"`
    S3Bucket string `json:"s3Bucket"`
    
    // Image Properties
    Format      string `json:"format"`
    ContentType string `json:"contentType"`
    Size        int64  `json:"size"`
    
    // Base64 Data (for Bedrock)
    Base64Data string `json:"base64Data"`
    Base64Size int64  `json:"base64Size"`
    
    // Metadata
    LastModified string `json:"lastModified"`
    ETag         string `json:"etag"`
}
```

### Bedrock Message Format

```go
// BedrockMessage for Bedrock API
type BedrockMessage struct {
    Role    string           `json:"role"`
    Content []BedrockContent `json:"content"`
}

// BedrockContent with image support
type BedrockContent struct {
    Type  string            `json:"type"`  // "text" or "image"
    Text  string            `json:"text,omitempty"`
    Image *BedrockImageData `json:"image,omitempty"`
}
```

## Using the Package

### Importing

```go
import "github.com/kootoro/vending-machine-verification/workflow-function/shared/schema"
```

### Working with Base64 Images

```go
// Building ImageInfo with Base64 data
builder := schema.NewImageInfoBuilder()
imageInfo, err := builder.
    WithS3Info(s3URL, s3Key, s3Bucket).
    WithImageData(imageBytes, contentType, filename).
    WithDimensions(width, height).
    Build()

// Validating Base64 data
errors := schema.ValidateImageInfo(imageInfo, true) // requireBase64 = true
```

### Building Bedrock Messages

```go
// Create a message with text and image
message := schema.NewBedrockMessageBuilder("user").
    AddText("Analyze this vending machine image").
    AddImage(imageInfo).
    Build()

// Build complete messages for Bedrock API
messages := schema.BuildBedrockMessages(promptText, "reference", imageData)
```

### Creating Prompts with Bedrock Support

```go
// Build CurrentPrompt with Bedrock messages
prompt := schema.NewCurrentPromptBuilder(1).
    WithIncludeImage("reference").
    WithBedrockMessages("Analyze the reference image", imageData).
    WithMetadata("promptVersion", "1.0").
    Build()
```

### Processing Image Data

```go
// Ensure images have Base64 data
processor := schema.NewImageDataProcessor()
err := processor.EnsureBase64Generated(imageData)

// Validate for Bedrock API
err = processor.ValidateForBedrock(imageData)
```

## Base64 Helper Functions

The package provides comprehensive helpers for Base64 operations:

```go
// Global helper instances
schema.Base64Helpers.ConvertToBase64(imageBytes)
schema.Base64Helpers.ValidateImageFormat(format)
schema.Base64Helpers.CheckBase64SizeLimit(base64Data, maxSize)

// Image processing
schema.ImageProcessor.EnsureBase64Generated(imageData)
schema.ImageProcessor.ValidateForBedrock(imageData)
```

## Validation Functions

Enhanced validation supports both legacy and Base64 formats:

```go
// Validate verification context
errors := schema.ValidateVerificationContext(context)

// Validate images with Base64 requirement
errors := schema.ValidateImageData(imageData, true)

// Validate Bedrock messages
errors := schema.ValidateBedrockMessages(messages)

// Validate current prompt with Bedrock messages
errors := schema.ValidateCurrentPrompt(prompt, true)
```

## Migration Guide

### From v1.0.0 to v1.1.0

The v1.1.0 update adds Base64 support while maintaining backward compatibility:

1. **Existing Code**: No changes required for existing Lambda functions
2. **New Features**: Optionally use Base64 features for Bedrock integration
3. **Schema Version**: Automatically handled in validation functions

### Migrating Lambda Functions

#### FetchImages Function
```go
// Old approach (S3 only)
imageInfo := &schema.ImageInfo{
    URL:      s3URL,
    S3Key:    s3Key,
    S3Bucket: s3Bucket,
}

// New approach (S3 + Base64)
imageInfo, err := schema.NewImageInfoBuilder().
    WithS3Info(s3URL, s3Key, s3Bucket).
    WithImageData(imageBytes, contentType, filename).
    Build()
```

#### Prompt Functions
```go
// Old approach (text only)
prompt := &schema.CurrentPrompt{
    Text:         promptText,
    TurnNumber:   1,
    IncludeImage: "reference",
}

// New approach (Bedrock messages)
prompt := schema.NewCurrentPromptBuilder(1).
    WithIncludeImage("reference").
    WithBedrockMessages(promptText, imageData).
    Build()
```

#### Execute Functions
```go
// Access Bedrock messages directly
for _, message := range prompt.Messages {
    for _, content := range message.Content {
        if content.Type == "image" {
            // Base64 data is available in content.Image.Source.Bytes
            bedrockResponse := callBedrock(messages)
        }
    }
}
```

## Best Practices

### 1. Always Store Both Formats
- Keep S3 references for logging and auditing
- Generate Base64 data for Bedrock API calls
- Use the `Base64Generated` flag to track conversion status

### 2. Validate Image Formats
```go
// Always validate format before Bedrock API calls
if err := schema.Base64Helpers.ValidateImageFormat(imageInfo.Format); err != nil {
    return fmt.Errorf("unsupported format: %w", err)
}
```

### 3. Check Size Limits
```go
// Validate Base64 size before API calls
if err := imageInfo.ValidateBase64Size(); err != nil {
    return fmt.Errorf("image too large: %w", err)
}
```

### 4. Use Builder Patterns
- Use `ImageInfoBuilder` for creating complete image structures
- Use `BedrockMessageBuilder` for constructing API messages
- Use `CurrentPromptBuilder` for creating prompts with Bedrock messages

### 5. Handle Backward Compatibility
```go
// Check for schema version and handle both formats
if prompt.Messages != nil && len(prompt.Messages) > 0 {
    // Use Bedrock message format
    response := callBedrockWithMessages(prompt.Messages)
} else {
    // Use legacy text format
    response := callBedrockWithText(prompt.Text)
}
```

## Error Handling

The package provides detailed error information for Base64 operations:

```go
// Validation errors include specific field information
errors := schema.ValidateImageData(imageData, true)
for _, err := range errors {
    log.Printf("Field: %s, Error: %s", err.Field, err.Message)
}

// Size validation errors
if err := imageInfo.ValidateBase64Size(); err != nil {
    log.Printf("Size validation failed: %v", err)
}
```

## Schema Evolution

The v1.1.0 update maintains full backward compatibility:

1. **Schema Version**: Automatically updated to "1.1.0"
2. **Legacy Support**: All existing functionality preserved
3. **Gradual Migration**: Lambda functions can be updated incrementally
4. **Feature Detection**: Functions can detect and use new features when available

## Performance Considerations

### Base64 Conversion
- Base64 encoding increases size by ~33%
- Convert images once in FetchImages function
- Cache Base64 data throughout the workflow

### Memory Usage
- Base64 strings consume more memory than binary data
- Consider image size limits for Lambda functions
- Monitor memory usage when processing large images

### Bedrock API Limits
- Maximum image size: 20MB (Base64)
- Validate size before API calls
- Consider image compression for large files

## Future Enhancements

The schema package is designed for extensibility:

1. **Additional Image Formats**: Easy to add new format support
2. **Compression Options**: Base64 data compression for large images
3. **Streaming Support**: Potential for streaming large images
4. **Caching Mechanisms**: Built-in Base64 data caching

For more detailed examples and advanced usage, see the test files and Lambda function implementations.