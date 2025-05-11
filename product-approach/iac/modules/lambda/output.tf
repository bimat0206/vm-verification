output "function_arns" {
  description = "ARNs of the Lambda functions"
  value       = { for k, v in aws_lambda_function.this : k => v.arn }
}

output "function_names" {
  description = "Names of the Lambda functions"
  value       = { for k, v in aws_lambda_function.this : k => v.function_name }
}

output "function_invoke_arns" {
  description = "Invoke ARNs of the Lambda functions"
  value       = { for k, v in aws_lambda_function.this : k => v.invoke_arn }
}

output "function_versions" {
  description = "Latest published Lambda function versions"
  value       = { for k, v in aws_lambda_function.this : k => v.version }
}

output "log_group_names" {
  description = "Names of the CloudWatch Log Groups for Lambda functions"
  value       = { for k, v in aws_cloudwatch_log_group.this : k => v.name }
}