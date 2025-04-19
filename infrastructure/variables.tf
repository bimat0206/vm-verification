# infrastructure/variables.tf
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for storing images"
  type        = string
  default     = "vending-verification-images"
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for verification results"
  type        = string
  default     = "VerificationResults"
}

variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30
}

variable "lambda_memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 256
}

variable "ecr_repository_name" {
  description = "Name of the ECR repository"
  type        = string
  default     = "vending-verification"
}

variable "ecr_image_tag_mutability" {
  description = "Image tag mutability for the ECR repository"
  type        = string
  default     = "MUTABLE"
}

variable "ecr_enable_scan_on_push" {
  description = "Whether to enable image scanning on push"
  type        = bool
  default     = true
}

variable "ecr_kms_key_arn" {
  description = "ARN of the KMS key for ECR encryption"
  type        = string
  default     = null
}

variable "ecr_max_image_count" {
  description = "Maximum number of images to keep in the ECR repository"
  type        = number
  default     = 30
}

variable "alb_certificate_arn" {
  description = "ARN of the SSL certificate for ALB HTTPS"
  type        = string
  default     = null
}

# VPC Variables
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "172.16.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["172.16.1.0/24", "172.16.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["172.16.3.0/24", "172.16.4.0/24"]
}

variable "enable_nat_gateway" {
  description = "Whether to enable NAT Gateway for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Whether to use a single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "ecr_image_uri" {
  description = "ECR image URI for Lambda container deployment"
  type        = string
  default     = null
}

# New variables for the updated architecture
variable "bedrock_api_key" {
  description = "API key for Bedrock service"
  type        = string
  default     = ""
  sensitive   = true
}

variable "step_functions_execution_timeout" {
  description = "Timeout for Step Functions execution in seconds"
  type        = number
  default     = 300
}

variable "enable_api_gateway_logging" {
  description = "Whether to enable logging for API Gateway"
  type        = bool
  default     = true
}

variable "api_gateway_throttling_rate_limit" {
  description = "Rate limit for API Gateway throttling"
  type        = number
  default     = 100
}

variable "api_gateway_throttling_burst_limit" {
  description = "Burst limit for API Gateway throttling"
  type        = number
  default     = 50
}

variable "cloudwatch_logs_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 90
}

variable "enable_alarm_notifications" {
  description = "Whether to enable alarm notifications"
  type        = bool
  default     = false
}

variable "alarm_email" {
  description = "Email address for alarm notifications"
  type        = string
  default     = ""
}
# Centralized Tags
variable "default_tags" {
  description = "Map of default tags to apply to all resources"
  type        = map(string)
  default     = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

variable "additional_tags" {
  description = "Map of additional tags to apply to resources (will be merged with default_tags)"
  type        = map(string)
  default     = {}
}
# infrastructure/variables.tf (add these variables)

# Variables for Lambda/ECR deployment
variable "skip_lambda_functions" {
  description = "Whether to skip Lambda function creation (for infrastructure-only deployments)"
  type        = bool
  default     = true
}

variable "push_placeholder_images" {
  description = "Whether to push placeholder nginx images to the ECR repositories"
  type        = bool
  default     = true
}

# Add to terraform.tfvars
# skip_lambda_functions = true
# push_placeholder_images = true