#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Updated script for the new project structure

# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-fetch-images-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-fetch-images-f6d3xl"
AWS_REGION="us-east-1"

echo -e "${YELLOW}Building FetchImages Lambda function...${NC}"
echo "Using ECR repository: $ECR_REPO"

# Verify we're in the FetchImages directory
if [ ! -f "go.mod" ] || [ ! -d "cmd/fetchimages" ]; then
    echo -e "${RED}Error: Cannot find required files. Make sure you're in the FetchImages directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/FetchImages/${NC}"
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

# Copy the FetchImages directory contents
echo -e "${YELLOW}Copying FetchImages directory...${NC}"
cp -r ./* "$BUILD_CONTEXT/"

# Create shared directory and copy shared modules
echo -e "${YELLOW}Copying shared modules...${NC}"
mkdir -p "$BUILD_CONTEXT/shared"

# Check if shared modules exist and copy them
if [ -d "../shared/logger" ]; then
    cp -r ../shared/logger "$BUILD_CONTEXT/shared/"
    echo "✓ Copied shared/logger"
else
    echo -e "${RED}Warning: ../shared/logger not found${NC}"
fi

if [ -d "../shared/schema" ]; then
    cp -r ../shared/schema "$BUILD_CONTEXT/shared/"
    echo "✓ Copied shared/schema"
else
    echo -e "${RED}Warning: ../shared/schema not found${NC}"
fi

if [ -d "../shared/s3state" ]; then
    cp -r ../shared/s3state "$BUILD_CONTEXT/shared/"
    echo "✓ Copied shared/s3state"
else
    echo -e "${RED}Warning: ../shared/s3state not found${NC}"
fi

# Create a modified go.mod file for Docker build with updated replace paths
echo -e "${YELLOW}Creating modified go.mod for Docker build...${NC}"
cat > "$BUILD_CONTEXT/go.mod" << 'EOF'
module workflow-function/FetchImages

go 1.24

require (
	github.com/aws/aws-lambda-go v1.48.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	github.com/aws/smithy-go v1.22.3
	workflow-function/shared/logger v0.0.0
	workflow-function/shared/s3state v0.0.0
	workflow-function/shared/schema v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
)

replace workflow-function/shared/schema => ./shared/schema
replace workflow-function/shared/logger => ./shared/logger
replace workflow-function/shared/s3state => ./shared/s3state
EOF

# Copy the go.sum file to ensure all dependencies are properly resolved
echo -e "${YELLOW}Copying go.sum file...${NC}"
cp go.sum "$BUILD_CONTEXT/go.sum"

# Verify the build context structure
echo -e "${YELLOW}Verifying build context structure...${NC}"
echo "Build context contents:"
ls -la "$BUILD_CONTEXT"
echo "Shared modules in build context:"
ls -la "$BUILD_CONTEXT/shared/" || echo "No shared directory found"

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
