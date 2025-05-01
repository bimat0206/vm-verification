# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "this" {
  dashboard_name = var.dashboard_name
  dashboard_body = jsonencode({
    widgets = concat(
      # Lambda Widgets
      [
        for name, function_name in var.lambda_function_names : {
          type   = "metric"
          width  = 12
          height = 6
          properties = {
            metrics = [
              ["AWS/Lambda", "Invocations", "FunctionName", function_name],
              ["AWS/Lambda", "Errors", "FunctionName", function_name],
              ["AWS/Lambda", "Duration", "FunctionName", function_name],
              ["AWS/Lambda", "Throttles", "FunctionName", function_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "Lambda: ${name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ],
      # Step Functions Widget
      var.step_function_name != "" ? [
        {
          type   = "metric"
          width  = 24
          height = 6
          properties = {
            metrics = [
              ["AWS/States", "ExecutionsStarted", "StateMachineArn", var.step_function_name],
              ["AWS/States", "ExecutionsFailed", "StateMachineArn", var.step_function_name],
              ["AWS/States", "ExecutionsSucceeded", "StateMachineArn", var.step_function_name],
              ["AWS/States", "ExecutionThrottled", "StateMachineArn", var.step_function_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "Step Functions: ${var.step_function_name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ] : [],
      # DynamoDB Widgets
      [
        for name, table_name in var.dynamodb_table_names : {
          type   = "metric"
          width  = 12
          height = 6
          properties = {
            metrics = [
              ["AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", table_name],
              ["AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", table_name],
              ["AWS/DynamoDB", "ThrottledRequests", "TableName", table_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "DynamoDB: ${name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ],
      # API Gateway Widget
      var.api_gateway_name != "" ? [
        {
          type   = "metric"
          width  = 24
          height = 6
          properties = {
            metrics = [
              ["AWS/ApiGateway", "Count", "ApiName", var.api_gateway_name],
              ["AWS/ApiGateway", "4XXError", "ApiName", var.api_gateway_name],
              ["AWS/ApiGateway", "5XXError", "ApiName", var.api_gateway_name],
              ["AWS/ApiGateway", "Latency", "ApiName", var.api_gateway_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "API Gateway: ${var.api_gateway_name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ] : [],
      # ECR Repository Widgets
      [
        for name, repo_name in var.ecr_repository_names : {
          type   = "metric"
          width  = 12
          height = 6
          properties = {
            metrics = [
              ["AWS/ECR", "RepositoryPullCount", "RepositoryName", repo_name],
              ["AWS/ECR", "ScanFindingsSeverityCritical", "RepositoryName", repo_name],
              ["AWS/ECR", "ScanFindingsSeverityHigh", "RepositoryName", repo_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "ECR: ${name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ],
      # App Runner Widget
      var.app_runner_service_name != "" ? [
        {
          type   = "metric"
          width  = 24
          height = 6
          properties = {
            metrics = [
              ["AWS/AppRunner", "Requests", "ServiceName", var.app_runner_service_name],
              ["AWS/AppRunner", "HTTP4xx", "ServiceName", var.app_runner_service_name],
              ["AWS/AppRunner", "HTTP5xx", "ServiceName", var.app_runner_service_name],
              ["AWS/AppRunner", "Latency", "ServiceName", var.app_runner_service_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "App Runner: ${var.app_runner_service_name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ] : []
    )
  })

  lifecycle {
    ignore_changes = [
      dashboard_body # Prevent accidental changes in the console from being overwritten
    ]
  }
}

# SNS Topic for Alarms
resource "aws_sns_topic" "alarms" {
  count = length(var.alarm_email_endpoints) > 0 ? 1 : 0
  name  = "${var.dashboard_name}-alarms"
  tags  = var.common_tags
}

# SNS Subscriptions for Email Endpoints
resource "aws_sns_topic_subscription" "email_alerts" {
  count     = length(var.alarm_email_endpoints)
  topic_arn = aws_sns_topic.alarms[0].arn
  protocol  = "email"
  endpoint  = var.alarm_email_endpoints[count.index]
}

# Lambda Error Rate Alarms
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  for_each = var.lambda_function_names

  alarm_name          = "${each.value}-error-rate-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 3
  alarm_description   = "Alarm for ${each.value} function error rate exceeding threshold"
  actions_enabled     = length(var.alarm_email_endpoints) > 0 ? true : false
  alarm_actions       = length(var.alarm_email_endpoints) > 0 ? [aws_sns_topic.alarms[0].arn] : []
  ok_actions          = length(var.alarm_email_endpoints) > 0 ? [aws_sns_topic.alarms[0].arn] : []

  dimensions = {
    FunctionName = each.value
  }

  tags = var.common_tags
}

# Step Functions Execution Failed Alarm
resource "aws_cloudwatch_metric_alarm" "step_functions_failed" {
  count = var.step_function_name != "" && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.step_function_name}-failed-executions-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ExecutionsFailed"
  namespace           = "AWS/States"
  period              = 300
  statistic           = "Sum"
  threshold           = 1
  alarm_description   = "Alarm for ${var.step_function_name} state machine failed executions"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    StateMachineArn = var.step_function_name
  }

  tags = var.common_tags
}

# DynamoDB Throttled Requests Alarm
resource "aws_cloudwatch_metric_alarm" "dynamodb_throttles" {
  for_each = var.dynamodb_table_names

  alarm_name          = "${each.value}-throttled-requests-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ThrottledRequests"
  namespace           = "AWS/DynamoDB"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Alarm for ${each.value} DynamoDB table throttled requests"
  actions_enabled     = length(var.alarm_email_endpoints) > 0 ? true : false
  alarm_actions       = length(var.alarm_email_endpoints) > 0 ? [aws_sns_topic.alarms[0].arn] : []
  ok_actions          = length(var.alarm_email_endpoints) > 0 ? [aws_sns_topic.alarms[0].arn] : []

  dimensions = {
    TableName = each.value
  }

  tags = var.common_tags
}

# API Gateway 5XX Error Alarm
resource "aws_cloudwatch_metric_alarm" "api_gateway_5xx" {
  count = var.api_gateway_name != "" && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.api_gateway_name}-5xx-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGateway"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Alarm for ${var.api_gateway_name} API Gateway 5XX errors"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    ApiName = var.api_gateway_name
  }

  tags = var.common_tags
}

# App Runner Error Alarm
resource "aws_cloudwatch_metric_alarm" "app_runner_5xx" {
  count = var.app_runner_service_name != "" && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.app_runner_service_name}-5xx-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "HTTP5xx"
  namespace           = "AWS/AppRunner"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Alarm for ${var.app_runner_service_name} App Runner 5XX errors"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    ServiceName = var.app_runner_service_name
  }

  tags = var.common_tags
}

# Log Group Configuration for Lambda Functions
resource "aws_cloudwatch_log_group" "lambda_log_groups" {
  for_each = var.lambda_function_names

  name              = "/aws/lambda/${each.value}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# Log Group for Step Functions
resource "aws_cloudwatch_log_group" "step_functions_log_group" {
  count = var.step_function_name != "" ? 1 : 0

  name              = "/aws/states/${var.step_function_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway_log_group" {
  count = var.api_gateway_name != "" ? 1 : 0

  name              = "/aws/apigateway/${var.api_gateway_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# Log Group for App Runner
resource "aws_cloudwatch_log_group" "app_runner_log_group" {
  count = var.app_runner_service_name != "" ? 1 : 0

  name              = "/aws/apprunner/${var.app_runner_service_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}