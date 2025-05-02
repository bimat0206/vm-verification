# S3 Buckets
module "s3_buckets" {
  source = "./modules/s3"
  count  = var.s3_buckets.create_buckets ? 1 : 0

  reference_bucket_name = local.s3_buckets.reference
  checking_bucket_name  = local.s3_buckets.checking
  results_bucket_name   = local.s3_buckets.results

  reference_lifecycle_rules = var.s3_buckets.lifecycle_rules.reference
  checking_lifecycle_rules  = var.s3_buckets.lifecycle_rules.checking
  results_lifecycle_rules   = var.s3_buckets.lifecycle_rules.results

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
  count  = var.ecr.create_repositories && var.lambda_functions.use_ecr ? 1 : 0

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
    module.s3_buckets[0].results_bucket_arn
  ] : []

  dynamodb_table_arns = var.dynamodb_tables.create_tables ? [
    module.dynamodb_tables[0].verification_results_table_arn,
    module.dynamodb_tables[0].layout_metadata_table_arn,
    module.dynamodb_tables[0].conversation_history_table_arn
  ] : []

  ecr_repository_arns = var.ecr.create_repositories && var.lambda_functions.use_ecr ? [
    for repo_name, repo_url in module.ecr_repositories[0].repository_arns : repo_url
  ] : []

  bedrock_model_arn = "arn:aws:bedrock:${var.aws_region}::foundation-model/${var.bedrock.model_id}"

  common_tags = local.common_tags
}

# Lambda Functions
module "lambda_functions" {
  source = "./modules/lambda"
  count  = var.lambda_functions.create_functions ? 1 : 0

  functions_config = local.lambda_functions

  execution_role_arn = module.lambda_iam[0].lambda_execution_role_arn

  use_ecr_repository = var.lambda_functions.use_ecr
  # Use the actual ECR repository URLs from the ECR module output
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
  eventbridge_source_arns       = null # To be populated if EventBridge rules are created

  common_tags = local.common_tags
}

# Step Functions State Machine
module "step_functions" {
  source = "./modules/step_functions"
  count  = var.step_functions.create_step_functions && var.lambda_functions.create_functions ? 1 : 0

  state_machine_name = local.step_function_name


  log_level = var.step_functions.log_level

  lambda_function_arns = {
    initialize                    = module.lambda_functions[0].function_arns["initialize"]
    fetch_historical_verification = module.lambda_functions[0].function_arns["fetch_historical_verification"]
    fetch_images                  = module.lambda_functions[0].function_arns["fetch_images"]
    prepare_system_prompt         = module.lambda_functions[0].function_arns["prepare_system_prompt"]
    prepare_turn_prompt           = module.lambda_functions[0].function_arns["prepare_turn_prompt"]
    invoke_bedrock                = module.lambda_functions[0].function_arns["invoke_bedrock"]
    process_turn1_response        = module.lambda_functions[0].function_arns["process_turn1_response"]
    process_turn2_response        = module.lambda_functions[0].function_arns["process_turn2_response"]
    finalize_results              = module.lambda_functions[0].function_arns["finalize_results"]
    store_results                 = module.lambda_functions[0].function_arns["store_results"]
    notify                        = module.lambda_functions[0].function_arns["notify"]
    handle_bedrock_error          = module.lambda_functions[0].function_arns["handle_bedrock_error"]
    finalize_with_error           = module.lambda_functions[0].function_arns["finalize_with_error"]
  }

  common_tags = local.common_tags
}

# API Gateway
module "api_gateway" {
  source = "./modules/api_gateway"
  count  = var.api_gateway.create_api_gateway && var.lambda_functions.create_functions ? 1 : 0

  api_name = local.api_gateway_name

  api_description = "Kootoro Vending Machine Verification API"

  stage_name = var.api_gateway.stage_name

  throttling_rate_limit  = var.api_gateway.throttling_rate_limit
  throttling_burst_limit = var.api_gateway.throttling_burst_limit

  cors_enabled    = var.api_gateway.cors_enabled
  metrics_enabled = var.api_gateway.metrics_enabled

  lambda_function_arns = {
    initialize = module.lambda_functions[0].function_arns["initialize"]
  }

  lambda_function_names = {
    initialize = module.lambda_functions[0].function_names["initialize"]
  }

  common_tags = local.common_tags
}

# App Runner service for frontend
module "app_runner" {
  source = "./modules/app_runner"
  count  = var.app_runner.create_app_runner ? 1 : 0

  service_name = local.app_runner_service_name

  image_uri       = var.app_runner.image_uri
  image_repo_type = "ECR_PUBLIC" # Explicitly set to ECR_PUBLIC for public images

  cpu    = var.app_runner.cpu
  memory = var.app_runner.memory

  auto_deployments_enabled = var.app_runner.auto_deployments_enabled

  environment_variables = merge(
    var.app_runner.environment_variables,
    {
      API_ENDPOINT = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_endpoint : ""
    }
  )

  common_tags = local.common_tags
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

  app_runner_service_name      = var.app_runner.create_app_runner ? module.app_runner[0].service_name : ""
  enable_app_runner_monitoring = var.app_runner.create_app_runner

  alarm_email_endpoints = var.monitoring.alarm_email_endpoints

  common_tags = local.common_tags
}
