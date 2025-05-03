# modules/secretsmanager/variables.tf

variable "secret_name" {
  description = "Name of the Secrets Manager secret"
  type        = string
}

variable "secret_description" {
  description = "Description of the Secrets Manager secret"
  type        = string
  default     = "Secret created by Terraform"
}

variable "secret_value" {
  description = "Value of the secret (e.g., API key)"
  type        = string
  sensitive   = true
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}