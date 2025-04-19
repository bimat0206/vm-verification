# infrastructure/modules/monitoring/main.tf
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = var.dashboard_name
  dashboard_body = templatefile("${path.module}/dashboard.json.tftpl", {
    region                  = var.aws_region
    lambda_functions        = var.lambda_functions
    state_machine_arn       = var.state_machine_arn
    dynamodb_table          = var.dynamodb_table
    api_gateway_api_name    = var.api_gateway_api_name
    api_gateway_stage_name  = var.api_gateway_stage_name
    s3_bucket_name          = var.s3_bucket_name
  })
}

# Create CloudWatch Alarms for Lambda Errors
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  for_each = toset(var.lambda_functions)
  
  alarm_name          = "${each.value}-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = var.lambda_error_threshold
  alarm_description   = "This alarm monitors errors in Lambda function ${each.value}"
  treat_missing_data  = "notBreaching"
  
  dimensions = {
    FunctionName = each.value
  }
  
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions
  
  tags = merge(
    {
      Name        = "${each.value}-errors"
      Environment = var.environment
    },
    var.tags
  )
}

# Create CloudWatch Alarm for API Gateway 5XX errors
resource "aws_cloudwatch_metric_alarm" "api_gateway_5xx" {
  alarm_name          = "${var.api_gateway_api_name}-5xx-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGateway"
  period              = 60
  statistic           = "Sum"
  threshold           = var.api_gateway_error_threshold
  alarm_description   = "This alarm monitors 5XX errors in API Gateway"
  treat_missing_data  = "notBreaching"
  
  dimensions = {
    ApiName = var.api_gateway_api_name
    Stage   = var.api_gateway_stage_name
  }
  
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions
  
  tags = merge(
    {
      Name        = "${var.api_gateway_api_name}-5xx-errors"
      Environment = var.environment
    },
    var.tags
  )
}

# Create CloudWatch Alarm for Step Functions execution failures
resource "aws_cloudwatch_metric_alarm" "step_functions_failures" {
  alarm_name          = "${var.state_machine_name}-failures"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ExecutionsFailed"
  namespace           = "AWS/States"
  period              = 60
  statistic           = "Sum"
  threshold           = var.step_functions_failure_threshold
  alarm_description   = "This alarm monitors failures in Step Functions state machine ${var.state_machine_name}"
  treat_missing_data  = "notBreaching"
  
  dimensions = {
    StateMachineArn = var.state_machine_arn
  }
  
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions
  
  tags = merge(
    {
      Name        = "${var.state_machine_name}-failures"
      Environment = var.environment
    },
    var.tags
  )
}

# Log Groups
resource "aws_cloudwatch_log_group" "lambda_logs" {
  for_each = toset(var.lambda_functions)
  
  name              = "/aws/lambda/${each.value}"
  retention_in_days = var.log_retention_days
  
  tags = merge(
    {
      Name        = "/aws/lambda/${each.value}"
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_cloudwatch_log_group" "api_gateway_logs" {
  name              = "API-Gateway-Execution-Logs_${var.api_gateway_api_name}/${var.api_gateway_stage_name}"
  retention_in_days = var.log_retention_days
  
  tags = merge(
    {
      Name        = "API-Gateway-Logs-${var.api_gateway_api_name}"
      Environment = var.environment
    },
    var.tags
  )
}

# Fix for the monitoring module's state machine log group
# Modify the log group definition to handle existing log groups gracefully

resource "aws_cloudwatch_log_group" "state_machine_logs" {
  # Add a count parameter to conditionally create this resource
  count             = var.create_state_machine_log_group ? 1 : 0
  name              = "/aws/states/${var.state_machine_name}"
  retention_in_days = var.log_retention_days
  
  tags = merge(
    {
      Name        = "/aws/states/${var.state_machine_name}"
      Environment = var.environment
    },
    var.tags
  )
  
  # Prevent errors when the log group already exists
  lifecycle {
    ignore_changes = [tags]
  }
}