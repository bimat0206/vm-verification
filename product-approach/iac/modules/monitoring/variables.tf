variable "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  type        = string
}

variable "region" {
  description = "AWS region for CloudWatch metrics"
  type        = string
  default     = "us-east-1"
}

variable "lambda_function_names" {
  description = "Map of Lambda function names to monitor"
  type        = map(string)
  default     = {}
}

variable "step_function_name" {
  description = "Name of the Step Functions state machine to monitor"
  type        = string
  default     = ""
}

variable "dynamodb_table_names" {
  description = "Map of DynamoDB table names to monitor"
  type        = map(string)
  default     = {}
}

variable "api_gateway_name" {
  description = "Name of the API Gateway to monitor"
  type        = string
  default     = ""
}

variable "ecr_repository_names" {
  description = "Map of ECR repository names to monitor"
  type        = map(string)
  default     = {}
}

variable "app_runner_service_name" {
  description = "Name of the App Runner service to monitor"
  type        = string
  default     = ""
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 14
}

variable "alarm_email_endpoints" {
  description = "List of email addresses to send alarm notifications to"
  type        = list(string)
  default     = []
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}