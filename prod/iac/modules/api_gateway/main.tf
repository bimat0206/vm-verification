# modules/api_gateway/main.tf

# This file has been reorganized into smaller, more manageable files:
# - resources.tf: Contains all API Gateway resource definitions
# - methods.tf: Contains all API Gateway method and integration definitions
# - deployment.tf: Contains deployment, stage, and API key configurations
# - models.tf: Contains model definitions for request/response validation
# - validators.tf: Contains request validators
# - error_responses.tf: Contains custom error responses
# - locals.tf: Contains local variable definitions
# - variables.tf: Contains input variable definitions
# - output.tf: Contains output definitions

# This reorganization improves maintainability and makes troubleshooting easier
# by grouping related resources together in logical files.