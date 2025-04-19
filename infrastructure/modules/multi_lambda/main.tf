# infrastructure/modules/multi_lambda/main.tf
# This module creates multiple Lambda functions for the Vending Machine Verification workflow

locals {
  common_tags = merge(
    var.tags,
    {
      Environment = var.environment
      Name        = "${var.name_prefix}-lambda"
    }
  )
  
  lambda_functions = {
    initialize = {
      name         = "${var.name_prefix}-initialize"
      description  = "Validates inputs and prepares for processing"
      memory_size  = 256
      timeout      = 30
      handler      = "initialize.handler"
      image_uri    = var.ecr_repository_urls["initialize"]
    }
    fetch_images = {
      name         = "${var.name_prefix}-fetch-images"
      description  = "Retrieves reference and checking images from S3"
      memory_size  = 512
      timeout      = 60
      handler      = "fetchImages.handler"
      image_uri    = var.ecr_repository_urls["fetch-images"]
    }
    prepare_prompt = {
      name         = "${var.name_prefix}-prepare-prompt"
      description  = "Formats data for Bedrock model"
      memory_size  = 256
      timeout      = 30
      handler      = "preparePrompt.handler"
      image_uri    = var.ecr_repository_urls["prepare-prompt"]
    }
    invoke_bedrock = {
      name         = "${var.name_prefix}-invoke-bedrock"
      description  = "Calls Bedrock API for image analysis"
      memory_size  = 1024
      timeout      = 120
      handler      = "invokeBedrock.handler"
      image_uri    = var.ecr_repository_urls["invoke-bedrock"]
    }
    process_results = {
      name         = "${var.name_prefix}-process-results"
      description  = "Processes and parses results from Bedrock"
      memory_size  = 512
      timeout      = 60
      handler      = "processResults.handler"
      image_uri    = var.ecr_repository_urls["process-results"]
    }
    store_results = {
      name         = "${var.name_prefix}-store-results"
      description  = "Saves results to DynamoDB"
      memory_size  = 256
      timeout      = 30
      handler      = "storeResults.handler"
      image_uri    = var.ecr_repository_urls["store-results"]
    }
    notify = {
      name         = "${var.name_prefix}-notify"
      description  = "Sends notifications upon completion"
      memory_size  = 128
      timeout      = 30
      handler      = "notify.handler"
      image_uri    = var.ecr_repository_urls["notify"]
    }
    get_comparison = {
      name         = "${var.name_prefix}-get-comparison"
      description  = "Retrieves comparison results from DynamoDB"
      memory_size  = 256
      timeout      = 30
      handler      = "getComparison.handler"
      image_uri    = var.ecr_repository_urls["get-comparison"]
    }
    get_images = {
      name         = "${var.name_prefix}-get-images"
      description  = "Lists available images for comparison"
      memory_size  = 256
      timeout      = 30
      handler      = "getImages.handler"
      image_uri    = var.ecr_repository_urls["get-images"]
    }
  }
  
  # Use local variable for secrets_arn check to avoid count issues
  has_secrets_arn = var.enable_secrets_access
}

# IAM Role for Lambda functions
resource "aws_iam_role" "lambda_role" {
  name = "${var.name_prefix}-lambda-role"

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

  tags = local.common_tags
}

# Basic Lambda policy (CloudWatch Logs)
resource "aws_iam_policy" "lambda_basic_policy" {
  name        = "${var.name_prefix}-lambda-basic-policy"
  description = "Basic Lambda permissions for CloudWatch Logs"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

# S3 access policy
resource "aws_iam_policy" "s3_access_policy" {
  name        = "${var.name_prefix}-s3-access-policy"
  description = "S3 access permissions for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Resource = [
          var.s3_bucket_arn,
          "${var.s3_bucket_arn}/*"
        ]
      }
    ]
  })
}

# DynamoDB access policy
resource "aws_iam_policy" "dynamodb_access_policy" {
  name        = "${var.name_prefix}-dynamodb-access-policy"
  description = "DynamoDB access permissions for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = var.dynamodb_table_arn
      }
    ]
  })
}

# Bedrock access policy
resource "aws_iam_policy" "bedrock_access_policy" {
  name        = "${var.name_prefix}-bedrock-access-policy"
  description = "Bedrock access permissions for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = "arn:aws:bedrock:${var.aws_region}::foundation-model/anthropic.claude-3*"
      }
    ]
  })
}

# Secrets Manager access policy
resource "aws_iam_policy" "secrets_access_policy" {
  count       = local.has_secrets_arn ? 1 : 0
  name        = "${var.name_prefix}-secrets-access-policy"
  description = "Secrets Manager access permissions for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = local.has_secrets_arn ? [var.secrets_arn] : []
      }
    ]
  })
}

# Policy attachments
resource "aws_iam_role_policy_attachment" "lambda_basic_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_basic_policy.arn
}

resource "aws_iam_role_policy_attachment" "s3_access_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.s3_access_policy.arn
}

resource "aws_iam_role_policy_attachment" "dynamodb_access_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.dynamodb_access_policy.arn
}

resource "aws_iam_role_policy_attachment" "bedrock_access_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.bedrock_access_policy.arn
}

resource "aws_iam_role_policy_attachment" "secrets_access_attachment" {
  count      = local.has_secrets_arn ? 1 : 0
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.secrets_access_policy[0].arn
}

# Create the Lambda functions
resource "aws_lambda_function" "function" {
  for_each = var.skip_lambda_function_creation ? {} : local.lambda_functions
  
  function_name    = each.value.name
  description      = each.value.description
  role             = aws_iam_role.lambda_role.arn
  architectures    = var.architectures
  timeout          = each.value.timeout
  memory_size      = each.value.memory_size
  
  # Determine package type based on whether we have an image_uri
  package_type     = each.value.image_uri != null ? "Image" : "Zip"
  
  # Set only one of image_uri OR filename, not both
  image_uri        = each.value.image_uri != null ? each.value.image_uri : null
  
  # Only include these if we're not using a container image
  #filename         = each.value.image_uri == null ? var.filename : null
  handler          = each.value.image_uri == null ? each.value.handler : null
  #runtime          = each.value.image_uri == null ? var.runtime : null
  
  # Optional image configuration
  dynamic "image_config" {
    for_each = each.value.image_uri != null && var.image_command != null ? [1] : []
    content {
      command = var.image_command
    }
  }
  
  environment {
    variables = merge(
      {
        ENVIRONMENT = var.environment
        S3_BUCKET   = var.s3_bucket_name
        DYNAMO_TABLE = var.dynamodb_table_name
        BEDROCK_MODEL = var.bedrock_model_id
      },
      var.environment_variables
    )
  }
  
  tags = merge(
    local.common_tags,
    {
      Name = each.value.name
    }
  )
}