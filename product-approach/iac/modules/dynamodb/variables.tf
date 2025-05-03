variable "verification_results_table_name" {
  description = "Name of the DynamoDB table for verification results"
  type        = string
}

variable "layout_metadata_table_name" {
  description = "Name of the DynamoDB table for layout metadata"
  type        = string
}

variable "conversation_history_table_name" {
  description = "Name of the DynamoDB table for conversation history"
  type        = string
}

variable "billing_mode" {
  description = "DynamoDB billing mode (PROVISIONED or PAY_PER_REQUEST)"
  type        = string
  default     = "PAY_PER_REQUEST"
  
  validation {
    condition     = contains(["PROVISIONED", "PAY_PER_REQUEST"], var.billing_mode)
    error_message = "Billing mode must be either PROVISIONED or PAY_PER_REQUEST."
  }
}

variable "read_capacity" {
  description = "Read capacity units for DynamoDB tables (used only with PROVISIONED billing mode)"
  type        = number
  default     = 5
}

variable "write_capacity" {
  description = "Write capacity units for DynamoDB tables (used only with PROVISIONED billing mode)"
  type        = number
  default     = 5
}

variable "point_in_time_recovery" {
  description = "Whether to enable point-in-time recovery for DynamoDB tables"
  type        = bool
  default     = false
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}