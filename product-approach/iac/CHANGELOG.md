# Changelog

## [1.7.0] - 2024-12-20

### Added
- Enhanced ECS Streamlit IAM policies to support dual AWS Secrets Manager access
- Added support for both CONFIG_SECRET and API_KEY_SECRET_NAME in IAM permissions
- Implemented comprehensive Secrets Manager resource ARN management for ECS tasks

### Changed
- **BREAKING**: Updated ECS Streamlit IAM policy to require both CONFIG_SECRET and API_KEY_SECRET_NAME environment variables
- Enhanced IAM policy logic to dynamically build secret ARNs based on configured environment variables
- Improved IAM policy conditions to create Secrets Manager access only when secrets are configured
- Updated IAM policy resource targeting to include multiple specific secret ARNs instead of wildcard access

### Security
- **MAJOR SECURITY IMPROVEMENT**: Implemented least-privilege access for Secrets Manager
- Restricted IAM permissions to only the specific secrets required by the application
- Enhanced security by removing wildcard Secrets Manager access in favor of targeted resource permissions
- Improved IAM policy structure for better security auditing and compliance

### Infrastructure
- Updated ECS task role IAM policies to support the new secrets-based configuration approach
- Enhanced Terraform locals to dynamically calculate required secret ARNs
- Improved IAM policy attachment logic based on configured environment variables
- Added backward compatibility for deployments without secrets configuration

### Migration Required
- **ACTION REQUIRED**: Apply Terraform changes to update ECS task role IAM permissions
- **ACTION REQUIRED**: Ensure both CONFIG_SECRET and API_KEY_SECRET_NAME are properly configured in ECS Task Definition
- **ACTION REQUIRED**: Verify that both secrets exist in AWS Secrets Manager before deployment
- **ACTION REQUIRED**: Restart ECS service after applying IAM policy changes

### Deployment Notes
- ECS service will automatically restart when IAM policies are updated
- Application will gain access to both configuration and API key secrets
- No downtime expected during IAM policy updates
- Monitor CloudWatch logs to verify successful secret access after deployment

## [1.6.0] - 2024-12-19

### Added
- Added new ECS Streamlit configuration secret management using AWS Secrets Manager
- Created `ecs_config_secret` module for centralized configuration storage
- Implemented CONFIG_SECRET environment variable approach for enhanced security

### Changed
- Replaced hardcoded environment variables (REGION, DYNAMODB_TABLE, S3_BUCKET, AWS_DEFAULT_REGION, API_ENDPOINT) with CONFIG_SECRET
- Updated ECS Streamlit module to use single CONFIG_SECRET environment variable
- Enhanced Secrets Manager module to support both simple strings and JSON configuration objects
- Updated IAM policies in ECS Streamlit module to support configuration secret access

### Security
- Improved security posture by moving sensitive configuration from environment variables to encrypted AWS Secrets Manager
- Centralized configuration management for better security and maintainability
- Reduced environment variable exposure in ECS task definitions

## [1.5.0] - 2025-05-19

### Fixed
- Fixed IAM policy error in Step Functions module: "MalformedPolicyDocument: Policy statement must contain resources"
- Removed S3 state management policy when s3_state_bucket_arn is null
- Added state lifecycle rules to terraform.tfvars for proper S3 bucket configuration
- Removed default values from variables.tf for s3_buckets to ensure values are defined in terraform.tfvars

### Changed
- Updated main.tf to properly handle Step Functions module integration with S3 state bucket
- Improved error handling in Step Functions module for S3 state management policy

## [1.4.0] - 2025-05-18

### Added
- Added S3 state management bucket configuration
- Added lifecycle rules for state bucket data management
- Added integration between Step Functions and S3 state bucket

### Changed
- Updated S3 module to support state management bucket
- Updated Step Functions module to access S3 state bucket