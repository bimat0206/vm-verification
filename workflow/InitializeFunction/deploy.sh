#!/bin/bash

# Colors for better readability
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default configuration
FUNCTION_NAME="vending-verification-initialize"
AWS_REGION="us-east-1"
LAMBDA_ROLE="arn:aws:iam::$(aws sts get-caller-identity --query 'Account' --output text):role/lambda-${FUNCTION_NAME}-role"
ECR_REPOSITORY="$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.${AWS_REGION}.amazonaws.com/${FUNCTION_NAME}"
IMAGE_TAG="latest"

# Lambda environment variables
DYNAMODB_LAYOUT_TABLE="VendingMachineLayoutMetadata"
DYNAMODB_VERIFICATION_TABLE="VerificationResults"
VERIFICATION_PREFIX="verif-"
REFERENCE_BUCKET="vending-machine-verification-image-reference-a11"
CHECKING_BUCKET="vending-machine-verification-image-checking-a12"

# Lambda configuration
LAMBDA_MEMORY="512"
LAMBDA_TIMEOUT="30"

# Print a message with proper formatting
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print a section header
print_section() {
    local message=$1
    echo -e "\n${BLUE}=== ${message} ===${NC}"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for prerequisites
check_prerequisites() {
    print_section "Checking Prerequisites"
    
    local missing=0
    
    if ! command_exists aws; then
        print_message "${RED}" "AWS CLI is not installed. Please install it first."
        missing=1
    else
        print_message "${GREEN}" "✓ AWS CLI is installed"
    fi
    
    if ! command_exists docker; then
        print_message "${RED}" "Docker is not installed. Please install it first."
        missing=1
    else
        print_message "${GREEN}" "✓ Docker is installed"
    fi
    
    # Check AWS CLI configuration
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        print_message "${RED}" "AWS CLI is not configured properly. Please run 'aws configure'."
        missing=1
    else
        print_message "${GREEN}" "✓ AWS CLI is configured"
        print_message "${GREEN}" "  Using AWS Account: $(aws sts get-caller-identity --query 'Account' --output text)"
        print_message "${GREEN}" "  Region: ${AWS_REGION}"
    fi
    
    if [ $missing -eq 1 ]; then
        print_message "${RED}" "Please install the missing prerequisites and try again."
        exit 1
    fi
}

# Create a JSON file with Lambda environment variables
create_env_file() {
    print_section "Creating Environment Variables File"
    cat > env-vars.json << EOF
{
  "Variables": {
    "DYNAMODB_LAYOUT_TABLE": "${DYNAMODB_LAYOUT_TABLE}",
    "DYNAMODB_VERIFICATION_TABLE": "${DYNAMODB_VERIFICATION_TABLE}",
    "VERIFICATION_PREFIX": "${VERIFICATION_PREFIX}",
    "REFERENCE_BUCKET": "${REFERENCE_BUCKET}",
    "CHECKING_BUCKET": "${CHECKING_BUCKET}"
  }
}
EOF
    print_message "${GREEN}" "Created env-vars.json with Lambda environment variables"
    print_message "${YELLOW}" "Environment file contents:"
    cat env-vars.json
}

# Build the Docker image
build_docker_image() {
    print_section "Building Docker Image"
    print_message "${YELLOW}" "Building image: ${FUNCTION_NAME}:${IMAGE_TAG}"
    
    docker build -t ${FUNCTION_NAME}:${IMAGE_TAG} . || {
        print_message "${RED}" "Warning: Docker build encountered issues, but continuing with deployment"
        print_message "${YELLOW}" "This may affect later steps if the image wasn't built correctly"
    }
    
    print_message "${GREEN}" "Docker build process completed"
}

# Push the Docker image to ECR
push_to_ecr() {
    print_section "Pushing Image to ECR"
    
    # Login to ECR
    print_message "${YELLOW}" "Logging into ECR..."
    aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin $(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.${AWS_REGION}.amazonaws.com || true
    
    # Create repository if it doesn't exist
    print_message "${YELLOW}" "Checking if ECR repository exists..."
    aws ecr describe-repositories --repository-names ${FUNCTION_NAME} --region ${AWS_REGION} >/dev/null 2>&1 || {
        print_message "${YELLOW}" "Creating ECR repository: ${FUNCTION_NAME}"
        aws ecr create-repository --repository-name ${FUNCTION_NAME} --region ${AWS_REGION} || true
    }
    
    # Tag and push the image
    print_message "${YELLOW}" "Tagging and pushing image to ECR..."
    docker tag ${FUNCTION_NAME}:${IMAGE_TAG} ${ECR_REPOSITORY}:${IMAGE_TAG} || true
    
    docker push ${ECR_REPOSITORY}:${IMAGE_TAG} || {
        print_message "${RED}" "Warning: Push to ECR had issues but continuing with deployment"
    }
    
    print_message "${GREEN}" "ECR push process completed"
}

# Create IAM role for Lambda
create_role() {
    print_section "Creating/Updating IAM Role"
    
    # Check if role exists
    aws iam get-role --role-name lambda-${FUNCTION_NAME}-role --region ${AWS_REGION} >/dev/null 2>&1
    local role_exists=$?
    
    if [ $role_exists -eq 0 ]; then
        print_message "${GREEN}" "IAM role already exists: lambda-${FUNCTION_NAME}-role"
    else
        print_message "${YELLOW}" "Creating IAM role: lambda-${FUNCTION_NAME}-role"
        aws iam create-role \
            --role-name lambda-${FUNCTION_NAME}-role \
            --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
            --region ${AWS_REGION} || true
        
        # Allow some time for IAM role propagation
        print_message "${YELLOW}" "Waiting for IAM role propagation..."
        sleep 10
    fi
    
    # Attach policies
    print_message "${YELLOW}" "Attaching policies to role..."
    
    # Lambda basic execution
    aws iam attach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole \
        --region ${AWS_REGION} 2>/dev/null || true
    
    # DynamoDB read
    aws iam attach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBReadOnlyAccess \
        --region ${AWS_REGION} 2>/dev/null || true
    
    # S3 read
    aws iam attach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
        --region ${AWS_REGION} 2>/dev/null || true
    
    # Custom policy for DynamoDB write
    print_message "${YELLOW}" "Creating custom policy for DynamoDB write permissions..."
    aws iam put-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-name ${FUNCTION_NAME}-dynamodb-write \
        --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:PutItem","dynamodb:UpdateItem","dynamodb:BatchWriteItem"],"Resource":"arn:aws:dynamodb:'${AWS_REGION}':'$(aws sts get-caller-identity --query "Account" --output text)':table/'${DYNAMODB_VERIFICATION_TABLE}'"}]}' \
        --region ${AWS_REGION} || true
    
    print_message "${GREEN}" "IAM role setup completed"
    
    # Additional delay to ensure role permissions have propagated
    print_message "${YELLOW}" "Giving IAM permissions time to propagate..."
    sleep 5
}

# Deploy or update Lambda function
deploy_lambda() {
    print_section "Deploying Lambda Function"
    create_env_file
    
    # Check if function exists
    aws lambda get-function --function-name ${FUNCTION_NAME} --region ${AWS_REGION} >/dev/null 2>&1
    local function_exists=$?
    
    if [ $function_exists -eq 0 ]; then
        print_message "${YELLOW}" "Updating existing Lambda function: ${FUNCTION_NAME}"
        
        # Update function code
        print_message "${YELLOW}" "Updating function code..."
        aws lambda update-function-code \
            --function-name ${FUNCTION_NAME} \
            --image-uri ${ECR_REPOSITORY}:${IMAGE_TAG} \
            --region ${AWS_REGION} || true
        
        # Wait for code update to finish
        print_message "${YELLOW}" "Waiting for code update to complete..."
        sleep 5
        
        # Update function configuration
        print_message "${YELLOW}" "Updating function configuration..."
        aws lambda update-function-configuration \
            --function-name ${FUNCTION_NAME} \
            --timeout ${LAMBDA_TIMEOUT} \
            --memory-size ${LAMBDA_MEMORY} \
            --environment file://env-vars.json \
            --region ${AWS_REGION} || true
    else
        print_message "${YELLOW}" "Creating new Lambda function: ${FUNCTION_NAME}"
        aws lambda create-function \
            --function-name ${FUNCTION_NAME} \
            --package-type Image \
            --code ImageUri=${ECR_REPOSITORY}:${IMAGE_TAG} \
            --role ${LAMBDA_ROLE} \
            --architectures arm64 \
            --timeout ${LAMBDA_TIMEOUT} \
            --memory-size ${LAMBDA_MEMORY} \
            --environment file://env-vars.json \
            --region ${AWS_REGION} || true
    fi
    
    print_message "${GREEN}" "Lambda function deployment initiated"
    print_message "${YELLOW}" "Note: The function update might take a few moments to complete"
    print_message "${YELLOW}" "You can check the status with option 7 (View Lambda logs)"
}

# Invoke Lambda with test payload
invoke_lambda() {
    print_section "Invoking Lambda Function"
    
    print_message "${YELLOW}" "Creating test payload..."
    local payload='{
  "referenceImageUrl": "s3://vending-machine-verification-image-reference-a11/processed/2025/04/22/41927_54mf04d1_reference_image.png",
  "checkingImageUrl": "s3://vending-machine-verification-image-checking-a12/AACZ 1.png",
  "vendingMachineId": "VM-3245",
  "layoutId": 41927,
  "layoutPrefix": "54mf04d1",
  "conversationConfig": {
    "type": "two-turn",
    "maxTurns": 2
  },
  "requestId": "req-8a72b936-d1c5-4f4a-9b7e-fb2c75e93d1b",
  "requestTimestamp": "2025-04-21T15:30:20Z"
}'
    
    print_message "${YELLOW}" "Invoking Lambda function..."
    aws lambda invoke \
        --function-name ${FUNCTION_NAME} \
        --payload "$payload" \
        --cli-binary-format raw-in-base64-out \
        response.json \
        --region ${AWS_REGION} >/dev/null 2>&1 || {
        print_message "${YELLOW}" "Invocation may have had issues, checking response file..."
    }
    
    if [ -f response.json ] && [ -s response.json ]; then
        print_message "${GREEN}" "Response received:"
        cat response.json
    else
        print_message "${YELLOW}" "No valid response received. Function may still be updating or had an error."
        print_message "${YELLOW}" "Check logs with option 7 (View Lambda logs) for more details."
    fi
}

# View Lambda logs
view_logs() {
    print_section "Viewing Lambda Logs"
    
    print_message "${YELLOW}" "Checking Lambda function status..."
    aws lambda get-function \
        --function-name ${FUNCTION_NAME} \
        --region ${AWS_REGION} \
        --query 'Configuration.[State,LastUpdateStatus]' \
        --output text 2>/dev/null || {
        print_message "${YELLOW}" "Function status check failed, but continuing to log retrieval"
    }
    
    print_message "${YELLOW}" "Fetching logs from CloudWatch..."
    aws logs filter-log-events \
        --log-group-name /aws/lambda/${FUNCTION_NAME} \
        --limit 25 \
        --region ${AWS_REGION} \
        --query 'events[*].message' \
        --output text 2>/dev/null || {
        print_message "${YELLOW}" "No logs found or log group doesn't exist yet"
        print_message "${YELLOW}" "This is normal for new functions or if the function hasn't been invoked yet"
    }
    
    print_message "${GREEN}" "Log retrieval process completed"
}

# Delete Lambda and associated resources
delete_resources() {
    print_section "Deleting Resources"
    
    read -p "Are you sure you want to delete all resources? (y/n): " confirm
    if [[ $confirm != [yY] ]]; then
        print_message "${YELLOW}" "Deletion cancelled"
        return
    fi
    
    # Delete Lambda function
    print_message "${YELLOW}" "Deleting Lambda function..."
    if aws lambda delete-function --function-name ${FUNCTION_NAME} --region ${AWS_REGION} >/dev/null 2>&1; then
        print_message "${GREEN}" "Lambda function deleted"
    else
        print_message "${YELLOW}" "Lambda function doesn't exist or already deleted"
    fi
    
    # Delete ECR repository
    print_message "${YELLOW}" "Deleting ECR repository..."
    if aws ecr delete-repository --repository-name ${FUNCTION_NAME} --force --region ${AWS_REGION} >/dev/null 2>&1; then
        print_message "${GREEN}" "ECR repository deleted"
    else
        print_message "${YELLOW}" "ECR repository doesn't exist or already deleted"
    fi
    
    # Delete IAM role and policies
    print_message "${YELLOW}" "Cleaning up IAM role and policies..."
    
    # Detach policies
    aws iam detach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole \
        --region ${AWS_REGION} 2>/dev/null || true
    
    aws iam detach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBReadOnlyAccess \
        --region ${AWS_REGION} 2>/dev/null || true
    
    aws iam detach-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
        --region ${AWS_REGION} 2>/dev/null || true
    
    # Delete custom policy
    aws iam delete-role-policy \
        --role-name lambda-${FUNCTION_NAME}-role \
        --policy-name ${FUNCTION_NAME}-dynamodb-write \
        --region ${AWS_REGION} 2>/dev/null || true
    
    # Delete role
    if aws iam delete-role \
        --role-name lambda-${FUNCTION_NAME}-role \
        --region ${AWS_REGION} 2>/dev/null; then
        print_message "${GREEN}" "IAM role deleted"
    else
        print_message "${YELLOW}" "IAM role doesn't exist or already deleted"
    fi
    
    # Clean up local files
    rm -f env-vars.json response.json 2>/dev/null
    
    print_message "${GREEN}" "All resources cleaned up successfully"
}

# Clean up local resources
clean_local() {
    print_section "Cleaning Local Resources"
    
    print_message "${YELLOW}" "Removing temporary files..."
    rm -f env-vars.json response.json 2>/dev/null
    
    print_message "${YELLOW}" "Removing Docker images..."
    if docker images | grep -q ${FUNCTION_NAME}; then
        docker rmi ${FUNCTION_NAME}:${IMAGE_TAG} 2>/dev/null || true
    fi
    
    if docker images | grep -q ${ECR_REPOSITORY}; then
        docker rmi ${ECR_REPOSITORY}:${IMAGE_TAG} 2>/dev/null || true
    fi
    
    print_message "${GREEN}" "Local resources cleaned up successfully"
}

# Display the main menu
show_menu() {
    echo ""
    print_message "${BLUE}" "=== Vending Machine Verification Lambda Deployment ==="
    echo ""
    print_message "${YELLOW}" "Configuration:"
    echo "  Function Name:    ${FUNCTION_NAME}"
    echo "  AWS Region:       ${AWS_REGION}"
    echo "  ECR Repository:   ${ECR_REPOSITORY}"
    echo ""
    print_message "${BLUE}" "Available Operations:"
    echo "  1) Check prerequisites"
    echo "  2) Build Docker image"
    echo "  3) Push image to ECR"
    echo "  4) Create/update IAM role"
    echo "  5) Deploy Lambda function"
    echo "  6) Invoke Lambda (test)"
    echo "  7) View Lambda logs"
    echo "  8) Clean local resources"
    echo "  9) Delete all AWS resources"
    echo ""
    echo "  A) Full deployment (steps 1-5)"
    echo "  B) Edit configuration"
    echo ""
    echo "  0) Exit"
    echo ""
    read -p "Enter your choice: " choice
    
    case $choice in
        1) check_prerequisites; show_menu ;;
        2) build_docker_image; show_menu ;;
        3) push_to_ecr; show_menu ;;
        4) create_role; show_menu ;;
        5) deploy_lambda; show_menu ;;
        6) invoke_lambda; show_menu ;;
        7) view_logs; show_menu ;;
        8) clean_local; show_menu ;;
        9) delete_resources; show_menu ;;
        [Aa]) 
            check_prerequisites
            build_docker_image
            push_to_ecr
            create_role
            deploy_lambda
            show_menu 
            ;;
        [Bb]) edit_configuration; show_menu ;;
        0) exit 0 ;;
        *) print_message "${RED}" "Invalid choice"; show_menu ;;
    esac
}

# Edit configuration
edit_configuration() {
    print_section "Edit Configuration"
    
    echo "Current Configuration:"
    echo "  1) Function Name:           ${FUNCTION_NAME}"
    echo "  2) AWS Region:              ${AWS_REGION}"
    echo "  3) Lambda Memory (MB):      ${LAMBDA_MEMORY}"
    echo "  4) Lambda Timeout (sec):    ${LAMBDA_TIMEOUT}"
    echo ""
    echo "Environment Variables:"
    echo "  5) DynamoDB Layout Table:   ${DYNAMODB_LAYOUT_TABLE}"
    echo "  6) DynamoDB Verif. Table:   ${DYNAMODB_VERIFICATION_TABLE}"
    echo "  7) Verification Prefix:     ${VERIFICATION_PREFIX}"
    echo "  8) Reference Bucket:        ${REFERENCE_BUCKET}"
    echo "  9) Checking Bucket:         ${CHECKING_BUCKET}"
    echo ""
    echo "  0) Back to main menu"
    echo ""
    read -p "Enter number to change (0-9): " config_choice
    
    case $config_choice in
        1) read -p "Enter new Function Name: " FUNCTION_NAME; edit_configuration ;;
        2) read -p "Enter new AWS Region: " AWS_REGION; edit_configuration ;;
        3) read -p "Enter new Lambda Memory (MB): " LAMBDA_MEMORY; edit_configuration ;;
        4) read -p "Enter new Lambda Timeout (sec): " LAMBDA_TIMEOUT; edit_configuration ;;
        5) read -p "Enter new DynamoDB Layout Table: " DYNAMODB_LAYOUT_TABLE; edit_configuration ;;
        6) read -p "Enter new DynamoDB Verification Table: " DYNAMODB_VERIFICATION_TABLE; edit_configuration ;;
        7) read -p "Enter new Verification Prefix: " VERIFICATION_PREFIX; edit_configuration ;;
        8) read -p "Enter new Reference Bucket: " REFERENCE_BUCKET; edit_configuration ;;
        9) read -p "Enter new Checking Bucket: " CHECKING_BUCKET; edit_configuration ;;
        0) return ;;
        *) print_message "${RED}" "Invalid choice"; edit_configuration ;;
    esac
    
    # Update role and repository based on updated function name
    LAMBDA_ROLE="arn:aws:iam::$(aws sts get-caller-identity --query 'Account' --output text):role/lambda-${FUNCTION_NAME}-role"
    ECR_REPOSITORY="$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.${AWS_REGION}.amazonaws.com/${FUNCTION_NAME}"
}

# Start the script
check_prerequisites
show_menu