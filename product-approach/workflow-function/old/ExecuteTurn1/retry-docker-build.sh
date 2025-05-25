#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default configuration
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-execute-turn1-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-execute-turn1-f6d3xl"
AWS_REGION="us-east-1"
SKIP_LAMBDA_UPDATE=false
SKIP_PUSH=false

# Parse command-line arguments
for arg in "$@"; do
  case $arg in
    --repo=*)
      ECR_REPO="${arg#*=}"
      shift
      ;;
    --function=*)
      FUNCTION_NAME="${arg#*=}"
      shift
      ;;
    --region=*)
      AWS_REGION="${arg#*=}"
      shift
      ;;
    --skip-lambda-update)
      SKIP_LAMBDA_UPDATE=true
      shift
      ;;
    --skip-push)
      SKIP_PUSH=true
      shift
      ;;
    --help)
      echo "=== ExecuteTurn1 Docker Build Script ==="
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --repo=REPO              ECR repository URI (default: $ECR_REPO)"
      echo "  --function=NAME          Lambda function name (default: $FUNCTION_NAME)"
      echo "  --region=REGION          AWS region (default: $AWS_REGION)"
      echo "  --skip-lambda-update     Skip Lambda function update"
      echo "  --skip-push              Skip pushing to ECR"
      echo "  --help                   Show this help message"
      exit 0
      ;;
  esac
done

echo -e "${YELLOW}Building ExecuteTurn1 Lambda function...${NC}"
echo "Using ECR repository: $ECR_REPO"

# Verify we're in the ExecuteTurn1 directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}Error: Cannot find required files. Make sure you're in the ExecuteTurn1 directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/ExecuteTurn1/${NC}"
    exit 1
fi

echo -e "${GREEN}Correct directory structure found${NC}"
echo -e "${YELLOW}Verifying files...${NC}"
ls -la | head -10

# Log in to ECR
echo -e "${YELLOW}Logging in to ECR...${NC}"
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin "$ECR_REPO"

# Create a temporary build context with shared modules
echo -e "${YELLOW}Creating temporary build context with shared modules...${NC}"
BUILD_CONTEXT=$(mktemp -d)
cp -r ./* "$BUILD_CONTEXT/"
mkdir -p "$BUILD_CONTEXT/shared"
cp -r ../shared/logger "$BUILD_CONTEXT/shared/"
cp -r ../shared/schema "$BUILD_CONTEXT/shared/"
cp -r ../shared/s3state "$BUILD_CONTEXT/shared/"
cp -r ../shared/errors "$BUILD_CONTEXT/shared/"
cp -r ../shared/bedrock "$BUILD_CONTEXT/shared/"

# Create a modified go.mod file for Docker build
echo -e "${YELLOW}Copying go.mod and go.sum to build context...${NC}"
cp go.mod "$BUILD_CONTEXT/go.mod"
cp go.sum "$BUILD_CONTEXT/go.sum"

# Update replace directives to point to the Docker build context paths
echo -e "${YELLOW}Updating module replace directives...${NC}"
sed -i '' "s|workflow-function/shared/bedrock => ../shared/bedrock|workflow-function/shared/bedrock => ./shared/bedrock|g" "$BUILD_CONTEXT/go.mod"
sed -i '' "s|workflow-function/shared/errors => ../shared/errors|workflow-function/shared/errors => ./shared/errors|g" "$BUILD_CONTEXT/go.mod"
sed -i '' "s|workflow-function/shared/logger => ../shared/logger|workflow-function/shared/logger => ./shared/logger|g" "$BUILD_CONTEXT/go.mod"
sed -i '' "s|workflow-function/shared/s3state => ../shared/s3state|workflow-function/shared/s3state => ./shared/s3state|g" "$BUILD_CONTEXT/go.mod"
sed -i '' "s|workflow-function/shared/schema => ../shared/schema|workflow-function/shared/schema => ./shared/schema|g" "$BUILD_CONTEXT/go.mod"

# Build the image from the temporary build context
echo -e "${YELLOW}Building Docker image...${NC}"
docker build -t "$ECR_REPO:latest" "$BUILD_CONTEXT"

# Clean up temporary directory
trap "rm -rf $BUILD_CONTEXT" EXIT

# Push the image
if [ "$SKIP_PUSH" = false ]; then
  echo -e "${YELLOW}Pushing image to ECR...${NC}"
  docker push "$ECR_REPO:latest"
else
  echo -e "${YELLOW}Skipping image push to ECR as requested${NC}"
fi

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
if [ "$SKIP_LAMBDA_UPDATE" = false ] && [ "$SKIP_PUSH" = false ]; then
  echo -e "${YELLOW}Deploying to AWS Lambda...${NC}"
  aws lambda update-function-code \
      --function-name "$FUNCTION_NAME" \
      --image-uri "$ECR_REPO:latest" \
      --region "$AWS_REGION" > /dev/null 2>&1
  
  echo -e "${GREEN}✅ Done! Lambda function $FUNCTION_NAME has been updated with the latest code.${NC}"
elif [ "$SKIP_LAMBDA_UPDATE" = true ]; then
  echo -e "${YELLOW}Skipping Lambda function update as requested${NC}"
elif [ "$SKIP_PUSH" = true ]; then
  echo -e "${YELLOW}Skipping Lambda function update because image push was skipped${NC}"
fi

echo -e "${GREEN}✅ Build process completed successfully${NC}"