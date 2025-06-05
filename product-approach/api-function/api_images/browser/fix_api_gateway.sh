#!/bin/bash

# Fix API Gateway integration for API Images Browser
# This script addresses the 405 Method Not Allowed error

set -e

# Configuration
AWS_REGION="us-east-1"
API_ID="hpux2uegnd"

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

# Get Lambda function ARN
get_lambda_arn() {
    log_info "Getting Lambda function ARN..."
    
    LAMBDA_FUNCTION_NAME=$(aws lambda list-functions --query "Functions[?contains(FunctionName, 'api-images-browser')].FunctionName" --output text 2>/dev/null | head -1)
    
    if [ -z "$LAMBDA_FUNCTION_NAME" ]; then
        log_error "Lambda function not found!"
        exit 1
    fi
    
    LAMBDA_ARN=$(aws lambda get-function --function-name $LAMBDA_FUNCTION_NAME --query 'Configuration.FunctionArn' --output text)
    log_success "Lambda ARN: $LAMBDA_ARN"
}

# Check current API Gateway resources
check_resources() {
    log_info "Checking API Gateway resources..."
    
    # Get all resources
    aws apigateway get-resources --rest-api-id $API_ID > resources.json
    
    # Check if browser resource exists
    BROWSER_RESOURCE_ID=$(cat resources.json | jq -r '.items[] | select(.pathPart == "browser") | .id' 2>/dev/null)
    
    if [ -z "$BROWSER_RESOURCE_ID" ] || [ "$BROWSER_RESOURCE_ID" = "null" ]; then
        log_error "Browser resource not found in API Gateway"
        log_info "Available resources:"
        cat resources.json | jq -r '.items[] | "\(.path) - \(.id)"'
        rm -f resources.json
        exit 1
    fi
    
    log_success "Browser resource found: $BROWSER_RESOURCE_ID"
    
    # Check if GET method exists
    GET_METHOD=$(aws apigateway get-method --rest-api-id $API_ID --resource-id $BROWSER_RESOURCE_ID --http-method GET 2>/dev/null || echo "not_found")
    
    if [ "$GET_METHOD" = "not_found" ]; then
        log_error "GET method not found for browser resource"
        exit 1
    fi
    
    log_success "GET method exists for browser resource"
    
    rm -f resources.json
}

# Check Lambda permissions
check_lambda_permissions() {
    log_info "Checking Lambda permissions..."
    
    # Check if API Gateway has permission to invoke Lambda
    POLICY=$(aws lambda get-policy --function-name $LAMBDA_FUNCTION_NAME 2>/dev/null || echo "no_policy")
    
    if [ "$POLICY" = "no_policy" ]; then
        log_warning "No resource policy found for Lambda function"
        add_lambda_permission
    else
        # Check if API Gateway permission exists
        if echo "$POLICY" | grep -q "apigateway.amazonaws.com"; then
            log_success "API Gateway permission exists"
        else
            log_warning "API Gateway permission not found"
            add_lambda_permission
        fi
    fi
}

# Add Lambda permission for API Gateway
add_lambda_permission() {
    log_info "Adding API Gateway permission to Lambda function..."
    
    STATEMENT_ID="api-gateway-invoke-$(date +%s)"
    
    aws lambda add-permission \
        --function-name $LAMBDA_FUNCTION_NAME \
        --statement-id $STATEMENT_ID \
        --action lambda:InvokeFunction \
        --principal apigateway.amazonaws.com \
        --source-arn "arn:aws:execute-api:$AWS_REGION:*:$API_ID/*/*" \
        --region $AWS_REGION
    
    log_success "Permission added successfully"
}

# Test API Gateway integration
test_integration() {
    log_info "Testing API Gateway integration..."
    
    # Test the endpoint directly
    API_URL="https://$API_ID.execute-api.$AWS_REGION.amazonaws.com/v1/api/images/browser?bucketType=reference"
    
    log_info "Testing URL: $API_URL"
    
    RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" "$API_URL")
    HTTP_STATUS=$(echo $RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    RESPONSE_BODY=$(echo $RESPONSE | sed -e 's/HTTPSTATUS:.*//g')
    
    log_info "HTTP Status: $HTTP_STATUS"
    
    if [ "$HTTP_STATUS" = "200" ]; then
        log_success "API Gateway integration is working!"
        log_info "Response: $RESPONSE_BODY"
    elif [ "$HTTP_STATUS" = "403" ]; then
        log_error "403 Forbidden - Check API key or authentication"
    elif [ "$HTTP_STATUS" = "405" ]; then
        log_error "405 Method Not Allowed - Method configuration issue"
        log_info "This suggests the GET method is not properly configured"
    elif [ "$HTTP_STATUS" = "500" ]; then
        log_error "500 Internal Server Error - Lambda function error"
        log_info "Response: $RESPONSE_BODY"
    else
        log_error "Unexpected status code: $HTTP_STATUS"
        log_info "Response: $RESPONSE_BODY"
    fi
}

# Create a new deployment
create_deployment() {
    log_info "Creating new API Gateway deployment..."
    
    DEPLOYMENT_ID=$(aws apigateway create-deployment \
        --rest-api-id $API_ID \
        --stage-name v1 \
        --description "Fix for API Images Browser - $(date)" \
        --query 'id' --output text)
    
    if [ $? -eq 0 ]; then
        log_success "New deployment created: $DEPLOYMENT_ID"
    else
        log_error "Failed to create deployment"
        exit 1
    fi
}

# Main fix function
fix_api_gateway() {
    log_info "Starting API Gateway fix for Images Browser..."
    echo "=================================================="
    
    get_lambda_arn
    echo "=================================================="
    
    check_resources
    echo "=================================================="
    
    check_lambda_permissions
    echo "=================================================="
    
    log_info "Creating new deployment to ensure changes are active..."
    create_deployment
    echo "=================================================="
    
    log_info "Waiting 10 seconds for deployment to propagate..."
    sleep 10
    
    test_integration
    echo "=================================================="
    
    log_info "Fix attempt completed!"
}

# Parse command line arguments
case "${1:-fix}" in
    "check")
        get_lambda_arn
        check_resources
        check_lambda_permissions
        ;;
    "test")
        test_integration
        ;;
    "deploy")
        create_deployment
        ;;
    "permission")
        get_lambda_arn
        add_lambda_permission
        ;;
    "fix"|"")
        fix_api_gateway
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  fix        Run full fix process [default]"
        echo "  check      Check current configuration"
        echo "  test       Test API endpoint"
        echo "  deploy     Create new deployment"
        echo "  permission Add Lambda permission"
        echo "  help       Show this help message"
        ;;
    *)
        log_error "Unknown command: $1"
        log_info "Use '$0 help' for usage information"
        exit 1
        ;;
esac
