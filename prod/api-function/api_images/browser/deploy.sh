#!/bin/bash

# Deploy script for API Images Browser Lambda Function
# This script builds the Docker image and pushes it to ECR

set -e

# Configuration
FUNCTION_NAME="kootoro-dev-lambda-api-images-browser-f6d3xl"
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-images-browser-f6d3xl"
IMAGE_TAG="latest"
AWS_REGION="us-east-1"



# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    log_success "All dependencies are available"
}

# Get ECR repository URL from AWS
get_ecr_repository() {
    log_info "Getting ECR repository URL..."

    # Get AWS Account ID
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

    # Get the repository name that contains 'api-images-browser'
    REPO_NAME=$(aws ecr describe-repositories --region $AWS_REGION --query "repositories[?contains(repositoryName, 'api-images-browser')].repositoryName" --output text 2>/dev/null | head -1)

    if [ -z "$REPO_NAME" ]; then
        log_error "ECR repository not found. Please ensure Terraform has been applied and the ECR repository exists."
        log_info "Expected repository name pattern: *api-images-browser*"
        log_info "Available repositories:"
        aws ecr describe-repositories --region $AWS_REGION --query "repositories[].repositoryName" --output table 2>/dev/null || log_warning "Could not list repositories"
        exit 1
    fi

    ECR_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${REPO_NAME}"
    log_success "ECR repository: $ECR_REPO"
}

# Login to ECR
ecr_login() {
    log_info "Logging into ECR..."
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
    log_success "ECR login successful"
}

# Build and test the Go application
build_and_test() {
    log_info "Building and testing Go application..."

    # Download dependencies
    go mod download
    go mod tidy

    # Run tests
    log_info "Running tests..."
    REFERENCE_BUCKET=test-ref CHECKING_BUCKET=test-check go test -v

    # Build binary
    log_info "Building binary..."
    go build -o api-images-browser *.go

    log_success "Build and test completed successfully"
}

# Build Docker image
build_docker_image() {
    log_info "Building Docker image..."

    IMAGE_TAG="${ECR_REPO}:latest"
    docker build -t $FUNCTION_NAME .
    docker tag $FUNCTION_NAME:latest $IMAGE_TAG

    log_success "Docker image built: $IMAGE_TAG"
}

# Push to ECR
push_to_ecr() {
    log_info "Pushing image to ECR..."

    IMAGE_TAG="${ECR_REPO}:latest"
    docker push $IMAGE_TAG

    log_success "Image pushed to ECR: $IMAGE_TAG"
}

# Update Lambda function
update_lambda() {
    log_info "Updating Lambda function..."

    # Get Lambda function name from Terraform or use pattern
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)

    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found. Please ensure Terraform has been applied and the Lambda function exists."
        exit 1
    fi

    IMAGE_URI="${ECR_REPO}:latest"

    aws lambda update-function-code \
        --function-name $LAMBDA_FUNCTION_NAME \
        --image-uri $IMAGE_URI \
        --region $AWS_REGION > /dev/null 2>&1

    log_success "Lambda function updated: $LAMBDA_FUNCTION_NAME"

    # Wait for update to complete
    log_info "Waiting for function update to complete..."
    aws lambda wait function-updated --function-name $LAMBDA_FUNCTION_NAME --region $AWS_REGION
    log_success "Function update completed"
}

# Test the deployed function
test_function() {
    log_info "Testing deployed function..."

    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)

    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_warning "Lambda function not found for testing"
        return
    fi

    # Create test payload file
    cat > test_payload.json << 'EOF'
{
  "httpMethod": "GET",
  "path": "/api/images/browser",
  "queryStringParameters": {
    "bucketType": "reference"
  },
  "pathParameters": null,
  "headers": {
    "Content-Type": "application/json"
  }
}
EOF

    log_info "Invoking function with test payload..."
    aws lambda invoke \
        --function-name $LAMBDA_FUNCTION_NAME \
        --payload file://test_payload.json \
        --region $AWS_REGION \
        response.json

    if [ $? -eq 0 ]; then
        log_success "Function invocation successful"
        log_info "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
        rm -f response.json test_payload.json
    else
        log_error "Function invocation failed"
        rm -f test_payload.json
        exit 1
    fi
}

# Main deployment function
deploy() {
    log_info "Starting deployment of API Images Browser Lambda Function..."

    check_dependencies
    get_ecr_repository
    ecr_login
    build_and_test
    build_docker_image
    push_to_ecr
    update_lambda
    test_function

    log_success "Deployment completed successfully!"
    log_info "The API Images Browser Lambda function is now deployed and ready to use."
    log_info "Endpoint: GET /api/images/browser"
    log_info "Endpoint: GET /api/images/browser/{path+}"
}

# Basic Go operations
go_build() {
    log_info "Building Go binary..."
    go build -o api-images-browser *.go
    log_success "Binary built: api-images-browser"
}

go_clean() {
    log_info "Cleaning up..."
    rm -f api-images-browser
    log_success "Cleanup completed"
}

go_test() {
    log_info "Running Go tests..."
    REFERENCE_BUCKET=test-ref CHECKING_BUCKET=test-check go test -v ./...
    log_success "Tests completed"
}

go_run() {
    log_info "Running Go application locally..."
    log_warning "Make sure to set environment variables:"
    log_info "  export REFERENCE_BUCKET=your-reference-bucket"
    log_info "  export CHECKING_BUCKET=your-checking-bucket"
    log_info "  export LOG_LEVEL=INFO"
    go run *.go
}

go_deps() {
    log_info "Downloading and tidying Go dependencies..."
    go mod download
    go mod tidy
    log_success "Dependencies updated"
}

go_fmt() {
    log_info "Formatting Go code..."
    go fmt ./...
    log_success "Code formatted"
}

# Parse command line arguments
case "${1:-deploy}" in
    "build")
        log_info "Building Docker image only..."
        check_dependencies
        build_and_test
        build_docker_image
        ;;
    "push")
        log_info "Building and pushing to ECR..."
        check_dependencies
        get_ecr_repository
        ecr_login
        build_and_test
        build_docker_image
        push_to_ecr
        ;;
    "update")
        log_info "Updating Lambda function only..."
        check_dependencies
        get_ecr_repository
        update_lambda
        ;;
    "test")
        log_info "Testing deployed function..."
        test_function
        ;;
    "deploy"|"")
        deploy
        ;;
    "go-build")
        go_build
        ;;
    "go-clean")
        go_clean
        ;;
    "go-test")
        go_test
        ;;
    "go-run")
        go_run
        ;;
    "go-deps")
        go_deps
        ;;
    "go-fmt")
        go_fmt
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Deployment Commands:"
        echo "  deploy    Full deployment (build, push, update) [default]"
        echo "  build     Build Docker image only"
        echo "  push      Build and push to ECR"
        echo "  update    Update Lambda function with latest ECR image"
        echo "  test      Test the deployed function"
        echo ""
        echo "Go Development Commands:"
        echo "  go-build  Build Go binary"
        echo "  go-clean  Clean up binary"
        echo "  go-test   Run Go tests"
        echo "  go-run    Run Go application locally"
        echo "  go-deps   Download and tidy Go dependencies"
        echo "  go-fmt    Format Go code"
        echo ""
        echo "  help      Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  AWS_REGION    AWS region (default: us-east-1)"
        ;;
    *)
        log_error "Unknown command: $1"
        log_info "Use '$0 help' for usage information"
        exit 1
        ;;
esac
