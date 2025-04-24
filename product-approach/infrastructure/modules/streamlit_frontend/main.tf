# Complete infrastructure/modules/streamlit_frontend/main.tf

# Create ECR repository for Streamlit container
resource "aws_ecr_repository" "streamlit_app" {
  name                 = "${var.name_prefix}-streamlit-app"
  image_tag_mutability = var.image_tag_mutability

  image_scanning_configuration {
    scan_on_push = var.enable_scan_on_push
  }

  encryption_configuration {
    encryption_type = var.kms_key_arn != null ? "KMS" : "AES256"
    kms_key         = var.kms_key_arn
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-streamlit-app"
      Environment = var.environment
    },
    var.tags
  )
}

# Add lifecycle policy to ECR repository
resource "aws_ecr_lifecycle_policy" "streamlit_app_policy" {
  repository = aws_ecr_repository.streamlit_app.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        action = {
          type = "expire"
        }
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = var.max_image_count
        }
        description = "Keep only the latest ${var.max_image_count} images"
      }
    ]
  })
}

# Create a secret to store Streamlit app configuration
resource "aws_secretsmanager_secret" "streamlit_config" {
  name        = "${var.name_prefix}-streamlit-config"
  description = "Configuration for Streamlit frontend application"
  
  recovery_window_in_days = 7
  
  tags = merge(
    {
      Name        = "${var.name_prefix}-streamlit-config"
      Environment = var.environment
    },
    var.tags
  )
}

# Store API endpoints and other configuration in the secret
resource "aws_secretsmanager_secret_version" "streamlit_config_version" {
  secret_id = aws_secretsmanager_secret.streamlit_config.id
  
  secret_string = jsonencode({
    api_endpoint          = var.api_endpoint,
    dynamodb_table_name   = var.dynamodb_table_name,
    s3_bucket_name        = var.s3_bucket_name,
    step_functions_arn    = var.step_functions_arn,
    additional_config     = var.additional_config
  })
}

# Security Groups
resource "aws_security_group" "alb_sg" {
  name        = "${var.name_prefix}-alb-sg"
  description = "Security group for ALB"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-alb-sg"
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_security_group" "ecs_sg" {
  name        = "${var.name_prefix}-ecs-sg"
  description = "Security group for ECS tasks"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = var.container_port
    to_port         = var.container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.alb_sg.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-ecs-sg"
      Environment = var.environment
    },
    var.tags
  )
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "${var.name_prefix}-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(
    {
      Name        = "${var.name_prefix}-ecs-task-execution-role"
      Environment = var.environment
    },
    var.tags
  )
}

# IAM Role for ECS Task
resource "aws_iam_role" "ecs_task_role" {
  name = "${var.name_prefix}-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(
    {
      Name        = "${var.name_prefix}-ecs-task-role"
      Environment = var.environment
    },
    var.tags
  )
}

# Policy for ECS Task Execution Role
resource "aws_iam_policy" "ecs_task_execution_policy" {
  name        = "${var.name_prefix}-ecs-task-execution-policy"
  description = "Policy for ECS task execution"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "*"
      },
      {
        Effect = "Allow",
        Action = [
          "secretsmanager:GetSecretValue"
        ],
        Resource = [
          aws_secretsmanager_secret.streamlit_config.arn
        ]
      }
    ]
  })
}

# Policy for ECS Task Role
resource "aws_iam_policy" "ecs_task_policy" {
  name        = "${var.name_prefix}-ecs-task-policy"
  description = "Policy for ECS task"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ],
        Resource = [
          "arn:aws:s3:::${var.s3_bucket_name}",
          "arn:aws:s3:::${var.s3_bucket_name}/*",
          "arn:aws:dynamodb:${var.aws_region}:*:table/${var.dynamodb_table_name}",
          "arn:aws:dynamodb:${var.aws_region}:*:table/${var.dynamodb_table_name}/*"
        ]
      },
      {
        Effect = "Allow",
        Action = [
          "states:StartExecution",
          "states:DescribeExecution"
        ],
        Resource = [
          var.step_functions_arn
        ]
      }
    ]
  })
}

# Attach policies to roles
resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = aws_iam_policy.ecs_task_execution_policy.arn
}

resource "aws_iam_role_policy_attachment" "ecs_task_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_task_policy.arn
}

# Attach AWS managed policy for ECR and CloudWatch Logs
resource "aws_iam_role_policy_attachment" "ecs_task_execution_managed_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Create ALB
resource "aws_lb" "streamlit_alb" {
  name               = "${var.name_prefix}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb_sg.id]
  subnets            = var.public_subnet_ids  # Use the public subnet IDs directly

  enable_deletion_protection = false

  tags = merge(
    {
      Name        = "${var.name_prefix}-alb"
      Environment = var.environment
    },
    var.tags
  )
}

# Create ALB Target Group
resource "aws_lb_target_group" "streamlit_tg" {
  name        = "${var.name_prefix}-tg"
  port        = var.container_port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    interval            = 30
    path                = "/"
    port                = "traffic-port"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    timeout             = 5
    matcher             = "200-499"  # Accept a wider range of status codes as healthy
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-tg"
      Environment = var.environment
    },
    var.tags
  )
}

# Create ALB Listener (HTTP)
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.streamlit_alb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.streamlit_tg.arn
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-http-listener"
      Environment = var.environment
    },
    var.tags
  )
}

# Create HTTPS listener if certificate ARN is provided
resource "aws_lb_listener" "https" {
  count             = var.certificate_arn != null ? 1 : 0
  load_balancer_arn = aws_lb.streamlit_alb.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.streamlit_tg.arn
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-https-listener"
      Environment = var.environment
    },
    var.tags
  )
}

# Add HTTP to HTTPS redirect if certificate ARN is provided
resource "aws_lb_listener" "http_redirect" {
  count             = var.certificate_arn != null ? 1 : 0
  load_balancer_arn = aws_lb.streamlit_alb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-http-redirect"
      Environment = var.environment
    },
    var.tags
  )
}

# Create CloudWatch Log Group
resource "aws_cloudwatch_log_group" "streamlit_logs" {
  name              = "/ecs/${var.name_prefix}-streamlit"
  retention_in_days = var.log_retention_days

  tags = merge(
    {
      Name        = "${var.name_prefix}-logs"
      Environment = var.environment
    },
    var.tags
  )
}

# Create ECS Cluster
resource "aws_ecs_cluster" "streamlit_cluster" {
  name = "${var.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-cluster"
      Environment = var.environment
    },
    var.tags
  )
}

# Create ECS Task Definition
resource "aws_ecs_task_definition" "streamlit_task" {
  family                   = "${var.name_prefix}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "${var.name_prefix}-container"
      image     = "${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag}"
      essential = true

      portMappings = [
        {
          containerPort = var.container_port
          hostPort      = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "SECRET_ARN"
          value = aws_secretsmanager_secret.streamlit_config.arn
        },
        {
          name  = "REGION"
          value = var.aws_region
        },
        {
          name  = "API_ENDPOINT"
          value = var.api_endpoint
        },
        {
          name  = "DYNAMODB_TABLE"
          value = var.dynamodb_table_name
        },
        {
          name  = "S3_BUCKET"
          value = var.s3_bucket_name
        },
        {
          name  = "PYTHONUNBUFFERED"
          value = "1"
        },
        {
          name  = "STREAMLIT_SERVER_PORT" 
          value = tostring(var.container_port)
        },
        {
          name  = "STREAMLIT_SERVER_ADDRESS"
          value = "0.0.0.0"
        },
        {
          name  = "STREAMLIT_SERVER_HEADLESS"
          value = "true"
        },
        {
          name  = "STREAMLIT_SERVER_ENABLE_CORS"
          value = "false"
        },
        {
          name  = "STREAMLIT_SERVER_FILE_WATCHER_TYPE"
          value = "none"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.streamlit_logs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${var.container_port}/ || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])

  tags = merge(
    {
      Name        = "${var.name_prefix}-task"
      Environment = var.environment
    },
    var.tags
  )
}

# Create ECS Service
resource "aws_ecs_service" "streamlit_service" {
  name                               = "${var.name_prefix}-service"
  cluster                            = aws_ecs_cluster.streamlit_cluster.id
  task_definition                    = aws_ecs_task_definition.streamlit_task.arn
  desired_count                      = var.min_capacity
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  health_check_grace_period_seconds  = 120
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200

  network_configuration {
    subnets          = var.private_subnet_ids  # Use the private subnet IDs directly
    security_groups  = [aws_security_group.ecs_sg.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.streamlit_tg.arn
    container_name   = "${var.name_prefix}-container"
    container_port   = var.container_port
  }

  depends_on = [
    aws_lb_listener.http,
    aws_iam_role_policy_attachment.ecs_task_execution_policy_attachment,
    aws_iam_role_policy_attachment.ecs_task_policy_attachment
  ]

  tags = merge(
    {
      Name        = "${var.name_prefix}-service"
      Environment = var.environment
    },
    var.tags
  )
}

# Create Auto Scaling for ECS Service
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${aws_ecs_cluster.streamlit_cluster.name}/${aws_ecs_service.streamlit_service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_policy_cpu" {
  name               = "${var.name_prefix}-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "ecs_policy_memory" {
  name               = "${var.name_prefix}-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# Create CloudWatch Alarms for ECS Service
resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  alarm_name          = "${var.name_prefix}-cpu-high"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 85
  alarm_description   = "This metric monitors ECS CPU utilization"
  alarm_actions       = []

  dimensions = {
    ClusterName = aws_ecs_cluster.streamlit_cluster.name
    ServiceName = aws_ecs_service.streamlit_service.name
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-cpu-high"
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_cloudwatch_metric_alarm" "ecs_memory_high" {
  alarm_name          = "${var.name_prefix}-memory-high"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 85
  alarm_description   = "This metric monitors ECS memory utilization"
  alarm_actions       = []

  dimensions = {
    ClusterName = aws_ecs_cluster.streamlit_cluster.name
    ServiceName = aws_ecs_service.streamlit_service.name
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-memory-high"
      Environment = var.environment
    },
    var.tags
  )
}

# Null resource to build and push initial Docker image to ECR
resource "null_resource" "docker_build_push" {
  count = var.build_and_push_image ? 1 : 0
  
  triggers = {
    ecr_repository_url = aws_ecr_repository.streamlit_app.repository_url
    source_code_hash   = var.source_code_hash != "" ? var.source_code_hash : timestamp()
  }
  
  provisioner "local-exec" {
    working_dir = var.app_source_path
    command     = <<EOF
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${aws_ecr_repository.streamlit_app.repository_url}
      docker build -t ${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag} .
      docker push ${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag}
    EOF
  }
  
  depends_on = [aws_ecr_repository.streamlit_app]
}