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
      # ECS Widget
      var.ecs_cluster_name != "" && var.ecs_service_name != "" ? [
        {
          type   = "metric"
          width  = 12
          height = 6
          properties = {
            metrics = [
              ["AWS/ECS", "CPUUtilization", "ClusterName", var.ecs_cluster_name, "ServiceName", var.ecs_service_name],
              ["AWS/ECS", "MemoryUtilization", "ClusterName", var.ecs_cluster_name, "ServiceName", var.ecs_service_name]
            ]
            period = 300
            stat   = "Average"
            title  = "ECS: ${var.ecs_service_name}"
            view   = "timeSeries"
            region = var.region
          }
        }
      ] : [],
      # ALB Widget
      var.alb_name != "" ? [
        {
          type   = "metric"
          width  = 12
          height = 6
          properties = {
            metrics = [
              ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", var.alb_name],
              ["AWS/ApplicationELB", "HTTPCode_Target_4XX_Count", "LoadBalancer", var.alb_name],
              ["AWS/ApplicationELB", "HTTPCode_Target_5XX_Count", "LoadBalancer", var.alb_name],
              ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", var.alb_name]
            ]
            period = 300
            stat   = "Sum"
            title  = "ALB: ${var.alb_name}"
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

  # Don't add tags here as they're provided by default_tags in the provider
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
}

# Step Functions Execution Failed Alarm
resource "aws_cloudwatch_metric_alarm" "step_functions_failed" {
  count = var.enable_step_function_monitoring && length(var.alarm_email_endpoints) > 0 ? 1 : 0

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
}

# API Gateway 5XX Error Alarm
resource "aws_cloudwatch_metric_alarm" "api_gateway_5xx" {
  count = var.enable_api_gateway_monitoring && length(var.alarm_email_endpoints) > 0 ? 1 : 0

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
}

# ECS CPU Utilization Alarm
resource "aws_cloudwatch_metric_alarm" "ecs_cpu_utilization" {
  count = var.enable_ecs_monitoring && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.ecs_service_name}-cpu-utilization-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Alarm for ${var.ecs_service_name} ECS service CPU utilization exceeding threshold"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    ClusterName = var.ecs_cluster_name
    ServiceName = var.ecs_service_name
  }
}

# ECS Memory Utilization Alarm
resource "aws_cloudwatch_metric_alarm" "ecs_memory_utilization" {
  count = var.enable_ecs_monitoring && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.ecs_service_name}-memory-utilization-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Alarm for ${var.ecs_service_name} ECS service memory utilization exceeding threshold"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    ClusterName = var.ecs_cluster_name
    ServiceName = var.ecs_service_name
  }
}

# ALB 5XX Error Alarm
resource "aws_cloudwatch_metric_alarm" "alb_5xx" {
  count = var.enable_alb_monitoring && length(var.alarm_email_endpoints) > 0 ? 1 : 0

  alarm_name          = "${var.alb_name}-5xx-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "HTTPCode_Target_5XX_Count"
  namespace           = "AWS/ApplicationELB"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Alarm for ${var.alb_name} ALB 5XX errors"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alarms[0].arn]
  ok_actions          = [aws_sns_topic.alarms[0].arn]

  dimensions = {
    LoadBalancer = var.alb_name
  }
}

# Note: Log groups are now created in their respective service modules
# - Lambda log groups are created in the lambda module
# - Step Functions log groups are created in the step_functions module
# - API Gateway log groups are created in the api_gateway module
# - App Runner log groups are created in the streamlit-frontend module
