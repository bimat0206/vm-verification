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
    module.s3_buckets[0].results_bucket_arn
  ] : []

  dynamodb_table_arns = var.dynamodb_tables.create_tables ? [
    module.dynamodb_tables[0].verification_results_table_arn,
    module.dynamodb_tables[0].layout_metadata_table_arn,
    module.dynamodb_tables[0].conversation_history_table_arn
  ] : []

  ecr_repository_arns = var.ecr.create_repositories ? [
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
# Module update in main.tf to enable API Gateway integration

module "step_functions" {
  source = "./modules/step_functions"
  count  = var.step_functions.create_step_functions && var.lambda_functions.create_functions ? 1 : 0

  state_machine_name   = local.step_function_name
  log_level            = var.step_functions.log_level
  enable_x_ray_tracing = var.step_functions.enable_x_ray_tracing

  # Enable API Gateway integration
  create_api_gateway_integration = var.api_gateway.create_api_gateway
  api_gateway_id                 = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_id : ""
  api_gateway_root_resource_id   = var.api_gateway.create_api_gateway ? module.api_gateway[0].root_resource_id : ""
  region                         = var.aws_region

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
    list_verifications            = module.lambda_functions[0].function_arns["list_verifications"]
    get_verification              = module.lambda_functions[0].function_arns["get_verification"]
    get_conversation              = module.lambda_functions[0].function_arns["get_conversation"]
    health_check                  = module.lambda_functions[0].function_arns["health_check"]

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
  openapi_definition     = "${path.module}/openapi.yaml"
  streamlit_service_url  = var.streamlit_frontend.create_streamlit ? module.streamlit_frontend[0].service_url : ""
  lambda_function_arns = {
    for k, v in module.lambda_functions[0].function_arns : k => v
  }

  lambda_function_names = {
    for k, v in module.lambda_functions[0].function_names : k => v
  }

  region = var.aws_region

  common_tags = local.common_tags
}

# Secrets Manager for API Key
# Secrets Manager for API Key
module "secretsmanager" {
  source = "./modules/secretsmanager"
  count  = var.api_gateway.create_api_gateway && var.api_gateway.use_api_key ? 1 : 0

  project_name       = var.project_name
  environment        = var.environment
  name_suffix        = local.name_suffix
  secret_base_name   = "api-key" # Replace "kootoro/api-key" with just "api-key"
  secret_description = "API key for Kootoro Vending Machine Verification API"
  secret_value       = module.api_gateway[0].api_key_value

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

  alarm_email_endpoints = var.monitoring.alarm_email_endpoints

  common_tags = local.common_tags
}

# Streamlit Frontend Service
module "streamlit_frontend" {
  source = "./modules/streamlit-frontend"
  count  = var.streamlit_frontend.create_streamlit ? 1 : 0

  service_name = var.streamlit_frontend.service_name
  environment  = var.environment
  name_suffix  = local.name_suffix

  image_uri                = var.streamlit_frontend.image_uri
  image_repository_type    = var.streamlit_frontend.image_repository_type
  cpu                      = var.streamlit_frontend.cpu
  memory                   = var.streamlit_frontend.memory
  port                     = var.streamlit_frontend.port
  auto_deployments_enabled = var.streamlit_frontend.auto_deployments_enabled

  health_check_path                = var.streamlit_frontend.health_check_path
  health_check_interval            = 15
  health_check_timeout             = 5
  health_check_healthy_threshold   = var.streamlit_frontend.health_check_healthy_threshold
  health_check_unhealthy_threshold = 5

  enable_auto_scaling = var.streamlit_frontend.enable_auto_scaling
  min_size            = var.streamlit_frontend.min_size
  max_size            = var.streamlit_frontend.max_size

  theme_mode         = var.streamlit_frontend.theme_mode
  log_retention_days = var.streamlit_frontend.log_retention_days

  # Access permissions
  enable_api_gateway_access = true
  enable_s3_access          = true
  enable_dynamodb_access    = true

  api_gateway_arn = var.api_gateway.create_api_gateway ? module.api_gateway[0].api_arn : ""

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


  environment_variables = merge(
    var.streamlit_frontend.environment_variables,
    {
      REGION              = var.aws_region
      DYNAMODB_TABLE      = local.dynamodb_tables.verification_results
      S3_BUCKET           = local.s3_buckets.reference
      AWS_DEFAULT_REGION  = var.aws_region
      API_KEY_SECRET_NAME = module.secretsmanager[0].secret_name # Use the output instead of hardcoded value
    }
  )

  enable_ecr_full_access = false

  common_tags = local.common_tags
}

# This resource will update the Streamlit environment variables after both resources are created
# This resource is commented out as it's causing errors and appears to be redundant with the API Gateway module
# The API Gateway module already creates a stage with the necessary configuration
# If additional configuration is needed, it should be added to the API Gateway module
/*
resource "aws_api_gateway_stage" "verification_api" {
  deployment_id = aws_api_gateway_deployment.verification_api.id
  rest_api_id  = aws_api_gateway_rest_api.verification_api.id
  stage_name   = var.environment
  
  variables = {
    verification_lookup_lambda = aws_lambda_function.verification_lookup.arn
    verification_initiate_lambda = aws_lambda_function.verification_initiate.arn
    verification_list_lambda = aws_lambda_function.verification_list.arn
    verification_get_lambda = aws_lambda_function.verification_get.arn
    verification_conversation_lambda = aws_lambda_function.verification_conversation.arn
    health_lambda = aws_lambda_function.health.arn
    image_view_lambda = aws_lambda_function.image_view.arn
    image_browser_lambda = aws_lambda_function.image_browser.arn
  }
  
  depends_on = [
    module.api_gateway,
    module.streamlit_frontend
  ]
}
*/

# Use a null_resource to update the Streamlit environment variables after deployment
# Use a null_resource to update the Streamlit environment variables after deployment
# Use a null_resource to update the Streamlit environment variables after deployment
# Use a null_resource to update the Streamlit environment variables after deployment
resource "null_resource" "update_streamlit_env" {
  count = var.streamlit_frontend.create_streamlit && var.api_gateway.create_api_gateway ? 1 : 0

  # Use local-exec to update the Streamlit environment variables
  provisioner "local-exec" {
    command = <<EOT
      # Create a JSON file to use with the AWS CLI command
      cat > update_env.json <<EOF
{
  "SourceConfiguration": {
    "ImageRepository": {
      "ImageConfiguration": {
        "RuntimeEnvironmentVariables": {
          "API_ENDPOINT": "${module.api_gateway[0].invoke_url}"
        }
      }
    }
  }
}
EOF
      # Use AWS CLI to update the service with the JSON file
      aws apprunner update-service \
        --service-arn ${module.streamlit_frontend[0].service_arn} \
        --source-configuration-update file://update_env.json
EOT
  }

  # Add a trigger to run this whenever the API Gateway URL changes
  triggers = {
    api_gateway_url = var.api_gateway.create_api_gateway ? module.api_gateway[0].invoke_url : "none"
  }

  depends_on = [
    module.api_gateway,
    module.streamlit_frontend
  ]
}
