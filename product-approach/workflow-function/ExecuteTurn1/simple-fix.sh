#!/bin/bash
set -euo pipefail

# Function to cleanup on exit or error
cleanup() {
  local exit_code=$?
  local temp_dir="${TEMP_DIR:-./docker_build}"
  
  echo "Cleaning up temporary build files..."
  if [ -d "$temp_dir" ]; then
    rm -rf "$temp_dir"
  fi
  
  if [ $exit_code -ne 0 ]; then
    echo "❌ Script failed with exit code $exit_code"
  fi
  
  exit $exit_code
}

# Set trap to ensure cleanup happens even on errors
trap cleanup EXIT INT TERM

# Print header
echo "=========================================================="
echo "    ExecuteTurn1 Lambda Function Builder & Deployer       "
echo "=========================================================="
echo "Started at: $(date '+%Y-%m-%d %H:%M:%S')"
echo

# Configuration
TEMP_DIR="./docker_build"
REPO_NAME="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-execute-turn1-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-execute-turn1-f6d3xl"
AWS_REGION="us-east-1"

# Track timing
START_TIME=$(date +%s)

# Clean up any existing directory
rm -rf $TEMP_DIR

# Create a new build directory
mkdir -p $TEMP_DIR/shared
mkdir -p $TEMP_DIR/cmd
mkdir -p $TEMP_DIR/internal/handler
mkdir -p $TEMP_DIR/internal/dependencies
mkdir -p $TEMP_DIR/internal/config
mkdir -p $TEMP_DIR/internal/models

# Copy files
cp cmd/main.go $TEMP_DIR/cmd/
cp internal/handler/*.go $TEMP_DIR/internal/handler/
cp internal/dependencies/*.go $TEMP_DIR/internal/dependencies/
cp internal/config/*.go $TEMP_DIR/internal/config/
cp internal/models/*.go $TEMP_DIR/internal/models/

# Copy and check shared modules
cp -r ../shared/errors $TEMP_DIR/shared/
cp -r ../shared/logger $TEMP_DIR/shared/
cp -r ../shared/schema $TEMP_DIR/shared/

# Create a fixed go.mod file
cat > $TEMP_DIR/go.mod << EOF
module workflow-function/ExecuteTurn1

go 1.22

require (
	github.com/aws/aws-lambda-go v1.46.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.27.9
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.7.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	github.com/aws/smithy-go v1.22.2
	github.com/google/uuid v1.6.0
	workflow-function/shared/errors v0.0.0
	workflow-function/shared/logger v0.0.0
	workflow-function/shared/schema v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.9 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.5 // indirect
)

replace (
	workflow-function/shared/errors => ./shared/errors
	workflow-function/shared/logger => ./shared/logger
	workflow-function/shared/schema => ./shared/schema
)
EOF

# Create simple go.mod in shared
cat > $TEMP_DIR/shared/errors/go.mod << EOF
module workflow-function/shared/errors

go 1.22
EOF

cat > $TEMP_DIR/shared/logger/go.mod << EOF
module workflow-function/shared/logger

go 1.22
EOF

cat > $TEMP_DIR/shared/schema/go.mod << EOF
module workflow-function/shared/schema

go 1.22
EOF

# Create Dockerfile
cat > $TEMP_DIR/Dockerfile << EOF
FROM --platform=linux/arm64 golang:1.22-alpine AS builder

# Install git and dependencies
RUN apk add --no-cache git build-base 

# Set up the working directory
WORKDIR /build

# Copy everything
COPY . .

# Build the application 
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/main.go

# Create the lambda runtime image
FROM --platform=linux/arm64 public.ecr.aws/lambda/provided:al2-arm64

# Add function version metadata
LABEL function.version="1.3.2" 
LABEL function.description="ExecuteTurn1 - Bedrock InvokeModel API implementation"

# Copy the compiled binary
COPY --from=builder /build/bootstrap /var/runtime/bootstrap

# Make the bootstrap file executable
RUN chmod +x /var/runtime/bootstrap

# Set the entry point for AWS Lambda
CMD ["bootstrap"]
EOF

# ECR Login with suppressed output
echo "Logging in to ECR..."
if ! aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin "$(echo "$REPO_NAME" | cut -d'/' -f1)" >/dev/null 2>&1; then
  echo "❌ ECR login failed"
  exit 1
else
  echo "✅ ECR login successful"
fi

# Build Docker image with concise output
echo "Building Docker image... (this may take a minute)"
if ! DOCKER_BUILDKIT=1 docker build --quiet --platform=linux/arm64 -t "${REPO_NAME}:latest" $TEMP_DIR >/dev/null; then
  echo "❌ Docker build failed"
  exit 1
else
  echo "✅ Docker build completed successfully"
fi

# Push Image automatically
echo "Pushing image to ECR..."
if ! docker push "${REPO_NAME}:latest" >/dev/null 2>&1; then
  echo "❌ Failed to push Docker image to ECR"
  exit 1
else
  echo "✅ Image successfully pushed to ECR"
fi

# Update Lambda function automatically
echo "Updating Lambda function..."
if ! aws lambda update-function-code \
    --function-name $FUNCTION_NAME \
    --image-uri "${REPO_NAME}:latest" \
    --region $AWS_REGION \
    --output json \
    >/dev/null 2>&1; then
  echo "❌ Failed to update Lambda function"
  exit 1
else 
  echo "✅ Lambda function updated successfully"
  
  # Get function configuration to verify update
  echo "Current function state:"
  aws lambda get-function \
    --function-name $FUNCTION_NAME \
    --region $AWS_REGION \
    --query 'Configuration.[LastUpdateStatus,State,LastModified]' \
    --output json
fi

# Calculate execution time
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
MINUTES=$((DURATION / 60))
SECONDS=$((DURATION % 60))

# Cleanup happens automatically via trap

echo "=========================================================="
echo "✅ Deployment completed successfully!"
echo "   Total execution time: ${MINUTES}m ${SECONDS}s"
echo "   Lambda function: $FUNCTION_NAME"
echo "   Image URI: ${REPO_NAME}:latest"
echo "=========================================================="