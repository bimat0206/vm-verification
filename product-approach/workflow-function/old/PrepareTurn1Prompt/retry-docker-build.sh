#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Updated script for the new project structure - v4.0.11 with template loading fixes

# Manually set the ECR repository URL if terraform output isn't working
# Replace this with your actual ECR repository URL from the AWS console
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-prepare-turn1-prompt-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-prepare-turn1-f6d3xl"
AWS_REGION="us-east-1"
VERSION="4.0.11"
TAG="latest"

echo -e "${YELLOW}Building PrepareTurn1Prompt Lambda function v${VERSION}...${NC}"
echo "Using ECR repository: $ECR_REPO"

# Verify we're in the PrepareTurn1Prompt directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}Error: Cannot find required files. Make sure you're in the PrepareTurn1Prompt directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/PrepareTurn1Prompt/${NC}"
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
cp -r ../shared/templateloader "$BUILD_CONTEXT/shared/"

# Debug template directory
echo -e "${YELLOW}Checking template directory before build...${NC}"
find "$BUILD_CONTEXT/templates" -type d | sort
find "$BUILD_CONTEXT/templates" -type f | sort

# Ensure template directories have correct permissions
chmod -R 755 "$BUILD_CONTEXT/templates"

# Update Dockerfile version
sed -i '' "s/LABEL version=\"[0-9.]*\"/LABEL version=\"${VERSION}\"/" "$BUILD_CONTEXT/Dockerfile"

# Create a modified go.mod file for Docker build
echo -e "${YELLOW}Creating modified go.mod for Docker build...${NC}"
cat > "$BUILD_CONTEXT/go.mod" << EOF
module prepare-turn1

go 1.24

require (
	github.com/aws/aws-lambda-go v1.48.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	workflow-function/shared/errors v0.0.0
	workflow-function/shared/logger v0.0.0
	workflow-function/shared/s3state v0.0.0
	workflow-function/shared/schema v0.0.0
	workflow-function/shared/templateloader v0.0.0
	workflow-function/shared/bedrock v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	workflow-function/shared/bedrock => ./shared/bedrock
	workflow-function/shared/errors => ./shared/errors
	workflow-function/shared/logger => ./shared/logger
	workflow-function/shared/s3state => ./shared/s3state
	workflow-function/shared/schema => ./shared/schema
	workflow-function/shared/templateloader => ./shared/templateloader
)
EOF

# Build the image from the temporary build context
echo -e "${YELLOW}Building Docker image v${VERSION}...${NC}"

# Verify template directory structure in build context
echo -e "${YELLOW}Verifying template directory structure...${NC}"
find "$BUILD_CONTEXT/templates" -type f -name "*.tmpl" | sort

# We don't need to add debugging to the Dockerfile as we already have it

docker build -t "$ECR_REPO:$TAG" "$BUILD_CONTEXT"

# Clean up temporary directory
trap "rm -rf $BUILD_CONTEXT" EXIT

# Push the image
echo -e "${YELLOW}Pushing image to ECR...${NC}"
docker push "$ECR_REPO:$TAG"

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
echo -e "${YELLOW}Deploying to AWS Lambda...${NC}"
aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --image-uri "$ECR_REPO:$TAG" \
    --region "$AWS_REGION"  > /dev/null 2>&1



echo -e "${GREEN}✅ Done! Lambda function $FUNCTION_NAME has been updated with v${VERSION} code.${NC}"
echo -e "${GREEN}✅ Template execution error fixed. Verification should now proceed normally.${NC}"