#!/bin/bash

# Deploy script for API Verifications Conversation Lambda Function
# This script builds the Docker image and pushes it to ECR

set -e

# Configuration from environment variables
ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-get-conversation-f6d3xl"
LAMBDA_FUNCTION_NAME="kootoro-dev-lambda-api-get-conversation-f6d3xl"
AWS_REGION="us-east-1"
IMAGE_TAG="latest"

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

# Check required environment variables
check_env_vars() {
    log_info "Checking required environment variables..."

    if [ -z "$ECR_REPO" ]; then
        log_error "ECR_REPO environment variable is required"
        log_info "Example: export ECR_REPO=123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo"
        exit 1
    fi

    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "LAMBDA_FUNCTION_NAME environment variable is required"
        log_info "Example: export LAMBDA_FUNCTION_NAME=my-lambda-function"
        exit 1
    fi

    log_success "All required environment variables are set"
    log_info "ECR Repository: $ECR_REPO"
    log_info "Lambda Function: $LAMBDA_FUNCTION_NAME"
    log_info "AWS Region: $AWS_REGION"
}

# Login to ECR
ecr_login() {
    log_info "Logging into ECR..."
    # Extract account ID from ECR_REPO
    AWS_ACCOUNT_ID=$(echo $ECR_REPO | cut -d'.' -f1)
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
    log_success "ECR login successful"
}

# Build and test the Go application
build_and_test() {
    log_info "Building and testing Go application..."

    # Download dependencies
    GOWORK=off go mod download
    GOWORK=off go mod tidy

    # Run tests
    log_info "Running tests..."
    GOWORK=off DYNAMODB_CONVERSATION_TABLE=test-table RESULTS_BUCKET=test-bucket go test -v

    # Build binary
    log_info "Building binary..."
    GOWORK=off go build -o api-verifications-conversation *.go

    log_success "Build and test completed successfully"
}

# Build Docker image
build_docker_image() {
    log_info "Building Docker image..."

    FULL_IMAGE_TAG="${ECR_REPO}:${IMAGE_TAG}"
    docker build -t api-verifications-conversation .
    docker tag api-verifications-conversation:latest $FULL_IMAGE_TAG

    log_success "Docker image built: $FULL_IMAGE_TAG"
}

# Push to ECR
push_to_ecr() {
    log_info "Pushing image to ECR..."

    FULL_IMAGE_TAG="${ECR_REPO}:${IMAGE_TAG}"
    docker push $FULL_IMAGE_TAG

    log_success "Image pushed to ECR: $FULL_IMAGE_TAG"
}

# Update Lambda function
update_lambda() {
    log_info "Updating Lambda function..."

    IMAGE_URI="${ECR_REPO}:${IMAGE_TAG}"

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

    # Create test payload file
    cat > test_payload.json << 'EOF'
{
  "httpMethod": "GET",
  "path": "/api/verifications/test-verification-id/conversation",
  "pathParameters": {
    "verificationId": "test-verification-id"
  },
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
    log_info "Starting deployment of API Verifications Conversation Lambda Function..."

    check_dependencies
    check_env_vars
    ecr_login
    build_and_test
    build_docker_image
    push_to_ecr
    update_lambda
    test_function

    log_success "Deployment completed successfully!"
    log_info "The API Verifications Conversation Lambda function is now deployed and ready to use."
    log_info "Endpoint: GET /api/verifications/{verificationId}/conversation"
}

# Basic Go operations
go_build() {
    log_info "Building Go binary..."
    GOWORK=off go build -o api-verifications-conversation *.go
    log_success "Binary built: api-verifications-conversation"
}

go_clean() {
    log_info "Cleaning up..."
    rm -f api-verifications-conversation
    log_success "Cleanup completed"
}

go_test() {
    log_info "Running Go tests..."
    GOWORK=off DYNAMODB_CONVERSATION_TABLE=test-table RESULTS_BUCKET=test-bucket go test -v
    log_success "Tests completed"
}

go_run() {
    log_info "Running Go application locally..."
    log_warning "Make sure to set environment variables:"
    log_info "  export DYNAMODB_CONVERSATION_TABLE=your-conversation-table"
    log_info "  export RESULTS_BUCKET=your-results-bucket"
    log_info "  export LOG_LEVEL=INFO"
    GOWORK=off go run *.go
}

go_deps() {
    log_info "Downloading and tidying Go dependencies..."
    GOWORK=off go mod download
    GOWORK=off go mod tidy
    log_success "Dependencies updated"
}

go_fmt() {
    log_info "Formatting Go code..."
    GOWORK=off go fmt ./...
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
        check_env_vars
        ecr_login
        build_and_test
        build_docker_image
        push_to_ecr
        ;;
    "update")
        log_info "Updating Lambda function only..."
        check_dependencies
        check_env_vars
        update_lambda
        ;;
    "test")
        log_info "Testing deployed function..."
        check_env_vars
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
        echo "Required Environment Variables:"
        echo "  ECR_REPO              ECR repository URI"
        echo "                        Example: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo"
        echo "  LAMBDA_FUNCTION_NAME  Lambda function name"
        echo "                        Example: my-lambda-function"
        echo ""
        echo "Optional Environment Variables:"
        echo "  AWS_REGION            AWS region (default: us-east-1)"
        echo "  IMAGE_TAG             Docker image tag (default: latest)"
        echo ""
        echo "Example usage:"
        echo "  export ECR_REPO=123456789012.dkr.ecr.us-east-1.amazonaws.com/my-conversation-repo"
        echo "  export LAMBDA_FUNCTION_NAME=my-conversation-function"
        echo "  ./deploy.sh"
        ;;
    *)
        log_error "Unknown command: $1"
        log_info "Use '$0 help' for usage information"
        exit 1
        ;;
esac
