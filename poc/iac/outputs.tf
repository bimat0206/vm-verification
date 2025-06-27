# S3 Bucket outputs
output "s3_bucket_names" {
  description = "Names of created S3 buckets"
  value = var.s3_buckets.create_buckets ? {
    reference = module.s3_buckets[0].reference_bucket_name
    checking  = module.s3_buckets[0].checking_bucket_name
    results   = module.s3_buckets[0].results_bucket_name
    state = module.s3_buckets[0].state_bucket_name
  } : {}
}

output "s3_bucket_arns" {
  description = "ARNs of created S3 buckets"
  value = var.s3_buckets.create_buckets ? {
    reference = module.s3_buckets[0].reference_bucket_arn
    checking  = module.s3_buckets[0].checking_bucket_arn
    results   = module.s3_buckets[0].results_bucket_arn
    state = module.s3_buckets[0].state_bucket_arn
  } : {}
}

# ECR Repository outputs
output "ecr_repository_urls" {
  description = "URLs of created ECR repositories"
  value       = var.ecr.create_repositories && var.lambda_functions.use_ecr ? module.ecr_repositories[0].repository_urls : {}
}

output "ecr_repository_arns" {
  description = "ARNs of created ECR repositories"
  value       = var.ecr.create_repositories && var.lambda_functions.use_ecr ? module.ecr_repositories[0].repository_arns : {}
}

# DynamoDB Table outputs
output "dynamodb_table_names" {
  description = "Names of created DynamoDB tables"
  value = var.dynamodb_tables.create_tables ? {
    verification_results = module.dynamodb_tables[0].verification_results_table_name
    layout_metadata      = module.dynamodb_tables[0].layout_metadata_table_name
    conversation_history = module.dynamodb_tables[0].conversation_history_table_name
  } : {}
}

output "dynamodb_table_arns" {
  description = "ARNs of created DynamoDB tables"
  value = var.dynamodb_tables.create_tables ? {
    verification_results = module.dynamodb_tables[0].verification_results_table_arn
    layout_metadata      = module.dynamodb_tables[0].layout_metadata_table_arn
    conversation_history = module.dynamodb_tables[0].conversation_history_table_arn
  } : {}
}

# Lambda Function outputs
output "lambda_function_names" {
  description = "Names of created Lambda functions"
  value       = var.lambda_functions.create_functions ? module.lambda_functions[0].function_names : {}
}

output "lambda_function_arns" {
  description = "ARNs of created Lambda functions"
  value       = var.lambda_functions.create_functions ? module.lambda_functions[0].function_arns : {}
}

# Step Functions outputs
output "step_functions_state_machine_name" {
  description = "Name of created Step Functions state machine"
  value       = var.step_functions.create_step_functions ? module.step_functions[0].state_machine_name : ""
}

output "step_functions_state_machine_arn" {
  description = "ARN of created Step Functions state machine"
  value       = var.step_functions.create_step_functions ? module.step_functions[0].state_machine_arn : ""
}

# API Gateway outputs
output "api_gateway_endpoint" {
  description = "Endpoint URL of created API Gateway"
  value       = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_endpoint : ""
}

output "api_gateway_id" {
  description = "ID of created API Gateway"
  value       = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_id : ""
}


# product-approach/iac/outputs.tf (append)

output "api_gateway_api_key" {
  description = "API key for the API Gateway"
  value       = var.api_gateway.create_api_gateway && var.api_gateway.use_api_key ? module.api_gateway[0].api_key_value : null
  sensitive   = true
}
# VPC outputs
output "vpc_id" {
  description = "ID of the created VPC"
  value       = var.vpc.create_vpc && var.streamlit_frontend.create_streamlit ? module.vpc[0].vpc_id : ""
}

output "vpc_public_subnet_ids" {
  description = "IDs of the public subnets in the VPC"
  value       = var.vpc.create_vpc && var.streamlit_frontend.create_streamlit ? module.vpc[0].public_subnet_ids : []
}

output "vpc_private_subnet_ids" {
  description = "IDs of the private subnets in the VPC"
  value       = var.vpc.create_vpc && var.streamlit_frontend.create_streamlit ? module.vpc[0].private_subnet_ids : []
}

# Streamlit Frontend outputs
output "streamlit_frontend_url" {
  description = "URL of created Streamlit frontend service"
  value       = var.streamlit_frontend.create_streamlit ? module.ecs_streamlit[0].service_url : ""
}

output "streamlit_frontend_https_url" {
  description = "HTTPS URL of created Streamlit frontend service (if HTTPS is enabled)"
  value       = var.streamlit_frontend.create_streamlit && var.streamlit_frontend.enable_https ? module.ecs_streamlit[0].service_https_url : ""
}

output "streamlit_frontend_alb_dns_name" {
  description = "DNS name of the ALB for Streamlit frontend"
  value       = var.streamlit_frontend.create_streamlit ? module.ecs_streamlit[0].alb_dns_name : ""
}

output "streamlit_frontend_ecs_cluster_id" {
  description = "ID of the ECS cluster for Streamlit frontend"
  value       = var.streamlit_frontend.create_streamlit ? module.ecs_streamlit[0].ecs_cluster_id : ""
}

output "streamlit_frontend_ecs_service_name" {
  description = "Name of the ECS service for Streamlit frontend"
  value       = var.streamlit_frontend.create_streamlit ? module.ecs_streamlit[0].ecs_service_name : ""
}

output "streamlit_frontend_log_group" {
  description = "CloudWatch log group for Streamlit frontend"
  value       = var.streamlit_frontend.create_streamlit ? module.ecs_streamlit[0].cloudwatch_log_group_name : ""
}
