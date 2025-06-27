# API Images View Lambda Function - Terraform Configuration

This document describes the Terraform configuration changes made to add the new `api_images_view` Lambda function to the infrastructure.

## Overview

The `api_images_view` Lambda function implements the `/api/images/{key}/view` endpoint to generate presigned URLs for S3 image viewing. This function resolves the 500 Internal Server Error occurring when the Initiate Verification page tries to load image previews.

## Files Modified

### 1. `locals.tf`

#### ECR Repository Configuration
Added ECR repository configuration for the new Lambda function:

```hcl
api_images_view = {
  name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-images-view", local.name_suffix])))
  image_tag_mutability = "MUTABLE"
  scan_on_push         = true
  force_delete         = false
  encryption_type      = "AES256"
  kms_key              = null
  lifecycle_policy     = null
  repository_policy    = null
},
```

#### Lambda Function Configuration
Added Lambda function configuration:

```hcl
api_images_view = {
  name        = lower(join("-", compact([local.name_prefix, "lambda", "api-images-view", local.name_suffix]))),
  description = "Generate presigned URLs for S3 image viewing via REST API",
  memory_size = 128,
  timeout     = 30,
  image_uri   = "879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-images-view-f6d3xl:latest"
  environment_variables = {
    REFERENCE_BUCKET = local.s3_buckets.reference
    CHECKING_BUCKET  = local.s3_buckets.checking
    LOG_LEVEL        = "INFO"
  }
}
```

### 2. `main.tf`

#### Step Functions Integration
Added the new function ARN to the step functions module:

```hcl
lambda_function_arns = {
  # ... existing functions ...
  api_images_browser            = module.lambda_functions[0].function_arns["api_images_browser"]
  api_images_view               = module.lambda_functions[0].function_arns["api_images_view"]
}
```

### 3. `modules/api_gateway/methods.tf`

#### API Gateway Integration
Updated the existing `/api/images/{key}/view` endpoint to use the new Lambda function:

```hcl
resource "aws_api_gateway_integration" "image_view_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.image_view.id
  http_method             = aws_api_gateway_method.image_view_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["api_images_view"]}/invocations"
}
```

**Note**: Changed from `fetch_images` to `api_images_view` function.

## Infrastructure Components Created

### ECR Repository
- **Name**: `kootoro-dev-ecr-api-images-view-f6d3xl`
- **Purpose**: Store Docker images for the Lambda function
- **Configuration**: Mutable tags, scan on push enabled

### Lambda Function
- **Name**: `kootoro-dev-lambda-api-images-view-f6d3xl`
- **Runtime**: Container (Go)
- **Memory**: 128 MB
- **Timeout**: 30 seconds
- **Environment Variables**:
  - `REFERENCE_BUCKET`: Reference S3 bucket name
  - `CHECKING_BUCKET`: Checking S3 bucket name
  - `LOG_LEVEL`: INFO

### API Gateway Integration
- **Endpoint**: `GET /api/images/{key}/view`
- **Integration**: AWS_PROXY with Lambda function
- **CORS**: Enabled with appropriate headers
- **API Key**: Required (if enabled)

## Environment Variables

The Lambda function receives the following environment variables:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `REFERENCE_BUCKET` | S3 bucket for reference images | `kootoro-dev-s3-reference-f6d3xl` |
| `CHECKING_BUCKET` | S3 bucket for checking images | `kootoro-dev-s3-checking-f6d3xl` |
| `LOG_LEVEL` | Logging level | `INFO` |

## IAM Permissions

The Lambda function inherits permissions from the shared Lambda execution role, which includes:

- **S3 Permissions**: 
  - `s3:GetObject` on reference and checking buckets
  - `s3:HeadObject` for object existence validation
  - `s3:GetObjectVersion` for versioned objects

- **CloudWatch Logs**:
  - `logs:CreateLogGroup`
  - `logs:CreateLogStream`
  - `logs:PutLogEvents`

- **X-Ray Tracing**:
  - `xray:PutTraceSegments`
  - `xray:PutTelemetryRecords`

## Deployment Process

### 1. Apply Terraform Changes
```bash
cd product-approach/iac
terraform plan
terraform apply
```

### 2. Build and Deploy Lambda Function
```bash
cd product-approach/api-function/api_images/view
./deploy.sh
```

### 3. Verify Deployment
```bash
# Test the endpoint
curl -X GET "https://your-api-gateway-endpoint/v1/api/images/test-image.jpg/view?bucketType=reference" \
  -H "x-api-key: your-api-key"
```

## Expected Resources

After applying the Terraform configuration, the following resources will be created:

1. **ECR Repository**: `kootoro-dev-ecr-api-images-view-f6d3xl`
2. **Lambda Function**: `kootoro-dev-lambda-api-images-view-f6d3xl`
3. **CloudWatch Log Group**: `/aws/lambda/kootoro-dev-lambda-api-images-view-f6d3xl`
4. **API Gateway Integration**: Updated to use new Lambda function

## Testing

### Unit Testing
```bash
cd product-approach/api-function/api_images/view
./deploy.sh go-test
```

### Integration Testing
```bash
# Test deployed function
./deploy.sh test

# Manual API testing
curl -X GET "https://api-endpoint/v1/api/images/folder/image.jpg/view" \
  -H "x-api-key: your-key" \
  -H "Content-Type: application/json"
```

## Monitoring

The function can be monitored through:

- **CloudWatch Logs**: `/aws/lambda/kootoro-dev-lambda-api-images-view-f6d3xl`
- **CloudWatch Metrics**: Lambda function metrics
- **X-Ray Tracing**: Request tracing and performance analysis
- **API Gateway Metrics**: Request count, latency, errors

## Troubleshooting

### Common Issues

1. **Function Not Found**: Ensure Terraform apply completed successfully
2. **Permission Denied**: Verify IAM role has S3 access permissions
3. **Image Not Found**: Check that ECR repository exists and contains image
4. **API Gateway 500 Error**: Check Lambda function logs for errors

### Debug Commands

```bash
# Check function exists
aws lambda get-function --function-name kootoro-dev-lambda-api-images-view-f6d3xl

# View recent logs
aws logs tail /aws/lambda/kootoro-dev-lambda-api-images-view-f6d3xl --follow

# Test function directly
aws lambda invoke --function-name kootoro-dev-lambda-api-images-view-f6d3xl \
  --payload '{"httpMethod":"GET","pathParameters":{"key":"test.jpg"}}' \
  response.json
```

## Next Steps

1. Apply Terraform changes to create infrastructure
2. Deploy Lambda function using the deploy script
3. Test the endpoint with real image keys
4. Monitor function performance and logs
5. Update frontend to use the new endpoint (already compatible)

This implementation should resolve the 500 Internal Server Error in the Initiate Verification page when loading image previews.
