# infrastructure/modules/monitoring/outputs.tf
output "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.main.dashboard_name
}

output "dashboard_arn" {
  description = "ARN of the CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.main.dashboard_arn
}

output "lambda_alarms" {
  description = "Map of Lambda function names to their alarm ARNs"
  value = {
    for name, alarm in aws_cloudwatch_metric_alarm.lambda_errors : name => alarm.arn
  }
}

output "api_gateway_alarm_arn" {
  description = "ARN of the API Gateway 5XX errors alarm"
  value       = aws_cloudwatch_metric_alarm.api_gateway_5xx.arn
}

output "step_functions_alarm_arn" {
  description = "ARN of the Step Functions execution failures alarm"
  value       = aws_cloudwatch_metric_alarm.step_functions_failures.arn
}

output "lambda_log_groups" {
  description = "Map of Lambda function names to their log group ARNs"
  value = {
    for name, log_group in aws_cloudwatch_log_group.lambda_logs : name => log_group.arn
  }
}

output "api_gateway_log_group_arn" {
  description = "ARN of the API Gateway log group"
  value       = aws_cloudwatch_log_group.api_gateway_logs.arn
}

output "state_machine_log_group_arn" {
  description = "ARN of the Step Functions state machine log group"
  value       = aws_cloudwatch_log_group.state_machine_logs.arn
}