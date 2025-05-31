# Auto-Deploy New API Scripts

This directory contains automation scripts to streamline the process of adding new API endpoints and deploying infrastructure changes.

## Scripts Overview

### 1. `auto-deploy-new-api.sh`
Comprehensive deployment automation script that handles:
- Terraform planning and applying
- ECR image updates for Lambda functions
- Deployment verification
- Backup and rollback capabilities
- Comprehensive logging

### 2. `add-new-api.sh`
Helper script to add new API endpoints to the infrastructure by:
- Adding Lambda function configurations to `locals.tf`
- Creating ECR repository configurations
- Generating API Gateway resources and methods
- Updating Step Functions integration

### 3. `update-lambda-image-uris.sh`
Existing script for updating Lambda function image URIs with matching ECR repositories.

## Quick Start

### Adding a New API Endpoint

1. **Add the API configuration:**
   ```bash
   ./add-new-api.sh --name user_management \
                    --description "User management API" \
                    --path "/api/users" \
                    --method GET \
                    --memory 512 \
                    --timeout 30
   ```

2. **Deploy the changes:**
   ```bash
   ./auto-deploy-new-api.sh --force --auto-approve
   ```

### Dry Run (Preview Changes)

```bash
# Preview what would be deployed
./auto-deploy-new-api.sh

# Preview only ECR updates
./auto-deploy-new-api.sh --skip-terraform
```

## Detailed Usage

### auto-deploy-new-api.sh Options

```bash
./auto-deploy-new-api.sh [options]

Options:
  -p, --profile           AWS profile to use (default: default)
  -r, --region            AWS region (default: from AWS configuration)
  -t, --tag               Image tag to use (default: latest)
  -f, --force             Apply changes without dry run
  --auto-approve          Auto-approve Terraform changes
  --skip-terraform        Skip Terraform plan/apply
  --skip-ecr-update       Skip ECR image updates
  --skip-verification     Skip deployment verification
  --no-backup             Disable backup creation
  --rollback-on-failure   Enable automatic rollback on failure
  -h, --help              Show help message
```

### add-new-api.sh Options

```bash
./add-new-api.sh [options]

Options:
  -n, --name              Name of the new API function (required)
  -d, --description       Description of the function
  -p, --path              API path (e.g., /api/new-endpoint)
  -m, --method            HTTP method (GET, POST, PUT, DELETE)
  --memory                Memory size in MB (default: 256)
  --timeout               Timeout in seconds (default: 30)
  --auto-deploy           Automatically run deployment after adding
  -h, --help              Show help message
```

## Common Workflows

### 1. Complete New API Development Workflow

```bash
# Step 1: Add new API configuration
./add-new-api.sh --name data_export \
                 --description "Data export functionality" \
                 --path "/api/export" \
                 --method POST \
                 --memory 1024 \
                 --timeout 120

# Step 2: Create Lambda function code (manual step)
# - Create the Lambda function implementation
# - Build Docker image
# - Push to ECR repository

# Step 3: Deploy infrastructure
./auto-deploy-new-api.sh --force --auto-approve
```

### 2. Update Existing Infrastructure

```bash
# Update only ECR images (after code changes)
./auto-deploy-new-api.sh --skip-terraform --force

# Full deployment with verification
./auto-deploy-new-api.sh --force --auto-approve

# Safe deployment with rollback capability
./auto-deploy-new-api.sh --force --rollback-on-failure
```

### 3. Development and Testing

```bash
# Preview changes without applying
./auto-deploy-new-api.sh

# Deploy to development environment
./auto-deploy-new-api.sh --profile dev --region us-east-1 --force

# Deploy specific image tag
./auto-deploy-new-api.sh --tag v1.2.3 --force
```

## Features

### Backup and Recovery
- Automatic backup of Terraform state and Lambda configurations
- Rollback capability on deployment failure
- Backup retention (30 days for logs, 7 days for plans)

### Logging and Monitoring
- Comprehensive logging to `logs/auto_deploy_TIMESTAMP.log`
- Color-coded console output
- Deployment verification with health checks

### Safety Features
- Dry run mode by default
- Prerequisites checking (tools, credentials, Terraform state)
- Terraform validation before deployment
- Deployment verification after changes

### Flexibility
- Modular execution (skip specific steps)
- Multiple AWS profile support
- Configurable image tags
- Auto-approval for CI/CD integration

## File Structure

```
iac/
├── auto-deploy-new-api.sh      # Main deployment script
├── add-new-api.sh              # API addition helper
├── update-lambda-image-uris.sh # ECR image updater
├── logs/                       # Deployment logs
├── backups/                    # State backups
├── locals.tf                   # Lambda and ECR configurations
├── main.tf                     # Main Terraform configuration
└── modules/
    └── api_gateway/
        ├── resources.tf        # API Gateway resources
        └── methods.tf          # API Gateway methods
```

## Prerequisites

### Required Tools
- `terraform` (>= 1.0)
- `aws` CLI (>= 2.0)
- `jq` (for JSON processing)
- `curl` (for health checks)

### AWS Configuration
- Valid AWS credentials configured
- Appropriate IAM permissions for:
  - Lambda functions
  - API Gateway
  - ECR repositories
  - Step Functions
  - CloudWatch
  - S3 buckets
  - DynamoDB tables

### Terraform Setup
- Terraform initialized (`terraform init`)
- Valid `terraform.tfvars` file
- Backend configuration (if using remote state)

## Troubleshooting

### Common Issues

1. **Terraform not initialized**
   ```bash
   terraform init
   ```

2. **AWS credentials not configured**
   ```bash
   aws configure --profile your-profile
   ```

3. **Missing required tools**
   ```bash
   # macOS
   brew install terraform awscli jq

   # Ubuntu/Debian
   apt-get install terraform awscli jq
   ```

4. **Permission denied on scripts**
   ```bash
   chmod +x *.sh
   ```

### Log Analysis
Check deployment logs for detailed error information:
```bash
tail -f logs/auto_deploy_TIMESTAMP.log
```

### Manual Rollback
If automatic rollback fails:
```bash
# Restore from backup
cp backups/TIMESTAMP/terraform.tfstate.backup terraform.tfstate
terraform apply
```

## Best Practices

1. **Always test in development first**
2. **Use dry run mode to preview changes**
3. **Enable backups for production deployments**
4. **Monitor logs during deployment**
5. **Verify deployment after completion**
6. **Use specific image tags for production**
7. **Keep deployment scripts in version control**

## Integration with CI/CD

Example GitHub Actions workflow:
```yaml
- name: Deploy Infrastructure
  run: |
    ./auto-deploy-new-api.sh \
      --profile production \
      --tag ${{ github.sha }} \
      --force \
      --auto-approve \
      --rollback-on-failure
```

## Support

For issues or questions:
1. Check the logs in `logs/` directory
2. Review Terraform plan output
3. Verify AWS permissions and credentials
4. Check prerequisite tools installation
