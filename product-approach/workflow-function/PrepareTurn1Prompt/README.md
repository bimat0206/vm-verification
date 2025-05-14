# Kootoro GenAI Vending Machine Verification - PrepareTurn1Prompt

This Lambda function generates Turn 1 prompts for the Kootoro GenAI Vending Machine Verification solution. 
It creates properly formatted first turn prompts for Amazon Bedrock (Claude 3.7 Sonnet) to analyze the reference image in the two-turn conversation workflow.

## Overview

The PrepareTurn1Prompt function is a critical component in the Kootoro vending machine verification workflow:

1. It receives verification context information and a system prompt
2. Loads the appropriate Turn 1 prompt template based on verification type
3. Validates input data and image format
4. Creates a Bedrock message containing the Turn 1 prompt text and reference image
5. Configures Bedrock parameters for optimal performance
6. Returns a complete Turn 1 prompt ready for the first conversation turn

## Architecture

The function follows a modular architecture with the following components:

- **main.go**: Lambda handler and core execution logic
- **types.go**: Data structures and type definitions
- **validator.go**: Input validation logic
- **templates.go**: Template loading and management
- **bedrock.go**: Bedrock message construction and configuration
- **processor.go**: Core business logic for creating Turn 1 prompt
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
docker build -t kootoro-prepare-turn1-prompt:v1.0.0 .

# Tag for ECR
docker tag kootoro-prepare-turn1-prompt:v1.0.0 ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-turn1-prompt:v1.0.0

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-turn1-prompt:v1.0.0

# Deploy Lambda using CloudFormation or AWS CLI
aws lambda create-function \
  --function-name kootoro-prepare-turn1-prompt \
  --package-type Image \
  --code ImageUri=${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/kootoro-prepare-turn1-prompt:v1.0.0 \
  --role arn:aws:iam::${AWS_ACCOUNT_ID}:role/KootoroLambdaExecutionRole \
  --environment "Variables={REFERENCE_BUCKET=your-reference-bucket,CHECKING_BUCKET=your-checking-bucket,TEMPLATE_BASE_PATH=/opt/templates}"
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| REFERENCE_BUCKET | S3 bucket for reference layout images | - | Yes |
| CHECKING_BUCKET | S3 bucket for checking images | - | Yes |
| TEMPLATE_BASE_PATH | Path to template directory | /opt/templates | No |
| ANTHROPIC_VERSION | Anthropic API version for Bedrock | bedrock-2023-05-31 | No |
| MAX_TOKENS | Maximum tokens for response | 24000 | No |
| BUDGET_TOKENS | Tokens for Claude's thinking process | 16000 | No |
| THINKING_TYPE | Claude's thinking mode | enabled | No |
| TURN1_TEMPLATE_VERSION_LAYOUT_VS_CHECKING | Template version for layout-vs-checking | 1.0.0 | No |
| TURN1_TEMPLATE_VERSION_PREVIOUS_VS_CURRENT | Template version for previous-vs-current | 1.0.0 | No |
| DEBUG | Enable debug logging | false | No |

### IAM Permissions

The Lambda function requires the following permissions:

- s3:GetObject on the reference and checking buckets
- logs:CreateLogGroup, logs:CreateLogStream, logs:PutLogEvents for CloudWatch logging
- lambda:InvokeFunction for calling other Lambda functions (if applicable)

## Template System

This function uses local templates stored within the container for generating Turn 1 prompts. Templates should be organized in the following structure:

```
/opt/templates/
├── layout-vs-checking/
│   ├── v1.0.0.tmpl              # System prompt template (used by PrepareSystemPrompt)
│   ├── turn1-v1.0.0.tmpl        # Turn 1 template (used by PrepareTurn1Prompt)
│   └── turn1-v1.1.0.tmpl        # Newer Turn 1 template version
└── previous-vs-current/
    ├── v1.0.0.tmpl              # System prompt template (used by PrepareSystemPrompt)
    ├── turn1-v1.0.0.tmpl        # Turn 1 template (used by PrepareTurn1Prompt)
    └── turn1-v1.1.0.tmpl        # Newer Turn 1 template version
```

Template versions follow semantic versioning (x.y.z) and the highest version is automatically selected unless overridden by environment variables.

### Template Format

Templates use Go's text/template format with the following context variables:

```go
type TemplateData struct {
    VerificationType   string
    VerificationID     string
    VerificationAt     string
    VendingMachineID   string
    Location           string
    TurnNumber         int               // Always 1 for Turn1Prompt
    
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
```

## Usage

### Step Functions Integration

This function is designed to be called from the verification workflow Step Functions state machine. The expected input is the verification context, system prompt, and reference image information.

### Example Input (Layout vs. Checking)

```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "SYSTEM_PROMPT_READY",
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
  },
  "turnNumber": 1,
  "includeImage": "reference"
}
```

### Example Output

```json
{
  "verificationContext": {
    "verificationId": "verif-2025042115302500",
    "verificationAt": "2025-04-21T15:30:25Z",
    "status": "TURN1_PROMPT_READY",
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
  "currentPrompt": {
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "Please analyze the FIRST image (Reference Image)..."
          },
          {
            "type": "image",
            "image": {
              "format": "png",
              "source": {
                "s3Location": {
                  "uri": "s3://your-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
                  "bucketOwner": "111122223333"
                }
              }
            }
          }
        ]
      }
    ],
    "turnNumber": 1,
    "promptId": "prompt-verif-2025042115302500-turn1",
    "createdAt": "2025-04-21T15:30:30Z",
    "promptVersion": "1.0.0",
    "imageIncluded": "reference"
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
```

## Development

### Local Development

```bash
# Setup local environment
go mod download

# Run tests
go test ./...

# Build locally
go build -o main cmd/main.go

# Local testing with AWS SAM
sam local invoke PrepareTurn1PromptFunction --event events/layout-vs-checking.json
```

### Adding New Templates

1. Create a new template file in the appropriate directory with the prefix "turn1-"
2. Use Go's text/template syntax
3. Test the template with sample data
4. Update the version information in the deployment

## Limitations and Requirements

- Image Formats: Only JPEG and PNG images are supported (Bedrock requirement)
- Bucket Names: Must be configured via environment variables
- Template Structure: Templates must be properly formatted and accessible at runtime
- Memory Usage: Function requires at least 512MB Lambda memory allocation
- Timeout: Recommended timeout is 30 seconds

## Related Components

- PrepareSystemPrompt: Generates system prompts for the verification workflow
- ExecuteTurn1: Invokes Bedrock with the Turn 1 prompt
- ProcessTurn1Response: Processes the Turn 1 response from Bedrock
- PrepareTurn2Prompt: Generates Turn 2 prompts for the verification workflow

## License

Proprietary - All Rights Reserved