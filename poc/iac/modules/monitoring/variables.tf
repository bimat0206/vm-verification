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

variable "enable_step_function_monitoring" {
  description = "Whether to enable Step Functions monitoring"
  type        = bool
  default     = false
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

variable "enable_api_gateway_monitoring" {
  description = "Whether to enable API Gateway monitoring"
  type        = bool
  default     = false
}

variable "ecr_repository_names" {
  description = "Map of ECR repository names to monitor"
  type        = map(string)
  default     = {}
}

variable "ecs_cluster_name" {
  description = "Name of the ECS cluster to monitor"
  type        = string
  default     = ""
}

variable "ecs_service_name" {
  description = "Name of the ECS service to monitor"
  type        = string
  default     = ""
}

variable "enable_ecs_monitoring" {
  description = "Whether to enable ECS monitoring"
  type        = bool
  default     = false
}

variable "alb_name" {
  description = "Name of the ALB to monitor"
  type        = string
  default     = ""
}

variable "enable_alb_monitoring" {
  description = "Whether to enable ALB monitoring"
  type        = bool
  default     = false
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 14
}


variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "cloudwatch_kms_key_id" {
  description = "KMS Key ID for CloudWatch Logs encryption"
  type        = string
  default     = null
}
