# infrastructure/outputs.tf
output "s3_bucket_name" {
  description = "Name of the S3 bucket for storing images"
  value       = module.images_bucket.bucket_id
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for storing images"
  value       = module.images_bucket.bucket_arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = module.verification_results.table_name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = module.verification_results.table_arn
}

output "lambda_function_names" {
  description = "Names of all Lambda functions"
  value       = module.lambda_functions.function_names
}

output "lambda_function_arns" {
  description = "ARNs of all Lambda functions"
  value       = module.lambda_functions.function_arns
}

output "state_machine_name" {
  description = "Name of the Step Functions state machine"
  value       = module.step_functions.state_machine_name
}

output "state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  value       = module.step_functions.state_machine_arn
}

output "api_gateway_invoke_url" {
  description = "Invoke URL for the API Gateway"
  value       = module.api_gateway.invoke_url
}

output "api_gateway_endpoint" {
  description = "API Gateway endpoint URL"
  value       = module.api_gateway.api_endpoint
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = module.ecr.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  value       = module.ecr.repository_arn
}

output "cloudwatch_dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = module.monitoring.dashboard_name
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret"
  value       = module.secrets.secret_arn
}