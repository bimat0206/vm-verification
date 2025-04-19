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