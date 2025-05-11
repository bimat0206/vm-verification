# IAM roles and policies for ECS Streamlit

locals {
  # Define the Secrets Manager resource ARN based on the API_KEY_SECRET_NAME
  secret_resource_arn = lookup(var.environment_variables, "API_KEY_SECRET_NAME", "") != "" ? (
    "arn:aws:secretsmanager:${data.aws_region.current.name}:*:secret:${lookup(var.environment_variables, "API_KEY_SECRET_NAME", "")}*"
    ) : (
    "arn:aws:secretsmanager:*:*:secret:*"
  )
}

# ECS Task Execution Role - Used by the ECS service to pull images and publish logs
resource "aws_iam_role" "ecs_execution_role" {
  name = "${local.service_name}-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.common_tags
}

# ECS Task Role - Used by the container itself to access AWS services
resource "aws_iam_role" "ecs_task_role" {
  name = "${local.service_name}-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.common_tags
}

# Attach the AWS managed ECS Task Execution Role policy
resource "aws_iam_role_policy_attachment" "ecs_execution_role_policy_attachment" {
  role       = aws_iam_role.ecs_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# CloudWatch access policy for the task role
resource "aws_iam_policy" "ecs_cloudwatch_policy" {
  name        = "${local.service_name}-cloudwatch-policy"
  description = "Policy for Streamlit ECS tasks to write logs and metrics"

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

# Attach CloudWatch policy to task role
resource "aws_iam_role_policy_attachment" "ecs_task_role_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_cloudwatch_policy.arn
}

# ECR access policy
resource "aws_iam_policy" "ecs_ecr_policy" {
  count = var.enable_ecr_full_access ? 1 : 0

  name        = "${local.service_name}-ecr-policy"
  description = "Policy for Streamlit ECS tasks to access ECR"

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

  tags = var.common_tags
}

# Attach ECR policy to task role if needed
resource "aws_iam_role_policy_attachment" "ecs_ecr_attachment" {
  count      = var.enable_ecr_full_access ? 1 : 0
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_ecr_policy[0].arn
}

# API Gateway access policy
resource "aws_iam_policy" "ecs_api_gateway_policy" {
  count = var.enable_api_gateway_access ? 1 : 0

  name        = "${local.service_name}-api-gateway-policy"
  description = "Policy for Streamlit ECS tasks to access API Gateway"

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

# Attach API Gateway policy to task role if needed
resource "aws_iam_role_policy_attachment" "ecs_api_gateway_attachment" {
  count      = var.enable_api_gateway_access ? 1 : 0
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_api_gateway_policy[0].arn
}

# S3 access policy
resource "aws_iam_policy" "ecs_s3_policy" {
  count = var.enable_s3_access ? 1 : 0

  name        = "${local.service_name}-s3-policy"
  description = "Policy for Streamlit ECS tasks to access S3 buckets"

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

# Attach S3 policy to task role if needed
resource "aws_iam_role_policy_attachment" "ecs_s3_attachment" {
  count      = var.enable_s3_access ? 1 : 0
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_s3_policy[0].arn
}

# DynamoDB access policy
resource "aws_iam_policy" "ecs_dynamodb_policy" {
  count = var.enable_dynamodb_access ? 1 : 0

  name        = "${local.service_name}-dynamodb-policy"
  description = "Policy for Streamlit ECS tasks to access DynamoDB tables"

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

# Attach DynamoDB policy to task role if needed
resource "aws_iam_role_policy_attachment" "ecs_dynamodb_attachment" {
  count      = var.enable_dynamodb_access ? 1 : 0
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_dynamodb_policy[0].arn
}

# Secrets Manager access policy
resource "aws_iam_policy" "ecs_secretsmanager_policy" {
  count = contains(keys(var.environment_variables), "API_KEY_SECRET_NAME") ? 1 : 0

  name        = "${local.service_name}-secretsmanager-policy"
  description = "Policy for Streamlit ECS tasks to access Secrets Manager"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Effect   = "Allow"
        Resource = local.secret_resource_arn
      }
    ]
  })

  tags = var.common_tags
}

# Attach Secrets Manager policy to task role if needed
resource "aws_iam_role_policy_attachment" "ecs_secretsmanager_attachment" {
  count      = contains(keys(var.environment_variables), "API_KEY_SECRET_NAME") ? 1 : 0
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_secretsmanager_policy[0].arn
}
