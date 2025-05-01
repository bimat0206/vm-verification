variable "state_machine_name" {
  description = "Name of the Step Functions state machine"
  type        = string
}

variable "lambda_function_arns" {
  description = "Map of Lambda function names to their ARNs"
  type        = map(string)
}

variable "log_level" {
  description = "Log level for Step Functions (ALL, ERROR, FATAL, OFF)"
  type        = string
  default     = "ALL"
  
  validation {
    condition     = contains(["ALL", "ERROR", "FATAL", "OFF"], var.log_level)
    error_message = "Log level must be one of: ALL, ERROR, FATAL, OFF."
  }
}

variable "log_retention_days" {
  description = "Number of days to retain Step Functions logs"
  type        = number
  default     = 14
}

variable "create_definition_file" {
  description = "Whether to create a local file with the generated state machine definition"
  type        = bool
  default     = false
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}