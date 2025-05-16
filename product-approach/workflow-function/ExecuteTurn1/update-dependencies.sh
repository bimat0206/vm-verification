#!/bin/bash
set -e

# This script adds the required S3 dependencies to the shared schema module

# Directory check
if [ ! -d "../shared/schema" ]; then
  echo "Error: ../shared/schema directory not found!"
  echo "This script must be run from the ExecuteTurn1 directory."
  exit 1
fi

echo "=== Updating shared schema dependencies ==="

# Navigate to schema directory
pushd ../shared/schema > /dev/null

# Add missing S3 dependencies
echo "Adding AWS S3 SDK dependencies..."
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/service/s3/types
go mod tidy

# Verify dependencies were added
if grep -q "github.com/aws/aws-sdk-go-v2/service/s3" go.mod; then
  echo "✅ S3 SDK dependency added successfully"
else
  echo "❌ Failed to add S3 SDK dependency"
  popd > /dev/null
  exit 1
fi

if grep -q "github.com/aws/aws-sdk-go-v2/service/s3/types" go.mod; then
  echo "✅ S3 types dependency added successfully"
else
  echo "❌ Failed to add S3 types dependency"
  popd > /dev/null
  exit 1
fi

# Return to original directory
popd > /dev/null

echo "✅ Dependencies updated successfully"
echo ""
echo "You can now run ./retry-docker-build.sh to build the Docker image"