# API Gateway Module

## Overview
This module creates and configures an AWS API Gateway REST API with multiple endpoints, Lambda integrations, and CORS support.

## Module Structure
The module has been organized into smaller, more manageable files to improve maintainability and make troubleshooting easier:

- **main.tf**: Entry point that references other files
- **resources.tf**: Contains all API Gateway resource definitions
- **methods.tf**: Contains all API Gateway method and integration definitions
- **deployment.tf**: Contains deployment, stage, and API key configurations
- **cors.tf**: Contains CORS configuration for API endpoints
- **locals.tf**: Contains local variable definitions
- **variables.tf**: Contains input variable definitions
- **output.tf**: Contains output definitions

## Benefits of This Structure

1. **Improved Readability**: Each file has a clear purpose and contains related resources
2. **Easier Troubleshooting**: When issues arise, you can quickly locate the relevant file
3. **Better Collaboration**: Team members can work on different aspects of the module simultaneously
4. **Simplified Maintenance**: Smaller files are easier to understand and modify
5. **Reduced Merge Conflicts**: Separating resources into logical files reduces the chance of conflicts

## Usage

```hcl
module "api_gateway" {
  source = "./modules/api_gateway"
  
  api_name        = "verification-api"
  api_description = "API for vending machine verification service"
  stage_name      = "dev"
  
  # Additional variables as needed
}
```

## Endpoints

The API Gateway exposes the following endpoints:

- `GET /api/v1/verifications/lookup`: Lookup historical verifications
- `POST /api/v1/verifications`: Initiate a new verification
- `GET /api/v1/verifications`: List all verifications
- `GET /api/v1/verifications/{verificationId}`: Get a specific verification
- `GET /api/v1/verifications/{verificationId}/conversation`: Get verification conversation
- `GET /api/v1/health`: Health check endpoint
- `GET /api/v1/images/{key}/view`: View a specific image
- `GET /api/v1/images/browser`: Browse available images