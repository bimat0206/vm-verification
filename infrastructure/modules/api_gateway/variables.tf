# infrastructure/modules/api_gateway/variables.tf
variable "api_name" {
  description = "Name of the API Gateway REST API"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "stage_name" {
  description = "Name of the deployment stage"
  type        = string
  default     = "v1"
}

variable "step_functions_state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  type        = string
}

variable "step_functions_invoke_arn" {
  description = "Invoke ARN for the Step Functions state machine"
  type        = string
}

variable "get_comparison_lambda_function_name" {
  description = "Name of the Lambda function for retrieving comparison results"
  type        = string
}

variable "get_comparison_lambda_invoke_arn" {
  description = "Invoke ARN for the Lambda function for retrieving comparison results"
  type        = string
}

variable "get_images_lambda_function_name" {
  description = "Name of the Lambda function for listing images"
  type        = string
}

variable "get_images_lambda_invoke_arn" {
  description = "Invoke ARN for the Lambda function for listing images"
  type        = string
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
# Create a temporary fix for the API Gateway module
# Add this to your infrastructure/variables.tf file

variable "skip_api_gateway_integration_response" {
  description = "Whether to skip creating the API Gateway integration response resource (useful for troubleshooting)"
  type        = bool
  default     = true
}

# Then modify your modules/api_gateway/main.tf file to add a count parameter to 
# the aws_api_gateway_integration_response.post_comparisons_integration_response resource like this:

# Find the line that looks like this:
# resource "aws_api_gateway_integration_response" "post_comparisons_integration_response" {

# And replace it with:
# resource "aws_api_gateway_integration_response" "post_comparisons_integration_response" {
#   count = var.skip_api_gateway_integration_response ? 0 : 1
#   rest_api_id = ... (rest of your configuration)

# Also add this to your modules/api_gateway/variables.tf file:
variable "skip_api_gateway_integration_response" {
  description = "Whether to skip creating the API Gateway integration response resource"
  type        = bool
  default     = true
}