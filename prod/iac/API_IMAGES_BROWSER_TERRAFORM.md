# API Images Browser - Terraform Configuration

This document describes the Terraform configuration for the API Images Browser Lambda function and its associated AWS resources.

## Overview

The API Images Browser is a Go-based AWS Lambda function that provides a REST API for browsing images stored in S3 buckets. It enables users to navigate through S3 bucket contents in a file-browser-like interface.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │───▶│ Lambda Function │───▶│   S3 Buckets    │
│                 │    │                 │    │                 │
│ /api/images/    │    │ api-images-     │    │ - Reference     │
│ browser         │    │ browser         │    │ - Checking      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CloudWatch    │    │   ECR Repository│    │   IAM Roles     │
│   Logs          │    │                 │    │   & Policies    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Resources Created

### 1. ECR Repository
- **Name**: `{project}-{env}-ecr-api-images-browser-{suffix}`
- **Purpose**: Stores the Docker container image for the Lambda function
- **Configuration**: 
  - Image tag mutability: MUTABLE
  - Scan on push: Enabled
  - Encryption: AES256

### 2. Lambda Function
- **Name**: `{project}-{env}-lambda-api-images-browser-{suffix}`
- **Runtime**: Container (Go 1.20)
- **Memory**: 256 MB
- **Timeout**: 30 seconds
- **Environment Variables**:
  - `REFERENCE_BUCKET`: Name of the reference images bucket
  - `CHECKING_BUCKET`: Name of the checking images bucket
  - `LOG_LEVEL`: Logging level (INFO)

### 3. API Gateway Integration
- **Endpoints**:
  - `GET /api/images/browser` - Browse root directory
  - `GET /api/images/browser/{path+}` - Browse specific path
  - `OPTIONS` methods for CORS support
- **Integration Type**: AWS_PROXY
- **CORS**: Enabled for web application access

### 4. IAM Permissions
- **S3 Permissions**:
  - `s3:ListBucket` on reference and checking buckets
  - `s3:GetObject` on bucket contents
  - `s3:GetBucketLocation` and `s3:GetBucketVersioning`
- **CloudWatch Logs**: Write permissions for function logs
- **ECR**: Pull permissions for container images

### 5. CloudWatch Resources
- **Log Group**: `/aws/lambda/{function-name}`
- **Log Retention**: Configurable (default from variables)
- **Metrics**: Standard Lambda metrics (Duration, Errors, Invocations)

## Configuration Files

### Modified Files

1. **`locals.tf`**
   - Added ECR repository configuration for `api_images_browser`
   - Added Lambda function configuration with environment variables

2. **`main.tf`**
   - Added `api_images_browser` to the lambda_function_arns mapping
   - Ensures the function is included in Step Functions integration

3. **`modules/api_gateway/methods.tf`**
   - Updated existing image browser integration to use new Lambda function
   - Added support for path parameter endpoint `{path+}`
   - Added proper CORS configuration for both endpoints

### New Files

4. **`api-images-browser.tf`**
   - Additional IAM policies for S3 access
   - CloudWatch log group configuration
   - Output values for ECR repository and Lambda function details
   - Example configurations for customization

## Deployment Process

### Prerequisites

1. **Terraform Applied**: Ensure the main Terraform configuration has been applied
2. **ECR Repository**: The ECR repository must exist before pushing images
3. **Lambda Function**: The Lambda function will be created but needs an initial image

### Step 1: Apply Terraform Configuration

```bash
cd product-approach/iac
terraform plan
terraform apply
```

This will create:
- ECR repository for the API Images Browser
- Lambda function (with placeholder image)
- API Gateway endpoints
- IAM roles and policies

### Step 2: Build and Deploy Lambda Function

```bash
cd product-approach/api-function/api_images/browser
./deploy.sh
```

This will:
- Build the Go application and run tests
- Create Docker image
- Push image to ECR
- Update Lambda function with new image
- Test the deployed function

### Step 3: Verify Deployment

Check the API Gateway endpoints:
```bash
# Get the API Gateway URL from Terraform output
API_ENDPOINT=$(cd ../../iac && terraform output -raw api_gateway_endpoint)

# Test the image browser endpoint
curl "$API_ENDPOINT/api/images/browser?bucketType=reference"

# Test with a specific path
curl "$API_ENDPOINT/api/images/browser/2025/01/20?bucketType=checking"
```

## Environment Variables

The Lambda function uses the following environment variables from CONFIG_SECRET:

| Variable | Description | Source |
|----------|-------------|---------|
| `REFERENCE_BUCKET` | Name of the S3 bucket for reference images | Terraform locals |
| `CHECKING_BUCKET` | Name of the S3 bucket for checking images | Terraform locals |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | Static configuration |

## API Endpoints

### GET /api/images/browser

Browse the root directory of a bucket.

**Query Parameters:**
- `bucketType` (optional): `reference` or `checking` (default: `reference`)

**Example:**
```bash
GET /api/images/browser?bucketType=reference
```

### GET /api/images/browser/{path+}

Browse a specific path within a bucket.

**Path Parameters:**
- `path`: URL-encoded path within the bucket

**Query Parameters:**
- `bucketType` (optional): `reference` or `checking` (default: `reference`)

**Example:**
```bash
GET /api/images/browser/layouts/machine-1?bucketType=checking
```

## Response Format

```json
{
  "currentPath": "folder/subfolder",
  "parentPath": "folder",
  "items": [
    {
      "name": "image.jpg",
      "path": "folder/subfolder/image.jpg",
      "type": "image",
      "size": 1024576,
      "lastModified": "2025-01-20T10:30:00Z"
    },
    {
      "name": "documents",
      "path": "folder/subfolder/documents",
      "type": "folder"
    }
  ],
  "totalItems": 2
}
```

## Monitoring and Logging

### CloudWatch Logs
- Log Group: `/aws/lambda/{function-name}`
- Log Level: Configurable via `LOG_LEVEL` environment variable
- Structured JSON logging with correlation IDs

### CloudWatch Metrics
- **Duration**: Function execution time
- **Errors**: Error count and rate
- **Invocations**: Total request count
- **Throttles**: Throttling events

### Optional Alarms
Uncomment the alarm configurations in `api-images-browser.tf` to enable:
- Error rate monitoring
- Duration threshold alerts
- Custom metric alarms

## Customization

### Memory and Timeout
Modify the Lambda function configuration in `locals.tf`:

```hcl
api_images_browser = {
  memory_size = 512  # Increase for better performance
  timeout     = 60   # Increase for large directories
  # ... other configuration
}
```

### Additional Environment Variables
Add custom environment variables in `locals.tf`:

```hcl
environment_variables = {
  REFERENCE_BUCKET = local.s3_buckets.reference
  CHECKING_BUCKET  = local.s3_buckets.checking
  LOG_LEVEL        = "INFO"
  MAX_ITEMS_PER_REQUEST = "500"
  ENABLE_CACHING        = "true"
}
```

### Custom Domain
Uncomment and configure the custom domain section in `api-images-browser.tf`.

## Troubleshooting

### Common Issues

1. **ECR Repository Not Found**
   - Ensure Terraform has been applied successfully
   - Check that ECR repository exists: `aws ecr describe-repositories`

2. **Lambda Function Update Fails**
   - Verify ECR image exists and is accessible
   - Check IAM permissions for Lambda service role

3. **API Gateway 403 Errors**
   - Verify API key configuration if enabled
   - Check CORS settings for web applications

4. **S3 Access Denied**
   - Verify bucket names in environment variables
   - Check IAM policies for S3 access permissions

### Debug Commands

```bash
# Check ECR repositories
aws ecr describe-repositories --query "repositories[?contains(repositoryName, 'api-images-browser')]"

# Check Lambda function
aws lambda get-function --function-name $(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text)

# Test Lambda function directly
aws lambda invoke --function-name <function-name> --payload '{"httpMethod":"GET","path":"/api/images/browser","queryStringParameters":{"bucketType":"reference"}}' response.json

# Check CloudWatch logs
aws logs tail /aws/lambda/<function-name> --follow
```

## Security Considerations

- **Bucket Access**: Function only accesses configured reference and checking buckets
- **Path Traversal**: Input paths are sanitized to prevent directory traversal attacks
- **Authentication**: Relies on API Gateway for authentication and authorization
- **CORS**: Configured for web application access with appropriate origins
- **IAM**: Principle of least privilege with minimal required permissions
