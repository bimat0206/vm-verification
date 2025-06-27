# API Verifications List Lambda Function - Terraform Configuration

This document describes the Terraform configuration for the new API Verifications List Lambda function that provides the `/api/verifications` GET endpoint.

## Overview

The API Verifications List Lambda function is a Go-based AWS Lambda function that provides a REST API endpoint for listing verification results with advanced filtering, pagination, and sorting capabilities. It integrates with DynamoDB to efficiently query verification records and supports the frontend Streamlit application.

## Terraform Resources Created

### 1. ECR Repository

**Resource**: `api_verifications_list` in `locals.tf`

```hcl
api_verifications_list = {
  name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-verifications-list", local.name_suffix])))
  image_tag_mutability = "MUTABLE"
  scan_on_push         = true
  force_delete         = false
  encryption_type      = "AES256"
  kms_key              = null
  lifecycle_policy     = null
  repository_policy    = null
}
```

**Purpose**: Stores the Docker container image for the Lambda function.

### 2. Lambda Function

**Resource**: `api_verifications_list` in `locals.tf`

```hcl
api_verifications_list = {
  name        = lower(join("-", compact([local.name_prefix, "lambda", "api-verifications-list", local.name_suffix])))
  description = "API endpoint for listing verification results with filtering and pagination"
  memory_size = 512
  timeout     = 30
  environment_variables = {
    VERIFICATION_TABLE = local.dynamodb_tables.verification_results
    LOG_LEVEL          = "INFO"
  }
}
```

**Configuration Details**:
- **Memory**: 512 MB (increased from default 256 MB for better performance)
- **Timeout**: 30 seconds
- **Environment Variables**:
  - `VERIFICATION_TABLE`: DynamoDB table name for verification results
  - `LOG_LEVEL`: Logging level (INFO, DEBUG, WARN, ERROR)

### 3. API Gateway Integration

**Resource**: Updated `verifications_get` integration in `modules/api_gateway/methods.tf`

```hcl
resource "aws_api_gateway_integration" "verifications_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verifications.id
  http_method             = aws_api_gateway_method.verifications_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["api_verifications_list"]}/invocations"
}
```

**Endpoint**: `GET /api/verifications`

**Query Parameters Supported**:
- `verificationStatus` (CORRECT, INCORRECT)
- `vendingMachineId` (string)
- `fromDate` and `toDate` (RFC3339 format)
- `limit` (1-100, default: 20)
- `offset` (default: 0)
- `sortBy` (verificationAt:desc/asc, overallAccuracy:desc/asc)

### 4. API Gateway Model

**Resource**: Updated `verification_list` model in `modules/api_gateway/models.tf`

The model has been enhanced to match the Go Lambda function's response structure:

```json
{
  "results": [
    {
      "verificationId": "string",
      "verificationAt": "string",
      "verificationStatus": "CORRECT|INCORRECT",
      "verificationType": "string",
      "vendingMachineId": "string",
      "referenceImageUrl": "string",
      "checkingImageUrl": "string",
      "layoutId": "integer|null",
      "layoutPrefix": "string|null",
      "overallAccuracy": "number|null",
      "correctPositions": "integer|null",
      "discrepantPositions": "integer|null",
      "result": "object|null",
      "verificationSummary": "object|null",
      "createdAt": "string|null",
      "updatedAt": "string|null"
    }
  ],
  "pagination": {
    "total": "integer",
    "limit": "integer",
    "offset": "integer",
    "nextOffset": "integer|null"
  }
}
```

## IAM Permissions

The Lambda function inherits permissions from the existing Lambda IAM role, which includes:

- **DynamoDB Permissions**:
  - `dynamodb:Query` - For GSI queries
  - `dynamodb:Scan` - For table scans when needed
  - `dynamodb:GetItem` - For individual record retrieval

- **CloudWatch Logs Permissions**:
  - `logs:CreateLogGroup`
  - `logs:CreateLogStream`
  - `logs:PutLogEvents`

## DynamoDB Integration

The function uses the existing verification results table with the following access patterns:

### Primary Access Pattern
- **Table**: `verification_results`
- **Primary Key**: `verificationId` (Hash) + `verificationAt` (Range)

### GSI Access Pattern
- **Index**: `VerificationStatusIndex`
- **Key**: `verificationStatus` (Hash) + `verificationAt` (Range)
- **Usage**: Efficient filtering by verification status

### Query Strategy
1. **Status-based queries**: Uses `VerificationStatusIndex` GSI for optimal performance
2. **General queries**: Falls back to table scan for flexibility
3. **Additional filtering**: Applied at application level for complex criteria

## Deployment Process

### 1. Infrastructure Deployment

```bash
# Navigate to IAC directory
cd product-approach/iac

# Plan the changes
terraform plan

# Apply the changes
terraform apply
```

### 2. Lambda Function Deployment

```bash
# Navigate to Lambda function directory
cd product-approach/api-function/api_verifications/list

# Deploy the function
./deploy.sh deploy
```

### 3. Verification

```bash
# Test the deployed function
./deploy.sh test

# Or test via API Gateway
curl -X GET "https://your-api-gateway-url/api/verifications?limit=5"
```

## Environment Variables

The Lambda function requires the following environment variables:

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `VERIFICATION_TABLE` | DynamoDB verification results table name | Yes | Set by Terraform |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | No | INFO |

## Performance Considerations

### Memory and Timeout
- **Memory**: 512 MB (optimized for DynamoDB operations and JSON processing)
- **Timeout**: 30 seconds (sufficient for most query operations)

### DynamoDB Optimization
- Uses GSI for status-based queries to minimize read capacity consumption
- Implements efficient pagination to handle large result sets
- Applies filtering at DynamoDB level where possible

### Caching Strategy
- Consider implementing caching for frequently accessed data
- Use DynamoDB DAX for microsecond latency if needed

## Monitoring and Logging

### CloudWatch Metrics
Monitor the following metrics:
- `Duration` - Function execution time
- `Errors` - Error count
- `Invocations` - Total invocation count
- `Throttles` - Throttling events

### Structured Logging
The function uses structured JSON logging with the following levels:
- **DEBUG**: Detailed DynamoDB operation information
- **INFO**: Request/response information and successful operations
- **WARN**: Non-critical issues (e.g., unmarshaling errors)
- **ERROR**: Critical errors and failures

### Log Analysis
```bash
# View recent logs
aws logs tail /aws/lambda/your-function-name --follow

# Filter for errors
aws logs filter-log-events --log-group-name /aws/lambda/your-function-name --filter-pattern "ERROR"
```

## Security Considerations

### API Gateway Security
- API key authentication (if enabled)
- CORS headers properly configured
- Input validation through API Gateway models

### Lambda Security
- Minimal IAM permissions following principle of least privilege
- Environment variables for sensitive configuration
- VPC configuration (if required for network isolation)

### Data Security
- DynamoDB encryption at rest
- CloudWatch logs encryption
- Secure handling of verification data

## Troubleshooting

### Common Issues

1. **"Table not found" errors**
   - Verify `VERIFICATION_TABLE` environment variable
   - Check DynamoDB table exists and is accessible

2. **"Access denied" errors**
   - Verify IAM permissions for DynamoDB access
   - Check Lambda execution role

3. **Timeout errors**
   - Monitor function duration
   - Optimize DynamoDB queries
   - Consider increasing timeout if needed

4. **Empty responses**
   - Verify verification records exist in DynamoDB
   - Check filter criteria
   - Review CloudWatch logs for errors

### Debug Mode
Enable debug logging by setting `LOG_LEVEL=DEBUG` to see detailed operation logs.

## Future Enhancements

### Planned Improvements
- **Caching**: Implement caching for frequently accessed data
- **Advanced Sorting**: Enhanced sorting capabilities for complex use cases
- **Batch Operations**: Support for batch queries and operations
- **Real-time Updates**: Integration with DynamoDB Streams
- **Analytics**: Enhanced analytics and reporting capabilities

### Performance Optimization
- Consider DynamoDB DAX for microsecond latency
- Implement connection pooling for better performance
- Add compression for large responses
- Optimize JSON serialization/deserialization
