# API Images View Lambda Function

This AWS Lambda function implements the `/api/images/{key}/view` endpoint to generate presigned URLs for S3 image viewing.

## Purpose

The function generates temporary presigned URLs that allow frontend applications to securely access images stored in S3 buckets without requiring direct AWS credentials. This is essential for the Initiate Verification page to display image previews and resolves the current 500 Internal Server Error when loading images.

## Endpoint

- **Method**: GET
- **Path**: `/api/images/{key}/view`
- **Path Parameter**: `{key}` - The S3 object key/path for the image file
- **Query Parameter**: `bucketType` - Either "reference" or "checking" (defaults to "reference")

## Request Examples

```bash
# View an image from the reference bucket
GET /api/images/folder/image.jpg/view

# View an image from the checking bucket
GET /api/images/uploads/2024/01/15/abc123_photo.png/view?bucketType=checking

# Handle special characters in filenames
GET /api/images/folder/image%20with%20spaces.jpg/view

# Real example with actual bucket structure
GET /api/images/products/electronics/smartphone.jpg/view?bucketType=reference
```

## Response Format

### Success Response (200)
```json
{
  "presignedUrl": "https://kootoro-dev-s3-reference-f6d3xl.s3.amazonaws.com/folder/image.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=..."
}
```

### Error Responses

#### 400 Bad Request - Missing Key
```json
{
  "error": "Missing key parameter",
  "message": "The image key must be provided in the URL path"
}
```

#### 400 Bad Request - Invalid Bucket Type
```json
{
  "error": "Invalid bucket type",
  "message": "bucketType must be 'reference' or 'checking'"
}
```

#### 404 Not Found
```json
{
  "error": "Image not found",
  "message": "The requested image does not exist or is not accessible"
}
```

#### 500 Internal Server Error - Configuration
```json
{
  "error": "Bucket configuration error",
  "message": "The reference bucket is not configured"
}
```

#### 500 Internal Server Error - Presigned URL
```json
{
  "error": "Failed to generate presigned URL",
  "message": "Unable to create temporary access URL for the image"
}
```

## Features

### URL Encoding Support
- Automatically handles URL-encoded keys with special characters
- Supports spaces, unicode characters, and other special symbols in filenames
- Properly decodes path parameters before S3 operations

### Multi-Bucket Support
- Supports both reference and checking buckets
- Bucket selection via `bucketType` query parameter
- Dynamic bucket name resolution from environment variables
- Validates bucket configuration at runtime

### Security
- Presigned URLs expire after 1 hour (configurable)
- CORS headers included for frontend access
- Object existence validation before URL generation
- No direct S3 credentials exposed to frontend

### Error Handling
- Comprehensive error responses with appropriate HTTP status codes
- Structured logging for debugging and monitoring
- Graceful handling of missing files and access errors
- Detailed error messages for troubleshooting

### CORS Support
- Includes all necessary CORS headers for frontend integration
- Supports preflight OPTIONS requests
- Allows credentials for authenticated requests
- Compatible with Streamlit frontend requirements

## Environment Variables

The function requires the following environment variables:

- `REFERENCE_BUCKET` - Name of the S3 bucket containing reference images (e.g., "kootoro-dev-s3-reference-f6d3xl")
- `CHECKING_BUCKET` - Name of the S3 bucket containing checking images (e.g., "kootoro-dev-s3-checking-f6d3xl")
- `LOG_LEVEL` - Logging level (debug, info, warn, error) - defaults to info

## Dependencies

- **AWS SDK for Go v2** - Latest S3 client with presigned URL support
- **AWS Lambda Go runtime** - Native Lambda execution environment
- **Logrus** - Structured JSON logging
- **Go 1.20+** - Modern Go runtime with generics support

## Building and Deployment

### Quick Start
```bash
# Full deployment (recommended)
./deploy.sh

# Or using make
make deploy
```

### Development Commands
```bash
# Build and test locally
./deploy.sh go-build
./deploy.sh go-test

# Format code
./deploy.sh go-fmt

# Clean up
./deploy.sh go-clean
```

### Deployment Commands
```bash
# Build Docker image only
./deploy.sh build

# Build and push to ECR
./deploy.sh push

# Update Lambda function only
./deploy.sh update

# Test deployed function
./deploy.sh test

# Show help
./deploy.sh help
```

### Manual Build Steps
```bash
# Download dependencies
go mod download && go mod tidy

# Build binary
go build -o api-images-view *.go

# Build Docker image
docker build -t api-images-view .

# Run tests
REFERENCE_BUCKET=test-ref CHECKING_BUCKET=test-check go test -v
```

## Testing

### Unit Tests
```bash
./deploy.sh go-test
```

### Integration Testing
```bash
# Test deployed function
./deploy.sh test

# Manual Lambda invocation
aws lambda invoke \
  --function-name kootoro-dev-lambda-api-images-view-f6d3xl \
  --payload '{"httpMethod":"GET","path":"/api/images/test.jpg/view","pathParameters":{"key":"test.jpg"},"queryStringParameters":{"bucketType":"reference"}}' \
  response.json
```

### Local Testing
```bash
# Set environment variables
export REFERENCE_BUCKET=kootoro-dev-s3-reference-f6d3xl
export CHECKING_BUCKET=kootoro-dev-s3-checking-f6d3xl
export LOG_LEVEL=debug

# Run locally (requires Lambda runtime emulator)
./deploy.sh go-run
```

## Integration

This function integrates with:

### Frontend Applications
- **Streamlit App**: Image browser and verification pages
- **API Client**: `get_image_url()` method in `api_client.py`
- **Image Display**: Direct integration with `st.image()` components

### AWS Services
- **S3 Buckets**: Reference and checking image storage
- **API Gateway**: RESTful API routing and request handling
- **Lambda**: Serverless execution environment
- **CloudWatch**: Logging and monitoring
- **ECR**: Container image storage

### Expected Integration Flow
1. Frontend calls `api_client.get_image_url(key)`
2. API Gateway routes to Lambda function
3. Lambda validates key and bucket configuration
4. Lambda checks S3 object existence
5. Lambda generates presigned URL with 1-hour expiration
6. Frontend receives `presignedUrl` in JSON response
7. Frontend displays image using presigned URL

## Troubleshooting

### Common Issues

#### 500 Internal Server Error
- **Cause**: Bucket environment variables not configured
- **Solution**: Verify `REFERENCE_BUCKET` and `CHECKING_BUCKET` in Lambda environment
- **Check**: `aws lambda get-function-configuration --function-name <function-name>`

#### 404 Not Found
- **Cause**: Image key doesn't exist in specified bucket
- **Solution**: Verify the image exists using AWS CLI: `aws s3 ls s3://bucket-name/path/`
- **Check**: Ensure correct bucket type parameter

#### CORS Errors
- **Cause**: Missing or incorrect CORS headers
- **Solution**: Function includes all required CORS headers automatically
- **Check**: Verify API Gateway CORS configuration

#### URL Encoding Issues
- **Cause**: Special characters in filenames not properly encoded
- **Solution**: Function automatically handles URL decoding
- **Check**: Ensure frontend properly encodes special characters

### Logging and Monitoring

#### Structured Logging Fields
- `original_key`: The URL-encoded key from the request
- `decoded_key`: The decoded S3 object key
- `bucket_name`: The target S3 bucket
- `bucket_type`: Either "reference" or "checking"
- `presigned_url`: The generated presigned URL (in success cases)
- `error`: Error details for failed requests

#### CloudWatch Monitoring
```bash
# View recent logs
aws logs tail /aws/lambda/kootoro-dev-lambda-api-images-view-f6d3xl --follow

# Check function metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Invocations \
  --dimensions Name=FunctionName,Value=kootoro-dev-lambda-api-images-view-f6d3xl \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 3600 \
  --statistics Sum
```

#### Debug Mode
```bash
# Enable debug logging
aws lambda update-function-configuration \
  --function-name kootoro-dev-lambda-api-images-view-f6d3xl \
  --environment Variables='{REFERENCE_BUCKET=bucket1,CHECKING_BUCKET=bucket2,LOG_LEVEL=debug}'
```

## Performance Considerations

- **Cold Start**: ~100-200ms for Go Lambda functions
- **Warm Execution**: ~10-50ms per request
- **Presigned URL Generation**: ~5-10ms
- **S3 HeadObject Check**: ~20-50ms
- **Memory Usage**: 128MB recommended (configurable)
- **Timeout**: 30 seconds (more than sufficient)

## Security Considerations

- **Presigned URLs**: 1-hour expiration limits exposure window
- **Object Validation**: Prevents generation of URLs for non-existent objects
- **Bucket Isolation**: Separate reference and checking buckets
- **No Credentials**: Frontend never receives AWS credentials
- **CORS**: Properly configured for cross-origin requests
- **Error Handling**: No sensitive information leaked in error messages

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for version history and updates.
