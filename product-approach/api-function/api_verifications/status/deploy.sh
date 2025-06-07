#!/bin/bash

# API Verifications Status Lambda Function Deployment Script
# This script builds, packages, and deploys the Go-based Lambda function for verification status checking

set -e

# Configuration
LAMBDA_FUNCTION_NAME="api-verifications-status"
ECR_REPO_NAME="api-verifications-status"
IMAGE_TAG="latest"
AWS_REGION="us-east-1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.20 or later."
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install AWS CLI."
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Check environment variables
check_env_vars() {
    log_info "Checking environment variables..."
    
    if [ -z "$AWS_REGION" ]; then
        log_error "AWS_REGION is not set"
        exit 1
    fi
    
    log_success "Environment variables are set"
}

# Get ECR repository URI
get_ecr_repository() {
    log_info "Getting ECR repository URI..."
    
    # Check if repository exists, create if it doesn't
    if ! aws ecr describe-repositories --repository-names $ECR_REPO_NAME --region $AWS_REGION > /dev/null 2>&1; then
        log_info "ECR repository doesn't exist. Creating..."
        aws ecr create-repository --repository-name $ECR_REPO_NAME --region $AWS_REGION > /dev/null
        log_success "ECR repository created: $ECR_REPO_NAME"
    fi
    
    ECR_REPO=$(aws ecr describe-repositories --repository-names $ECR_REPO_NAME --region $AWS_REGION --query 'repositories[0].repositoryUri' --output text)
    log_success "ECR repository URI: $ECR_REPO"
}

# Login to ECR
ecr_login() {
    log_info "Logging in to ECR..."
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_REPO
    log_success "Successfully logged in to ECR"
}

# Build and test Go application
build_and_test() {
    log_info "Building and testing Go application..."
    
    # Download dependencies
    go mod tidy
    
    # Build the application
    GOWORK=off go build -o api-verifications-status *.go
    
    if [ $? -eq 0 ]; then
        log_success "Go application built successfully"
    else
        log_error "Failed to build Go application"
        exit 1
    fi
}

# Build Docker image
build_docker_image() {
    log_info "Building Docker image..."
    
    FULL_IMAGE_TAG="${ECR_REPO}:${IMAGE_TAG}"
    docker build -t $FULL_IMAGE_TAG .
    
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
  "path": "/api/verifications/status/test-verification-id",
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
    log_info "Starting deployment of API Verifications Status Lambda Function..."

    check_dependencies
    check_env_vars
    get_ecr_repository
    ecr_login
    build_and_test
    build_docker_image
    push_to_ecr
    update_lambda
    test_function

    log_success "Deployment completed successfully!"
    log_info "The API Verifications Status Lambda function is now deployed and ready to use."
    log_info "Endpoint: GET /api/verifications/status/{verificationId}"
}

# Basic Go operations
go_build() {
    log_info "Building Go binary..."
    GOWORK=off go build -o api-verifications-status *.go
    log_success "Binary built: api-verifications-status"
}

go_clean() {
    log_info "Cleaning up..."
    rm -f api-verifications-status
    log_success "Cleanup completed"
}

# Help function
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  deploy    Deploy the Lambda function (default)"
    echo "  build     Build the Go binary only"
    echo "  clean     Clean up build artifacts"
    echo "  help      Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  AWS_REGION              AWS region (default: us-east-1)"
    echo "  LAMBDA_FUNCTION_NAME    Lambda function name (default: api-verifications-status)"
    echo "  ECR_REPO_NAME          ECR repository name (default: api-verifications-status)"
    echo ""
}

# Main script logic
case "${1:-deploy}" in
    "deploy")
        deploy
        ;;
    "build")
        go_build
        ;;
    "clean")
        go_clean
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
