# Validator Package

This package provides validation functionality for the ProcessTurn1Response Lambda function. It ensures that data processed from Bedrock responses meets the required standards before proceeding to the next phase of the vending machine verification workflow.

## Overview

The validator package is responsible for ensuring data quality and integrity in the ProcessTurn1Response Lambda function. It performs comprehensive validation of extracted data from Turn 1 responses, checking for structural validity, completeness, and logical consistency.

## Components

### Core Components

1. **ValidatorInterface**:
   - Defines the contract for validator implementations
   - Ensures consistency across validation methods

2. **Validator**:
   - Primary implementation of the validator interface
   - Provides comprehensive validation of different data aspects
   - Integrates with shared schema types for consistency

3. **LegacyValidator**:
   - Provides backward compatibility with existing code
   - Adapts the new validator implementation to maintain existing interfaces

### Validation Types

The validator supports multiple validation scenarios:

1. **Reference Analysis Validation**:
   - Validates the complete reference analysis extracted from Bedrock responses
   - Ensures all required fields are present and valid
   - Checks for completeness of the analysis

2. **Machine Structure Validation**:
   - Validates the machine layout structure (rows, columns)
   - Ensures consistency between row/column counts and their orders
   - Uses shared schema types for validation

3. **Processing Result Validation**:
   - Validates the complete Turn 1 processing result
   - Includes validation specific to each processing path
   - Checks context preparation for Turn 2

4. **Historical Enhancement Validation**:
   - Validates historical data enhancement with visual confirmation
   - Ensures proper integration of baseline and visual data
   - Checks for required fields in enriched baseline

5. **Extracted State Validation**:
   - Validates the extracted state of the vending machine
   - Ensures row states are consistent with machine structure
   - Validates counts and position mappings

## Usage

### Basic Usage

```go
import (
    "workflow-function/ProcessTurn1Response/internal/validator"
    "workflow-function/shared/logger"
)

// Create a validator instance
log := logger.NewLogger()
validator := validator.NewValidator(log)

// Validate machine structure
structure := &schema.MachineStructure{
    RowCount:      3,
    ColumnsPerRow: 4,
    RowOrder:      []string{"A", "B", "C"},
    ColumnOrder:   []string{"1", "2", "3", "4"},
}
err := validator.ValidateMachineStructure(structure)

// Validate processing result
result := &types.Turn1ProcessingResult{
    // ... result data
}
err = validator.ValidateProcessingResult(result)
```

### Backward Compatibility

For code that uses the original validator interface:

```go
// Create a legacy validator instance
validator := validator.NewLegacyValidator(log)

// Use with existing code
err := validator.ValidateReferenceAnalysis(analysis)
```

## Validation Results

Validation results can include:

1. **Errors**: Critical issues that prevent further processing
2. **Warnings**: Non-critical issues that should be addressed
3. **Completeness**: Measurement of data completeness (0.0 to 1.0)

## Integration

The validator integrates with:

1. **Shared schema types**: For validation against standardized data structures
2. **Error handling**: Uses the shared errors package for consistent error reporting
3. **Logging**: Provides detailed logging of validation results and issues

## Best Practices

When using the validator:

1. Always validate extracted data before processing
2. Use the appropriate validation method for each data type
3. Check for warnings even when validation passes
4. Handle validation errors properly, with clear user feedback
5. Use the completeness score to assess data quality