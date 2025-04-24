#!/bin/bash
set -e  # Exit immediately if a command fails




# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-verification-initialize"

echo "Using ECR repository: $ECR_REPO"

# Log in to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REPO"

# Build the image
docker build -t "$ECR_REPO:latest" .

# Push the image
docker push "$ECR_REPO:latest"

echo "Docker image built and pushed successfully to $ECR_REPO:latest"