# Lambda Configuration Guide

## Environment Variables Setup

After deploying the Lambda function, you **must** configure the following environment variables in the AWS Lambda console:

### Step 1: Navigate to Lambda Function
1. Go to AWS Lambda console
2. Find your `api-images-upload-render` function
3. Click on "Configuration" tab
4. Click on "Environment variables" in the left sidebar
5. Click "Edit"

### Step 2: Add Required Environment Variables

**Required Variables:**
```
REFERENCE_BUCKET = your-reference-bucket-name
CHECKING_BUCKET = your-checking-bucket-name
```

**Optional Variables:**
```
JSON_RENDER_PATH = s3://your-reference-bucket-name/raw/
DYNAMODB_LAYOUT_TABLE = VendingMachineLayoutMetadata
AWS_REGION = us-east-1
LOG_LEVEL = info
```

### Step 3: Configure Function Settings

**General Configuration:**
- **Timeout:** 300 seconds (5 minutes)
- **Memory:** 1024 MB
- **Architecture:** x86_64

**Execution Role Permissions:**
Your Lambda execution role needs the following permissions:

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
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
```

### Step 4: Test Configuration

You can test the configuration by:

1. **Upload Test:** Upload a regular image file to test basic upload functionality
2. **Render Test:** Upload a JSON layout file to the `/raw` path to test rendering

### Environment Variable Details

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REFERENCE_BUCKET` | ✅ | - | S3 bucket for reference files |
| `CHECKING_BUCKET` | ✅ | - | S3 bucket for checking files |
| `JSON_RENDER_PATH` | ❌ | `raw` | S3 path where JSON files trigger rendering. Supports full S3 URI (s3://bucket/path/) or simple path |
| `DYNAMODB_LAYOUT_TABLE` | ❌ | - | DynamoDB table for layout metadata |
| `AWS_REGION` | ❌ | `us-east-1` | AWS region |
| `LOG_LEVEL` | ❌ | `info` | Logging level (debug, info, warn, error) |

### JSON_RENDER_PATH Examples

**Full S3 URI Format (Recommended):**
```
JSON_RENDER_PATH=s3://kootoro-dev-s3-reference-f6d3xl/raw/
```

**Simple Path Format (Backward Compatibility):**
```
JSON_RENDER_PATH=raw
```

**Cross-Bucket Rendering:**
```
JSON_RENDER_PATH=s3://special-layout-bucket/input/
```

### Troubleshooting

**Common Issues:**

1. **"REFERENCE_BUCKET and CHECKING_BUCKET environment variables are required"**
   - Solution: Set the required environment variables in Lambda configuration

2. **"Failed to upload file to S3"**
   - Solution: Check IAM permissions for S3 access

3. **"Failed to store layout metadata"**
   - Solution: Check DynamoDB table exists and IAM permissions

4. **Timeout errors**
   - Solution: Increase Lambda timeout to 300 seconds

5. **Memory errors during rendering**
   - Solution: Increase Lambda memory to 1024 MB or higher
