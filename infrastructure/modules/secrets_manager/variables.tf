# infrastructure/modules/secrets_manager/variables.tf
variable "secret_name" {
  description = "Name of the secret"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "description" {
  description = "Description of the secret"
  type        = string
  default     = ""
}

variable "recovery_window_in_days" {
  description = "Number of days that AWS Secrets Manager waits before it can delete the secret"
  type        = number
  default     = 30
  validation {
    condition     = var.recovery_window_in_days >= 0 && var.recovery_window_in_days <= 30
    error_message = "Recovery window must be between 0 and 30 days"
  }
}

variable "kms_key_id" {
  description = "ARN or ID of the AWS KMS key to be used to encrypt the secret values"
  type        = string
  default     = null
}

variable "create_secret_version" {
  description = "Whether to create a secret version with the provided value"
  type        = bool
  default     = true
}

variable "secret_string_value" {
  description = "Text to store in the secret"
  type        = string
  default     = null
  sensitive   = true
}

variable "secret_string_map" {
  description = "Key/value map to store as JSON in the secret"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "create_bedrock_policy" {
  description = "Whether to create an IAM policy for Bedrock access"
  type        = bool
  default     = false
}

variable "aws_region" {
  description = "AWS region where the Bedrock service is located"
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "Additional tags to apply to the secret"
  type        = map(string)
  default     = {}
}