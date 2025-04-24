# infrastructure/modules/multi_ecr/variables.tf

variable "repository_prefix" {
  description = "Prefix for ECR repository names"
  type        = string
  default     = "vending-verification"
}
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

variable "aws_region" {
  description = "AWS region where the repositories will be created"
  type        = string
  default     = "us-east-1"
}

variable "image_tag_mutability" {
  description = "The tag mutability setting for the repositories"
  type        = string
  default     = "MUTABLE"
  validation {
    condition     = contains(["MUTABLE", "IMMUTABLE"], var.image_tag_mutability)
    error_message = "image_tag_mutability must be either MUTABLE or IMMUTABLE"
  }
}

variable "enable_scan_on_push" {
  description = "Indicates whether images are scanned after being pushed to the repositories"
  type        = bool
  default     = true
}

variable "max_image_count" {
  description = "Maximum number of images to keep in each repository"
  type        = number
  default     = 5
}

variable "kms_key_arn" {
  description = "The ARN of the KMS key to use for encryption"
  type        = string
  default     = null
}

variable "push_placeholder_images" {
  description = "Whether to push placeholder nginx images to the repositories"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags to apply to the repositories"
  type        = map(string)
  default     = {}
}