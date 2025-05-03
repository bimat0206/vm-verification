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

# CloudWatch Log Group for Step Functions
resource "aws_cloudwatch_log_group" "step_functions_logs" {
  name              = "/aws/states/${var.state_machine_name}"
  retention_in_days = var.log_retention_days
  
  tags = var.common_tags
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

  tracing_configuration {
    enabled = var.enable_x_ray_tracing
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

# API Gateway resource
resource "aws_api_gateway_resource" "step_functions" {
  count       = var.create_api_gateway_integration ? 1 : 0
  rest_api_id = var.api_gateway_id
  parent_id   = var.api_gateway_root_resource_id
  path_part   = "executions"
}

# API Gateway method for starting executions
resource "aws_api_gateway_method" "step_functions_start" {
  count         = var.create_api_gateway_integration ? 1 : 0
  rest_api_id   = var.api_gateway_id
  resource_id   = aws_api_gateway_resource.step_functions[0].id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = true
}

# IAM Role for API Gateway to invoke Step Functions
resource "aws_iam_role" "api_gateway_step_functions_role" {
  count = var.create_api_gateway_integration ? 1 : 0
  name  = "${var.state_machine_name}-apigw-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })
  
  tags = var.common_tags
}

# IAM Policy for API Gateway to invoke Step Functions
resource "aws_iam_policy" "api_gateway_step_functions_policy" {
  count       = var.create_api_gateway_integration ? 1 : 0
  name        = "${var.state_machine_name}-apigw-policy"
  description = "Allows API Gateway to invoke Step Functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "states:StartExecution"
        Effect   = "Allow"
        Resource = aws_sfn_state_machine.verification_workflow.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "api_gateway_step_functions_attachment" {
  count      = var.create_api_gateway_integration ? 1 : 0
  role       = aws_iam_role.api_gateway_step_functions_role[0].name
  policy_arn = aws_iam_policy.api_gateway_step_functions_policy[0].arn
}
