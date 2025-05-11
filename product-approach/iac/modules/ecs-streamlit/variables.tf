# Variables for ECS Streamlit module

variable "service_name" {
  description = "Name of the service"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = ""
}

variable "name_suffix" {
  description = "Suffix to append to resource names"
  type        = string
  default     = ""
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Container configuration
variable "image_uri" {
  description = "URI of the container image"
  type        = string
}

variable "image_repository_type" {
  description = "Type of image repository (ECR or ECR_PUBLIC)"
  type        = string
  default     = "ECR"
}

variable "port" {
  description = "Port on which the container is listening"
  type        = number
  default     = 8501  # Default Streamlit port
}

variable "cpu" {
  description = "CPU units for the task (e.g., 256, 512, 1024, 2048, 4096)"
  type        = number
  default     = 256
}

variable "memory" {
  description = "Memory for the task in MiB (e.g., 512, 1024, 2048, 4096, 8192)"
  type        = number
  default     = 512
}

variable "environment_variables" {
  description = "Environment variables for the container"
  type        = map(string)
  default     = {}
}

variable "theme_mode" {
  description = "Streamlit theme mode (light or dark)"
  type        = string
  default     = "light"
}

# ECS configuration
variable "enable_container_insights" {
  description = "Whether to enable CloudWatch Container Insights for the ECS cluster"
  type        = bool
  default     = false
}

variable "enable_execute_command" {
  description = "Whether to enable execute command functionality for the ECS service"
  type        = bool
  default     = false
}

variable "min_capacity" {
  description = "Minimum number of tasks to run"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum number of tasks to run"
  type        = number
  default     = 10
}

variable "enable_auto_scaling" {
  description = "Whether to enable auto scaling for the ECS service"
  type        = bool
  default     = true
}

variable "cpu_threshold" {
  description = "CPU utilization threshold for scaling"
  type        = number
  default     = 70
}

variable "memory_threshold" {
  description = "Memory utilization threshold for scaling"
  type        = number
  default     = 70
}

# VPC and networking
variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the ALB"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for the ECS tasks"
  type        = list(string)
}

variable "alb_security_group_id" {
  description = "ID of the security group for the ALB"
  type        = string
}

variable "ecs_security_group_id" {
  description = "ID of the security group for the ECS tasks"
  type        = string
}

# ALB configuration
variable "internal_alb" {
  description = "Whether the ALB is internal"
  type        = bool
  default     = false
}

variable "enable_deletion_protection" {
  description = "Whether to enable deletion protection for the ALB"
  type        = bool
  default     = false
}

variable "alb_idle_timeout" {
  description = "Idle timeout for the ALB in seconds"
  type        = number
  default     = 60
}

variable "alb_access_logs_bucket" {
  description = "S3 bucket for ALB access logs"
  type        = string
  default     = ""
}

variable "alb_access_logs_prefix" {
  description = "S3 prefix for ALB access logs"
  type        = string
  default     = ""
}

variable "deregistration_delay" {
  description = "Deregistration delay for the target group in seconds"
  type        = number
  default     = 30
}

# Health check configuration
variable "health_check_path" {
  description = "Path for health checks"
  type        = string
  default     = "/_stcore/health"
}

variable "health_check_interval" {
  description = "Interval between health checks in seconds"
  type        = number
  default     = 30
}

variable "health_check_timeout" {
  description = "Timeout for health checks in seconds"
  type        = number
  default     = 5
}

variable "health_check_healthy_threshold" {
  description = "Number of consecutive successful health checks to be considered healthy"
  type        = number
  default     = 2
}

variable "health_check_unhealthy_threshold" {
  description = "Number of consecutive failed health checks to be considered unhealthy"
  type        = number
  default     = 3
}

# HTTPS configuration
variable "enable_https" {
  description = "Whether to enable HTTPS"
  type        = bool
  default     = false
}

variable "ssl_policy" {
  description = "SSL policy for the HTTPS listener"
  type        = string
  default     = "ELBSecurityPolicy-2016-08"
}

variable "certificate_arn" {
  description = "ARN of the SSL certificate"
  type        = string
  default     = ""
}

# Logging
variable "log_retention_days" {
  description = "Number of days to retain logs"
  type        = number
  default     = 30
}

# Access permissions
variable "enable_ecr_full_access" {
  description = "Whether to enable full access to ECR"
  type        = bool
  default     = false
}

variable "enable_api_gateway_access" {
  description = "Whether to enable access to API Gateway"
  type        = bool
  default     = false
}

variable "api_gateway_arn" {
  description = "ARN of the API Gateway"
  type        = string
  default     = ""
}

variable "enable_s3_access" {
  description = "Whether to enable access to S3"
  type        = bool
  default     = false
}

variable "s3_bucket_arns" {
  description = "List of S3 bucket ARNs"
  type        = list(string)
  default     = []
}

variable "enable_dynamodb_access" {
  description = "Whether to enable access to DynamoDB"
  type        = bool
  default     = false
}

variable "dynamodb_table_arns" {
  description = "List of DynamoDB table ARNs"
  type        = list(string)
  default     = []
}

# Auto deployments
variable "auto_deployments_enabled" {
  description = "Whether to enable automatic deployments"
  type        = bool
  default     = true
}
