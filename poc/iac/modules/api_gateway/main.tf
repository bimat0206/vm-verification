# modules/api_gateway/main.tf

# This file has been reorganized into smaller, more manageable files:
# - resources.tf: Contains all API Gateway resource definitions
# - methods.tf: Contains all API Gateway method and integration definitions
# - deployment.tf: Contains deployment, stage, and API key configurations
# - models.tf: Contains model definitions for request/response validation
# - validators.tf: Contains request validators
# - error_responses.tf: Contains custom error responses
# - locals.tf: Contains local variable definitions
# - variables.tf: Contains input variable definitions
# - output.tf: Contains output definitions

# This reorganization improves maintainability and makes troubleshooting easier
# by grouping related resources together in logical files.

# IAM Role for API Gateway to invoke Step Functions
resource "aws_iam_role" "api_gateway_step_functions_role" {
  name = "${var.api_name}-step-functions-role"

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
  name        = "${var.api_name}-step-functions-policy"
  description = "Allows API Gateway to start Step Functions executions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "states:StartExecution"
        Effect   = "Allow"
        Resource = var.step_functions_state_machine_arn
      }
    ]
  })
}

# Attach policy to API Gateway role
resource "aws_iam_role_policy_attachment" "api_gateway_step_functions_attachment" {
  role       = aws_iam_role.api_gateway_step_functions_role.name
  policy_arn = aws_iam_policy.api_gateway_step_functions_policy.arn
}