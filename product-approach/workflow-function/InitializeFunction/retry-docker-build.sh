#!/bin/bash
set -e  # Exit immediately if a command fails




# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-initialize-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-initialize-f6d3xl"
AWS_REGION="us-east-1"

echo "Using ECR repository: $ECR_REPO"

# Log in to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REPO"

# Get current directory
CURRENT_DIR=$(pwd)
PARENT_DIR=$(dirname "$CURRENT_DIR")

# Create a temporary directory for the build context
TEMP_DIR=$(mktemp -d)
echo "Created temporary build directory: $TEMP_DIR"

# Create necessary directories
mkdir -p "$TEMP_DIR/workflow-function/shared/schema"
mkdir -p "$TEMP_DIR/workflow-function/shared/logger"
mkdir -p "$TEMP_DIR/workflow-function/shared/s3utils"
mkdir -p "$TEMP_DIR/workflow-function/shared/dbutils"
mkdir -p "$TEMP_DIR/workflow-function/InitializeFunction"

# Copy all shared packages
cp -r "$PARENT_DIR/shared/schema"/* "$TEMP_DIR/workflow-function/shared/schema/"
cp -r "$PARENT_DIR/shared/logger"/* "$TEMP_DIR/workflow-function/shared/logger/"
cp -r "$PARENT_DIR/shared/s3utils"/* "$TEMP_DIR/workflow-function/shared/s3utils/"
cp -r "$PARENT_DIR/shared/dbutils"/* "$TEMP_DIR/workflow-function/shared/dbutils/"

# Copy InitializeFunction files
cp -r "$CURRENT_DIR"/* "$TEMP_DIR/workflow-function/InitializeFunction/"

# Create a temporary Dockerfile for the build
cat > "$TEMP_DIR/Dockerfile" << 'EOF'
FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set working directory
WORKDIR /app

# Copy with correct paths for build context
COPY workflow-function/ /app/workflow-function/

# Build from the function directory
WORKDIR /app/workflow-function/InitializeFunction

# Build the application for AWS Lambda ARM64 (Graviton)
# Use build flags to create a statically linked binary
RUN GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o /main

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]
EOF

# Move to the temp directory for building
cd "$TEMP_DIR"

# Build the image
echo "Building Docker image from temp directory: $TEMP_DIR"
docker build -t "$ECR_REPO:latest" .

# Clean up temp directory
cd "$CURRENT_DIR"
rm -rf "$TEMP_DIR"
echo "Removed temporary build directory"

# Push the image
docker push "$ECR_REPO:latest"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
echo "Deploying to AWS Lambda..."
aws lambda update-function-code \
 		--function-name "$FUNCTION_NAME" \
 		--image-uri "$ECR_REPO:latest" \
 		--region "$AWS_REGION" > /dev/null 2>&1

echo "Docker image built and pushed successfully to $ECR_REPO:latest"