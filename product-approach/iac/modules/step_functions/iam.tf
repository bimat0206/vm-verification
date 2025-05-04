# modules/step_functions/iam.tf

# IAM Role for Step Functions
resource "aws_iam_role" "step_functions_role" {
  name = "${var.state_machine_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "states.amazonaws.com"
        }
      }
    ]
  })
  
  tags = var.common_tags
}

# IAM Policy for Step Functions to invoke Lambda functions
resource "aws_iam_policy" "lambda_invoke_policy" {
  name        = "${var.state_machine_name}-lambda-invoke-policy"
  description = "Allows Step Functions to invoke Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "lambda:InvokeFunction"
        Effect   = "Allow"
        Resource = values(var.lambda_function_arns)
      }
    ]
  })
}

# IAM Policy for Step Functions logging
resource "aws_iam_policy" "cloudwatch_logs_policy" {
  name        = "${var.state_machine_name}-cloudwatch-logs-policy"
  description = "Allows Step Functions to write logs to CloudWatch"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogDelivery",
          "logs:GetLogDelivery",
          "logs:UpdateLogDelivery",
          "logs:DeleteLogDelivery",
          "logs:ListLogDeliveries",
          "logs:PutResourcePolicy",
          "logs:DescribeResourcePolicies",
          "logs:DescribeLogGroups"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# IAM Policy for X-Ray tracing (added for better monitoring)
resource "aws_iam_policy" "xray_policy" {
  name        = "${var.state_machine_name}-xray-policy"
  description = "Allows Step Functions to use X-Ray tracing"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "xray:PutTraceSegments",
          "xray:PutTelemetryRecords",
          "xray:GetSamplingRules",
          "xray:GetSamplingTargets"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Add DynamoDB policy for direct integration
resource "aws_iam_policy" "dynamodb_policy" {
  name        = "${var.state_machine_name}-dynamodb-policy"
  description = "Allows Step Functions to access DynamoDB tables"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query"
        ],
        Effect   = "Allow",
        Resource = var.dynamodb_table_arns
      }
    ]
  })
}

# Attach policies to the Step Functions IAM role
resource "aws_iam_role_policy_attachment" "lambda_invoke_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.lambda_invoke_policy.arn
}

resource "aws_iam_role_policy_attachment" "cloudwatch_logs_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.cloudwatch_logs_policy.arn
}

resource "aws_iam_role_policy_attachment" "xray_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.xray_policy.arn
}

# Add the attachment for the DynamoDB policy
resource "aws_iam_role_policy_attachment" "dynamodb_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.dynamodb_policy.arn
}

# CloudWatch Log Group for Step Functions
resource "aws_cloudwatch_log_group" "step_functions_logs" {
  name              = "/aws/states/${var.state_machine_name}"
  retention_in_days = var.log_retention_days
  
  tags = var.common_tags
}

# Create policy for Lambda to invoke Step Functions
resource "aws_iam_policy" "lambda_to_sfn_policy" {
  name        = "${var.state_machine_name}-lambda-start-execution-policy"
  description = "Allows Lambda functions to start Step Functions executions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "states:StartExecution"
        Effect   = "Allow"
        Resource = "arn:aws:states:${var.region}:*:stateMachine:${var.state_machine_name}"
      }
    ]
  })
}


