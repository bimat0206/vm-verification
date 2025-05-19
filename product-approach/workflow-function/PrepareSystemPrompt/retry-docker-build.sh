#!/bin/bash
set -e  # Exit immediately if a command fails

# Configuration - Edit these values if needed
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-prepare-system-prompt-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-prepare-system-prompt-f6d3xl"
AWS_REGION="us-east-1"
IMAGE_TAG="latest"  # Updated version tag to match changelog

echo "=== Kootoro PrepareSystemPrompt Docker Build ==="
echo "ECR Repository: $ECR_REPO"
echo "AWS Region: $AWS_REGION"
echo "=============================================="

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check required tools
if ! command_exists docker; then
    echo "Error: Docker is not installed or not in PATH"
    exit 1
fi

if ! command_exists aws; then
    echo "Error: AWS CLI is not installed or not in PATH"
    exit 1
fi

# Log in to ECR
echo "Logging into ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin "$ECR_REPO"

# Create build context
echo "Preparing build context..."
BUILD_CONTEXT=$(mktemp -d)
trap "rm -rf $BUILD_CONTEXT" EXIT

# Copy the new directory structure
echo "Copying application code..."
mkdir -p "$BUILD_CONTEXT/cmd" \
         "$BUILD_CONTEXT/internal/adapters" \
         "$BUILD_CONTEXT/internal/config" \
         "$BUILD_CONTEXT/internal/handlers" \
         "$BUILD_CONTEXT/internal/models" \
         "$BUILD_CONTEXT/internal/processors"

# Copy all subdirectories and files
cp -r ./cmd/* "$BUILD_CONTEXT/cmd/"
cp -r ./internal/adapters/* "$BUILD_CONTEXT/internal/adapters/"
cp -r ./internal/config/* "$BUILD_CONTEXT/internal/config/"
cp -r ./internal/handlers/* "$BUILD_CONTEXT/internal/handlers/"
cp -r ./internal/models/* "$BUILD_CONTEXT/internal/models/"
cp -r ./internal/processors/* "$BUILD_CONTEXT/internal/processors/"
cp -r ./templates "$BUILD_CONTEXT/"
cp go.mod go.sum "$BUILD_CONTEXT/"
cp *.md "$BUILD_CONTEXT/" 2>/dev/null || true

# Create shared modules directory in build context
mkdir -p "$BUILD_CONTEXT/shared"

# Copy shared modules and fix import paths
echo "Copying shared modules and fixing import paths..."
for module in schema s3state templateloader logger errors; do
    if [ -d "../shared/$module" ]; then
        cp -r "../shared/$module" "$BUILD_CONTEXT/shared/"
        echo "  ✓ Copied shared/$module"
        
        # Fix import paths in the copied module
        find "$BUILD_CONTEXT/shared/$module" -name "*.go" -exec sed -i 's|product-approach/workflow-function/shared/|workflow-function/shared/|g' {} \;
    else
        echo "  ✗ Warning: shared/$module not found"
    fi
done

# Create optimized Dockerfile
cat > "$BUILD_CONTEXT/Dockerfile" << 'EOF'
# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install required tools
RUN apk add --no-cache git

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Copy source code with new directory structure
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY *.md ./

# Copy shared modules
COPY ./shared ./shared

# Update go.mod to use local shared modules
RUN go mod edit -replace=workflow-function/shared/schema=./shared/schema \
    && go mod edit -replace=workflow-function/shared/templateloader=./shared/templateloader \
    && go mod edit -replace=workflow-function/shared/logger=./shared/logger \
    && go mod edit -replace=workflow-function/shared/s3state=./shared/s3state \
    && go mod edit -replace=workflow-function/shared/error=./shared/errors

# Download dependencies and build
RUN go mod download && go mod tidy
RUN GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

# Final stage
FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

# Copy the binary
COPY --from=build /main /main

# Create templates directory and copy templates
RUN mkdir -p /opt/templates
COPY templates/ /opt/templates/

# Set environment variables 
ENV COMPONENT_NAME="PrepareSystemPrompt"
ENV DATE_PARTITION_TIMEZONE="UTC"
ENV TEMPLATE_BASE_PATH="/opt/templates"
ENV DEBUG="false"

ENTRYPOINT ["/main"]
EOF

# Build the Docker image
echo "Building Docker image..."
docker build -t "$ECR_REPO:$IMAGE_TAG" "$BUILD_CONTEXT"

# Push the image
echo "Pushing image to ECR..."
docker push "$ECR_REPO:$IMAGE_TAG"

# Update Lambda function
echo "Updating Lambda function..."
aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --image-uri "$ECR_REPO:$IMAGE_TAG" \
    --region "$AWS_REGION" \
    --no-cli-pager

# Verify the update
echo "Verifying Lambda function update..."
aws lambda get-function \
    --function-name "$FUNCTION_NAME" \
    --region "$AWS_REGION" \
    --query 'Code.ImageUri' \
    --output text > /dev/null 2>&1

echo "=== Build and Deployment Complete ==="
echo "Image: $ECR_REPO:$IMAGE_TAG"
echo "Function: $FUNCTION_NAME"
echo "========================================="