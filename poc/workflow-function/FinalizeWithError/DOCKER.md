# Docker Containerization Guide - FinalizeWithError

This document describes the Docker containerization setup for the FinalizeWithError Lambda function.

## Overview

The FinalizeWithError function is containerized using a multi-stage Docker build optimized for AWS Lambda ARM64 (Graviton) deployment. This setup provides:

- **Optimized builds** with multi-stage Docker containers
- **ARM64 support** for AWS Lambda Graviton processors
- **Automated deployment** with ECR integration
- **Shared module handling** for consistent dependency management

## Files

- `Dockerfile` - Multi-stage Docker build configuration
- `retry-docker-build.sh` - Automated build and deployment script

## Quick Start

### Prerequisites

1. **AWS CLI** configured with appropriate permissions
2. **Docker** installed and running
3. **ECR repository** created for the function
4. **Lambda function** created in AWS

### Build and Deploy

```bash
# Basic usage (update script with your ECR repo and function name first)
./retry-docker-build.sh

# With custom parameters
./retry-docker-build.sh \
  --repo=your-account.dkr.ecr.region.amazonaws.com/your-repo \
  --function=your-lambda-function-name \
  --region=us-east-1
```

## Configuration

### ECR Repository and Lambda Function

Update the default values in `retry-docker-build.sh`:

```bash
ECR_REPO="your-account.dkr.ecr.region.amazonaws.com/your-repo"
FUNCTION_NAME="your-lambda-function-name"
AWS_REGION="your-region"
```

### Command Line Options

- `--repo=<ECR_REPO_URI>` - ECR repository URI
- `--function=<LAMBDA_FUNCTION_NAME>` - Lambda function name
- `--region=<AWS_REGION>` - AWS region (default: us-east-1)
- `--help` - Show help message

## Build Process

The build script performs the following steps:

1. **Validation** - Verifies directory structure and required files
2. **ECR Login** - Authenticates with AWS ECR
3. **Build Context** - Creates temporary build context with shared modules
4. **Module Copying** - Copies shared modules (logger, schema, s3state, errors)
5. **Go Module Setup** - Creates Docker-compatible go.mod with local paths
6. **Docker Build** - Builds multi-stage container for ARM64
7. **ECR Push** - Pushes image to ECR repository
8. **Lambda Deploy** - Updates Lambda function with new image
9. **Cleanup** - Removes temporary build context

## Docker Architecture

### Multi-Stage Build

**Stage 1: Builder**
- Base: `golang:1.24-alpine`
- Installs build dependencies (ca-certificates, git, tzdata)
- Copies source code and shared modules
- Downloads Go dependencies
- Compiles for `linux/arm64` with Lambda optimizations

**Stage 2: Runtime**
- Base: `public.ecr.aws/lambda/provided:al2-arm64`
- Copies compiled binary and certificates
- Sets up Lambda entrypoint

### Shared Module Handling

The build process handles shared modules by:
1. Copying shared modules to temporary build context
2. Modifying go.mod to use local paths (`./shared/module`)
3. Ensuring all dependencies are available during Docker build

## Troubleshooting

### Common Issues

**Build fails with "module not found"**
- Ensure shared modules exist in `../shared/` directory
- Check that go.mod replace directives are correct

**ECR login fails**
- Verify AWS CLI configuration and permissions
- Check ECR repository exists and is accessible

**Lambda deployment fails**
- Verify Lambda function exists and you have update permissions
- Check function name and region are correct

### Debug Mode

Add debug output to the build script:
```bash
set -x  # Add after set -e for verbose output
```

## Development

### Local Testing

Build locally without deployment:
```bash
# Build image only
docker build -t finalize-with-error:local .

# Run locally (requires Lambda runtime emulator)
docker run -p 9000:8080 finalize-with-error:local
```

### Shared Module Updates

When shared modules are updated:
1. The build script automatically copies latest versions
2. No manual intervention required
3. Dependencies are resolved during Docker build

## Security

- Uses official AWS Lambda base images
- Includes only necessary certificates
- Builds with security-optimized flags (`-s -w`)
- No sensitive data in container layers
