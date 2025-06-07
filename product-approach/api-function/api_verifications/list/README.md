# API Verifications List Lambda Function

This is a Go-based AWS Lambda function that provides a REST API for listing verification results with filtering and pagination capabilities. It serves as the backend for the `/api/verifications` endpoint, enabling users to query and browse verification records stored in DynamoDB.

## Features

- **DynamoDB Integration**: Efficiently queries verification records from DynamoDB tables
- **Advanced Filtering**: Support for filtering by verification status, vending machine ID, and date ranges
- **Pagination**: Built-in pagination with configurable limits and offset-based navigation
- **Sorting**: Support for sorting by verification date and overall accuracy
- **Performance Optimization**: Uses DynamoDB GSI (Global Secondary Index) for efficient queries
- **CORS Support**: Properly configured for web application integration
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes
- **Structured Logging**: Detailed logging with configurable log levels

## API Endpoint

### GET `/api/verifications`

List verification results with optional filtering, pagination, and sorting.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `verificationStatus` | string | No | - | Filter by verification status (`CORRECT` or `INCORRECT`) |
| `vendingMachineId` | string | No | - | Filter by specific vending machine ID |
| `fromDate` | string | No | - | Filter results from this date (RFC3339 format) |
| `toDate` | string | No | - | Filter results until this date (RFC3339 format) |
| `limit` | integer | No | `20` | Number of results per page (1-100) |
| `offset` | integer | No | `0` | Number of results to skip for pagination |
| `sortBy` | string | No | `verificationAt:desc` | Sort order (`verificationAt:desc`, `verificationAt:asc`, `overallAccuracy:desc`, `overallAccuracy:asc`) |

#### Response Format

```json
{
  "results": [
    {
      "verificationId": "a041e458-3171-43e9-a149-f63c5916d3a2",
      "verificationAt": "2025-01-02T10:30:00Z",
      "verificationStatus": "CORRECT",
      "verificationType": "LAYOUT_VS_CHECKING",
      "vendingMachineId": "VM-001",
      "referenceImageUrl": "s3://bucket/reference.jpg",
      "checkingImageUrl": "s3://bucket/checking.jpg",
      "layoutId": 12345,
      "layoutPrefix": "test_prefix",
      "overallAccuracy": 95.5,
      "correctPositions": 18,
      "discrepantPositions": 2,
      "result": {
        "outcome": "CORRECT",
        "confidence": 0.95
      },
      "verificationSummary": {
        "totalPositions": 20,
        "accuracyScore": 95.5
      },
      "createdAt": "2025-01-02T10:25:00Z",
      "updatedAt": "2025-01-02T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "nextOffset": 20
  }
}
```

#### Example Requests

```bash
# Get all verifications (default pagination)
GET /api/verifications

# Filter by verification status
GET /api/verifications?verificationStatus=CORRECT

# Filter by vending machine and date range
GET /api/verifications?vendingMachineId=VM-001&fromDate=2025-01-01T00:00:00Z&toDate=2025-01-02T23:59:59Z

# Paginated results with custom limit
GET /api/verifications?limit=50&offset=100

# Sort by accuracy (highest first)
GET /api/verifications?sortBy=overallAccuracy:desc

# Combined filters
GET /api/verifications?verificationStatus=INCORRECT&limit=10&sortBy=verificationAt:asc
```

## Environment Variables

The function requires the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `DYNAMODB_VERIFICATION_TABLE` | Name of the DynamoDB verification results table | Yes |
| `DYNAMODB_CONVERSATION_TABLE` | Name of the DynamoDB conversation table | Yes |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | No (default: INFO) |

## DynamoDB Table Structure

The function expects a DynamoDB table with the following structure:

### Primary Key
- **Hash Key**: `verificationId` (String)
- **Range Key**: `verificationAt` (String)

### Global Secondary Indexes (GSI)
- **VerificationStatusIndex**: `verificationStatus` (Hash) + `verificationAt` (Range)
- **VerificationTypeIndex**: `verificationType` (Hash) + `verificationAt` (Range)
- **CheckingImageIndex**: `checkingImageUrl` (Hash) + `verificationAt` (Range)
- **ReferenceImageIndex**: `referenceImageUrl` (Hash) + `verificationAt` (Range)
- **LayoutIndex**: `layoutId` (Hash) + `verificationAt` (Range)

## IAM Permissions

The Lambda function requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:GetItem"
      ],
      "Resource": [
        "arn:aws:dynamodb:${AWS_REGION}:${AWS_ACCOUNT_ID}:table/${DYNAMODB_VERIFICATION_TABLE}",
        "arn:aws:dynamodb:${AWS_REGION}:${AWS_ACCOUNT_ID}:table/${DYNAMODB_VERIFICATION_TABLE}/index/*",
        "arn:aws:dynamodb:${AWS_REGION}:${AWS_ACCOUNT_ID}:table/${DYNAMODB_CONVERSATION_TABLE}",
        "arn:aws:dynamodb:${AWS_REGION}:${AWS_ACCOUNT_ID}:table/${DYNAMODB_CONVERSATION_TABLE}/index/*"
      ]
    }
  ]
}
```

## Building and Deployment

### Prerequisites

- Go 1.20 or higher
- Docker
- AWS CLI configured with appropriate permissions
- Access to AWS ECR repository

### Local Build

```bash
# Install dependencies
go mod download

# Build the binary
go build -o api-verifications-list

# Run tests
go test ./...
```

### Docker Build

```bash
# Build Docker image
docker build -t api-verifications-list .

# Test locally with Docker
docker run -e DYNAMODB_VERIFICATION_TABLE=my-verification-table \
           -e DYNAMODB_CONVERSATION_TABLE=my-conversation-table \
           -e LOG_LEVEL=INFO \
           api-verifications-list
```

### Deploy to AWS Lambda

Use the provided deployment script:

```bash
# Full deployment (build, push, update)
./deploy.sh

# Build and push to ECR only
./deploy.sh push

# Update Lambda function only
./deploy.sh update

# Test deployed function
./deploy.sh test
```

## Performance Considerations

### Query Optimization

1. **Status-based queries**: Uses `VerificationStatusIndex` GSI for efficient filtering
2. **Date range filtering**: Applied at the DynamoDB level when possible
3. **Machine ID filtering**: Applied as filter expressions to reduce data transfer
4. **Pagination**: Implemented at the application level after DynamoDB queries

### Scaling

- **DynamoDB**: Uses on-demand billing mode for automatic scaling
- **Lambda**: Configured with appropriate memory and timeout settings
- **Caching**: Consider implementing caching for frequently accessed data

## Error Responses

The API returns structured error responses:

```json
{
  "error": "Invalid query parameters",
  "message": "invalid limit: must be between 1 and 100",
  "code": "HTTP_400"
}
```

### Common Error Codes

- **400**: Bad Request (invalid parameters, malformed dates)
- **405**: Method Not Allowed (non-GET request)
- **500**: Internal Server Error (DynamoDB errors, processing failures)

## Monitoring and Logging

### CloudWatch Metrics

Monitor the following CloudWatch metrics:

- `Duration`: Function execution time
- `Errors`: Error count
- `Invocations`: Total invocation count
- `Throttles`: Throttling events

### Structured Logging

The function uses structured logging with the following log levels:

- **DEBUG**: Detailed DynamoDB operation information
- **INFO**: Request/response information and successful operations
- **WARN**: Non-critical issues (e.g., unmarshaling errors)
- **ERROR**: Critical errors and failures

## Local Development

For local testing, you can use the AWS Lambda Runtime Interface Emulator:

```bash
# Build the container
docker build -t api-verifications-list .

# Run with Lambda RIE
docker run -p 9000:8080 \
  -e DYNAMODB_VERIFICATION_TABLE=test-verification-table \
  -e DYNAMODB_CONVERSATION_TABLE=test-conversation-table \
  -e LOG_LEVEL=DEBUG \
  api-verifications-list

# Test the function
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"httpMethod":"GET","path":"/api/verifications","queryStringParameters":{"limit":"5"}}'
```

## Troubleshooting

### Common Issues

1. **"Table not found" errors**: Verify table name in environment variables
2. **"Access denied" errors**: Check IAM permissions for DynamoDB access
3. **"Invalid parameters" errors**: Ensure query parameters match expected format
4. **Empty responses**: Check if verification records exist in the table

### Debug Mode

Enable debug logging by setting `LOG_LEVEL=DEBUG` to see detailed DynamoDB operation logs.
