# Kootoro GenAI Vending Machine Verification - PrepareSystemPrompt

This Lambda function generates system prompts for the Kootoro GenAI Vending Machine Verification solution. 
It creates properly formatted system prompts for Amazon Bedrock (Claude 3.7 Sonnet) based on the verification type.

## Overview

The PrepareSystemPrompt function is a critical component in the Kootoro vending machine verification workflow:

1. It receives verification context information
2. Loads the appropriate prompt template based on verification type
3. Validates input data and image formats
4. Injects dynamic data like machine structure and product mappings into the template
5. Configures Bedrock parameters for optimal performance
6. Returns a complete system prompt ready for the two-turn conversation flow

## Architecture

![System Architecture](docs/images/architecture.png)

The function follows a modular architecture with the following components:

- **main.go**: Lambda handler and core execution logic
- **types.go**: Type definitions for all data structures
- **validator.go**: Input validation logic
- **templates.go**: Template loading and management
- **bedrock.go**: Bedrock configuration and integration
- **utils.go**: Helper functions for various operations

## Installation

### Prerequisites

- Go 1.19+
- Docker
- AWS CLI configured with appropriate permissions

### Build and Deploy

Build and deploy the Lambda function as a container:

```bash
# Build the container
docker build -t kootoro-prepare-system-prompt:v1.0.0 .

# Tag for ECR
docker tag kootoro-prepare-system-prompt:v1.0.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v1.0.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v1.0.0

# Deploy Lambda using CloudFormation or AWS CLI
aws lambda create-function \
  --function-name kootoro-prepare-system-prompt \
  --package-type Image \
  --code ImageUri=${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-system-prompt:v1.0.0 \
  --role arn:aws:iam::${AWS_ACCOUNT_ID}:role/KootoroLambdaExecutionRole \
  --environment "Variables={REFERENCE_BUCKET=your-reference-bucket,CHECKING_BUCKET=your-checking-bucket,TEMPLATE_BASE_PATH=/opt/templates}"

  Configuration
Environment Variables
VariableDescriptionDefaultRequiredREFERENCE_BUCKETS3 bucket for reference layout images-YesCHECKING_BUCKETS3 bucket for checking images-YesTEMPLATE_BASE_PATHPath to template directory/opt/templatesNoANTHROPIC_VERSIONAnthropic API version for Bedrockbedrock-2023-05-31NoMAX_TOKENSMaximum tokens for response24000NoBUDGET_TOKENSTokens for Claude's thinking process16000NoTHINKING_TYPEClaude's thinking modeenabledNoPROMPT_VERSIONDefault prompt version1.0.0NoDEBUGEnable debug loggingfalseNo
IAM Permissions
The Lambda function requires the following permissions:

s3:GetObject on the reference and checking buckets
dynamodb:GetItem on the LayoutMetadata table (optional)
logs:CreateLogGroup, logs:CreateLogStream, logs:PutLogEvents for CloudWatch logging

Local Templates
This function uses local templates stored within the container for generating system prompts. Templates should be organized in the following structure:
/opt/templates/
├── layout-vs-checking/
│   ├── v1.0.0.tmpl
│   ├── v1.1.0.tmpl
│   └── v1.2.3.tmpl
└── previous-vs-current/
    ├── v1.0.0.tmpl
    └── v1.1.0.tmpl

Alternatively, a flatter structure can be used:
/opt/templates/
├── layout-vs-checking.tmpl
└── previous-vs-current.tmpl

Template versions can be controlled via environment variables (TEMPLATE_VERSION_LAYOUT_VS_CHECKING, TEMPLATE_VERSION_PREVIOUS_VS_CURRENT) or through the discovery mechanism that selects the latest version available.

Template Format
Templates use Go's text/template format with the following context variables:
type TemplateData struct {
    VerificationType   string
    VerificationID     string
    VerificationAt     string
    VendingMachineID   string
    Location           string
    
    MachineStructure   *MachineStructure
    RowCount           int
    ColumnCount        int
    RowLabels          string
    ColumnLabels       string
    TotalPositions     int
    
    ProductMappings    []ProductMapping
    
    PreviousVerificationID     string
    PreviousVerificationAt     string
    PreviousVerificationStatus string
    HoursSinceLastVerification float64
    VerificationSummary        *VerificationSummary
}

Usage
Example Input (Layout vs. Checking)
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "INITIALIZED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e",
    "referenceImageUrl": "s3://your-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
    "checkingImageUrl": "s3://your-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 6,
      "columnsPerRow": 10,
      "rowOrder": ["A", "B", "C", "D", "E", "F"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    },
    "productPositionMap": {
      "A01": {
        "productId": 3486,
        "productName": "Mì Hảo Hảo"
      }
    },
    "location": "Office Building A, Floor 3"
  }
}{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "INITIALIZED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e",
    "referenceImageUrl": "s3://your-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
    "checkingImageUrl": "s3://your-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 6,
      "columnsPerRow": 10,
      "rowOrder": ["A", "B", "C", "D", "E", "F"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    },
    "productPositionMap": {
      "A01": {
        "productId": 3486,
        "productName": "Mì Hảo Hảo"
      }
    },
    "location": "Office Building A, Floor 3"
  }
}

Example Output
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "SYSTEM_PROMPT_READY",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 6,
      "columnsPerRow": 10,
      "rowOrder": ["A", "B", "C", "D", "E", "F"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    }
  },
  "systemPrompt": {
    "content": "You are an AI assistant specialized in analyzing vending machine product placement...",
    "promptId": "prompt-20250421-23ds-system",
    "createdAt": "2025-04-21T15:30:28Z",
    "promptVersion": "1.2.3"
  },
  "bedrockConfig": {
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 24000,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 16000
    }
  }
}
Development
Local Development
bash# Setup local environment
go mod download

# Run tests
go test ./...

# Build locally
go build -o main cmd/main.go

# Local testing with AWS SAM
sam local invoke PrepareSystemPromptFunction --event events/layout-vs-checking.json
Adding New Templates

Create a new template file in the appropriate directory
Use Go's text/template syntax
Test the template with sample data
Update the version information in the deployment

Limitations and Requirements

Image Formats: Only JPEG and PNG images are supported (Bedrock requirement)
Bucket Names: Must be configured via environment variables
Template Structure: Templates must be properly formatted and accessible at runtime
Memory Usage: Function requires at least 512MB Lambda memory allocation
Timeout: Recommended timeout is 30 seconds

