#!/bin/bash
set -eo pipefail  # Exit immediately if a command fails with improved error propagation

# Default Configuration
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-execute-turn1-f6d3xl"
FUNCTION_NAME="kootoro-dev-lambda-execute-turn1-f6d3xl"
AWS_REGION="us-east-1"
SHARED_PACKAGES_DIR="../shared"
TEMP_BUILD_DIR="./temp_build"
IMAGE_TAG="latest"
SKIP_LAMBDA_UPDATE=false
SKIP_PUSH=false
SKIP_ECR_LOGIN=false
NO_CACHE=false

# Script directory for absolute paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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
    --shared-dir=*)
      SHARED_PACKAGES_DIR="${arg#*=}"
      shift
      ;;
    --temp-dir=*)
      TEMP_BUILD_DIR="${arg#*=}"
      shift
      ;;
    --tag=*)
      IMAGE_TAG="${arg#*=}"
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
    --skip-ecr-login)
      SKIP_ECR_LOGIN=true
      shift
      ;;
    --no-cache)
      NO_CACHE=true
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
      echo "  --shared-dir=DIR         Shared packages directory (default: $SHARED_PACKAGES_DIR)"
      echo "  --temp-dir=DIR           Temporary build directory (default: $TEMP_BUILD_DIR)"
      echo "  --tag=TAG                Image tag (default: $IMAGE_TAG)"
      echo "  --skip-lambda-update     Skip Lambda function update"
      echo "  --skip-push              Skip pushing to ECR"
      echo "  --skip-ecr-login         Skip ECR login"
      echo "  --no-cache               Build without Docker cache"
      echo "  --help                   Show this help message"
      exit 0
      ;;
  esac
done

# Setup cleanup trap to always remove temporary files
cleanup() {
  local exit_code=$?
  
  if [ -d "$TEMP_BUILD_DIR" ]; then
    echo "Cleaning up temporary build directory..."
    rm -rf "$TEMP_BUILD_DIR"
  fi
  
  if [ $exit_code -ne 0 ]; then
    echo "❌ Build failed!"
  fi
}
trap cleanup EXIT

echo "=== ExecuteTurn1 Docker Build Script ==="
echo "Using ECR repository: $ECR_REPO"
echo "Function name: $FUNCTION_NAME"
echo "AWS region: $AWS_REGION"

# Check if shared packages directory exists
if [ ! -d "$SHARED_PACKAGES_DIR" ]; then
  echo "Error: Shared packages directory not found at $SHARED_PACKAGES_DIR"
  exit 1
fi

# Clean up any existing temporary build directory
if [ -d "$TEMP_BUILD_DIR" ]; then
  echo "Cleaning up existing temporary build directory..."
  rm -rf "$TEMP_BUILD_DIR"
fi

# Create temporary build directory with proper structure for Docker build
echo "Preparing build context with shared packages..."
mkdir -p "$TEMP_BUILD_DIR/workflow-function/shared"

# Copy shared packages to the temporary build directory with the correct structure
echo "Copying shared packages..."
cp -r "$SHARED_PACKAGES_DIR"/* "$TEMP_BUILD_DIR/workflow-function/shared/"

# Verify shared packages were copied correctly
if [ ! -d "$TEMP_BUILD_DIR/workflow-function/shared/errors" ] || \
   [ ! -d "$TEMP_BUILD_DIR/workflow-function/shared/schema" ] || \
   [ ! -d "$TEMP_BUILD_DIR/workflow-function/shared/logger" ]; then
  echo "Error: Failed to copy shared packages correctly"
  echo "Looking for: errors, schema, and logger packages"
  echo "Found: $(ls -la $TEMP_BUILD_DIR/workflow-function/shared/)"
  rm -rf "$TEMP_BUILD_DIR"
  exit 1
fi

# Copy current function files to the temporary build directory
echo "Copying function files..."
cp -r ./* "$TEMP_BUILD_DIR/"

# Fix module paths in go.mod for Docker build
echo "Fixing go.mod replacements for Docker build..."
if [ -f "$TEMP_BUILD_DIR/go.mod" ]; then
  sed -i.bak 's|replace (|// Commented out for Docker build\n// replace (|g' "$TEMP_BUILD_DIR/go.mod"
  sed -i.bak 's|workflow-function/shared/errors => ../shared/errors|// workflow-function/shared/errors => ../shared/errors|g' "$TEMP_BUILD_DIR/go.mod"
  sed -i.bak 's|workflow-function/shared/logger => ../shared/logger|// workflow-function/shared/logger => ../shared/logger|g' "$TEMP_BUILD_DIR/go.mod"
  sed -i.bak 's|workflow-function/shared/schema => ../shared/schema|// workflow-function/shared/schema => ../shared/schema|g' "$TEMP_BUILD_DIR/go.mod"
  sed -i.bak 's|)|// )|g' "$TEMP_BUILD_DIR/go.mod"
else
  echo "Warning: go.mod not found in temp directory - cannot fix replacements"
fi

# Use our simple Dockerfile for Docker build
if [ -f "Dockerfile.simple" ]; then
  echo "Using simplified build Dockerfile..."
  cp -f "Dockerfile.simple" "$TEMP_BUILD_DIR/Dockerfile"
else
  echo "Warning: Simplified Dockerfile.simple not found!"
fi

# Create simple module files for each shared package
echo "Creating simple module files for shared packages..."
for pkg in "errors" "schema" "logger"; do
  # Create a simple go.mod file in each module directory if it doesn't exist
  if [ ! -f "$TEMP_BUILD_DIR/workflow-function/shared/$pkg/go.mod" ]; then
    cat > "$TEMP_BUILD_DIR/workflow-function/shared/$pkg/go.mod" << EOF
module workflow-function/shared/$pkg

go 1.22
EOF
  fi
done

# Log in to ECR if not skipped
if [[ "$SKIP_ECR_LOGIN" == false ]]; then
  echo "Logging in to ECR..."
  if ! aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$(echo "$ECR_REPO" | cut -d'/' -f1)"; then
    echo "Error: Failed to log in to ECR"
    exit 1
  fi
fi

# Build the image from the temporary directory
echo "Building Docker image..."
BUILD_ARGS=""
if [[ "$NO_CACHE" == true ]]; then
  BUILD_ARGS="--no-cache"
fi

# Get host architecture
HOST_ARCH=$(uname -m)
if [[ "$HOST_ARCH" == "arm64" ]]; then
  echo "Building for ARM64 architecture..."
  PLATFORM_ARG="--platform=linux/arm64"
else
  echo "Building for AMD64 architecture..."
  # Auto-detect if we should use emulation
  PLATFORM_ARG="--platform=linux/arm64"
fi

if ! docker build $PLATFORM_ARG $BUILD_ARGS -t "${ECR_REPO}:${IMAGE_TAG}" "$TEMP_BUILD_DIR"; then
  echo "Error: Docker build failed"
  exit 1
fi

# Push the image if not skipped
if [[ "$SKIP_PUSH" == false ]]; then
  echo "Pushing image to ECR..."
  if ! docker push "${ECR_REPO}:${IMAGE_TAG}"; then
    echo "Error: Failed to push image to ECR"
    exit 1
  fi
  echo "✅ Image pushed successfully to ${ECR_REPO}:${IMAGE_TAG}"
else
  echo "✅ Image built successfully (push skipped)"
fi

# Deploy to AWS Lambda if not skipped
if [[ "$SKIP_LAMBDA_UPDATE" == false && "$SKIP_PUSH" == false ]]; then
  echo "Deploying to AWS Lambda..."
  if ! aws lambda update-function-code \
      --function-name "$FUNCTION_NAME" \
      --image-uri "${ECR_REPO}:${IMAGE_TAG}" \
      --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "Warning: Failed to update Lambda function code"
    echo "You may need to update the Lambda function manually"
  else
    echo "✅ Lambda function $FUNCTION_NAME updated successfully"
  fi
elif [[ "$SKIP_LAMBDA_UPDATE" == true ]]; then
  echo "ℹ️ Lambda function update skipped as requested"
elif [[ "$SKIP_PUSH" == true ]]; then
  echo "ℹ️ Lambda function update skipped because image push was skipped"
fi

echo "✅ Build process completed successfully"
