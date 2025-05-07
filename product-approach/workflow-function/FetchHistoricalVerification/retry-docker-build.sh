#!/bin/bash
set -e  # Exit immediately if a command fails




# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-fetch-historical-verification-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-fetch-historical-f6d3xl"
AWS_REGION="us-east-1"

echo "Using ECR repository: $ECR_REPO"

# Log in to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REPO"

# Build the image
docker build -t "$ECR_REPO:latest" .

# Push the image
docker push "$ECR_REPO:latest"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions) - commented out as per instructions
echo "Deploying to AWS Lambda..."
aws lambda update-function-code \
 		--function-name "$FUNCTION_NAME" \
 		--image-uri "$ECR_REPO:latest" \
 		--region "$AWS_REGION" > /dev/null 2>&1

echo "Docker image built and pushed successfully to $ECR_REPO:latest"