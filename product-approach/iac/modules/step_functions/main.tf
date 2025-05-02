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
  
  # Don't add tags here as they're provided by default_tags in the provider
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

# Attach policies to the Step Functions IAM role
resource "aws_iam_role_policy_attachment" "lambda_invoke_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.lambda_invoke_policy.arn
}

resource "aws_iam_role_policy_attachment" "cloudwatch_logs_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.cloudwatch_logs_policy.arn
}

# CloudWatch Log Group for Step Functions
resource "aws_cloudwatch_log_group" "step_functions_logs" {
  name              = "/aws/states/${var.state_machine_name}"
  retention_in_days = var.log_retention_days
  
  # Don't add tags here as they're provided by default_tags in the provider
}

# Step Functions State Machine
resource "aws_sfn_state_machine" "verification_workflow" {
  name     = var.state_machine_name
  role_arn = aws_iam_role.step_functions_role.arn

  definition = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns = var.lambda_function_arns
  })

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.step_functions_logs.arn}:*"
    include_execution_data = true
    level                  = var.log_level
  }

  type = "STANDARD"

  tags = merge(
    var.common_tags,
    {
      Name = var.state_machine_name
    }
  )
}

# Create the state machine definition template file
resource "local_file" "state_machine_definition" {
  count = var.create_definition_file ? 1 : 0
  
  content  = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns = var.lambda_function_arns
  })
  filename = "${path.module}/generated_definition.json"
}
