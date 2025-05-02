variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "vending-machine-verification"
}

variable "environment" {
  description = "Environment (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "resource_name_suffix" {
  description = "Optional custom suffix to append to resource names (will be combined with auto-generated random suffix)"
  type        = string
  default     = ""
}

variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "s3_buckets" {
  description = "Configuration for S3 buckets"
  type = object({
    create_buckets = bool
    force_destroy  = bool
    lifecycle_rules = object({
      reference = list(any)
      checking  = list(any)
      results   = list(any)
    })
  })
  default = {
    create_buckets = true
    force_destroy  = false
    lifecycle_rules = {
      reference = []
      checking  = []
      results   = []
    }
  }
}

variable "dynamodb_tables" {
  description = "Configuration for DynamoDB tables"
  type = object({
    create_tables         = bool
    billing_mode          = string
    read_capacity         = number
    write_capacity        = number
    point_in_time_recovery = bool
  })
  default = {
    create_tables         = true
    billing_mode          = "PAY_PER_REQUEST"
    read_capacity         = 5
    write_capacity        = 5
    point_in_time_recovery = true
  }
}

variable "ecr" {
  description = "Configuration for ECR repositories"
  type = object({
    create_repositories = bool
  })
  default = {
    create_repositories = true
  }
}

variable "lambda_functions" {
  description = "Configuration for Lambda functions"
  type = object({
    create_functions        = bool
    use_ecr                 = bool
    default_image_uri       = string
    image_tag               = string
    architectures           = list(string)
    log_retention_days      = number
    s3_trigger_functions    = list(string)
    eventbridge_trigger_functions = list(string)
  })
  default = {
    create_functions        = true
    use_ecr                 = true
    default_image_uri       = ""
    image_tag               = "latest"
    architectures           = ["x86_64"]
    log_retention_days      = 7
    s3_trigger_functions    = []
    eventbridge_trigger_functions = []
  }
}

variable "step_functions" {
  description = "Configuration for Step Functions"
  type = object({
    create_step_functions = bool
    log_level             = string
  })
  default = {
    create_step_functions = true
    log_level             = "ALL"
  }
}

variable "api_gateway" {
  description = "Configuration for API Gateway"
  type = object({
    create_api_gateway     = bool
    stage_name             = string
    throttling_rate_limit  = number
    throttling_burst_limit = number
    cors_enabled           = bool
    metrics_enabled        = bool
  })
  default = {
    create_api_gateway     = true
    stage_name             = "v1"
    throttling_rate_limit  = 100
    throttling_burst_limit = 50
    cors_enabled           = true
    metrics_enabled        = true
  }
}

variable "app_runner" {
  description = "Configuration for App Runner service"
  type = object({
    create_app_runner       = bool
    image_uri               = string
    cpu                     = string
    memory                  = string
    auto_deployments_enabled = bool
    environment_variables   = map(string)
  })
  default = {
    create_app_runner       = false
    image_uri               = ""
    cpu                     = "1 vCPU"
    memory                  = "2 GB"
    auto_deployments_enabled = true
    environment_variables   = {}
  }
}

variable "monitoring" {
  description = "Configuration for monitoring resources"
  type = object({
    create_dashboard      = bool
    log_retention_days    = number
    alarm_email_endpoints = list(string)
  })
  default = {
    create_dashboard      = true
    log_retention_days    = 7
    alarm_email_endpoints = []
  }
}

variable "bedrock" {
  description = "Configuration for Amazon Bedrock"
  type = object({
    model_id = string
  })
  default = {
    model_id = "anthropic.claude-3-sonnet-20240229-v1:0"
  }
}
