#!/bin/bash
set -e  # Exit immediately if a command fails

# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-prepare-system-prompt-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-prepare-system-prompt-f6d3xl"
AWS_REGION="us-east-1"

echo "Using ECR repository: $ECR_REPO"

# Log in to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REPO"

# Create Dockerfile.local
cat > Dockerfile.local << 'EOF'
# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app
ENV GO111MODULE=on

# Install required tools
RUN apk add --no-cache git

# Create directories for shared modules
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./templates ./templates
COPY ./events ./events
COPY *.md ./
COPY go.mod go.sum ./

# Create vendor directory
RUN go mod edit -replace=workflow-function/shared/schema=/app/shared/schema \
    && go mod edit -replace=workflow-function/shared/s3utils=/app/shared/s3utils \
    && go mod edit -replace=workflow-function/shared/templateloader=/app/shared/templateloader \
    && go mod edit -replace=workflow-function/shared/logger=/app/shared/logger

# Copy shared modules
COPY ./shared/schema /app/shared/schema
COPY ./shared/s3utils /app/shared/s3utils
COPY ./shared/templateloader /app/shared/templateloader
COPY ./shared/logger /app/shared/logger

# Build the application
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

COPY --from=build /main /main

RUN mkdir -p /opt/templates

COPY templates/ /opt/templates/

# Set component name for logging
ENV COMPONENT_NAME="PrepareSystemPrompt"

ENTRYPOINT ["/main"]
EOF

# Create a temporary directory for the build
mkdir -p ./shared
echo "Copying shared modules..."

# Copy shared modules 
cp -r ../shared/logger ./shared/
cp -r ../shared/s3utils ./shared/
cp -r ../shared/schema ./shared/
cp -r ../shared/templateloader ./shared/

# Build the image using the local Dockerfile
docker build -f Dockerfile.local -t "$ECR_REPO:latest" .

# Push the image
docker push "$ECR_REPO:latest"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions) - commented out as per instructions
echo "Deploying to AWS Lambda..."
aws lambda update-function-code \
 		--function-name "$FUNCTION_NAME" \
 		--image-uri "$ECR_REPO:latest" \
 		--region "$AWS_REGION" > /dev/null 2>&1

# Clean up
echo "Cleaning up temporary files"
rm -rf ./shared
rm -f Dockerfile.local

echo "Docker image built and pushed successfully to $ECR_REPO:latest"