# modules/api_gateway/main.tf

resource "aws_api_gateway_rest_api" "api" {
  name        = var.api_name
  description = var.api_description
  body = templatefile(var.openapi_definition, {
    cors_allowed_origins                 = var.streamlit_service_url != "" ? [startswith(var.streamlit_service_url, "http") ? var.streamlit_service_url : "https://${var.streamlit_service_url}"] : ["*"]
    verification_lookup_lambda_arn       = var.lambda_function_arns["fetch_historical_verification"]
    verification_initiate_lambda_arn     = var.lambda_function_arns["initialize"]
    verification_list_lambda_arn         = var.lambda_function_arns["fetch_historical_verification"]
    verification_get_lambda_arn          = var.lambda_function_arns["fetch_historical_verification"]
    verification_conversation_lambda_arn = var.lambda_function_arns["fetch_historical_verification"]
    health_lambda_arn                    = var.lambda_function_arns["initialize"]
    image_view_lambda_arn                = var.lambda_function_arns["fetch_images"]
    image_browser_lambda_arn             = var.lambda_function_arns["fetch_images"]
  })

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = var.common_tags
}

resource "aws_api_gateway_deployment" "deployment" {
  rest_api_id = aws_api_gateway_rest_api.api.id


  depends_on = [
    aws_api_gateway_rest_api.api
  ]
}

resource "aws_api_gateway_stage" "stage" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  deployment_id = aws_api_gateway_deployment.deployment.id
  stage_name    = var.stage_name

  cache_cluster_enabled = false
  xray_tracing_enabled  = var.metrics_enabled

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

resource "aws_cloudwatch_log_group" "api_log_group" {
  name              = "/aws/apigateway/${var.api_name}/${var.stage_name}"
  retention_in_days = 7
  tags              = var.common_tags
}

resource "aws_api_gateway_method_settings" "settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = var.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled        = var.metrics_enabled
    logging_level          = "INFO"
    throttling_rate_limit  = var.throttling_rate_limit
    throttling_burst_limit = var.throttling_burst_limit
  }
}

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
