# Kootoro GenAI Vending Machine Verification Infrastructure

This repository contains Terraform code to deploy the infrastructure for Kootoro's GenAI Vending Machine Verification Solution.

## Architecture

The solution uses the following AWS services:
- AWS Lambda (container-based) for serverless computing
- Amazon ECR for container image storage
- Amazon DynamoDB for data storage
- Amazon S3 for file storage
- AWS Step Functions for workflow orchestration
- Amazon API Gateway for API management
- AWS App Runner for frontend hosting
- Amazon Bedrock for AI image processing
- AWS CloudWatch for monitoring and logging

### API Gateway and Step Functions Integration

The architecture features a direct integration between API Gateway and Step Functions:

1. API Gateway receives requests at the POST /api/verifications endpoint
2. API Gateway directly invokes the Step Functions state machine using the StartExecution action
3. Step Functions Initialize state maps the input parameters to ensure consistent structure
4. The workflow executes through the state machine, invoking Lambda functions as needed
5. Results are stored in DynamoDB and S3 for retrieval

This integration pattern ensures consistent input handling regardless of whether the workflow is invoked via API Gateway or directly through the Step Functions API.

## Prerequisites

- Terraform >= 1.0.0
- AWS CLI configured with appropriate permissions
- Docker for building and pushing container images
- Access to AWS services including Bedrock

## Repository Structure

```
kootoro-terraform/
├── main.tf                  # Main configuration with module calls
├── variables.tf             # Input variables definition
├── outputs.tf               # Output values
├── providers.tf             # Provider configuration
├── versions.tf              # Terraform version constraints
├── locals.tf                # Local variables for naming and tags
├── env/
│   ├── dev.tfvars           # Development environment variables
│   ├── test.tfvars          # Testing environment variables
│   └── prod.tfvars          # Production environment variables
└── modules/
    ├── s3/                  # S3 buckets module
    ├── dynamodb/            # DynamoDB tables module
    ├── ecr/                 # ECR repositories module
    ├── lambda/              # Lambda functions module
    ├── step_functions/      # Step Functions module
    ├── api_gateway/         # API Gateway module
    ├── app_runner/          # App Runner module
    ├── iam/                 # IAM roles and policies module
    └── monitoring/          # CloudWatch resources module
```

## Getting Started

### 1. Initialize Terraform

```bash
terraform init
```

### 2. Select Environment Configuration

The project uses `.tfvars` files for environment-specific configurations. Choose the appropriate environment:

```bash
# For development environment
terraform plan -var-file=env/dev.tfvars

# For production environment
terraform plan -var-file=env/prod.tfvars
```

### 3. Deploy Infrastructure

```bash
# For development environment
terraform apply -var-file=env/dev.tfvars

# For production environment
terraform apply -var-file=env/prod.tfvars
```

## Initial Deployment

The first deployment will create ECR repositories and Lambda functions using placeholder images (nginx). After the infrastructure is in place:

1. Build your Lambda function container images
2. Push them to the created ECR repositories
3. The Lambda functions will automatically use these images on the next deployment

Example workflow for a Lambda function image:

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com

# Build image
docker build -t kootoro-dev-ecr-initialize-dev01:latest ./function-code/initialize/

# Tag image with ECR repository
docker tag kootoro-dev-ecr-initialize-dev01:latest ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-initialize-dev01:latest

# Push image to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-initialize-dev01:latest
```

## Custom Configurations

The infrastructure can be customized by modifying the `.tfvars` files. Major configuration sections include:

- **General**: AWS region, project name, environment
- **S3**: Lifecycle rules, bucket creation
- **DynamoDB**: Billing mode, capacity units
- **ECR**: Repository configuration, lifecycle policies
- **Lambda**: Memory sizes, timeouts, architecture
- **API Gateway**: Throttling, CORS, metrics
- **Step Functions**: Logging levels
- **App Runner**: CPU, memory, auto-deployments
- **Bedrock**: Model ID, token configuration
- **Monitoring**: Log retention, alarm endpoints

## Output Values

After deployment, Terraform will output useful information:

- S3 bucket names and ARNs
- ECR repository URLs
- DynamoDB table names and ARNs
- Lambda function names and ARNs
- Step Functions state machine ARN
- API Gateway endpoint URL
- App Runner service URL

## Cleanup

To destroy the infrastructure:

```bash
terraform destroy -var-file=env/dev.tfvars
```

**IMPORTANT**: In production, S3 buckets have `force_destroy` set to `false` to prevent accidental data loss. To destroy these buckets, first empty them or update the configuration.

## Security Considerations

- IAM roles follow the principle of least privilege
- S3 buckets are configured with appropriate lifecycle policies
- DynamoDB tables use point-in-time recovery in production
- ECR repositories scan images on push
- All resources are properly tagged for governance
