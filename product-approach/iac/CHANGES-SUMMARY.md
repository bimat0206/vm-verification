# API Gateway and Step Functions Integration Changes

## Overview
The infrastructure has been updated to improve the integration between API Gateway and Step Functions. Previously, the API Gateway POST /api/verifications endpoint was integrated with a Lambda function (initialize), which then triggered the Step Functions workflow. Now, API Gateway integrates directly with Step Functions, providing a more streamlined architecture and consistent input handling.

## Changes Made

### 1. API Gateway Integration Update
- Changed the POST /api/verifications endpoint to integrate directly with Step Functions
- Updated the integration type to AWS with StartExecution action
- Added request template to format the input for Step Functions
- Configured proper IAM permissions for API Gateway to invoke Step Functions

### 2. Step Functions Initialize State Update
- Added Parameters mapping to the Initialize state in the Step Functions template
- Ensured consistent input structure regardless of invocation source
- Mapped all required parameters: verificationType, referenceImageUrl, checkingImageUrl, vendingMachineId, layoutId, layoutPrefix, notificationEnabled, requestId, and requestTimestamp

### 3. Infrastructure Support Changes
- Added necessary outputs for the Step Functions module
- Added required variables for API Gateway and Step Functions modules
- Updated main.tf to pass the correct parameters between modules

### 4. File Changes
The following files were modified:
- `step_functions/templates/state_machine_definition.tftpl`: Added Parameters mapping to Initialize state
- `api_gateway/methods.tf`: Updated integration for POST /api/verifications
- `api_gateway/variables.tf`: Added Step Functions related variables
- `step_functions/variables.tf`: Added API Gateway endpoint variable
- `step_functions/output.tf`: Renamed to `outputs.tf` to match naming convention in main.tf
- `api_gateway/output.tf`: Renamed to `outputs.tf` to match naming convention in main.tf
- `step_functions/outputs.tf`: Added outputs for API Gateway integration
- `main.tf`: Updated module calls to pass the correct parameters between modules

## Benefits
- Simplified architecture by removing an unnecessary Lambda invocation
- Consistent input structure for the Step Functions workflow
- More direct control over the Step Functions execution
- Improved error handling and monitoring
- Better alignment with AWS best practices for serverless architectures

## Testing Recommendations
After deploying these changes, verify that:
1. The POST /api/verifications endpoint successfully starts a Step Functions execution
2. The Initialize Lambda function receives the correct parameters
3. The workflow completes successfully with the new integration
4. Error handling works as expected
