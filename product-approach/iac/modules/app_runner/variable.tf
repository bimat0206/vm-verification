variable "service_name" {
  description = "Name of the App Runner service"
  type        = string
}

variable "image_uri" {
  description = "URI of the container image for App Runner service"
  type        = string
}

variable "image_repo_type" {
  description = "Type of the image repository (ECR or ECR_PUBLIC)"
  type        = string
  default     = "ECR_PUBLIC"
  
  validation {
    condition     = contains(["ECR", "ECR_PUBLIC"], var.image_repo_type)
    error_message = "Image repository type must be ECR or ECR_PUBLIC."
  }
}

variable "port" {
  description = "Port that your application listens to in the container"
  type        = number
  default     = 8080
}

variable "cpu" {
  description = "CPU units for App Runner service"
  type        = number
  default     = 1
  
  validation {
    condition     = contains([0.25, 0.5, 1, 2, 4], var.cpu)
    error_message = "CPU must be one of 0.25, 0.5, 1, 2, or 4."
  }
}

variable "memory" {
  description = "Memory in MB for App Runner service"
  type        = number
  default     = 2048
  
  validation {
    condition     = contains([0.5, 1, 2, 3, 4], floor(var.memory / 1024))
    error_message = "Memory must be approximately 0.5, 1, 2, 3, or 4 GB (512, 1024, 2048, 3072, or 4096 MB)."
  }
}

variable "environment_variables" {
  description = "Environment variables for App Runner service"
  type        = map(string)
  default     = {}
}

variable "auto_deployments_enabled" {
  description = "Whether to automatically deploy new images when pushed to the repository"
  type        = bool
  default     = false
}

variable "health_check_protocol" {
  description = "Protocol for health checks (TCP or HTTP)"
  type        = string
  default     = "HTTP"
  
  validation {
    condition     = contains(["TCP", "HTTP"], var.health_check_protocol)
    error_message = "Health check protocol must be TCP or HTTP."
  }
}

variable "health_check_path" {
  description = "Path for HTTP health checks"
  type        = string
  default     = "/"
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
  default     = 10
}

variable "is_publicly_accessible" {
  description = "Whether the service is publicly accessible"
  type        = bool
  default     = true
}

variable "ingress_vpc_configuration" {
  description = "VPC configuration for App Runner service"
  type = object({
    subnets         = list(string)
    security_groups = list(string)
  })
  default = null
}

variable "custom_domain_name" {
  description = "Custom domain name for App Runner service"
  type        = string
  default     = ""
}

variable "enable_www_subdomain" {
  description = "Whether to enable the www subdomain for the custom domain"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Number of days to retain App Runner logs"
  type        = number
  default     = 14
}

variable "api_gateway_access" {
  description = "Whether to grant access to API Gateway"
  type        = bool
  default     = false
}

variable "api_gateway_arn" {
  description = "ARN of the API Gateway to grant access to"
  type        = string
  default     = ""
}

variable "s3_bucket_arns" {
  description = "List of S3 bucket ARNs to grant read access to"
  type        = list(string)
  default     = []
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}