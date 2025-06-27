# AWS Secrets Manager Configuration

This document describes the new configuration approach for the Streamlit application using AWS Secrets Manager instead of hardcoded environment variables.

## Overview

The application now uses two AWS Secrets Manager secrets for configuration:

1. **CONFIG_SECRET**: Contains application configuration parameters
2. **API_KEY_SECRET_NAME**: Contains the API key for authentication

## ECS Task Definition Environment Variables

The ECS Task Definition should include these environment variables:

```json
{
  "environment": [
    {
      "name": "CONFIG_SECRET",
      "value": "your-config-secret-name"
    },
    {
      "name": "API_KEY_SECRET_NAME", 
      "value": "your-api-key-secret-name"
    },
    {
      "name": "AWS_DEFAULT_REGION",
      "value": "us-east-1"
    }
  ]
}
```

## AWS Secrets Manager Secret Structure

### CONFIG_SECRET Secret

This secret should contain a JSON object with the following structure:

```json
{
  "API_ENDPOINT": "https://your-api-endpoint.com/v1",
  "CHECKING_BUCKET": "your-checking-bucket-name",
  "DYNAMODB_CONVERSATION_TABLE": "your-conversation-table-name",
  "DYNAMODB_VERIFICATION_TABLE": "your-verification-table-name", 
  "REFERENCE_BUCKET": "your-reference-bucket-name",
  "REGION": "us-east-1"
}
```

### API_KEY_SECRET_NAME Secret

This secret should contain a JSON object with the following structure:

```json
{
  "api_key": "your-actual-api-key-value"
}
```

## Benefits

1. **Security**: Sensitive configuration is stored securely in AWS Secrets Manager
2. **Centralized Management**: Configuration can be updated without rebuilding Docker images
3. **Environment Separation**: Different environments can use different secrets
4. **Audit Trail**: AWS CloudTrail logs access to secrets
5. **Rotation**: API keys can be rotated using AWS Secrets Manager rotation features

## Backward Compatibility

The application maintains backward compatibility with individual environment variables. If `CONFIG_SECRET` is not provided, the application will fall back to reading individual environment variables like `API_ENDPOINT`, `REGION`, etc.

## IAM Permissions

The ECS task role must have the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:region:account:secret:your-config-secret-name*",
        "arn:aws:secretsmanager:region:account:secret:your-api-key-secret-name*"
      ]
    }
  ]
}
```

## Migration Steps

1. Create the CONFIG_SECRET in AWS Secrets Manager with the required configuration
2. Create the API_KEY_SECRET_NAME in AWS Secrets Manager with the API key
3. Update the ECS Task Definition to include CONFIG_SECRET and API_KEY_SECRET_NAME environment variables
4. Remove individual configuration environment variables from the ECS Task Definition
5. Deploy the updated application

## Troubleshooting

- Check CloudWatch logs for configuration loading messages
- Verify IAM permissions for the ECS task role
- Ensure secrets exist in the correct AWS region
- Validate JSON structure in the secrets
