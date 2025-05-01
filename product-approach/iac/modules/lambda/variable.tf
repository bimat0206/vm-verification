variable "functions_config" {
  description = "Configuration for Lambda functions"
  type = map(object({
    name                 = string
    description          = string
    memory_size          = number
    timeout              = number
    environment_variables = map(string)
    reserved_concurrent_executions = optional(number, -1)
  }))
}

variable "execution_role_arn" {
  description = "ARN of the IAM role for Lambda execution"
  type        = string
}

variable "use_ecr_repository" {
  description = "Whether to use ECR repository for Lambda function images"
  type        = bool
  default     = true
}

variable "ecr_repository_url" {
  description = "Base URL for ECR repository (without repository name)"
  type        = string
  default     = ""
}

variable "image_uri" {
  description = "URI of the container image to use for Lambda functions when not using ECR"
  type        = string
  default     = "public.ecr.aws/nginx/nginx:latest" # Placeholder image
}

variable "image_tag" {
  description = "Tag of the ECR image to use for Lambda functions"
  type        = string
  default     = "latest"
}

variable "architectures" {
  description = "Instruction set architecture for Lambda functions"
  type        = list(string)
  default     = ["x86_64"]
  validation {
    condition     = length([for arch in var.architectures : arch if contains(["x86_64", "arm64"], arch)]) == length(var.architectures)
    error_message = "Supported architectures are x86_64 and arm64 only."
  }
}

variable "log_retention_days" {
  description = "Number of days to retain Lambda function logs"
  type        = number
  default     = 14
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "api_gateway_source_arn" {
  description = "ARN of the API Gateway for Lambda permissions"
  type        = string
  default     = null
}

variable "s3_source_arns" {
  description = "Map of function names to S3 bucket ARNs for Lambda permissions"
  type        = map(string)
  default     = null
}

variable "s3_trigger_functions" {
  description = "List of Lambda function names that should be triggered by S3 events"
  type        = list(string)
  default     = []
}

variable "eventbridge_source_arns" {
  description = "Map of function names to EventBridge rule ARNs for Lambda permissions"
  type        = map(string)
  default     = null
}

variable "eventbridge_trigger_functions" {
  description = "List of Lambda function names that should be triggered by EventBridge events"
  type        = list(string)
  default     = []
}