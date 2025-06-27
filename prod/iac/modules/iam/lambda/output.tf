output "lambda_execution_role_arn" {
  description = "ARN of the Lambda execution IAM role"
  value       = aws_iam_role.lambda_execution_role.arn
}

output "lambda_execution_role_name" {
  description = "Name of the Lambda execution IAM role"
  value       = aws_iam_role.lambda_execution_role.name
}
