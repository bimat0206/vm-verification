# infrastructure/modules/streamlit_frontend_ecs/variables.tf

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "vending-verification"
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

# VPC configuration
variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

# ECR configuration
variable "image_tag_mutability" {
  description = "Image tag mutability setting for the ECR repository"
  type        = string
  default     = "MUTABLE"
  validation {
    condition     = contains(["MUTABLE", "IMMUTABLE"], var.image_tag_mutability)
    error_message = "Allowed values are MUTABLE or IMMUTABLE"
  }
}

variable "enable_scan_on_push" {
  description = "Whether to enable vulnerability scanning on image push"
  type        = bool
  default     = true
}

variable "kms_key_arn" {
  description = "ARN of KMS key for ECR repository encryption"
  type        = string
  default     = null
}

variable "max_image_count" {
  description = "Maximum number of images to keep in ECR repository"
  type        = number
  default     = 5
}

# Application configuration
variable "api_endpoint" {
  description = "API Gateway endpoint URL"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket"
  type        = string
}

variable "step_functions_arn" {
  description = "ARN of the Step Functions state machine"
  type        = string
}

variable "additional_config" {
  description = "Additional configuration for Streamlit app"
  type        = map(string)
  default     = {}
}

# Container configuration
variable "container_port" {
  description = "Port exposed by the Streamlit container"
  type        = number
  default     = 8501
}

variable "image_tag" {
  description = "Tag for the Docker image"
  type        = string
  default     = "latest"
}

# ECS configuration
variable "cpu" {
  description = "CPU units for ECS task (1024 = 1 vCPU)"
  type        = string
  default     = "1024"
}

variable "memory" {
  description = "Memory for ECS task in MB"
  type        = string
  default     = "2048"
}

variable "min_capacity" {
  description = "Minimum number of ECS tasks"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum number of ECS tasks"
  type        = number
  default     = 5
}

# ALB configuration
variable "certificate_arn" {
  description = "ARN of the SSL certificate for HTTPS"
  type        = string
  default     = null
}

# CloudWatch configuration
variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
}

# Docker build configuration
variable "build_and_push_image" {
  description = "Whether to build and push Docker image during Terraform apply"
  type        = bool
  default     = false
}

variable "app_source_path" {
  description = "Path to the Streamlit app source code"
  type        = string
  default     = "./frontend"
}

variable "source_code_hash" {
  description = "Hash of the source code to trigger rebuild when changed"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}