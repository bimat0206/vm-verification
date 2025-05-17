#!/bin/bash

# Set ECR repository details
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-fetch-historical-verification-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-fetch-historical-f6d3xl"
AWS_REGION="us-east-1"

echo "Building and deploying FetchHistoricalVerification v1.1.0"
echo "Using ECR repository: $ECR_REPO"

# Log in to ECR
echo "Logging in to Amazon ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin "$ECR_REPO"

# Build the image from the parent directory (workflow-function) to include shared packages
cd .. 
echo "Building Docker image..."
docker build -t "$ECR_REPO:latest" -f ./FetchHistoricalVerification/Dockerfile .

# Push the image (still in the parent directory)
echo "Pushing image to ECR..."
docker push "$ECR_REPO:latest"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
echo "Deploying to AWS Lambda..."
aws lambda update-function-code \
        --function-name "$FUNCTION_NAME" \
        --image-uri "$ECR_REPO:latest" \
        --region "$AWS_REGION" > /dev/null 2>&1

# Return to the original directory
cd FetchHistoricalVerification

echo "Deployment complete. Image URI: $ECR_REPO:latest"
echo "Function: $FUNCTION_NAME"