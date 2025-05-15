#!/bin/bash
set -e  # Exit immediately if a command fails

# Configuration
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-prepare-system-prompt-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-prepare-system-prompt-f6d3xl"
AWS_REGION="us-east-1"
IMAGE_TAG="latest"

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

# Create build context with shared modules
echo "Preparing build context..."

# Create temporary directory for build context
BUILD_CONTEXT=$(mktemp -d)
trap "rm -rf $BUILD_CONTEXT" EXIT

# Copy function code
cp -r ./cmd "$BUILD_CONTEXT/"
cp -r ./internal "$BUILD_CONTEXT/"
cp -r ./templates "$BUILD_CONTEXT/"
cp go.mod go.sum "$BUILD_CONTEXT/"
cp *.md "$BUILD_CONTEXT/" 2>/dev/null || true

# Create shared modules directory in build context
mkdir -p "$BUILD_CONTEXT/shared"

# Copy shared modules
echo "Copying shared modules..."
for module in schema s3utils templateloader logger; do
    if [ -d "../shared/$module" ]; then
        cp -r "../shared/$module" "$BUILD_CONTEXT/shared/"
        echo "  ✓ Copied shared/$module"
    else
        echo "  ✗ Warning: shared/$module not found"
    fi
done

# Create Dockerfile for build
cat > "$BUILD_CONTEXT/Dockerfile" << 'EOF'
# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app
ENV GO111MODULE=on

# Install required tools
RUN apk add --no-cache git

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Copy source code
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY *.md ./

# Copy shared modules
COPY ./shared ./shared

# Update go.mod to use local shared modules
RUN go mod edit -replace=workflow-function/shared/schema=./shared/schema \
    && go mod edit -replace=workflow-function/shared/s3utils=./shared/s3utils \
    && go mod edit -replace=workflow-function/shared/templateloader=./shared/templateloader \
    && go mod edit -replace=workflow-function/shared/logger=./shared/logger

# Download dependencies and build
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

# Final stage
FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

# Copy the binary
COPY --from=build /main /main

# Create templates directory and copy templates
RUN mkdir -p /opt/templates
COPY templates/ /opt/templates/

# Set component name for logging
ENV COMPONENT_NAME="PrepareSystemPrompt"

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
    --output text   > /dev/null 2>&1

echo "=== Build and Deployment Complete ==="
echo "Image: $ECR_REPO:$IMAGE_TAG"
echo "Function: $FUNCTION_NAME"
echo "========================================="