# infrastructure/modules/monitoring/variables.tf
variable "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "lambda_functions" {
  description = "List of Lambda function names to monitor"
  type        = list(string)
  default     = []
}

variable "state_machine_arn" {
  description = "ARN of the Step Functions state machine to monitor"
  type        = string
}

variable "state_machine_name" {
  description = "Name of the Step Functions state machine to monitor"
  type        = string
}

variable "dynamodb_table" {
  description = "Name of the DynamoDB table to monitor"
  type        = string
}

variable "api_gateway_api_name" {
  description = "Name of the API Gateway API to monitor"
  type        = string
}

variable "api_gateway_stage_name" {
  description = "Name of the API Gateway stage to monitor"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket to monitor"
  type        = string
}

variable "lambda_error_threshold" {
  description = "Threshold for Lambda error alarm"
  type        = number
  default     = 1
}

variable "api_gateway_error_threshold" {
  description = "Threshold for API Gateway 5XX error alarm"
  type        = number
  default     = 1
}

variable "step_functions_failure_threshold" {
  description = "Threshold for Step Functions execution failure alarm"
  type        = number
  default     = 1
}

variable "log_retention_days" {
  description = "Number of days to retain logs"
  type        = number
  default     = 90
}

variable "alarm_actions" {
  description = "List of ARNs to notify when the alarm transitions to the ALARM state"
  type        = list(string)
  default     = []
}

variable "ok_actions" {
  description = "List of ARNs to notify when the alarm transitions to the OK state"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}