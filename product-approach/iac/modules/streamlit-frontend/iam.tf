# Add Secret Manager permissions if needed
# This policy is created automatically for API key access
# The API key is stored in Secrets Manager
resource "aws_iam_policy" "streamlit_secretsmanager_policy" {
  # Create this policy if API Gateway is enabled and uses API key
  # The API key is stored in Secrets Manager with name from environment variable API_KEY_SECRET_NAME
  count = contains(keys(local.environment_variables), "API_KEY_SECRET_NAME") ? 1 : 0

  name        = "${local.service_name}-secretsmanager-policy"
  description = "Policy for Streamlit App Runner to access Secrets Manager"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:secretsmanager:*:*:secret:*"
      }
    ]
  })

  tags = var.common_tags
}

resource "aws_iam_role_policy_attachment" "streamlit_secretsmanager_attachment" {
  count      = contains(keys(local.environment_variables), "API_KEY_SECRET_NAME") ? 1 : 0
  role       = aws_iam_role.streamlit_instance_role.name
  policy_arn = aws_iam_policy.streamlit_secretsmanager_policy[0].arn
}
