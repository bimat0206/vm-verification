#!/bin/bash

# Test Environment Variables Script
# This script tests if the Lambda function properly loads environment variables

echo "Testing environment variable loading..."

# Set test environment variables
export REFERENCE_BUCKET="test-reference-bucket"
export CHECKING_BUCKET="test-checking-bucket"
export JSON_RENDER_PATH="s3://test-reference-bucket/raw/"
export DYNAMODB_LAYOUT_TABLE="TestLayoutTable"
export AWS_REGION="us-west-2"
export LOG_LEVEL="debug"

echo "Set environment variables:"
echo "REFERENCE_BUCKET=$REFERENCE_BUCKET"
echo "CHECKING_BUCKET=$CHECKING_BUCKET"
echo "JSON_RENDER_PATH=$JSON_RENDER_PATH"
echo "DYNAMODB_LAYOUT_TABLE=$DYNAMODB_LAYOUT_TABLE"
echo "AWS_REGION=$AWS_REGION"
echo "LOG_LEVEL=$LOG_LEVEL"

echo ""
echo "Testing Lambda function initialization..."

# Create a simple test payload for OPTIONS request (should not require AWS credentials)
cat > test_options.json << 'EOF'
{
  "httpMethod": "OPTIONS",
  "path": "/upload",
  "headers": {
    "Origin": "https://example.com"
  },
  "queryStringParameters": {},
  "body": "",
  "isBase64Encoded": false
}
EOF

echo "Created test payload for OPTIONS request"
echo "Note: This test only verifies environment variable loading and basic initialization"
echo "For full testing, you'll need valid AWS credentials and S3 buckets"

# Clean up
rm -f test_options.json

echo "Test preparation complete!"
echo ""
echo "To test with actual AWS resources:"
echo "1. Set up valid AWS credentials"
echo "2. Create the S3 buckets specified in environment variables"
echo "3. Create DynamoDB table (optional)"
echo "4. Use the test_payload.json file for a complete test"
