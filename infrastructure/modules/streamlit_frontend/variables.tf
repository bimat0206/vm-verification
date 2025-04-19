# infrastructure/modules/streamlit_frontend/variables.tf

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

variable "auto_deployments_enabled" {
  description = "Whether to enable automatic deployments"
  type        = bool
  default     = true
}

variable "cpu" {
  description = "CPU units for App Runner service (1024 units = 1 vCPU)"
  type        = string
  default     = "1024"
  validation {
    condition     = contains(["1024", "2048", "4096"], var.cpu)
    error_message = "CPU value must be 1024 (1 vCPU), 2048 (2 vCPU), or 4096 (4 vCPU)."
  }
}

variable "memory" {
  description = "Memory for App Runner service in MB"
  type        = string
  default     = "2048"
  validation {
    condition     = contains(["2048", "3072", "4096", "6144", "8192", "10240", "12288"], var.memory)
    error_message = "Memory value must be one of: 2048, 3072, 4096, 6144, 8192, 10240, or 12288 MB."
  }
}

variable "max_concurrency" {
  description = "Maximum requests that can be processed concurrently"
  type        = number
  default     = 100
}

variable "max_size" {
  description = "Maximum number of instances"
  type        = number
  default     = 5
}

variable "min_size" {
  description = "Minimum number of instances"
  type        = number
  default     = 1
}

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