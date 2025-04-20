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

# Create env.json with SAM_LOCAL flag
echo '{
  "RenderFunction": {
    "USE_MOCK_S3": "true",
    "AWS_SAM_LOCAL": "true",
    "TEMP_DIR": "/tmp/s3-mock",
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
export USE_MOCK_S3=true
export AWS_SAM_LOCAL=true

# Run SAM CLI to invoke the Lambda
echo "Running SAM CLI to invoke Lambda..."

# Option 1: Run Lambda with debug output to extract container ID
echo "Starting Lambda with container ID capture..."
CONTAINER_OUTPUT=$(sam local invoke RenderFunction \
  --env-vars env.json \
  -e test-event.json \
  --debug 2>&1)

# Try to extract container ID from debug output
CONTAINER_ID=$(echo "$CONTAINER_OUTPUT" | grep -o "Container Id: [a-zA-Z0-9]*" | sed 's/Container Id: //')

# If container ID extraction failed, try other methods
if [ -z "$CONTAINER_ID" ]; then
  echo "Container ID not found in debug output, trying Docker ps..."
  # Get latest container ID from Docker for Lambda runtime
  CONTAINER_ID=$(docker ps -a --filter "ancestor=public.ecr.aws/lambda/nodejs:22-rapid-arm64" --format "{{.ID}}" | head -n 1)
fi

if [ -n "$CONTAINER_ID" ]; then
  echo "Found Lambda container ID: $CONTAINER_ID"
  
  # Wait for Lambda execution to complete
  sleep 2
  
  # Copy output files from container to local
  echo "Copying files from container to host..."
  
  # Create temporary script to run inside the container
  cat > copy_files.sh << 'EOL'
#!/bin/bash
# List all files in /tmp/s3-mock for debugging
echo "Files in container /tmp/s3-mock:"
find /tmp/s3-mock -type f | sort

# Check if rendered file exists
if [ -f "/tmp/s3-mock/rendered-layout/layout_test123.png" ]; then
  echo "Found output image in container: /tmp/s3-mock/rendered-layout/layout_test123.png"
  ls -la /tmp/s3-mock/rendered-layout/layout_test123.png
else
  echo "No output image found in container!"
fi
EOL
  
  chmod +x copy_files.sh
  docker cp copy_files.sh $CONTAINER_ID:/tmp/
  docker exec $CONTAINER_ID /tmp/copy_files.sh
  
  # Copy all files from container's /tmp/s3-mock to local tmp/s3-mock
  echo "Copying all files from container to host..."
  docker cp $CONTAINER_ID:/tmp/s3-mock/. ./tmp/s3-mock/
  
  # Fix permissions
  chmod -R 755 ./tmp/s3-mock/
else
  echo "Warning: Could not extract container ID. Output may not be copied."
  echo "$CONTAINER_OUTPUT"
fi

# Check if output file was created
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

# Option 2: Alternative approach without container ID
# This creates a direct test that doesn't rely on SAM
echo "Running alternative direct test (Node.js test)..."
cat > direct-test.js << 'EOL'
// Simple direct test that bypasses SAM
process.env.USE_MOCK_S3 = 'true';
process.env.AWS_SAM_LOCAL = 'true';
process.env.TEMP_DIR = '/tmp/s3-mock';

const { handler } = require('./index');
const event = require('./test-event.json');

console.log('Starting direct test...');
handler(event)
  .then(result => {
    console.log('Test completed successfully:', result);
    process.exit(0);
  })
  .catch(error => {
    console.error('Test failed:', error);
    process.exit(1);
  });
EOL

# Run the Node.js test
echo "Running direct Node.js test..."
docker run --rm -v "$(pwd):/var/task" -e USE_MOCK_S3=true -e AWS_SAM_LOCAL=true -e TEMP_DIR=/tmp/s3-mock public.ecr.aws/lambda/nodejs:22-rapid-arm64 direct-test.js

# Copy again from the most recent container (from direct test)
RECENT_CONTAINER=$(docker ps -a --format "{{.ID}}" | head -n 1)
if [ -n "$RECENT_CONTAINER" ]; then
  echo "Copying files from most recent container: $RECENT_CONTAINER"
  docker cp $RECENT_CONTAINER:/tmp/s3-mock/. ./tmp/s3-mock/
fi

# Final check for output files after both tests
echo "Final check for output files..."
find tmp/s3-mock -type f | sort