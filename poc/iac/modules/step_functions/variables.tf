# modules/step_functions/variables.tf

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

variable "enable_x_ray_tracing" {
  description = "Whether to enable X-Ray tracing for the state machine"
  type        = bool
  default     = true
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}


variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# Add new variables for DynamoDB integration
variable "dynamodb_table_arns" {
  description = "List of DynamoDB table ARNs that Step Functions needs access to"
  type        = list(string)
  default     = []
}

variable "dynamodb_verification_table" {
  description = "Name of the DynamoDB table for verification results"
  type        = string
  default     = ""
}

variable "dynamodb_conversation_table" {
  description = "Name of the DynamoDB table for conversation history"
  type        = string
  default     = ""
}

