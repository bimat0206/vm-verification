# ECS React Module using ALB and Fargate

locals {
  # Create a shorter service name prefix for resources with name length limitations
  short_service_name = var.environment != "" ? "${substr(var.service_name, 0, 5)}-${var.environment}" : substr(var.service_name, 0, 10)

  name_prefix = var.environment != "" ? "${var.service_name}-${var.environment}" : var.service_name
  name_suffix = var.name_suffix != "" ? var.name_suffix : ""

  service_name = lower(join("-", compact([local.name_prefix, "react", local.name_suffix])))

  # Default environment variables for React/Next.js
  default_react_env_vars = {
    PORT                    = tostring(var.port)
    NODE_ENV               = "production"
    NEXT_TELEMETRY_DISABLED = "1"
  }

  # Merge user-provided env vars with defaults
  environment_variables = merge(local.default_react_env_vars, var.environment_variables)
  
  # Convert environment variables to ECS format
  ecs_environment = [
    for key, value in local.environment_variables : {
      name  = key
      value = value
    }
  ]
}

# ECS Cluster
resource "aws_ecs_cluster" "react" {
  name = "${local.service_name}-cluster"

  setting {
    name  = "containerInsights"
    value = var.enable_container_insights ? "enabled" : "disabled"
  }

  tags = merge(
    var.common_tags,
    {
      Name = "${local.service_name}-cluster"
    }
  )
}

# ECS Task Definition
resource "aws_ecs_task_definition" "react" {
  family                   = local.service_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name         = local.service_name
      image        = var.image_uri
      essential    = true
      portMappings = [
        {
          containerPort = var.port
          hostPort      = var.port
          protocol      = "tcp"
        }
      ]
      environment = local.ecs_environment
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.react_logs.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = merge(
    var.common_tags,
    {
      Name = local.service_name
    }
  )
}

# ECS Service
resource "aws_ecs_service" "react" {
  name                               = local.service_name
  cluster                            = aws_ecs_cluster.react.id
  task_definition                    = aws_ecs_task_definition.react.arn
  desired_count                      = var.min_capacity
  launch_type                        = "FARGATE"
  platform_version                   = "LATEST"
  health_check_grace_period_seconds  = 60
  enable_execute_command             = var.enable_execute_command
  
  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.react.arn
    container_name   = local.service_name
    container_port   = var.port
  }

  # Auto-scaling is handled by aws_appautoscaling_* resources below
  
  lifecycle {
    ignore_changes = [desired_count]
  }

  depends_on = [
    aws_lb_listener.http,
    aws_iam_role_policy_attachment.ecs_execution_role_policy_attachment,
    aws_iam_role_policy_attachment.ecs_task_role_policy_attachment
  ]

  tags = merge(
    var.common_tags,
    {
      Name = local.service_name
    }
  )
}

# CloudWatch log group for ECS
resource "aws_cloudwatch_log_group" "react_logs" {
  name              = "/aws/ecs/${local.service_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# Auto Scaling Target
resource "aws_appautoscaling_target" "ecs_target" {
  count              = var.enable_auto_scaling ? 1 : 0
  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${aws_ecs_cluster.react.name}/${aws_ecs_service.react.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Auto Scaling Policy - CPU
resource "aws_appautoscaling_policy" "ecs_policy_cpu" {
  count              = var.enable_auto_scaling ? 1 : 0
  name               = "${local.service_name}-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = var.cpu_threshold
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# Auto Scaling Policy - Memory
resource "aws_appautoscaling_policy" "ecs_policy_memory" {
  count              = var.enable_auto_scaling ? 1 : 0
  name               = "${local.service_name}-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = var.memory_threshold
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# Current AWS Region
data "aws_region" "current" {}
