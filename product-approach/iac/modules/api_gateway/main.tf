# Create API Gateway v2 HTTP API
resource "aws_apigatewayv2_api" "this" {
  name          = var.api_name
  description   = var.api_description
  protocol_type = "HTTP"
  
  cors_configuration {
    allow_headers = var.cors_enabled ? ["Content-Type", "X-Amz-Date", "Authorization", "X-Api-Key", "X-Amz-Security-Token", "X-Amz-User-Agent"] : []
    allow_methods = var.cors_enabled ? ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"] : []
    allow_origins = var.cors_enabled ? var.cors_allowed_origins : []
    allow_credentials = var.cors_enabled
    max_age = var.cors_enabled ? 7200 : null
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.api_name
    }
  )
}

# Create API Gateway stage
resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = var.stage_name
  auto_deploy = true

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
      }
    )
  }

  default_route_settings {
    throttling_burst_limit = var.throttling_burst_limit
    throttling_rate_limit  = var.throttling_rate_limit
    detailed_metrics_enabled = var.metrics_enabled
  }

  tags = var.common_tags
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigatewayv2/${var.api_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# API Key (optional)
resource "aws_apigatewayv2_api_key" "this" {
  count = var.use_api_key ? 1 : 0
  name  = "${var.api_name}-key"
}

# Lambda integration for initialize function
resource "aws_apigatewayv2_integration" "initialize" {
  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = "AWS_PROXY"
  integration_uri        = lookup(var.lambda_function_arns, "initialize", null)
  integration_method     = "POST"
  payload_format_version = "2.0"
  timeout_milliseconds   = 30000
}

# Add routes for verification endpoints
resource "aws_apigatewayv2_route" "verifications_post" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "POST /api/${var.stage_name}/verifications"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

resource "aws_apigatewayv2_route" "verifications_get" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/verifications"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

resource "aws_apigatewayv2_route" "verification_lookup" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/verifications/lookup"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

resource "aws_apigatewayv2_route" "verification_by_id" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/verifications/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

resource "aws_apigatewayv2_route" "verification_conversation" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/verifications/{id}/conversation"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

# Health check endpoint
resource "aws_apigatewayv2_route" "health" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/health"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
}

# Image view endpoint
resource "aws_apigatewayv2_route" "image_view" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/images/{key}/view"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

# Image browser endpoint
resource "aws_apigatewayv2_route" "image_browser" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "GET /api/${var.stage_name}/images/browser/{path+}"
  target    = "integrations/${aws_apigatewayv2_integration.initialize.id}"
  
  authorization_type = var.use_api_key ? "API_KEY" : null
  api_key_required   = var.use_api_key
}

# Permission for API Gateway to invoke Lambda
resource "aws_lambda_permission" "api_gateway" {
  for_each = var.lambda_function_names

  statement_id  = "AllowExecutionFromAPIGateway-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = each.value
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*"
}

# Create usage plan if API key is used
resource "aws_api_gateway_usage_plan" "this" {
  count = var.use_api_key ? 1 : 0
  
  name        = "${var.api_name}-usage-plan"
  description = "Usage plan for ${var.api_name}"
  
  api_stages {
    api_id = aws_apigatewayv2_api.this.id
    stage  = aws_apigatewayv2_stage.this.id
  }
  
  quota_settings {
    limit  = var.api_quota_limit
    period = var.api_quota_period
  }
  
  throttle_settings {
    burst_limit = var.throttling_burst_limit
    rate_limit  = var.throttling_rate_limit
  }
  
  tags = var.common_tags
}

# Associate API key with usage plan
resource "aws_api_gateway_usage_plan_key" "this" {
  count = var.use_api_key ? 1 : 0
  
  key_id        = aws_apigatewayv2_api_key.this[0].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.this[0].id
}