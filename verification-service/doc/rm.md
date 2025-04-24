# Running the Verification Service Application

Here's a comprehensive guide to running the Vending Machine Verification Service application.

## Prerequisites

- Go 1.20 or higher
- Docker (for containerization)
- AWS credentials configured (for accessing S3, DynamoDB, and Bedrock)

## Local Development Setup

### 1. Clone the Repository

First, create a project directory and initialize the Go module:

```bash
mkdir -p verification-service
cd verification-service
go mod init verification-service
```

### 2. Set Up the Project Structure

Create the directory structure as outlined in the code:

```bash
mkdir -p cmd/server
mkdir -p internal/api/handlers internal/api/middleware
mkdir -p internal/app/services internal/app/dto
mkdir -p internal/domain/models internal/domain/engines internal/domain/services
mkdir -p internal/infrastructure/bedrock internal/infrastructure/dynamodb internal/infrastructure/s3 internal/infrastructure/logger
mkdir -p internal/config
mkdir -p pkg/errors pkg/utils
```

### 3. Copy the Generated Code Files

Copy all the code files to their respective directories as generated above.

### 4. Install Dependencies

Install the required Go dependencies:

```bash
go mod tidy
```

### 5. Configure Environment Variables

Create a `.env` file in the project root with the necessary configuration:

```env
# Server configuration
SERVER_PORT=3000
SERVER_READ_TIMEOUT_SECS=30
SERVER_WRITE_TIMEOUT_SECS=30
SERVER_IDLE_TIMEOUT_SECS=60

# AWS configuration
AWS_REGION=us-east-1

# DynamoDB configuration
DYNAMODB_VERIFICATION_TABLE=VerificationResults
DYNAMODB_LAYOUT_TABLE=LayoutMetadata

# S3 configuration
S3_REFERENCE_BUCKET=kootoro-reference-bucket
S3_CHECKING_BUCKET=kootoro-checking-bucket
S3_RESULTS_BUCKET=kootoro-results-bucket

# Bedrock configuration
BEDROCK_MODEL_ID=anthropic.claude-3-7-sonnet-20250219
BEDROCK_MAX_RETRIES=3

# Logging configuration
LOG_LEVEL=INFO
```

Make sure to replace the bucket names and other values with your actual AWS resources.

## Running the Application Locally

### 1. Set AWS Credentials

Ensure your AWS credentials are configured either through environment variables or the AWS credentials file:

```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
```

Or configure the `~/.aws/credentials` file.

### 2. Build and Run the Application

Build and run the application from the project root:

```bash
go build -o bin/verification-service ./cmd/server
./bin/verification-service
```

The application should start and listen on the configured port (default: 3000).

## Building and Running with Docker

### 1. Build the Docker Image

From the project root, build the Docker image:

```bash
docker build -t verification-service:latest .
```

### 2. Run the Docker Container

Run the container, passing in the required environment variables:

```bash
docker run -p 3000:3000 \
  -e AWS_ACCESS_KEY_ID=your_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret_key \
  -e AWS_REGION=us-east-1 \
  -e DYNAMODB_VERIFICATION_TABLE=VerificationResults \
  -e DYNAMODB_LAYOUT_TABLE=LayoutMetadata \
  -e S3_REFERENCE_BUCKET=kootoro-reference-bucket \
  -e S3_CHECKING_BUCKET=kootoro-checking-bucket \
  -e S3_RESULTS_BUCKET=kootoro-results-bucket \
  -e BEDROCK_MODEL_ID=anthropic.claude-3-7-sonnet-20250219 \
  verification-service:latest
```

## Deploying to AWS ECS

### 1. Push the Docker Image to ECR

Create an ECR repository if you don't have one:

```bash
aws ecr create-repository --repository-name verification-service
```

Tag and push the image:

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin your-account-id.dkr.ecr.us-east-1.amazonaws.com
docker tag verification-service:latest your-account-id.dkr.ecr.us-east-1.amazonaws.com/verification-service:latest
docker push your-account-id.dkr.ecr.us-east-1.amazonaws.com/verification-service:latest
```

### 2. Create ECS Task Definition

Create a task definition JSON file (`task-definition.json`):

```json
{
  "family": "verification-service",
  "executionRoleArn": "arn:aws:iam::your-account-id:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::your-account-id:role/verification-service-role",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "name": "verification-container",
      "image": "your-account-id.dkr.ecr.us-east-1.amazonaws.com/verification-service:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 3000,
          "hostPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "SERVER_PORT", "value": "3000" },
        { "name": "AWS_REGION", "value": "us-east-1" },
        { "name": "DYNAMODB_VERIFICATION_TABLE", "value": "VerificationResults" },
        { "name": "DYNAMODB_LAYOUT_TABLE", "value": "LayoutMetadata" },
        { "name": "S3_REFERENCE_BUCKET", "value": "kootoro-reference-bucket" },
        { "name": "S3_CHECKING_BUCKET", "value": "kootoro-checking-bucket" },
        { "name": "S3_RESULTS_BUCKET", "value": "kootoro-results-bucket" },
        { "name": "BEDROCK_MODEL_ID", "value": "anthropic.claude-3-7-sonnet-20250219" }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/verification-service",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "verification"
        }
      },
      "cpu": 1024,
      "memory": 2048
    }
  ],
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048"
}
```

Register the task definition:

```bash
aws ecs register-task-definition --cli-input-json file://task-definition.json
```

### 3. Create the ECS Service

Create an ECS service to run the task:

```bash
aws ecs create-service \
  --cluster verification-cluster \
  --service-name verification-service \
  --task-definition verification-service:1 \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx,subnet-yyy],securityGroups=[sg-zzz],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:us-east-1:your-account-id:targetgroup/verification-service-tg/xxx,containerName=verification-container,containerPort=3000"
```

## Testing the API Endpoints

Once the application is running, you can test the following endpoints:

### Health Check

```bash
curl http://localhost:3000/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-04-24T12:00:00Z"
}
```

### Initiate Verification

```bash
curl -X POST http://localhost:3000/api/v1/verification \
  -H "Content-Type: application/json" \
  -d '{
    "referenceImageUrl": "s3://kootoro-reference-bucket/processed/2025-04-21/14-25-10/23591_v1_abc_1q2w3e/image.png",
    "checkingImageUrl": "s3://kootoro-checking-bucket/2025-04-21/VM-3245/check_15-30-25.jpg",
    "vendingMachineId": "VM-3245",
    "layoutId": 23591,
    "layoutPrefix": "1q2w3e"
  }'
```

Expected response:
```json
{
  "verificationId": "verif-20250424120000",
  "verificationAt": "2025-04-24T12:00:00Z",
  "status": "INITIALIZED",
  "message": "Verification has been successfully initiated."
}
```

### Get Verification Status

```bash
curl http://localhost:3000/api/v1/verification/verif-20250424120000
```

The response will vary depending on the current status of the verification process.

### List Verifications

```bash
curl "http://localhost:3000/api/v1/verification?vendingMachineId=VM-3245&limit=10"
```

This will return a list of verifications matching the filters.

## Troubleshooting

### AWS Connectivity Issues

If the application can't connect to AWS services:

1. Verify AWS credentials are correctly configured
2. Check that the regions match between your configuration and AWS resources
3. Ensure IAM permissions include necessary actions for S3, DynamoDB, and Bedrock

### API Errors

If API endpoints return errors:

1. Check the application logs for detailed error messages
2. Verify the request format matches the expected input
3. Ensure required environment variables are correctly set

### Container Issues

If the Docker container fails to start:

1. Check the Docker logs: `docker logs <container_id>`
2. Verify that environment variables are correctly passed to the container
3. Ensure container networking is properly configured

## Development Tips

1. For faster development cycles, use hot reloading with tools like `air`
2. Use mocks for external services during local development
3. Set `LOG_LEVEL=DEBUG` for more verbose logging during development

This setup guide should help you get the Vending Machine Verification Service up and running in various environments.