variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "filename" {
  description = "Path to the Lambda function deployment package"
  type        = string
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
}

variable "runtime" {
  description = "Lambda function runtime"
  type        = string
}

variable "architectures" {
  description = "Lambda function architectures"
  type        = list(string)
  default     = ["x86_64"]
}

variable "timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30
}

variable "memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 128
}

variable "environment_variables" {
  description = "Environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "additional_policy_actions" {
  description = "Additional IAM policy actions"
  type        = list(string)
  default     = []
}

variable "additional_policy_resources" {
  description = "Additional IAM policy resources"
  type        = list(string)
  default     = []
}

variable "enable_api_gateway_integration" {
  description = "Enable API Gateway integration"
  type        = bool
  default     = false
}

variable "api_gateway_source_arn" {
  description = "API Gateway source ARN"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags to apply to the Lambda function"
  type        = map(string)
  default     = {}
}

variable "ecr_image_uri" {
  description = "ECR repository URI for Lambda container image"
  type        = string
  default     = null
}

variable "image_command" {
  description = "Command to run in the container image"
  type        = list(string)
  default     = null
}
