#!/bin/bash

# Setup script for local development of the Streamlit application
# This script automatically retrieves cloud resource information and configures the app

set -e

echo "🚀 Setting up Streamlit app for local development..."

# Check if AWS CLI is available and configured
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI not found. Please install AWS CLI first."
    exit 1
fi

if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured. Run 'aws configure' first."
    exit 1
fi

# Get AWS account info
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
REGION=$(aws configure get region || echo "us-east-1")

echo "✅ AWS credentials configured"
echo "   Account: $ACCOUNT_ID"
echo "   Region: $REGION"

# Check if .streamlit directory exists
if [ ! -d ".streamlit" ]; then
    echo "📁 Creating .streamlit directory..."
    mkdir -p .streamlit
fi

# Check if secrets.toml already exists
if [ -f ".streamlit/secrets.toml" ]; then
    echo "⚠️  .streamlit/secrets.toml already exists."
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Setup cancelled. Existing secrets.toml preserved."
        exit 1
    fi
fi

echo ""
echo "🔍 Retrieving cloud resource information..."

# Function to find resources by pattern
find_api_gateway() {
    echo "🔍 Looking for API Gateway..." >&2
    local api_id=$(aws apigateway get-rest-apis --query 'items[?contains(name, `vending`) || contains(name, `verification`) || contains(name, `kootoro`)].id' --output text 2>/dev/null | tr '\t' '\n' | head -1)
    if [ -n "$api_id" ] && [ "$api_id" != "None" ] && [ "$api_id" != "" ]; then
        echo "https://${api_id}.execute-api.${REGION}.amazonaws.com/v1"
    else
        echo ""
    fi
}

find_api_key() {
    echo "🔍 Looking for API Key..." >&2
    local api_key=$(aws apigateway get-api-keys --query 'items[?contains(name, `vending`) || contains(name, `verification`) || contains(name, `kootoro`)].value' --output text 2>/dev/null | tr '\t' '\n' | head -1)
    if [ -n "$api_key" ] && [ "$api_key" != "None" ] && [ "$api_key" != "" ]; then
        echo "$api_key"
    else
        echo ""
    fi
}

find_s3_buckets() {
    echo "🔍 Looking for S3 buckets..." >&2
    local reference_bucket=$(aws s3api list-buckets --query 'Buckets[?contains(Name, `reference`) && (contains(Name, `vending`) || contains(Name, `verification`) || contains(Name, `kootoro`))].Name' --output text 2>/dev/null | tr '\t' '\n' | head -1)
    local checking_bucket=$(aws s3api list-buckets --query 'Buckets[?contains(Name, `checking`) && (contains(Name, `vending`) || contains(Name, `verification`) || contains(Name, `kootoro`))].Name' --output text 2>/dev/null | tr '\t' '\n' | head -1)

    echo "REFERENCE:$reference_bucket"
    echo "CHECKING:$checking_bucket"
}

find_dynamodb_tables() {
    echo "🔍 Looking for DynamoDB tables..." >&2
    local verification_table=$(aws dynamodb list-tables --query 'TableNames[?contains(@, `verification`) && (contains(@, `vending`) || contains(@, `kootoro`))]' --output text 2>/dev/null | tr '\t' '\n' | head -1)
    local conversation_table=$(aws dynamodb list-tables --query 'TableNames[?contains(@, `conversation`) && (contains(@, `vending`) || contains(@, `kootoro`))]' --output text 2>/dev/null | tr '\t' '\n' | head -1)

    echo "VERIFICATION:$verification_table"
    echo "CONVERSATION:$conversation_table"
}

# Retrieve cloud resources
API_ENDPOINT=$(find_api_gateway)
API_KEY=$(find_api_key)
S3_INFO=$(find_s3_buckets)
DYNAMODB_INFO=$(find_dynamodb_tables)

# Parse S3 info
REFERENCE_BUCKET=$(echo "$S3_INFO" | grep "REFERENCE:" | cut -d: -f2)
CHECKING_BUCKET=$(echo "$S3_INFO" | grep "CHECKING:" | cut -d: -f2)

# Parse DynamoDB info
VERIFICATION_TABLE=$(echo "$DYNAMODB_INFO" | grep "VERIFICATION:" | cut -d: -f2)
CONVERSATION_TABLE=$(echo "$DYNAMODB_INFO" | grep "CONVERSATION:" | cut -d: -f2)

# Create secrets.toml file
echo "📝 Creating .streamlit/secrets.toml with discovered resources..."

cat > .streamlit/secrets.toml << EOF
# Auto-generated Streamlit secrets for local development
# Generated on $(date)

# API Configuration
API_ENDPOINT = "${API_ENDPOINT:-https://your-api-gateway-endpoint.execute-api.${REGION}.amazonaws.com/v1}"
API_KEY = "${API_KEY:-your-api-key-here}"

# AWS Configuration
REGION = "${REGION}"
AWS_DEFAULT_REGION = "${REGION}"

# S3 Buckets
REFERENCE_BUCKET = "${REFERENCE_BUCKET:-your-reference-bucket-name}"
CHECKING_BUCKET = "${CHECKING_BUCKET:-your-checking-bucket-name}"

# DynamoDB Tables
DYNAMODB_VERIFICATION_TABLE = "${VERIFICATION_TABLE:-your-verification-table-name}"
DYNAMODB_CONVERSATION_TABLE = "${CONVERSATION_TABLE:-your-conversation-table-name}"

# Legacy support (if needed)
DYNAMODB_TABLE = "${VERIFICATION_TABLE:-your-legacy-table-name}"
S3_BUCKET = "${REFERENCE_BUCKET:-your-legacy-bucket-name}"
EOF

echo "✅ Created .streamlit/secrets.toml"

# Show what was found
echo ""
echo "📋 Discovered resources:"
echo "   API Endpoint: ${API_ENDPOINT:-❌ Not found}"
echo "   API Key: ${API_KEY:+✅ Found}${API_KEY:-❌ Not found}"
echo "   Reference Bucket: ${REFERENCE_BUCKET:-❌ Not found}"
echo "   Checking Bucket: ${CHECKING_BUCKET:-❌ Not found}"
echo "   Verification Table: ${VERIFICATION_TABLE:-❌ Not found}"
echo "   Conversation Table: ${CONVERSATION_TABLE:-❌ Not found}"

# Check for missing resources
MISSING_RESOURCES=()
[ -z "$API_ENDPOINT" ] && MISSING_RESOURCES+=("API Gateway")
[ -z "$API_KEY" ] && MISSING_RESOURCES+=("API Key")
[ -z "$REFERENCE_BUCKET" ] && MISSING_RESOURCES+=("Reference S3 Bucket")
[ -z "$CHECKING_BUCKET" ] && MISSING_RESOURCES+=("Checking S3 Bucket")
[ -z "$VERIFICATION_TABLE" ] && MISSING_RESOURCES+=("Verification DynamoDB Table")
[ -z "$CONVERSATION_TABLE" ] && MISSING_RESOURCES+=("Conversation DynamoDB Table")

if [ ${#MISSING_RESOURCES[@]} -gt 0 ]; then
    echo ""
    echo "⚠️  Some resources were not found automatically:"
    for resource in "${MISSING_RESOURCES[@]}"; do
        echo "   - $resource"
    done
    echo ""
    echo "📝 Please edit .streamlit/secrets.toml manually to add missing values."
    echo "💡 You can find resources using:"
    echo "   - API Gateway: aws apigateway get-rest-apis"
    echo "   - API Keys: aws apigateway get-api-keys"
    echo "   - S3 Buckets: aws s3 ls"
    echo "   - DynamoDB Tables: aws dynamodb list-tables"
fi

echo ""
echo "🏃 To run the app: streamlit run app.py"
echo "📖 For more details, see LOCAL_DEVELOPMENT.md"
echo ""
echo "⚠️  Remember: Never commit .streamlit/secrets.toml to version control!"
echo ""
echo "🎉 Local development setup complete!"
