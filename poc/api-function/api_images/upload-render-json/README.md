# API Images Upload and Render JSON Lambda

This Lambda function combines two functionalities:
1. **File Upload API** - Handles file uploads to S3 buckets
2. **JSON Layout Rendering** - Automatically renders JSON layout files to images when uploaded to specific paths

## Features

### Upload API
- Supports multiple file types (images, JSON, PDF, text files)
- Configurable upload paths
- Support for both reference and checking buckets
- File size validation (10MB limit)
- CORS support for web applications

### JSON Rendering
- Automatically detects JSON layout files uploaded to configured paths
- Renders vending machine layouts to PNG images
- Stores rendered images in organized S3 structure
- Optional DynamoDB metadata storage
- Comprehensive logging and error handling

## Environment Variables

**Important:** All environment variables must be configured in the AWS Lambda function configuration, not in the Docker image.

### Required
- `REFERENCE_BUCKET` - S3 bucket for reference files
- `CHECKING_BUCKET` - S3 bucket for checking files

### Optional
- `JSON_RENDER_PATH` - S3 path where JSON files should be automatically rendered. Supports two formats:
  - Full S3 URI: `s3://bucket-name/path/` (e.g., `s3://my-reference-bucket/raw/`)
  - Simple path: `raw` (uses REFERENCE_BUCKET, default: "raw")
- `DYNAMODB_LAYOUT_TABLE` - DynamoDB table for storing layout metadata (if not set, metadata storage is skipped)
- `AWS_REGION` - AWS region (default: "us-east-1")
- `LOG_LEVEL` - Logging level (default: "info")

### Example Lambda Environment Configuration
```
REFERENCE_BUCKET=my-vending-machine-reference-bucket
CHECKING_BUCKET=my-vending-machine-checking-bucket
JSON_RENDER_PATH=s3://my-vending-machine-reference-bucket/raw/
DYNAMODB_LAYOUT_TABLE=VendingMachineLayoutMetadata
AWS_REGION=us-east-1
LOG_LEVEL=info
```

### JSON_RENDER_PATH Format Examples
```bash
# Full S3 URI format (recommended)
JSON_RENDER_PATH=s3://kootoro-dev-s3-reference-f6d3xl/raw/

# Simple path format (backward compatibility)
JSON_RENDER_PATH=raw

# Different bucket and path
JSON_RENDER_PATH=s3://my-special-bucket/layouts/input/
```

## API Usage

### Upload Files

**Endpoint:** `POST /upload`

**Query Parameters:**
- `bucketType` - "reference" or "checking" (default: "reference")
- `path` - Upload path within bucket (optional)
- `fileName` - Name of the file being uploaded (required)

**Request:**
- Content-Type: `multipart/form-data`
- Body: File content

**Response:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "files": [
    {
      "originalName": "layout.json",
      "key": "raw/layout.json",
      "size": 1024,
      "contentType": "application/json",
      "bucket": "my-reference-bucket"
    }
  ],
  "renderResult": {
    "rendered": true,
    "layoutId": 12345,
    "layoutPrefix": "20240101-120000-ABC12",
    "processedKey": "processed/2024/01/01/12345_20240101-120000-ABC12_reference_image.png",
    "message": "Layout rendered successfully"
  }
}
```

## Automatic JSON Rendering

When a JSON file is uploaded to the configured render path (default: `/raw`), the system will:

1. **Validate** the JSON file as a valid layout structure
2. **Check** file size limits (10MB max)
3. **Parse** the layout data
4. **Render** the layout to a PNG image
5. **Upload** the rendered image to S3 with organized path structure
6. **Store** metadata in DynamoDB (if configured)

### Layout JSON Structure

The JSON file must contain a valid vending machine layout structure:

```json
{
  "layoutId": 12345,
  "subLayoutList": [
    {
      "trayList": [
        {
          "trayCode": "A",
          "trayNo": 1,
          "slotList": [
            {
              "vmLayoutSlotId": 1,
              "productId": 100,
              "productTemplateId": 200,
              "maxQuantity": 10,
              "slotNo": 1,
              "status": 1,
              "position": "A1",
              "slotIndexCode": 1,
              "cellNumber": 1,
              "productTemplateName": "Coca Cola",
              "productTemplateImage": "https://example.com/image.jpg"
            }
          ]
        }
      ]
    }
  ]
}
```

### Rendered Image Output

Rendered images are stored with the following path structure:
```
processed/{year}/{month}/{date}/{layoutId}_{layoutPrefix}_reference_image.png
```

Example: `processed/2024/01/15/12345_20240115-143022-XYZ89_reference_image.png`

## Deployment

### Build Docker Image
```bash
docker build -t api-upload-render .
```

### Deploy to AWS Lambda
1. Push image to ECR
2. Create Lambda function with container image
3. **Configure environment variables in Lambda console** (see Environment Variables section above)
4. Set appropriate timeout (recommended: 300 seconds)
5. Set appropriate memory (recommended: 1024 MB)
6. Configure API Gateway integration

### Required IAM Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:HeadObject"
      ],
      "Resource": [
        "arn:aws:s3:::your-reference-bucket/*",
        "arn:aws:s3:::your-checking-bucket/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:PutItem",
        "dynamodb:GetItem"
      ],
      "Resource": "arn:aws:dynamodb:region:account:table/VendingMachineLayoutMetadata"
    }
  ]
}
```

## Error Handling

The API provides detailed error responses for various scenarios:
- Invalid file types
- File size exceeded
- Invalid JSON structure
- S3 upload failures
- Rendering errors

## Logging

All operations are logged with structured JSON format including:
- Request details
- Upload progress
- Render process steps
- Error conditions
- Performance metrics
