# Shared Schema Package

This package provides standardized data models and validation for the vending machine verification workflow. It ensures consistency across all Lambda functions and Step Functions states, with S3-only storage for Base64-encoded images for Bedrock API integration.

## Overview

The shared schema package defines:

1. Core data structures used throughout the workflow
2. Status constants for explicit state transitions
3. Validation functions to ensure data integrity
4. Helper functions for common operations
5. **S3-based Base64 image support for Bedrock API calls**
6. **Bedrock message formatting utilities**

## Key Features (v2.0.0)

### S3-Only Image Support
- **S3 Storage**: Images are stored in S3 with references for traceability
- **Automatic Conversion**: Built-in helpers for converting between formats
- **Size Validation**: Automatic validation of Base64 data within Bedrock limits
- **Format Detection**: Intelligent detection of image formats from content type or filename
- **Modular Codebase**: Split into multiple files for better maintainability

### Bedrock Integration
- **Native Message Format**: Direct support for Bedrock API message structure
- **S3-based Base64 Image Retrieval**: Images are retrieved from S3 for Bedrock messages
- **Builder Patterns**: Easy-to-use builders for constructing Bedrock messages and prompts

## File Structure

The package is now organized into multiple files for better maintainability:

- **constants.go**: All constants used throughout the package
- **core.go**: Core types and helper functions
- **image_info.go**: ImageInfo struct and its methods
- **image_data.go**: ImageData struct and its methods
- **bedrock.go**: Bedrock-related types and functions
- **s3_helpers.go**: S3 storage helpers (replacing base64_helpers.go)
- **validation.go**: Validation functions
- **types.go**: Additional type definitions

## Key Components

### Core Data Structures

- `VerificationContext`: The central context object that flows through the entire workflow
- `ImageData`: Enhanced structure with S3-based Base64 support for Bedrock
- `ImageInfo`: Complete image information including S3 metadata and S3 references for Base64 data
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

### S3-Based Image Structures

```go
// ImageInfo with S3-based Base64 support
type ImageInfo struct {
    // S3 References (for traceability)
    URL      string `json:"url"`
    S3Key    string `json:"s3Key"`
    S3Bucket string `json:"s3Bucket"`
    
    // Image Properties
    Format      string `json:"format"`
    ContentType string `json:"contentType"`
    Size        int64  `json:"size"`
    
    // S3 Storage for Base64
    Base64Size       int64  `json:"base64Size,omitempty"`
    Base64S3Bucket   string `json:"base64S3Bucket,omitempty"`
    Base64S3Key      string `json:"base64S3Key,omitempty"`
    
    // Metadata
    LastModified string `json:"lastModified"`
    ETag         string `json:"etag"`
}
```

### Bedrock Message Format (v2.0.0)

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

// BedrockImageData represents image data for Bedrock API
type BedrockImageData struct {
    Format string             `json:"format"`         // "png", "jpeg", etc.
    Source BedrockImageSource `json:"source"`
}

// BedrockImageSource represents the source of image data for Bedrock
type BedrockImageSource struct {
    Type      string `json:"type"`           // "base64" for v2.0.0 schema
    Media_type string `json:"media_type"`    // "image/jpeg", "image/png", etc.
    Data      string `json:"data"`           // Base64-encoded image data
}
```

## Using the Package

### Importing

```go
import "github.com/kootoro/vending-machine-verification/workflow-function/shared/schema"
```

### Working with S3-Based Base64 Images

```go
// Initialize S3 client
s3Client := s3.NewFromConfig(cfg)

// Create S3 storage config
s3Config := &schema.S3StorageConfig{
    TempBase64Bucket:       "temp-base64-bucket",
    Base64RetrievalTimeout: 5000, // 5 seconds
}

// Building ImageInfo with S3-based Base64 storage
builder := schema.NewS3ImageInfoBuilder(s3Config, s3Client)
imageInfo, err := builder.
    WithS3Info(s3URL, s3Key, s3Bucket).
    WithImageDataAndS3Storage(imageBytes, contentType, filename, "reference").
    WithDimensions(width, height).
    Build()

// Upload Base64 data to S3
err = builder.Upload()

// Validating Base64 data
errors := schema.ValidateImageInfo(imageInfo, true) // requireBase64 = true
```

### Building Bedrock Messages with S3 Retrieval

```go
// Initialize S3 retriever
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)

// Create a message with text and image
message := schema.NewBedrockMessageBuilder("user", retriever).
    AddText("Analyze this vending machine image").
    AddImageWithS3Retrieval(imageInfo).
    Build()

// For backward compatibility
messages := schema.BuildBedrockMessages(promptText, "reference", imageData)
```

### Creating Prompts with S3-Based Bedrock Support

```go
// Build CurrentPrompt with S3-based Bedrock messages
prompt := schema.NewCurrentPromptBuilder(1, retriever).
    WithIncludeImage("reference").
    WithBedrockMessagesS3("Analyze the reference image", imageData).
    WithMetadata("promptVersion", "1.0").
    Build()
```

### Processing Image Data with S3 Storage

```go
// Initialize S3 processor
processor := schema.NewS3ImageDataProcessor(retriever, builder)

// Ensure images have S3-based Base64 data
err := processor.EnsureS3Base64Generated(imageData)

// Validate for Bedrock API
err = processor.ValidateForBedrockS3(imageData)

// For backward compatibility
legacyProcessor := schema.NewImageDataProcessor()
err = legacyProcessor.ValidateForBedrock(imageData)
```

## S3 Helper Functions

The package provides comprehensive helpers for S3-based Base64 operations:

```go
// Global helper instances
schema.Base64Helpers.ConvertToBase64(imageBytes)
schema.Base64Helpers.ValidateImageFormat(format)
schema.Base64Helpers.CheckBase64SizeLimit(base64Data, maxSize)

// S3-based image processing
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)
builder := schema.NewS3ImageInfoBuilder(s3Config, s3Client)
processor := schema.S3ImageProcessor(retriever, builder)
err = processor.ValidateForBedrockS3(imageData)
```

## Validation Functions

Enhanced validation supports S3-based Base64 storage:

```go
// Validate verification context
errors := schema.ValidateVerificationContext(context)

// Validate images with S3-based Base64 requirement
errors := schema.ValidateImageData(imageData, true)

// Validate Bedrock messages
errors := schema.ValidateBedrockMessages(messages)

// Validate current prompt with Bedrock messages
errors := schema.ValidateCurrentPrompt(prompt, true)
```

## Migration Guide

### From v1.3.0 to v2.0.0

The v2.0.0 update aligns with the JSON schema definitions while maintaining compatibility:

1. **Bedrock Message Format**: Updated to match the JSON schema v2.0.0 with new field names
2. **Field Naming**: Changed `Source.Type` from "bytes" to "base64" and updated field names
3. **Validation**: Enhanced validation for the new Bedrock message structure
4. **Schema Version**: Automatically updated to "2.0.0"

### From v1.0.0 to v1.3.0

The v1.3.0 update migrates to S3-only storage for Base64 data while maintaining backward compatibility:

1. **Existing Code**: Minimal changes required for existing Lambda functions
2. **New Features**: S3-based Base64 storage for Bedrock integration
3. **Schema Version**: Automatically updated to "1.3.0"
4. **File Structure**: Split into multiple files for better maintainability

### Migrating Lambda Functions

#### FetchImages Function
```go
// Old approach (S3 only)
imageInfo := &schema.ImageInfo{
    URL:      s3URL,
    S3Key:    s3Key,
    S3Bucket: s3Bucket,
}

// New approach (S3-based Base64)
s3Config := &schema.S3StorageConfig{
    TempBase64Bucket: "temp-base64-bucket",
}
builder := schema.NewS3ImageInfoBuilder(s3Config, s3Client)
imageInfo, err := builder.
    WithS3Info(s3URL, s3Key, s3Bucket).
    WithImageDataAndS3Storage(imageBytes, contentType, filename, "reference").
    Build()
err = builder.Upload()
```

#### Prompt Functions
```go
// Old approach (text only)
prompt := &schema.CurrentPrompt{
    Text:         promptText,
    TurnNumber:   1,
    IncludeImage: "reference",
}

// New approach (S3-based Bedrock messages)
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)
prompt := schema.NewCurrentPromptBuilder(1, retriever).
    WithIncludeImage("reference").
    WithBedrockMessagesS3(promptText, imageData).
    Build()
```

#### Execute Functions
```go
// Access Bedrock messages directly
for _, message := range prompt.Messages {
    for _, content := range message.Content {
        if content.Type == "image" {
            // Base64 data is retrieved from S3 by the retriever
            bedrockResponse := callBedrock(messages)
        }
    }
}
```

## Best Practices

### 1. Use S3 for Base64 Storage
- Keep S3 references for logging and auditing
- Store Base64 data in S3 for Bedrock API calls
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
- Use `S3ImageInfoBuilder` for creating complete image structures
- Use `BedrockMessageBuilder` with S3 retriever for constructing API messages
- Use `CurrentPromptBuilder` with S3 retriever for creating prompts with Bedrock messages

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

The package provides detailed error information for S3-based Base64 operations:

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

// S3 retrieval errors
retriever := schema.NewS3Base64Retriever(s3Client, s3Config)
base64Data, err := retriever.RetrieveBase64Data(imageInfo)
if err != nil {
    log.Printf("S3 retrieval failed: %v", err)
}
```

## Schema Evolution

### Version 2.0.0 Updates
The v2.0.0 update aligns with the JSON schema definitions while maintaining backward compatibility:

1. **Schema Version**: Automatically updated to "2.0.0"
2. **Bedrock Message Structure**: Updated to match the JSON schema 2.0.0 with new field names:
   ```go
   // BedrockImageSource in v2.0.0
   type BedrockImageSource struct {
       Type      string `json:"type"`           // "base64" for v2.0.0 schema
       Media_type string `json:"media_type"`    // "image/jpeg", "image/png", etc.
       Data      string `json:"data"`           // Base64-encoded image data
   }
   
   // Previous BedrockImageSource (v1.3.0)
   // type BedrockImageSource struct {
   //     Type  string `json:"type"`           // "bytes" 
   //     Bytes string `json:"bytes"`          // Base64-encoded image data
   // }
   ```
3. **Builder Pattern Improvements**: Enhanced builders with more explicit naming and better error handling
4. **Validation Enhancement**: Improved validation for Bedrock message format with detailed error reporting
5. **Legacy Support**: Maintained backward compatibility through careful refactoring

### Migration from v1.3.0 to v2.0.0

When updating code from v1.3.0 to v2.0.0, you should:

1. Use the updated BedrockMessageBuilder for creating Bedrock messages
2. Update any direct references to BedrockImageSource fields to use the new names
3. Take advantage of the enhanced validation for better error reporting

Example upgrade:
```go
// v2.0.0 approach
messageBuilder := schema.NewBedrockMessageBuilder("user", retriever)
messageBuilder.AddText("Analyze this vending machine image")
err := messageBuilder.AddImageWithS3Retrieval(imageInfo)
if err != nil {
    // Enhanced error handling
    return fmt.Errorf("failed to add image: %w", err)
}
message := messageBuilder.Build()
```

### Version 1.3.0 Features
The v1.3.0 update maintained backward compatibility:

1. **Schema Version**: Automatically updated to "1.3.0"
2. **Legacy Support**: Most existing functionality preserved
3. **Gradual Migration**: Lambda functions can be updated incrementally
4. **Feature Detection**: Functions can detect and use new features when available
5. **File Structure**: Split into multiple files for better maintainability

## Performance Considerations

### S3-Based Base64 Storage
- Base64 encoding increases size by ~33%
- Convert images once in FetchImages function and store in S3
- Retrieve Base64 data from S3 when needed

### Memory Usage
- S3-based storage reduces Lambda memory usage
- Base64 data is stored in S3 instead of in memory
- Consider S3 retrieval timeouts for large files

### Bedrock API Limits
- Maximum image size: 20MB (Base64)
- Validate size before API calls
- Consider image compression for large files

## Future Enhancements

The schema package is designed for extensibility:

1. **Additional Image Formats**: Easy to add new format support
2. **Compression Options**: Base64 data compression for large images
3. **Streaming Support**: Potential for streaming large images from S3
4. **Caching Mechanisms**: Built-in S3 caching for Base64 data

For more detailed examples and advanced usage, see the test files and Lambda function implementations.
