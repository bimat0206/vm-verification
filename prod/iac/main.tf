# S3 Buckets
module "s3_buckets" {
  source = "./modules/s3"
  count  = var.s3_buckets.create_buckets ? 1 : 0

  reference_bucket_name = local.s3_buckets.reference
  checking_bucket_name  = local.s3_buckets.checking
  state_bucket_name     = local.s3_buckets.state

  reference_lifecycle_rules = var.s3_buckets.lifecycle_rules.reference
  checking_lifecycle_rules  = var.s3_buckets.lifecycle_rules.checking

  force_destroy = var.s3_buckets.force_destroy

  common_tags = local.common_tags
}

# DynamoDB Tables
module "dynamodb_tables" {
  source = "./modules/dynamodb"
  count  = var.dynamodb_tables.create_tables ? 1 : 0

  verification_results_table_name = local.dynamodb_tables.verification_results
  layout_metadata_table_name      = local.dynamodb_tables.layout_metadata
  conversation_history_table_name = local.dynamodb_tables.conversation_history

  billing_mode   = var.dynamodb_tables.billing_mode
  read_capacity  = var.dynamodb_tables.read_capacity
  write_capacity = var.dynamodb_tables.write_capacity

  point_in_time_recovery = var.dynamodb_tables.point_in_time_recovery

  common_tags = local.common_tags
}

# ECR Repositories
module "ecr_repositories" {
  source = "./modules/ecr"
  count  = var.ecr.create_repositories ? 1 : 0

  repositories = local.ecr_repositories

  common_tags = local.common_tags
}

# IAM Roles for Lambda Functions
module "lambda_iam" {
  source = "./modules/iam/lambda"
  count  = var.lambda_functions.create_functions ? 1 : 0

  project_name = var.project_name
  environment  = var.environment
  name_suffix  = local.name_suffix

  s3_bucket_arns = var.s3_buckets.create_buckets ? [
    module.s3_buckets[0].reference_bucket_arn,
    module.s3_buckets[0].checking_bucket_arn,
    module.s3_buckets[0].state_bucket_arn
  ] : []

  dynamodb_table_arns = var.dynamodb_tables.create_tables ? [
    module.dynamodb_tables[0].verification_results_table_arn,
    module.dynamodb_tables[0].layout_metadata_table_arn,
    module.dynamodb_tables[0].conversation_history_table_arn
  ] : []

  ecr_repository_arns = var.ecr.create_repositories ? [
    for repo_name, repo_url in module.ecr_repositories[0].repository_arns : repo_url
  ] : []

  bedrock_model_arn = "arn:aws:bedrock:*::foundation-model/${var.bedrock.model_id}"

  common_tags = local.common_tags
}

# Lambda Functions
module "lambda_functions" {
  source = "./modules/lambda"
  count  = var.lambda_functions.create_functions ? 1 : 0

  functions_config = local.lambda_functions

  execution_role_arn = module.lambda_iam[0].lambda_execution_role_arn

  use_ecr_repository = var.lambda_functions.use_ecr
  ecr_repository_url = var.lambda_functions.use_ecr && var.ecr.create_repositories ? local.ecr_repository_base_url : ""
  image_uri          = var.lambda_functions.default_image_uri
  default_image_uri  = var.lambda_functions.default_image_uri
  image_tag          = var.lambda_functions.image_tag

  architectures      = var.lambda_functions.architectures
  log_retention_days = var.lambda_functions.log_retention_days

  s3_trigger_functions = var.lambda_functions.s3_trigger_functions
  s3_source_arns = var.s3_buckets.create_buckets ? {
    for func_name in var.lambda_functions.s3_trigger_functions :
    func_name => module.s3_buckets[0].reference_bucket_arn
  } : null

  eventbridge_trigger_functions = var.lambda_functions.eventbridge_trigger_functions
  eventbridge_source_arns       = null

  common_tags = local.common_tags
}

# Step Functions State Machine

module "step_functions" {
  source = "./modules/step_functions"
  count  = var.step_functions.create_step_functions && var.lambda_functions.create_functions ? 1 : 0

  state_machine_name   = local.step_function_name
  log_level            = var.step_functions.log_level
  enable_x_ray_tracing = var.step_functions.enable_x_ray_tracing

  region = var.aws_region

  lambda_function_arns = {
    initialize                    = module.lambda_functions[0].function_arns["initialize"]
    fetch_historical_verification = module.lambda_functions[0].function_arns["fetch_historical_verification"]
    fetch_images                  = module.lambda_functions[0].function_arns["fetch_images"]
    prepare_system_prompt         = module.lambda_functions[0].function_arns["prepare_system_prompt"]
    execute_turn1_combined        = module.lambda_functions[0].function_arns["execute_turn1_combined"]
    execute_turn2_combined        = module.lambda_functions[0].function_arns["execute_turn2_combined"]
    finalize_results              = module.lambda_functions[0].function_arns["finalize_results"]
    finalize_with_error           = module.lambda_functions[0].function_arns["finalize_with_error"]
  }

  # Add DynamoDB table ARNs
  dynamodb_table_arns = var.dynamodb_tables.create_tables ? [
    module.dynamodb_tables[0].verification_results_table_arn,
    module.dynamodb_tables[0].layout_metadata_table_arn,
    module.dynamodb_tables[0].conversation_history_table_arn
  ] : []

  # Add DynamoDB table names
  dynamodb_verification_table = var.dynamodb_tables.create_tables ? module.dynamodb_tables[0].verification_results_table_name : ""
  dynamodb_conversation_table = var.dynamodb_tables.create_tables ? module.dynamodb_tables[0].conversation_history_table_name : ""

  common_tags = local.common_tags
}

# API Gateway
# The API Gateway module has been reorganized into smaller, more manageable files:
# - resources.tf: Contains all API Gateway resource definitions
# - methods.tf: Contains all API Gateway method and integration definitions
# - deployment.tf: Contains deployment, stage, and API key configurations
# - cors.tf: Contains CORS configuration for API endpoints
# - locals.tf: Contains local variable definitions
# - variables.tf: Contains input variable definitions
# - output.tf: Contains output definitions
module "api_gateway" {
  source = "./modules/api_gateway"
  count  = var.api_gateway.create_api_gateway && var.lambda_functions.create_functions ? 1 : 0

  api_name               = local.api_gateway_name
  api_description        = "Kootoro Vending Machine Verification API"
  stage_name             = var.api_gateway.stage_name
  throttling_rate_limit  = var.api_gateway.throttling_rate_limit
  throttling_burst_limit = var.api_gateway.throttling_burst_limit
  cors_enabled           = var.api_gateway.cors_enabled
  metrics_enabled        = var.api_gateway.metrics_enabled
  use_api_key            = var.api_gateway.use_api_key
  lambda_function_arns = {
    for k, v in module.lambda_functions[0].function_arns : k => v
  }

  lambda_function_names = {
    for k, v in module.lambda_functions[0].function_names : k => v
  }

  # Add Step Functions integration parameters
  step_functions_state_machine_arn = var.step_functions.create_step_functions ? module.step_functions[0].state_machine_arn : ""

  region = var.aws_region

  common_tags = local.common_tags
}

# Secrets Manager for API Key
module "secretsmanager" {
  source = "./modules/secretsmanager"
  count  = var.api_gateway.create_api_gateway && var.api_gateway.use_api_key ? 1 : 0

  project_name       = var.project_name
  environment        = var.environment
  name_suffix        = local.name_suffix
  secret_base_name   = "api-key"
  secret_description = "API key for Kootoro Vending Machine Verification API"
  secret_value       = module.api_gateway[0].api_key_value

  common_tags = local.common_tags
}


# Secrets Manager for ECS React Configuration
module "ecs_react_config_secret" {
  source = "./modules/secretsmanager"
  count  = var.react_frontend.create_react ? 1 : 0

  project_name       = var.project_name
  environment        = var.environment
  name_suffix        = local.name_suffix
  secret_base_name   = "ecs-react-config"
  secret_description = "Configuration settings for ECS React application"
  secret_value = jsonencode({
    REGION                      = var.aws_region
    DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
    DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
    API_ENDPOINT                = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_endpoint : ""
    REFERENCE_BUCKET            = local.s3_buckets.reference
    CHECKING_BUCKET             = local.s3_buckets.checking
  })

  common_tags = local.common_tags

  depends_on = [
    module.api_gateway,
    module.secretsmanager
  ]
}

# CloudWatch Monitoring Resources
module "monitoring" {
  source = "./modules/monitoring"
  count  = var.monitoring.create_dashboard ? 1 : 0

  dashboard_name = local.dashboard_name

  log_retention_days = var.monitoring.log_retention_days

  lambda_function_names = var.lambda_functions.create_functions ? module.lambda_functions[0].function_names : {}

  step_function_name              = var.step_functions.create_step_functions ? module.step_functions[0].state_machine_name : ""
  enable_step_function_monitoring = var.step_functions.create_step_functions

  api_gateway_name              = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_name : ""
  enable_api_gateway_monitoring = var.api_gateway.create_api_gateway

  dynamodb_table_names = var.dynamodb_tables.create_tables ? {
    verification_results = module.dynamodb_tables[0].verification_results_table_name
    layout_metadata      = module.dynamodb_tables[0].layout_metadata_table_name
    conversation_history = module.dynamodb_tables[0].conversation_history_table_name
  } : {}

  ecr_repository_names = var.ecr.create_repositories && var.lambda_functions.use_ecr ? module.ecr_repositories[0].repository_names : {}

  # ECS and ALB monitoring
  ecs_cluster_name      = ""
  ecs_service_name      = ""
  enable_ecs_monitoring = false

  alb_name              = ""
  enable_alb_monitoring = false


  common_tags = local.common_tags
}

# VPC for ALB and ECS
module "vpc" {
  source = "./modules/vpc"
  count  = var.vpc.create_vpc && var.react_frontend.create_react ? 1 : 0

  vpc_cidr           = var.vpc.vpc_cidr
  availability_zones = var.vpc.availability_zones
  create_nat_gateway = var.vpc.create_nat_gateway
  name_prefix        = local.vpc_name
  container_port     = 3000

  common_tags = local.common_tags
}


# React Frontend Service using ECS and ALB
module "ecs_react" {
  source = "./modules/ecs-react"
  count  = var.react_frontend.create_react ? 1 : 0

  service_name = var.react_frontend.service_name
  environment  = var.environment
  name_suffix  = local.name_suffix

  # Container configuration
  image_uri             = var.react_frontend.image_uri
  image_repository_type = var.react_frontend.image_repository_type
  cpu                   = var.react_frontend.cpu
  memory                = var.react_frontend.memory
  port                  = var.react_frontend.port

  # ECS configuration
  enable_container_insights = var.react_frontend.enable_container_insights
  enable_execute_command    = var.react_frontend.enable_execute_command
  min_capacity              = var.react_frontend.min_size
  max_capacity              = var.react_frontend.max_capacity
  enable_auto_scaling       = var.react_frontend.enable_auto_scaling
  cpu_threshold             = var.react_frontend.cpu_threshold
  memory_threshold          = var.react_frontend.memory_threshold
  auto_deployments_enabled  = var.react_frontend.auto_deployments_enabled

  # VPC and networking
  vpc_id                = module.vpc[0].vpc_id
  public_subnet_ids     = module.vpc[0].public_subnet_ids
  private_subnet_ids    = module.vpc[0].private_subnet_ids
  alb_security_group_id = module.vpc[0].alb_security_group_id
  ecs_security_group_id = module.vpc[0].ecs_react_security_group_id

  # ALB configuration
  internal_alb               = var.react_frontend.internal_alb
  enable_deletion_protection = false
  enable_https               = var.react_frontend.enable_https
  certificate_arn            = var.react_frontend.certificate_arn

  # Health check configuration
  health_check_path                = var.react_frontend.health_check_path
  health_check_interval            = var.react_frontend.health_check_interval
  health_check_timeout             = var.react_frontend.health_check_timeout
  health_check_healthy_threshold   = var.react_frontend.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.react_frontend.health_check_unhealthy_threshold

  # Logging
  log_retention_days = var.react_frontend.log_retention_days

  # Access permissions
  enable_api_gateway_access = true
  enable_s3_access          = true
  enable_dynamodb_access    = true
  enable_ecr_full_access    = true

  api_gateway_arn = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_arn : ""

  s3_bucket_arns = var.s3_buckets.create_buckets ? [
    module.s3_buckets[0].reference_bucket_arn,
    module.s3_buckets[0].checking_bucket_arn
  ] : []

  dynamodb_table_arns = var.dynamodb_tables.create_tables ? [
    module.dynamodb_tables[0].verification_results_table_arn,
    module.dynamodb_tables[0].layout_metadata_table_arn,
    module.dynamodb_tables[0].conversation_history_table_arn
  ] : []

  environment_variables = merge(
    var.react_frontend.environment_variables,
    {
      CONFIG_SECRET       = var.react_frontend.create_react ? module.ecs_react_config_secret[0].secret_name : ""
      API_KEY_SECRET_NAME = var.api_gateway.create_api_gateway && var.api_gateway.use_api_key ? module.secretsmanager[0].secret_name : ""
    }
  )

  common_tags = local.common_tags

  depends_on = [
    module.api_gateway,
    module.vpc,
    module.secretsmanager,
    module.ecs_react_config_secret
  ]
}
