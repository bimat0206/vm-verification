#!/bin/bash

# Example deployment script for API Verifications Conversation Lambda Function
# This script shows how to set the required environment variables and deploy

# Set required environment variables
export ECR_REPO="123456789012.dkr.ecr.us-east-1.amazonaws.com/my-conversation-repo"
export LAMBDA_FUNCTION_NAME="my-conversation-lambda-function"

# Optional environment variables
export AWS_REGION="us-east-1"
export IMAGE_TAG="latest"

# Run the deployment
echo "Starting deployment with the following configuration:"
echo "ECR Repository: $ECR_REPO"
echo "Lambda Function: $LAMBDA_FUNCTION_NAME"
echo "AWS Region: $AWS_REGION"
echo "Image Tag: $IMAGE_TAG"
echo ""

# Execute the deployment
./deploy.sh
