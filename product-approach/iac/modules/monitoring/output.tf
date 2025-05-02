output "dashboard_name" {
  description = "Name of the created CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.this.dashboard_name
}

output "dashboard_arn" {
  description = "ARN of the created CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.this.dashboard_arn
}

output "alarm_sns_topic_arn" {
  description = "ARN of the SNS topic for CloudWatch alarms"
  value       = length(var.alarm_email_endpoints) > 0 ? aws_sns_topic.alarms[0].arn : ""
}

output "lambda_alarms" {
  description = "Map of Lambda function alarm names"
  value       = { for name, function_name in var.lambda_function_names : name => aws_cloudwatch_metric_alarm.lambda_errors[name].alarm_name }
}

output "dynamodb_alarms" {
  description = "Map of DynamoDB table alarm names"
  value       = { for name, table_name in var.dynamodb_table_names : name => aws_cloudwatch_metric_alarm.dynamodb_throttles[name].alarm_name }
}

output "log_groups" {
  description = "Map of resource types to their CloudWatch log group names (now managed by respective service modules)"
  value = {
    lambda        = { for name, function_name in var.lambda_function_names : name => "/aws/lambda/${function_name}" }
    step_function = var.step_function_name != "" ? { "${var.step_function_name}" = "/aws/states/${var.step_function_name}" } : {}
    api_gateway   = var.api_gateway_name != "" ? { "${var.api_gateway_name}" = "/aws/apigateway/${var.api_gateway_name}" } : {}
    app_runner    = var.app_runner_service_name != "" ? { "${var.app_runner_service_name}" = "/aws/apprunner/${var.app_runner_service_name}" } : {}
  }
}
