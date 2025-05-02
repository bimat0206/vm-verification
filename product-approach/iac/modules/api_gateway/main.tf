# product-approach/iac/modules/api_gateway/main.tf

# Create REST API Gateway using OpenAPI specification
# product-approach/iac/modules/api_gateway/main.tf

# Create REST API Gateway using OpenAPI specification
resource "aws_api_gateway_rest_api" "this" {
  name        = var.api_name
  description = var.api_description

  body = templatefile(var.openapi_definition, {
    verification_lookup_lambda_arn       = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "fetch_historical_verification", "")}/invocations"
    verification_initiate_lambda_arn     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    verification_list_lambda_arn         = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    verification_get_lambda_arn          = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    verification_conversation_lambda_arn = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    health_lambda_arn                    = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    image_view_lambda_arn                = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    image_browser_lambda_arn             = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${lookup(var.lambda_function_arns, "initialize", "")}/invocations"
    # Format CORS allowed origins as a JSON array string for proper OpenAPI parsing
    cors_allowed_origins = jsonencode(var.cors_allowed_origins)
  })

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.api_name
    }
  )
}

# Create API Gateway deployment
resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.this.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Create API Gateway stage
resource "aws_api_gateway_stage" "this" {
  deployment_id = aws_api_gateway_deployment.this.id
  rest_api_id   = aws_api_gateway_rest_api.this.id
  stage_name    = var.stage_name

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
    })
  }

  variables = {
    "throttlingRateLimit"  = var.throttling_rate_limit
    "throttlingBurstLimit" = var.throttling_burst_limit
    "metricsEnabled"       = var.metrics_enabled ? "true" : "false"
  }
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.api_name}"
  retention_in_days = var.log_retention_days
}

# API Key (optional)
resource "aws_api_gateway_api_key" "this" {
  count       = var.use_api_key ? 1 : 0
  name        = "${var.api_name}-key"
  enabled     = true
  description = "API key for ${var.api_name}"
}

# Create usage plan if API key is used
resource "aws_api_gateway_usage_plan" "this" {
  count = var.use_api_key ? 1 : 0

  name        = "${var.api_name}-usage-plan"
  description = "Usage plan for ${var.api_name}"

  api_stages {
    api_id = aws_api_gateway_rest_api.this.id
    stage  = aws_api_gateway_stage.this.stage_name
  }

  quota_settings {
    limit  = var.api_quota_limit
    period = var.api_quota_period
  }

  throttle_settings {
    burst_limit = var.throttling_burst_limit
    rate_limit  = var.throttling_rate_limit
  }
}

# Associate API key with usage plan
resource "aws_api_gateway_usage_plan_key" "this" {
  count = var.use_api_key ? 1 : 0

  key_id        = aws_api_gateway_api_key.this[0].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.this[0].id
}

# Permission for API Gateway to invoke Lambda
resource "aws_lambda_permission" "api_gateway" {
  for_each = var.lambda_function_names

  statement_id  = "AllowExecutionFromAPIGateway-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = each.value
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/*/*"
}
