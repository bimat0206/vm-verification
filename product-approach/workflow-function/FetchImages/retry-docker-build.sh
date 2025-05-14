#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-fetch-images-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-fetch-images-f6d3xl"
AWS_REGION="us-east-1"

echo -e "${YELLOW}Building FetchImages Lambda function...${NC}"
echo "Using ECR repository: $ECR_REPO"

# Check if we're in the FetchImages directory and move up to workflow-function
if [ ! -d "../shared" ]; then
    echo -e "${RED}Error: Cannot find shared directory. Make sure you're in the FetchImages directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/FetchImages/${NC}"
    echo -e "${RED}Expected structure:${NC}"
    echo -e "${RED}  workflow-function/${NC}"
    echo -e "${RED}    ├── FetchImages/ (you are here)${NC}"
    echo -e "${RED}    └── shared/${NC}"
    exit 1
fi

echo -e "${GREEN}Correct directory structure found${NC}"

# Move to the parent directory (workflow-function) for the build context
cd ..

echo -e "${YELLOW}Changed to build context directory: $(pwd)${NC}"
echo -e "${YELLOW}Verifying structure...${NC}"
ls -la FetchImages/ shared/ | head -10

# Log in to ECR
echo -e "${YELLOW}Logging in to ECR...${NC}"
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REPO"

# Build the image from the workflow-function directory
echo -e "${YELLOW}Building Docker image...${NC}"
docker build -f FetchImages/Dockerfile -t "$ECR_REPO:latest" .

# Push the image
echo -e "${YELLOW}Pushing image to ECR...${NC}"
docker push "$ECR_REPO:latest"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
echo -e "${YELLOW}Deploying to AWS Lambda...${NC}"
aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --image-uri "$ECR_REPO:latest" \
    --region "$AWS_REGION" > /dev/null 2>&1

echo -e "${GREEN}✅ Done! Lambda function $FUNCTION_NAME has been updated with the latest code.${NC}"