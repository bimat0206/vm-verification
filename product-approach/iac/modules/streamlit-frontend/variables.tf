variable "service_name" {
  description = "Name of the Streamlit application service"
  type        = string
}

variable "environment" {
  description = "Environment (e.g., dev, staging, prod)"
  type        = string
  default     = ""
}

variable "name_suffix" {
  description = "Suffix to add to resource names (optional)"
  type        = string
  default     = ""
}

variable "image_uri" {
  description = "URI of the Docker image for Streamlit app"
  type        = string
}

variable "image_repository_type" {
  description = "Type of the image repository (ECR or ECR_PUBLIC)"
  type        = string
  default     = "ECR_PUBLIC"

  validation {
    condition     = contains(["ECR", "ECR_PUBLIC"], var.image_repository_type)
    error_message = "Image repository type must be ECR or ECR_PUBLIC."
  }
}

variable "port" {
  description = "Port that Streamlit listens to in the container"
  type        = number
  default     = 8501
}

variable "cpu" {
  description = "CPU units for App Runner service"
  type        = string
  default     = "1 vCPU"

  validation {
    condition     = contains(["0.25 vCPU", "0.5 vCPU", "1 vCPU", "2 vCPU", "4 vCPU"], var.cpu)
    error_message = "CPU must be one of: 0.25 vCPU, 0.5 vCPU, 1 vCPU, 2 vCPU, or 4 vCPU."
  }
}

variable "memory" {
  description = "Memory for App Runner service"
  type        = string
  default     = "2 GB"

  validation {
    condition     = contains(["0.5 GB", "1 GB", "2 GB", "3 GB", "4 GB", "6 GB", "8 GB", "10 GB"], var.memory)
    error_message = "Memory must be one of: 0.5 GB, 1 GB, 2 GB, 3 GB, 4 GB, 6 GB, 8 GB, or 10 GB."
  }
}

variable "environment_variables" {
  description = "Environment variables for Streamlit application"
  type        = map(string)
  default     = {}
}

variable "auto_deployments_enabled" {
  description = "Whether to automatically deploy new images when pushed to the repository"
  type        = bool
  default     = true
}

variable "theme_mode" {
  description = "Streamlit theme mode (light or dark)"
  type        = string
  default     = "dark"

  validation {
    condition     = contains(["light", "dark"], var.theme_mode)
    error_message = "Theme mode must be 'light' or 'dark'."
  }
}

variable "health_check_path" {
  description = "Path for HTTP health checks"
  type        = string
  default     = "/_stcore/health"
}

variable "health_check_interval" {
  description = "Time interval, in seconds, between health checks"
  type        = number
  default     = 5
}

variable "health_check_timeout" {
  description = "Time, in seconds, to wait for a health check response"
  type        = number
  default     = 2
}

variable "health_check_healthy_threshold" {
  description = "Number of consecutive successful health checks before considering healthy"
  type        = number
  default     = 1
}

variable "health_check_unhealthy_threshold" {
  description = "Number of consecutive failed health checks before considering unhealthy"
  type        = number
  default     = 5
}

variable "enable_auto_scaling" {
  description = "Whether to enable auto scaling for App Runner service"
  type        = bool
  default     = false
}

variable "max_concurrency" {
  description = "Maximum number of concurrent requests per instance"
  type        = number
  default     = 100
}

variable "min_size" {
  description = "Minimum number of instances"
  type        = number
  default     = 1
}

variable "max_size" {
  description = "Maximum number of instances"
  type        = number
  default     = 5
}

variable "is_publicly_accessible" {
  description = "Whether the service is publicly accessible"
  type        = bool
  default     = true
}



variable "log_retention_days" {
  description = "Number of days to retain App Runner logs"
  type        = number
  default     = 14
}

variable "api_gateway_arn" {
  description = "ARN of the API Gateway to grant access to"
  type        = string
  default     = ""
}

variable "enable_api_gateway_access" {
  description = "Whether to enable API Gateway access for the Streamlit app"
  type        = bool
  default     = false
}

variable "s3_bucket_arns" {
  description = "List of S3 bucket ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "enable_s3_access" {
  description = "Whether to enable S3 access for the Streamlit app"
  type        = bool
  default     = false
}

variable "dynamodb_table_arns" {
  description = "List of DynamoDB table ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "enable_dynamodb_access" {
  description = "Whether to enable DynamoDB access for the Streamlit app"
  type        = bool
  default     = false
}

variable "enable_ecr_full_access" {
  description = "Whether to enable full ECR access for the App Runner instance role"
  type        = bool
  default     = true
}

# Secrets Manager access is now handled automatically based on environment variables

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
