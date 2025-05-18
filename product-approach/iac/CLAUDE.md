# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Infrastructure Overview

This repository contains Terraform infrastructure-as-code (IaC) for Kootoro's GenAI Vending Machine Verification Solution. The solution uses multiple AWS services in a serverless architecture to verify vending machine layouts using generative AI technology.

### Key AWS Services

- **AWS Lambda** (container-based): Serverless computing functions
- **Amazon ECR**: Container image storage 
- **Amazon DynamoDB**: Data storage (verification results, layout metadata, conversation history)
- **Amazon S3**: File storage (reference images, checking images, results)
- **AWS Step Functions**: Workflow orchestration of the verification process
- **Amazon API Gateway**: API management with direct Step Functions integration
- **AWS ECS with ALB**: Frontend hosting (Streamlit application)
- **Amazon Bedrock**: AI image processing
- **AWS CloudWatch**: Monitoring and logging

## Terraform Commands

### Basic Commands

```bash
# Initialize Terraform in current directory
terraform init

# Plan changes using development variables 
terraform plan -var-file=terraform.tfvars

# Apply changes
terraform apply -var-file=terraform.tfvars

# Validate configuration
./terraform-validator.sh
```

### Terraform Validator

The repository includes a custom validation script (`terraform-validator.sh`) that:

1. Checks AWS CLI and Terraform installation
2. Validates AWS credentials and account
3. Runs `terraform validate` to check configuration syntax
4. Executes `terraform plan` and saves the output
5. Checks for best practices (no hardcoded credentials, proper tagging)
6. Manages plan files, state backups, and tfvars backups

```bash
# Run validator script
./terraform-validator.sh

# Run with specific profile
./terraform-validator.sh -p profile_name
```

## Architecture Highlights

### Modular Structure

The infrastructure is organized in a modular structure where each AWS service component has its own module:

- **api_gateway**: API Gateway configuration with direct Step Functions integration
- **dynamodb**: DynamoDB tables for storing verification data
- **ecr**: ECR repositories for Lambda container images
- **ecs-streamlit**: ECS configuration for frontend Streamlit app
- **iam**: IAM roles and policies
- **lambda**: Lambda functions configuration
- **s3**: S3 buckets for image and data storage
- **secretsmanager**: Secrets management
- **step_functions**: Step Functions state machine for verification workflow
- **vpc**: VPC configuration
- **monitoring**: CloudWatch resources

### Step Functions Workflow

The core workflow is implemented as a Step Functions state machine:

1. **Initialize**: Sets up the verification context
2. **CheckVerificationType**: Determines verification type
3. **FetchHistoricalVerification**: (Optional) Retrieves history
4. **FetchImages**: Gets images from S3
5. **PrepareSystemPrompt**: Prepares prompt for Bedrock
6. **ExecuteTurn1/2**: Two-turn conversation with Bedrock
7. **FinalizeResults**: Processes verification results
8. **StoreResults**: Stores results in DynamoDB
9. **Notify**: (Optional) Sends notifications

### API Gateway Integration

API Gateway directly invokes the Step Functions state machine:

1. API Gateway receives requests at the POST /api/verifications endpoint
2. API Gateway directly invokes the Step Functions state machine
3. The workflow executes through the state machine, invoking Lambda functions

## Best Practices

1. When modifying State Machine definitions, ensure JSONPath references are correct
2. When adding new Lambda functions, update the IAM roles with proper permissions
3. Use `terraform validate` and the custom validator script before applying changes
4. Follow the established pattern for error handling with Retry and Catch blocks
5. Keep CHANGELOG.md files updated when making changes to modules
6. Maintain backward compatibility with parameter structures expected by Lambda functions