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
- **outputs.tf**: Contains output definitions
- **CHANGELOG.md**: Documents changes to the module

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

- `GET /api/verifications/lookup`: Lookup historical verifications
- `POST /api/verifications`: Initiate a new verification (directly integrated with Step Functions)
- `GET /api/verifications`: List all verifications
- `GET /api/verifications/{verificationId}`: Get a specific verification
- `GET /api/verifications/{verificationId}/conversation`: Get verification conversation
- `GET /api/health`: Health check endpoint
- `GET /api/images/{key}/view`: View a specific image
- `GET /api/images/browser`: Browse available images
- `POST /api/workflow`: Direct access to the Step Functions workflow (if enabled)

## Step Functions Integration

The API Gateway now integrates directly with AWS Step Functions for the `POST /api/verifications` endpoint. This integration:

1. Eliminates the need for an intermediary Lambda function
2. Directly starts a Step Functions execution when the endpoint is called
3. Passes the request payload as input to the state machine
4. Uses IAM roles to authorize the API Gateway to invoke Step Functions

### Configuration

To enable Step Functions integration, provide the following variables:

```hcl
module "api_gateway" {
  source = "./modules/api_gateway"
  
  # Other variables...
  
  step_functions_role_arn = module.step_functions.api_gateway_role_arn
  step_functions_state_machine_arn = module.step_functions.state_machine_arn
}
```

The integration uses a request template to format the input for the Step Functions StartExecution API:

```json
{
  "input": "$util.escapeJavaScript($input.json('$'))",
  "stateMachineArn": "${var.step_functions_state_machine_arn}"
}
```

## API Base Path

The API uses `/api/` as its base path. This simplifies the final invoke URL structure, which follows the pattern:
```
https://{api-id}.execute-api.{region}.amazonaws.com/{stage}/api/...
```

For example, with a stage name of `v1`, the health endpoint would be accessible at:
```
https://abc123.execute-api.ap-southeast-1.amazonaws.com/v1/api/health
```

## CORS Configuration

CORS (Cross-Origin Resource Sharing) is enabled for all endpoints when the `cors_enabled` variable is set to `true`. This allows browsers to make cross-origin requests to the API.

### CORS Implementation

The module implements CORS through:

1. **OPTIONS Method**: Each endpoint has an OPTIONS method that responds to preflight requests
2. **Integration Responses**: Each GET/POST method has an integration response that adds CORS headers
3. **Method Responses**: Each method response includes the necessary CORS headers
4. **Gateway Responses**: Global gateway responses include CORS headers for error cases

### CORS Headers

The following CORS headers are included in responses:

- `Access-Control-Allow-Origin`: Set to `*` by default (configurable via `cors_origin` local variable)
- `Access-Control-Allow-Methods`: Configured per endpoint (GET, POST, OPTIONS as appropriate)
- `Access-Control-Allow-Headers`: Common headers including `Content-Type`, `Authorization`, etc.

### Enabling CORS

To enable CORS for the API Gateway, set the `cors_enabled` variable to `true`:

```hcl
module "api_gateway" {
  source = "./modules/api_gateway"
  
  cors_enabled = true
  # Other variables...
}
```
