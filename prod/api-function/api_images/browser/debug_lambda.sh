#!/bin/bash

# Debug script for API Images Browser Lambda Function
# This script helps diagnose issues with the Lambda function

set -e

# Configuration
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

# Check Lambda function status
check_lambda_function() {
    log_info "Checking Lambda function status..."
    
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)
    
    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found!"
        log_info "Available Lambda functions:"
        aws lambda list-functions --query "Functions[].FunctionName" --output table
        return 1
    fi
    
    log_success "Found Lambda function: $LAMBDA_FUNCTION_NAME"
    
    # Get function configuration
    log_info "Getting function configuration..."
    aws lambda get-function-configuration --function-name $LAMBDA_FUNCTION_NAME --region $AWS_REGION > lambda_config.json
    
    # Check environment variables
    log_info "Environment variables:"
    cat lambda_config.json | jq '.Environment.Variables' || log_warning "Could not parse environment variables"
    
    # Check function state
    STATE=$(cat lambda_config.json | jq -r '.State')
    log_info "Function state: $STATE"
    
    if [ "$STATE" != "Active" ]; then
        log_warning "Function is not in Active state!"
    fi
    
    # Check last update status
    LAST_UPDATE_STATUS=$(cat lambda_config.json | jq -r '.LastUpdateStatus')
    log_info "Last update status: $LAST_UPDATE_STATUS"
    
    if [ "$LAST_UPDATE_STATUS" != "Successful" ]; then
        log_warning "Last update was not successful!"
        LAST_UPDATE_STATUS_REASON=$(cat lambda_config.json | jq -r '.LastUpdateStatusReason')
        log_info "Reason: $LAST_UPDATE_STATUS_REASON"
    fi
    
    rm -f lambda_config.json
}

# Check S3 buckets
check_s3_buckets() {
    log_info "Checking S3 buckets..."
    
    # Get bucket names from Lambda environment variables
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)
    
    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found for bucket check"
        return 1
    fi
    
    REFERENCE_BUCKET=$(aws lambda get-function-configuration --function-name $LAMBDA_FUNCTION_NAME --region $AWS_REGION --query 'Environment.Variables.REFERENCE_BUCKET' --output text 2>/dev/null)
    CHECKING_BUCKET=$(aws lambda get-function-configuration --function-name $LAMBDA_FUNCTION_NAME --region $AWS_REGION --query 'Environment.Variables.CHECKING_BUCKET' --output text 2>/dev/null)
    
    log_info "Reference bucket: $REFERENCE_BUCKET"
    log_info "Checking bucket: $CHECKING_BUCKET"
    
    # Check if buckets exist
    if [ "$REFERENCE_BUCKET" != "None" ] && [ -n "$REFERENCE_BUCKET" ]; then
        if aws s3 ls "s3://$REFERENCE_BUCKET" > /dev/null 2>&1; then
            log_success "Reference bucket exists and is accessible"
        else
            log_error "Reference bucket does not exist or is not accessible"
        fi
    else
        log_error "REFERENCE_BUCKET environment variable not set"
    fi
    
    if [ "$CHECKING_BUCKET" != "None" ] && [ -n "$CHECKING_BUCKET" ]; then
        if aws s3 ls "s3://$CHECKING_BUCKET" > /dev/null 2>&1; then
            log_success "Checking bucket exists and is accessible"
        else
            log_error "Checking bucket does not exist or is not accessible"
        fi
    else
        log_error "CHECKING_BUCKET environment variable not set"
    fi
}

# Check API Gateway integration
check_api_gateway() {
    log_info "Checking API Gateway integration..."
    
    # Find API Gateway that contains 'verification'
    API_ID=$(aws apigateway get-rest-apis --query "items[?contains(name, 'verification')].id" --output text 2>/dev/null | head -1)
    
    if [ -z "$API_ID" ]; then
        log_warning "API Gateway not found"
        log_info "Available APIs:"
        aws apigateway get-rest-apis --query "items[].[name,id]" --output table 2>/dev/null || log_warning "Could not list APIs"
        return 1
    fi
    
    log_success "Found API Gateway: $API_ID"
    
    # Get API Gateway URL
    API_URL="https://${API_ID}.execute-api.${AWS_REGION}.amazonaws.com/v1"
    log_info "API Gateway URL: $API_URL"
    
    # Test the endpoint
    log_info "Testing API endpoint..."
    curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/images/browser?bucketType=reference" > response_code.txt
    RESPONSE_CODE=$(cat response_code.txt)
    rm -f response_code.txt
    
    log_info "Response code: $RESPONSE_CODE"
    
    if [ "$RESPONSE_CODE" = "200" ]; then
        log_success "API endpoint is working"
    else
        log_error "API endpoint returned error code: $RESPONSE_CODE"
    fi
}

# Check CloudWatch logs
check_logs() {
    log_info "Checking CloudWatch logs..."
    
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)
    
    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found for log check"
        return 1
    fi
    
    LOG_GROUP="/aws/lambda/$LAMBDA_FUNCTION_NAME"
    
    log_info "Log group: $LOG_GROUP"
    
    # Check if log group exists
    if aws logs describe-log-groups --log-group-name-prefix "$LOG_GROUP" --region $AWS_REGION | grep -q "$LOG_GROUP"; then
        log_success "Log group exists"
        
        # Get recent log events
        log_info "Recent log events (last 10):"
        aws logs filter-log-events --log-group-name "$LOG_GROUP" --region $AWS_REGION --limit 10 --query 'events[].[timestamp,message]' --output table 2>/dev/null || log_warning "Could not retrieve log events"
    else
        log_warning "Log group does not exist"
    fi
}

# Test Lambda function directly
test_lambda_direct() {
    log_info "Testing Lambda function directly..."
    
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)
    
    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found for direct test"
        return 1
    fi
    
    # Create test payload
    cat > test_payload.json << 'EOF'
{
  "httpMethod": "GET",
  "path": "/api/images/browser",
  "queryStringParameters": {
    "bucketType": "reference"
  },
  "pathParameters": {},
  "headers": {
    "Content-Type": "application/json"
  }
}
EOF
    
    log_info "Invoking Lambda function directly..."
    aws lambda invoke \
        --function-name $LAMBDA_FUNCTION_NAME \
        --payload file://test_payload.json \
        --region $AWS_REGION \
        response.json
    
    if [ $? -eq 0 ]; then
        log_info "Lambda response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
        
        # Check for errors in response
        if cat response.json | grep -q '"statusCode": 500'; then
            log_error "Lambda function returned 500 error"
        elif cat response.json | grep -q '"statusCode": 200'; then
            log_success "Lambda function executed successfully"
        fi
    else
        log_error "Lambda invocation failed"
    fi
    
    rm -f test_payload.json response.json
}

# Main diagnostic function
diagnose() {
    log_info "Starting diagnostic for API Images Browser Lambda Function..."
    echo "=================================================="
    
    check_lambda_function
    echo "=================================================="
    
    check_s3_buckets
    echo "=================================================="
    
    check_api_gateway
    echo "=================================================="
    
    check_logs
    echo "=================================================="
    
    test_lambda_direct
    echo "=================================================="
    
    log_info "Diagnostic completed!"
}

# Parse command line arguments
case "${1:-diagnose}" in
    "lambda")
        check_lambda_function
        ;;
    "buckets")
        check_s3_buckets
        ;;
    "api")
        check_api_gateway
        ;;
    "logs")
        check_logs
        ;;
    "test")
        test_lambda_direct
        ;;
    "diagnose"|"")
        diagnose
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  diagnose  Run full diagnostic [default]"
        echo "  lambda    Check Lambda function status"
        echo "  buckets   Check S3 buckets"
        echo "  api       Check API Gateway integration"
        echo "  logs      Check CloudWatch logs"
        echo "  test      Test Lambda function directly"
        echo "  help      Show this help message"
        ;;
    *)
        log_error "Unknown command: $1"
        log_info "Use '$0 help' for usage information"
        exit 1
        ;;
esac
