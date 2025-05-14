# ProcessTurn1Response Function

## Overview
Processes and extracts structured data from Turn 1 Bedrock responses, transforming raw AI output into normalized data structures for Turn 2 analysis.

## Responsibilities
- Parse Bedrock response content and thinking sections
- Extract reference state information
- Handle both use cases (LAYOUT_VS_CHECKING and PREVIOUS_VS_CURRENT)
- Prepare context for Turn 2 processing

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