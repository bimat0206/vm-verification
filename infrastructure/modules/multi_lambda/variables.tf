# infrastructure/modules/multi_lambda/variables.tf
variable "name_prefix" {
  description = "Prefix for Lambda function names"
  type        = string
  default     = "vending-verification"
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "filename" {
  description = "Path to the Lambda function deployment package"
  type        = string
  default     = null
}

#variable "runtime" {
#  description = "Lambda function runtime"
#  type        = string
#}

variable "architectures" {
  description = "Lambda function architectures"
  type        = list(string)
  default     = ["arm64"]
}

variable "environment_variables" {
  description = "Additional environment variables for the Lambda functions"
  type        = map(string)
  default     = {}
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket for storing images"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for storing images"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table for verification results"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for verification results"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "secrets_arn" {
  description = "ARN of the Secrets Manager secret for Bedrock API key"
  type        = string
  default     = null
}

variable "enable_secrets_access" {
  description = "Whether to enable access to Secrets Manager"
  type        = bool
  default     = false
}

variable "bedrock_model_id" {
  description = "Bedrock model ID for Claude"
  type        = string
}

variable "ecr_image_uri" {
  description = "ECR repository URI for Lambda container image (used for backwards compatibility)"
  type        = string
  default     = null
}

variable "ecr_repository_urls" {
  description = "Map of function names to ECR repository URLs"
  type        = map(string)
  default     = {}
}

variable "image_command" {
  description = "Command to run in the container image"
  type        = list(string)
  default     = null
}

variable "tags" {
  description = "Additional tags to apply to the Lambda functions"
  type        = map(string)
  default     = {}
}

variable "skip_lambda_function_creation" {
  description = "Whether to skip Lambda function creation (for infrastructure-only deployments)"
  type        = bool
  default     = false
}