#!/bin/bash
# Script to test Lambda function with SAM CLI

# Create test directories if they don't exist
mkdir -p tmp/s3-mock/raw
mkdir -p tmp/s3-mock/rendered-layout
mkdir -p tmp/s3-mock/logs


# Create event file if it doesn't exist
if [ ! -f "test-event.json" ]; then
  echo '{
  "version": "0",
  "id": "17793124-05d4-b198-2fde-7ededc63b103",
  "detail-type": "Object Created",
  "source": "aws.s3",
  "account": "111122223333",
  "time": "2021-11-12T00:00:00Z",
  "region": "us-east-1",
  "resources": ["arn:aws:s3:::vending-machine-verification-image-reference-a11/"],
  "detail": {
    "version": "0",
    "bucket": {"name": "vending-machine-verification-image-reference-a11"},
    "object": {"key": "raw/DataInput.json", "size": 5},
    "request-id": "N4N7GDK58NMKJ12R",
    "requester": "123456789012"
  }
}' > test-event.json
fi

# Create env.json BEFORE invoking SAM
echo '{
  "RenderFunction": {
    "USE_MOCK_S3": "true",
    "TEMP_DIR": "'`pwd`'/tmp/s3-mock",
    "AWS_REGION": "us-east-1",
    "NODE_ENV": "test",
    "LOG_BUCKET": "vending-machine-verification-image-reference-a11"
  }
}' > env.json

# Install required dependencies if needed
if [ ! -d "node_modules/pngjs" ]; then
  echo "Installing dependencies..."
  npm install --silent pngjs canvas@2.11.2
fi

# Export environment variables to help with testing
export TEMP_DIR=`pwd`/tmp/s3-mock
export USE_MOCK_S3=true

# Run SAM CLI to invoke the Lambda
echo "Running SAM CLI to invoke Lambda..."

# Invoke Lambda with SAM CLI
sam local invoke RenderFunction \
  --env-vars env.json \
  -e test-event.json

# Check if output file was created (after SAM test)
echo "Checking for output files..."
if [ -f "tmp/s3-mock/rendered-layout/layout_test123.png" ]; then
  echo "Success! Output file created: tmp/s3-mock/rendered-layout/layout_test123.png"
  echo "Output file size: $(wc -c < tmp/s3-mock/rendered-layout/layout_test123.png) bytes"
else
  echo "Warning: Output file not found in tmp/s3-mock/rendered-layout/"
  # List all files in the directory to debug
  echo "Files in tmp/s3-mock:"
  find tmp/s3-mock -type f | sort
fi

# Check if log file was created
if [ -n "$(find tmp/s3-mock/logs -name "*.log" 2>/dev/null)" ]; then
  echo "Log file created in tmp/s3-mock/logs/"
  ls -la tmp/s3-mock/logs/
else
  echo "Warning: No log file found in tmp/s3-mock/logs/"
fi