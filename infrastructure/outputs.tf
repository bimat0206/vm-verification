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

output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = module.api_lambda.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = module.api_lambda.function_arn
}

output "lambda_invoke_arn" {
  description = "Invocation ARN of the Lambda function"
  value       = module.api_lambda.invoke_arn
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "alb_zone_id" {
  description = "Canonical hosted zone ID of the Application Load Balancer"
  value       = module.alb.alb_zone_id
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = module.ecr.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  value       = module.ecr.repository_arn
}