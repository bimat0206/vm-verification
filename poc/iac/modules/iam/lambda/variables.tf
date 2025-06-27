variable "project_name" {
  description = "Project name to be used in resource names"
  type        = string
}

variable "environment" {
  description = "Environment name to be used in resource names"
  type        = string
}

variable "name_suffix" {
  description = "Suffix to be used in resource names"
  type        = string
}

variable "s3_bucket_arns" {
  description = "List of ARNs of S3 buckets Lambda functions need access to"
  type        = list(string)
  default     = []
}

variable "dynamodb_table_arns" {
  description = "List of ARNs of DynamoDB tables Lambda functions need access to"
  type        = list(string)
  default     = []
}

variable "ecr_repository_arns" {
  description = "List of ARNs of ECR repositories Lambda functions need access to"
  type        = list(string)
  default     = []
}

variable "bedrock_model_arn" {
  description = "ARN of the Bedrock model Lambda functions need access to"
  type        = string
  default     = ""
}

variable "sns_topic_arns" {
  description = "List of ARNs of SNS topics Lambda functions need to publish to"
  type        = list(string)
  default     = null
}

variable "step_functions_arns" {
  description = "List of ARNs of Step Functions state machines Lambda functions need to interact with"
  type        = list(string)
  default     = null
}

variable "secrets_manager_arns" {
  description = "List of ARNs of Secrets Manager secrets Lambda functions need access to"
  type        = list(string)
  default     = null
}

variable "enable_xray" {
  description = "Whether to enable X-Ray tracing for Lambda functions"
  type        = bool
  default     = false
}

variable "lambda_in_vpc" {
  description = "Whether Lambda functions will be deployed in a VPC"
  type        = bool
  default     = false
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}