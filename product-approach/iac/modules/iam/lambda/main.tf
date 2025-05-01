# IAM Role for Lambda Functions
resource "aws_iam_role" "lambda_execution_role" {
  name = "${var.project_name}-${var.environment}-lambda-execution-role-${var.name_suffix}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = var.common_tags
}

# Basic Lambda execution policy attachment
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# S3 access policy
resource "aws_iam_policy" "s3_access" {
  count = length(var.s3_bucket_arns) > 0 ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-s3-access-${var.name_suffix}"
  description = "Policy for Lambda to access S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:DeleteObject",
          "s3:GetObjectVersion",
          "s3:GetObjectTagging",
          "s3:PutObjectTagging"
        ]
        Effect   = "Allow"
        Resource = concat(var.s3_bucket_arns, [for arn in var.s3_bucket_arns : "${arn}/*"])
      }
    ]
  })
}

# Attach S3 policy to Lambda role
resource "aws_iam_role_policy_attachment" "s3_access" {
  count      = length(var.s3_bucket_arns) > 0 ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.s3_access[0].arn
}

# DynamoDB access policy
resource "aws_iam_policy" "dynamodb_access" {
  count = length(var.dynamodb_table_arns) > 0 ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-dynamodb-access-${var.name_suffix}"
  description = "Policy for Lambda to access DynamoDB tables"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:BatchGetItem",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchWriteItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem"
        ]
        Effect   = "Allow"
        Resource = var.dynamodb_table_arns
      },
      {
        Action = [
          "dynamodb:Query"
        ]
        Effect   = "Allow"
        Resource = [for arn in var.dynamodb_table_arns : "${arn}/index/*"]
      }
    ]
  })
}

# Attach DynamoDB policy to Lambda role
resource "aws_iam_role_policy_attachment" "dynamodb_access" {
  count      = length(var.dynamodb_table_arns) > 0 ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.dynamodb_access[0].arn
}

# ECR access policy
resource "aws_iam_policy" "ecr_access" {
  count = length(var.ecr_repository_arns) > 0 ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-ecr-access-${var.name_suffix}"
  description = "Policy for Lambda to access ECR repositories"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetRepositoryPolicy",
          "ecr:DescribeRepositories",
          "ecr:ListImages",
          "ecr:DescribeImages"
        ]
        Effect   = "Allow"
        Resource = var.ecr_repository_arns
      },
      {
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Attach ECR policy to Lambda role
resource "aws_iam_role_policy_attachment" "ecr_access" {
  count      = length(var.ecr_repository_arns) > 0 ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.ecr_access[0].arn
}

# Bedrock access policy
resource "aws_iam_policy" "bedrock_access" {
  count = var.bedrock_model_arn != "" ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-bedrock-access-${var.name_suffix}"
  description = "Policy for Lambda to access Amazon Bedrock models"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Effect   = "Allow"
        Resource = var.bedrock_model_arn
      },
      {
        Action = [
          "bedrock:ListFoundationModels",
          "bedrock:GetFoundationModel"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Attach Bedrock policy to Lambda role
resource "aws_iam_role_policy_attachment" "bedrock_access" {
  count      = var.bedrock_model_arn != "" ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.bedrock_access[0].arn
}

# SNS publish policy
resource "aws_iam_policy" "sns_publish" {
  count = var.sns_topic_arns != null ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-sns-publish-${var.name_suffix}"
  description = "Policy for Lambda to publish messages to SNS topics"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sns:Publish"
        ]
        Effect   = "Allow"
        Resource = var.sns_topic_arns
      }
    ]
  })
}

# Attach SNS policy to Lambda role
resource "aws_iam_role_policy_attachment" "sns_publish" {
  count      = var.sns_topic_arns != null ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.sns_publish[0].arn
}

# Step Functions access policy
resource "aws_iam_policy" "step_functions_access" {
  count = var.step_functions_arns != null ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-step-functions-access-${var.name_suffix}"
  description = "Policy for Lambda to start and describe Step Functions executions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "states:StartExecution",
          "states:DescribeExecution",
          "states:GetExecutionHistory"
        ]
        Effect   = "Allow"
        Resource = var.step_functions_arns
      }
    ]
  })
}

# Attach Step Functions policy to Lambda role
resource "aws_iam_role_policy_attachment" "step_functions_access" {
  count      = var.step_functions_arns != null ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.step_functions_access[0].arn
}

# Secrets Manager access policy
resource "aws_iam_policy" "secrets_manager_access" {
  count = var.secrets_manager_arns != null ? 1 : 0
  
  name        = "${var.project_name}-${var.environment}-lambda-secrets-manager-access-${var.name_suffix}"
  description = "Policy for Lambda to access Secrets Manager secrets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Effect   = "Allow"
        Resource = var.secrets_manager_arns
      }
    ]
  })
}

# Attach Secrets Manager policy to Lambda role
resource "aws_iam_role_policy_attachment" "secrets_manager_access" {
  count      = var.secrets_manager_arns != null ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.secrets_manager_access[0].arn
}

# X-Ray access policy
resource "aws_iam_role_policy_attachment" "xray_access" {
  count      = var.enable_xray ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}

# VPC access policy
resource "aws_iam_role_policy_attachment" "vpc_access" {
  count      = var.lambda_in_vpc ? 1 : 0
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}