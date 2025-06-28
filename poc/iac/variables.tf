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
    })
  })
  default = {
    create_buckets = true
    force_destroy  = false
    lifecycle_rules = {
      reference = []
      checking  = []
    }
  }
}

variable "dynamodb_tables" {
  description = "Configuration for DynamoDB tables"
  type = object({
    create_tables          = bool
    billing_mode           = string
    read_capacity          = number
    write_capacity         = number
    point_in_time_recovery = bool
  })
  default = {
    create_tables          = true
    billing_mode           = "PAY_PER_REQUEST"
    read_capacity          = 5
    write_capacity         = 5
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
    create_functions              = bool
    use_ecr                       = bool
    default_image_uri             = string
    image_tag                     = string
    architectures                 = list(string)
    log_retention_days            = number
    s3_trigger_functions          = list(string)
    eventbridge_trigger_functions = list(string)
  })
  default = {
    create_functions              = true
    use_ecr                       = true
    default_image_uri             = ""
    image_tag                     = "latest"
    architectures                 = ["x86_64"]
    log_retention_days            = 7
    s3_trigger_functions          = []
    eventbridge_trigger_functions = []
  }
}

variable "step_functions" {
  description = "Configuration for Step Functions"
  type = object({
    create_step_functions = bool
    log_level             = string
    enable_x_ray_tracing  = bool
  })
  default = {
    create_step_functions = true
    log_level             = "ALL"
    enable_x_ray_tracing  = true
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
    use_api_key            = bool
  })
}

variable "monitoring" {
  description = "Configuration for monitoring resources"
  type = object({
    create_dashboard   = bool
    log_retention_days = number
  })
  default = {
    create_dashboard   = true
    log_retention_days = 7
  }
}

variable "bedrock" {
  description = "Configuration for Amazon Bedrock"
  type = object({
    model_id          = string
    anthropic_version = string
    max_tokens        = number
    budget_tokens     = number
  })
}

variable "vpc" {
  description = "Configuration for VPC"
  type = object({
    create_vpc         = bool
    vpc_cidr           = string
    availability_zones = list(string)
    create_nat_gateway = bool
  })
  /*
  default = {
    create_vpc         = true
    vpc_cidr           = "10.0.0.0/16"
    availability_zones = ["us-east-1a", "us-east-1b"]
    create_nat_gateway = true
  }
  */
}


variable "react_frontend" {
  description = "Configuration for React frontend application"
  type = object({
    create_react                   = bool
    service_name                   = string
    image_uri                      = string
    image_repository_type          = string
    cpu                            = number
    memory                         = number
    port                           = number
    auto_deployments_enabled       = bool
    enable_auto_scaling            = bool
    min_size                       = number
    max_capacity                   = number
    max_size                       = number
    cpu_threshold                  = number
    memory_threshold               = number
    log_retention_days             = number
    health_check_path              = string
    health_check_interval          = number
    health_check_timeout           = number
    health_check_healthy_threshold = number
    health_check_unhealthy_threshold = number
    enable_https                   = bool
    internal_alb                   = bool
    enable_container_insights      = bool
    enable_execute_command         = bool
    environment_variables          = map(string)
    certificate_arn                = string
  })
  /*
  default = {
    create_react                   = false
    service_name                   = "react-frontend"
    image_uri                      = ""
    image_repository_type          = "ECR"
    cpu                            = 1024
    memory                         = 2048
    port                           = 3000
    auto_deployments_enabled       = true
    enable_auto_scaling            = true
    min_size                       = 1
    max_capacity                   = 10
    max_size                       = 3
    cpu_threshold                  = 70
    memory_threshold               = 70
    log_retention_days             = 30
    health_check_path              = "/api/health"
    health_check_interval          = 30
    health_check_timeout           = 5
    health_check_healthy_threshold = 2
    health_check_unhealthy_threshold = 3
    enable_https                   = false
    internal_alb                   = false
    enable_container_insights      = false
    enable_execute_command         = true
    environment_variables          = {}
    certificate_arn                = ""
  }
  */
}
