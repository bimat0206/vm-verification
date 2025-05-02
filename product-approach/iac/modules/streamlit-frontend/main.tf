# Streamlit frontend module using AWS App Runner

locals {
  # Create a shorter service name prefix for resources with name length limitations
  short_service_name = var.environment != "" ? "${substr(var.service_name, 0, 5)}-${var.environment}" : substr(var.service_name, 0, 10)

  name_prefix = var.environment != "" ? "${var.service_name}-${var.environment}" : var.service_name
  name_suffix = var.name_suffix != "" ? var.name_suffix : ""

  service_name = lower(join("-", compact([local.name_prefix, "streamlit", local.name_suffix])))

  # Default environment variables for Streamlit
  default_streamlit_env_vars = {
    STREAMLIT_SERVER_PORT                = tostring(var.port)
    STREAMLIT_SERVER_ADDRESS             = "0.0.0.0"
    STREAMLIT_SERVER_HEADLESS            = "true"
    STREAMLIT_BROWSER_GATHER_USAGE_STATS = "false"
    STREAMLIT_THEME_BASE                 = var.theme_mode
  }

  # Merge user-provided env vars with defaults
  environment_variables = merge(local.default_streamlit_env_vars, var.environment_variables)
}

# IAM Role for App Runner
resource "aws_iam_role" "streamlit_role" {
  name = "${local.service_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "build.apprunner.amazonaws.com",
            "ecr.amazonaws.com"
          ]
        }
      }
    ]
  })

  tags = var.common_tags
}

# IAM Policy for App Runner to access ECR
resource "aws_iam_policy" "streamlit_ecr_policy" {
  name        = "${local.service_name}-ecr-policy"
  description = "Policy for Streamlit App Runner to access ECR"

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
          "ecr:ListImages",
          "ecr:CreateRepository",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:BatchDeleteImage",
          "ecr:DeleteRepository",
          "ecr:SetRepositoryPolicy",
          "ecr:TagResource",
          "ecr:UntagResource",
          "ecr:PutLifecyclePolicy",
          "ecr:PutImageTagMutability",
          "ecr:PutImageScanningConfiguration"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })

  tags = var.common_tags
}

# Attach ECR policy to App Runner role
resource "aws_iam_role_policy_attachment" "streamlit_ecr_attachment" {
  role       = aws_iam_role.streamlit_role.name
  policy_arn = aws_iam_policy.streamlit_ecr_policy.arn
}

# IAM Instance Role for App Runner service 
resource "aws_iam_role" "streamlit_instance_role" {
  name = "${local.service_name}-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "tasks.apprunner.amazonaws.com",
            "ecr.amazonaws.com"
          ]
        }
      }
    ]
  })

  tags = var.common_tags
}

# CloudWatch access policy for the instance role
resource "aws_iam_policy" "streamlit_cloudwatch_policy" {
  name        = "${local.service_name}-cloudwatch-policy"
  description = "Policy for Streamlit App Runner to write logs and metrics"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "cloudwatch:PutMetricData"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })

  tags = var.common_tags
}

# Attach CloudWatch policy to instance role
resource "aws_iam_role_policy_attachment" "streamlit_cloudwatch_attachment" {
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_cloudwatch_policy.arn
}

# Attach ECR policy to instance role
resource "aws_iam_role_policy_attachment" "streamlit_instance_ecr_attachment" {
  count      = var.enable_ecr_full_access ? 1 : 0
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_ecr_policy.arn
}

# API Gateway access policy (if needed)
resource "aws_iam_policy" "streamlit_api_gateway_policy" {
  count = var.enable_api_gateway_access ? 1 : 0

  name        = "${local.service_name}-api-gateway-policy"
  description = "Policy for Streamlit App Runner to access API Gateway"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "execute-api:Invoke",
          "execute-api:ManageConnections"
        ]
        Effect   = "Allow"
        Resource = var.api_gateway_arn != "" ? "${var.api_gateway_arn}/*" : "*"
      }
    ]
  })

  tags = var.common_tags
}

# Attach API Gateway policy to instance role if needed
resource "aws_iam_role_policy_attachment" "streamlit_api_gateway_attachment" {
  count      = var.enable_api_gateway_access ? 1 : 0
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_api_gateway_policy[0].arn
}

# S3 access policy
resource "aws_iam_policy" "streamlit_s3_policy" {
  count = var.enable_s3_access ? 1 : 0

  name        = "${local.service_name}-s3-policy"
  description = "Policy for Streamlit App Runner to access S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:GetBucketLocation"
        ]
        Effect   = "Allow"
        Resource = length(var.s3_bucket_arns) > 0 ? concat(var.s3_bucket_arns, [for arn in var.s3_bucket_arns : "${arn}/*"]) : ["*"]
      }
    ]
  })

  tags = var.common_tags
}

# Attach S3 policy to instance role if needed
resource "aws_iam_role_policy_attachment" "streamlit_s3_attachment" {
  count      = var.enable_s3_access ? 1 : 0
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_s3_policy[0].arn
}

# DynamoDB access policy
resource "aws_iam_policy" "streamlit_dynamodb_policy" {
  count = var.enable_dynamodb_access ? 1 : 0

  name        = "${local.service_name}-dynamodb-policy"
  description = "Policy for Streamlit App Runner to access DynamoDB tables"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:BatchGetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem"
        ]
        Effect   = "Allow"
        Resource = length(var.dynamodb_table_arns) > 0 ? var.dynamodb_table_arns : ["*"]
      }
    ]
  })

  tags = var.common_tags
}

# Attach DynamoDB policy to instance role if needed
resource "aws_iam_role_policy_attachment" "streamlit_dynamodb_attachment" {
  count      = var.enable_dynamodb_access ? 1 : 0
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_dynamodb_policy[0].arn
}

# App Runner service
resource "aws_apprunner_service" "streamlit" {
  service_name = local.service_name

  source_configuration {
    image_repository {
      image_configuration {
        port                          = var.port
        runtime_environment_variables = local.environment_variables
      }
      image_identifier      = var.image_uri
      image_repository_type = var.image_repository_type
    }

    # Only include authentication configuration for private ECR repositories
    dynamic "authentication_configuration" {
      for_each = var.image_repository_type == "ECR" ? [1] : []
      content {
        access_role_arn = aws_iam_role.streamlit_role.arn
      }
    }

    auto_deployments_enabled = var.auto_deployments_enabled
  }

  instance_configuration {
    cpu               = var.cpu
    memory            = var.memory
    instance_role_arn = aws_iam_role.streamlit_instance_role.arn
  }

  health_check_configuration {
    protocol            = "HTTP"
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

  # Directly reference the auto scaling configuration if enabled
  auto_scaling_configuration_arn = var.enable_auto_scaling ? aws_apprunner_auto_scaling_configuration_version.streamlit_scaling[0].arn : null

  tags = merge(
    var.common_tags,
    {
      Name = local.service_name
    }
  )
}

# CloudWatch log group for App Runner
resource "aws_cloudwatch_log_group" "streamlit_logs" {
  name              = "/aws/apprunner/${local.service_name}"
  retention_in_days = var.log_retention_days

  tags = var.common_tags
}

# Auto scaling configuration (if enabled)
resource "aws_apprunner_auto_scaling_configuration_version" "streamlit_scaling" {
  count = var.enable_auto_scaling ? 1 : 0

  auto_scaling_configuration_name = "${local.short_service_name}-st-as"

  max_concurrency = var.max_concurrency
  max_size        = var.max_size
  min_size        = var.min_size

  tags = var.common_tags
}

# Auto scaling configuration is now directly attached in the aws_apprunner_service resource


resource "aws_apprunner_observability_configuration" "streamlit" {
  observability_configuration_name = "streamlit-observability"

  trace_configuration {
    vendor = "AWSXRAY"
  }
}
