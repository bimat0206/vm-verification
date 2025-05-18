# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This repository contains a Terraform module for AWS Step Functions that implements a workflow for vending machine verification. The module creates an AWS Step Functions state machine with integration for Lambda functions, DynamoDB, and API Gateway.

## Module Architecture

### Core Components

- **State Machine**: Defines a workflow that orchestrates Lambda functions for vending machine verification
- **IAM Roles and Policies**: Configures permissions for Step Functions to invoke Lambda functions and access other AWS resources
- **API Gateway Integration**: Optional integration allowing direct invocation of the state machine via API Gateway
- **CloudWatch Logging**: Configuration for state machine execution logs

### Key Files

- `main.tf`: Creates the Step Functions state machine and API Gateway integration
- `iam.tf`: Defines IAM roles and policies for Step Functions
- `variables.tf`: Contains input variable definitions
- `outputs.tf`: Contains output definitions
- `templates/state_machine_definition.tftpl`: Template for the state machine definition

## Workflow Process

The Step Functions state machine implements a verification workflow with the following high-level steps:

1. **Initialize**: Sets up the verification context and parameters
2. **CheckVerificationType**: Determines the verification type (LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)
3. **FetchHistoricalVerification**: (Optional) Retrieves historical verification data
4. **FetchImages**: Gets images from S3
5. **PrepareSystemPrompt**: Prepares the system prompt for Bedrock
6. **Execute Turn 1/2**: Two-turn conversation with Bedrock for image verification
7. **FinalizeResults**: Processes and finalizes the verification results
8. **StoreResults**: Stores the results in DynamoDB
9. **Notify**: (Optional) Sends notifications

## Working with the Module

### Testing Changes

```bash
# Initialize Terraform in the parent directory
cd ../..
terraform init

# Validate Terraform configurations
terraform validate

# Plan changes to see what will be modified
terraform plan -target=module.step_functions

# Apply changes (only when ready)
terraform apply -target=module.step_functions
```

### Common Issues and Fixes

1. **JSONPath errors**: When modifying state machine definition, ensure JSONPath references are correct and data is properly structured
2. **IAM permission issues**: When adding new Lambda functions or AWS resources, ensure IAM roles have proper permissions
3. **State machine validation errors**: Use AWS Step Functions Workflow Studio to validate state machine definitions visually

## Recent Fixes

### FetchImages Parameter Fix (2025-05-14)
- Removed `historicalContext.$: "$.historicalContext"` from FetchImages state Parameters
- This fixed an issue where the workflow would fail for LAYOUT_VS_CHECKING verification types that don't have historicalContext
- The FetchImages Lambda function already handles missing historicalContext by creating an empty object

## Best Practices

1. Use the `templatefile` function in Terraform to pass dynamic values to the state machine definition
2. Follow the established pattern for error handling with Retry and Catch blocks
3. Maintain backward compatibility with the parameter structures expected by Lambda functions
4. Add proper documentation in code comments and update the README.md when making changes
5. Test state machine changes in isolation using the `-target` flag with Terraform