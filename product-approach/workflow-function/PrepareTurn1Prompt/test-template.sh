#!/bin/bash
set -e  # Exit immediately if a command fails

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
TEST_INPUT_FILE="test-input.json"

# Verify we're in the PrepareTurn1Prompt directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}Error: Cannot find required files. Make sure you're in the PrepareTurn1Prompt directory${NC}"
    echo -e "${RED}Expected to be in: workflow-function/PrepareTurn1Prompt/${NC}"
    exit 1
fi

echo -e "${GREEN}Correct directory structure found${NC}"

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

# Create a modified go.mod file for local testing
echo -e "${YELLOW}Creating modified go.mod for build...${NC}"
cat > "$BUILD_CONTEXT/go.mod" << EOF
module prepare-turn1

go 1.24.0

require (
    github.com/aws/aws-lambda-go v1.48.0
    github.com/aws/aws-sdk-go-v2 v1.36.3
    github.com/aws/aws-sdk-go-v2/config v1.29.14
    github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
    workflow-function/shared/errors v0.0.0-00010101000000-000000000000
    workflow-function/shared/logger v0.0.0-00010101000000-000000000000
    workflow-function/shared/s3state v0.0.0-00010101000000-000000000000
    workflow-function/shared/schema v0.0.0-00010101000000-000000000000
    workflow-function/shared/templateloader v0.0.0-00010101000000-000000000000
    workflow-function/shared/bedrock v0.0.0-00010101000000-000000000000
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

# Copy templates directory with the updated template
cp -r templates/ "$BUILD_CONTEXT/"

# Create test input file to use real data from the failing verification
cat > "$BUILD_CONTEXT/test-template-input.json" << EOF
{
  "schemaVersion": "2.0",
  "s3References": {
    "images_metadata": {
      "bucket": "kootoro-dev-s3-state-f6d3xl",
      "key": "2025/05/20/verif-20250520082751-de81/images/metadata.json"
    },
    "processing_initialization": {
      "bucket": "kootoro-dev-s3-state-f6d3xl",
      "key": "2025/05/20/verif-20250520082751-de81/processing/initialization.json"
    },
    "processing_layout-metadata": {
      "bucket": "kootoro-dev-s3-state-f6d3xl",
      "key": "2025/05/20/verif-20250520082751-de81/processing/layout-metadata.json"
    },
    "prompts_system": {
      "bucket": "kootoro-dev-s3-state-f6d3xl",
      "key": "2025/05/20/verif-20250520082751-de81/prompts/system-prompt.json"
    }
  },
  "verificationId": "verif-20250520082751-de81",
  "verificationType": "LAYOUT_VS_CHECKING",
  "status": "PROMPT_PREPARED"
}
EOF

# Build the application
echo -e "${YELLOW}Building application...${NC}"
cd "$BUILD_CONTEXT"
go build -o main ./cmd/main.go

# Set environment variables for testing
export STATE_BUCKET="kootoro-dev-s3-state-f6d3xl"
export TEMPLATE_BASE_PATH="./templates"
export INPUT_FILE="test-template-input.json"
export DEBUG="true"

# Run the application
echo -e "${YELLOW}Running test with updated template handling...${NC}"
./main

# Clean up temporary directory
cd - > /dev/null
trap "rm -rf $BUILD_CONTEXT" EXIT

echo -e "${GREEN}✅ Test completed${NC}"