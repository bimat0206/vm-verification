#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-execute-turn2-combined-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-execute-turn2-combined-f6d3xl"
AWS_REGION="us-east-1"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --repo=*)
      ECR_REPO="${1#*=}"
      shift
      ;;
    --function=*)
      FUNCTION_NAME="${1#*=}"
      shift
      ;;
    --region=*)
      AWS_REGION="${1#*=}"
      shift
      ;;
    *)
      echo -e "${RED}Unknown parameter: $1${NC}"
      echo "Usage: $0 --repo=<ECR_REPO_URI> --function=<LAMBDA_FUNCTION_NAME> [--region=<AWS_REGION>]"
      exit 1
      ;;
  esac
done

# Validate required parameters
if [ -z "$ECR_REPO" ] || [ -z "$FUNCTION_NAME" ]; then
  echo -e "${RED}Error: Missing required parameters${NC}"
  echo "Usage: $0 --repo=<ECR_REPO_URI> --function=<LAMBDA_FUNCTION_NAME> [--region=<AWS_REGION>]"
  exit 1
fi

echo -e "${YELLOW}Building ExecuteTurn2Combined Lambda function...${NC}"
echo "Using ECR repository: $ECR_REPO"
echo "Function name: $FUNCTION_NAME"
echo "AWS region: $AWS_REGION"

# Verify we're in the ExecuteTurn2Combined directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}Error: Cannot find required files. Make sure you're in the ExecuteTurn2Combined directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/ExecuteTurn2Combined/${NC}"
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
mkdir -p "$BUILD_CONTEXT/templates"
cp -r ../shared/bedrock "$BUILD_CONTEXT/shared/"
cp -r ../shared/logger "$BUILD_CONTEXT/shared/"
cp -r ../shared/schema "$BUILD_CONTEXT/shared/"
cp -r ../shared/s3state "$BUILD_CONTEXT/shared/"
cp -r ../shared/errors "$BUILD_CONTEXT/shared/"
cp -r ../shared/templateloader "$BUILD_CONTEXT/shared/"

# Copy template files to build context
echo -e "${YELLOW}Copying template files...${NC}"
if [ -d "templates" ] && [ "$(ls -A templates)" ]; then
    cp -r ./templates/* "$BUILD_CONTEXT/templates/"
    echo -e "${GREEN}Template files copied successfully${NC}"
    ls -la "$BUILD_CONTEXT/templates/"
else
    echo -e "${RED}Warning: No template files found in ./templates/${NC}"
fi

# Create a modified go.mod file for Docker build
echo -e "${YELLOW}Creating modified go.mod for Docker build...${NC}"
cat > "$BUILD_CONTEXT/go.mod" << EOF
module workflow-function/ExecuteTurn2Combined

go 1.24.0

require (
	github.com/aws/aws-lambda-go v1.48.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	workflow-function/shared/bedrock v0.0.0
	workflow-function/shared/errors v0.0.0
	workflow-function/shared/logger v0.0.0
	workflow-function/shared/s3state v0.0.0
	workflow-function/shared/schema v0.0.0
	workflow-function/shared/templateloader v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.30.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace workflow-function/shared/bedrock => ./shared/bedrock

replace workflow-function/shared/errors => ./shared/errors

replace workflow-function/shared/logger => ./shared/logger

replace workflow-function/shared/s3state => ./shared/s3state

replace workflow-function/shared/schema => ./shared/schema

replace workflow-function/shared/templateloader => ./shared/templateloader
EOF

# Build the image from the temporary build context
echo -e "${YELLOW}Building Docker image...${NC}"
docker build -t "$ECR_REPO:latest" "$BUILD_CONTEXT"

# Clean up temporary directory
trap "rm -rf $BUILD_CONTEXT" EXIT

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