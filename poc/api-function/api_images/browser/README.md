# API Images Browser Lambda Function

This is a Go-based AWS Lambda function that provides a REST API for browsing images stored in S3 buckets. It serves as the backend for the `/api/images/browser` endpoint, enabling users to navigate through S3 bucket contents in a file-browser-like interface.

## Features

- **S3 Bucket Browsing**: Navigate through folders and files in S3 buckets
- **Multi-Bucket Support**: Browse both reference and checking buckets
- **Image Detection**: Automatically identifies image files and marks them appropriately
- **Folder Navigation**: Support for hierarchical folder structure with parent navigation
- **CORS Support**: Properly configured for web application integration
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes
- **Logging**: Structured logging with configurable log levels

## API Endpoint

### GET `/api/images/browser/{path+}`

Browse S3 bucket contents at the specified path.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `bucketType` | string | No | `reference` | Type of bucket to browse (`reference` or `checking`) |

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | No | Path within the bucket to browse (URL-encoded) |

#### Response Format

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
      "lastModified": "2025-01-20T10:30:00Z",
      "contentType": "image/jpeg"
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

#### Item Types

- **`folder`**: Directory/prefix in S3
- **`image`**: Image files (jpg, jpeg, png, gif, bmp, webp, tiff, tif)
- **`file`**: Other file types

#### Example Requests

```bash
# Browse root of reference bucket
GET /api/images/browser?bucketType=reference

# Browse specific folder in checking bucket
GET /api/images/browser/2025/01/20?bucketType=checking

# Browse nested folder structure
GET /api/images/browser/layouts/vending-machine-1/images?bucketType=reference
```

## Environment Variables

The function requires the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `REFERENCE_BUCKET` | Name of the S3 bucket for reference images | Yes |
| `CHECKING_BUCKET` | Name of the S3 bucket for checking images | Yes |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | No (default: INFO) |

## IAM Permissions

The Lambda function requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::${REFERENCE_BUCKET}",
        "arn:aws:s3:::${REFERENCE_BUCKET}/*",
        "arn:aws:s3:::${CHECKING_BUCKET}",
        "arn:aws:s3:::${CHECKING_BUCKET}/*"
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
go build -o api-images-browser

# Run tests (if any)
go test ./...
```

### Docker Build

```bash
# Build Docker image
docker build -t api-images-browser .

# Test locally with Docker
docker run -e REFERENCE_BUCKET=my-ref-bucket \
           -e CHECKING_BUCKET=my-check-bucket \
           -e LOG_LEVEL=INFO \
           api-images-browser
```

### Deploy to AWS Lambda

1. **Build and push to ECR**:
```bash
# Tag for ECR
docker tag api-images-browser:latest ${ECR_REPO}:latest

# Push to ECR
docker push ${ECR_REPO}:latest
```

2. **Update Lambda function**:
```bash
# Update function code
aws lambda update-function-code \
  --function-name api-images-browser \
  --image-uri ${ECR_REPO}:latest
```

3. **Set environment variables**:
```bash
aws lambda update-function-configuration \
  --function-name api-images-browser \
  --environment Variables='{
    "REFERENCE_BUCKET":"your-reference-bucket",
    "CHECKING_BUCKET":"your-checking-bucket",
    "LOG_LEVEL":"INFO"
  }'
```

## Integration with API Gateway

This Lambda function is designed to be integrated with AWS API Gateway using the following configuration:

### Resource Configuration

- **Resource Path**: `/api/images/browser/{proxy+}`
- **HTTP Method**: `GET`, `OPTIONS`
- **Integration Type**: Lambda Proxy Integration
- **Lambda Function**: `api-images-browser`

### CORS Configuration

The function handles CORS internally and returns appropriate headers:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token`

## Error Responses

The API returns structured error responses:

```json
{
  "error": "Invalid bucket type",
  "message": "Bucket type must be 'reference' or 'checking', got: invalid",
  "code": "HTTP_400"
}
```

### Common Error Codes

- **400**: Bad Request (invalid parameters)
- **404**: Not Found (bucket or path doesn't exist)
- **405**: Method Not Allowed (non-GET request)
- **500**: Internal Server Error (AWS service errors)

## Security Considerations

- **Bucket Access**: Function only accesses configured reference and checking buckets
- **Path Traversal**: Input paths are sanitized to prevent directory traversal attacks
- **Authentication**: Relies on API Gateway for authentication and authorization
- **CORS**: Configured for web application access

## Monitoring and Logging

The function uses structured logging with the following log levels:

- **DEBUG**: Detailed S3 operation information
- **INFO**: Request/response information and successful operations
- **WARN**: Non-critical issues (e.g., invalid file types)
- **ERROR**: Critical errors and failures

### CloudWatch Metrics

Monitor the following CloudWatch metrics:

- `Duration`: Function execution time
- `Errors`: Error count
- `Invocations`: Total invocation count
- `Throttles`: Throttling events

## Local Development

For local testing, you can use the AWS Lambda Runtime Interface Emulator:

```bash
# Build the container
docker build -t api-images-browser .

# Run with Lambda RIE
docker run -p 9000:8080 \
  -e REFERENCE_BUCKET=test-ref-bucket \
  -e CHECKING_BUCKET=test-check-bucket \
  -e LOG_LEVEL=DEBUG \
  api-images-browser

# Test the function
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"httpMethod":"GET","path":"/api/images/browser","queryStringParameters":{"bucketType":"reference"}}'
```

## Troubleshooting

### Common Issues

1. **"Bucket not found" errors**: Verify bucket names in environment variables
2. **"Access denied" errors**: Check IAM permissions for S3 access
3. **"Invalid bucket type" errors**: Ensure bucketType parameter is 'reference' or 'checking'
4. **Empty responses**: Check if the specified path exists in the bucket

### Debug Mode

Enable debug logging by setting `LOG_LEVEL=DEBUG` to see detailed S3 operation logs.
