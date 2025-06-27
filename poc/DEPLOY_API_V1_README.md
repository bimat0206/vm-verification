# API Gateway v1 Deployment Script

This document describes how to use the `deploy-api-v1.sh` script to deploy the API Gateway to stage v1 for the vending machine verification project.

## Overview

The `deploy-api-v1.sh` script is a comprehensive deployment tool that handles:
- Infrastructure deployment using Terraform
- Lambda function image updates
- API Gateway deployment to stage v1
- Post-deployment verification
- Backup creation and rollback capabilities

## Prerequisites

Before running the deployment script, ensure you have:

1. **Required Tools:**
   - `terraform` (>= 1.0)
   - `aws` CLI (>= 2.0)
   - `jq` (for JSON processing)
   - `curl` (for API testing)

2. **AWS Configuration:**
   - AWS credentials configured (`aws configure` or environment variables)
   - Appropriate IAM permissions for API Gateway, Lambda, and other AWS services

3. **Project Setup:**
   - Run from the project root directory
   - Terraform initialized in the `iac/` directory

## Usage

### Basic Usage

```bash
# Deploy with default settings (dry run first)
./deploy-api-v1.sh --dry-run

# Deploy to stage v1
./deploy-api-v1.sh --auto-approve
```

### Command Line Options

```bash
./deploy-api-v1.sh [options]

Options:
  -p, --profile           AWS profile to use (default: default)
  -r, --region            AWS region (default: us-east-1)
  -t, --tag               Docker image tag to use (default: latest)
  --dry-run               Preview changes without applying them
  --auto-approve          Auto-approve all changes without prompts
  --skip-backup           Skip creating backup before deployment
  --skip-verification     Skip post-deployment verification
  --force-redeploy        Force redeployment even if no changes detected
  -h, --help              Show help message
```

### Examples

```bash
# Preview deployment changes
./deploy-api-v1.sh --dry-run

# Deploy with automatic approval
./deploy-api-v1.sh --auto-approve

# Deploy with specific AWS profile and region
./deploy-api-v1.sh --profile prod --region us-west-2 --auto-approve

# Deploy with specific Docker image tag
./deploy-api-v1.sh --tag v1.2.3 --auto-approve

# Deploy without backup (faster, but less safe)
./deploy-api-v1.sh --skip-backup --auto-approve

# Deploy without post-deployment verification
./deploy-api-v1.sh --skip-verification --auto-approve
```

## Using with Makefile

You can also use the deployment script through the Makefile in the `iac/` directory:

```bash
cd iac/

# Dry run deployment
make deploy

# Deploy with force
make deploy-force

# Deploy to production
make deploy-prod

# Update only ECR images
make ecr-update

# Deploy with specific AWS profile
make deploy-force AWS_PROFILE=prod

# Deploy with specific image tag
make deploy-force IMAGE_TAG=v1.2.3
```

## Deployment Process

The script follows this process:

1. **Prerequisites Check:** Validates required tools and AWS credentials
2. **Backup Creation:** Creates backup of current state (unless skipped)
3. **Terraform Validation:** Validates Terraform configuration
4. **Terraform Plan:** Creates execution plan for infrastructure changes
5. **Terraform Apply:** Applies infrastructure changes (if not dry run)
6. **Lambda Updates:** Updates Lambda function images with specified tag
7. **Verification:** Tests API endpoints and verifies deployment
8. **Summary:** Provides deployment summary and endpoint URLs

## Configuration

The script uses configuration from:
- `iac/terraform.tfvars` - Main Terraform variables
- `iac/deploy-config.env` - Deployment configuration (if exists)
- Command line arguments (highest priority)

Key configuration values:
- **Stage Name:** `v1` (hardcoded for this script)
- **AWS Region:** `us-east-1` (default, can be overridden)
- **Image Tag:** `latest` (default, can be overridden)

## Output and Logging

The script provides:
- **Console Output:** Color-coded status messages
- **Log Files:** Detailed logs in `iac/logs/deploy_api_v1_TIMESTAMP.log`
- **Backup Files:** State backups in `iac/backups/TIMESTAMP/`

## Verification

After deployment, the script verifies:
- ✅ API Gateway stage v1 exists and is accessible
- ✅ Lambda functions are in "Active" state
- ✅ API endpoints respond correctly
- ✅ Health check endpoint (if available)

## Troubleshooting

### Common Issues

1. **AWS Credentials Error:**
   ```bash
   # Check AWS configuration
   aws sts get-caller-identity
   
   # Configure if needed
   aws configure
   ```

2. **Terraform Not Initialized:**
   ```bash
   cd iac/
   terraform init
   ```

3. **Permission Denied:**
   ```bash
   chmod +x deploy-api-v1.sh
   ```

4. **API Gateway Not Found:**
   - Ensure Terraform has been applied at least once
   - Check AWS region is correct

### Debug Mode

For detailed debugging, check the log file:
```bash
tail -f iac/logs/deploy_api_v1_TIMESTAMP.log
```

## Safety Features

- **Dry Run Mode:** Preview changes before applying
- **Backup Creation:** Automatic state backup before changes
- **Verification:** Post-deployment health checks
- **Error Handling:** Graceful error handling and cleanup
- **Rollback:** Manual rollback using backup files (if needed)

## API Endpoints

After successful deployment, the API will be available at:
```
https://API_ID.execute-api.REGION.amazonaws.com/v1
```

Key endpoints:
- `GET /api/health` - Health check
- `GET /api/verifications` - List verifications
- `POST /api/verifications` - Create verification
- `GET /api/verifications/{id}` - Get specific verification

## Support

For issues or questions:
1. Check the log files in `iac/logs/`
2. Review the Terraform plan output
3. Verify AWS permissions and configuration
4. Check the project documentation in `iac/README.md`