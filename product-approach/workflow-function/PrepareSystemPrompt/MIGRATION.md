# Shared Package Migration Guide

This document outlines the process of migrating the PrepareSystemPrompt function to use the shared packages.

## Shared Packages Used

We've refactored the PrepareSystemPrompt function to use the following shared packages:

1. **shared/schema**: Defines common data structures and validation logic
2. **shared/templateloader**: Provides template loading, caching, and rendering
3. **shared/s3utils**: Utilities for working with S3 including URL validation

## Changes Made

### 1. Dependencies

The `go.mod` file was updated to include the shared packages. We set up proper replacement directives to point to the relative path of the shared packages.

### 2. Type Definitions

The original `internal/types.go` was refactored to:

- Use the `schema.WorkflowState` as the primary data structure
- Keep some local types like `TemplateData` which are specific to this function
- Provide helper methods to extract data from the generic maps in `WorkflowState`

### 3. Template Management

The original `internal/templates.go` was replaced with:

- Integration with `shared/templateloader` for template loading and caching
- Helper functions to maintain backward compatibility with existing code

### 4. Validation

The validation was updated to:

- Use `shared/schema/validation` for standard validation
- Use `shared/s3utils` for S3 URL validation
- Keep function-specific validation logic

### 5. Bedrock Integration

The Bedrock integration was updated to:

- Use `schema.BedrockConfig` and `schema.Thinking` types
- Keep function-specific request/response handling

## Challenges

During migration, we encountered several challenges:

1. **Module Path Resolution**: Go module paths needed careful configuration to properly resolve the shared packages.

2. **Type Compatibility**: The shared types sometimes required adaptation to work with existing code.

3. **Dependency Management**: Managing dependencies across multiple packages required careful attention.

## Integration Steps for Other Functions

To integrate shared packages into other functions, follow these steps:

1. **Update go.mod**: Add the shared packages and set up proper replacement directives.

2. **Adapt Types**: Update code to use shared types where applicable, keeping function-specific types as needed.

3. **Use Shared Utilities**: Replace custom implementations with shared utilities.

4. **Test Thoroughly**: Ensure the function still works with sample events.

## Benefits

Using shared packages provides several benefits:

1. **Standardized Data Structures**: Consistent types across functions
2. **Code Reuse**: Avoid duplicating logic
3. **Maintainability**: Easier to maintain and update
4. **Consistency**: Ensure consistent behavior across functions

## Future Improvements

Some potential future improvements:

1. **Better Documentation**: Add more documentation to shared packages
2. **More Test Coverage**: Add tests for shared functionality
3. **CI/CD Integration**: Ensure CI/CD verifies compatibility with shared packages