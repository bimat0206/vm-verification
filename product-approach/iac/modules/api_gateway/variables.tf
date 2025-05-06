# modules/api_gateway/variables.tf

variable "api_name" {
  description = "Name of the API Gateway"
  type        = string
}

variable "api_description" {
  description = "Description of the API Gateway"
  type        = string
}

variable "stage_name" {
  description = "Name of the API Gateway stage"
  type        = string
}

variable "throttling_rate_limit" {
  description = "Throttling rate limit for API Gateway"
  type        = number
}

variable "throttling_burst_limit" {
  description = "Throttling burst limit for API Gateway"
  type        = number
}

variable "cors_enabled" {
  description = "Enable CORS for API Gateway"
  type        = bool
}

variable "metrics_enabled" {
  description = "Enable metrics for API Gateway"
  type        = bool
}

variable "use_api_key" {
  description = "Enable API key authentication"
  type        = bool
}

variable "openapi_definition" {
  description = "Path to OpenAPI definition file"
  type        = string
}

variable "streamlit_service_url" {
  description = "URL of the Streamlit frontend service for CORS configuration"
  type        = string
  default     = ""
}

variable "lambda_function_arns" {
  description = "Map of Lambda function ARNs"
  type        = map(string)
}

variable "lambda_function_names" {
  description = "Map of Lambda function names"
  type        = map(string)
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "step_functions_role_arn" {
  description = "ARN of the IAM role for API Gateway to invoke Step Functions"
  type        = string
}

variable "step_functions_state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  type        = string
}
