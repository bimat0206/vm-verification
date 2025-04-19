

# Merge default and additional tags with environment-specific overrides
locals {
  common_tags = merge(
    var.additional_tags,
    {
      Environment = var.environment
    }
  )
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  public_subnet_cidrs = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  enable_nat_gateway = var.enable_nat_gateway
  single_nat_gateway = var.single_nat_gateway
  tags               = local.common_tags
}

# S3 bucket for storing images
module "images_bucket" {
  source = "./modules/s3"

  bucket_name     = var.s3_bucket_name
  environment     = var.environment
  enable_versioning = true
  encryption_algorithm = "AES256"
  tags            = local.common_tags
}

# DynamoDB table for verification results
module "verification_results" {
  source = "./modules/dynamodb"

  table_name     = var.dynamodb_table_name
  environment    = var.environment
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "comparisonId"
  range_key      = "timestamp"
  enable_streams = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  attributes = [
    {
      name = "comparisonId"
      type = "S"
    },
    {
      name = "timestamp"
      type = "S"
    },
    {
      name = "vendingMachineId"
      type = "S"
    }
  ]
  global_secondary_indexes = [
    {
      name               = "VendingMachineIndex"
      hash_key           = "vendingMachineId"
      range_key          = "timestamp"
      projection_type    = "ALL"
      read_capacity      = null
      write_capacity     = null
    }
  ]
  tags            = local.common_tags
}

# Secrets Manager for storing API keys and credentials
module "secrets" {
  source = "./modules/secrets_manager"

  secret_name     = "vending-verification-bedrock-api-key"
  environment     = var.environment
  description     = "API key for Bedrock Claude model"
  create_bedrock_policy = true
  aws_region      = var.aws_region
  create_secret_version = true
  secret_string_map = {
    "api_key"     = var.bedrock_api_key
    "model_id"    = "anthropic.claude-3-sonnet-20240229-v1:0"
  }
  tags            = local.common_tags
}

# ECR Repository for Docker images


# Lambda functions for the workflow


# Step Functions state machine
# Step Functions state machine
module "step_functions" {
  source = "./modules/step_functions"

  state_machine_name = "vending-verification-workflow"
  environment        = var.environment
  
  initialize_function_arn      = module.lambda_functions.initialize_function_arn
  fetch_images_function_arn    = module.lambda_functions.fetch_images_function_arn
  prepare_prompt_function_arn  = module.lambda_functions.prepare_prompt_function_arn
  invoke_bedrock_function_arn  = module.lambda_functions.invoke_bedrock_function_arn
  process_results_function_arn = module.lambda_functions.process_results_function_arn
  store_results_function_arn   = module.lambda_functions.store_results_function_arn
  notify_function_arn          = module.lambda_functions.notify_function_arn
  
  # Add log retention days
  log_retention_days = var.cloudwatch_logs_retention_days
  
  tags            = local.common_tags
}

# API Gateway
# API Gateway
module "api_gateway" {
  source = "./modules/api_gateway"

  api_name        = "vending-verification-api"
  environment     = var.environment
  stage_name      = "v1"
  
  step_functions_state_machine_arn = module.step_functions.state_machine_arn
  step_functions_invoke_arn        = "arn:aws:apigateway:${var.aws_region}:states:action/StartExecution"
  
  get_comparison_lambda_function_name = module.lambda_functions.get_comparison_function_name
  get_comparison_lambda_invoke_arn    = module.lambda_functions.get_comparison_invoke_arn
  
  get_images_lambda_function_name     = module.lambda_functions.get_images_function_name
  get_images_lambda_invoke_arn        = module.lambda_functions.get_images_invoke_arn
  
  # Add this line to skip the integration response
  skip_api_gateway_integration_response = true
  
  tags            = local.common_tags
}

# CloudWatch monitoring
# Updated CloudWatch monitoring module
module "monitoring" {
  source = "./modules/monitoring"

  dashboard_name       = "vending-verification-dashboard"
  environment          = var.environment
  aws_region           = var.aws_region
  
  lambda_functions     = [
    module.lambda_functions.initialize_function_name,
    module.lambda_functions.fetch_images_function_name,
    module.lambda_functions.prepare_prompt_function_name,
    module.lambda_functions.invoke_bedrock_function_name,
    module.lambda_functions.process_results_function_name,
    module.lambda_functions.store_results_function_name,
    module.lambda_functions.notify_function_name,
    module.lambda_functions.get_comparison_function_name,
    module.lambda_functions.get_images_function_name
  ]
  
  state_machine_arn     = module.step_functions.state_machine_arn
  state_machine_name    = module.step_functions.state_machine_name
  dynamodb_table        = module.verification_results.table_name
  api_gateway_api_name  = "vending-verification-api"
  api_gateway_stage_name = "v1"
  s3_bucket_name        = module.images_bucket.bucket_id
  
  lambda_error_threshold = 1
  api_gateway_error_threshold = 1
  step_functions_failure_threshold = 1
  
  log_retention_days = var.cloudwatch_logs_retention_days
  
  # Set to false since the Step Functions module already creates this log group
  create_state_machine_log_group = false
  
  tags               = local.common_tags
}
module "ecr" {
  source = "./modules/ecr"

  repository_prefix  = var.ecr_repository_name
  environment        = var.environment
  aws_region         = var.aws_region
  image_tag_mutability = var.ecr_image_tag_mutability
  enable_scan_on_push = var.ecr_enable_scan_on_push
  kms_key_arn        = var.ecr_kms_key_arn
  max_image_count    = var.ecr_max_image_count
  
  # Set to true to automatically push placeholder nginx images
  push_placeholder_images = var.push_placeholder_images
  
  tags               = local.common_tags
}

# Lambda functions for the workflow
# Update to main.tf (lambda_functions module section)

# Lambda functions for the workflow
# Lambda functions for the workflow
# Lambda functions for the workflow
module "lambda_functions" {
  source = "./modules/multi_lambda"

  name_prefix      = "vending-verification"
  environment      = var.environment
  architectures    = ["arm64"]
  
  s3_bucket_name   = module.images_bucket.bucket_id
  s3_bucket_arn    = module.images_bucket.bucket_arn
  
  dynamodb_table_name = module.verification_results.table_name
  dynamodb_table_arn  = module.verification_results.table_arn
  
  aws_region       = var.aws_region
  secrets_arn      = module.secrets.secret_arn
  enable_secrets_access = true
  
  bedrock_model_id = "anthropic.claude-3-7-sonnet-20250219-v1:0"
  
  # Use the correct variable name
  skip_lambda_function_creation = var.skip_lambda_functions
  
  # Pass ECR repository URLs to the Lambda module
  ecr_repository_urls = {
    initialize      = module.ecr.initialize_repository_url
    "fetch-images"  = module.ecr.fetch_images_repository_url
    "prepare-prompt" = module.ecr.prepare_prompt_repository_url
    "invoke-bedrock" = module.ecr.invoke_bedrock_repository_url
    "process-results" = module.ecr.process_results_repository_url
    "store-results" = module.ecr.store_results_repository_url
    notify          = module.ecr.notify_repository_url
    "get-comparison" = module.ecr.get_comparison_repository_url
    "get-images"    = module.ecr.get_images_repository_url
  }
  
  environment_variables = {
    SECRETS_ARN    = module.secrets.secret_arn
    API_STAGE      = "v1"
  }
  
  tags            = local.common_tags
  
  depends_on = [
    module.ecr.repository_urls
  ]
}

# Add this to your infrastructure/main.tf file

# Streamlit Frontend
module "streamlit_frontend" {
  source = "./modules/streamlit_frontend"

  name_prefix      = "vending-verification"
  environment      = var.environment
  aws_region       = var.aws_region
  
  # ECR Configuration
  image_tag_mutability = var.ecr_image_tag_mutability
  enable_scan_on_push  = var.ecr_enable_scan_on_push
  kms_key_arn          = var.ecr_kms_key_arn
  max_image_count      = var.ecr_max_image_count
  
  # API Configuration
  api_endpoint        = module.api_gateway.invoke_url # Replace with module.api_gateway.invoke_url
  dynamodb_table_name = module.verification_results.table_name
  s3_bucket_name      = module.images_bucket.bucket_id
  step_functions_arn  = module.step_functions.state_machine_arn
  
  # App Runner Configuration
  container_port          = 8501
  image_tag               = "latest"
  auto_deployments_enabled = true
  cpu                     = "1024"  # 1 vCPU
  memory                  = "2048"  # 2 GB
  max_concurrency         = 50
  max_size                = 2
  min_size                = 1
  
  # Docker Build Configuration (optional)
  build_and_push_image   = false  # Set to true to build and push during terraform apply
  app_source_path        = "${path.module}/../frontend"
  
  # Additional configuration (optional)
  additional_config = {
    APP_TITLE          = "Vending Machine Verification"
    ENABLE_ANALYTICS   = "false"
    LOG_LEVEL          = "INFO"
  }
  
  tags = local.common_tags
}