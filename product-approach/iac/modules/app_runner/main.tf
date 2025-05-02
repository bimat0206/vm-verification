# IAM Role for App Runner
resource "aws_iam_role" "app_runner_role" {
  name = "${var.service_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "build.apprunner.amazonaws.com"
        }
      }
    ]
  })

  # tags = var.common_tags
}

# IAM Policy for App Runner to access ECR
resource "aws_iam_policy" "app_runner_ecr_policy" {
  name        = "${var.service_name}-ecr-policy"
  description = "Policy for App Runner to access ECR"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetAuthorizationToken",
          "ecr:DescribeRepositories",
          "ecr:DescribeImages",
          "ecr:ListImages"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Attach ECR policy to App Runner role
resource "aws_iam_role_policy_attachment" "app_runner_ecr_attachment" {
  role       = aws_iam_role.app_runner_role.name
  policy_arn = aws_iam_policy.app_runner_ecr_policy.arn
}

# IAM Instance Role for App Runner service (for API access)
resource "aws_iam_role" "app_runner_instance_role" {
  name = "${var.service_name}-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "tasks.apprunner.amazonaws.com"
        }
      }
    ]
  })

  # tags = var.common_tags
}

# API Gateway access policy for App Runner instance role
resource "aws_iam_policy" "app_runner_api_gateway_policy" {
  count = var.api_gateway_access ? 1 : 0

  name        = "${var.service_name}-api-gateway-policy"
  description = "Policy for App Runner to access API Gateway"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "execute-api:Invoke",
          "execute-api:ManageConnections"
        ]
        Effect   = "Allow"
        Resource = "${var.api_gateway_arn}/*"
      }
    ]
  })
}

# Attach API Gateway policy to App Runner instance role
resource "aws_iam_role_policy_attachment" "app_runner_api_gateway_attachment" {
  count      = var.api_gateway_access ? 1 : 0
  role       = aws_iam_role.app_runner_instance_role.name
  policy_arn = aws_iam_policy.app_runner_api_gateway_policy[0].arn
}

# S3 access policy for App Runner instance role
resource "aws_iam_policy" "app_runner_s3_policy" {
  count = length(var.s3_bucket_arns) > 0 ? 1 : 0

  name        = "${var.service_name}-s3-policy"
  description = "Policy for App Runner to access S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Effect   = "Allow"
        Resource = concat(var.s3_bucket_arns, [for arn in var.s3_bucket_arns : "${arn}/*"])
      }
    ]
  })
}

# Attach S3 policy to App Runner instance role
resource "aws_iam_role_policy_attachment" "app_runner_s3_attachment" {
  count      = length(var.s3_bucket_arns) > 0 ? 1 : 0
  role       = aws_iam_role.app_runner_instance_role.name
  policy_arn = aws_iam_policy.app_runner_s3_policy[0].arn
}

# App Runner service
resource "aws_apprunner_service" "this" {
  service_name = var.service_name

  source_configuration {
    image_repository {
      image_configuration {
        port                          = var.port
        runtime_environment_variables = var.environment_variables
      }
      image_identifier      = var.image_uri
      image_repository_type = var.image_repo_type
    }

    # Only include authentication configuration for private ECR repositories
    dynamic "authentication_configuration" {
      for_each = var.image_repo_type == "ECR" ? [1] : []
      content {
        access_role_arn = aws_iam_role.app_runner_role.arn
      }
    }

    auto_deployments_enabled = var.auto_deployments_enabled
  }

  instance_configuration {
    cpu               = "${var.cpu} vCPU"
    memory            = var.memory
    instance_role_arn = aws_iam_role.app_runner_instance_role.arn
  }

  health_check_configuration {
    protocol            = var.health_check_protocol
    path                = var.health_check_path
    interval            = var.health_check_interval
    timeout             = var.health_check_timeout
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold
  }

  network_configuration {
    ingress_configuration {
      is_publicly_accessible = var.is_publicly_accessible
    }
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.service_name
    }
  )
}

# CloudWatch log group for App Runner
resource "aws_cloudwatch_log_group" "app_runner" {
  name              = "/aws/apprunner/${var.service_name}"
  retention_in_days = var.log_retention_days

  # tags = var.common_tags
}

# Auto scaling configuration (optional)
resource "aws_apprunner_auto_scaling_configuration_version" "this" {
  count = var.enable_auto_scaling ? 1 : 0

  auto_scaling_configuration_name = "${var.service_name}-auto-scaling"

  max_concurrency = var.max_concurrency
  max_size        = var.max_size
  min_size        = var.min_size

  tags = var.common_tags
}

# Associate auto scaling configuration with service
resource "aws_apprunner_custom_domain_association" "this" {
  count = var.custom_domain_name != "" ? 1 : 0

  domain_name          = var.custom_domain_name
  service_arn          = aws_apprunner_service.this.arn
  enable_www_subdomain = var.enable_www_subdomain
}
