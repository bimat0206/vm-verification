#!/bin/bash
# Script to manually push nginx images to ECR repositories
# Use this if you prefer not to use the Terraform null_resource

# Set variables
AWS_REGION="us-east-1"  # Change this to your desired region
PROJECT_PREFIX="vending-verification"

# List of functions that need repositories
FUNCTIONS=(
  "initialize"
  "fetch-images"
  "prepare-prompt"
  "invoke-bedrock"
  "process-results"
  "store-results"
  "notify"
  "get-comparison"
  "get-images"
)

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Login to ECR
echo "Logging in to Amazon ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Pull nginx image if not already available
if ! docker image inspect nginx:latest &> /dev/null; then
  echo "Pulling nginx image..."
  docker pull nginx:latest
fi

# Tag and push nginx image to each repository
for func in "${FUNCTIONS[@]}"; do
  REPO_NAME="${PROJECT_PREFIX}-${func}"
  REPO_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${REPO_NAME}"
  
  echo "Checking if repository $REPO_NAME exists..."
  if aws ecr describe-repositories --repository-names $REPO_NAME --region $AWS_REGION &> /dev/null; then
    echo "Tagging and pushing image to $REPO_NAME"
    docker tag nginx:latest $REPO_URI:latest
    docker push $REPO_URI:latest
    echo "Successfully pushed nginx image to $REPO_NAME"
  else
    echo "Repository $REPO_NAME does not exist. Please create it first."
  fi
done

echo "All placeholder images have been pushed."