# API Images Upload Lambda Function

This Lambda function handles file uploads to S3 buckets for the vending machine verification system.

## Overview

The upload function accepts POST requests to upload files to either the reference bucket or checking bucket, as specified by environment variables.

## Environment Variables

- `REFERENCE_BUCKET`: S3 bucket for reference images
- `CHECKING_BUCKET`: S3 bucket for checking images
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR)

## API Endpoints

### POST /api/images/upload

Upload a file to the specified S3 bucket.

#### Query Parameters

- `bucketType` (optional): Target bucket type - "reference" or "checking" (default: "reference")
- `fileName` (required): Name of the file to upload
- `path` (optional): Directory path within the bucket to organize uploads

#### Request Headers

- `Content-Type`: Should be set appropriately for the file type

#### Request Body

The file content as binary data.

#### Response

```json
{
  "success": true,
  "message": "File uploaded successfully",
  "files": [
    {
      "originalName": "example.jpg",
      "key": "path/to/example.jpg",
      "size": 12345,
      "contentType": "image/jpeg",
      "bucket": "my-reference-bucket"
    }
  ]
}
```

#### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "errors": ["Detailed error messages"]
}
```

## Supported File Types

- Images: .jpg, .jpeg, .png, .gif, .bmp, .webp, .tiff, .tif
- Documents: .pdf, .txt, .json, .csv, .xml

## File Size Limits

- Maximum file size: 10MB (AWS Lambda limit)

## Usage Examples

### Upload to Reference Bucket

```bash
curl -X POST "https://api.example.com/api/images/upload?bucketType=reference&fileName=test.jpg" \
  -H "Content-Type: image/jpeg" \
  --data-binary @test.jpg
```

### Upload to Checking Bucket with Path

```bash
curl -X POST "https://api.example.com/api/images/upload?bucketType=checking&fileName=sample.png&path=products/category1" \
  -H "Content-Type: image/png" \
  --data-binary @sample.png
```

## Development

### Local Development

1. Set environment variables:
```bash
export REFERENCE_BUCKET=your-reference-bucket
export CHECKING_BUCKET=your-checking-bucket
export LOG_LEVEL=INFO
```

2. Run locally:
```bash
./deploy.sh go-run
```

### Building

```bash
./deploy.sh go-build
```

### Testing

```bash
./deploy.sh go-test
```

### Deployment

```bash
./deploy.sh deploy
```

## Architecture

The function follows the same structure as other API functions in this project:

- **main.go**: Lambda handler and core logic
- **Dockerfile**: Multi-stage Docker build
- **deploy.sh**: Deployment automation script
- **go.mod**: Go module dependencies

## Security

- CORS headers are configured for cross-origin requests
- File type validation prevents unauthorized file uploads
- File size limits prevent abuse
- Path sanitization prevents directory traversal attacks

## Error Handling

The function provides detailed error responses for:
- Invalid file types
- File size exceeded
- Missing required parameters
- S3 upload failures
- Invalid bucket types

## Logging

All operations are logged with structured logging using logrus:
- Request details
- Upload progress
- Error conditions
- Performance metrics
