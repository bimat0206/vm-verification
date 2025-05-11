resource "aws_lambda_function" "this" {
  for_each = var.functions_config

  function_name = each.value.name
  description   = each.value.description
  role          = var.execution_role_arn

  # Image configuration for container-based Lambda
  package_type = "Image"
  # Use ECR repository if enabled, otherwise use default image_uri
  image_uri = var.use_ecr_repository ? "${var.ecr_repository_url}/${each.key}:${var.image_tag}" : var.default_image_uri

  # Configure resources
  memory_size   = each.value.memory_size
  timeout       = each.value.timeout
  architectures = var.architectures

  # Environment variables
  environment {
    variables = each.value.environment_variables
  }

  # Event configuration
  reserved_concurrent_executions = try(each.value.reserved_concurrent_executions, -1)

  # Tracing configuration
  tracing_config {
    mode = "Active"
  }

  tags = merge(
    var.common_tags,
    {
      Name = each.value.name
    }
  )
}

# Configure logging
resource "aws_cloudwatch_log_group" "this" {
  for_each = var.functions_config

  name              = "/aws/lambda/${each.value.name}"
  retention_in_days = var.log_retention_days

  # tags = var.common_tags
}

# Lambda function permissions for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  for_each = var.api_gateway_source_arn != null ? var.functions_config : {}

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = var.api_gateway_source_arn
}

# Lambda function permissions for S3
resource "aws_lambda_permission" "s3" {
  for_each = var.s3_source_arns != null ? { for k, v in var.functions_config : k => v if contains(var.s3_trigger_functions, k) } : {}

  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[each.key].function_name
  principal     = "s3.amazonaws.com"
  source_arn    = var.s3_source_arns[each.key]
}

# Lambda function permissions for EventBridge
resource "aws_lambda_permission" "eventbridge" {
  for_each = var.eventbridge_source_arns != null ? { for k, v in var.functions_config : k => v if contains(var.eventbridge_trigger_functions, k) } : {}

  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[each.key].function_name
  principal     = "events.amazonaws.com"
  source_arn    = var.eventbridge_source_arns[each.key]
}
