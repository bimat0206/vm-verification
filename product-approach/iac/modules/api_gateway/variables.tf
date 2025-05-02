# product-approach/iac/modules/api_gateway/variable.tf

variable "api_name" {
  description = "Name of the API Gateway"
  type        = string
}

variable "api_description" {
  description = "Description of the API Gateway"
  type        = string
  default     = "API Gateway created by Terraform"
}

variable "stage_name" {
  description = "Name of the API Gateway stage"
  type        = string
  default     = "v1"
}

variable "throttling_rate_limit" {
  description = "Rate limit for API Gateway throttling"
  type        = number
  default     = 100
}

variable "throttling_burst_limit" {
  description = "Burst limit for API Gateway throttling"
  type        = number
  default     = 50
}

variable "cors_enabled" {
  description = "Whether CORS is enabled for the API Gateway"
  type        = bool
  default     = true
}

variable "cors_allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = ["*"]
}

variable "metrics_enabled" {
  description = "Whether detailed metrics are enabled for the API Gateway"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Number of days to retain API Gateway logs"
  type        = number
  default     = 7
}

variable "use_api_key" {
  description = "Whether to use API key for the API Gateway"
  type        = bool
  default     = false
}

variable "lambda_function_arns" {
  description = "Map of Lambda function ARNs to integrate with API Gateway"
  type        = map(string)
}

variable "lambda_function_names" {
  description = "Map of Lambda function names to integrate with API Gateway"
  type        = map(string)
}

variable "api_quota_limit" {
  description = "Quota limit for API Gateway usage plan"
  type        = number
  default     = 1000
}

variable "api_quota_period" {
  description = "Quota period for API Gateway usage plan (DAY, WEEK, or MONTH)"
  type        = string
  default     = "DAY"
}

variable "openapi_definition" {
  description = "Path to the OpenAPI definition file"
  type        = string
  default     = "openapi.yaml"
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "region" {
  description = "AWS region where the API Gateway is deployed"
  type        = string
  default     = "us-east-1"
}
