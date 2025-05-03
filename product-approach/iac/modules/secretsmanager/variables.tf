variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment (e.g., dev, staging, prod)"
  type        = string
  default     = ""
}

variable "name_suffix" {
  description = "Suffix to add to resource names"
  type        = string
  default     = ""
}

variable "secret_base_name" {
  description = "Base name for the secret (without prefix/suffix)"
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