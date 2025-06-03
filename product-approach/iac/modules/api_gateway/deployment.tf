# modules/api_gateway/deployment.tf

# API Gateway Deployment
resource "aws_api_gateway_deployment" "deployment" {
  rest_api_id = aws_api_gateway_rest_api.api.id

  # Force a new deployment when any of the methods or integrations change
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_method.verifications_lookup_get.id,
      aws_api_gateway_integration.verifications_lookup_get.id,
      aws_api_gateway_method_response.verifications_lookup_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.verifications_lookup_get[0].id : "",

      aws_api_gateway_method.verifications_lookup_options.id,
      aws_api_gateway_integration.verifications_lookup_options.id,
      aws_api_gateway_method_response.verifications_lookup_options.id,
      aws_api_gateway_integration_response.verifications_lookup_options.id,

      aws_api_gateway_method.verifications_post.id,
      aws_api_gateway_integration.verifications_post.id,
      aws_api_gateway_method_response.verifications_post.id,
      var.cors_enabled ? aws_api_gateway_integration_response.verifications_post[0].id : "",

      aws_api_gateway_method.verifications_options.id,
      aws_api_gateway_integration.verifications_options.id,
      aws_api_gateway_method_response.verifications_options.id,
      aws_api_gateway_integration_response.verifications_options.id,

      aws_api_gateway_method.verifications_get.id,
      aws_api_gateway_integration.verifications_get.id,
      aws_api_gateway_method_response.verifications_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.verifications_get[0].id : "",

      aws_api_gateway_method.verification_id_get.id,
      aws_api_gateway_integration.verification_id_get.id,
      aws_api_gateway_method_response.verification_id_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.verification_id_get[0].id : "",

      aws_api_gateway_method.verification_id_options.id,
      aws_api_gateway_integration.verification_id_options.id,
      aws_api_gateway_method_response.verification_id_options.id,
      aws_api_gateway_integration_response.verification_id_options.id,

      aws_api_gateway_method.verification_conversation_get.id,
      aws_api_gateway_integration.verification_conversation_get.id,
      aws_api_gateway_method_response.verification_conversation_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.verification_conversation_get[0].id : "",

      aws_api_gateway_method.verification_conversation_options.id,
      aws_api_gateway_integration.verification_conversation_options.id,
      aws_api_gateway_method_response.verification_conversation_options.id,
      aws_api_gateway_integration_response.verification_conversation_options.id,

      aws_api_gateway_method.health_get.id,
      aws_api_gateway_integration.health_get.id,
      aws_api_gateway_method_response.health_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.health_get[0].id : "",

      aws_api_gateway_method.health_options.id,
      aws_api_gateway_integration.health_options.id,
      aws_api_gateway_method_response.health_options.id,
      aws_api_gateway_integration_response.health_options.id,

      aws_api_gateway_method.image_view_get.id,
      aws_api_gateway_integration.image_view_get.id,
      aws_api_gateway_method_response.image_view_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.image_view_get[0].id : "",

      aws_api_gateway_method.image_view_options.id,
      aws_api_gateway_integration.image_view_options.id,
      aws_api_gateway_method_response.image_view_options.id,
      aws_api_gateway_integration_response.image_view_options.id,

      aws_api_gateway_method.image_browser_get.id,
      aws_api_gateway_integration.image_browser_get.id,
      aws_api_gateway_method_response.image_browser_get.id,
      var.cors_enabled ? aws_api_gateway_integration_response.image_browser_get[0].id : "",

      aws_api_gateway_method.image_browser_options.id,
      aws_api_gateway_integration.image_browser_options.id,
      aws_api_gateway_method_response.image_browser_options.id,
      aws_api_gateway_integration_response.image_browser_options.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    # GET method integrations
    aws_api_gateway_integration.verifications_lookup_get,
    aws_api_gateway_integration.verifications_post,
    aws_api_gateway_integration.verifications_get,
    aws_api_gateway_integration.verification_id_get,
    aws_api_gateway_integration.verification_conversation_get,
    aws_api_gateway_integration.health_get,
    aws_api_gateway_integration.image_view_get,
    aws_api_gateway_integration.image_browser_get,

    # Method responses
    aws_api_gateway_method_response.verifications_lookup_get,
    aws_api_gateway_method_response.verifications_post,
    aws_api_gateway_method_response.verifications_get,
    aws_api_gateway_method_response.verification_id_get,
    aws_api_gateway_method_response.verification_conversation_get,
    aws_api_gateway_method_response.health_get,
    aws_api_gateway_method_response.image_view_get,
    aws_api_gateway_method_response.image_browser_get,

    # Integration responses for CORS
    aws_api_gateway_integration_response.verifications_lookup_get,
    aws_api_gateway_integration_response.verifications_post,
    aws_api_gateway_integration_response.verifications_get,
    aws_api_gateway_integration_response.verification_id_get,
    aws_api_gateway_integration_response.verification_conversation_get,
    aws_api_gateway_integration_response.health_get,
    aws_api_gateway_integration_response.image_view_get,
    aws_api_gateway_integration_response.image_browser_get,

    # OPTIONS method integrations and responses
    aws_api_gateway_integration.verifications_lookup_options,
    aws_api_gateway_method_response.verifications_lookup_options,
    aws_api_gateway_integration_response.verifications_lookup_options,

    aws_api_gateway_integration.verifications_options,
    aws_api_gateway_method_response.verifications_options,
    aws_api_gateway_integration_response.verifications_options,

    aws_api_gateway_integration.verification_id_options,
    aws_api_gateway_method_response.verification_id_options,
    aws_api_gateway_integration_response.verification_id_options,

    aws_api_gateway_integration.verification_conversation_options,
    aws_api_gateway_method_response.verification_conversation_options,
    aws_api_gateway_integration_response.verification_conversation_options,

    aws_api_gateway_integration.health_options,
    aws_api_gateway_method_response.health_options,
    aws_api_gateway_integration_response.health_options,

    aws_api_gateway_integration.image_view_options,
    aws_api_gateway_method_response.image_view_options,
    aws_api_gateway_integration_response.image_view_options,

    aws_api_gateway_integration.image_browser_options,
    aws_api_gateway_method_response.image_browser_options,
    aws_api_gateway_integration_response.image_browser_options
  ]
}

# API Gateway Stage
resource "aws_api_gateway_stage" "stage" {
  deployment_id = aws_api_gateway_deployment.deployment.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = var.stage_name

  variables = {
    verification_lookup_lambda       = var.lambda_function_names["fetch_historical_verification"]
    verification_initiate_lambda     = var.lambda_function_names["initialize"]
    verification_list_lambda         = var.lambda_function_names["api_verifications_list"]
    verification_get_lambda          = var.lambda_function_names["api_verifications_list"]
    verification_conversation_lambda = var.lambda_function_names["get_conversation"]
    health_lambda                    = var.lambda_function_names["health_check"]
    image_view_lambda                = var.lambda_function_names["api_images_view"]
    image_browser_lambda             = var.lambda_function_names["api_images_browser"]
  }

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_log_group.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      caller         = "$context.identity.caller"
      user           = "$context.identity.user"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      resourcePath   = "$context.resourcePath"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }

  tags = var.common_tags
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_log_group" {
  name              = "/aws/apigateway/${var.api_name}/${var.stage_name}"
  retention_in_days = 7
  tags              = var.common_tags
}

# API Gateway Method Settings
resource "aws_api_gateway_method_settings" "settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.stage.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled        = var.metrics_enabled
    logging_level          = "INFO"
    throttling_rate_limit  = var.throttling_rate_limit
    throttling_burst_limit = var.throttling_burst_limit
  }

  depends_on = [
    aws_api_gateway_stage.stage
  ]
}

# API Key Configuration
resource "aws_api_gateway_api_key" "api_key" {
  count       = var.use_api_key ? 1 : 0
  name        = "${var.api_name}-key"
  description = "API key for ${var.api_name}"
  enabled     = true
  tags        = var.common_tags
}

resource "aws_api_gateway_usage_plan" "usage_plan" {
  count       = var.use_api_key ? 1 : 0
  name        = "${var.api_name}-usage-plan"
  description = "Usage plan for ${var.api_name}"

  api_stages {
    api_id = aws_api_gateway_rest_api.api.id
    stage  = aws_api_gateway_stage.stage.stage_name
  }

  throttle_settings {
    rate_limit  = var.throttling_rate_limit
    burst_limit = var.throttling_burst_limit
  }

  tags = var.common_tags
}

resource "aws_api_gateway_usage_plan_key" "usage_plan_key" {
  count         = var.use_api_key ? 1 : 0
  key_id        = aws_api_gateway_api_key.api_key[0].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.usage_plan[0].id
}
