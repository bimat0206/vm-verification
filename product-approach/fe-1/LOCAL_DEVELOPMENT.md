# Local Development Setup

This guide explains how to run the Streamlit application locally for development while connecting to cloud resources.

## Quick Setup (Recommended)

**Automated Setup Script:**
```bash
./setup-local-dev.sh
```

This script will:
- ✅ Automatically discover your cloud resources (API Gateway, S3 buckets, DynamoDB tables)
- ✅ Generate `.streamlit/secrets.toml` with the correct values
- ✅ Validate AWS credentials and configuration
- ✅ Show you what resources were found and what might be missing

## Configuration Options

The application supports multiple configuration sources with the following priority:

1. **AWS Secrets Manager** (Cloud deployment)
2. **Environment Variables** (Local/Cloud)
3. **Streamlit Secrets** (Local development)

## Manual Local Development Setup

### Option 1: Using Streamlit Secrets (Recommended for Local Development)

1. **Copy the example secrets file:**
   ```bash
   cp .streamlit/secrets.toml.example .streamlit/secrets.toml
   ```

2. **Update `.streamlit/secrets.toml` with your cloud values:**
   ```toml
   # API Configuration
   API_ENDPOINT = "https://your-api-gateway-endpoint.execute-api.us-east-1.amazonaws.com/v1"
   API_KEY = "your-actual-api-key"

   # AWS Configuration
   REGION = "us-east-1"
   AWS_DEFAULT_REGION = "us-east-1"

   # S3 Buckets
   REFERENCE_BUCKET = "your-reference-bucket-name"
   CHECKING_BUCKET = "your-checking-bucket-name"

   # DynamoDB Tables
   DYNAMODB_VERIFICATION_TABLE = "your-verification-table-name"
   DYNAMODB_CONVERSATION_TABLE = "your-conversation-table-name"
   ```

3. **Run the application:**
   ```bash
   streamlit run app.py
   ```

### Option 2: Using Environment Variables

1. **Set environment variables:**
   ```bash
   export API_ENDPOINT="https://your-api-gateway-endpoint.execute-api.us-east-1.amazonaws.com/v1"
   export API_KEY="your-actual-api-key"
   export REGION="us-east-1"
   export REFERENCE_BUCKET="your-reference-bucket-name"
   export CHECKING_BUCKET="your-checking-bucket-name"
   export DYNAMODB_VERIFICATION_TABLE="your-verification-table-name"
   export DYNAMODB_CONVERSATION_TABLE="your-conversation-table-name"
   ```

2. **Run the application:**
   ```bash
   streamlit run app.py
   ```

### Option 3: Using .env File

1. **Create a `.env` file:**
   ```bash
   API_ENDPOINT=https://your-api-gateway-endpoint.execute-api.us-east-1.amazonaws.com/v1
   API_KEY=your-actual-api-key
   REGION=us-east-1
   REFERENCE_BUCKET=your-reference-bucket-name
   CHECKING_BUCKET=your-checking-bucket-name
   DYNAMODB_VERIFICATION_TABLE=your-verification-table-name
   DYNAMODB_CONVERSATION_TABLE=your-conversation-table-name
   ```

2. **Load environment variables and run:**
   ```bash
   source .env
   streamlit run app.py
   ```

## Getting Cloud Resource Information

To get the actual values for your configuration:

### API Gateway Endpoint
```bash
aws apigateway get-rest-apis --query 'items[?name==`your-api-name`].[id,name]' --output table
# Then construct: https://{api-id}.execute-api.{region}.amazonaws.com/v1
```

### API Key
```bash
aws apigateway get-api-keys --query 'items[?name==`your-api-key-name`].value' --output text
```

### S3 Buckets
```bash
aws s3 ls | grep your-project-name
```

### DynamoDB Tables
```bash
aws dynamodb list-tables --query 'TableNames[?contains(@, `your-project-name`)]' --output table
```

## AWS Credentials

Ensure your local AWS credentials are configured:

```bash
aws configure
# or
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-east-1
```

## Troubleshooting

### Common Issues

1. **"API_ENDPOINT not found in configuration"**
   - Ensure you've set the API_ENDPOINT in one of the configuration sources
   - Check that `.streamlit/secrets.toml` exists and has the correct values

2. **"API key not found"**
   - Ensure you've set the API_KEY in environment variables or secrets.toml
   - Verify the API key is valid and not expired

3. **AWS credential errors**
   - Run `aws sts get-caller-identity` to verify your AWS credentials
   - Ensure your AWS user/role has permissions to access the required resources

### Debug Mode

Enable debug logging by setting:
```bash
export STREAMLIT_LOGGER_LEVEL=debug
```

Or add to your configuration:
```toml
[logger]
level = "debug"
```

## Security Notes

- **Never commit `.streamlit/secrets.toml`** to version control
- The `.gitignore` already excludes `.streamlit/` directory
- Use IAM roles with minimal required permissions
- Rotate API keys regularly

## Cloud Deployment

When deploying to ECS, the application will automatically use:
- `CONFIG_SECRET` environment variable pointing to AWS Secrets Manager
- `API_KEY_SECRET_NAME` environment variable pointing to API key secret
- No local configuration files are needed in cloud deployment
