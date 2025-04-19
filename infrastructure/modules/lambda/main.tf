resource "aws_iam_role" "lambda_role" {
  name = "${var.function_name}-role"

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

  tags = merge(
    {
      Name        = "${var.function_name}-role"
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_iam_role_policy" "lambda_policy" {
  name = "${var.function_name}-policy"
  role = aws_iam_role.lambda_role.id

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
      },
      {
        Effect = "Allow"
        Action = var.additional_policy_actions
        Resource = var.additional_policy_resources
      }
    ]
  })
}

resource "aws_lambda_function" "function" {
  function_name    = var.function_name
  role            = aws_iam_role.lambda_role.arn
  architectures   = var.architectures
  timeout         = var.timeout
  memory_size     = var.memory_size
  
  # Explicitly set package_type based on deployment method
  package_type    = var.ecr_image_uri != null ? "Image" : "Zip"
  
  # Use either image_uri or filename based on deployment type
  dynamic "image_config" {
    for_each = var.ecr_image_uri != null ? [1] : []
    content {
      command = var.image_command
    }
  }
  
  # Only set image_uri when using container image
  image_uri = var.ecr_image_uri
  
  # Only use these if not using container image
  filename = var.ecr_image_uri == null ? var.filename : null
  handler = var.ecr_image_uri == null ? var.handler : null
  runtime = var.ecr_image_uri == null ? var.runtime : null

  environment {
    variables = merge(
      {
        ENVIRONMENT = var.environment
      },
      var.environment_variables
    )
  }

  tags = merge(
    {
      Name = var.function_name
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_lambda_permission" "api_gw" {
  count         = var.enable_api_gateway_integration ? 1 : 0
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.function.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = var.api_gateway_source_arn
}
